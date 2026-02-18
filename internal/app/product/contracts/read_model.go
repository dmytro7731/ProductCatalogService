package contracts

import (
	"context"
	"time"
)

// ProductReadModel represents a product for read operations.
// This is a DTO optimized for queries, not domain logic.
type ProductReadModel struct {
	ID                   string
	Name                 string
	Description          string
	Category             string
	BasePriceNumerator   int64
	BasePriceDenominator int64
	EffectivePriceNum    int64
	EffectivePriceDenom  int64
	DiscountPercent      *int64
	DiscountStartDate    *time.Time
	DiscountEndDate      *time.Time
	Status               string
	CreatedAt            time.Time
	UpdatedAt            time.Time
	ArchivedAt           *time.Time
}

// ProductListFilters defines filters for listing products.
type ProductListFilters struct {
	Category   *string
	Status     *string
	ActiveOnly bool
}

// Pagination defines pagination parameters.
type Pagination struct {
	Limit  int
	Offset int
}

// ProductListResult contains the result of a product list query.
type ProductListResult struct {
	Products   []*ProductReadModel
	TotalCount int64
	HasMore    bool
}

// ProductReadModelRepository defines the interface for product read operations.
// This interface is for queries (CQRS read side) and may bypass domain for optimization.
type ProductReadModelRepository interface {
	// GetByID retrieves a product read model by ID.
	GetByID(ctx context.Context, id string) (*ProductReadModel, error)

	// List retrieves a paginated list of products with optional filters.
	List(ctx context.Context, filters ProductListFilters, pagination Pagination) (*ProductListResult, error)

	// CountByCategory counts products in a category.
	CountByCategory(ctx context.Context, category string) (int64, error)
}
