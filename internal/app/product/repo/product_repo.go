package repo

import (
	"context"
	"math/big"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/product-catalog-service/internal/app/product/domain"
	"github.com/product-catalog-service/internal/models/m_product"
)

// ProductRepo implements the ProductRepository interface for Spanner.
type ProductRepo struct {
	client *spanner.Client
	model  *m_product.Model
}

// NewProductRepo creates a new ProductRepo.
func NewProductRepo(client *spanner.Client) *ProductRepo {
	return &ProductRepo{
		client: client,
		model:  m_product.NewModel(),
	}
}

// GetByID retrieves a product by its ID.
func (r *ProductRepo) GetByID(ctx context.Context, id string) (*domain.Product, error) {
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

	return r.rowToProduct(row)
}

// GetByIDWithTxn retrieves a product within a transaction.
func (r *ProductRepo) GetByIDWithTxn(ctx context.Context, txn *spanner.ReadWriteTransaction, id string) (*domain.Product, error) {
	row, err := txn.ReadRow(
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

	return r.rowToProduct(row)
}

// InsertMut returns a mutation for inserting a new product.
func (r *ProductRepo) InsertMut(product *domain.Product) *spanner.Mutation {
	if !product.IsNew() {
		return nil
	}

	dbProduct := r.productToDBModel(product)
	return r.model.InsertMut(dbProduct)
}

// UpdateMut returns a mutation for updating an existing product.
// Only includes fields that have been modified.
func (r *ProductRepo) UpdateMut(product *domain.Product) *spanner.Mutation {
	if product.IsNew() {
		return nil
	}

	changes := product.Changes()
	if !changes.HasChanges() {
		return nil
	}

	updates := make(map[string]interface{})

	if changes.Dirty(domain.FieldName) {
		updates[m_product.Name] = product.Name()
	}

	if changes.Dirty(domain.FieldDescription) {
		updates[m_product.Description] = product.Description()
	}

	if changes.Dirty(domain.FieldCategory) {
		updates[m_product.Category] = product.Category()
	}

	if changes.Dirty(domain.FieldStatus) {
		updates[m_product.Status] = string(product.Status())
	}

	if changes.Dirty(domain.FieldDiscount) {
		if d := product.Discount(); d != nil {
			updates[m_product.DiscountPercent] = spanner.NullNumeric{
				Numeric: *big.NewRat(d.Percentage(), 1),
				Valid:   true,
			}
			updates[m_product.DiscountStartDate] = spanner.NullTime{
				Time:  d.StartDate(),
				Valid: true,
			}
			updates[m_product.DiscountEndDate] = spanner.NullTime{
				Time:  d.EndDate(),
				Valid: true,
			}
		} else {
			updates[m_product.DiscountPercent] = spanner.NullNumeric{Valid: false}
			updates[m_product.DiscountStartDate] = spanner.NullTime{Valid: false}
			updates[m_product.DiscountEndDate] = spanner.NullTime{Valid: false}
		}
	}

	if changes.Dirty(domain.FieldArchivedAt) {
		if archivedAt := product.ArchivedAt(); archivedAt != nil {
			updates[m_product.ArchivedAt] = spanner.NullTime{
				Time:  *archivedAt,
				Valid: true,
			}
		} else {
			updates[m_product.ArchivedAt] = spanner.NullTime{Valid: false}
		}
	}

	if len(updates) == 0 {
		return nil
	}

	updates[m_product.UpdatedAt] = product.UpdatedAt()

	return r.model.UpdateMut(product.ID(), updates)
}

func (r *ProductRepo) productToDBModel(p *domain.Product) *m_product.Product {
	dbProduct := &m_product.Product{
		ProductID:            p.ID(),
		Name:                 p.Name(),
		Description:          p.Description(),
		Category:             p.Category(),
		BasePriceNumerator:   p.BasePrice().Numerator(),
		BasePriceDenominator: p.BasePrice().Denominator(),
		Status:               string(p.Status()),
		CreatedAt:            p.CreatedAt(),
		UpdatedAt:            p.UpdatedAt(),
	}

	if d := p.Discount(); d != nil {
		dbProduct.DiscountPercent = spanner.NullNumeric{
			Numeric: *big.NewRat(d.Percentage(), 1),
			Valid:   true,
		}
		dbProduct.DiscountStartDate = spanner.NullTime{
			Time:  d.StartDate(),
			Valid: true,
		}
		dbProduct.DiscountEndDate = spanner.NullTime{
			Time:  d.EndDate(),
			Valid: true,
		}
	}

	if archivedAt := p.ArchivedAt(); archivedAt != nil {
		dbProduct.ArchivedAt = spanner.NullTime{
			Time:  *archivedAt,
			Valid: true,
		}
	}

	return dbProduct
}

func (r *ProductRepo) rowToProduct(row *spanner.Row) (*domain.Product, error) {
	var (
		productID            string
		name                 string
		description          string
		category             string
		basePriceNumerator   int64
		basePriceDenominator int64
		discountPercent      spanner.NullNumeric
		discountStartDate    spanner.NullTime
		discountEndDate      spanner.NullTime
		status               string
		createdAt            time.Time
		updatedAt            time.Time
		archivedAt           spanner.NullTime
	)

	err := row.Columns(
		&productID,
		&name,
		&description,
		&category,
		&basePriceNumerator,
		&basePriceDenominator,
		&discountPercent,
		&discountStartDate,
		&discountEndDate,
		&status,
		&createdAt,
		&updatedAt,
		&archivedAt,
	)
	if err != nil {
		return nil, err
	}

	basePrice, err := domain.NewMoney(basePriceNumerator, basePriceDenominator)
	if err != nil {
		return nil, err
	}

	var discount *domain.Discount
	if discountPercent.Valid && discountStartDate.Valid && discountEndDate.Valid {
		percentage, _ := discountPercent.Numeric.Float64()
		discount, err = domain.NewDiscount(int64(percentage), discountStartDate.Time, discountEndDate.Time)
		if err != nil {
			return nil, err
		}
	}

	var archivedAtPtr *time.Time
	if archivedAt.Valid {
		archivedAtPtr = &archivedAt.Time
	}

	return domain.Reconstitute(
		productID,
		name,
		description,
		category,
		basePrice,
		discount,
		domain.ProductStatus(status),
		createdAt,
		updatedAt,
		archivedAtPtr,
	), nil
}
