package inventory

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rovilay/ecommerce-service/domains/product"
)

type productServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewProductServiceClient(productClientUrl string) *productServiceClient {
	return &productServiceClient{baseURL: productClientUrl, httpClient: &http.Client{}}
}

func (c *productServiceClient) CheckProductExists(ctx context.Context, productID int) (bool, error) {
	url := fmt.Sprintf("%s/products/%d", c.baseURL, productID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	prod := &product.Product{}

	if err := json.NewDecoder(resp.Body).Decode(prod); err != nil {
		return false, err
	}

	return prod.ID == productID, nil
}
