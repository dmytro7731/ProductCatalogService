// Package queries contains the read-side query handlers following CQRS pattern.
//
// Queries are optimized for read operations and may bypass the domain layer
// for performance. They return DTOs (Data Transfer Objects) rather than
// domain entities.
//
// Available queries:
//   - get_product: Retrieve a single product by ID with effective price calculation
//   - list_products: Paginated listing with filtering by category and status
//
// Query handlers are stateless and produce no side effects.
package queries
