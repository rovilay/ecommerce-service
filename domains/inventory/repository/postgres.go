package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/rovilay/ecommerce-service/domains/inventory"
	"github.com/rovilay/ecommerce-service/domains/inventory/model"
	"github.com/rs/zerolog"
)

type postgresInventoryRepository struct {
	db  *sqlx.DB
	log *zerolog.Logger
}

func NewPostgresInventoryRepository(db *sqlx.DB, log *zerolog.Logger) *postgresInventoryRepository {
	repoLogger := log.With().Str("repository", "postgresInventoryRepository").Logger()

	return &postgresInventoryRepository{
		db:  db,
		log: &repoLogger,
	}
}

func (r *postgresInventoryRepository) CreateInventoryItem(ctx context.Context, productID int, quantity uint) (*model.InventoryItem, error) {
	log := r.log.With().Str("method", "CreateInventoryItem").Logger()

	var ivn model.InventoryItem
	query := `INSERT INTO inventory_items (id, product_id, quantity)
		values ($1, $2)
		RETURNING id, product_id, quantity
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		productID,
		quantity,
	).Scan(
		&ivn.ID, &ivn.ProductID, &ivn.Quantity,
	)
	if err != nil {
		return nil, r.mapDatabaseError(err, &log)
	}

	return &ivn, nil
}

func (r *postgresInventoryRepository) GetInventoryItemByProductID(ctx context.Context, productID int) (*model.InventoryItem, error) {
	log := r.log.With().Str("method", "GetInventoryItemByProductID").Logger()

	var inventoryItem model.InventoryItem
	query := `SELECT id, product_id, quantity FROM inventory_items WHERE product_id = $1`
	err := r.db.GetContext(ctx, &inventoryItem, query, productID)
	if err != nil {
		return nil, r.mapDatabaseError(err, &log)
	}
	return &inventoryItem, nil
}

func (r *postgresInventoryRepository) UpdateInventoryQuantity(ctx context.Context, productID int, quantityDelta int) error {
	log := r.log.With().Str("method", "UpdateInventoryQuantity").Logger()

	tx, err := r.db.Begin()
	if err != nil {
		return r.mapDatabaseError(err, &log)
	}
	defer tx.Rollback()

	query := `UPDATE inventory_items SET quantity = quantity + $1 WHERE product_id = $2 AND quantity + $1 >= 0`
	result, err := tx.ExecContext(ctx, query, quantityDelta, productID)
	if err != nil {
		return r.mapDatabaseError(err, &log)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Implies attempted overselling
		if quantityDelta < 0 {
			return inventory.ErrInsufficientStock
		}

		return inventory.ErrNotFound
	}

	return tx.Commit()
}

func (r *postgresInventoryRepository) mapDatabaseError(err error, log *zerolog.Logger) error {
	log.Err(err)

	if errors.Is(err, sql.ErrNoRows) {
		return inventory.ErrNotFound
	} else if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code {
		case "23505": // Unique constraint violation
			return inventory.ErrDuplicateEntry
		case "23503": // Foreign key violation
			return inventory.ErrForeignKeyViolation
		default:
			return fmt.Errorf("database error (%s): %w", pqErr.Code, err)
		}
	} else {
		return fmt.Errorf("database error: %w", err)
	}
}
