package m_product

// Table name
const TableName = "products"

// Column names for the products table.
const (
	ProductID            = "product_id"
	Name                 = "name"
	Description          = "description"
	Category             = "category"
	BasePriceNumerator   = "base_price_numerator"
	BasePriceDenominator = "base_price_denominator"
	DiscountPercent      = "discount_percent"
	DiscountStartDate    = "discount_start_date"
	DiscountEndDate      = "discount_end_date"
	Status               = "status"
	CreatedAt            = "created_at"
	UpdatedAt            = "updated_at"
	ArchivedAt           = "archived_at"
)

// AllColumns returns all column names.
func AllColumns() []string {
	return []string{
		ProductID,
		Name,
		Description,
		Category,
		BasePriceNumerator,
		BasePriceDenominator,
		DiscountPercent,
		DiscountStartDate,
		DiscountEndDate,
		Status,
		CreatedAt,
		UpdatedAt,
		ArchivedAt,
	}
}

// InsertColumns returns columns used for insert operations.
func InsertColumns() []string {
	return []string{
		ProductID,
		Name,
		Description,
		Category,
		BasePriceNumerator,
		BasePriceDenominator,
		DiscountPercent,
		DiscountStartDate,
		DiscountEndDate,
		Status,
		CreatedAt,
		UpdatedAt,
		ArchivedAt,
	}
}
