package apply_discount

import (
	"context"
	"time"

	"github.com/product-catalog-service/internal/app/product/domain"
	"github.com/product-catalog-service/internal/app/product/repo"
	"github.com/product-catalog-service/internal/pkg/clock"
	"github.com/product-catalog-service/internal/pkg/committer"
)

// Request represents the input for applying a discount.
type Request struct {
	ProductID  string
	Percentage int64
	StartDate  time.Time
	EndDate    time.Time
}

// Interactor handles the apply discount use case.
type Interactor struct {
	productRepo *repo.ProductRepo
	outboxRepo  *repo.OutboxRepo
	committer   committer.Committer
	clock       clock.Clock
}

// NewInteractor creates a new apply discount interactor.
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

// Execute applies a discount to a product.
func (it *Interactor) Execute(ctx context.Context, req Request) error {
	// 1. Load existing product aggregate
	product, err := it.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return err
	}

	// 2. Create discount value object
	discount, err := domain.NewDiscount(req.Percentage, req.StartDate, req.EndDate)
	if err != nil {
		return err
	}

	// 3. Apply domain logic
	if err := product.ApplyDiscount(discount, it.clock.Now()); err != nil {
		return err
	}

	// 4. Build commit plan
	plan := committer.NewPlan()

	// 5. Get update mutation from repository
	if mut := it.productRepo.UpdateMut(product); mut != nil {
		plan.Add(mut)
	}

	// 6. Add outbox events
	for _, event := range product.DomainEvents() {
		outboxMut, err := it.outboxRepo.InsertFromDomainEventMut(event)
		if err != nil {
			return err
		}
		plan.Add(outboxMut)
	}

	// 7. Apply plan atomically
	if err := it.committer.Apply(ctx, plan); err != nil {
		return err
	}

	return nil
}
