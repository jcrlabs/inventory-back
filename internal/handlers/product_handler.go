package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jonathanCaamano/inventory-back/internal/middleware"
	"github.com/jonathanCaamano/inventory-back/internal/models"
	"github.com/jonathanCaamano/inventory-back/internal/repository"
	"github.com/jonathanCaamano/inventory-back/internal/services"
)

type ProductHandler struct {
	productRepo  *repository.ProductRepository
	categoryRepo *repository.CategoryRepository
	minioSvc     *services.MinIOService
}

func NewProductHandler(
	productRepo *repository.ProductRepository,
	categoryRepo *repository.CategoryRepository,
	minioSvc *services.MinIOService,
) *ProductHandler {
	return &ProductHandler{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		minioSvc:     minioSvc,
	}
}

type CreateProductRequest struct {
	Name        string     `json:"name" binding:"required,min=1,max=200"`
	Description string     `json:"description"`
	Price       float64    `json:"price" binding:"gte=0"`
	Stock       int        `json:"stock" binding:"gte=0"`
	SKU         string     `json:"sku"`
	CategoryID  *uuid.UUID `json:"category_id"`
	Active      *bool      `json:"active"`
}

type UpdateProductRequest struct {
	Name        string     `json:"name" binding:"omitempty,min=1,max=200"`
	Description string     `json:"description"`
	Price       *float64   `json:"price" binding:"omitempty,gte=0"`
	Stock       *int       `json:"stock" binding:"omitempty,gte=0"`
	SKU         string     `json:"sku"`
	CategoryID  *uuid.UUID `json:"category_id"`
	Active      *bool      `json:"active"`
}

func (h *ProductHandler) List(c *gin.Context) {
	filter := repository.ProductFilter{
		Search:   c.Query("search"),
		Page:     1,
		PageSize: 20,
	}

	if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 {
		filter.Page = page
	}
	if size, err := strconv.Atoi(c.Query("page_size")); err == nil && size > 0 && size <= 100 {
		filter.PageSize = size
	}
	if catID, err := uuid.Parse(c.Query("category_id")); err == nil {
		filter.CategoryID = &catID
	}
	if activeStr := c.Query("active"); activeStr != "" {
		active := activeStr == "true"
		filter.Active = &active
	}

	products, total, err := h.productRepo.FindAll(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch products"})
		return
	}

	// Enrich with presigned image URLs
	for i := range products {
		if products[i].ImageKey != "" && h.minioSvc != nil {
			url, _ := h.minioSvc.GetPresignedURL(products[i].ImageKey, time.Hour)
			products[i].ImageURL = url
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      products,
		"total":     total,
		"page":      filter.Page,
		"page_size": filter.PageSize,
	})
}

func (h *ProductHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	product, err := h.productRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	if product.ImageKey != "" && h.minioSvc != nil {
		url, _ := h.minioSvc.GetPresignedURL(product.ImageKey, time.Hour)
		product.ImageURL = url
	}

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	active := true
	if req.Active != nil {
		active = *req.Active
	}

	product := &models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		SKU:         req.SKU,
		CategoryID:  req.CategoryID,
		CreatedByID: userID,
		Active:      active,
	}

	if err := h.productRepo.Create(product); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "failed to create product (SKU may already exist)"})
		return
	}

	// Reload with associations
	product, _ = h.productRepo.FindByID(product.ID)
	c.JSON(http.StatusCreated, product)
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	product, err := h.productRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}
	if req.SKU != "" {
		product.SKU = req.SKU
	}
	if req.CategoryID != nil {
		product.CategoryID = req.CategoryID
	}
	if req.Active != nil {
		product.Active = *req.Active
	}

	if err := h.productRepo.Update(product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update product"})
		return
	}

	if product.ImageKey != "" && h.minioSvc != nil {
		url, _ := h.minioSvc.GetPresignedURL(product.ImageKey, time.Hour)
		product.ImageURL = url
	}

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	product, err := h.productRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	// Delete image from MinIO if exists
	if product.ImageKey != "" && h.minioSvc != nil {
		_ = h.minioSvc.DeleteObject(product.ImageKey)
	}

	if err := h.productRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}

func (h *ProductHandler) UploadImage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	product, err := h.productRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file required"})
		return
	}
	defer file.Close()

	if h.minioSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "image storage not available"})
		return
	}

	// Delete old image if exists
	if product.ImageKey != "" {
		_ = h.minioSvc.DeleteObject(product.ImageKey)
	}

	objectKey, err := h.minioSvc.UploadProductImage(file, header)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product.ImageKey = objectKey
	if err := h.productRepo.Update(product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update product image"})
		return
	}

	imageURL, _ := h.minioSvc.GetPresignedURL(objectKey, time.Hour)
	c.JSON(http.StatusOK, gin.H{"image_url": imageURL})
}

// Category handlers

type CategoryHandler struct {
	categoryRepo *repository.CategoryRepository
}

func NewCategoryHandler(categoryRepo *repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{categoryRepo: categoryRepo}
}

type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description"`
}

func (h *CategoryHandler) List(c *gin.Context) {
	categories, err := h.categoryRepo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch categories"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": categories, "total": len(categories)})
}

func (h *CategoryHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	category, err := h.categoryRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	c.JSON(http.StatusOK, category)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category := &models.Category{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.categoryRepo.Create(category); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "category name already exists"})
		return
	}

	c.JSON(http.StatusCreated, category)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	category, err := h.categoryRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category.Name = req.Name
	category.Description = req.Description

	if err := h.categoryRepo.Update(category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update category"})
		return
	}

	c.JSON(http.StatusOK, category)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	if err := h.categoryRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "category deleted"})
}
