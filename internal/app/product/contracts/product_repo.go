package contracts

import (
	"context"

	"cloud.google.com/go/spanner"

	"github.com/product-catalog-service/internal/app/product/domain"
)

// ProductRepository defines the interface for product persistence operations.
// Following the pattern: repositories return mutations, NEVER apply them.
type ProductRepository interface {
	// GetByID retrieves a product by its ID.
	GetByID(ctx context.Context, id string) (*domain.Product, error)

	// InsertMut returns a mutation for inserting a new product.
	// Returns nil if the product is not new.
	InsertMut(product *domain.Product) *spanner.Mutation

	// UpdateMut returns a mutation for updating an existing product.
	// Only includes fields that have been modified (using change tracker).
	// Returns nil if there are no changes.
	UpdateMut(product *domain.Product) *spanner.Mutation
}
