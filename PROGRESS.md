# MiniCloud Progress

## Current Phase

Phase 1 — Core Control Plane

## Status

IN PROGRESS

## Current Branch

`phase-1-control-plane-remaining` (3 commits ahead of main)

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
- **SQLC Integration**: Complete SQLC queries for all resources (users, projects, applications, deployments) generated and integrated.
- **Repository Layer**: Repository layer created using SQLC-generated code to handle database operations.
- **Handler Refactoring**: All control-plane handlers refactored to use repository layer instead of raw SQL.
- **Service Layer**: Service layer created to separate business logic from HTTP handlers and database persistence.
- **Business Rules**: Project/application/deployment relationship validation, ownership checks, and status validation moved to service layer.
- **Transaction Orchestration**: Audit and outbox event transaction handling moved to service layer.
- **Service-Level Errors**: Clear service error types introduced (ErrUserNotFound, ErrProjectNotFound, ErrApplicationNotFound, ErrDeploymentNotFound, ErrInvalidRelationship, ErrInvalidStatus, ErrResourceConflict).
- **Handler Refactoring**: All handlers refactored to use service layer with proper error translation to HTTP responses.

### In Progress
- **OpenAPI Schemas**: Endpoints defined but request/response schemas incomplete.
- **Testing**: Only basic router tests exist; no handler or integration tests.
- **Console Integration**: Console still shows Phase 0 messaging; no Phase 1 resource UI.

### Remaining Work
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

1. Add comprehensive test coverage (handler tests, integration tests).
2. Complete OpenAPI schemas and generate type clients.
3. Update console UI for Phase 1 resources.
4. Update documentation to reflect Phase 1 progress.

## Decisions Made

- Phase 0 remains a Go modular monolith plus Next.js console.
- Readiness checks local dependency TCP reachability; liveness remains process-only.
- OpenAPI contains only current endpoints and grows phase by phase.
- Phase 1 uses PostgreSQL for persistent storage with custom migration runner.
- SQLC configured and integrated for type-safe database operations through repository layer.
- Service layer introduced to separate business logic from HTTP handlers and database persistence.
- Audit logs and outbox events infrastructure added for future event-driven features.

## Problems / Technical Debt

- Minimal test coverage; only basic router and service validation tests exist.
- OpenAPI schemas incomplete (endpoints defined but schemas minimal).
- Console not integrated with Phase 1 resources.
- Pagination not implemented (hardcoded limit 100).
- Error handling inconsistent across handlers (service layer improves this but not complete).
- Audit logging incomplete (actor_id not set, some operations missing).
- Outbox events written but no publisher implemented.
- Documentation outdated (PROGRESS.md, README.md, OpenAPI README).

## Important Notes For Next Agent

- Complete Phase 1 before starting Phase 2.
- Focus on testing and OpenAPI completion as highest priority.
- Do not add resource/business APIs beyond Phase 1 scope.
- Do not create node-agent, scheduler, or worker before their phases.
- Preserve contract-first approach (OpenAPI and SQLC).
- Ensure all changes are tested before considering Phase 1 complete.

## Last Updated

2026-09-28
