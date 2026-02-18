package domain_test

import (
	"math/big"
	"testing"

	"github.com/product-catalog-service/internal/app/product/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMoney(t *testing.T) {
	tests := []struct {
		name        string
		numerator   int64
		denominator int64
		wantErr     error
	}{
		{
			name:        "valid money",
			numerator:   1999,
			denominator: 100,
			wantErr:     nil,
		},
		{
			name:        "zero numerator",
			numerator:   0,
			denominator: 100,
			wantErr:     nil,
		},
		{
			name:        "zero denominator",
			numerator:   1999,
			denominator: 0,
			wantErr:     domain.ErrInvalidMoney,
		},
		{
			name:        "negative numerator",
			numerator:   -1999,
			denominator: 100,
			wantErr:     domain.ErrNegativeMoney,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			money, err := domain.NewMoney(tt.numerator, tt.denominator)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, money)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, money)
				// Note: big.Rat normalizes fractions, so we check the value is correct
				// rather than exact numerator/denominator (e.g., 0/100 becomes 0/1)
				expectedRat := big.NewRat(tt.numerator, tt.denominator)
				assert.True(t, money.Amount().Cmp(expectedRat) == 0,
					"expected %v, got %v", expectedRat, money.Amount())
			}
		})
	}
}

func TestMoney_Add(t *testing.T) {
	m1, _ := domain.NewMoney(1000, 100) // $10.00
	m2, _ := domain.NewMoney(500, 100)  // $5.00

	result := m1.Add(m2)

	// $10.00 + $5.00 = $15.00
	expected := big.NewRat(1500, 100)
	assert.True(t, result.Amount().Cmp(expected) == 0)
}

func TestMoney_Subtract(t *testing.T) {
	m1, _ := domain.NewMoney(1000, 100) // $10.00
	m2, _ := domain.NewMoney(500, 100)  // $5.00

	result, err := m1.Subtract(m2)
	require.NoError(t, err)

	// $10.00 - $5.00 = $5.00
	expected := big.NewRat(500, 100)
	assert.True(t, result.Amount().Cmp(expected) == 0)
}

func TestMoney_Subtract_Negative(t *testing.T) {
	m1, _ := domain.NewMoney(500, 100)  // $5.00
	m2, _ := domain.NewMoney(1000, 100) // $10.00

	_, err := m1.Subtract(m2)
	assert.ErrorIs(t, err, domain.ErrNegativeMoney)
}

func TestMoney_ApplyPercentage(t *testing.T) {
	m, _ := domain.NewMoney(10000, 100) // $100.00

	result := m.ApplyPercentage(20) // 20%

	// $100.00 * 20% = $20.00
	expected := big.NewRat(2000, 100)
	assert.True(t, result.Amount().Cmp(expected) == 0)
}

func TestMoney_SubtractPercentage(t *testing.T) {
	m, _ := domain.NewMoney(10000, 100) // $100.00

	result := m.SubtractPercentage(20) // 20% off

	// $100.00 - 20% = $80.00
	expected := big.NewRat(8000, 100)
	assert.True(t, result.Amount().Cmp(expected) == 0)
}

func TestMoney_Comparison(t *testing.T) {
	m1, _ := domain.NewMoney(1000, 100) // $10.00
	m2, _ := domain.NewMoney(500, 100)  // $5.00
	m3, _ := domain.NewMoney(1000, 100) // $10.00

	assert.True(t, m1.GreaterThan(m2))
	assert.True(t, m2.LessThan(m1))
	assert.True(t, m1.Equals(m3))
	assert.False(t, m1.Equals(m2))
}

func TestMoney_IsZero(t *testing.T) {
	zero := domain.Zero()
	nonZero, _ := domain.NewMoney(100, 100)

	assert.True(t, zero.IsZero())
	assert.False(t, nonZero.IsZero())
}

func TestMoney_String(t *testing.T) {
	m, _ := domain.NewMoney(1999, 100) // $19.99
	assert.Equal(t, "19.99", m.String())
}
