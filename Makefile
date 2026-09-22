.DEFAULT_GOAL := help

.PHONY: help setup up down restart logs api console test lint generate verify reset compose-config

help:
	@echo "MiniCloud development commands"
	@echo "  setup          Install JavaScript dependencies"
	@echo "  up/down        Start or stop local infrastructure"
	@echo "  logs           Follow infrastructure logs"
	@echo "  api/console    Start an application locally"
	@echo "  test/lint      Run available checks"
	@echo "  generate       Validate the OpenAPI contract"
	@echo "  verify         Run non-destructive Phase 0 checks"

setup:
	pnpm install --frozen-lockfile=false

up:
	docker compose up -d

down:
	docker compose down

restart: down up

logs:
	docker compose logs -f

api:
	go run ./apps/api/cmd/minicloud

console:
	pnpm --filter @minicloud/console dev

test:
	go test ./apps/api/...
	pnpm test

lint:
	gofmt -w apps/api
	go vet ./apps/api/...
	pnpm lint

generate:
	pnpm openapi:validate

compose-config:
	docker compose config --quiet

verify: compose-config generate test lint

reset:
	docker compose down -v --remove-orphans
