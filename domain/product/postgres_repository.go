package product

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
)

type postgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) GetProductByID(ctx context.Context, id int) (*Product, error) {
	var product Product
	query := `SELECT id, name, description, price, sku, image_url, category_id, created_at, updated_at
		FROM products
		WHERE id = $1 AND deleted_at IS NULL
	`
	err := r.db.GetContext(ctx, &product, query, id)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *postgresRepository) GetAllProducts(ctx context.Context, limit, offset int) ([]*Product, error) {
	query := `
        SELECT id, name, description, price, sku 
        FROM products
		WHERE deleted_at IS NULL
        LIMIT $1 OFFSET $2
    `

	var products []*Product
	err := r.db.SelectContext(ctx, &products, query, limit, offset)
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (r *postgresRepository) CreateProduct(ctx context.Context, p *Product) (*Product, error) {
	query := `
        INSERT INTO products (name, description, price, sku, image_url, category_id) 
        VALUES ($1, $2, $3, $4, $5, $6) 
        RETURNING id, name, description, price, sku, image_url, category_id, created_at, updated_at
    `

	err := r.db.QueryRowContext(
		ctx,
		query,
		p.Name,
		p.Description,
		p.Price,
		p.SKU,
		p.ImageURL,
		p.CategoryID,
	).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.SKU, &p.ImageURL, &p.CategoryID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (r *postgresRepository) UpdateProduct(ctx context.Context, p *Product) (*Product, error) {
	query := `
        UPDATE products 
        SET name = $1, description = $2, price = $3, sku = $4, image_url = $5, category_id = $6, updated_at = NOW()
        WHERE id = $7 AND deleted_at IS NULL
    `

	result, err := r.db.ExecContext(ctx, query, p.Name, p.Description, p.Price, p.SKU, p.ImageURL, p.CategoryID, p.ID)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, errors.New("product not found or no changes made")
	}

	return r.GetProductByID(ctx, p.ID)
}

func (r *postgresRepository) GetProductsByCategory(ctx context.Context, categoryID int) ([]*Product, error) {
	query := `
        SELECT id, name, description, price, sku, image_url, category_id, created_at, updated_at
        FROM products
        WHERE category_id = $1 AND deleted_at IS NULL
    `
	var products []*Product
	err := r.db.SelectContext(ctx, &products, query, categoryID)
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (r *postgresRepository) DeleteProduct(ctx context.Context, id int) error {
	query := `UPDATE products SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("product not found")
	}

	return nil
}

func (r *postgresRepository) CountProducts(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*) as total_products
		FROM products
        WHERE deleted_at IS NULL
	`

	var count int
	err := r.db.SelectContext(ctx, &count, query)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *postgresRepository) SearchProductsByName(ctx context.Context, searchTerm string) ([]*Product, error) {
	query := `SELECT *
		FROM products
		WHERE name ILIKE $1 AND deleted_at IS NULL
	`
	searchTerm = "%" + searchTerm + "%"

	var products []*Product
	err := r.db.SelectContext(ctx, &products, query, searchTerm)
	if err != nil {
		return nil, err
	}

	return products, nil
}
