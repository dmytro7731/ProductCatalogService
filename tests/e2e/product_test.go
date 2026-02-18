package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/iterator"

	"github.com/product-catalog-service/internal/app/product/domain"
	"github.com/product-catalog-service/internal/app/product/queries/get_product"
	"github.com/product-catalog-service/internal/app/product/queries/list_products"
	"github.com/product-catalog-service/internal/app/product/usecases/activate_product"
	"github.com/product-catalog-service/internal/app/product/usecases/apply_discount"
	"github.com/product-catalog-service/internal/app/product/usecases/archive_product"
	"github.com/product-catalog-service/internal/app/product/usecases/create_product"
	"github.com/product-catalog-service/internal/app/product/usecases/deactivate_product"
	"github.com/product-catalog-service/internal/app/product/usecases/remove_discount"
	"github.com/product-catalog-service/internal/app/product/usecases/update_product"
	"github.com/product-catalog-service/internal/models/m_outbox"
	"github.com/product-catalog-service/internal/pkg/clock"
	"github.com/product-catalog-service/internal/services"
)

var (
	testContainer *services.Container
	testClient    *spanner.Client
	testClock     *clock.MockClock
)

func TestMain(m *testing.M) {
	// Check if we're running with emulator
	if os.Getenv("SPANNER_EMULATOR_HOST") == "" {
		fmt.Println("Skipping E2E tests: SPANNER_EMULATOR_HOST not set")
		os.Exit(0)
	}

	// Setup
	ctx := context.Background()
	var err error

	database := fmt.Sprintf(
		"projects/%s/instances/%s/databases/%s",
		getEnv("SPANNER_PROJECT", "test-project"),
		getEnv("SPANNER_INSTANCE", "test-instance"),
		getEnv("SPANNER_DATABASE", "product-catalog"),
	)

	testClient, err = spanner.NewClient(ctx, database)
	if err != nil {
		fmt.Printf("Failed to create Spanner client: %v\n", err)
		os.Exit(1)
	}

	testClock = clock.NewMockClock(time.Date(2026, 2, 18, 12, 0, 0, 0, time.UTC))
	testContainer = services.NewContainerWithClock(testClient, testClock)

	// Run tests
	code := m.Run()

	// Cleanup
	testClient.Close()

	os.Exit(code)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func cleanupDatabase(t *testing.T, ctx context.Context) {
	// Delete all products
	_, err := testClient.Apply(ctx, []*spanner.Mutation{
		spanner.Delete("products", spanner.AllKeys()),
		spanner.Delete("outbox_events", spanner.AllKeys()),
	})
	require.NoError(t, err)
}

func getOutboxEvents(t *testing.T, ctx context.Context, aggregateID string) []outboxEvent {
	query := fmt.Sprintf(
		"SELECT event_id, event_type, aggregate_id, payload, status, created_at FROM %s WHERE aggregate_id = @aggregateID ORDER BY created_at",
		m_outbox.TableName,
	)

	stmt := spanner.Statement{
		SQL:    query,
		Params: map[string]interface{}{"aggregateID": aggregateID},
	}

	iter := testClient.Single().Query(ctx, stmt)
	defer iter.Stop()

	var events []outboxEvent
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		require.NoError(t, err)

		var event outboxEvent
		var payload spanner.NullJSON
		err = row.Columns(&event.ID, &event.EventType, &event.AggregateID, &payload, &event.Status, &event.CreatedAt)
		require.NoError(t, err)

		if payload.Valid {
			event.Payload = payload.Value
		}

		events = append(events, event)
	}

	return events
}

type outboxEvent struct {
	ID          string
	EventType   string
	AggregateID string
	Payload     interface{}
	Status      string
	CreatedAt   time.Time
}

// TestProductCreationFlow tests the complete product creation flow
func TestProductCreationFlow(t *testing.T) {
	ctx := context.Background()
	cleanupDatabase(t, ctx)

	// Create product
	req := create_product.Request{
		Name:                 "Test Product",
		Description:          "A test product for E2E testing",
		Category:             "Electronics",
		BasePriceNumerator:   1999,
		BasePriceDenominator: 100,
	}

	productID, err := testContainer.CreateProductUsecase.Execute(ctx, req)
	require.NoError(t, err)
	require.NotEmpty(t, productID)

	// Verify: Query returns correct data
	product, err := testContainer.GetProductQuery.Execute(ctx, get_product.Request{ProductID: productID})
	require.NoError(t, err)

	assert.Equal(t, productID, product.ID)
	assert.Equal(t, "Test Product", product.Name)
	assert.Equal(t, "A test product for E2E testing", product.Description)
	assert.Equal(t, "Electronics", product.Category)
	assert.Equal(t, int64(1999), product.BasePriceNumerator)
	assert.Equal(t, int64(100), product.BasePriceDenominator)
	assert.Equal(t, "draft", product.Status)

	// Verify: Outbox event was created
	events := getOutboxEvents(t, ctx, productID)
	require.Len(t, events, 1)
	assert.Equal(t, "product.created", events[0].EventType)
	assert.Equal(t, m_outbox.StatusPending, events[0].Status)
}

// TestProductUpdateFlow tests product update functionality
func TestProductUpdateFlow(t *testing.T) {
	ctx := context.Background()
	cleanupDatabase(t, ctx)

	// Setup: Create a product
	productID := createTestProduct(t, ctx)

	// Update product
	err := testContainer.UpdateProductUsecase.Execute(ctx, update_product.Request{
		ProductID:   productID,
		Name:        "Updated Product Name",
		Description: "Updated description",
		Category:    "Updated Category",
	})
	require.NoError(t, err)

	// Verify changes
	product, err := testContainer.GetProductQuery.Execute(ctx, get_product.Request{ProductID: productID})
	require.NoError(t, err)

	assert.Equal(t, "Updated Product Name", product.Name)
	assert.Equal(t, "Updated description", product.Description)
	assert.Equal(t, "Updated Category", product.Category)

	// Verify outbox event
	events := getOutboxEvents(t, ctx, productID)
	require.GreaterOrEqual(t, len(events), 2, "should have at least 2 events (created + updated)")
	// Find the updated event
	var hasUpdatedEvent bool
	for _, e := range events {
		if e.EventType == "product.updated" {
			hasUpdatedEvent = true
			break
		}
	}
	assert.True(t, hasUpdatedEvent, "should have product.updated event")
}

// TestProductActivationDeactivationFlow tests activation and deactivation
func TestProductActivationDeactivationFlow(t *testing.T) {
	ctx := context.Background()
	cleanupDatabase(t, ctx)

	// Setup: Create a product
	productID := createTestProduct(t, ctx)

	// Activate
	err := testContainer.ActivateProductUsecase.Execute(ctx, activate_product.Request{
		ProductID: productID,
	})
	require.NoError(t, err)

	product, _ := testContainer.GetProductQuery.Execute(ctx, get_product.Request{ProductID: productID})
	assert.Equal(t, "active", product.Status)

	// Deactivate
	err = testContainer.DeactivateProductUsecase.Execute(ctx, deactivate_product.Request{
		ProductID: productID,
	})
	require.NoError(t, err)

	product, _ = testContainer.GetProductQuery.Execute(ctx, get_product.Request{ProductID: productID})
	assert.Equal(t, "inactive", product.Status)

	// Verify events
	events := getOutboxEvents(t, ctx, productID)
	require.GreaterOrEqual(t, len(events), 3, "should have at least 3 events")
	// Find the expected events
	var hasActivatedEvent, hasDeactivatedEvent bool
	for _, e := range events {
		if e.EventType == "product.activated" {
			hasActivatedEvent = true
		}
		if e.EventType == "product.deactivated" {
			hasDeactivatedEvent = true
		}
	}
	assert.True(t, hasActivatedEvent, "should have product.activated event")
	assert.True(t, hasDeactivatedEvent, "should have product.deactivated event")
}

// TestDiscountApplicationFlow tests discount functionality
func TestDiscountApplicationFlow(t *testing.T) {
	ctx := context.Background()
	cleanupDatabase(t, ctx)

	// Setup: Create and activate product
	productID := createAndActivateProduct(t, ctx)

	now := testClock.Now()
	startDate := now
	endDate := now.Add(7 * 24 * time.Hour)

	// Apply 20% discount
	err := testContainer.ApplyDiscountUsecase.Execute(ctx, apply_discount.Request{
		ProductID:  productID,
		Percentage: 20,
		StartDate:  startDate,
		EndDate:    endDate,
	})
	require.NoError(t, err)

	// Verify discount was applied
	product, err := testContainer.GetProductQuery.Execute(ctx, get_product.Request{ProductID: productID})
	require.NoError(t, err)

	assert.NotNil(t, product.DiscountPercent)
	assert.Equal(t, int64(20), *product.DiscountPercent)

	// Verify effective price is calculated correctly
	// Base: $19.99, Discount: 20%, Effective: $15.992 (stored as 15992/1000 or simplified)
	expectedEffectiveNum := product.BasePriceNumerator * (100 - 20)
	expectedEffectiveDenom := product.BasePriceDenominator * 100
	assert.Equal(t, expectedEffectiveNum, product.EffectivePriceNum)
	assert.Equal(t, expectedEffectiveDenom, product.EffectivePriceDenom)

	// Verify outbox event
	events := getOutboxEvents(t, ctx, productID)
	var discountEvent *outboxEvent
	for i := range events {
		if events[i].EventType == "product.discount_applied" {
			discountEvent = &events[i]
			break
		}
	}
	require.NotNil(t, discountEvent)
}

// TestDiscountRemoval tests removing a discount
func TestDiscountRemoval(t *testing.T) {
	ctx := context.Background()
	cleanupDatabase(t, ctx)

	// Setup: Create product with discount
	productID := createProductWithDiscount(t, ctx)

	// Remove discount
	err := testContainer.RemoveDiscountUsecase.Execute(ctx, remove_discount.Request{
		ProductID: productID,
	})
	require.NoError(t, err)

	// Verify discount removed
	product, err := testContainer.GetProductQuery.Execute(ctx, get_product.Request{ProductID: productID})
	require.NoError(t, err)

	assert.Nil(t, product.DiscountPercent)
	assert.Equal(t, product.BasePriceNumerator, product.EffectivePriceNum)
	assert.Equal(t, product.BasePriceDenominator, product.EffectivePriceDenom)
}

// TestProductArchiving tests soft delete functionality
func TestProductArchiving(t *testing.T) {
	ctx := context.Background()
	cleanupDatabase(t, ctx)

	// Setup: Create a product (in draft status)
	productID := createTestProduct(t, ctx)

	// Archive
	err := testContainer.ArchiveProductUsecase.Execute(ctx, archive_product.Request{
		ProductID: productID,
	})
	require.NoError(t, err)

	// Verify archived
	product, err := testContainer.GetProductQuery.Execute(ctx, get_product.Request{ProductID: productID})
	require.NoError(t, err)
	assert.Equal(t, "archived", product.Status)

	// Verify doesn't appear in active listings
	result, err := testContainer.ListProductsQuery.Execute(ctx, list_products.Request{
		ActiveOnly: true,
		Limit:      100,
	})
	require.NoError(t, err)

	for _, p := range result.Products {
		assert.NotEqual(t, productID, p.ID, "Archived product should not appear in active listing")
	}
}

// TestBusinessRuleValidation tests domain error handling
func TestBusinessRuleValidation(t *testing.T) {
	ctx := context.Background()
	cleanupDatabase(t, ctx)

	t.Run("cannot apply discount to inactive product", func(t *testing.T) {
		productID := createTestProduct(t, ctx) // draft status

		now := testClock.Now()
		err := testContainer.ApplyDiscountUsecase.Execute(ctx, apply_discount.Request{
			ProductID:  productID,
			Percentage: 20,
			StartDate:  now,
			EndDate:    now.Add(24 * time.Hour),
		})

		assert.ErrorIs(t, err, domain.ErrProductNotActive)
	})

	t.Run("cannot activate already active product", func(t *testing.T) {
		cleanupDatabase(t, ctx)
		productID := createAndActivateProduct(t, ctx)

		err := testContainer.ActivateProductUsecase.Execute(ctx, activate_product.Request{
			ProductID: productID,
		})

		assert.ErrorIs(t, err, domain.ErrProductAlreadyActive)
	})

	t.Run("cannot archive active product", func(t *testing.T) {
		cleanupDatabase(t, ctx)
		productID := createAndActivateProduct(t, ctx)

		err := testContainer.ArchiveProductUsecase.Execute(ctx, archive_product.Request{
			ProductID: productID,
		})

		assert.ErrorIs(t, err, domain.ErrCannotArchiveActive)
	})

	t.Run("cannot remove non-existent discount", func(t *testing.T) {
		cleanupDatabase(t, ctx)
		productID := createAndActivateProduct(t, ctx)

		err := testContainer.RemoveDiscountUsecase.Execute(ctx, remove_discount.Request{
			ProductID: productID,
		})

		assert.ErrorIs(t, err, domain.ErrNoDiscountToRemove)
	})
}

// TestProductListing tests pagination and filtering
func TestProductListing(t *testing.T) {
	ctx := context.Background()
	cleanupDatabase(t, ctx)

	// Create multiple products in different categories
	createProductInCategory(t, ctx, "Electronics", true)
	createProductInCategory(t, ctx, "Electronics", true)
	createProductInCategory(t, ctx, "Books", true)
	createProductInCategory(t, ctx, "Electronics", false) // inactive

	t.Run("list all active products", func(t *testing.T) {
		result, err := testContainer.ListProductsQuery.Execute(ctx, list_products.Request{
			ActiveOnly: true,
			Limit:      100,
		})
		require.NoError(t, err)

		assert.Equal(t, int64(3), result.TotalCount)
		assert.Len(t, result.Products, 3)
	})

	t.Run("filter by category", func(t *testing.T) {
		category := "Electronics"
		result, err := testContainer.ListProductsQuery.Execute(ctx, list_products.Request{
			Category:   &category,
			ActiveOnly: true,
			Limit:      100,
		})
		require.NoError(t, err)

		assert.Equal(t, int64(2), result.TotalCount)
		for _, p := range result.Products {
			assert.Equal(t, "Electronics", p.Category)
		}
	})

	t.Run("pagination", func(t *testing.T) {
		result, err := testContainer.ListProductsQuery.Execute(ctx, list_products.Request{
			ActiveOnly: true,
			Limit:      2,
			Offset:     0,
		})
		require.NoError(t, err)

		assert.Equal(t, int64(3), result.TotalCount)
		assert.Len(t, result.Products, 2)
		assert.True(t, result.HasMore)

		// Next page
		result, err = testContainer.ListProductsQuery.Execute(ctx, list_products.Request{
			ActiveOnly: true,
			Limit:      2,
			Offset:     2,
		})
		require.NoError(t, err)

		assert.Len(t, result.Products, 1)
		assert.False(t, result.HasMore)
	})
}

// TestOutboxEventPayload verifies event payload structure
func TestOutboxEventPayload(t *testing.T) {
	ctx := context.Background()
	cleanupDatabase(t, ctx)

	productID := createTestProduct(t, ctx)

	events := getOutboxEvents(t, ctx, productID)
	require.Len(t, events, 1)

	payload, ok := events[0].Payload.(map[string]interface{})
	require.True(t, ok, "Payload should be a map")

	assert.Contains(t, payload, "event_type")
	assert.Contains(t, payload, "aggregate_id")
	assert.Contains(t, payload, "occurred_at")
	assert.Contains(t, payload, "name")
	assert.Contains(t, payload, "category")
}

// Helper functions

func createTestProduct(t *testing.T, ctx context.Context) string {
	productID, err := testContainer.CreateProductUsecase.Execute(ctx, create_product.Request{
		Name:                 fmt.Sprintf("Test Product %d", time.Now().UnixNano()),
		Description:          "Test description",
		Category:             "Test Category",
		BasePriceNumerator:   1999,
		BasePriceDenominator: 100,
	})
	require.NoError(t, err)
	return productID
}

func createAndActivateProduct(t *testing.T, ctx context.Context) string {
	productID := createTestProduct(t, ctx)
	err := testContainer.ActivateProductUsecase.Execute(ctx, activate_product.Request{
		ProductID: productID,
	})
	require.NoError(t, err)
	return productID
}

func createProductWithDiscount(t *testing.T, ctx context.Context) string {
	productID := createAndActivateProduct(t, ctx)
	now := testClock.Now()
	err := testContainer.ApplyDiscountUsecase.Execute(ctx, apply_discount.Request{
		ProductID:  productID,
		Percentage: 20,
		StartDate:  now,
		EndDate:    now.Add(7 * 24 * time.Hour),
	})
	require.NoError(t, err)
	return productID
}

func createProductInCategory(t *testing.T, ctx context.Context, category string, activate bool) string {
	productID, err := testContainer.CreateProductUsecase.Execute(ctx, create_product.Request{
		Name:                 fmt.Sprintf("Product %d", time.Now().UnixNano()),
		Description:          "Description",
		Category:             category,
		BasePriceNumerator:   1000,
		BasePriceDenominator: 100,
	})
	require.NoError(t, err)

	if activate {
		err = testContainer.ActivateProductUsecase.Execute(ctx, activate_product.Request{
			ProductID: productID,
		})
		require.NoError(t, err)
	}

	return productID
}
