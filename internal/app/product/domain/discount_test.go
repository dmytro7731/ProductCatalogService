package domain_test

import (
	"testing"
	"time"

	"github.com/product-catalog-service/internal/app/product/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDiscount(t *testing.T) {
	now := time.Now()
	startDate := now
	endDate := now.Add(7 * 24 * time.Hour) // 1 week

	tests := []struct {
		name       string
		percentage int64
		startDate  time.Time
		endDate    time.Time
		wantErr    error
	}{
		{
			name:       "valid discount",
			percentage: 20,
			startDate:  startDate,
			endDate:    endDate,
			wantErr:    nil,
		},
		{
			name:       "100% discount",
			percentage: 100,
			startDate:  startDate,
			endDate:    endDate,
			wantErr:    nil,
		},
		{
			name:       "zero percentage",
			percentage: 0,
			startDate:  startDate,
			endDate:    endDate,
			wantErr:    domain.ErrInvalidDiscountPercentage,
		},
		{
			name:       "negative percentage",
			percentage: -10,
			startDate:  startDate,
			endDate:    endDate,
			wantErr:    domain.ErrInvalidDiscountPercentage,
		},
		{
			name:       "percentage over 100",
			percentage: 101,
			startDate:  startDate,
			endDate:    endDate,
			wantErr:    domain.ErrInvalidDiscountPercentage,
		},
		{
			name:       "end date before start date",
			percentage: 20,
			startDate:  endDate,
			endDate:    startDate,
			wantErr:    domain.ErrInvalidDiscountPeriod,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			discount, err := domain.NewDiscount(tt.percentage, tt.startDate, tt.endDate)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, discount)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, discount)
				assert.Equal(t, tt.percentage, discount.Percentage())
			}
		})
	}
}

func TestDiscount_IsValidAt(t *testing.T) {
	startDate := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 2, 28, 23, 59, 59, 0, time.UTC)

	discount, err := domain.NewDiscount(20, startDate, endDate)
	require.NoError(t, err)

	tests := []struct {
		name     string
		checkAt  time.Time
		expected bool
	}{
		{
			name:     "before start",
			checkAt:  time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC),
			expected: false,
		},
		{
			name:     "at start",
			checkAt:  startDate,
			expected: true,
		},
		{
			name:     "during period",
			checkAt:  time.Date(2026, 2, 15, 12, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "at end",
			checkAt:  endDate,
			expected: true,
		},
		{
			name:     "after end",
			checkAt:  time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, discount.IsValidAt(tt.checkAt))
		})
	}
}

func TestDiscount_Apply(t *testing.T) {
	discount, err := domain.NewDiscount(
		20,
		time.Now(),
		time.Now().Add(24*time.Hour),
	)
	require.NoError(t, err)

	price, err := domain.NewMoney(10000, 100) // $100.00
	require.NoError(t, err)

	discountedPrice := discount.Apply(price)

	// $100.00 - 20% = $80.00
	assert.Equal(t, "80.00", discountedPrice.String())
}

func TestDiscount_IsExpired(t *testing.T) {
	now := time.Now()
	pastDiscount, _ := domain.NewDiscount(20, now.Add(-48*time.Hour), now.Add(-24*time.Hour))
	futureDiscount, _ := domain.NewDiscount(20, now.Add(24*time.Hour), now.Add(48*time.Hour))

	assert.True(t, pastDiscount.IsExpired(now))
	assert.False(t, futureDiscount.IsExpired(now))
}
