package handler

import (
	"github.com/rovilay/ecommerce-service/domains/product"
	"github.com/rs/zerolog"
)

type ProductHandler struct {
	service *product.Service
	log     *zerolog.Logger
}

func NewProductHandler(s *product.Service, l *zerolog.Logger) *ProductHandler {
	logger := l.With().Str("component", "ProductHandler").Logger()

	return &ProductHandler{
		service: s,
		log:     &logger,
	}
}

type CategoryHandler struct {
	service *product.Service
	log     *zerolog.Logger
}

func NewCategoryHandler(s *product.Service, l *zerolog.Logger) *CategoryHandler {
	logger := l.With().Str("component", "CategoryHandler").Logger()

	return &CategoryHandler{
		service: s,
		log:     &logger,
	}
}
