package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/jonathanCaamano/inventory-back/internal/http/dto"
	"github.com/jonathanCaamano/inventory-back/internal/service"
)

type productRow struct {
	ID           uuid.UUID    `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Price        float64      `json:"price"`
	Paid         bool         `json:"paid"`
	Status       string       `json:"status"`
	EntryDate    time.Time    `json:"entry_date"`
	ExitDate     *time.Time   `json:"exit_date"`
	Observations string       `json:"observations"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	Contact      *dto.Contact `json:"contact,omitempty"`
	Images       []dto.Image  `json:"images,omitempty"`
}

func (r *ProductRepo) Create(ctx context.Context, in dto.ProductCreate) (any, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var id uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO products (name, description, price, paid, status, entry_date, exit_date, observations)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id`,
		in.Name, in.Description, in.Price, in.Paid, in.Status, in.EntryDate, in.ExitDate, in.Observations,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	if in.Contact != nil {
		_, err = tx.Exec(ctx, `
			INSERT INTO product_contacts (product_id, first_name, last_name, phone_number)
			VALUES ($1,$2,$3,$4)`,
			id, in.Contact.FirstName, in.Contact.LastName, in.Contact.PhoneNumber,
		)
		if err != nil {
			return nil, err
		}
	}

	for _, img := range in.Images {
		_, err = tx.Exec(ctx, `
			INSERT INTO product_images (product_id, image_url, position)
			VALUES ($1,$2,$3)`,
			id, img.ImageURL, img.Position,
		)
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

func (r *ProductRepo) GetByID(ctx context.Context, id uuid.UUID) (any, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, coalesce(description,''), price, paid, status, entry_date, exit_date, coalesce(observations,''), created_at, updated_at
		FROM products WHERE id=$1`, id)

	var p productRow
	if err := row.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Paid, &p.Status, &p.EntryDate, &p.ExitDate, &p.Observations, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}

	cRow := r.pool.QueryRow(ctx, `SELECT first_name,last_name,phone_number FROM product_contacts WHERE product_id=$1`, id)
	var c dto.Contact
	if err := cRow.Scan(&c.FirstName, &c.LastName, &c.PhoneNumber); err == nil {
		p.Contact = &c
	}

	imgs, err := r.pool.Query(ctx, `SELECT image_url, position FROM product_images WHERE product_id=$1 ORDER BY position ASC, created_at ASC`, id)
	if err != nil {
		return nil, err
	}
	defer imgs.Close()

	for imgs.Next() {
		var img dto.Image
		if err := imgs.Scan(&img.ImageURL, &img.Position); err != nil {
			return nil, err
		}
		p.Images = append(p.Images, img)
	}

	return p, nil
}

func (r *ProductRepo) Update(ctx context.Context, id uuid.UUID, in dto.ProductUpdate) (any, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	ct, err := tx.Exec(ctx, `
		UPDATE products SET name=$1, description=$2, price=$3, paid=$4, status=$5, entry_date=$6, exit_date=$7, observations=$8
		WHERE id=$9`,
		in.Name, in.Description, in.Price, in.Paid, in.Status, in.EntryDate, in.ExitDate, in.Observations, id,
	)
	if err != nil {
		return nil, err
	}
	if ct.RowsAffected() == 0 {
		return nil, errors.New("not_found")
	}

	if in.Contact != nil {
		_, err = tx.Exec(ctx, `
			INSERT INTO product_contacts (product_id, first_name, last_name, phone_number)
			VALUES ($1,$2,$3,$4)
			ON CONFLICT (product_id) DO UPDATE SET first_name=EXCLUDED.first_name, last_name=EXCLUDED.last_name, phone_number=EXCLUDED.phone_number`,
			id, in.Contact.FirstName, in.Contact.LastName, in.Contact.PhoneNumber,
		)
		if err != nil {
			return nil, err
		}
	}

	_, _ = tx.Exec(ctx, `INSERT INTO audit_logs (entity_id, entity_type, action) VALUES ($1,'product','update')`, id)

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
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

func (r *ProductRepo) Search(ctx context.Context, req service.SearchRequest) (any, error) {
	where := []string{"1=1"}
	args := []any{}
	arg := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	if req.Status != "" {
		where = append(where, "status="+arg(req.Status))
	}
	if req.Paid != "" {
		if req.Paid == "true" || req.Paid == "false" {
			where = append(where, "paid="+arg(req.Paid == "true"))
		} else {
			return nil, errors.New("invalid_paid")
		}
	}
	if req.MinPrice != nil {
		where = append(where, "price>="+arg(*req.MinPrice))
	}
	if req.MaxPrice != nil {
		where = append(where, "price<="+arg(*req.MaxPrice))
	}
	if req.FromEntry != nil {
		where = append(where, "entry_date>="+arg(*req.FromEntry))
	}
	if req.ToEntry != nil {
		where = append(where, "entry_date<="+arg(*req.ToEntry))
	}

	sort := "updated_at DESC"
	selectRank := ""
	searchParamIdx := -1
	if req.Search != "" {
		args = append(args, req.Search)
		searchParamIdx = len(args)
		selectRank = fmt.Sprintf(", ts_rank(search_vector, websearch_to_tsquery('simple', $%d)) AS rank", searchParamIdx)
		where = append(where, fmt.Sprintf("search_vector @@ websearch_to_tsquery('simple', $%d)", searchParamIdx))
		sort = "rank DESC, updated_at DESC"
	}

	switch req.Sort {
	case "price_asc":
		sort = "price ASC, updated_at DESC"
	case "price_desc":
		sort = "price DESC, updated_at DESC"
	case "newest":
		sort = "created_at DESC"
	}

	limit := arg(req.Limit)
	offset := arg(req.Offset)

	q := `
		SELECT id, name, coalesce(description,''), price, paid, status, entry_date, exit_date, coalesce(observations,''), created_at, updated_at
		` + selectRank + `
		FROM products
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY ` + sort + `
		LIMIT ` + limit + ` OFFSET ` + offset

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []productRow{}
	for rows.Next() {
		var p productRow
		if selectRank != "" {
			var rank float32
			if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Paid, &p.Status, &p.EntryDate, &p.ExitDate, &p.Observations, &p.CreatedAt, &p.UpdatedAt, &rank); err != nil {
				return nil, err
			}
		} else {
			if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Paid, &p.Status, &p.EntryDate, &p.ExitDate, &p.Observations, &p.CreatedAt, &p.UpdatedAt); err != nil {
				return nil, err
			}
		}
		items = append(items, p)
	}

	return map[string]any{"items": items, "limit": req.Limit, "offset": req.Offset}, nil
}

func (r *ProductRepo) AddImage(ctx context.Context, id uuid.UUID, img dto.Image) (any, error) {
	_, err := r.pool.Exec(ctx, `INSERT INTO product_images (product_id, image_url, position) VALUES ($1,$2,$3)`, id, img.ImageURL, img.Position)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *ProductRepo) UpsertContact(ctx context.Context, id uuid.UUID, c dto.Contact) (any, error) {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO product_contacts (product_id, first_name, last_name, phone_number)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (product_id) DO UPDATE SET first_name=EXCLUDED.first_name, last_name=EXCLUDED.last_name, phone_number=EXCLUDED.phone_number`,
		id, c.FirstName, c.LastName, c.PhoneNumber,
	)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}
