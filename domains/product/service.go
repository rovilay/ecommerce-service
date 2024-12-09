package product

import (
	"context"

	"github.com/rovilay/ecommerce-service/common/events"
)

type Service struct {
	repo      Repository
	msgBroker *events.RabbitClient
}

func NewService(repo Repository, b *events.RabbitClient) (*Service, error) {
	s := &Service{repo: repo, msgBroker: b}
	err := s.setupQueues()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Service) setupQueues() error {
	// create queues and bindings here
	productCreatedQ, err := s.msgBroker.CreateQueue(string(events.ProductCreated), true, false)
	if err != nil {
		return err
	}

	err = s.msgBroker.CreateBinding(events.Product, productCreatedQ.Name, events.ProductCreated)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetProduct(ctx context.Context, id int) (*Product, error) {
	return s.repo.GetProductByID(ctx, id)
}

func (s *Service) CreateProduct(ctx context.Context, data *Product) (*Product, error) {
	p, err := s.repo.CreateProduct(ctx, data)
	if err != nil {
		return nil, err
	}

	// publish event
	e := events.EventData{
		Event: events.ProductCreated,
		Data:  p,
	}
	err = s.msgBroker.Send(ctx, string(events.Product), string(events.ProductCreated), e)
	if err != nil {
		return nil, err
	}

	return p, err
}

func (s *Service) ListProducts(ctx context.Context, limit int, offset int) (*PaginationResult[*Product], error) {
	pwp := &PaginationResult[*Product]{Limit: limit, Offset: offset}

	total, err := s.repo.CountProducts(ctx)
	if err != nil {
		return nil, err
	}

	pwp.Total = total

	if offset >= total {
		return pwp, nil
	}

	products, err := s.repo.GetAllProducts(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	pwp.Items = products

	return pwp, nil
}

func (s *Service) UpdateProduct(ctx context.Context, id int, data *Product) (*Product, error) {
	data.ID = id
	return s.repo.UpdateProduct(ctx, data)
}

func (s *Service) DeleteProduct(ctx context.Context, id int) error {
	return s.repo.DeleteProduct(ctx, id)
}

func (s *Service) SearchProductsByName(ctx context.Context, searchTerm string) ([]*Product, error) {
	return s.repo.SearchProductsByName(ctx, searchTerm)
}

func (s *Service) GetCategory(ctx context.Context, id int) (*Category, error) {
	return s.repo.GetCategoryByID(ctx, id)
}

func (s *Service) CreateCategory(ctx context.Context, data *Category) (*Category, error) {
	return s.repo.CreateCategory(ctx, data.Name)
}

func (s *Service) ListCategories(ctx context.Context, limit int, offset int) (*PaginationResult[*Category], error) {
	cwp := &PaginationResult[*Category]{Limit: limit, Offset: offset}

	total, err := s.repo.CountCategories(ctx)
	if err != nil {
		return nil, err
	}

	cwp.Total = total

	if offset >= total {
		return cwp, nil
	}

	categories, err := s.repo.GetAllCategories(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	cwp.Items = categories

	return cwp, nil
}

func (s *Service) UpdateCategory(ctx context.Context, id int, data *Category) (*Category, error) {
	return s.repo.UpdateCategory(ctx, id, data.Name)
}

func (s *Service) SearchCategoriesByName(ctx context.Context, searchTerm string) ([]*Category, error) {
	return s.repo.SearchCategoriesByName(ctx, searchTerm)
}
