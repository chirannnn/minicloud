# MiniCloud Progress

## Current Phase

Phase 1 — Core Control Plane

## Status

IN PROGRESS

## Current Branch

`phase-1-control-plane-remaining` (7 commits ahead of main)

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
- **Service-Level Errors**: Clear service error types introduced (ErrUserNotFound, ErrProjectNotFound, ErrApplicationNotFound, ErrDeploymentNotFound, ErrInvalidRelationship, ErrInvalidStatus, ErrResourceConflict, ErrValidation, ErrInternal).
- **Handler Refactoring**: All handlers refactored to use service layer with proper error translation to HTTP responses.
- **Validation Standardization**: Dedicated validation module created with standardized validation methods for email, display name, slug, name, description, UUID, deployment version, and deployment status.
- **Error Handling Standardization**: Consistent HTTP error response structure (code, message) and error-to-HTTP status mapping implemented.
- **HTTP Error Codes**: Standardized error codes (VALIDATION_FAILED, NOT_FOUND, CONFLICT, INVALID_RELATIONSHIP, INTERNAL) defined and used consistently.
- **Error Mapping**: ToHTTPError function converts service errors to appropriate HTTP responses with correct status codes.
- **Handler Error Translation**: All handlers updated to use consistent error translation via ToHTTPError instead of manual error checking.
- **Validation Tests**: Comprehensive validation tests added for email, display name, slug, name, UUID, deployment version, and deployment status validation.
- **Error Mapping Tests**: Tests added for error-to-HTTP response mapping to ensure consistent error responses.
- **Pagination Implementation**: Pagination parameters (limit, offset) added to all list endpoints (users, projects, applications, deployments).
- **Pagination Validation**: ValidatePagination method added to validator with default limit of 20, maximum limit of 100, and validation for negative values and invalid integers.
- **Pagination Metadata**: List responses now include limit and offset metadata for consistency.
- **SQL-Based Pagination**: All pagination applied at SQL level using LIMIT and OFFSET clauses (not in-memory slicing).
- **Deterministic Ordering**: All list queries ordered by created_at for consistent pagination results.
- **Pagination Tests**: Comprehensive pagination tests added for defaults, custom limits, offsets, invalid parameters, negative values, and maximum limit enforcement.
- **Audit Logging Coverage**: All mutation operations now create audit records with appropriate action, resource type, resource ID, and metadata.
- **Outbox Event Coverage**: All mutation operations now create outbox events with event type, aggregate type, aggregate ID, and payload.
- **Transaction Safety**: Audit and outbox inserts are wrapped in the same transaction as the resource mutation, ensuring atomicity.
- **Audit Helper Functions**: insertAuditLog helper function created to standardize audit log insertion with JSONB metadata.
- **Outbox Helper Functions**: insertOutboxEvent helper function created to standardize outbox event insertion with JSONB payload.
- **Actor ID Handling**: Actor ID set to uuid.Nil for operations where authentication is not yet implemented (Phase 11).
- **Metadata Safety**: Sensitive data (deployment configuration) excluded from audit/outbox payloads to avoid storing secrets.
- **Audit Coverage**: CreateUser, CreateProject, UpdateProject, DeleteProject, CreateApplication, UpdateApplication, DeleteApplication, CreateDeployment, UpdateDeploymentStatus.
- **Outbox Coverage**: CreateUser, CreateProject, UpdateProject, DeleteProject, CreateApplication, UpdateApplication, DeleteApplication, CreateDeployment, UpdateDeploymentStatus.
- **OpenAPI Completion**: Complete OpenAPI 3.1 specification for all Phase 1 endpoints with accurate request/response schemas.
- **OpenAPI Endpoints**: All 17 Phase 1 endpoints documented with operation IDs, parameters, request bodies, and responses.
- **OpenAPI Schemas**: Complete schemas for User, Project, Application, Deployment resources with validation constraints.
- **Pagination Documentation**: Pagination parameters (limit, offset) documented with default 20 and max 100 values.
- **Error Response Schemas**: Standardized error schema with code (VALIDATION_FAILED, NOT_FOUND, CONFLICT, INVALID_RELATIONSHIP, INTERNAL) and message.
- **HTTP Status Mapping**: Accurate HTTP status codes documented for all success and error responses.
- **TypeScript Generation**: openapi-typescript configured to generate TypeScript types from OpenAPI specification.
- **Generated Types**: types.ts file generated with complete TypeScript interfaces for all API endpoints and schemas.
- **Type Generation Script**: Added generate script to openapi package that validates, generates, and formats types.
- **CI Integration**: Added type generation and freshness check to CI workflow to ensure types stay up to date.
- **OpenAPI Validation**: OpenAPI specification validated with swagger-cli.
- **Git Ignore**: Generated types.ts added to .gitignore to avoid committing generated files.
- **Console UI**: Full Phase 1 control plane console UI built with real API integration.
- **Console Pages**: Dashboard, Projects list/detail, Applications list/detail, Deployments list/detail pages implemented.
- **Console API Integration**: Shared API helper created for all API requests with proper error handling.
- **Console Workflows**: Create/read/delete for projects, applications, and deployments with real API calls.
- **Console Forms**: Create forms for projects, applications, and deployments with validation.
- **Console Navigation**: Navigation between dashboard, projects, applications, and deployments.
- **Console States**: Loading, empty, and error states handled across all pages.
- **Console Styling**: Enhanced CSS with form controls, buttons, and improved UI elements.
- **TypeScript Path Aliases**: Configured @/* path alias for cleaner imports.
- **Next.js Params**: Updated to use Promise-based params for Next.js 15 compatibility.

### In Progress
- **Testing**: Only basic router tests exist; no handler or integration tests.

### Remaining Work
- Add comprehensive handler tests.
- Add integration tests with database.
- Update documentation to reflect Phase 1 progress.

## Currently Working On

Phase 1 implementation - completing control plane resources with contract-backed API, validation, and tests.

## Next Steps

1. Add comprehensive test coverage (handler tests, integration tests).
2. Update console UI for Phase 1 resources.
3. Update documentation to reflect Phase 1 progress.

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
- Console not integrated with Phase 1 resources.
- Outbox events written but no publisher implemented (deferred to Phase 4 MiniEvents).
- Documentation outdated (PROGRESS.md, README.md, OpenAPI README).

## Important Notes For Next Agent

- Complete Phase 1 before starting Phase 2.
- Focus on testing and OpenAPI completion as highest priority.
- Do not add resource/business APIs beyond Phase 1 scope.
- Do not create node-agent, scheduler, or worker before their phases.
- Preserve contract-first approach (OpenAPI and SQLC).
- Ensure all changes are tested before considering Phase 1 complete.

## Last Updated

2026-10-12
