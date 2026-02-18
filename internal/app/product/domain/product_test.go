package domain_test

import (
	"testing"
	"time"

	"github.com/product-catalog-service/internal/app/product/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProduct(t *testing.T) {
	now := time.Now()
	basePrice, _ := domain.NewMoney(1999, 100)

	tests := []struct {
		name        string
		productName string
		description string
		category    string
		basePrice   *domain.Money
		wantErr     error
	}{
		{
			name:        "valid product",
			productName: "Test Product",
			description: "A test product description",
			category:    "Electronics",
			basePrice:   basePrice,
			wantErr:     nil,
		},
		{
			name:        "empty name",
			productName: "",
			description: "Description",
			category:    "Electronics",
			basePrice:   basePrice,
			wantErr:     domain.ErrEmptyProductName,
		},
		{
			name:        "empty category",
			productName: "Product",
			description: "Description",
			category:    "",
			basePrice:   basePrice,
			wantErr:     domain.ErrEmptyCategory,
		},
		{
			name:        "nil base price",
			productName: "Product",
			description: "Description",
			category:    "Electronics",
			basePrice:   nil,
			wantErr:     domain.ErrZeroPrice,
		},
		{
			name:        "zero base price",
			productName: "Product",
			description: "Description",
			category:    "Electronics",
			basePrice:   domain.Zero(),
			wantErr:     domain.ErrZeroPrice,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, err := domain.NewProduct(
				"test-id",
				tt.productName,
				tt.description,
				tt.category,
				tt.basePrice,
				now,
			)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, product)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, product)
				assert.Equal(t, "test-id", product.ID())
				assert.Equal(t, tt.productName, product.Name())
				assert.Equal(t, tt.category, product.Category())
				assert.Equal(t, domain.ProductStatusDraft, product.Status())
				assert.True(t, product.IsNew())

				// Should have ProductCreatedEvent
				events := product.DomainEvents()
				require.Len(t, events, 1)
				assert.Equal(t, "product.created", events[0].EventType())
			}
		})
	}
}

func TestProduct_Update(t *testing.T) {
	now := time.Now()
	basePrice, _ := domain.NewMoney(1999, 100)
	product, err := domain.NewProduct("test-id", "Original Name", "Original Description", "Category1", basePrice, now)
	require.NoError(t, err)
	product.ClearEvents()

	err = product.Update("New Name", "New Description", "Category2", now.Add(time.Hour))
	require.NoError(t, err)

	assert.Equal(t, "New Name", product.Name())
	assert.Equal(t, "New Description", product.Description())
	assert.Equal(t, "Category2", product.Category())

	// Check change tracker
	assert.True(t, product.Changes().Dirty(domain.FieldName))
	assert.True(t, product.Changes().Dirty(domain.FieldDescription))
	assert.True(t, product.Changes().Dirty(domain.FieldCategory))

	// Check event
	events := product.DomainEvents()
	require.Len(t, events, 1)
	assert.Equal(t, "product.updated", events[0].EventType())
}

func TestProduct_UpdateArchived(t *testing.T) {
	product := createArchivedProduct(t)

	err := product.Update("New Name", "New Description", "Category", time.Now())
	assert.ErrorIs(t, err, domain.ErrCannotUpdateArchived)
}

func TestProduct_Activate(t *testing.T) {
	now := time.Now()
	basePrice, _ := domain.NewMoney(1999, 100)
	product, err := domain.NewProduct("test-id", "Product", "Description", "Category", basePrice, now)
	require.NoError(t, err)
	product.ClearEvents()

	err = product.Activate(now)
	require.NoError(t, err)

	assert.Equal(t, domain.ProductStatusActive, product.Status())
	assert.True(t, product.IsActive())
	assert.True(t, product.Changes().Dirty(domain.FieldStatus))

	events := product.DomainEvents()
	require.Len(t, events, 1)
	assert.Equal(t, "product.activated", events[0].EventType())
}

func TestProduct_ActivateAlreadyActive(t *testing.T) {
	product := createActiveProduct(t)

	err := product.Activate(time.Now())
	assert.ErrorIs(t, err, domain.ErrProductAlreadyActive)
}

func TestProduct_ActivateArchived(t *testing.T) {
	product := createArchivedProduct(t)

	err := product.Activate(time.Now())
	assert.ErrorIs(t, err, domain.ErrCannotActivateArchived)
}

func TestProduct_Deactivate(t *testing.T) {
	product := createActiveProduct(t)
	product.ClearEvents()

	err := product.Deactivate(time.Now())
	require.NoError(t, err)

	assert.Equal(t, domain.ProductStatusInactive, product.Status())
	assert.False(t, product.IsActive())

	events := product.DomainEvents()
	require.Len(t, events, 1)
	assert.Equal(t, "product.deactivated", events[0].EventType())
}

func TestProduct_DeactivateArchived(t *testing.T) {
	product := createArchivedProduct(t)

	err := product.Deactivate(time.Now())
	assert.ErrorIs(t, err, domain.ErrCannotDeactivateArchived)
}

func TestProduct_Archive(t *testing.T) {
	now := time.Now()
	basePrice, _ := domain.NewMoney(1999, 100)
	product, err := domain.NewProduct("test-id", "Product", "Description", "Category", basePrice, now)
	require.NoError(t, err)
	product.ClearEvents()

	err = product.Archive(now)
	require.NoError(t, err)

	assert.Equal(t, domain.ProductStatusArchived, product.Status())
	assert.True(t, product.IsArchived())
	assert.NotNil(t, product.ArchivedAt())

	events := product.DomainEvents()
	require.Len(t, events, 1)
	assert.Equal(t, "product.archived", events[0].EventType())
}

func TestProduct_ArchiveActive(t *testing.T) {
	product := createActiveProduct(t)

	err := product.Archive(time.Now())
	assert.ErrorIs(t, err, domain.ErrCannotArchiveActive)
}

func TestProduct_ApplyDiscount(t *testing.T) {
	product := createActiveProduct(t)
	product.ClearEvents()

	now := time.Now()
	discount, err := domain.NewDiscount(20, now, now.Add(7*24*time.Hour))
	require.NoError(t, err)

	err = product.ApplyDiscount(discount, now)
	require.NoError(t, err)

	assert.NotNil(t, product.Discount())
	assert.True(t, product.HasActiveDiscount(now))
	assert.True(t, product.Changes().Dirty(domain.FieldDiscount))

	events := product.DomainEvents()
	require.Len(t, events, 1)
	assert.Equal(t, "product.discount_applied", events[0].EventType())
}

func TestProduct_ApplyDiscountToInactive(t *testing.T) {
	now := time.Now()
	basePrice, _ := domain.NewMoney(1999, 100)
	product, _ := domain.NewProduct("test-id", "Product", "Description", "Category", basePrice, now)

	discount, _ := domain.NewDiscount(20, now, now.Add(7*24*time.Hour))

	err := product.ApplyDiscount(discount, now)
	assert.ErrorIs(t, err, domain.ErrProductNotActive)
}

func TestProduct_RemoveDiscount(t *testing.T) {
	product := createProductWithDiscount(t)
	product.ClearEvents()

	err := product.RemoveDiscount(time.Now())
	require.NoError(t, err)

	assert.Nil(t, product.Discount())
	assert.True(t, product.Changes().Dirty(domain.FieldDiscount))

	events := product.DomainEvents()
	require.Len(t, events, 1)
	assert.Equal(t, "product.discount_removed", events[0].EventType())
}

func TestProduct_RemoveNonExistentDiscount(t *testing.T) {
	product := createActiveProduct(t)

	err := product.RemoveDiscount(time.Now())
	assert.ErrorIs(t, err, domain.ErrNoDiscountToRemove)
}

func TestProduct_EffectivePrice(t *testing.T) {
	now := time.Now()
	basePrice, _ := domain.NewMoney(10000, 100) // $100.00
	product, _ := domain.NewProduct("test-id", "Product", "Description", "Category", basePrice, now)
	product.Activate(now)

	// Without discount
	effectivePrice := product.EffectivePrice(now)
	assert.Equal(t, "100.00", effectivePrice.String())

	// With 20% discount
	discount, _ := domain.NewDiscount(20, now, now.Add(7*24*time.Hour))
	product.ApplyDiscount(discount, now)

	effectivePrice = product.EffectivePrice(now)
	assert.Equal(t, "80.00", effectivePrice.String())
}

func TestProduct_EffectivePriceExpiredDiscount(t *testing.T) {
	now := time.Now()
	basePrice, _ := domain.NewMoney(10000, 100) // $100.00
	product, _ := domain.NewProduct("test-id", "Product", "Description", "Category", basePrice, now)
	product.Activate(now)

	// Apply discount that was valid yesterday
	discount, _ := domain.NewDiscount(20, now.Add(-48*time.Hour), now.Add(-24*time.Hour))
	product.ApplyDiscount(discount, now.Add(-36*time.Hour))

	// Effective price should be base price since discount expired
	effectivePrice := product.EffectivePrice(now)
	assert.Equal(t, "100.00", effectivePrice.String())
}

// Helper functions

func createActiveProduct(t *testing.T) *domain.Product {
	now := time.Now()
	basePrice, _ := domain.NewMoney(1999, 100)
	product, err := domain.NewProduct("test-id", "Active Product", "Description", "Category", basePrice, now)
	require.NoError(t, err)
	err = product.Activate(now)
	require.NoError(t, err)
	return product
}

func createArchivedProduct(t *testing.T) *domain.Product {
	now := time.Now()
	basePrice, _ := domain.NewMoney(1999, 100)
	product, err := domain.NewProduct("test-id", "Archived Product", "Description", "Category", basePrice, now)
	require.NoError(t, err)
	err = product.Archive(now)
	require.NoError(t, err)
	return product
}

func createProductWithDiscount(t *testing.T) *domain.Product {
	now := time.Now()
	product := createActiveProduct(t)
	discount, err := domain.NewDiscount(20, now, now.Add(7*24*time.Hour))
	require.NoError(t, err)
	err = product.ApplyDiscount(discount, now)
	require.NoError(t, err)
	return product
}
