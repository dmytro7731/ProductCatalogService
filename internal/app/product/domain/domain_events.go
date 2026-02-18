package domain

import (
	"time"
)

// DomainEvent is the base interface for all domain events.
// Domain events are simple structs that capture what happened in the domain.
type DomainEvent interface {
	EventType() string
	OccurredAt() time.Time
	AggregateID() string
}

// BaseEvent contains common fields for all domain events.
type BaseEvent struct {
	aggregateID string
	occurredAt  time.Time
}

func (e BaseEvent) AggregateID() string {
	return e.aggregateID
}

func (e BaseEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// ProductCreatedEvent is raised when a new product is created.
type ProductCreatedEvent struct {
	BaseEvent
	Name        string
	Description string
	Category    string
	BasePrice   *Money
}

func (e ProductCreatedEvent) EventType() string {
	return "product.created"
}

func NewProductCreatedEvent(id, name, description, category string, basePrice *Money, occurredAt time.Time) *ProductCreatedEvent {
	return &ProductCreatedEvent{
		BaseEvent: BaseEvent{
			aggregateID: id,
			occurredAt:  occurredAt,
		},
		Name:        name,
		Description: description,
		Category:    category,
		BasePrice:   basePrice,
	}
}

// ProductUpdatedEvent is raised when product details are changed.
type ProductUpdatedEvent struct {
	BaseEvent
	Name        string
	Description string
	Category    string
}

func (e ProductUpdatedEvent) EventType() string {
	return "product.updated"
}

func NewProductUpdatedEvent(id, name, description, category string, occurredAt time.Time) *ProductUpdatedEvent {
	return &ProductUpdatedEvent{
		BaseEvent: BaseEvent{
			aggregateID: id,
			occurredAt:  occurredAt,
		},
		Name:        name,
		Description: description,
		Category:    category,
	}
}

// ProductActivatedEvent is raised when a product is activated.
type ProductActivatedEvent struct {
	BaseEvent
}

func (e ProductActivatedEvent) EventType() string {
	return "product.activated"
}

func NewProductActivatedEvent(id string, occurredAt time.Time) *ProductActivatedEvent {
	return &ProductActivatedEvent{
		BaseEvent: BaseEvent{
			aggregateID: id,
			occurredAt:  occurredAt,
		},
	}
}

// ProductDeactivatedEvent is raised when a product is deactivated.
type ProductDeactivatedEvent struct {
	BaseEvent
}

func (e ProductDeactivatedEvent) EventType() string {
	return "product.deactivated"
}

func NewProductDeactivatedEvent(id string, occurredAt time.Time) *ProductDeactivatedEvent {
	return &ProductDeactivatedEvent{
		BaseEvent: BaseEvent{
			aggregateID: id,
			occurredAt:  occurredAt,
		},
	}
}

// ProductArchivedEvent is raised when a product is archived (soft deleted).
type ProductArchivedEvent struct {
	BaseEvent
}

func (e ProductArchivedEvent) EventType() string {
	return "product.archived"
}

func NewProductArchivedEvent(id string, occurredAt time.Time) *ProductArchivedEvent {
	return &ProductArchivedEvent{
		BaseEvent: BaseEvent{
			aggregateID: id,
			occurredAt:  occurredAt,
		},
	}
}

// DiscountAppliedEvent is raised when a discount is applied to a product.
type DiscountAppliedEvent struct {
	BaseEvent
	Percentage int64
	StartDate  time.Time
	EndDate    time.Time
}

func (e DiscountAppliedEvent) EventType() string {
	return "product.discount_applied"
}

func NewDiscountAppliedEvent(id string, percentage int64, startDate, endDate, occurredAt time.Time) *DiscountAppliedEvent {
	return &DiscountAppliedEvent{
		BaseEvent: BaseEvent{
			aggregateID: id,
			occurredAt:  occurredAt,
		},
		Percentage: percentage,
		StartDate:  startDate,
		EndDate:    endDate,
	}
}

// DiscountRemovedEvent is raised when a discount is removed from a product.
type DiscountRemovedEvent struct {
	BaseEvent
}

func (e DiscountRemovedEvent) EventType() string {
	return "product.discount_removed"
}

func NewDiscountRemovedEvent(id string, occurredAt time.Time) *DiscountRemovedEvent {
	return &DiscountRemovedEvent{
		BaseEvent: BaseEvent{
			aggregateID: id,
			occurredAt:  occurredAt,
		},
	}
}
