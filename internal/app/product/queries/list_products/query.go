package list_products

import (
	"context"

	"github.com/product-catalog-service/internal/app/product/contracts"
)

// Request represents the input for listing products.
type Request struct {
	Category   *string
	Status     *string
	ActiveOnly bool
	Limit      int
	Offset     int
}

// Query handles the list products query.
type Query struct {
	readModel contracts.ProductReadModelRepository
}

// NewQuery creates a new list products query handler.
func NewQuery(readModel contracts.ProductReadModelRepository) *Query {
	return &Query{
		readModel: readModel,
	}
}

// Execute retrieves a paginated list of products.
func (q *Query) Execute(ctx context.Context, req Request) (*ListResultDTO, error) {
	filters := contracts.ProductListFilters{
		Category:   req.Category,
		Status:     req.Status,
		ActiveOnly: req.ActiveOnly,
	}

	pagination := contracts.Pagination{
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	// Apply defaults
	if pagination.Limit <= 0 {
		pagination.Limit = 20
	}
	if pagination.Limit > 100 {
		pagination.Limit = 100
	}

	result, err := q.readModel.List(ctx, filters, pagination)
	if err != nil {
		return nil, err
	}

	return mapToListResult(result), nil
}

func mapToListResult(result *contracts.ProductListResult) *ListResultDTO {
	products := make([]*ProductListItemDTO, len(result.Products))

	for i, p := range result.Products {
		products[i] = &ProductListItemDTO{
			ID:                   p.ID,
			Name:                 p.Name,
			Description:          p.Description,
			Category:             p.Category,
			BasePriceNumerator:   p.BasePriceNumerator,
			BasePriceDenominator: p.BasePriceDenominator,
			EffectivePriceNum:    p.EffectivePriceNum,
			EffectivePriceDenom:  p.EffectivePriceDenom,
			DiscountPercent:      p.DiscountPercent,
			Status:               p.Status,
			CreatedAt:            p.CreatedAt,
		}
	}

	return &ListResultDTO{
		Products:   products,
		TotalCount: result.TotalCount,
		HasMore:    result.HasMore,
	}
}
