package m_outbox

// Table name
const TableName = "outbox_events"

// Column names for the outbox_events table.
const (
	EventID     = "event_id"
	EventType   = "event_type"
	AggregateID = "aggregate_id"
	Payload     = "payload"
	Status      = "status"
	CreatedAt   = "created_at"
	ProcessedAt = "processed_at"
)

// Event status constants.
const (
	StatusPending   = "pending"
	StatusProcessed = "processed"
	StatusFailed    = "failed"
)

// AllColumns returns all column names.
func AllColumns() []string {
	return []string{
		EventID,
		EventType,
		AggregateID,
		Payload,
		Status,
		CreatedAt,
		ProcessedAt,
	}
}

// InsertColumns returns columns used for insert operations.
func InsertColumns() []string {
	return []string{
		EventID,
		EventType,
		AggregateID,
		Payload,
		Status,
		CreatedAt,
	}
}
