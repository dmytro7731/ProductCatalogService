package product

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/product-catalog-service/internal/app/product/queries/get_product"
	"github.com/product-catalog-service/internal/app/product/queries/list_products"
	"github.com/product-catalog-service/internal/app/product/usecases/activate_product"
	"github.com/product-catalog-service/internal/app/product/usecases/apply_discount"
	"github.com/product-catalog-service/internal/app/product/usecases/archive_product"
	"github.com/product-catalog-service/internal/app/product/usecases/create_product"
	"github.com/product-catalog-service/internal/app/product/usecases/deactivate_product"
	"github.com/product-catalog-service/internal/app/product/usecases/remove_discount"
	"github.com/product-catalog-service/internal/app/product/usecases/update_product"
	pb "github.com/product-catalog-service/proto/product/v1"
)

// Commands holds all command usecases.
type Commands struct {
	CreateProduct     *create_product.Interactor
	UpdateProduct     *update_product.Interactor
	ActivateProduct   *activate_product.Interactor
	DeactivateProduct *deactivate_product.Interactor
	ArchiveProduct    *archive_product.Interactor
	ApplyDiscount     *apply_discount.Interactor
	RemoveDiscount    *remove_discount.Interactor
}

// Queries holds all query handlers.
type Queries struct {
	GetProduct   *get_product.Query
	ListProducts *list_products.Query
}

// Handler implements the ProductServiceServer interface.
type Handler struct {
	pb.UnimplementedProductServiceServer
	commands Commands
	queries  Queries
}

// NewHandler creates a new product gRPC handler.
func NewHandler(commands Commands, queries Queries) *Handler {
	return &Handler{
		commands: commands,
		queries:  queries,
	}
}

// CreateProduct creates a new product.
func (h *Handler) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductReply, error) {
	// 1. Validate proto request
	if err := validateCreateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 2. Map proto to application request
	appReq := mapToCreateProductRequest(req)

	// 3. Call usecase
	productID, err := h.commands.CreateProduct.Execute(ctx, appReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	// 4. Return response
	return &pb.CreateProductReply{
		ProductId: productID,
	}, nil
}

// UpdateProduct updates an existing product.
func (h *Handler) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.UpdateProductReply, error) {
	if err := validateUpdateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	appReq := mapToUpdateProductRequest(req)

	if err := h.commands.UpdateProduct.Execute(ctx, appReq); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.UpdateProductReply{}, nil
}

// ActivateProduct activates a product.
func (h *Handler) ActivateProduct(ctx context.Context, req *pb.ActivateProductRequest) (*pb.ActivateProductReply, error) {
	if err := validateActivateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	appReq := activate_product.Request{
		ProductID: req.GetProductId(),
	}

	if err := h.commands.ActivateProduct.Execute(ctx, appReq); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.ActivateProductReply{}, nil
}

// DeactivateProduct deactivates a product.
func (h *Handler) DeactivateProduct(ctx context.Context, req *pb.DeactivateProductRequest) (*pb.DeactivateProductReply, error) {
	if err := validateDeactivateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	appReq := deactivate_product.Request{
		ProductID: req.GetProductId(),
	}

	if err := h.commands.DeactivateProduct.Execute(ctx, appReq); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.DeactivateProductReply{}, nil
}

// ArchiveProduct archives (soft deletes) a product.
func (h *Handler) ArchiveProduct(ctx context.Context, req *pb.ArchiveProductRequest) (*pb.ArchiveProductReply, error) {
	if err := validateArchiveRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	appReq := archive_product.Request{
		ProductID: req.GetProductId(),
	}

	if err := h.commands.ArchiveProduct.Execute(ctx, appReq); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.ArchiveProductReply{}, nil
}

// ApplyDiscount applies a discount to a product.
func (h *Handler) ApplyDiscount(ctx context.Context, req *pb.ApplyDiscountRequest) (*pb.ApplyDiscountReply, error) {
	if err := validateApplyDiscountRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	appReq := mapToApplyDiscountRequest(req)

	if err := h.commands.ApplyDiscount.Execute(ctx, appReq); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.ApplyDiscountReply{}, nil
}

// RemoveDiscount removes a discount from a product.
func (h *Handler) RemoveDiscount(ctx context.Context, req *pb.RemoveDiscountRequest) (*pb.RemoveDiscountReply, error) {
	if err := validateRemoveDiscountRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	appReq := remove_discount.Request{
		ProductID: req.GetProductId(),
	}

	if err := h.commands.RemoveDiscount.Execute(ctx, appReq); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.RemoveDiscountReply{}, nil
}

// GetProduct retrieves a product by ID.
func (h *Handler) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductReply, error) {
	if err := validateGetProductRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	queryReq := mapToGetProductRequest(req)

	product, err := h.queries.GetProduct.Execute(ctx, queryReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.GetProductReply{
		Product: mapProductDTOToProto(product),
	}, nil
}

// ListProducts retrieves a paginated list of products.
func (h *Handler) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsReply, error) {
	if err := validateListProductsRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	queryReq := mapToListProductsRequest(req)

	result, err := h.queries.ListProducts.Execute(ctx, queryReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return mapListResultToProto(result), nil
}
