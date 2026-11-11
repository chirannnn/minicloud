#!/bin/bash
# Integration test runner for MiniCloud Phase 1
# This script starts the test PostgreSQL container and runs integration tests

set -e

echo "Starting test PostgreSQL container..."
docker compose --profile test up -d postgres-test

echo "Waiting for PostgreSQL to be ready..."
timeout 60 bash -c 'until docker compose exec -T postgres-test pg_isready -U minicloud -d minicloud_test; do sleep 2; done'

echo "Running integration tests..."
cd apps/api
TEST_DATABASE_URL="postgres://minicloud:minicloud_dev_password@localhost:5433/minicloud_test?sslmode=disable" go test -v ./internal/modules/controlplane -run TestIntegration

echo "Stopping test PostgreSQL container..."
cd ../..
docker compose --profile test down

echo "Integration tests completed."
