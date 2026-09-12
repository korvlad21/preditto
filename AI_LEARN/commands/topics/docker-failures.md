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
