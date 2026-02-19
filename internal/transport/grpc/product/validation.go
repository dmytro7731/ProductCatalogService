package product

import (
	"errors"

	pb "github.com/product-catalog-service/proto/product/v1"
)

var (
	ErrMissingProductID   = errors.New("product_id is required")
	ErrMissingName        = errors.New("name is required")
	ErrMissingCategory    = errors.New("category is required")
	ErrMissingBasePrice   = errors.New("base_price is required")
	ErrInvalidPercentage  = errors.New("percentage must be between 1 and 100")
	ErrMissingStartDate   = errors.New("start_date is required")
	ErrMissingEndDate     = errors.New("end_date is required")
	ErrInvalidDenominator = errors.New("base_price denominator must be positive")
	ErrInvalidNumerator   = errors.New("base_price numerator must be positive")
)

// validateCreateRequest validates CreateProductRequest.
func validateCreateRequest(req *pb.CreateProductRequest) error {
	if req.GetName() == "" {
		return ErrMissingName
	}
	if req.GetCategory() == "" {
		return ErrMissingCategory
	}
	if req.GetBasePrice() == nil {
		return ErrMissingBasePrice
	}
	if req.GetBasePrice().GetDenominator() <= 0 {
		return ErrInvalidDenominator
	}
	if req.GetBasePrice().GetNumerator() <= 0 {
		return ErrInvalidNumerator
	}
	return nil
}

// validateUpdateRequest validates UpdateProductRequest.
func validateUpdateRequest(req *pb.UpdateProductRequest) error {
	if req.GetProductId() == "" {
		return ErrMissingProductID
	}
	if req.GetName() == "" {
		return ErrMissingName
	}
	if req.GetCategory() == "" {
		return ErrMissingCategory
	}
	return nil
}

// validateActivateRequest validates ActivateProductRequest.
func validateActivateRequest(req *pb.ActivateProductRequest) error {
	if req.GetProductId() == "" {
		return ErrMissingProductID
	}
	return nil
}

// validateDeactivateRequest validates DeactivateProductRequest.
func validateDeactivateRequest(req *pb.DeactivateProductRequest) error {
	if req.GetProductId() == "" {
		return ErrMissingProductID
	}
	return nil
}

// validateArchiveRequest validates ArchiveProductRequest.
func validateArchiveRequest(req *pb.ArchiveProductRequest) error {
	if req.GetProductId() == "" {
		return ErrMissingProductID
	}
	return nil
}

// validateApplyDiscountRequest validates ApplyDiscountRequest.
func validateApplyDiscountRequest(req *pb.ApplyDiscountRequest) error {
	if req.GetProductId() == "" {
		return ErrMissingProductID
	}
	if req.GetPercentage() < 1 || req.GetPercentage() > 100 {
		return ErrInvalidPercentage
	}
	if req.GetStartDate() == nil {
		return ErrMissingStartDate
	}
	if req.GetEndDate() == nil {
		return ErrMissingEndDate
	}
	return nil
}

// validateRemoveDiscountRequest validates RemoveDiscountRequest.
func validateRemoveDiscountRequest(req *pb.RemoveDiscountRequest) error {
	if req.GetProductId() == "" {
		return ErrMissingProductID
	}
	return nil
}

// validateGetProductRequest validates GetProductRequest.
func validateGetProductRequest(req *pb.GetProductRequest) error {
	if req.GetProductId() == "" {
		return ErrMissingProductID
	}
	return nil
}

// validateListProductsRequest validates ListProductsRequest.
func validateListProductsRequest(req *pb.ListProductsRequest) error {
	// Limit and offset have sensible defaults, so no validation needed
	return nil
}
