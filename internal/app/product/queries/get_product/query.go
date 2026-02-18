package get_product

import (
	"context"

	"github.com/product-catalog-service/internal/app/product/contracts"
)

// Request represents the input for getting a product.
type Request struct {
	ProductID string
}

// Query handles the get product query.
type Query struct {
	readModel contracts.ProductReadModelRepository
}

// NewQuery creates a new get product query handler.
func NewQuery(readModel contracts.ProductReadModelRepository) *Query {
	return &Query{
		readModel: readModel,
	}
}

// Execute retrieves a product by ID.
func (q *Query) Execute(ctx context.Context, req Request) (*ProductDTO, error) {
	product, err := q.readModel.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	return mapToDTO(product), nil
}

func mapToDTO(rm *contracts.ProductReadModel) *ProductDTO {
	dto := &ProductDTO{
		ID:                   rm.ID,
		Name:                 rm.Name,
		Description:          rm.Description,
		Category:             rm.Category,
		BasePriceNumerator:   rm.BasePriceNumerator,
		BasePriceDenominator: rm.BasePriceDenominator,
		EffectivePriceNum:    rm.EffectivePriceNum,
		EffectivePriceDenom:  rm.EffectivePriceDenom,
		Status:               rm.Status,
		CreatedAt:            rm.CreatedAt,
		UpdatedAt:            rm.UpdatedAt,
	}

	if rm.DiscountPercent != nil {
		dto.DiscountPercent = rm.DiscountPercent
		dto.DiscountStartDate = rm.DiscountStartDate
		dto.DiscountEndDate = rm.DiscountEndDate
	}

	return dto
}
