# MiniCloud

MiniCloud is a small, educational cloud engineering platform. It uses real infrastructure locally while keeping the control plane intentionally small. Its eventual differentiator is a Live Engineering Observatory that visualizes real requests, events, workloads, and recovery.

## Current status

Phase 0 — Foundation and Development Environment.

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

Read [PROJECT.md](PROJECT.md), [PLAN.md](PLAN.md), and [PROGRESS.md](PROGRESS.md) before implementing further phases.
