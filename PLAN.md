# MiniCloud Delivery Plan

| Phase | Objective | Definition of done |
| --- | --- | --- |
| 0 — Foundation | Monorepo, API/console skeletons, Compose dependencies, docs, contracts, CI | A fresh clone can configure, start infrastructure, run both apps, and verify health/readiness. |
| 1 — Core Control Plane | Users, projects, applications, deployments | Migrations, contract-backed API, validation, and tests are runnable. |
| 2 — Console + Observatory | Real resource views, WebSocket activity stream, topology base | UI displays genuine correlated state/activity. |
| 3 — MiniCache | Redis cache abstraction | Hit/miss/set behavior is real and observable. |
| 4 — MiniEvents | Kafka topic/producer/consumer abstraction | Offsets, retries, ordering, and lag are visible. |
| 5 — MiniQueue + Workers | Redis Streams worker jobs | Retry, backoff, dead letter, and lifecycle work. |
| 6 — MiniStorage | MinIO S3-style objects | Bucket/object operations and metadata work. |
| 7 — MiniCompute | Docker workloads via node-agent | Deploy/control/log/health flows work safely. |
| 8 — Scheduler + Networking | Placement, discovery, routing | Healthy workloads are scheduled and routed. |
| 9 — Autoscaling + Self-Healing | Reconciliation and replacement | Desired/actual state, scale, and recovery work. |
| 10 — Observability | Full logs, metrics, traces, dashboards | Signals correlate across operational and product views. |
| 11 — Security + Quality | Auth, RBAC, testing, CI/CD | Security boundaries and broad automated checks pass. |
| 12 — Kubernetes + AWS + Terraform | Terraform and optional cloud mapping | Local concepts map reproducibly to cloud infrastructure. |

Each phase depends on the preceding phase, must remain locally runnable, include proportionate tests, and update documentation. Phase progress is tracked in `PROGRESS.md`.

## Phase Details

### Phase 0 — Foundation
**Objective**: Establish the development environment and project structure.

**What will be built**:
- Monorepo structure with pnpm workspaces and Go workspace
- API skeleton with health/readiness endpoints and middleware
- Console skeleton with basic Next.js setup
- Docker Compose configuration for local infrastructure
- OpenAPI contract seed with health endpoints
- CI pipeline for backend, frontend, and infrastructure validation
- Architecture Decision Records (ADRs) for key decisions
- Documentation (PROJECT.md, PLAN.md, PROGRESS.md, AGENTS.md, README.md)

**Definition of done**: A fresh clone can configure, start infrastructure, run both apps, and verify health/readiness.

### Phase 1 — Core Control Plane
**Objective**: Implement core resource management for users, projects, applications, and deployments.

**What will be built**:
- Database schema for users, projects, applications, deployments, audit_logs, outbox_events
- Migration system with transaction support
- SQLC configuration and type-safe database queries
- CRUD API endpoints for all resources
- Input validation and error handling
- Basic transaction support for write operations
- Audit logging infrastructure
- Outbox event infrastructure
- OpenAPI schemas for all endpoints
- Handler tests and integration tests
- Console UI for resource management

**Definition of done**: Migrations, contract-backed API, validation, and tests are runnable.

### Phase 2 — Console + Live Engineering Observatory
**Objective**: Build the console UI and real-time activity visualization.

**What will be built**:
- Real resource views (users, projects, applications, deployments)
- WebSocket activity stream
- Real-time topology visualization
- Activity correlation and filtering
- Console integration with Phase 1 APIs

**Definition of done**: UI displays genuine correlated state/activity.

### Phase 3 — MiniCache
**Objective**: Implement Redis cache abstraction.

**What will be built**:
- Redis client integration
- Cache abstraction layer
- Hit/miss/set/delete operations
- Cache metrics and observability

**Definition of done**: Hit/miss/set behavior is real and observable.

### Phase 4 — MiniEvents
**Objective**: Implement Kafka event streaming.

**What will be built**:
- Kafka topic management
- Event producer and consumer abstractions
- Offset management and consumer groups
- Event contracts and versioning
- Retry and error handling
- Lag monitoring

**Definition of done**: Offsets, retries, ordering, and lag are visible.

### Phase 5 — MiniQueue + Workers
**Objective**: Implement job queue with Redis Streams.

**What will be built**:
- Redis Streams queue implementation
- Worker pool management
- Job retry with exponential backoff
- Dead letter queue
- Job lifecycle tracking

**Definition of done**: Retry, backoff, dead letter, and lifecycle work.

### Phase 6 — MiniStorage
**Objective**: Implement MinIO S3-style object storage.

**What will be built**:
- MinIO client integration
- Bucket operations (create, delete, list)
- Object operations (put, get, delete, list)
- Metadata management
- Storage metrics

**Definition of done**: Bucket/object operations and metadata work.

### Phase 7 — MiniCompute
**Objective**: Implement Docker workload management via node-agent.

**What will be built**:
- Node-agent service for Docker operations
- Workload deployment API
- Container lifecycle management
- Log streaming
- Health checks
- Node registration and discovery

**Definition of done**: Deploy/control/log/health flows work safely.

### Phase 8 — Scheduler + Networking
**Objective**: Implement workload scheduling and load balancing.

**What will be built**:
- Scheduler for workload placement
- Service discovery
- Load balancer integration
- Health-based routing
- Capacity management

**Definition of done**: Healthy workloads are scheduled and routed.

### Phase 9 — Autoscaling + Self-Healing
**Objective**: Implement automatic scaling and failure recovery.

**What will be built**:
- Reconciliation loop for desired vs actual state
- Autoscaling policies (metric-based, schedule-based)
- Failure detection and replacement
- Rolling updates
- State validation

**Definition of done**: Desired/actual state, scale, and recovery work.

### Phase 10 — Observability
**Objective**: Complete the observability stack.

**What will be built**:
- Comprehensive logging with structured formats
- Application metrics (business and operational)
- Distributed tracing across all services
- Grafana dashboards for all signals
- Alert integration
- Log and trace correlation

**Definition of done**: Signals correlate across operational and product views.

### Phase 11 — Security + Quality
**Objective**: Implement security, testing, and CI/CD maturity.

**What will be built**:
- Authentication and authorization
- Role-based access control (RBAC)
- Security audit logging
- Comprehensive test suite (unit, integration, E2E)
- CI/CD pipeline enhancements
- Security scanning
- Performance testing

**Definition of done**: Security boundaries and broad automated checks pass.

### Phase 12 — Kubernetes + AWS + Terraform
**Objective**: Explore cloud deployment options.

**What will be built**:
- Terraform configurations for infrastructure
- Kubernetes deployment manifests
- AWS service integration (optional)
- Cloud-native adaptations
- Documentation for cloud deployment

**Definition of done**: Local concepts map reproducibly to cloud infrastructure.
