package contracts

import (
	"cloud.google.com/go/spanner"
	"github.com/product-catalog-service/internal/app/product/domain"
)

// OutboxEvent represents an enriched event ready for persistence.
type OutboxEvent struct {
	ID          string
	EventType   string
	AggregateID string
	Payload     []byte
	Status      string
}

// OutboxRepository defines the interface for outbox event persistence.
type OutboxRepository interface {
	// InsertMut returns a mutation for inserting an outbox event.
	InsertMut(event *OutboxEvent) *spanner.Mutation

	// InsertFromDomainEventMut creates an outbox event from a domain event and returns its mutation.
	InsertFromDomainEventMut(event domain.DomainEvent) (*spanner.Mutation, error)
}
