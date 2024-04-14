package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/rovilay/ecommerce-service/domains/order"
	"github.com/rovilay/ecommerce-service/domains/order/models"
	"github.com/rs/zerolog"
)

type postgresOrderRepository struct {
	db  *sqlx.DB
	log *zerolog.Logger
}

func NewPostgresOrderRepository(ctx context.Context, db *sqlx.DB, log *zerolog.Logger) *postgresOrderRepository {
	repoLogger := log.With().Str("repository", "postgresOrderRepository").Logger()

	// ping db
	if err := db.PingContext(ctx); err != nil {
		repoLogger.Fatal().Err(fmt.Errorf("failed to connect to postgres: %w", err))
	}

	return &postgresOrderRepository{
		db:  db,
		log: &repoLogger,
	}
}

func (r *postgresOrderRepository) CreateOrder(ctx context.Context, order *models.Order) (*models.Order, error) {
	log := r.log.With().Str("method", "CreateOrder").Logger()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, r.mapDatabaseError(err, &log)
	}
	defer tx.Rollback()

	var orderID struct {
		ID int
	}
	// 1. Insert Order
	query1 := `
        INSERT INTO orders (user_id, status, total_price, shipping_address)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `
	err = tx.QueryRowContext(ctx, query1, order.UserID, string(order.Status), order.TotalPrice, order.ShippingAddress).Scan(&orderID.ID)
	if err != nil {
		return nil, r.mapDatabaseError(err, &log)
	}

	// 2. Insert Order Items
	query2 := `
        INSERT INTO order_items (order_id, product_id, quantity, price)
        VALUES ($1, $2, $3, $4)
		RETURNING id
    `
	for i, item := range order.OrderItems {
		err = tx.QueryRowContext(ctx, query2, orderID.ID, item.ProductID, item.Quantity, item.Price).
			Scan(&item.ID)
		if err != nil {
			return nil, r.mapDatabaseError(err, &log)
		}

		item.OrderID = orderID.ID
		order.OrderItems[i] = item
	}

	// If everything succeeds:
	err = tx.Commit()
	if err != nil {
		return nil, r.mapDatabaseError(err, &log)
	}

	order.ID = orderID.ID

	return order, nil
}

func (r *postgresOrderRepository) GetOrderByID(ctx context.Context, orderID int) (*models.Order, error) {
	log := r.log.With().Str("method", "GetOrderByID").Logger()

	query := `
        SELECT o.id, o.user_id, o.status, o.total_price, o.shipping_address, o.created_at, o.updated_at, 
               coalesce(json_agg(oi) FILTER (WHERE oi.id IS NOT NULL), '[]') AS order_items 
        FROM orders o
        LEFT JOIN order_items oi ON o.id = oi.order_id
        WHERE o.id = $1
        GROUP BY o.id 
    `

	var order models.Order
	var orderItemsJSON string // To store aggregated JSON

	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&order.ID, &order.UserID, &order.Status, &order.TotalPrice, &order.ShippingAddress, &order.CreatedAt, &order.UpdatedAt, &orderItemsJSON,
	)
	if err != nil {
		return nil, r.mapDatabaseError(err, &log)
	}

	// Unmarshal order items
	err = json.Unmarshal([]byte(orderItemsJSON), &order.OrderItems)
	if err != nil {
		return nil, r.mapDatabaseError(err, &log)
	}

	return &order, nil
}

func (r *postgresOrderRepository) GetOrdersByUser(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Order, error) {
	log := r.log.With().Str("method", "GetOrderByUser").Logger()

	query := `
		SELECT o.id, o.user_id, o.status, o.total_price, o.shipping_address, o.created_at, o.updated_at
		FROM orders o
		WHERE o.user_id = $1
		ORDER BY o.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID.String(), limit, offset)
	if err != nil {
		return nil, r.mapDatabaseError(err, &log)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.ID, &order.UserID, &order.Status, &order.TotalPrice,
			&order.ShippingAddress, &order.CreatedAt, &order.UpdatedAt,
		); err != nil {
			return nil, r.mapDatabaseError(err, &log)
		}

		orders = append(orders, &order)
	}

	return orders, nil
}

func (r *postgresOrderRepository) CountUserOrders(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT count(*) as total_orders
		FROM orders
        WHERE user_id = $1;
	`

	var count int
	err := r.db.GetContext(ctx, &count, query, userID.String())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *postgresOrderRepository) UpdateOrderStatus(ctx context.Context, orderID int, newStatus models.OrderStatus) error {
	log := r.log.With().Str("method", "UpdateOrderStatus").Logger()

	query := `UPDATE orders SET status = $1 WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, string(newStatus), orderID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return r.mapDatabaseError(err, &log)
	}

	if rowsAffected == 0 {
		return order.ErrNotFound
	}

	return nil
}

func (r *postgresOrderRepository) mapDatabaseError(err error, log *zerolog.Logger) error {
	log.Err(err).Msg("database operation failed!")

	var pqErr *pgconn.PgError
	if ok := errors.As(err, &pqErr); ok {
		log.Debug().Msg(fmt.Sprintf("%v:%v", ok, pqErr.SQLState()))

		switch pqErr.Code {
		case "23505": // Unique constraint violation
			return order.ErrDuplicateEntry
		case "23503": // Foreign key violation
			return order.ErrForeignKeyViolation
		default:
			return fmt.Errorf("database error (%s): %w", pqErr.Code, err)
		}
	} else if errors.Is(err, sql.ErrNoRows) {
		return order.ErrNotFound
	} else {
		return fmt.Errorf("database error: %w", err)
	}
}
