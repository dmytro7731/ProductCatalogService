.PHONY: all build run test test-unit test-e2e lint proto clean docker-up docker-down migrate

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOMOD=$(GOCMD) mod
BINARY_NAME=product-catalog-server

# Proto parameters
PROTO_DIR=proto
PROTO_OUT=proto

all: build

# Build the application
build:
	$(GOBUILD) -o bin/$(BINARY_NAME) ./cmd/server

# Run the application
run: build
	./bin/$(BINARY_NAME)

# Run all tests
test: test-unit test-e2e

# Run unit tests
test-unit:
	$(GOTEST) -v -race -short ./internal/...

# Run e2e tests
test-e2e:
	$(GOTEST) -v -race ./tests/e2e/...

# Run linter
lint:
	golangci-lint run ./...

# Vet code
vet:
	$(GOVET) ./...

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Generate protobuf code
proto:
	protoc --go_out=$(PROTO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/product/v1/product_service.proto

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out

# Start Docker services
docker-up:
	docker-compose up -d

# Stop Docker services
docker-down:
	docker-compose down -v

# Run migrations (requires Spanner emulator)
migrate:
	@echo "Running migrations..."
	@docker-compose exec spanner-setup gcloud spanner databases ddl update product-catalog \
		--instance=test-instance \
		--ddl-file=/migrations/001_initial_schema.sql

# Build and run in Docker
docker-build:
	docker-compose build product-catalog

# Logs
logs:
	docker-compose logs -f product-catalog

# Format code
fmt:
	gofmt -s -w .

# Run tests with coverage
coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out ./internal/...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Help
help:
	@echo "Available targets:"
	@echo "  build       - Build the application"
	@echo "  run         - Build and run the application"
	@echo "  test        - Run all tests"
	@echo "  test-unit   - Run unit tests"
	@echo "  test-e2e    - Run e2e tests"
	@echo "  lint        - Run linter"
	@echo "  vet         - Vet code"
	@echo "  deps        - Download dependencies"
	@echo "  proto       - Generate protobuf code"
	@echo "  clean       - Clean build artifacts"
	@echo "  docker-up   - Start Docker services"
	@echo "  docker-down - Stop Docker services"
	@echo "  migrate     - Run migrations"
	@echo "  coverage    - Run tests with coverage"
	@echo "  fmt         - Format code"
