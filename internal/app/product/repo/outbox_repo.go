package repo

import (
	"encoding/json"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"

	"github.com/product-catalog-service/internal/app/product/contracts"
	"github.com/product-catalog-service/internal/app/product/domain"
	"github.com/product-catalog-service/internal/models/m_outbox"
	"github.com/product-catalog-service/internal/pkg/clock"
)

// OutboxRepo implements the OutboxRepository interface for Spanner.
type OutboxRepo struct {
	model *m_outbox.Model
	clock clock.Clock
}

// NewOutboxRepo creates a new OutboxRepo.
func NewOutboxRepo(clock clock.Clock) *OutboxRepo {
	return &OutboxRepo{
		model: m_outbox.NewModel(),
		clock: clock,
	}
}

// InsertMut returns a mutation for inserting an outbox event.
func (r *OutboxRepo) InsertMut(event *contracts.OutboxEvent) *spanner.Mutation {
	dbEvent := &m_outbox.OutboxEvent{
		EventID:     event.ID,
		EventType:   event.EventType,
		AggregateID: event.AggregateID,
		Payload: spanner.NullJSON{
			Value: json.RawMessage(event.Payload),
			Valid: true,
		},
		Status:    event.Status,
		CreatedAt: r.clock.Now(),
	}

	return r.model.InsertMut(dbEvent)
}

// InsertFromDomainEventMut creates an outbox event from a domain event.
func (r *OutboxRepo) InsertFromDomainEventMut(event domain.DomainEvent) (*spanner.Mutation, error) {
	payload, err := r.serializeEvent(event)
	if err != nil {
		return nil, err
	}

	outboxEvent := &contracts.OutboxEvent{
		ID:          uuid.New().String(),
		EventType:   event.EventType(),
		AggregateID: event.AggregateID(),
		Payload:     payload,
		Status:      m_outbox.StatusPending,
	}

	return r.InsertMut(outboxEvent), nil
}

func (r *OutboxRepo) serializeEvent(event domain.DomainEvent) ([]byte, error) {
	eventData := map[string]interface{}{
		"event_type":   event.EventType(),
		"aggregate_id": event.AggregateID(),
		"occurred_at":  event.OccurredAt(),
	}

	switch e := event.(type) {
	case *domain.ProductCreatedEvent:
		eventData["name"] = e.Name
		eventData["description"] = e.Description
		eventData["category"] = e.Category
		eventData["base_price"] = map[string]int64{
			"numerator":   e.BasePrice.Numerator(),
			"denominator": e.BasePrice.Denominator(),
		}

	case *domain.ProductUpdatedEvent:
		eventData["name"] = e.Name
		eventData["description"] = e.Description
		eventData["category"] = e.Category

	case *domain.ProductActivatedEvent:
		// No additional data

	case *domain.ProductDeactivatedEvent:
		// No additional data

	case *domain.ProductArchivedEvent:
		// No additional data

	case *domain.DiscountAppliedEvent:
		eventData["percentage"] = e.Percentage
		eventData["start_date"] = e.StartDate
		eventData["end_date"] = e.EndDate

	case *domain.DiscountRemovedEvent:
		// No additional data
	}

	return json.Marshal(eventData)
}
