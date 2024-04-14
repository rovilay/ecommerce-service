package order

import "errors"

var ErrInsufficientStock = errors.New("insufficient stock")
var ErrNotFound = errors.New("order not found")
var ErrItemNotFound = errors.New("order item not found")
var ErrDuplicateEntry = errors.New("duplicate resource entry")
var ErrForeignKeyViolation = errors.New("foreign key violation (invalid order reference?)")
var ErrInvalidQuantity = errors.New("quantity must not be negative or zero")
var ErrInvalidProduct = errors.New("product not found")
var ErrInvalidCart = errors.New("cart not found")
var ErrInvalidJWToken = errors.New("unauthorized, invalid token")
