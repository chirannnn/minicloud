# ADR 001: Monorepo

MiniCloud uses pnpm workspaces, Turborepo, and a Go workspace. This keeps the console, API, contracts, and infrastructure changes coordinated while allowing each application to run independently.
