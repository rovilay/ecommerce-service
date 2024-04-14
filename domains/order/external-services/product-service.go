package externalservices

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rovilay/ecommerce-service/domains/order"
)

type Product struct {
	ID    int     `json:"id"`
	Price float32 `json:"price"`
}

type ProductService interface {
	GetProduct(ctx context.Context, productID int) (*Product, error)
}

type HTTPProductService struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPProductService(baseURL string) *HTTPProductService {
	return &HTTPProductService{
		httpClient: &http.Client{},
		baseURL:    baseURL,
	}
}

func (s *HTTPProductService) GetProduct(ctx context.Context, productID int) (*Product, error) {
	url := fmt.Sprintf("%s/products/%d", s.baseURL, productID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var prd Product

	if resp.StatusCode != http.StatusOK {
		return nil, order.ErrInvalidProduct
	}

	if err = json.NewDecoder(resp.Body).Decode(&prd); err != nil {
		return nil, err
	}

	return &prd, nil
}
