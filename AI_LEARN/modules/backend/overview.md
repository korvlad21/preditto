# Backend

[Migration 000003](../../../backend/migrations/000003_create_user_info.up.sql) adds user_info. Its unique, required user_id references users with ON DELETE CASCADE; nullable favorite_team_id references teams with ON DELETE SET NULL. updated_at has only a CURRENT_TIMESTAMP default, with no application trigger; future Go updates must maintain it explicitly. The [down migration](../../../backend/migrations/000003_create_user_info.down.sql) drops user_info and its owned identity sequence. Isolated PostgreSQL checks verified schema, constraints, both deletion rules, unchanged updated_at during an ordinary update, and rollback to version 2 preserving users and teams.

Verified on 2026-09-12: PostgreSQL schema migrations are owned by the separate `migrate` service in [Compose](../../../docker-compose.yaml). The backend waits for successful migration completion; migrations wait for the existing PostgreSQL healthcheck. No Go migration runner is used.

The schema, trigger behavior, environment contract, and manual commands are documented in [the migration README](../../../backend/migrations/README.md). SQL files live in `backend/migrations`; migrate tracks their version in `schema_migrations`.

[Migration 000002](../../../backend/migrations/000002_create_teams.up.sql) adds teams with an identity primary key, required name and unique slug, nullable short_name and logo_url, and a microsecond-precision created_at default. Its [down migration](../../../backend/migrations/000002_create_teams.down.sql) drops only teams and its owned identity sequence. Isolated PostgreSQL checks verified schema, constraints, defaults, apply to version 2, and rollback to version 1 with users preserved.

Isolated Docker checks with PostgreSQL 18.6 and migrate v4.20.1 verified initial apply, repeated up, column definitions, identity generation, uniqueness, null rejection, length limits, updated_at behavior, down 1, and reapply. A lightweight backend probe verified the Compose dependency gate; the actual Go API was not rebuilt or started in this iteration.
