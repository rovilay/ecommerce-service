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
	logger := l.With().Str("component", "productHandler").Logger()

	return &ProductHandler{
		service: s,
		log:     &logger,
	}
}
