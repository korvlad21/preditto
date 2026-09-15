# Docker Command Patterns

<a id="pattern-20260912-001"></a>
## [docker] Run PostgreSQL migrations as a Compose service

- ID: PATTERN-20260912-001
- Preflight: project root, Docker daemon available, current Compose configuration and environment checked; PostgreSQL must become healthy. SQL files mount read-only at `/migrations`.
- Commands: `docker compose config --quiet`, `docker compose run --rm migrate up`, `docker compose run --rm migrate down 1`.
- Verification: configuration validation passed on the project Compose file. Runtime commands passed using `docker compose -p <isolated-project> -f <isolated-compose.json>` with temporary PostgreSQL storage, synthetic credentials, and the actual SQL mount and migration entrypoint. Repeated up returned `no change`; down and reapply passed.
- Automatic startup: `docker compose up -d --build` follows PostgreSQL health, migrate completion, then backend startup. The dependency sequence was exercised with a lightweight backend probe; the full application build was outside this check.
- Avoid: rolling back a real database as a verification step. The first down drops users and its data. Do not assume restarting only an existing backend reruns migrations.
- Related failure: [FAILURE-20260912-001](docker-failures.md#failure-20260912-001).
- Source: [migration commands](../../../backend/migrations/README.md).

<a id="pattern-20260915-001"></a>
## [docker] Change a schema after its migration version was applied

- ID: PATTERN-20260915-001
- When to use: changing a PostgreSQL object after `schema_migrations` has recorded the version that created it.
- Preflight: inspect the current version and database objects; keep the previously applied SQL consistent with its down migration.
- Correct approach: add the next numbered `up/down` pair in `backend/migrations`, then run normal Compose migration `up` and verify the schema and `dirty=false` in PostgreSQL. Use an isolated database for apply/down/reapply checks.
- Avoid: editing an already applied `up.sql` and expecting Docker to replay it; forcing a version to apply a routine schema change.
- Project note: version 7 retained `roles.updated_at` despite an edit to migration 4. In this local development checkout the RBAC migrations were still untracked and all four RBAC tables were empty, so the application was stopped, versions 7-4 were rolled back, migration 4 was revised, and 4-7 reapplied. Version 7 now creates roles without `updated_at`; do not use this rewrite for shared databases or populated RBAC tables.
- Related failure: [FAILURE-20260915-001](docker-failures.md#failure-20260915-001).
