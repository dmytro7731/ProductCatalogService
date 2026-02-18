package services

import (
	"cloud.google.com/go/spanner"
	"github.com/product-catalog-service/internal/app/product/queries/get_product"
	"github.com/product-catalog-service/internal/app/product/queries/list_products"
	"github.com/product-catalog-service/internal/app/product/repo"
	"github.com/product-catalog-service/internal/app/product/usecases/activate_product"
	"github.com/product-catalog-service/internal/app/product/usecases/apply_discount"
	"github.com/product-catalog-service/internal/app/product/usecases/archive_product"
	"github.com/product-catalog-service/internal/app/product/usecases/create_product"
	"github.com/product-catalog-service/internal/app/product/usecases/deactivate_product"
	"github.com/product-catalog-service/internal/app/product/usecases/remove_discount"
	"github.com/product-catalog-service/internal/app/product/usecases/update_product"
	grpcHandler "github.com/product-catalog-service/internal/transport/grpc/product"
	"github.com/product-catalog-service/internal/pkg/clock"
	"github.com/product-catalog-service/internal/pkg/committer"
)

// Container holds all service dependencies.
type Container struct {
	// Infrastructure
	SpannerClient *spanner.Client
	Clock         clock.Clock
	Committer     committer.Committer

	// Repositories
	ProductRepo   *repo.ProductRepo
	OutboxRepo    *repo.OutboxRepo
	ReadModelRepo *repo.ReadModelRepo

	// Commands
	CreateProductUsecase     *create_product.Interactor
	UpdateProductUsecase     *update_product.Interactor
	ActivateProductUsecase   *activate_product.Interactor
	DeactivateProductUsecase *deactivate_product.Interactor
	ArchiveProductUsecase    *archive_product.Interactor
	ApplyDiscountUsecase     *apply_discount.Interactor
	RemoveDiscountUsecase    *remove_discount.Interactor

	// Queries
	GetProductQuery   *get_product.Query
	ListProductsQuery *list_products.Query

	// gRPC Handler
	ProductHandler *grpcHandler.Handler
}

// NewContainer creates a new dependency injection container.
func NewContainer(spannerClient *spanner.Client) *Container {
	c := &Container{
		SpannerClient: spannerClient,
	}

	// Initialize clock
	c.Clock = clock.NewRealClock()

	// Initialize committer
	c.Committer = committer.NewSpannerCommitter(spannerClient)

	// Initialize repositories
	c.ProductRepo = repo.NewProductRepo(spannerClient)
	c.OutboxRepo = repo.NewOutboxRepo(c.Clock)
	c.ReadModelRepo = repo.NewReadModelRepo(spannerClient, c.Clock)

	// Initialize usecases
	c.CreateProductUsecase = create_product.NewInteractor(
		c.ProductRepo,
		c.OutboxRepo,
		c.Committer,
		c.Clock,
	)

	c.UpdateProductUsecase = update_product.NewInteractor(
		c.ProductRepo,
		c.OutboxRepo,
		c.Committer,
		c.Clock,
	)

	c.ActivateProductUsecase = activate_product.NewInteractor(
		c.ProductRepo,
		c.OutboxRepo,
		c.Committer,
		c.Clock,
	)

	c.DeactivateProductUsecase = deactivate_product.NewInteractor(
		c.ProductRepo,
		c.OutboxRepo,
		c.Committer,
		c.Clock,
	)

	c.ArchiveProductUsecase = archive_product.NewInteractor(
		c.ProductRepo,
		c.OutboxRepo,
		c.Committer,
		c.Clock,
	)

	c.ApplyDiscountUsecase = apply_discount.NewInteractor(
		c.ProductRepo,
		c.OutboxRepo,
		c.Committer,
		c.Clock,
	)

	c.RemoveDiscountUsecase = remove_discount.NewInteractor(
		c.ProductRepo,
		c.OutboxRepo,
		c.Committer,
		c.Clock,
	)

	// Initialize queries
	c.GetProductQuery = get_product.NewQuery(c.ReadModelRepo)
	c.ListProductsQuery = list_products.NewQuery(c.ReadModelRepo)

	// Initialize gRPC handler
	commands := grpcHandler.Commands{
		CreateProduct:     c.CreateProductUsecase,
		UpdateProduct:     c.UpdateProductUsecase,
		ActivateProduct:   c.ActivateProductUsecase,
		DeactivateProduct: c.DeactivateProductUsecase,
		ArchiveProduct:    c.ArchiveProductUsecase,
		ApplyDiscount:     c.ApplyDiscountUsecase,
		RemoveDiscount:    c.RemoveDiscountUsecase,
	}

	queries := grpcHandler.Queries{
		GetProduct:   c.GetProductQuery,
		ListProducts: c.ListProductsQuery,
	}

	c.ProductHandler = grpcHandler.NewHandler(commands, queries)

	return c
}

// NewContainerWithClock creates a container with a custom clock (for testing).
func NewContainerWithClock(spannerClient *spanner.Client, clk clock.Clock) *Container {
	c := &Container{
		SpannerClient: spannerClient,
		Clock:         clk,
	}

	// Initialize committer
	c.Committer = committer.NewSpannerCommitter(spannerClient)

	// Initialize repositories
	c.ProductRepo = repo.NewProductRepo(spannerClient)
	c.OutboxRepo = repo.NewOutboxRepo(c.Clock)
	c.ReadModelRepo = repo.NewReadModelRepo(spannerClient, c.Clock)

	// Initialize usecases
	c.CreateProductUsecase = create_product.NewInteractor(
		c.ProductRepo,
		c.OutboxRepo,
		c.Committer,
		c.Clock,
	)

	c.UpdateProductUsecase = update_product.NewInteractor(
		c.ProductRepo,
		c.OutboxRepo,
		c.Committer,
		c.Clock,
	)

	c.ActivateProductUsecase = activate_product.NewInteractor(
		c.ProductRepo,
		c.OutboxRepo,
		c.Committer,
		c.Clock,
	)

	c.DeactivateProductUsecase = deactivate_product.NewInteractor(
		c.ProductRepo,
		c.OutboxRepo,
		c.Committer,
		c.Clock,
	)

	c.ArchiveProductUsecase = archive_product.NewInteractor(
		c.ProductRepo,
		c.OutboxRepo,
		c.Committer,
		c.Clock,
	)

	c.ApplyDiscountUsecase = apply_discount.NewInteractor(
		c.ProductRepo,
		c.OutboxRepo,
		c.Committer,
		c.Clock,
	)

	c.RemoveDiscountUsecase = remove_discount.NewInteractor(
		c.ProductRepo,
		c.OutboxRepo,
		c.Committer,
		c.Clock,
	)

	// Initialize queries
	c.GetProductQuery = get_product.NewQuery(c.ReadModelRepo)
	c.ListProductsQuery = list_products.NewQuery(c.ReadModelRepo)

	// Initialize gRPC handler
	commands := grpcHandler.Commands{
		CreateProduct:     c.CreateProductUsecase,
		UpdateProduct:     c.UpdateProductUsecase,
		ActivateProduct:   c.ActivateProductUsecase,
		DeactivateProduct: c.DeactivateProductUsecase,
		ArchiveProduct:    c.ArchiveProductUsecase,
		ApplyDiscount:     c.ApplyDiscountUsecase,
		RemoveDiscount:    c.RemoveDiscountUsecase,
	}

	queries := grpcHandler.Queries{
		GetProduct:   c.GetProductQuery,
		ListProducts: c.ListProductsQuery,
	}

	c.ProductHandler = grpcHandler.NewHandler(commands, queries)

	return c
}
