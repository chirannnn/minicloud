# MiniCloud OpenAPI Contract

`openapi.yaml` is the source of truth for MiniCloud HTTP transport contracts. Phase 0 defines only health and readiness endpoints. Future phases add APIs incrementally and generate Go and TypeScript transport types from this contract.

Run `pnpm --filter @minicloud/openapi validate` to validate it.
