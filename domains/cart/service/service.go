package service

import (
	"context"

	"github.com/rovilay/ecommerce-service/domains/auth"
	"github.com/rovilay/ecommerce-service/domains/cart"
	"github.com/rovilay/ecommerce-service/domains/cart/models"
	"github.com/rovilay/ecommerce-service/domains/cart/repository"
	"github.com/rs/zerolog"
)

type CartService struct {
	repo        repository.CartRepository
	authService auth.AuthService
	log         *zerolog.Logger
}

func NewCartService(r repository.CartRepository, s auth.AuthService, l *zerolog.Logger) *CartService {
	logger := l.With().Str("service", "CartService").Logger()

	return &CartService{
		log:         &logger,
		authService: s,
		repo:        r,
	}
}

func (s *CartService) GetCart(ctx context.Context, authToken string) (*models.Cart, error) {
	log := s.log.With().Str("method", "GetCart").Logger()

	userID, err := s.authService.ValidateJWT(ctx, authToken)
	if err != nil {
		log.Err(err).Msg("error validating token")
		return nil, cart.ErrInvalidJWToken
	}

	return s.repo.GetCartByUserID(ctx, userID)
}

func (s *CartService) AddItemToCart(ctx context.Context, authToken string, item models.CartItem) (*models.CartItem, error) {
	log := s.log.With().Str("method", "AddItemToCart").Logger()

	userID, err := s.authService.ValidateJWT(ctx, authToken)
	if err != nil {
		log.Err(err).Msg("error validating token")
		return nil, cart.ErrInvalidJWToken
	}

	return s.repo.AddItemToCart(ctx, userID, item.ProductID, item.Quantity)
}

func (s *CartService) UpdateCartItemQuantity(ctx context.Context, authToken string, item models.CartItem) error {
	log := s.log.With().Str("method", "UpdateCartItemQuantity").Logger()

	userID, err := s.authService.ValidateJWT(ctx, authToken)
	if err != nil {
		log.Err(err).Msg("error validating token")
		return cart.ErrInvalidJWToken
	}

	return s.repo.UpdateCartItemQuantity(ctx, userID, item.ID, item.Quantity)
}

func (s *CartService) RemoveItemFromCart(ctx context.Context, authToken string, cartItemID int) error {
	log := s.log.With().Str("method", "RemoveItemFromCart").Logger()

	userID, err := s.authService.ValidateJWT(ctx, authToken)
	if err != nil {
		log.Err(err).Msg("error validating token")
		return cart.ErrInvalidJWToken
	}

	return s.repo.RemoveItemFromCart(ctx, userID, cartItemID)
}

func (s *CartService) ClearCart(ctx context.Context, authToken string) error {
	log := s.log.With().Str("method", "ClearCart").Logger()

	userID, err := s.authService.ValidateJWT(ctx, authToken)
	if err != nil {
		log.Err(err).Msg("error validating token")
		return cart.ErrInvalidJWToken
	}

	return s.repo.ClearCartByUserID(ctx, userID)
}
