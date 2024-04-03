package inventory

import "errors"

var ErrInsufficientStock = errors.New("insufficient stock")
var ErrNotFound = errors.New("inventory item not found does")
var ErrDuplicateEntry = errors.New("duplicate inventory entry")
var ErrForeignKeyViolation = errors.New("foreign key violation (invalid product reference?)")
