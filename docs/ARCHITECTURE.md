# Architecture baseline v1.0

Modular monolith (Go 1.25+, PostgreSQL, Redis, REST `/api/v1`).

## Layers per bounded context

- `domain` — entities, value objects, repository ports
- `application` — use cases
- `adapter/http` — REST delivery
- `adapter/persistence` — PostgreSQL implementations
- `adapter/eventhandler` — in-process event subscribers (achievement, activity)

## Composition root

`internal/platform/app` wires config, postgres, redis, event bus, authorization, HTTP server, and route registration.

## Operations

- API: `cmd/api`
- Migrations: `cmd/migrate` (golang-migrate, SQL under `migrations/`)
- Docker: [`docker-compose.yml`](./docker-compose.yml), образ — [`Dockerfile`](./Dockerfile)

Full endpoint and table lists are defined in the project architecture baseline (v1.0).

Database design: see [SCHEMA.md](./SCHEMA.md) and [INDEXES.md](./INDEXES.md).
