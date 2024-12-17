package externalservices

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type InventoryService interface {
	CheckAvailability(ctx context.Context, productID int, quantity int) (bool, error)
	UpdateInventory(ctx context.Context, descrease bool, productID int, quantity int) error
}

type HTTPInventoryService struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPInventoryService(baseURL string) *HTTPInventoryService {

	return &HTTPInventoryService{
		httpClient: &http.Client{},
		baseURL:    baseURL,
	}
}

func (s *HTTPInventoryService) CheckAvailability(ctx context.Context, productID int, quantity int) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/inventory/products/%d/available?qty=%d", s.baseURL, productID, quantity)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var availableRes struct {
		Available bool   `json:"available"`
		Error     string `json:"error"`
	}

	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	if err = json.NewDecoder(resp.Body).Decode(&availableRes); err != nil {
		return false, err
	}

	if !availableRes.Available {
		return false, nil
	}

	if availableRes.Error != "" {
		return false, errors.New(availableRes.Error)
	}

	return true, nil
}

func (s *HTTPInventoryService) UpdateInventory(ctx context.Context, descrease bool, productID int, quantity int) error {
	ops := "increase"
	if descrease {
		ops = "decrease"
	}

	var payload struct {
		Quantity int `json:"quantity"`
	}

	payload.Quantity = quantity

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/inventory/products/%d/%s", s.baseURL, productID, ops)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to %s inventory for product: %d", ops, productID)
	}

	return nil
}
