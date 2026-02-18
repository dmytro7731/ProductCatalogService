#!/bin/bash
# Setup script for Spanner emulator
# Run this after starting the emulator with: docker run -d -p 9010:9010 -p 9020:9020 gcr.io/cloud-spanner-emulator/emulator

set -e

export SPANNER_EMULATOR_HOST=localhost:9010

echo "Waiting for Spanner emulator to be ready..."
sleep 5

# Install gcloud if needed, or use curl directly
# Using the REST API to create instance and database

echo "Creating instance via REST API..."
curl -X POST "http://localhost:9020/v1/projects/test-project/instances" \
  -H "Content-Type: application/json" \
  -d '{
    "instanceId": "test-instance",
    "instance": {
      "config": "projects/test-project/instanceConfigs/emulator-config",
      "displayName": "Test Instance",
      "nodeCount": 1
    }
  }' || true

echo ""
echo "Creating database via REST API..."
curl -X POST "http://localhost:9020/v1/projects/test-project/instances/test-instance/databases" \
  -H "Content-Type: application/json" \
  -d '{
    "createStatement": "CREATE DATABASE `product-catalog`"
  }' || true

echo ""
echo "Waiting for database to be created..."
sleep 3

echo "Applying schema..."
curl -X PATCH "http://localhost:9020/v1/projects/test-project/instances/test-instance/databases/product-catalog/ddl" \
  -H "Content-Type: application/json" \
  -d '{
    "statements": [
      "CREATE TABLE products (product_id STRING(36) NOT NULL, name STRING(255) NOT NULL, description STRING(MAX), category STRING(100) NOT NULL, base_price_numerator INT64 NOT NULL, base_price_denominator INT64 NOT NULL, discount_percent NUMERIC, discount_start_date TIMESTAMP, discount_end_date TIMESTAMP, status STRING(20) NOT NULL, created_at TIMESTAMP NOT NULL, updated_at TIMESTAMP NOT NULL, archived_at TIMESTAMP) PRIMARY KEY (product_id)",
      "CREATE TABLE outbox_events (event_id STRING(36) NOT NULL, event_type STRING(100) NOT NULL, aggregate_id STRING(36) NOT NULL, payload JSON NOT NULL, status STRING(20) NOT NULL, created_at TIMESTAMP NOT NULL, processed_at TIMESTAMP) PRIMARY KEY (event_id)",
      "CREATE INDEX idx_outbox_status ON outbox_events(status, created_at)",
      "CREATE INDEX idx_products_category ON products(category, status)",
      "CREATE INDEX idx_products_status ON products(status, created_at DESC)"
    ]
  }' || true

echo ""
echo "Setup complete!"
echo "Spanner emulator is ready at localhost:9010"
echo ""
echo "To use it, set: export SPANNER_EMULATOR_HOST=localhost:9010"
