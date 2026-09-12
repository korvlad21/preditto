# Backend

Verified on 2026-09-12: PostgreSQL schema migrations are owned by the separate `migrate` service in [Compose](../../../docker-compose.yaml). The backend waits for successful migration completion; migrations wait for the existing PostgreSQL healthcheck. No Go migration runner is used.

The schema, trigger behavior, environment contract, and manual commands are documented in [the migration README](../../../backend/migrations/README.md). SQL files live in `backend/migrations`; migrate tracks their version in `schema_migrations`.

Isolated Docker checks with PostgreSQL 18.6 and migrate v4.20.1 verified initial apply, repeated up, column definitions, identity generation, uniqueness, null rejection, length limits, updated_at behavior, down 1, and reapply. A lightweight backend probe verified the Compose dependency gate; the actual Go API was not rebuilt or started in this iteration.
