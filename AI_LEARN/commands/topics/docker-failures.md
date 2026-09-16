# Docker Command Failures

<a id="failure-20260912-001"></a>
## [docker] Empty PostgreSQL URL loses its driver syntax

- ID: FAILURE-20260912-001
- Date: 2026-09-12
- Context: Docker Compose v2.6.1, migrate/migrate:v4.20.1, PostgreSQL 18.6; temporary isolated project, root working directory, no application Go execution.
- Original invocation: `docker compose -p <isolated-project> -f <isolated-compose.json> up -d backend`; migration entrypoint used `migrate -path /migrations -database postgres:// up` with PG environment variables.
- Symptom / exit code: migrate exited 1 with `missing "=" after "postgres:" in connection info string`; Compose blocked the dependent backend and exited 1.
- Cause / incorrect assumption: URL normalization removes empty authority markers from `postgres://`, yielding `postgres:`; that is then treated as malformed keyword connection information. An empty URL authority alone does not preserve the expected syntax.
- Correction: use `postgres:///`, whose slash preserves URL syntax, with connection settings in `PGHOST`, `PGPORT`, `PGUSER`, `PGPASSWORD`, `PGDATABASE`, and `PGSSLMODE`.
- Verification: same Compose startup invocation succeeded after correction. Real up, no-change up, down 1, and reapply passed, including a password containing URL-reserved characters.
- Rule: preserve the trailing slash when using a PostgreSQL URL with environment-only settings. Keep passwords out of URL construction.
- Related pattern: [PATTERN-20260912-001](docker-patterns.md#pattern-20260912-001).

<a id="failure-20260915-001"></a>
## [docker] Dirty version after failed transactional SQL

- ID: FAILURE-20260915-001
- Date: 2026-09-15
- Context: PostgreSQL 18.6, `golang-migrate` v4.20.1, existing Compose database.
- Original invocation: `docker compose up -d --build` with a trailing comma in `000004_create_roles.up.sql`.
- Symptom / exit code: migration 4 failed at `CREATE TABLE roles`; subsequent Compose startup exited 1 with `Dirty database version 4`.
- Cause: PostgreSQL rolled back the failed migration's SQL, while `schema_migrations` recorded attempted version 4 with `dirty=true`.
- Correct invocation: fix migration 4; audit version, roles objects, and existing users; rehearse up/down on isolated PostgreSQL. Only after confirming version-4 objects were absent, run `docker compose run --rm migrate force 3`, then `docker compose run --rm migrate up`.
- Verification: the existing database reached version 7 with `dirty=false`; version 8 was later applied, then rolled back during the empty local RBAC baseline rewrite. Existing user count remained three.
- Rule: never clear `dirty` or delete a volume before verifying the last fully applied version and repairing the failed SQL.
- Related pattern: [PATTERN-20260912-001](docker-patterns.md#pattern-20260912-001).

<a id="failure-20260915-002"></a>
## [docker] Seed SQL contains an unintended leading character

- ID: FAILURE-20260915-002
- Date: 2026-09-15
- Context: normal Compose startup after migrations succeeded; seeds run in one transaction.
- Original invocation: `docker compose up -d --build`.
- Symptom / exit code: seed exited 1 with `pq: syntax error at or near "а" at position 1:1`; backend dependency remained blocked.
- Cause: a Cyrillic `а` followed the opening SQL raw-string delimiter in `backend/seeds/development/user_info.go`.
- Correct invocation: remove the stray character from the SQL text, validate UTF-8, and rerun `docker compose up -d --build`.
- Verification: seed exited 0, backend and frontend started, and `/health` returned `{"status":"ok"}`; the failed seed transaction rolled back.
- Rule: when migration succeeds but Compose startup stops, inspect the next dependency's logs before changing migration state.

<a id="failure-20260915-003"></a>
## [docker] Migration 3 down leaves its trigger function

- ID: FAILURE-20260915-003
- Date: 2026-09-15
- Context: temporary isolated PostgreSQL; fresh migration version 7, no application data.
- Original invocation: `docker compose -p <isolated-project> -f <isolated-compose.json> run --rm migrate down 5` followed by `up`.
- Symptom / exit code: down succeeded but reached version 2, not the intended version 3; reapply exited 1 because `set_user_info_updated_at()` already existed.
- Cause: five steps from version 7 include migration 3. Its current down SQL drops `user_info` but not its standalone trigger function, while its up SQL creates that function.
- Correct invocation: for the RBAC-only rollback from version 7, use `down 4` to stop at version 3. After resetting the disposable isolated database, fresh up, down 4, and reapply succeeded. The existing application database never rolled back version 3.
- Rule: compute the rollback count from the current version before running down. Do not use a blind `force` to mask a leftover function.
- Resolution (2026-09-16): `000003_create_user_info.down.sql` now drops the trigger before the table and drops the function afterward, both with `IF EXISTS`. Isolated PostgreSQL execution verified cleanup with and without these objects, then reapplication of the current up migration. Do not roll back a populated live database solely to verify this historical case.
- Related pattern: [PATTERN-20260915-001](docker-patterns.md#pattern-20260915-001).

<a id="failure-20260916-001"></a>
## [docker] Dirty migration 3 with partially committed user_info

- Date: 2026-09-16.
- Context: Existing Compose PostgreSQL database, migration version 3 with dirty=true; Compose migrate service exited 1 and blocked seeds and application startup.
- Symptom: `docker compose run --rm migrate up` reported `Dirty database version 3. Fix and force version.` The original migration error was not present in the available current logs.
- Observed schema: `user_info` existed with all expected columns and constraints but zero rows; `set_user_info_updated_at()` and `user_info_set_updated_at` were absent. The up migration committed the table before creating the function and trigger, allowing a partial state if a later statement failed.
- Correction: Verified the table was empty, removed only that partial table and any optional trigger/function, ran `docker compose run --rm migrate force 2`, then `docker compose run --rm migrate up`. Moved migration 3's COMMIT after trigger creation so future application is atomic.
- Verification: Migration reached version 8 with dirty=false; isolated up/down/up/down passed; Compose seed exited 0, backend became healthy, and expected seeded row counts were present.
- Rule: `force` changes version metadata only. Inspect data and objects first, repair the partial schema, then select the last fully applied version. Do not drop a populated table as part of this recipe.
