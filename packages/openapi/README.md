# MiniCloud OpenAPI Contract

`openapi.yaml` is the source of truth for MiniCloud HTTP transport contracts. Phase 0 defined health and readiness endpoints. Phase 1 adds control plane APIs for users, projects, applications, and deployments. Future phases will add APIs incrementally and generate Go and TypeScript transport types from this contract.

Run `pnpm --filter @minicloud/openapi validate` to validate it.
