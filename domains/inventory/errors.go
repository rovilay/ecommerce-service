package inventory

import "errors"

var ErrInsufficientStock = errors.New("insufficient stock")
var ErrNotFound = errors.New("inventory item not found for this product")
var ErrDuplicateEntry = errors.New("duplicate inventory entry")
var ErrForeignKeyViolation = errors.New("foreign key violation (invalid product reference?)")
var ErrInvalidQuantity = errors.New("initial quantity cannot be negative")
var ErrInvalidProduct = errors.New("product not found")
