# MiniCloud Project Specification

MiniCloud is a small, functional cloud-provider learning platform. Users will eventually deploy workloads and use storage, databases, caches, queues, and events while observing genuine activity and failures in real time.

## Goals and boundaries

MiniCloud favors real underlying technology, simple explanations, observable behavior, and runnable phases. It is not an AWS, Kubernetes, Kafka, Redis, or database reimplementation. The initial architecture is a Go modular monolith and a Next.js console; independent services are introduced only when later phases need them.

## Technology and architecture

### Control Plane
- Go 1.25+ modular monolith
- REST API with OpenAPI 3.1 contract-first design
- PostgreSQL with pgx/v5 and sqlc for type-safe database operations
- OpenTelemetry for distributed tracing
- Standard library HTTP server with custom middleware

### Console
- Next.js 15 with React 19
- TypeScript
- Tailwind CSS (planned, shadcn/ui, React Flow, Recharts for later phases)
- WebSocket (planned for Phase 2 Live Observatory)

### Infrastructure
- Docker Compose for local development
- PostgreSQL 17 for persistent storage
- Redis 7 (planned for Phase 3 MiniCache)
- Kafka 3.9.0 (planned for Phase 4 MiniEvents)
- MinIO (planned for Phase 6 MiniStorage)
- OTel Collector, Prometheus, Grafana, Loki, Tempo for observability

### Architectural Principles
- `packages/openapi/openapi.yaml` is the HTTP API source of truth
- Async events will use versioned contracts under `packages/contracts/events`, with event IDs, request/trace correlation, source, subject, timestamp, status, and payload
- The Phase 2 Live Engineering Observatory will publish real normalized activity events to the console; it must not animate fabricated events
- Prefer real system behavior over fake simulations
- Keep each phase incremental and understandable
- Avoid unnecessary abstractions and over-engineering

## Project structure

```
minicloud/
├── apps/
│   ├── api/                    # Go modular monolith
│   │   ├── cmd/minicloud/      # Main entry point
│   │   ├── internal/
│   │   │   ├── config/         # Configuration
│   │   │   ├── database/       # Database connection & migrations
│   │   │   ├── logger/         # Logging
│   │   │   ├── modules/        # Feature modules (controlplane, etc.)
│   │   │   ├── observability/  # OpenTelemetry
│   │   │   ├── router/         # HTTP routing & middleware
│   │   │   └── server/         # HTTP server
│   │   ├── migrations/         # SQL migrations
│   │   ├── queries/            # SQLC queries
│   │   └── sqlc.yaml           # SQLC configuration
│   └── console/                # Next.js console
│       └── app/                # Next.js app directory
├── packages/
│   └── openapi/                # OpenAPI contract
├── infrastructure/
│   └── docker/                 # Infrastructure configs
├── docs/
│   └── decisions/              # Architecture Decision Records
└── .github/workflows/          # CI configuration
```

## Development philosophy

- Small, testable, documented changes
- Contract-first development (OpenAPI, SQLC)
- Modular internal structure within the monolith
- Real infrastructure over mocks where possible
- Observability as a first-class concern
- Security and quality treated as cross-cutting concerns

## Live Engineering Observatory

The Live Engineering Observatory is MiniCloud's key differentiator. It visualizes real backend activity including:
- HTTP requests and responses
- Database operations
- Event publishing and consumption
- Workload deployments and status changes
- System failures and recovery

The Observatory must display genuine correlated activity from the running system, not fabricated or randomized events for visual effect.

## Data and event principles

- All changes to persistent state should generate audit events
- Use outbox pattern for reliable event publishing
- Preserve correlation information: request ID, trace ID, event ID, causation ID
- Events are versioned with backward-compatible contracts
- Event consumers should be idempotent

## Roadmap

Phase 0 creates the foundation. Phase 1 adds core control-plane resources. Phase 2 adds the console and live observatory. Phases 3–9 progressively add cache, events, queue, storage, compute, scheduling, load balancing, autoscaling, and self-healing. Phase 10 completes observability; Phase 11 adds security/testing/CI maturity; Phase 12 explores Kubernetes, AWS, and Terraform after the local system is stable.

## Cross-phase rules

- Each phase depends on the preceding phase
- Each phase must remain locally runnable
- Each phase must include proportionate tests
- Each phase must update documentation
- Do not implement features from future phases prematurely
- Respect infrastructure boundaries (e.g., Docker socket access belongs to node-agent, not API)
- Maintain the modular monolith until a genuine runtime boundary requires separation

## Definition of done

A phase is complete when:
- All specified functionality is implemented
- Tests are passing and provide meaningful coverage
- Documentation is updated to reflect the changes
- The system remains locally runnable
- CI checks pass
- Manual verification confirms the intended behavior
