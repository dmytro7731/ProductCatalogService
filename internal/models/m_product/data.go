package m_product

import (
	"time"

	"cloud.google.com/go/spanner"
)

// Product represents the database model for a product.
type Product struct {
	ProductID            string
	Name                 string
	Description          string
	Category             string
	BasePriceNumerator   int64
	BasePriceDenominator int64
	DiscountPercent      spanner.NullNumeric
	DiscountStartDate    spanner.NullTime
	DiscountEndDate      spanner.NullTime
	Status               string
	CreatedAt            time.Time
	UpdatedAt            time.Time
	ArchivedAt           spanner.NullTime
}

// Model provides methods for creating Spanner mutations.
type Model struct{}

// NewModel creates a new Model instance.
func NewModel() *Model {
	return &Model{}
}

// InsertMut creates an insert mutation for a product.
func (m *Model) InsertMut(p *Product) *spanner.Mutation {
	return spanner.InsertMap(TableName, map[string]interface{}{
		ProductID:            p.ProductID,
		Name:                 p.Name,
		Description:          p.Description,
		Category:             p.Category,
		BasePriceNumerator:   p.BasePriceNumerator,
		BasePriceDenominator: p.BasePriceDenominator,
		DiscountPercent:      p.DiscountPercent,
		DiscountStartDate:    p.DiscountStartDate,
		DiscountEndDate:      p.DiscountEndDate,
		Status:               p.Status,
		CreatedAt:            p.CreatedAt,
		UpdatedAt:            p.UpdatedAt,
		ArchivedAt:           p.ArchivedAt,
	})
}

// UpdateMut creates an update mutation for specific columns.
func (m *Model) UpdateMut(productID string, updates map[string]interface{}) *spanner.Mutation {
	updates[ProductID] = productID
	return spanner.UpdateMap(TableName, updates)
}

// InsertOrUpdateMut creates an insert or update mutation.
func (m *Model) InsertOrUpdateMut(p *Product) *spanner.Mutation {
	return spanner.InsertOrUpdateMap(TableName, map[string]interface{}{
		ProductID:            p.ProductID,
		Name:                 p.Name,
		Description:          p.Description,
		Category:             p.Category,
		BasePriceNumerator:   p.BasePriceNumerator,
		BasePriceDenominator: p.BasePriceDenominator,
		DiscountPercent:      p.DiscountPercent,
		DiscountStartDate:    p.DiscountStartDate,
		DiscountEndDate:      p.DiscountEndDate,
		Status:               p.Status,
		CreatedAt:            p.CreatedAt,
		UpdatedAt:            p.UpdatedAt,
		ArchivedAt:           p.ArchivedAt,
	})
}
