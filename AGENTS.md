# MiniCloud Agent Instructions

Before modifying code, read `PROJECT.md`, `PLAN.md`, and `PROGRESS.md`, then inspect relevant files. Implement only the user-requested current phase; do not pre-build future components.

- Preserve the modular Go-monolith and contract-first architecture.
- Keep changes small, testable, and documented. Avoid unnecessary services, abstractions, and large catch-all files.
- Never commit secrets; use `.env.example` for documented safe local defaults.
- Run relevant formatting, linting, tests, builds, OpenAPI checks, and infrastructure checks.
- Update `PROGRESS.md` only with facts actually verified. Add ADRs for meaningful architecture changes.
- Do not rewrite working code outside scope. End work with the standard handoff: completed work, changed files, tests, infrastructure verification, current phase, next step, and important notes.
