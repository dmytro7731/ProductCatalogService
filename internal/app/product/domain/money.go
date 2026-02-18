package domain

import (
	"fmt"
	"math/big"
)

// Money represents a monetary value with precise decimal arithmetic.
// It uses big.Rat internally to avoid floating-point precision issues.
type Money struct {
	amount *big.Rat
}

// NewMoney creates a new Money value from numerator and denominator.
// For example, $19.99 would be NewMoney(1999, 100).
func NewMoney(numerator, denominator int64) (*Money, error) {
	if denominator == 0 {
		return nil, ErrInvalidMoney
	}
	if numerator < 0 {
		return nil, ErrNegativeMoney
	}

	return &Money{
		amount: big.NewRat(numerator, denominator),
	}, nil
}

// NewMoneyFromRat creates Money from an existing big.Rat.
func NewMoneyFromRat(amount *big.Rat) (*Money, error) {
	if amount == nil {
		return nil, ErrInvalidMoney
	}
	if amount.Sign() < 0 {
		return nil, ErrNegativeMoney
	}

	return &Money{
		amount: new(big.Rat).Set(amount),
	}, nil
}

// Zero returns a Money value of zero.
func Zero() *Money {
	return &Money{
		amount: big.NewRat(0, 1),
	}
}

// Amount returns a copy of the underlying big.Rat.
func (m *Money) Amount() *big.Rat {
	return new(big.Rat).Set(m.amount)
}

// Numerator returns the numerator of the money value.
func (m *Money) Numerator() int64 {
	return m.amount.Num().Int64()
}

// Denominator returns the denominator of the money value.
func (m *Money) Denominator() int64 {
	return m.amount.Denom().Int64()
}

// Add returns a new Money that is the sum of m and other.
func (m *Money) Add(other *Money) *Money {
	result := new(big.Rat).Add(m.amount, other.amount)
	return &Money{amount: result}
}

// Subtract returns a new Money that is the difference of m and other.
// Returns error if result would be negative.
func (m *Money) Subtract(other *Money) (*Money, error) {
	result := new(big.Rat).Sub(m.amount, other.amount)
	if result.Sign() < 0 {
		return nil, ErrNegativeMoney
	}
	return &Money{amount: result}, nil
}

// Multiply returns a new Money multiplied by the given factor.
func (m *Money) Multiply(factor *big.Rat) (*Money, error) {
	if factor.Sign() < 0 {
		return nil, ErrNegativeMoney
	}
	result := new(big.Rat).Mul(m.amount, factor)
	return &Money{amount: result}, nil
}

// ApplyPercentage returns a new Money after applying a percentage.
// percentage should be 0-100 (e.g., 20 for 20%).
func (m *Money) ApplyPercentage(percentage int64) *Money {
	factor := big.NewRat(percentage, 100)
	result := new(big.Rat).Mul(m.amount, factor)
	return &Money{amount: result}
}

// SubtractPercentage returns the money after subtracting a percentage.
// percentage should be 0-100 (e.g., 20 for 20% off).
func (m *Money) SubtractPercentage(percentage int64) *Money {
	discount := m.ApplyPercentage(percentage)
	result := new(big.Rat).Sub(m.amount, discount.amount)
	return &Money{amount: result}
}

// IsZero returns true if the money value is zero.
func (m *Money) IsZero() bool {
	return m.amount.Sign() == 0
}

// IsPositive returns true if the money value is greater than zero.
func (m *Money) IsPositive() bool {
	return m.amount.Sign() > 0
}

// Equals returns true if two money values are equal.
func (m *Money) Equals(other *Money) bool {
	if other == nil {
		return false
	}
	return m.amount.Cmp(other.amount) == 0
}

// GreaterThan returns true if m is greater than other.
func (m *Money) GreaterThan(other *Money) bool {
	return m.amount.Cmp(other.amount) > 0
}

// LessThan returns true if m is less than other.
func (m *Money) LessThan(other *Money) bool {
	return m.amount.Cmp(other.amount) < 0
}

// String returns a human-readable representation of the money value.
func (m *Money) String() string {
	f, _ := m.amount.Float64()
	return fmt.Sprintf("%.2f", f)
}
