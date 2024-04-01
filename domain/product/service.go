package product

import (
	"context"
)

type Service struct {
	repo Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: *repo}
}

func (s *Service) GetProduct(ctx context.Context, id int) (*Product, error) {
	return s.repo.GetProductByID(ctx, id)
}

func (s *Service) CreateProduct(ctx context.Context, data *Product) (*Product, error) {
	return s.repo.CreateProduct(ctx, data)
}

func (s *Service) ListProducts(ctx context.Context, limit int, offset int) (*ProductsWithPagination, error) {
	pwp := &ProductsWithPagination{limit: limit, offset: offset}

	total, err := s.repo.CountProducts(ctx)
	if err != nil {
		return nil, err
	}

	pwp.total = total

	if offset >= total {
		return pwp, nil
	}

	products, err := s.repo.GetAllProducts(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	pwp.products = products

	return pwp, nil

}

func (s *Service) UpateProduct(ctx context.Context, id int, data *Product) (*Product, error) {
	data.ID = id
	return s.repo.UpdateProduct(ctx, data)
}

func (s *Service) DeleteProduct(ctx context.Context, id int) error {
	return s.repo.DeleteProduct(ctx, id)
}

func (s *Service) SearchProductsByName(ctx context.Context, searchTerm string) ([]*Product, error) {
	return s.repo.SearchProductsByName(ctx, searchTerm)
}
