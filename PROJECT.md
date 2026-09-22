# MiniCloud Project Specification

MiniCloud is a small, functional cloud-provider learning platform. Users will eventually deploy workloads and use storage, databases, caches, queues, and events while observing genuine activity and failures in real time.

## Goals and boundaries

MiniCloud favors real underlying technology, simple explanations, observable behavior, and runnable phases. It is not an AWS, Kubernetes, Kafka, Redis, or database reimplementation. The initial architecture is a Go modular monolith and a Next.js console; independent services are introduced only when later phases need them.

## Technology and architecture

The console uses Next.js, TypeScript, Tailwind, shadcn/ui, React Flow, Recharts, and WebSocket. The control plane uses Go, REST, WebSocket, OpenAPI 3.1, pgx/sqlc/PostgreSQL, Redis, Kafka, MinIO, Docker, and OpenTelemetry. Local dependencies run through Docker Compose. Prometheus, Grafana, Loki, Tempo, and the OTel Collector provide operational telemetry.

`packages/openapi/openapi.yaml` is the HTTP source of truth. Async events will use versioned contracts under `packages/contracts/events`, with event IDs, request/trace correlation, source, subject, timestamp, status, and payload. The Phase 2 Observatory will publish real normalized activity events to the console; it must not animate fabricated events.

## Roadmap

Phase 0 creates the foundation. Phase 1 adds core control-plane resources. Phase 2 adds the console and live observatory. Phases 3–9 progressively add cache, events, queue, storage, compute, scheduling, load balancing, autoscaling, and self-healing. Phase 10 completes observability; Phase 11 adds security/testing/CI maturity; Phase 12 explores Kubernetes, AWS, and Terraform after the local system is stable.
