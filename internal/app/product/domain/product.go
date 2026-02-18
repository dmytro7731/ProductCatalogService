package domain

import (
	"time"
)

// Field constants for change tracking.
const (
	FieldName        = "name"
	FieldDescription = "description"
	FieldCategory    = "category"
	FieldBasePrice   = "base_price"
	FieldDiscount    = "discount"
	FieldStatus      = "status"
	FieldArchivedAt  = "archived_at"
)

// ProductStatus represents the status of a product.
type ProductStatus string

const (
	ProductStatusDraft    ProductStatus = "draft"
	ProductStatusActive   ProductStatus = "active"
	ProductStatusInactive ProductStatus = "inactive"
	ProductStatusArchived ProductStatus = "archived"
)

// IsValid checks if the status is a valid ProductStatus.
func (s ProductStatus) IsValid() bool {
	switch s {
	case ProductStatusDraft, ProductStatusActive, ProductStatusInactive, ProductStatusArchived:
		return true
	}
	return false
}

// String returns the string representation of the status.
func (s ProductStatus) String() string {
	return string(s)
}

// ChangeTracker tracks which fields have been modified.
type ChangeTracker struct {
	dirtyFields map[string]bool
}

// NewChangeTracker creates a new change tracker.
func NewChangeTracker() *ChangeTracker {
	return &ChangeTracker{
		dirtyFields: make(map[string]bool),
	}
}

// MarkDirty marks a field as dirty (modified).
func (ct *ChangeTracker) MarkDirty(field string) {
	ct.dirtyFields[field] = true
}

// Dirty returns true if the field has been modified.
func (ct *ChangeTracker) Dirty(field string) bool {
	return ct.dirtyFields[field]
}

// HasChanges returns true if any field has been modified.
func (ct *ChangeTracker) HasChanges() bool {
	return len(ct.dirtyFields) > 0
}

// DirtyFields returns a list of all dirty fields.
func (ct *ChangeTracker) DirtyFields() []string {
	fields := make([]string, 0, len(ct.dirtyFields))
	for field := range ct.dirtyFields {
		fields = append(fields, field)
	}
	return fields
}

// Reset clears all tracked changes.
func (ct *ChangeTracker) Reset() {
	ct.dirtyFields = make(map[string]bool)
}

// Product is the aggregate root for the product domain.
type Product struct {
	id          string
	name        string
	description string
	category    string
	basePrice   *Money
	discount    *Discount
	status      ProductStatus
	createdAt   time.Time
	updatedAt   time.Time
	archivedAt  *time.Time

	changes *ChangeTracker
	events  []DomainEvent
	isNew   bool
}

// NewProduct creates a new product in draft status.
func NewProduct(id, name, description, category string, basePrice *Money, now time.Time) (*Product, error) {
	if name == "" {
		return nil, ErrEmptyProductName
	}
	if len(name) > MaxProductNameLength {
		return nil, ErrProductNameTooLong
	}
	if category == "" {
		return nil, ErrEmptyCategory
	}
	if len(category) > MaxCategoryLength {
		return nil, ErrCategoryTooLong
	}
	if basePrice == nil || basePrice.IsZero() {
		return nil, ErrZeroPrice
	}

	p := &Product{
		id:          id,
		name:        name,
		description: description,
		category:    category,
		basePrice:   basePrice,
		status:      ProductStatusDraft,
		createdAt:   now,
		updatedAt:   now,
		changes:     NewChangeTracker(),
		events:      make([]DomainEvent, 0),
		isNew:       true,
	}

	p.events = append(p.events, NewProductCreatedEvent(id, name, description, category, basePrice, now))

	return p, nil
}

// Reconstitute recreates a product from persistence without triggering events.
func Reconstitute(
	id, name, description, category string,
	basePrice *Money,
	discount *Discount,
	status ProductStatus,
	createdAt, updatedAt time.Time,
	archivedAt *time.Time,
) *Product {
	return &Product{
		id:          id,
		name:        name,
		description: description,
		category:    category,
		basePrice:   basePrice,
		discount:    discount,
		status:      status,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		archivedAt:  archivedAt,
		changes:     NewChangeTracker(),
		events:      make([]DomainEvent, 0),
		isNew:       false,
	}
}

// ID returns the product ID.
func (p *Product) ID() string {
	return p.id
}

// Name returns the product name.
func (p *Product) Name() string {
	return p.name
}

// Description returns the product description.
func (p *Product) Description() string {
	return p.description
}

// Category returns the product category.
func (p *Product) Category() string {
	return p.category
}

// BasePrice returns the base price.
func (p *Product) BasePrice() *Money {
	return p.basePrice
}

// Discount returns the current discount (may be nil).
func (p *Product) Discount() *Discount {
	return p.discount
}

// Status returns the product status.
func (p *Product) Status() ProductStatus {
	return p.status
}

// CreatedAt returns the creation timestamp.
func (p *Product) CreatedAt() time.Time {
	return p.createdAt
}

// UpdatedAt returns the last update timestamp.
func (p *Product) UpdatedAt() time.Time {
	return p.updatedAt
}

// ArchivedAt returns the archive timestamp (nil if not archived).
func (p *Product) ArchivedAt() *time.Time {
	return p.archivedAt
}

// IsNew returns true if this is a new product that hasn't been persisted.
func (p *Product) IsNew() bool {
	return p.isNew
}

// Changes returns the change tracker for this product.
func (p *Product) Changes() *ChangeTracker {
	return p.changes
}

// DomainEvents returns all domain events captured by this aggregate.
func (p *Product) DomainEvents() []DomainEvent {
	return p.events
}

// ClearEvents clears all captured domain events.
func (p *Product) ClearEvents() {
	p.events = make([]DomainEvent, 0)
}

// IsActive returns true if the product is active.
func (p *Product) IsActive() bool {
	return p.status == ProductStatusActive
}

// IsArchived returns true if the product is archived.
func (p *Product) IsArchived() bool {
	return p.status == ProductStatusArchived
}

// EffectivePrice calculates the current effective price considering any active discount.
func (p *Product) EffectivePrice(now time.Time) *Money {
	if p.discount == nil || !p.discount.IsValidAt(now) {
		return p.basePrice
	}
	return p.discount.Apply(p.basePrice)
}

// HasActiveDiscount returns true if the product has an active discount at the given time.
func (p *Product) HasActiveDiscount(now time.Time) bool {
	return p.discount != nil && p.discount.IsValidAt(now)
}

// Update updates the product details.
func (p *Product) Update(name, description, category string, now time.Time) error {
	if p.IsArchived() {
		return ErrCannotUpdateArchived
	}

	if name == "" {
		return ErrEmptyProductName
	}
	if len(name) > MaxProductNameLength {
		return ErrProductNameTooLong
	}
	if category == "" {
		return ErrEmptyCategory
	}
	if len(category) > MaxCategoryLength {
		return ErrCategoryTooLong
	}

	changed := false

	if p.name != name {
		p.name = name
		p.changes.MarkDirty(FieldName)
		changed = true
	}

	if p.description != description {
		p.description = description
		p.changes.MarkDirty(FieldDescription)
		changed = true
	}

	if p.category != category {
		p.category = category
		p.changes.MarkDirty(FieldCategory)
		changed = true
	}

	if changed {
		p.updatedAt = now
		p.events = append(p.events, NewProductUpdatedEvent(p.id, name, description, category, now))
	}

	return nil
}

// Activate activates the product.
func (p *Product) Activate(now time.Time) error {
	if p.IsArchived() {
		return ErrCannotActivateArchived
	}
	if p.IsActive() {
		return ErrProductAlreadyActive
	}

	p.status = ProductStatusActive
	p.updatedAt = now
	p.changes.MarkDirty(FieldStatus)
	p.events = append(p.events, NewProductActivatedEvent(p.id, now))

	return nil
}

// Deactivate deactivates the product.
func (p *Product) Deactivate(now time.Time) error {
	if p.IsArchived() {
		return ErrCannotDeactivateArchived
	}
	if p.status == ProductStatusInactive {
		return ErrProductInactive
	}

	p.status = ProductStatusInactive
	p.updatedAt = now
	p.changes.MarkDirty(FieldStatus)
	p.events = append(p.events, NewProductDeactivatedEvent(p.id, now))

	return nil
}

// Archive archives (soft deletes) the product.
func (p *Product) Archive(now time.Time) error {
	if p.IsArchived() {
		return ErrProductArchived
	}
	if p.IsActive() {
		return ErrCannotArchiveActive
	}

	p.status = ProductStatusArchived
	p.archivedAt = &now
	p.updatedAt = now
	p.changes.MarkDirty(FieldStatus)
	p.changes.MarkDirty(FieldArchivedAt)
	p.events = append(p.events, NewProductArchivedEvent(p.id, now))

	return nil
}

// ApplyDiscount applies a discount to the product.
func (p *Product) ApplyDiscount(discount *Discount, now time.Time) error {
	if !p.IsActive() {
		return ErrProductNotActive
	}

	if !discount.IsValidAt(now) && !discount.HasStarted(now) {
		if discount.IsExpired(now) {
			return ErrDiscountExpired
		}
	}

	p.discount = discount
	p.updatedAt = now
	p.changes.MarkDirty(FieldDiscount)
	p.events = append(p.events, NewDiscountAppliedEvent(
		p.id,
		discount.Percentage(),
		discount.StartDate(),
		discount.EndDate(),
		now,
	))

	return nil
}

// RemoveDiscount removes any discount from the product.
func (p *Product) RemoveDiscount(now time.Time) error {
	if p.discount == nil {
		return ErrNoDiscountToRemove
	}

	p.discount = nil
	p.updatedAt = now
	p.changes.MarkDirty(FieldDiscount)
	p.events = append(p.events, NewDiscountRemovedEvent(p.id, now))

	return nil
}
