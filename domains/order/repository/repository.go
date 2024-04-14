package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/rovilay/ecommerce-service/domains/order/models"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *models.Order) (*models.Order, error)
	GetOrderByID(ctx context.Context, orderID int) (*models.Order, error)
	GetOrdersByUser(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Order, error)
	CountUserOrders(ctx context.Context, userID uuid.UUID) (int, error)
	UpdateOrderStatus(ctx context.Context, orderID int, newStatus models.OrderStatus) error
}
