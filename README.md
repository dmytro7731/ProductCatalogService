# Product Catalog Service

A microservice for managing product catalogs with pricing and discounts, built with Clean Architecture and Domain-Driven Design principles.

## Overview

This service manages products and their pricing with:
- Product lifecycle management (create, update, activate, deactivate, archive)
- Percentage-based discounts with validity periods
- Precise decimal arithmetic for money calculations
- Event sourcing via transactional outbox pattern
- CQRS (Command Query Responsibility Segregation)

## Technology Stack

- **Language**: Go 1.21+
- **Database**: Google Cloud Spanner (emulator for local development)
- **Transport**: gRPC with Protocol Buffers
- **Architecture**: Clean Architecture / Hexagonal Architecture
- **Patterns**: DDD, CQRS, Transactional Outbox, Repository Pattern

## Project Structure

```
product-catalog-service/
├── cmd/server/                 # Application entry point
├── internal/
│   ├── app/product/
│   │   ├── domain/            # Domain layer (pure business logic)
│   │   │   ├── product.go     # Product aggregate
│   │   │   ├── discount.go    # Discount value object
│   │   │   ├── money.go       # Money value object
│   │   │   ├── domain_events.go
│   │   │   ├── domain_errors.go
│   │   │   └── services/      # Domain services
│   │   ├── usecases/          # Application layer (commands)
│   │   ├── queries/           # CQRS read side
│   │   ├── contracts/         # Repository interfaces
│   │   └── repo/              # Spanner implementations
│   ├── models/                # Database models
│   ├── transport/grpc/        # gRPC handlers
│   ├── services/              # DI container
│   └── pkg/                   # Shared utilities
├── proto/                     # Protocol Buffers definitions
├── migrations/                # Database migrations
├── tests/e2e/                 # End-to-end tests
├── docker-compose.yml
├── Makefile
└── README.md
```

## Architecture

### Directory Structure Rationale

The project follows a **vertical slice architecture** for use cases and queries, where each operation has its own directory:

```
usecases/
├── create_product/    # CreateProduct command
├── update_product/    # UpdateProduct command
├── activate_product/  # ActivateProduct command
└── ...
```

This pattern is intentional and provides:
- **Isolation**: Each use case is self-contained with its own request/response types
- **Scalability**: New features don't touch existing code
- **Navigation**: Easy to find code by feature name
- **Testing**: Tests co-located with implementation

While this increases folder count, it's a recognized Go pattern for larger services and follows the "package by feature" principle.

### Domain Layer Purity

The domain layer contains only pure Go business logic without external dependencies:
- No `context.Context` imports
- No database libraries
- No proto definitions
- No framework dependencies

### The Golden Mutation Pattern

Every write operation follows this pattern:
1. Load or create aggregate
2. Call domain methods (validation happens here)
3. Build commit plan
4. Get mutations from repository (repo returns, doesn't apply)
5. Add outbox events
6. Apply plan atomically (usecase applies, not handler)

### CQRS

- **Commands**: Go through domain aggregate, use CommitPlan
- **Queries**: May bypass domain for optimization, direct database access

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)
- Make (optional, for convenience)

### Start Spanner Emulator

```bash
# Start the Spanner emulator
docker-compose up -d

# Run the setup script to create instance, database, and schema
chmod +x scripts/setup-emulator.sh
./scripts/setup-emulator.sh
```

### Run Migrations

Migrations are applied by the setup script. To manually apply with gcloud:

```bash
gcloud config set auth/disable_credentials true
gcloud config set project test-project
gcloud config set api_endpoint_overrides/spanner http://localhost:9020/

gcloud spanner databases ddl update product-catalog \
  --instance=test-instance \
  --ddl-file=migrations/001_initial_schema.sql
```

### Run Tests

```bash
# Set emulator host
export SPANNER_EMULATOR_HOST=localhost:9010

# Run unit tests
make test-unit

# Run E2E tests (requires running emulator)
make test-e2e

# Run all tests
make test
```

### Start the Server

```bash
# Using Make
make run

# Or directly
export SPANNER_EMULATOR_HOST=localhost:9010
go run ./cmd/server
```

The gRPC server starts on port 50051 by default.

### Using Docker

```bash
# Build and start
docker-compose up --build

# Just the service (after emulator is ready)
docker-compose up product-catalog
```

## API Usage

### gRPC Endpoints

| Method | Description |
|--------|-------------|
| `CreateProduct` | Create a new product |
| `UpdateProduct` | Update product details |
| `ActivateProduct` | Activate a product |
| `DeactivateProduct` | Deactivate a product |
| `ArchiveProduct` | Soft delete a product |
| `ApplyDiscount` | Apply percentage discount |
| `RemoveDiscount` | Remove discount |
| `GetProduct` | Get product by ID |
| `ListProducts` | List products with filters |

### Example with grpcurl

```bash
# Create a product
grpcurl -plaintext -d '{
  "name": "Laptop",
  "description": "High-performance laptop",
  "category": "Electronics",
  "base_price": {"numerator": 99999, "denominator": 100}
}' localhost:50051 product.v1.ProductService/CreateProduct

# Get product
grpcurl -plaintext -d '{"product_id": "<id>"}' \
  localhost:50051 product.v1.ProductService/GetProduct

# Activate product
grpcurl -plaintext -d '{"product_id": "<id>"}' \
  localhost:50051 product.v1.ProductService/ActivateProduct

# Apply 20% discount
grpcurl -plaintext -d '{
  "product_id": "<id>",
  "percentage": 20,
  "start_date": "2026-02-18T00:00:00Z",
  "end_date": "2026-02-28T23:59:59Z"
}' localhost:50051 product.v1.ProductService/ApplyDiscount

# List active products
grpcurl -plaintext -d '{"active_only": true, "limit": 10}' \
  localhost:50051 product.v1.ProductService/ListProducts
```

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `GRPC_ADDRESS` | `:50051` | gRPC server address |
| `SPANNER_PROJECT` | `test-project` | GCP project ID |
| `SPANNER_INSTANCE` | `test-instance` | Spanner instance |
| `SPANNER_DATABASE` | `product-catalog` | Database name |
| `SPANNER_EMULATOR_HOST` | - | Emulator host (enables emulator mode) |

## Design Decisions & Trade-offs

### Money Representation

Money is stored as numerator/denominator (rational numbers) using `math/big.Rat` to avoid floating-point precision issues. This ensures accurate financial calculations.

```go
// $19.99 is stored as numerator=1999, denominator=100
price := big.NewRat(1999, 100)
```

### Change Tracking

The domain aggregate tracks which fields have changed, allowing the repository to generate targeted updates instead of full-row replacements:

```go
if product.Changes().Dirty(domain.FieldName) {
    updates[m_product.Name] = product.Name()
}
```

### Transactional Outbox

Domain events are stored in the `outbox_events` table within the same transaction as the aggregate changes. This ensures reliable event publishing without distributed transactions.

### Status State Machine

Products follow a state machine:
- `draft` → `active` or `archived`
- `active` → `inactive` or (deactivate first, then `archived`)
- `inactive` → `active` or `archived`
- `archived` → (terminal state)

### Discounts

- Only one active discount per product at a time
- Discount period must be valid (end > start)
- Can only apply discount to active products
- Effective price is calculated on read based on current time

## Domain Events

| Event | Trigger |
|-------|---------|
| `product.created` | New product created |
| `product.updated` | Product details changed |
| `product.activated` | Product activated |
| `product.deactivated` | Product deactivated |
| `product.archived` | Product soft deleted |
| `product.discount_applied` | Discount added |
| `product.discount_removed` | Discount removed |

## CI/CD

The project includes GitHub Actions workflows for automated quality gates:

### Pipeline Stages

1. **Lint**: Runs `golangci-lint` with project-specific configuration
2. **Test**: Executes unit tests and E2E tests against Spanner emulator
3. **Build**: Compiles the binary and builds Docker image

### Running Locally

```bash
# Run linter
make lint

# Run all tests
make test

# Build binary
make build
```

### Configuration

- `.github/workflows/ci.yml`: CI pipeline definition
- `.golangci.yml`: Linter configuration

## Development

### Generate Protobuf

```bash
make proto
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Run vet
make vet

# Coverage report
make coverage
```

### Project Commands

```bash
make help  # Show all available commands
```

## What's NOT Implemented

Per requirements, the following are intentionally omitted:
- Authentication/authorization
- Background outbox processor
- Actual Pub/Sub publishing
- Metrics/monitoring
- REST API

## Author

Built with care from Ukraine 🇺🇦

## License

MIT License
