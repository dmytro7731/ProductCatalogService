package create_product

import (
	"context"

	"github.com/google/uuid"

	"github.com/product-catalog-service/internal/app/product/domain"
	"github.com/product-catalog-service/internal/app/product/repo"
	"github.com/product-catalog-service/internal/pkg/clock"
	"github.com/product-catalog-service/internal/pkg/committer"
)

// Request represents the input for creating a product.
type Request struct {
	Name                 string
	Description          string
	Category             string
	BasePriceNumerator   int64
	BasePriceDenominator int64
}

// Interactor handles the create product use case.
type Interactor struct {
	productRepo *repo.ProductRepo
	outboxRepo  *repo.OutboxRepo
	committer   committer.Committer
	clock       clock.Clock
}

// NewInteractor creates a new create product interactor.
func NewInteractor(
	productRepo *repo.ProductRepo,
	outboxRepo *repo.OutboxRepo,
	committer committer.Committer,
	clock clock.Clock,
) *Interactor {
	return &Interactor{
		productRepo: productRepo,
		outboxRepo:  outboxRepo,
		committer:   committer,
		clock:       clock,
	}
}

// Execute creates a new product.
func (it *Interactor) Execute(ctx context.Context, req Request) (string, error) {
	// 1. Create base price value object
	basePrice, err := domain.NewMoney(req.BasePriceNumerator, req.BasePriceDenominator)
	if err != nil {
		return "", err
	}

	// 2. Create new product aggregate
	productID := uuid.New().String()
	product, err := domain.NewProduct(
		productID,
		req.Name,
		req.Description,
		req.Category,
		basePrice,
		it.clock.Now(),
	)
	if err != nil {
		return "", err
	}

	// 3. Build commit plan
	plan := committer.NewPlan()

	// 4. Get insert mutation from repository
	if mut := it.productRepo.InsertMut(product); mut != nil {
		plan.Add(mut)
	}

	// 5. Add outbox events
	for _, event := range product.DomainEvents() {
		outboxMut, err := it.outboxRepo.InsertFromDomainEventMut(event)
		if err != nil {
			return "", err
		}
		plan.Add(outboxMut)
	}

	// 6. Apply plan atomically
	if err := it.committer.Apply(ctx, plan); err != nil {
		return "", err
	}

	return product.ID(), nil
}
