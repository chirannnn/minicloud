# MiniCloud Agent Instructions

## Required Reading

Before modifying code, read these documents in order:
1. `PROJECT.md` - Project specification, architecture, and principles
2. `PLAN.md` - Implementation roadmap and phase definitions
3. `PROGRESS.md` - Current implementation status and what's been completed
4. Inspect relevant files for the current phase

## Phase Discipline

- Implement only the user-requested current phase
- Do not pre-build future components or features
- Verify the current phase status in `PROGRESS.md` before starting
- Complete the current phase's definition of done before moving to the next phase
- Respect scope boundaries defined in `PLAN.md`

## Scope Boundaries

### Current Phase Boundaries
- Phase 0: Foundation only - no resource APIs
- Phase 1: Core control plane only - users, projects, applications, deployments
- Phase 2: Console and observability only - no new infrastructure services
- Phase 3-12: Each phase has specific scope - do not cross boundaries

### Infrastructure Boundaries
- Docker socket access belongs to node-agent (Phase 7), not the API
- Kafka integration belongs to MiniEvents (Phase 4)
- Redis cache belongs to MiniCache (Phase 3)
- MinIO operations belong to MiniStorage (Phase 6)
- Do not add infrastructure services before their designated phase

### Architectural Boundaries
- Preserve the modular Go-monolith architecture
- Introduce independent services only when a later phase genuinely requires them
- Maintain contract-first development (OpenAPI and SQLC)
- Keep the monolith until a genuine runtime boundary requires separation

## Coding Principles

### Architecture
- Use the existing modular structure within the monolith
- Follow the separation of concerns: config, database, modules, observability, router, server
- Prefer raw SQL only until SQLC is fully integrated (Phase 1 completion)
- Add new modules under `internal/modules/` for new features

### Code Quality
- Keep changes small, testable, and documented
- Avoid unnecessary services, abstractions, and large catch-all files
- Follow Go conventions and existing code style
- Use standard library features where appropriate
- Add comprehensive tests for new functionality

### Contract-First Development
- `packages/openapi/openapi.yaml` is the HTTP API source of truth
- Update OpenAPI before implementing new endpoints
- Use SQLC for type-safe database operations (when fully integrated)
- Maintain versioned event contracts under `packages/contracts/events` (Phase 4+)

### Data and Events
- Preserve correlation information: request ID, trace ID, event ID, causation ID
- Use audit logs for state changes
- Use outbox pattern for reliable event publishing
- Events should be idempotent

## Testing Requirements

### Required Tests
- Unit tests for business logic
- Integration tests for database operations
- HTTP tests for API endpoints
- Migration tests for schema changes
- Transaction rollback tests for critical operations

### Verification
- Run `go test ./apps/api/...` for backend tests
- Run `pnpm test` for frontend tests
- Run `make verify` or equivalent for comprehensive checks
- Manually verify critical paths

## Verification Requirements

### Before Considering Work Complete
- All tests pass
- Code compiles without errors
- Linting passes
- OpenAPI validation passes
- Infrastructure configuration is valid
- Manual testing confirms intended behavior
- Documentation is updated

### Infrastructure Verification
- `docker compose config --quiet` must pass
- All services start successfully
- Health checks pass for all services
- API health/readiness endpoints return 200

## Git/Branch Rules

### Branch Strategy
- Work on the appropriate phase branch (e.g., `phase-1-control-plane`)
- Do not modify main directly
- Keep branch history clean and meaningful

### Commit Expectations
- Write clear, concise commit messages
- Use conventional commit format (feat:, fix:, docs:, etc.)
- Include relevant context in commit messages
- Do not commit secrets or sensitive data
- Do not commit large generated files without consideration

### Sensitive Data
- Never commit secrets
- Use `.env.example` for documented safe local defaults
- Do not add `.env` to git
- Use environment variables for configuration

## Documentation Update Rules

### Required Updates
- Update `PROGRESS.md` with completed work (only verified facts)
- Update relevant documentation when architecture changes
- Add ADRs for meaningful architecture changes
- Update OpenAPI when API changes are made
- Update README.md when user-facing changes are made

### Documentation Standards
- Keep documentation accurate to the actual code
- Remove outdated information
- Update phase status when phases complete
- Document decisions and trade-offs
- Keep documentation concise and maintainable

## Completion Handoff

### Standard Handoff Format
When completing work, provide:
1. Completed work summary
2. Changed files list
3. Tests added/modified
4. Infrastructure verification results
5. Current phase status
6. Recommended next step
7. Important notes or caveats

### Phase Completion
- Verify the phase's definition of done is met
- Update `PROGRESS.md` to mark phase complete
- Document any technical debt carried forward
- Note any deviations from the original plan

## Prohibited Actions

### Do Not
- Implement features from future phases
- Add unnecessary infrastructure services
- Break the modular monolith architecture prematurely
- Create fake or randomized activity for the Observatory
- Commit secrets or sensitive data
- Rewrite working code outside the current scope
- Skip testing or verification
- Leave documentation outdated
- Implement authentication/RBAC before Phase 11
- Implement Kubernetes/AWS before Phase 12

## Important MiniCloud Principles

1. MiniCloud is a small functional cloud-provider learning platform
2. It uses real infrastructure locally
3. It is not intended to reimplement AWS, Kubernetes, Kafka, Redis, PostgreSQL, or Docker
4. Start with a Go modular monolith and Next.js console
5. Introduce independent services only when later phases genuinely require them
6. Prefer real system behavior over fake simulations
7. The Live Engineering Observatory must visualize real backend activity
8. Do not fabricate/randomize activity merely to make the UI look alive
9. `packages/openapi/openapi.yaml` is the HTTP API source of truth
10. Async event contracts are versioned
11. Correlation information should be preserved where appropriate
12. Infrastructure boundaries must be respected
13. Keep each phase incremental and understandable
14. Do not prematurely implement future phases
15. Avoid unnecessary abstractions and over-engineering
