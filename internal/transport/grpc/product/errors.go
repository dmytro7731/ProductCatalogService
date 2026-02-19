package product

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/product-catalog-service/internal/app/product/domain"
)

// mapDomainErrorToGRPC converts domain errors to gRPC status errors.
func mapDomainErrorToGRPC(err error) error {
	if err == nil {
		return nil
	}

	// Not found errors
	if errors.Is(err, domain.ErrProductNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}

	// Validation errors (invalid argument)
	validationErrors := []error{
		domain.ErrEmptyProductName,
		domain.ErrEmptyCategory,
		domain.ErrInvalidProductStatus,
		domain.ErrProductNameTooLong,
		domain.ErrCategoryTooLong,
		domain.ErrInvalidMoney,
		domain.ErrNegativeMoney,
		domain.ErrZeroPrice,
		domain.ErrInvalidDiscountPercentage,
		domain.ErrInvalidDiscountPeriod,
	}

	for _, validationErr := range validationErrors {
		if errors.Is(err, validationErr) {
			return status.Error(codes.InvalidArgument, err.Error())
		}
	}

	// Business rule violations (failed precondition)
	businessErrors := []error{
		domain.ErrProductNotActive,
		domain.ErrProductAlreadyActive,
		domain.ErrProductArchived,
		domain.ErrProductInactive,
		domain.ErrDiscountNotActive,
		domain.ErrDiscountAlreadyExists,
		domain.ErrNoDiscountToRemove,
		domain.ErrDiscountExpired,
		domain.ErrDiscountNotStarted,
		domain.ErrCannotActivateArchived,
		domain.ErrCannotDeactivateArchived,
		domain.ErrCannotArchiveActive,
		domain.ErrCannotUpdateArchived,
	}

	for _, businessErr := range businessErrors {
		if errors.Is(err, businessErr) {
			return status.Error(codes.FailedPrecondition, err.Error())
		}
	}

	// Default to internal error
	return status.Error(codes.Internal, "internal server error")
}
