package repo

import (
	"context"
	"fmt"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"

	"github.com/product-catalog-service/internal/app/product/contracts"
	"github.com/product-catalog-service/internal/app/product/domain"
	"github.com/product-catalog-service/internal/models/m_product"
	"github.com/product-catalog-service/internal/pkg/clock"
)

// ReadModelRepo implements ProductReadModelRepository for Spanner.
type ReadModelRepo struct {
	client *spanner.Client
	clock  clock.Clock
}

// NewReadModelRepo creates a new ReadModelRepo.
func NewReadModelRepo(client *spanner.Client, clock clock.Clock) *ReadModelRepo {
	return &ReadModelRepo{
		client: client,
		clock:  clock,
	}
}

// GetByID retrieves a product read model by ID.
func (r *ReadModelRepo) GetByID(ctx context.Context, id string) (*contracts.ProductReadModel, error) {
	row, err := r.client.Single().ReadRow(
		ctx,
		m_product.TableName,
		spanner.Key{id},
		m_product.AllColumns(),
	)
	if err != nil {
		if spanner.ErrCode(err) == 5 { // NotFound
			return nil, domain.ErrProductNotFound
		}
		return nil, err
	}

	return r.rowToReadModel(row)
}

// List retrieves a paginated list of products with optional filters.
func (r *ReadModelRepo) List(
	ctx context.Context,
	filters contracts.ProductListFilters,
	pagination contracts.Pagination,
) (*contracts.ProductListResult, error) {
	// Build query
	query := fmt.Sprintf("SELECT %s FROM %s WHERE 1=1",
		buildSelectColumns(),
		m_product.TableName,
	)

	params := make(map[string]interface{})

	// Apply filters
	if filters.ActiveOnly {
		query += fmt.Sprintf(" AND %s = @status", m_product.Status)
		params["status"] = string(domain.ProductStatusActive)
	} else if filters.Status != nil {
		query += fmt.Sprintf(" AND %s = @status", m_product.Status)
		params["status"] = *filters.Status
	}

	if filters.Category != nil && *filters.Category != "" {
		query += fmt.Sprintf(" AND %s = @category", m_product.Category)
		params["category"] = *filters.Category
	}

	// Exclude archived by default
	query += fmt.Sprintf(" AND %s != @archivedStatus", m_product.Status)
	params["archivedStatus"] = string(domain.ProductStatusArchived)

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE 1=1", m_product.TableName)
	if filters.ActiveOnly {
		countQuery += fmt.Sprintf(" AND %s = @status", m_product.Status)
	} else if filters.Status != nil {
		countQuery += fmt.Sprintf(" AND %s = @status", m_product.Status)
	}
	if filters.Category != nil && *filters.Category != "" {
		countQuery += fmt.Sprintf(" AND %s = @category", m_product.Category)
	}
	countQuery += fmt.Sprintf(" AND %s != @archivedStatus", m_product.Status)

	var totalCount int64
	countRow := r.client.Single().Query(ctx, spanner.Statement{
		SQL:    countQuery,
		Params: params,
	})
	defer countRow.Stop()

	countRowData, err := countRow.Next()
	if err != nil && err != iterator.Done {
		return nil, err
	}
	if countRowData != nil {
		if err := countRowData.Columns(&totalCount); err != nil {
			return nil, err
		}
	}

	// Add ordering and pagination
	query += fmt.Sprintf(" ORDER BY %s DESC", m_product.CreatedAt)
	query += " LIMIT @limit OFFSET @offset"
	params["limit"] = int64(pagination.Limit)
	params["offset"] = int64(pagination.Offset)

	stmt := spanner.Statement{
		SQL:    query,
		Params: params,
	}

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	products := make([]*contracts.ProductReadModel, 0)

	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		product, err := r.rowToReadModel(row)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	hasMore := int64(pagination.Offset+len(products)) < totalCount

	return &contracts.ProductListResult{
		Products:   products,
		TotalCount: totalCount,
		HasMore:    hasMore,
	}, nil
}

// CountByCategory counts products in a category.
func (r *ReadModelRepo) CountByCategory(ctx context.Context, category string) (int64, error) {
	query := fmt.Sprintf(
		"SELECT COUNT(*) FROM %s WHERE %s = @category AND %s != @archivedStatus",
		m_product.TableName,
		m_product.Category,
		m_product.Status,
	)

	stmt := spanner.Statement{
		SQL: query,
		Params: map[string]interface{}{
			"category":       category,
			"archivedStatus": string(domain.ProductStatusArchived),
		},
	}

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err != nil {
		return 0, err
	}

	var count int64
	if err := row.Columns(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (r *ReadModelRepo) rowToReadModel(row *spanner.Row) (*contracts.ProductReadModel, error) {
	var dbProduct m_product.Product

	err := row.Columns(
		&dbProduct.ProductID,
		&dbProduct.Name,
		&dbProduct.Description,
		&dbProduct.Category,
		&dbProduct.BasePriceNumerator,
		&dbProduct.BasePriceDenominator,
		&dbProduct.DiscountPercent,
		&dbProduct.DiscountStartDate,
		&dbProduct.DiscountEndDate,
		&dbProduct.Status,
		&dbProduct.CreatedAt,
		&dbProduct.UpdatedAt,
		&dbProduct.ArchivedAt,
	)
	if err != nil {
		return nil, err
	}

	readModel := &contracts.ProductReadModel{
		ID:                   dbProduct.ProductID,
		Name:                 dbProduct.Name,
		Description:          dbProduct.Description,
		Category:             dbProduct.Category,
		BasePriceNumerator:   dbProduct.BasePriceNumerator,
		BasePriceDenominator: dbProduct.BasePriceDenominator,
		Status:               dbProduct.Status,
		CreatedAt:            dbProduct.CreatedAt,
		UpdatedAt:            dbProduct.UpdatedAt,
	}

	// Calculate effective price
	readModel.EffectivePriceNum = dbProduct.BasePriceNumerator
	readModel.EffectivePriceDenom = dbProduct.BasePriceDenominator

	if dbProduct.DiscountPercent.Valid {
		percentage, _ := dbProduct.DiscountPercent.Numeric.Float64()
		pct := int64(percentage)
		readModel.DiscountPercent = &pct

		// Check if discount is active
		now := r.clock.Now()
		if dbProduct.DiscountStartDate.Valid && dbProduct.DiscountEndDate.Valid {
			startDate := dbProduct.DiscountStartDate.Time
			endDate := dbProduct.DiscountEndDate.Time
			readModel.DiscountStartDate = &startDate
			readModel.DiscountEndDate = &endDate

			if !now.Before(startDate) && !now.After(endDate) {
				// Apply discount: effective = base * (100 - percent) / 100
				readModel.EffectivePriceNum = dbProduct.BasePriceNumerator * (100 - pct)
				readModel.EffectivePriceDenom = dbProduct.BasePriceDenominator * 100
			}
		}
	}

	if dbProduct.ArchivedAt.Valid {
		readModel.ArchivedAt = &dbProduct.ArchivedAt.Time
	}

	return readModel, nil
}

func buildSelectColumns() string {
	columns := m_product.AllColumns()
	result := ""
	for i, col := range columns {
		if i > 0 {
			result += ", "
		}
		result += col
	}
	return result
}
