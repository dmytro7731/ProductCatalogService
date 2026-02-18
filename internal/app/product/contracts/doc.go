// Package contracts defines the interfaces (ports) for the product domain.
//
// Following hexagonal architecture, these interfaces decouple the domain
// and use case layers from infrastructure implementations like databases.
//
// Interfaces:
//   - ProductRepository: Persistence operations for the Product aggregate
//   - OutboxRepository: Transactional outbox for reliable event publishing
//   - ProductReadModelRepository: Optimized read queries for CQRS
//
// Implementations of these interfaces reside in the repo package.
package contracts
