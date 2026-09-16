# Backend

The `preditto/seeds/development` [seed package](../../../backend/seeds/README.md) owns repeatable initial data and is run by [cmd/seed](../../../backend/cmd/seed/main.go). It reuses config and database packages under the `preditto` Go module. Compose now gates startup in the order PostgreSQL healthy, migrate success, seed success, backend. The seed service shares the backend build and environment. The runner executes countries, teams, users, user_info, roles, and user_roles in one transaction. It upserts the specified 36 teams by unique slug, matches the three users by username or email with bcrypt password hashes, and upserts profiles using actual user IDs with requested favorite team IDs 5/19 and SQL NULL avatars; its input validation, preservation rules, extension pattern, and manual command are documented in the package README.

[Migration 000004](../../../backend/migrations/000004_create_user_info.up.sql) adds user_info. Its unique, required user_id references users with ON DELETE CASCADE; nullable favorite_team_id references teams with ON DELETE SET NULL. The up migration creates the table, `BEFORE UPDATE` trigger, and function in one transaction. The [down migration](../../../backend/migrations/000004_create_user_info.down.sql) removes the trigger, table, and function. The current database is at version 8.

Verified on 2026-09-12: PostgreSQL schema migrations are owned by the separate `migrate` service in [Compose](../../../docker-compose.yaml). Seeds wait for successful migration completion; migrations wait for the existing PostgreSQL healthcheck. No Go migration runner is used.

The schema, trigger behavior, environment contract, and manual commands are documented in [the migration README](../../../backend/migrations/README.md). SQL files live in `backend/migrations`; migrate tracks their version in `schema_migrations`.

[RBAC migrations 000005-000008](../../../backend/migrations/000005_create_roles.up.sql) create roles, permissions, user_roles, and role_permissions. Roles have no `updated_at` column or trigger. The current Compose database is at version 8 with dirty=false.

[Migration 000003](../../../backend/migrations/000003_create_teams.up.sql) adds teams after countries exists, with a required country code referencing countries.short_name. It also has an identity primary key, required name and unique slug, nullable short_name and logo_url, and a microsecond-precision created_at default. Its [down migration](../../../backend/migrations/000003_create_teams.down.sql) drops teams before migration 2 drops countries.

Earlier isolated Docker checks with PostgreSQL 18.6 and migrate v4.20.1 verified the initial migrations and a lightweight dependency probe. On 2026-09-15, normal Compose startup built and started the Go API; `/health` returned `{"status":"ok"}`.

Verified on 2026-09-16: [role seeds](../../../backend/seeds/development/roles.go) upsert names by code; [assignment seeds](../../../backend/seeds/development/user_roles.go) resolve users by username and roles by code. [Repository functions](../../../backend/internal/repository/user_roles.go) own idempotent assignment/removal SQL and accept either a database or an existing transaction. No service layer exists in the inspected backend. Targeted Go tests and vet passed; PostgreSQL integration coverage was added but not executed in this iteration.

Countries: [migration 000002](../../../backend/migrations/000002_create_countries.up.sql) creates the table with a required unique short_name. The [countries seed](../../../backend/seeds/development/countries.go) validates 16 country entries and upserts names by three-letter uppercase code, preserving IDs and created_at. PostgreSQL seed integration tests pass in disposable schemas.

The [team seed](../../../backend/seeds/development/teams.go) assigns country codes to all 36 teams. On 2026-09-16, a fresh isolated schema passed migration 1-8 up and down, and PostgreSQL seed integration tests passed. The local database was rebuilt from the reordered migrations and reports version 8, dirty=false, with 36 valid team country links.
