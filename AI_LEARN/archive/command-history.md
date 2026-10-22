# Historical command records

These dated records preserve prior executions, including old module names, migration numbering and environment versions. They are not current runbooks. Use the [command index](../commands/index.md) for current guidance; do not replay historical recovery commands without inspecting the current database.

# Docker Command Patterns

<a id="pattern-20260912-001"></a>
## [docker] Run PostgreSQL migrations as a Compose service

- ID: PATTERN-20260912-001
- Preflight: project root, Docker daemon available, current Compose configuration and environment checked; PostgreSQL must become healthy. SQL files mount read-only at `/migrations`.
- Commands: `docker compose config --quiet`, `docker compose run --rm migrate up`, `docker compose run --rm migrate down 1`.
- Verification: configuration validation passed on the project Compose file. Runtime commands passed using `docker compose -p <isolated-project> -f <isolated-compose.json>` with temporary PostgreSQL storage, synthetic credentials, and the actual SQL mount and migration entrypoint. Repeated up returned `no change`; down and reapply passed.
- Automatic startup: `docker compose up -d --build` follows PostgreSQL health, migrate completion, then backend startup. The dependency sequence was exercised with a lightweight backend probe; the full application build was outside this check.
- Avoid: rolling back a real database as a verification step. The first down drops users and its data. Do not assume restarting only an existing backend reruns migrations.
- Related failure: [FAILURE-20260912-001](command-history.md#failure-20260912-001).
- Source: [migration commands](../../backend/migrations/README.md).

<a id="pattern-20260915-001"></a>
## [docker] Change a schema after its migration version was applied

- ID: PATTERN-20260915-001
- When to use: changing a PostgreSQL object after `schema_migrations` has recorded the version that created it.
- Preflight: inspect the current version and database objects; keep the previously applied SQL consistent with its down migration.
- Correct approach: add the next numbered `up/down` pair in `backend/migrations`, then run normal Compose migration `up` and verify the schema and `dirty=false` in PostgreSQL. Use an isolated database for apply/down/reapply checks.
- Avoid: editing an already applied `up.sql` and expecting Docker to replay it; forcing a version to apply a routine schema change.
- Project note: version 7 retained `roles.updated_at` despite an edit to migration 4. In this local development checkout the RBAC migrations were still untracked and all four RBAC tables were empty, so the application was stopped, versions 7-4 were rolled back, migration 4 was revised, and 4-7 reapplied. Version 7 now creates roles without `updated_at`; do not use this rewrite for shared databases or populated RBAC tables.
- Related failure: [FAILURE-20260915-001](command-history.md#failure-20260915-001).


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
- Related pattern: [PATTERN-20260912-001](../commands/topics/docker.md#pattern-20260912-001).

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
- Related pattern: [PATTERN-20260912-001](../commands/topics/docker.md#pattern-20260912-001).

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
- Resolution (2026-09-16): the user_info down migration now drops the trigger before the table and drops the function afterward, both with `IF EXISTS`. It was renumbered to `000004` during the country/teams reorder. Isolated PostgreSQL execution verified cleanup with and without these objects. Do not roll back a populated live database solely to verify this historical case.
- Related pattern: [PATTERN-20260915-001](../commands/topics/docker.md#pattern-20260915-001).

<a id="failure-20260916-001"></a>
## [docker] Dirty migration 3 with partially committed user_info

- Date: 2026-09-16.
- Context: Existing Compose PostgreSQL database, migration version 3 with dirty=true; Compose migrate service exited 1 and blocked seeds and application startup.
- Symptom: `docker compose run --rm migrate up` reported `Dirty database version 3. Fix and force version.` The original migration error was not present in the available current logs.
- Observed schema: `user_info` existed with all expected columns and constraints but zero rows; `set_user_info_updated_at()` and `user_info_set_updated_at` were absent. The up migration committed the table before creating the function and trigger, allowing a partial state if a later statement failed.
- Correction: Verified the table was empty, removed only that partial table and any optional trigger/function, ran `docker compose run --rm migrate force 2`, then `docker compose run --rm migrate up`. Moved migration 3's COMMIT after trigger creation so future application is atomic.
- Verification: Migration reached version 8 with dirty=false; isolated up/down/up/down passed; Compose seed exited 0, backend became healthy, and expected seeded row counts were present.
- Rule: `force` changes version metadata only. Inspect data and objects first, repair the partial schema, then select the last fully applied version. Do not drop a populated table as part of this recipe.

<a id="failure-20260916-002"></a>
## [docker] Teams referenced countries before countries existed

- Date: 2026-09-16.
- Symptom: migration 2 failed with `pq: relation "countries" does not exist`; the later transaction-aborted advisory unlock message was secondary.
- Cause: teams migration 2 had been changed to add a country foreign key, while countries was created only in migration 8.
- Corrected project layout: countries is migration 2, teams is migration 3 with the foreign key, and the prior migrations 3-7 are renumbered 4-8. Team seeds provide valid country codes.
- Verification: migration 1-8 up/down passed in an isolated PostgreSQL schema; PostgreSQL seed integration tests passed. The local database was backed up, rebuilt using the reordered files, and reached version 8 with dirty=false and valid team country links.
- Rule: referenced tables must be created before foreign keys. Renumbering migrations already recorded in schema_migrations requires a deliberate data-preserving backup and rebuild in a local development database; do not simply force the version marker.

<a id="failure-20260919-001"></a>
## [docker] Duplicate migration version after countries renumbering

- ID: FAILURE-20260919-001
- Date: 2026-09-19.
- Context: normal Compose startup; `golang-migrate` reads the read-only `backend/migrations` mount before connecting to PostgreSQL.
- Symptom: migrate exited 1 with `duplicate migration file: 000008_create_role_permissions.up.sql`.
- Cause: the obsolete one-way file `000008_create_countries.up.sql` remained beside the valid `000008_create_role_permissions.up.sql` and `.down.sql` pair after countries moved to version 2. Migration versions must be unique across the source directory.
- Correction: remove only the obsolete `000008_create_countries.up.sql`; retain `000002_create_countries.{up,down}.sql` and the complete version-8 role-permissions pair.
- Verification: migration source loading and normal Compose migration up passed after removal; the database remained at version 8 with `dirty=false`.
- Rule: after renumbering migrations, check that every numeric version has exactly one up file and one down file. Do not leave an old one-way file under a reused version.


# Go Command Patterns

<a id="go-pattern-20260908-001"></a>
## [go] Query the local Go version

- ID: PATTERN-20260908-001
- When to use: A preflight check needs the locally installed Go toolchain version.
- Preflight: Confirm `go` is available. Use `backend/` as the project Go working directory and inspect `backend/go.mod` separately for module requirements. No package, test, build tags, or application environment is needed for this version query.
- Correct command / template: `GOTOOLCHAIN=local go version`
- Verification: Executed on 2026-09-08 from `backend/`; printed the Go version and exited with code `0`.
- Avoid: `go --version`; it is not a supported version-query flag.
- Project note: `GOTOOLCHAIN=local` selects the local toolchain for this inspection. Do not carry this override into builds or tests without checking module requirements. This command does not validate dependencies, compile the backend, or run tests; the current local version is not a project version requirement.
- Related failures: [FAILURE-20260908-001](../commands/topics/go.md#failure-20260908-001).


# Go Command Failures

<a id="go-failure-20261017-001"></a>
## [go] Seed profile freshness window under race instrumentation

- ID: FAILURE-20261017-001
- Date: 2026-10-17
- Context: Go 1.23.3, module `preditto`, no build tags, both PostgreSQL test DSNs set to a disposable database; working directory `backend/`.
- Original command (test URLs omitted): `PREDITTO_AUTH_TEST_DSN='<disposable PostgreSQL URL>' PREDITTO_SEED_TEST_DSN='<same URL>' go test -race -mod=readonly -count=1 ./...`.
- Symptom / exit code: `TestUsersAndUserInfoPostgres` reported correct identity, status, and profile fields but `recent=false`; exit 1. The individual test took 30.36 seconds. Other tested packages passed, including HTTP integration tests.
- Cause: [the test](../../backend/seeds/test/users_postgres_test.go) requires `updated_at` within the last ten seconds. The [seed runner](../../backend/seeds/development/seed.go) starts one transaction before [bcrypt hashing for 20 users](../../backend/seeds/development/users.go); PostgreSQL `CURRENT_TIMESTAMP` remains the transaction-start time. Race instrumentation made the work exceed the assertion's window.
- Incorrect assumption: a full race run would fit the existing fixed freshness window. Literal substitution confirmed the constant extraction preserved seed values and execution logic; the test was unchanged.
- Correct validation invocation (same disposable test DSNs): `go test -mod=readonly -count=1 ./...`.
- Verification: the ordinary full backend suite passed, including PostgreSQL seed and HTTP tests; full-module vet passed. Removing race instrumentation is an alternative validation scope, not a repair of the timing-sensitive test.
- Rule: diagnose `recent=false` against the SQL interval and transaction duration before changing seed data or status constants. Reliable full race coverage needs timestamp bounds tied to the seed invocation in a separate test-maintenance change; that correction is not implemented or verified here.

<a id="go-failure-20260912-001"></a>
## [go] Driver configuration validation is deferred by sql.Open

- ID: FAILURE-20260912-001
- Date: 2026-09-12
- Context: Go 1.23.3, `lib/pq v1.12.3`, no build tags or special Go flags; module `github.com/korvlad21/firstGoWeb`, working directory `backend/`.
- Original command: `go test -mod=readonly ./internal/config ./internal/database`.
- Symptom / exit code: `TestNewPostgresReturnsOpenError` expected `open PostgreSQL:` for invalid SSL mode but received `ping PostgreSQL:`; exit code 1.
- Cause / incorrect assumption: `sql.Open` was assumed to validate this driver's settings immediately. This driver implements `Driver.Open`, so parsing was deferred until connection establishment.
- Correction: Parse with `pq.NewConfig`, create a connector with `pq.NewConnectorConfig`, then use `sql.OpenDB` and `PingContext`. Propagate the caller deadline into `ConnectTimeout` to bound the handshake too.
- Correct invocation: The same targeted test command after correcting initialization.
- Verification: Targeted tests, `go test -mod=readonly ./...`, and `go vet -mod=readonly ./...` pass.
- Rule: Verify the selected driver's initialization and context behavior rather than assuming `sql.Open` establishes or validates a connection. See [implementation](../../backend/internal/database/postgres.go) and [tests](../../backend/internal/database/postgres_test.go).

<a id="go-failure-20260908-001"></a>
## [go] Unsupported version flag

- ID: FAILURE-20260908-001
- Date: 2026-09-08
- Classification: Intentional, user-requested workflow exercise; not an accidental failure or application defect. The CLI syntax correction is reusable.
- Context: Host shell invocation with `GOTOOLCHAIN=local`; no container, build tags, package execution, or dependency changes.
- Working directory: `backend/`, relative to the project root.
- Module / package / test selection: `backend/go.mod` declares `github.com/korvlad21/firstGoWeb`; no package or test was selected.
- Original command: `GOTOOLCHAIN=local go --version`
- Symptom / error / exit code: `flag provided but not defined: -version`; exit code `2`. The usage output lists `version` as a subcommand.
- Cause: `--version` was passed as a top-level flag instead of using the `version` subcommand.
- Incorrect assumption: Go accepts the generic `--version` convention used by some other CLIs.
- Correct invocation: `GOTOOLCHAIN=local go version`
- Verification result: The corrected invocation printed the local Go version and exited with code `0` in the same directory. This verifies the invocation, not the backend build or tests.
- How to recognize early: For a Go version check, use the documented subcommand. Do not assume flags are shared between tools or retry the unsupported flag.
- Related pattern: [PATTERN-20260908-001](../commands/topics/go.md#pattern-20260908-001).
