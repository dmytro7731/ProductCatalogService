package domain

import (
	"time"
)

// Discount represents a percentage-based discount with validity period.
type Discount struct {
	percentage int64
	startDate  time.Time
	endDate    time.Time
}

// NewDiscount creates a new Discount value object.
// percentage should be 1-100.
func NewDiscount(percentage int64, startDate, endDate time.Time) (*Discount, error) {
	if percentage <= 0 || percentage > 100 {
		return nil, ErrInvalidDiscountPercentage
	}

	if endDate.Before(startDate) {
		return nil, ErrInvalidDiscountPeriod
	}

	return &Discount{
		percentage: percentage,
		startDate:  startDate,
		endDate:    endDate,
	}, nil
}

// Percentage returns the discount percentage (1-100).
func (d *Discount) Percentage() int64 {
	return d.percentage
}

// StartDate returns the discount start date.
func (d *Discount) StartDate() time.Time {
	return d.startDate
}

// EndDate returns the discount end date.
func (d *Discount) EndDate() time.Time {
	return d.endDate
}

// IsValidAt checks if the discount is valid at the given time.
func (d *Discount) IsValidAt(t time.Time) bool {
	return !t.Before(d.startDate) && !t.After(d.endDate)
}

// IsExpired checks if the discount has expired.
func (d *Discount) IsExpired(now time.Time) bool {
	return now.After(d.endDate)
}

// HasStarted checks if the discount period has started.
func (d *Discount) HasStarted(now time.Time) bool {
	return !now.Before(d.startDate)
}

// Apply applies the discount to the given money and returns the discounted price.
func (d *Discount) Apply(price *Money) *Money {
	return price.SubtractPercentage(d.percentage)
}

// Equals checks if two discounts are equal.
func (d *Discount) Equals(other *Discount) bool {
	if other == nil {
		return false
	}
	return d.percentage == other.percentage &&
		d.startDate.Equal(other.startDate) &&
		d.endDate.Equal(other.endDate)
}
