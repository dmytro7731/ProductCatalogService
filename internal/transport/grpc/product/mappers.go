package product

import (
	"github.com/product-catalog-service/internal/app/product/queries/get_product"
	"github.com/product-catalog-service/internal/app/product/queries/list_products"
	"github.com/product-catalog-service/internal/app/product/usecases/apply_discount"
	"github.com/product-catalog-service/internal/app/product/usecases/create_product"
	"github.com/product-catalog-service/internal/app/product/usecases/update_product"
	pb "github.com/product-catalog-service/proto/product/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// mapToCreateProductRequest converts proto request to application request.
func mapToCreateProductRequest(req *pb.CreateProductRequest) create_product.Request {
	var num, denom int64 = 0, 1
	if req.GetBasePrice() != nil {
		num = req.GetBasePrice().GetNumerator()
		denom = req.GetBasePrice().GetDenominator()
		if denom == 0 {
			denom = 1
		}
	}

	return create_product.Request{
		Name:                 req.GetName(),
		Description:          req.GetDescription(),
		Category:             req.GetCategory(),
		BasePriceNumerator:   num,
		BasePriceDenominator: denom,
	}
}

// mapToUpdateProductRequest converts proto request to application request.
func mapToUpdateProductRequest(req *pb.UpdateProductRequest) update_product.Request {
	return update_product.Request{
		ProductID:   req.GetProductId(),
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Category:    req.GetCategory(),
	}
}

// mapToApplyDiscountRequest converts proto request to application request.
func mapToApplyDiscountRequest(req *pb.ApplyDiscountRequest) apply_discount.Request {
	return apply_discount.Request{
		ProductID:  req.GetProductId(),
		Percentage: req.GetPercentage(),
		StartDate:  pb.TimestampToTime(req.GetStartDate()),
		EndDate:    pb.TimestampToTime(req.GetEndDate()),
	}
}

// mapToGetProductRequest converts proto request to query request.
func mapToGetProductRequest(req *pb.GetProductRequest) get_product.Request {
	return get_product.Request{
		ProductID: req.GetProductId(),
	}
}

// mapToListProductsRequest converts proto request to query request.
func mapToListProductsRequest(req *pb.ListProductsRequest) list_products.Request {
	var category, status *string

	if req.Category != nil {
		cat := req.GetCategory()
		category = &cat
	}

	if req.Status != nil {
		st := req.GetStatus()
		status = &st
	}

	return list_products.Request{
		Category:   category,
		Status:     status,
		ActiveOnly: req.GetActiveOnly(),
		Limit:      int(req.GetLimit()),
		Offset:     int(req.GetOffset()),
	}
}

// mapProductDTOToProto converts a product DTO to proto message.
func mapProductDTOToProto(dto *get_product.ProductDTO) *pb.Product {
	product := &pb.Product{
		Id:          dto.ID,
		Name:        dto.Name,
		Description: dto.Description,
		Category:    dto.Category,
		BasePrice: &pb.Money{
			Numerator:   dto.BasePriceNumerator,
			Denominator: dto.BasePriceDenominator,
		},
		EffectivePrice: &pb.Money{
			Numerator:   dto.EffectivePriceNum,
			Denominator: dto.EffectivePriceDenom,
		},
		Status:    dto.Status,
		CreatedAt: timestamppb.New(dto.CreatedAt),
		UpdatedAt: timestamppb.New(dto.UpdatedAt),
	}

	if dto.DiscountPercent != nil {
		product.Discount = &pb.Discount{
			Percentage: *dto.DiscountPercent,
		}
		if dto.DiscountStartDate != nil {
			product.Discount.StartDate = timestamppb.New(*dto.DiscountStartDate)
		}
		if dto.DiscountEndDate != nil {
			product.Discount.EndDate = timestamppb.New(*dto.DiscountEndDate)
		}
	}

	return product
}

// mapProductListItemDTOToProto converts a product list item DTO to proto message.
func mapProductListItemDTOToProto(dto *list_products.ProductListItemDTO) *pb.ProductListItem {
	item := &pb.ProductListItem{
		Id:          dto.ID,
		Name:        dto.Name,
		Description: dto.Description,
		Category:    dto.Category,
		BasePrice: &pb.Money{
			Numerator:   dto.BasePriceNumerator,
			Denominator: dto.BasePriceDenominator,
		},
		EffectivePrice: &pb.Money{
			Numerator:   dto.EffectivePriceNum,
			Denominator: dto.EffectivePriceDenom,
		},
		Status:    dto.Status,
		CreatedAt: timestamppb.New(dto.CreatedAt),
	}

	if dto.DiscountPercent != nil {
		item.DiscountPercent = dto.DiscountPercent
	}

	return item
}

// mapListResultToProto converts a list result DTO to proto response.
func mapListResultToProto(result *list_products.ListResultDTO) *pb.ListProductsReply {
	products := make([]*pb.ProductListItem, len(result.Products))
	for i, p := range result.Products {
		products[i] = mapProductListItemDTOToProto(p)
	}

	return &pb.ListProductsReply{
		Products:   products,
		TotalCount: result.TotalCount,
		HasMore:    result.HasMore,
	}
}
