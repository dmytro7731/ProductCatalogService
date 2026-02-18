package list_products

import (
	"time"
)

// ProductListItemDTO represents a product item in a list.
type ProductListItemDTO struct {
	ID                   string
	Name                 string
	Description          string
	Category             string
	BasePriceNumerator   int64
	BasePriceDenominator int64
	EffectivePriceNum    int64
	EffectivePriceDenom  int64
	DiscountPercent      *int64
	Status               string
	CreatedAt            time.Time
}

// HasActiveDiscount returns true if the product has an active discount.
func (p *ProductListItemDTO) HasActiveDiscount() bool {
	return p.DiscountPercent != nil && *p.DiscountPercent > 0
}

// BasePriceFloat returns the base price as a float64.
func (p *ProductListItemDTO) BasePriceFloat() float64 {
	if p.BasePriceDenominator == 0 {
		return 0
	}
	return float64(p.BasePriceNumerator) / float64(p.BasePriceDenominator)
}

// EffectivePriceFloat returns the effective price as a float64.
func (p *ProductListItemDTO) EffectivePriceFloat() float64 {
	if p.EffectivePriceDenom == 0 {
		return 0
	}
	return float64(p.EffectivePriceNum) / float64(p.EffectivePriceDenom)
}

// ListResultDTO represents the result of a product list query.
type ListResultDTO struct {
	Products   []*ProductListItemDTO
	TotalCount int64
	HasMore    bool
}
