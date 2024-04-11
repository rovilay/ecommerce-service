package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/rovilay/ecommerce-service/domains/cart"
	"github.com/rovilay/ecommerce-service/domains/cart/models"
	"github.com/rs/zerolog"
)

type postgresCartRepository struct {
	db  *sqlx.DB
	log *zerolog.Logger
}

func NewPostgresCartRepository(ctx context.Context, db *sqlx.DB, log *zerolog.Logger) *postgresCartRepository {
	repoLogger := log.With().Str("repository", "postgresCartRepository").Logger()

	// ping db
	if err := db.PingContext(ctx); err != nil {
		repoLogger.Fatal().Err(fmt.Errorf("failed to connect to postgres: %w", err))
	}

	return &postgresCartRepository{
		db:  db,
		log: &repoLogger,
	}
}

func (r *postgresCartRepository) GetCartByUserID(ctx context.Context, userID string) (*models.Cart, error) {
	log := r.log.With().Str("method", "GetCartByUserID").Logger()
	query := `
        SELECT c.id, c.user_id, ci.id, ci.product_id, ci.quantity
        FROM carts c
        JOIN cart_items ci ON c.id = ci.cart_id
        WHERE c.user_id = $1
    `

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, r.mapDatabaseError(err, &log)
	}
	defer rows.Close()

	cart := &models.Cart{UserID: userID}
	for rows.Next() {
		var item models.CartItem
		if err := rows.Scan(&cart.ID, &cart.UserID, &item.ID, &item.ProductID, &item.Quantity); err != nil {
			return nil, r.mapDatabaseError(err, &log)
		}
		cart.CartItems = append(cart.CartItems, item)
	}

	if cart.ID == 0 { // Check if a cart was actually found
		return nil, r.mapDatabaseError(sql.ErrNoRows, &log)
	}

	return cart, nil
}

func (r *postgresCartRepository) AddItemToCart(ctx context.Context, userID string, productID int, quantity int) (*models.CartItem, error) {
	log := r.log.With().Str("method", "AddItemToCart").Logger()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, r.mapDatabaseError(sql.ErrNoRows, &log)
	}
	defer tx.Rollback()

	var item models.CartItem

	// 1. Get or Create Cart
	var cartID int
	query1 := `
	INSERT INTO carts (user_id) 
	VALUES ($1) 
	ON CONFLICT (user_id) DO UPDATE
		SET updated_at = now()
	RETURNING id`

	err = tx.QueryRowContext(ctx, query1, userID).Scan(&cartID)
	if err != nil {
		return nil, r.mapDatabaseError(sql.ErrNoRows, &log)
	}

	// 2. Check for Existing Item
	var existingQuantity int
	query2 := `
	SELECT quantity FROM cart_items 
	WHERE cart_id = $1 AND product_id = $2`

	err = tx.QueryRowContext(ctx, query2, cartID, productID).Scan(&existingQuantity)

	// 3. Insert or Update Based on Existence
	if err != nil && err != sql.ErrNoRows {
		return nil, r.mapDatabaseError(sql.ErrNoRows, &log)
	} else if err == sql.ErrNoRows {
		query3 := `
		INSERT INTO cart_items (cart_id, product_id, quantity)
		VALUES ($1, $2, $3)`
		err = tx.QueryRowContext(ctx, query3, cartID, productID, quantity).Scan(&item)
	} else {
		query4 := `
		UPDATE cart_items SET quantity = quantity + $1 
		WHERE cart_id = $2 AND product_id = $3`
		err = tx.QueryRowContext(ctx, query4, quantity, cartID, productID).Scan(&item)
	}
	if err != nil {
		return nil, r.mapDatabaseError(sql.ErrNoRows, &log)
	}

	err = tx.Commit()
	if err != nil {
		return nil, r.mapDatabaseError(err, &log)
	}

	return &item, nil
}

func (r *postgresCartRepository) UpdateCartItemQuantity(ctx context.Context, userID string, cartItemID int, newQuantity int) error {
	log := r.log.With().Str("method", "UpdateCartItemQuantity").Logger()

	query := `
        UPDATE cart_items 
        SET quantity = $1, updated_at = now()
        WHERE id = $2 AND cart_id IN (SELECT id FROM carts WHERE user_id = $3)
    `
	result, err := r.db.ExecContext(ctx, query, newQuantity, cartItemID, userID)
	if err != nil {
		return r.mapDatabaseError(err, &log)
	}

	// Check if any rows were actually affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return r.mapDatabaseError(err, &log)
	}
	if rowsAffected == 0 {
		return cart.ErrItemNotFound
	}

	return nil
}

func (r *postgresCartRepository) RemoveItemFromCart(ctx context.Context, userID string, cartItemID int) error {
	log := r.log.With().Str("method", "RemoveItemFromCart").Logger()

	query := `
		DELETE FROM cart_items
		WHERE id = $1 AND cart_id IN (SELECT id FROM carts WHERE user_id = $2)
	`
	result, err := r.db.ExecContext(ctx, query, cartItemID, userID)
	if err != nil {
		return r.mapDatabaseError(err, &log)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return r.mapDatabaseError(err, &log)
	}
	if rowsAffected == 0 {
		return cart.ErrItemNotFound
	}

	return nil
}

func (r *postgresCartRepository) ClearCartByUserID(ctx context.Context, userID string) error {
	log := r.log.With().Str("method", "ClearCartByUserID").Logger()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return r.mapDatabaseError(err, &log)
	}
	defer tx.Rollback()

	query := `DELETE FROM carts WHERE user_id = $1`
	_, err = tx.ExecContext(ctx, query, userID)
	if err != nil {
		return r.mapDatabaseError(err, &log)
	}

	err = tx.Commit()
	if err != nil {
		return r.mapDatabaseError(err, &log)
	}

	return nil
}

func (r *postgresCartRepository) mapDatabaseError(err error, log *zerolog.Logger) error {
	log.Err(err).Msg("database operation failed!")

	var pqErr *pgconn.PgError
	if ok := errors.As(err, &pqErr); ok {
		log.Debug().Msg(fmt.Sprintf("%v:%v", ok, pqErr.SQLState()))

		switch pqErr.Code {
		case "23505": // Unique constraint violation
			return cart.ErrDuplicateEntry
		case "23503": // Foreign key violation
			return cart.ErrForeignKeyViolation
		default:
			return fmt.Errorf("database error (%s): %w", pqErr.Code, err)
		}
	} else if errors.Is(err, sql.ErrNoRows) {
		return cart.ErrNotFound
	} else {
		return fmt.Errorf("database error: %w", err)
	}
}
