# ADR 006: Docker-First Local Infrastructure

Stateful local dependencies run in Docker Compose with named volumes, health checks, and one private bridge network. Applications initially run on the developer machine for rapid iteration.
