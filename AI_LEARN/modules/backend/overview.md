# Backend

Verified on 2026-09-12 through targeted tests, full-module tests, static analysis, and Compose configuration validation.

The module name in [go.mod](../../../backend/go.mod) is `preditto`; internal imports use `preditto/internal/...`. The former `github.com/korvlad21/firstGoWeb` prefix was removed from the backend. These packages are local and remain required by API startup and database code. Full-module tests passed after the rename.

- [config.Load](../../../backend/internal/config/config.go) requires `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, and `POSTGRES_SSLMODE` from the process environment. It validates the port and supplies no database defaults.
- [NewPostgres](../../../backend/internal/database/postgres.go) uses `database/sql` and `lib/pq`. URL encoding preserves special characters. Driver configuration is validated before opening the pool; `PingContext` verifies connectivity. Errors preserve their causes, and failed Ping closes the pool.
- [API startup](../../../backend/cmd/api/main.go) verifies PostgreSQL before serving HTTP. Its five-second deadline also bounds the driver handshake. The caller closes the pool when `run` exits.
- [Compose](../../../docker-compose.yaml) forwards all six variables. The ignored root `.env` supplies local values; [.env.example](../../../.env.example) documents Compose defaults. Direct host execution requires exported variables and a host-reachable PostgreSQL address; the backend does not load `.env` itself.
- [go.mod](../../../backend/go.mod) adds only `github.com/lib/pq v1.12.3`; the Go directive and existing dependency versions are unchanged. No ORM is used.

Tests cover configuration errors, DSN escaping, canceled Ping, rejected connections, and stalled-handshake timeout. A successful real PostgreSQL connection remains unverified.
