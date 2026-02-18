package get_product

import (
	"time"
)

// ProductDTO represents a product for query responses.
type ProductDTO struct {
	ID                   string
	Name                 string
	Description          string
	Category             string
	BasePriceNumerator   int64
	BasePriceDenominator int64
	EffectivePriceNum    int64
	EffectivePriceDenom  int64
	DiscountPercent      *int64
	DiscountStartDate    *time.Time
	DiscountEndDate      *time.Time
	Status               string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// HasActiveDiscount returns true if the product has an active discount.
func (p *ProductDTO) HasActiveDiscount() bool {
	return p.DiscountPercent != nil && *p.DiscountPercent > 0
}

// BasePriceFloat returns the base price as a float64.
func (p *ProductDTO) BasePriceFloat() float64 {
	if p.BasePriceDenominator == 0 {
		return 0
	}
	return float64(p.BasePriceNumerator) / float64(p.BasePriceDenominator)
}

// EffectivePriceFloat returns the effective price as a float64.
func (p *ProductDTO) EffectivePriceFloat() float64 {
	if p.EffectivePriceDenom == 0 {
		return 0
	}
	return float64(p.EffectivePriceNum) / float64(p.EffectivePriceDenom)
}
