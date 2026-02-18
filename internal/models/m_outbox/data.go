package m_outbox

import (
	"time"

	"cloud.google.com/go/spanner"
)

// OutboxEvent represents the database model for an outbox event.
type OutboxEvent struct {
	EventID     string
	EventType   string
	AggregateID string
	Payload     spanner.NullJSON
	Status      string
	CreatedAt   time.Time
	ProcessedAt spanner.NullTime
}

// Model provides methods for creating Spanner mutations.
type Model struct{}

// NewModel creates a new Model instance.
func NewModel() *Model {
	return &Model{}
}

// InsertMut creates an insert mutation for an outbox event.
func (m *Model) InsertMut(e *OutboxEvent) *spanner.Mutation {
	return spanner.InsertMap(TableName, map[string]interface{}{
		EventID:     e.EventID,
		EventType:   e.EventType,
		AggregateID: e.AggregateID,
		Payload:     e.Payload,
		Status:      e.Status,
		CreatedAt:   e.CreatedAt,
	})
}

// UpdateStatusMut creates an update mutation for changing event status.
func (m *Model) UpdateStatusMut(eventID, status string, processedAt *time.Time) *spanner.Mutation {
	updates := map[string]interface{}{
		EventID: eventID,
		Status:  status,
	}
	if processedAt != nil {
		updates[ProcessedAt] = *processedAt
	}
	return spanner.UpdateMap(TableName, updates)
}

// MarkProcessedMut creates a mutation to mark an event as processed.
func (m *Model) MarkProcessedMut(eventID string, processedAt time.Time) *spanner.Mutation {
	return m.UpdateStatusMut(eventID, StatusProcessed, &processedAt)
}

// MarkFailedMut creates a mutation to mark an event as failed.
func (m *Model) MarkFailedMut(eventID string) *spanner.Mutation {
	return m.UpdateStatusMut(eventID, StatusFailed, nil)
}
