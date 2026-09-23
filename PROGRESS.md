# MiniCloud Progress

## Current Phase

Phase 1 — Core Control Plane

## Status

IN PROGRESS

## Current Branch

`phase-1-control-plane` (5 commits ahead of main)

## Phase 0 — Foundation: COMPLETE

### Completed
- Repository foundation, documentation, OpenAPI seed, API skeleton, console skeleton, Compose definition, and CI configuration created.
- `docker compose config --quiet` passed.
- All Compose services started successfully; PostgreSQL, Redis, Kafka, MinIO, Mailpit, OTel Collector, Prometheus, Grafana, Loki, and Tempo reported healthy. Kafka UI was reachable at HTTP 200.
- Go formatting, tests, vet, and API build passed.
- OpenAPI validation passed.
- Console lint, typecheck, and production build passed.
- API `/health`, `/ready`, and `/api/v1/health` all returned HTTP 200 against the live infrastructure.
- API CORS allowed the console origin (`http://localhost:3000`), validating the console's Phase 0 health request path.

## Phase 1 — Core Control Plane: IN PROGRESS

### Completed
- **Database Schema**: Complete schema with users, projects, applications, deployments, audit_logs, and outbox_events tables with proper indexes and constraints.
- **Migration Runner**: Custom Go-based migration system with transaction support and idempotent execution.
- **API Endpoints**: 17 CRUD endpoints implemented for users, projects, applications, and deployments.
- **Basic Validation**: Email format, UUID parsing, slug validation, and deployment status validation.
- **Transaction Support**: Transactions for project and application creation (audit logs + outbox events).
- **Ownership Boundaries**: Foreign key constraints and scope validation implemented.
- **OpenAPI Contract**: Phase 1 endpoints added to OpenAPI specification.
- **SQLC Configuration**: sqlc.yaml configured for PostgreSQL with pgx/v5.

### In Progress
- **SQLC Integration**: Configuration exists but generated code not yet used (handlers use raw SQL).
- **OpenAPI Schemas**: Endpoints defined but request/response schemas incomplete.
- **Testing**: Only basic router tests exist; no handler or integration tests.
- **Console Integration**: Console still shows Phase 0 messaging; no Phase 1 resource UI.

### Remaining Work
- Complete SQLC queries for all resources (currently only 3 user queries exist).
- Generate SQLC code and refactor handlers to use it.
- Add comprehensive handler tests.
- Add integration tests with database.
- Complete OpenAPI schemas with detailed request/response definitions.
- Implement pagination (currently hardcoded limit 100).
- Improve error handling consistency.
- Complete audit logging for all operations.
- Complete outbox events for all operations.
- Update console UI to display Phase 1 resources.
- Update documentation to reflect Phase 1 progress.

## Currently Working On

Phase 1 implementation - completing control plane resources with contract-backed API, validation, and tests.

## Next Steps

1. Complete SQLC integration (queries, generation, handler refactoring).
2. Add comprehensive test coverage (handler tests, integration tests).
3. Complete OpenAPI schemas and generate type clients.
4. Update console UI for Phase 1 resources.
5. Update documentation to reflect Phase 1 completion.

## Decisions Made

- Phase 0 remains a Go modular monolith plus Next.js console.
- Readiness checks local dependency TCP reachability; liveness remains process-only.
- OpenAPI contains only current endpoints and grows phase by phase.
- Phase 1 uses PostgreSQL for persistent storage with custom migration runner.
- SQLC configured for type-safe database operations (not yet utilized).
- Audit logs and outbox events infrastructure added for future event-driven features.

## Problems / Technical Debt

- SQLC configured but not yet used; handlers currently use raw SQL.
- Minimal test coverage; only basic router tests exist.
- OpenAPI schemas incomplete (endpoints defined but schemas minimal).
- Console not integrated with Phase 1 resources.
- Pagination not implemented (hardcoded limit 100).
- Error handling inconsistent across handlers.
- Audit logging incomplete (actor_id not set, some operations missing).
- Outbox events written but no publisher implemented.
- Documentation outdated (PROGRESS.md, README.md, OpenAPI README).

## Important Notes For Next Agent

- Complete Phase 1 before starting Phase 2.
- Focus on SQLC integration and testing as highest priority.
- Do not add resource/business APIs beyond Phase 1 scope.
- Do not create node-agent, scheduler, or worker before their phases.
- Preserve contract-first approach (OpenAPI and SQLC).
- Ensure all changes are tested before considering Phase 1 complete.

## Last Updated

2026-09-23
