package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/rovilay/ecommerce-service/domains/inventory"
	"github.com/rovilay/ecommerce-service/domains/inventory/model"
	"github.com/rs/zerolog"
)

type postgresInventoryRepository struct {
	db  *sqlx.DB
	log *zerolog.Logger
}

func NewPostgresInventoryRepository(ctx context.Context, db *sqlx.DB, log *zerolog.Logger) *postgresInventoryRepository {
	repoLogger := log.With().Str("repository", "postgresInventoryRepository").Logger()

	// ping db
	if err := db.PingContext(ctx); err != nil {
		repoLogger.Fatal().Err(fmt.Errorf("failed to connect to postgres: %w", err))
	}

	return &postgresInventoryRepository{
		db:  db,
		log: &repoLogger,
	}
}

func (r *postgresInventoryRepository) CreateInventoryItem(ctx context.Context, productID int, quantity uint) (*model.InventoryItem, error) {
	log := r.log.With().Str("method", "CreateInventoryItem").Logger()

	var ivn model.InventoryItem
	query := `INSERT INTO inventory_items (product_id, quantity)
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

	err = tx.Commit()
	if err != nil {
		return r.mapDatabaseError(err, &log)
	}

	return nil
}

func (r *postgresInventoryRepository) mapDatabaseError(err error, log *zerolog.Logger) error {
	log.Err(err).Msg("database operation failed!")

	var pqErr *pgconn.PgError
	if ok := errors.As(err, &pqErr); ok {
		log.Debug().Msg(fmt.Sprintf("%v:%v", ok, pqErr.SQLState()))

		switch pqErr.Code {
		case "23505": // Unique constraint violation
			return inventory.ErrDuplicateEntry
		case "23503": // Foreign key violation
			return inventory.ErrForeignKeyViolation
		default:
			return fmt.Errorf("database error (%s): %w", pqErr.Code, err)
		}
	} else if errors.Is(err, sql.ErrNoRows) {
		return inventory.ErrNotFound
	} else {
		return fmt.Errorf("database error: %w", err)
	}
}
