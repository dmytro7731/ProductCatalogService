package services

import (
	"math/big"
	"time"

	"github.com/product-catalog-service/internal/app/product/domain"
)

// PricingCalculator is a domain service for complex pricing calculations.
type PricingCalculator struct{}

// NewPricingCalculator creates a new pricing calculator.
func NewPricingCalculator() *PricingCalculator {
	return &PricingCalculator{}
}

// CalculateEffectivePrice calculates the effective price of a product at a given time.
func (pc *PricingCalculator) CalculateEffectivePrice(product *domain.Product, now time.Time) *domain.Money {
	return product.EffectivePrice(now)
}

// CalculateDiscountAmount calculates the discount amount in money.
func (pc *PricingCalculator) CalculateDiscountAmount(basePrice *domain.Money, discount *domain.Discount) *domain.Money {
	if discount == nil {
		return domain.Zero()
	}
	return basePrice.ApplyPercentage(discount.Percentage())
}

// CalculateSavings calculates savings when buying multiple items with discount.
func (pc *PricingCalculator) CalculateSavings(basePrice *domain.Money, discount *domain.Discount, quantity int64) *domain.Money {
	if discount == nil || quantity <= 0 {
		return domain.Zero()
	}

	singleDiscount := pc.CalculateDiscountAmount(basePrice, discount)
	factor := big.NewRat(quantity, 1)
	total, _ := singleDiscount.Multiply(factor)
	return total
}

// PriceBreakdown contains a detailed breakdown of product pricing.
type PriceBreakdown struct {
	BasePrice       *domain.Money
	DiscountPercent int64
	DiscountAmount  *domain.Money
	EffectivePrice  *domain.Money
	HasDiscount     bool
}

// GetPriceBreakdown returns a detailed price breakdown for a product.
func (pc *PricingCalculator) GetPriceBreakdown(product *domain.Product, now time.Time) *PriceBreakdown {
	breakdown := &PriceBreakdown{
		BasePrice:      product.BasePrice(),
		EffectivePrice: product.EffectivePrice(now),
		HasDiscount:    product.HasActiveDiscount(now),
	}

	if product.HasActiveDiscount(now) {
		discount := product.Discount()
		breakdown.DiscountPercent = discount.Percentage()
		breakdown.DiscountAmount = pc.CalculateDiscountAmount(product.BasePrice(), discount)
	} else {
		breakdown.DiscountPercent = 0
		breakdown.DiscountAmount = domain.Zero()
	}

	return breakdown
}

// ValidateDiscountApplication checks if a discount can be applied.
func (pc *PricingCalculator) ValidateDiscountApplication(
	product *domain.Product,
	discount *domain.Discount,
	now time.Time,
) error {
	if !product.IsActive() {
		return domain.ErrProductNotActive
	}

	if discount.IsExpired(now) {
		return domain.ErrDiscountExpired
	}

	return nil
}
