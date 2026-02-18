// Package usecases contains the application layer use cases for the product catalog service.
//
// Each use case follows the "Golden Mutation Pattern":
//  1. Load or create the domain aggregate
//  2. Execute domain logic (validation happens in domain)
//  3. Build a commit plan with mutations
//  4. Get mutations from repository (repository returns, doesn't apply)
//  5. Add outbox events for reliable event publishing
//  6. Apply the plan atomically
//
// Use cases are responsible for:
//   - Orchestrating domain operations
//   - Managing transactions via CommitPlan
//   - Ensuring domain events are persisted in the outbox
//
// Available use cases:
//   - create_product: Create new products in the catalog
//   - update_product: Update product details (name, description, category)
//   - activate_product: Transition product to active status
//   - deactivate_product: Transition product to inactive status
//   - archive_product: Soft delete a product
//   - apply_discount: Apply percentage-based discount to a product
//   - remove_discount: Remove discount from a product
package usecases
