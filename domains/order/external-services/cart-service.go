package externalservices

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rovilay/ecommerce-service/domains/order"
)

type Cart struct {
	UserID    string     `json:"user_id"`
	CartItems []CartItem `json:"cart_items"`
}

type CartItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type CartService interface {
	GetCart(ctx context.Context, authToken string) (*Cart, error)
	ClearCart(ctx context.Context, authToken string) error
}

type HTTPCartService struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPCartService(baseURL string) *HTTPCartService {
	return &HTTPCartService{
		httpClient: &http.Client{},
		baseURL:    baseURL,
	}
}

func (s *HTTPCartService) GetCart(ctx context.Context, authToken string) (*Cart, error) {
	url := fmt.Sprintf("%s/cart", s.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var cart Cart

	if resp.StatusCode != http.StatusOK {
		return nil, order.ErrInvalidCart
	}

	if err = json.NewDecoder(resp.Body).Decode(&cart); err != nil {
		return nil, err
	}

	return &cart, nil
}

func (s *HTTPCartService) ClearCart(ctx context.Context, authToken string) error {
	url := fmt.Sprintf("%s/cart", s.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return order.ErrInvalidCart
	}

	return nil
}
