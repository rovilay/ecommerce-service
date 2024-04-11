package repository

import (
	"context"

	"github.com/rovilay/ecommerce-service/domains/cart/models"
)

type CartRepository interface {
	GetCartByUserID(ctx context.Context, userID string) (*models.Cart, error)
	AddItemToCart(ctx context.Context, userID string, productID int, quantity int) (*models.CartItem, error)
	UpdateCartItemQuantity(ctx context.Context, userID string, cartItemID int, newQuantity int) error
	RemoveItemFromCart(ctx context.Context, userID string, cartItemID int) error
	ClearCartByUserID(ctx context.Context, userID string) error
}
