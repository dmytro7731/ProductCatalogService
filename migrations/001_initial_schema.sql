-- Migration: 001_initial_schema
-- Description: Create initial schema for Product Catalog Service
-- Created: 2026-02-18

-- Products table stores product catalog information
CREATE TABLE products (
    product_id STRING(36) NOT NULL,
    name STRING(255) NOT NULL,
    description STRING(MAX),
    category STRING(100) NOT NULL,
    base_price_numerator INT64 NOT NULL,
    base_price_denominator INT64 NOT NULL,
    discount_percent NUMERIC,
    discount_start_date TIMESTAMP,
    discount_end_date TIMESTAMP,
    status STRING(20) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    archived_at TIMESTAMP,
) PRIMARY KEY (product_id);

-- Outbox events table for transactional outbox pattern
CREATE TABLE outbox_events (
    event_id STRING(36) NOT NULL,
    event_type STRING(100) NOT NULL,
    aggregate_id STRING(36) NOT NULL,
    payload JSON NOT NULL,
    status STRING(20) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    processed_at TIMESTAMP,
) PRIMARY KEY (event_id);

-- Index for querying outbox events by status
CREATE INDEX idx_outbox_status ON outbox_events(status, created_at);

-- Index for filtering products by category and status
CREATE INDEX idx_products_category ON products(category, status);

-- Index for filtering products by status (for active products listing)
CREATE INDEX idx_products_status ON products(status, created_at DESC);
