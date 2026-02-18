package domain

import "errors"

// Domain errors are sentinel errors that represent business rule violations.
// These errors are pure domain concepts and do not depend on any infrastructure.

var (
	// Product errors
	ErrProductNotFound      = errors.New("product not found")
	ErrProductNotActive     = errors.New("product is not active")
	ErrProductAlreadyActive = errors.New("product is already active")
	ErrProductArchived      = errors.New("product is archived")
	ErrProductInactive      = errors.New("product is inactive")

	// Validation errors
	ErrEmptyProductName     = errors.New("product name cannot be empty")
	ErrEmptyCategory        = errors.New("category cannot be empty")
	ErrInvalidProductStatus = errors.New("invalid product status")
	ErrProductNameTooLong   = errors.New("product name exceeds maximum length")
	ErrCategoryTooLong      = errors.New("category exceeds maximum length")

	// Money errors
	ErrInvalidMoney  = errors.New("invalid money value")
	ErrNegativeMoney = errors.New("money cannot be negative")
	ErrZeroPrice     = errors.New("price cannot be zero")

	// Discount errors
	ErrInvalidDiscountPercentage = errors.New("discount percentage must be between 1 and 100")
	ErrInvalidDiscountPeriod     = errors.New("discount end date must be after start date")
	ErrDiscountNotActive         = errors.New("discount is not active at current time")
	ErrDiscountAlreadyExists     = errors.New("product already has an active discount")
	ErrNoDiscountToRemove        = errors.New("product has no discount to remove")
	ErrDiscountExpired           = errors.New("discount period has expired")
	ErrDiscountNotStarted        = errors.New("discount period has not started yet")

	// State transition errors
	ErrCannotActivateArchived   = errors.New("cannot activate archived product")
	ErrCannotDeactivateArchived = errors.New("cannot deactivate archived product")
	ErrCannotArchiveActive      = errors.New("must deactivate product before archiving")
	ErrCannotUpdateArchived     = errors.New("cannot update archived product")
)

// MaxProductNameLength is the maximum allowed length for product names.
const MaxProductNameLength = 255

// MaxCategoryLength is the maximum allowed length for categories.
const MaxCategoryLength = 100
