package product

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type postgresRepository struct {
	db  *sqlx.DB
	log zerolog.Logger
}

var ErrNotExist = errors.New("resource does not exist")

func NewPostgresRepository(ctx context.Context, db *sqlx.DB, log zerolog.Logger) *postgresRepository {
	logger := log.With().Str("repository", "postgresRepository").Logger()

	// ping db
	if err := db.PingContext(ctx); err != nil {
		logger.Fatal().Err(fmt.Errorf("failed to connect to postgres: %w", err)).Msg("something went wrong!")
	}

	return &postgresRepository{db: db, log: logger}
}

func (r *postgresRepository) GetProductByID(ctx context.Context, id int) (*Product, error) {
	var product Product
	query := `SELECT id, name, description, price, sku, image_url, category_id, created_at, updated_at
		FROM products
		WHERE id = $1 AND deleted_at IS NULL
	`
	err := r.db.GetContext(ctx, &product, query, id)
	if err != nil {
		r.log.Err(err).Str("method", "GetProductByID").Msg(err.Error())
		return nil, ErrNotExist
	}
	return &product, nil
}

// gets product, deleted or not.
func (r *postgresRepository) getProductByID(ctx context.Context, id int) (*Product, error) {
	var product Product
	query := `SELECT *
		FROM products
		WHERE id = $1
	`
	err := r.db.GetContext(ctx, &product, query, id)
	if err != nil {
		r.log.Err(err).Str("method", "GetProductByID").Msg(err.Error())
		return nil, ErrNotExist
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

	_, err := r.db.ExecContext(ctx, query, p.Name, p.Description, p.Price, p.SKU, p.ImageURL, p.CategoryID, p.ID)
	if err != nil {
		return nil, err
	}

	up, err := r.GetProductByID(ctx, p.ID)
	if err != nil {
		r.log.Err(err).Str("method", "UpdateProduct").Msg(err.Error())
		return nil, ErrNotExist
	}

	return up, nil
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
		r.log.Err(err).Str("method", "GetProductsByCategory").Msg(err.Error())
		return nil, ErrNotExist
	}
	return products, nil
}

func (r *postgresRepository) DeleteProduct(ctx context.Context, id int) error {
	query := `UPDATE products SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	p, err := r.getProductByID(ctx, id)
	if err != nil {
		r.log.Err(err).Str("method", "DeleteProduct").Msg(err.Error())
		return err
	}

	if p == nil {
		return ErrNotExist
	}

	r.log.Info().Msgf("My struct: %v", p)

	if p.DeletedAt == "" {
		return errors.New("delete failed")
	}

	return nil
}

func (r *postgresRepository) CountProducts(ctx context.Context) (int, error) {
	query := `
		SELECT count(*) as total_products
		FROM products
        WHERE deleted_at IS NULL;
	`

	var count int
	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *postgresRepository) SearchProductsByName(ctx context.Context, searchTerm string) ([]*Product, error) {
	query := `SELECT id, name, description, price, sku, image_url, category_id, created_at, updated_at
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

func (r *postgresRepository) GetCategoryByID(ctx context.Context, id int) (*Category, error) {
	c := &Category{}
	query := `SELECT id, name, created_at, updated_at
		FROM categories
		WHERE id = $1
	`
	err := r.db.GetContext(ctx, c, query, id)
	if err != nil {
		r.log.Err(err).Str("method", "GetCategoryByID").Msg(err.Error())
		return nil, ErrNotExist
	}
	return c, nil
}

func (r *postgresRepository) GetAllCategories(ctx context.Context, limit, offset int) ([]*Category, error) {
	query := `
        SELECT id, name, created_at, updated_at
        FROM categories
        LIMIT $1 OFFSET $2
    `

	var ctgry []*Category
	err := r.db.SelectContext(ctx, &ctgry, query, limit, offset)
	if err != nil {
		return nil, err
	}

	return ctgry, nil
}

func (r *postgresRepository) CreateCategory(ctx context.Context, name string) (*Category, error) {
	c := &Category{}
	query := `
        INSERT INTO categories (name) 
        VALUES ($1) 
        RETURNING id, name, created_at, updated_at
    `

	err := r.db.QueryRowContext(
		ctx,
		query,
		name,
	).Scan(
		&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (r *postgresRepository) UpdateCategory(ctx context.Context, id int, name string) (*Category, error) {
	query := `
        UPDATE categories 
        SET name = $1, updated_at = NOW()
        WHERE id = $2
    `

	_, err := r.db.ExecContext(ctx, query, name, id)
	if err != nil {
		return nil, err
	}

	c, err := r.GetCategoryByID(ctx, id)
	if err != nil {
		r.log.Err(err).Str("method", "UpdateCategory").Msg(err.Error())
		return nil, ErrNotExist
	}

	if c.Name != name {
		return nil, errors.New("update failed, name not changed")
	}

	return c, nil
}

func (r *postgresRepository) SearchCategoriesByName(ctx context.Context, searchTerm string) ([]*Category, error) {
	query := `SELECT *
		FROM categories
		WHERE name ILIKE $1
	`
	searchTerm = "%" + searchTerm + "%"

	var c []*Category
	err := r.db.SelectContext(ctx, &c, query, searchTerm)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (r *postgresRepository) CountCategories(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*) as total_categories
		FROM categories
	`

	var count int
	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, err
	}

	return count, nil
}
