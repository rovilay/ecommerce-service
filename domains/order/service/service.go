package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rovilay/ecommerce-service/domains/auth"
	"github.com/rovilay/ecommerce-service/domains/order"
	externalservices "github.com/rovilay/ecommerce-service/domains/order/external-services"
	"github.com/rovilay/ecommerce-service/domains/order/models"
	"github.com/rovilay/ecommerce-service/domains/order/repository"
	"github.com/rs/zerolog"
)

type OrderService struct {
	repo             repository.OrderRepository
	authService      auth.AuthService
	inventoryService externalservices.InventoryService
	prdService       externalservices.ProductService
	cartService      externalservices.CartService
	log              *zerolog.Logger
}

func NewOrderService(repo repository.OrderRepository, a auth.AuthService, i externalservices.InventoryService,
	p externalservices.ProductService, c externalservices.CartService, l *zerolog.Logger,
) *OrderService {
	logger := l.With().Str("service", "OrderService").Logger()

	return &OrderService{
		repo:             repo,
		authService:      a,
		inventoryService: i,
		prdService:       p,
		cartService:      c,
		log:              &logger,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, authToken string, data *models.Order, fromCart bool) (*models.Order, error) {
	log := s.log.With().Str("method", "GetCart").Logger()

	userID, err := s.authService.ValidateJWT(ctx, authToken)
	if err != nil {
		log.Err(err).Msg("error validating token")
		return nil, order.ErrInvalidJWToken
	}

	data.UserID, err = uuid.Parse(userID)
	if err != nil {
		log.Err(err).Msg("error parsing userID")
		return nil, order.ErrInvalidJWToken
	}

	if fromCart {
		orderItemsFromCart, err := s.getOrderItemsFromCart(ctx, authToken)
		if err != nil {
			log.Err(err).Msg("failed to get order items from cart")
			return nil, err
		}

		data.OrderItems = orderItemsFromCart
	}

	duration := time.Duration(30 * len(data.OrderItems))
	if duration.Minutes() > 5 {
		duration = 30 * 5
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*duration)
	defer cancel()

	validOrderItems, totalPrice, err := s.validateOrderItems(timeoutCtx, data.OrderItems)
	if err != nil {
		log.Err(err).Msg("error validating order items")
		return nil, err
	}

	data.OrderItems = validOrderItems
	data.TotalPrice = totalPrice
	data.Status = models.OrderStatusPending

	order, err := s.repo.CreateOrder(ctx, data)
	if err != nil {
		return nil, err
	}

	_, err = s.updateInventory(ctx, order)
	if err != nil {
		log.Err(err).Msg("inventory update failed")
	}

	if fromCart {
		err = s.cartService.ClearCart(ctx, authToken)
		if err != nil {
			log.Err(err).Msg("cart clearance failed")
		}
	}

	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, authToken string, orderID int) (*models.Order, error) {
	log := s.log.With().Str("method", "GetOrder").Logger()

	userID, err := s.authService.ValidateJWT(ctx, authToken)
	if err != nil {
		log.Err(err).Msg("error validating token")
		return nil, order.ErrInvalidJWToken
	}

	_, err = uuid.Parse(userID)
	if err != nil {
		log.Err(err).Msg("error parsing userID")
		return nil, order.ErrInvalidJWToken
	}

	return s.repo.GetOrderByID(ctx, orderID)
}

func (s *OrderService) GetUserOrders(ctx context.Context, authToken string, limit int, offset int) (*models.PaginationResult[*models.Order], error) {
	log := s.log.With().Str("method", "GetUserOrders").Logger()

	userID, err := s.authService.ValidateJWT(ctx, authToken)
	if err != nil {
		log.Err(err).Msg("error validating token")
		return nil, order.ErrInvalidJWToken
	}

	uUserID, err := uuid.Parse(userID)
	if err != nil {
		log.Err(err).Msg("error parsing userID")
		return nil, order.ErrInvalidJWToken
	}

	totalOrders, err := s.repo.CountUserOrders(ctx, uUserID)
	if err != nil {
		return nil, err
	}

	orders, err := s.repo.GetOrdersByUser(ctx, uUserID, limit, offset)
	if err != nil {
		return nil, err
	}

	var res models.PaginationResult[*models.Order]

	res.Items = orders
	res.Limit = limit
	res.Offset = offset
	res.Total = totalOrders

	return &res, nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, authToken string, orderID int, newStatus models.OrderStatus) error {
	log := s.log.With().Str("method", "GetUserOrders").Logger()

	userID, err := s.authService.ValidateJWT(ctx, authToken)
	if err != nil {
		log.Err(err).Msg("error validating token")
		return order.ErrInvalidJWToken
	}

	_, err = uuid.Parse(userID)
	if err != nil {
		log.Err(err).Msg("error parsing userID")
		return order.ErrInvalidJWToken
	}

	return s.repo.UpdateOrderStatus(ctx, orderID, newStatus)
}

type validationResult struct {
	productID int
	price     float32
	available bool
}

func (s *OrderService) validateOrderItems(ctx context.Context, items []models.OrderItem) ([]models.OrderItem, float32, error) {
	numItems := len(items)
	results := make([]models.OrderItem, numItems)
	var validationErrors []string

	var wg sync.WaitGroup
	wg.Add(numItems)

	for i, item := range items {
		go func(index int, item *models.OrderItem) {
			defer wg.Done()

			res, err := s.validateOrderItem(ctx, item)
			if err != nil {
				validationErrors = append(validationErrors, err.Error())
				return
			}

			if res.available {
				results[index] = *item
				results[index].ProductID = res.productID
				results[index].Price = res.price
			}
		}(i, &item)
	}

	wg.Wait()
	totalPrice := s.calculateTotalPrice(results)

	if len(validationErrors) > 0 {
		return nil, 0, fmt.Errorf("validation errors: %v", validationErrors)
	}

	return results, totalPrice, nil
}

func (s *OrderService) validateOrderItem(ctx context.Context, orderItem *models.OrderItem) (*validationResult, error) {
	vRes := validationResult{}

	numRoutines := 2
	productChan := make(chan *externalservices.Product)
	availabilityChan := make(chan bool)
	errChan := make(chan error)

	go func() {
		prd, err := s.prdService.GetProduct(ctx, orderItem.ProductID)
		if err != nil {
			errChan <- err
			return
		}
		productChan <- prd
	}()

	go func() {
		available, err := s.inventoryService.CheckAvailability(ctx, orderItem.ProductID, orderItem.Quantity)
		if err != nil {
			errChan <- err
			return
		}
		availabilityChan <- available
	}()

	for i := 0; i < numRoutines; i++ {
		select {
		case err := <-errChan:
			return nil, err
		case prd := <-productChan:
			vRes.productID = prd.ID
			vRes.price = prd.Price
		case available := <-availabilityChan:
			vRes.available = available
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return &vRes, nil
}

func (s *OrderService) calculateTotalPrice(items []models.OrderItem) float32 {
	totalPrice := float32(0.0)

	for _, item := range items {
		totalPrice += item.Price
	}

	return totalPrice
}

type updateError struct {
	id       int
	err      error
	quantity int
}

func (s *OrderService) updateInventory(ctx context.Context, order *models.Order) ([]updateError, error) {
	numItems := len(order.OrderItems)
	var errs []updateError

	var wg sync.WaitGroup
	wg.Add(numItems)
	for _, item := range order.OrderItems {
		go func(orderItem *models.OrderItem) {
			defer wg.Done()

			err := s.inventoryService.UpdateInventory(ctx, true, orderItem.ProductID, orderItem.Quantity)
			errs = append(errs, updateError{id: orderItem.ProductID, err: err, quantity: orderItem.Quantity})
		}(&item)
	}

	wg.Wait()

	if len(errs) == 0 {
		return nil, nil
	}

	return errs, fmt.Errorf("validation errors: %v", errs)
}

func (s *OrderService) getOrderItemsFromCart(ctx context.Context, authToken string) ([]models.OrderItem, error) {
	cart, err := s.cartService.GetCart(ctx, authToken)
	if err != nil {
		return nil, err
	}

	var items []models.OrderItem
	for _, cartItem := range cart.CartItems {
		items = append(items, models.OrderItem{ProductID: cartItem.ProductID, Quantity: cartItem.Quantity})
	}

	return items, nil
}
