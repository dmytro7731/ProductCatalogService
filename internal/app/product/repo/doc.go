// Package repo contains the Spanner repository implementations.
//
// Repositories follow the pattern where they return mutations rather than
// applying them directly. This allows the use case layer to build atomic
// commit plans that include multiple mutations.
//
// Key implementations:
//   - ProductRepo: Handles product aggregate persistence with change tracking
//   - OutboxRepo: Handles transactional outbox event persistence
//   - ReadModelRepo: Optimized read-only queries for CQRS read side
//
// Repositories use change tracking to generate targeted updates, only
// persisting fields that have actually changed in the domain aggregate.
package repo
