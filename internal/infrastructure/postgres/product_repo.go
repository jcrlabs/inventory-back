package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jonathanCaamano/inventory-back/internal/domain/product"
)

type ProductRepo struct {
	pool *pgxpool.Pool
}

func NewProductRepo(pool *pgxpool.Pool) *ProductRepo {
	return &ProductRepo{pool: pool}
}

func (r *ProductRepo) Create(ctx context.Context, p *product.Product) (*product.Product, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var id uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO products (group_id, name, description, price, paid, status, entry_date, exit_date, observations)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id`,
		p.GroupID, p.Name, p.Description, p.Price, p.Paid, string(p.Status), p.EntryDate, p.ExitDate, p.Observations,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	if p.Contact != nil {
		_, err = tx.Exec(ctx, `
			INSERT INTO product_contacts (product_id, first_name, last_name, phone_number)
			VALUES ($1,$2,$3,$4)`, id, p.Contact.FirstName, p.Contact.LastName, p.Contact.PhoneNumber)
		if err != nil {
			return nil, err
		}
	}
	for _, img := range p.Images {
		_, err = tx.Exec(ctx, `
			INSERT INTO product_images (product_id, image_url, position)
			VALUES ($1,$2,$3)`, id, img.URL, img.Position)
		if err != nil {
			return nil, err
		}
	}
	_, _ = tx.Exec(ctx, `INSERT INTO audit_logs (entity_id, entity_type, action) VALUES ($1,'product','create')`, id)
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *ProductRepo) GetByID(ctx context.Context, id uuid.UUID) (*product.Product, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, group_id, name, coalesce(description,''), price, paid, status, entry_date, exit_date, coalesce(observations,''), created_at, updated_at
		FROM products WHERE id=$1`, id)
	var p product.Product
	var status string
	err := row.Scan(&p.ID, &p.GroupID, &p.Name, &p.Description, &p.Price, &p.Paid, &status, &p.EntryDate, &p.ExitDate, &p.Observations, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	st, err := product.ParseStatus(status)
	if err != nil {
		return nil, err
	}
	p.Status = st

	cRow := r.pool.QueryRow(ctx, `SELECT first_name,last_name,phone_number FROM product_contacts WHERE product_id=$1`, id)
	var c product.Contact
	if err := cRow.Scan(&c.FirstName, &c.LastName, &c.PhoneNumber); err == nil {
		p.Contact = &c
	}

	imgs, err := r.pool.Query(ctx, `SELECT id, image_url, position, created_at FROM product_images WHERE product_id=$1 ORDER BY position ASC, created_at ASC`, id)
	if err != nil {
		return nil, err
	}
	defer imgs.Close()
	for imgs.Next() {
		var img product.Image
		if err := imgs.Scan(&img.ID, &img.URL, &img.Position, &img.CreatedAt); err != nil {
			return nil, err
		}
		p.Images = append(p.Images, img)
	}

	return &p, nil
}

func (r *ProductRepo) Update(ctx context.Context, p *product.Product) (*product.Product, error) {
	_, err := r.pool.Exec(ctx, `
		UPDATE products SET name=$2, description=$3, price=$4, paid=$5, status=$6, entry_date=$7, exit_date=$8, observations=$9
		WHERE id=$1`,
		p.ID, p.Name, p.Description, p.Price, p.Paid, string(p.Status), p.EntryDate, p.ExitDate, p.Observations)
	if err != nil {
		return nil, err
	}
	_, _ = r.pool.Exec(ctx, `INSERT INTO audit_logs (entity_id, entity_type, action) VALUES ($1,'product','update')`, p.ID)
	return r.GetByID(ctx, p.ID)
}

func (r *ProductRepo) Delete(ctx context.Context, id uuid.UUID) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM products WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("not_found")
	}
	_, _ = r.pool.Exec(ctx, `INSERT INTO audit_logs (entity_id, entity_type, action) VALUES ($1,'product','delete')`, id)
	return nil
}

func (r *ProductRepo) GetGroupID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	var gid uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT group_id FROM products WHERE id=$1`, id).Scan(&gid)
	return gid, err
}

func (r *ProductRepo) AddImage(ctx context.Context, productID uuid.UUID, img product.Image) (*product.Product, error) {
	_, err := r.pool.Exec(ctx, `INSERT INTO product_images (product_id, image_url, position) VALUES ($1,$2,$3)`, productID, img.URL, img.Position)
	if err != nil {
		return nil, err
	}
	_, _ = r.pool.Exec(ctx, `INSERT INTO audit_logs (entity_id, entity_type, action) VALUES ($1,'product','add_image')`, productID)
	return r.GetByID(ctx, productID)
}

func (r *ProductRepo) UpsertContact(ctx context.Context, productID uuid.UUID, c product.Contact) (*product.Product, error) {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO product_contacts (product_id, first_name, last_name, phone_number)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (product_id) DO UPDATE SET first_name=EXCLUDED.first_name, last_name=EXCLUDED.last_name, phone_number=EXCLUDED.phone_number`,
		productID, c.FirstName, c.LastName, c.PhoneNumber)
	if err != nil {
		return nil, err
	}
	_, _ = r.pool.Exec(ctx, `INSERT INTO audit_logs (entity_id, entity_type, action) VALUES ($1,'product','upsert_contact')`, productID)
	return r.GetByID(ctx, productID)
}

func (r *ProductRepo) Search(ctx context.Context, q product.SearchQuery, isAdmin bool) (items []*product.Product, total int, err error) {
	where := []string{}
	args := []any{}
	idx := 1

	if !isAdmin {
		where = append(where, fmt.Sprintf("group_id=$%d", idx))
		args = append(args, q.GroupID)
		idx++
	} else if q.GroupID != uuid.Nil {
		where = append(where, fmt.Sprintf("group_id=$%d", idx))
		args = append(args, q.GroupID)
		idx++
	}

	if strings.TrimSpace(q.Search) != "" {
		where = append(where, fmt.Sprintf("search_vector @@ plainto_tsquery('simple', $%d)", idx))
		args = append(args, q.Search)
		idx++
	}
	if strings.TrimSpace(q.Status) != "" {
		where = append(where, fmt.Sprintf("status=$%d", idx))
		args = append(args, q.Status)
		idx++
	}
	if strings.TrimSpace(q.Paid) != "" {
		if q.Paid == "true" {
			where = append(where, "paid=true")
		} else if q.Paid == "false" {
			where = append(where, "paid=false")
		}
	}
	if q.MinPrice != nil {
		where = append(where, fmt.Sprintf("price >= $%d", idx))
		args = append(args, *q.MinPrice)
		idx++
	}
	if q.MaxPrice != nil {
		where = append(where, fmt.Sprintf("price <= $%d", idx))
		args = append(args, *q.MaxPrice)
		idx++
	}
	if q.FromEntry != nil {
		where = append(where, fmt.Sprintf("entry_date >= $%d", idx))
		args = append(args, *q.FromEntry)
		idx++
	}
	if q.ToEntry != nil {
		where = append(where, fmt.Sprintf("entry_date <= $%d", idx))
		args = append(args, *q.ToEntry)
		idx++
	}

	w := ""
	if len(where) > 0 {
		w = "WHERE " + strings.Join(where, " AND ")
	}

	countQ := `SELECT count(*) FROM products ` + w
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	order := "ORDER BY created_at DESC"
	sort := strings.TrimSpace(q.Sort)
	switch sort {
	case "created_at_asc":
		order = "ORDER BY created_at ASC"
	case "price_asc":
		order = "ORDER BY price ASC"
	case "price_desc":
		order = "ORDER BY price DESC"
	case "entry_date_asc":
		order = "ORDER BY entry_date ASC"
	case "entry_date_desc":
		order = "ORDER BY entry_date DESC"
	}

	args2 := append([]any{}, args...)
	args2 = append(args2, q.Limit, q.Offset)
	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	query := fmt.Sprintf(`
		SELECT id FROM products %s %s LIMIT $%d OFFSET $%d`, w, order, limitPos, offsetPos)

	rows, err := r.pool.Query(ctx, query, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	ids := []uuid.UUID{}
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, 0, err
		}
		ids = append(ids, id)
	}

	items = make([]*product.Product, 0, len(ids))
	for _, id := range ids {
		p, err := r.GetByID(ctx, id)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, p)
	}
	return items, total, nil
}

var _ product.Repository = (*ProductRepo)(nil)
