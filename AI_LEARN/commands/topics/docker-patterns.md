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
