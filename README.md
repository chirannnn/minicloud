# MiniCloud

MiniCloud is a small, educational cloud engineering platform. It uses real infrastructure locally while keeping the control plane intentionally small. Its eventual differentiator is a Live Engineering Observatory that visualizes real requests, events, workloads, and recovery.

## Current status

Phase 1 — Core Control Plane (in progress)

Phase 0 (Foundation) is complete. Phase 1 adds users, projects, applications, and deployments with database migrations, API endpoints, and validation.

## Requirements

- Docker Desktop with Docker Compose
- Go 1.25+
- Node.js 22+
- pnpm 10+
- GNU Make or a compatible `make` implementation

## Setup

```text
copy .env.example .env        # Windows PowerShell: Copy-Item .env.example .env
pnpm install
docker compose up -d
go run ./apps/api/cmd/minicloud
pnpm --filter @minicloud/console dev
```

The API is available at `http://localhost:8080`; the console is at `http://localhost:3000`.

## Local infrastructure

| Service | URL / port |
| --- | --- |
| PostgreSQL | localhost:5432 |
| Redis | localhost:6379 |
| Kafka | localhost:9092 |
| Kafka UI | http://localhost:8081 |
| MinIO API / Console | http://localhost:9000 / http://localhost:9001 |
| Mailpit | http://localhost:8025 |
| Prometheus | http://localhost:9090 |
| Grafana | http://localhost:3001 |
| Loki / Tempo | localhost:3100 / http://localhost:3200 |
| OTel Collector | localhost:4317, localhost:4318 |

## Useful commands

`make up`, `make down`, `make logs`, `make api`, `make console`, `make test`, `make lint`, `make generate`, and `make verify` are available where Make is installed. The equivalent direct commands are documented in the Makefile.

## Architecture

```text
Next.js Console → Go API → PostgreSQL / Redis / Kafka / MinIO
                       └→ OTel Collector → Prometheus / Loki / Tempo → Grafana
```

The control plane is a Go modular monolith. Independent services are introduced only when later phases require them.

## High-level roadmap

- **Phase 0**: Foundation (complete) - Monorepo, API/console skeletons, infrastructure, CI
- **Phase 1**: Core Control Plane (in progress) - Users, projects, applications, deployments
- **Phase 2**: Console + Observatory - Real resource views, WebSocket activity stream
- **Phase 3-9**: Progressive infrastructure services (cache, events, queue, storage, compute, scheduling, autoscaling)
- **Phase 10-12**: Observability maturity, security/testing, and cloud exploration

Read [PROJECT.md](PROJECT.md) for the project specification, [PLAN.md](PLAN.md) for detailed phase definitions, [PROGRESS.md](PROGRESS.md) for current implementation status, and [AGENTS.md](AGENTS.md) for AI agent instructions.
