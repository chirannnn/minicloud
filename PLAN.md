# MiniCloud Delivery Plan

| Phase | Objective | Definition of done |
| --- | --- | --- |
| 0 — Foundation | Monorepo, API/console skeletons, Compose dependencies, docs, contracts, CI | A fresh clone can configure, start infrastructure, run both apps, and verify health/readiness. |
| 1 — Core Control Plane | Users, projects, applications, deployments | Migrations, contract-backed API, validation, and tests are runnable. |
| 2 — Console + Observatory | Real resource views, WebSocket activity stream, topology base | UI displays genuine correlated state/activity. |
| 3 — MiniCache | Redis cache abstraction | Hit/miss/set behavior is real and observable. |
| 4 — MiniEvents | Kafka topic/producer/consumer abstraction | Offsets, retries, ordering, and lag are visible. |
| 5 — MiniQueue | Redis Streams worker jobs | Retry, backoff, dead letter, and lifecycle work. |
| 6 — MiniStorage | MinIO S3-style objects | Bucket/object operations and metadata work. |
| 7 — MiniCompute | Docker workloads via node-agent | Deploy/control/log/health flows work safely. |
| 8 — Scheduler + LB | Placement, discovery, routing | Healthy workloads are scheduled and routed. |
| 9 — Autoscaling + Healing | Reconciliation and replacement | Desired/actual state, scale, and recovery work. |
| 10 — Observability | Full logs, metrics, traces, dashboards | Signals correlate across operational and product views. |
| 11 — Security + Quality | Auth, RBAC, testing, CI/CD | Security boundaries and broad automated checks pass. |
| 12 — Kubernetes/AWS | Terraform and optional cloud mapping | Local concepts map reproducibly to cloud infrastructure. |

Each phase depends on the preceding phase, must remain locally runnable, include proportionate tests, and update documentation. Detailed Phase 0 tasks are tracked in `PROGRESS.md`.
