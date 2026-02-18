// Package domain contains the core business logic for the product catalog service.
//
// This package follows Domain-Driven Design principles and contains:
//   - Product: The aggregate root representing a product in the catalog
//   - Money: A value object for precise monetary calculations using big.Rat
//   - Discount: A value object representing percentage-based discounts with validity periods
//   - Domain events: Captured as intents when business state changes
//   - Domain errors: Sentinel errors representing business rule violations
//
// The domain layer is intentionally pure and has no external dependencies on:
//   - context.Context
//   - Database libraries
//   - Proto definitions
//   - External frameworks
//
// This ensures the business logic remains testable, portable, and focused solely
// on expressing the business rules of the product catalog domain.
package domain
