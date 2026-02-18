package remove_discount

import (
	"context"

	"github.com/product-catalog-service/internal/app/product/repo"
	"github.com/product-catalog-service/internal/pkg/clock"
	"github.com/product-catalog-service/internal/pkg/committer"
)

// Request represents the input for removing a discount.
type Request struct {
	ProductID string
}

// Interactor handles the remove discount use case.
type Interactor struct {
	productRepo *repo.ProductRepo
	outboxRepo  *repo.OutboxRepo
	committer   committer.Committer
	clock       clock.Clock
}

// NewInteractor creates a new remove discount interactor.
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

// Execute removes a discount from a product.
func (it *Interactor) Execute(ctx context.Context, req Request) error {
	// 1. Load existing product aggregate
	product, err := it.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return err
	}

	// 2. Apply domain logic
	if err := product.RemoveDiscount(it.clock.Now()); err != nil {
		return err
	}

	// 3. Build commit plan
	plan := committer.NewPlan()

	// 4. Get update mutation from repository
	if mut := it.productRepo.UpdateMut(product); mut != nil {
		plan.Add(mut)
	}

	// 5. Add outbox events
	for _, event := range product.DomainEvents() {
		outboxMut, err := it.outboxRepo.InsertFromDomainEventMut(event)
		if err != nil {
			return err
		}
		plan.Add(outboxMut)
	}

	// 6. Apply plan atomically
	if err := it.committer.Apply(ctx, plan); err != nil {
		return err
	}

	return nil
}
