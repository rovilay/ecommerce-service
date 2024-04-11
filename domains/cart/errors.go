package cart

import "errors"

var ErrInsufficientStock = errors.New("insufficient stock")
var ErrNotFound = errors.New("cart not found")
var ErrItemNotFound = errors.New("cart item not found")
var ErrDuplicateEntry = errors.New("duplicate inventory entry")
var ErrForeignKeyViolation = errors.New("foreign key violation (invalid cart reference?)")
var ErrInvalidQuantity = errors.New("quantity must not be negative or zero")
var ErrInvalidProduct = errors.New("product not found")
var ErrInvalidJWToken = errors.New("unauthorized, invalid token")
