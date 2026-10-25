# Preditto

Go/Gin/PostgreSQL backend with a React/TypeScript/Vite frontend. The API implements authentication, current-user profiles, teams/countries and tournament listing/selection. The frontend currently contains starter UI and static team logos, with no API client or authentication flow.

## Development with Docker Compose

From the project root, copy [.env.example](.env.example) to an untracked `.env` and configure all values. PostgreSQL requires POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB and POSTGRES_SSLMODE. Inside Compose, the database host is normally `postgres`.

Set two independent JWT_ACCESS_SECRET and JWT_REFRESH_SECRET values, each at least 32 bytes. For example, run `openssl rand -hex 32` separately for each secret. Keep secrets out of commits and command output shared in documentation. JWT_ACCESS_TOKEN_TTL must be whole seconds between 1s and 1h; JWT_REFRESH_TOKEN_TTL must be whole seconds and longer. JWT_REFRESH_TOKEN_SHORT_TTL defaults to 12h and must be longer than access TTL but shorter than JWT_REFRESH_TOKEN_TTL. The example uses 15m access, 12h short refresh and 720h remembered/legacy refresh. See [Remember Me](backend/docs/auth.md#remember-me) for optional request behavior.

```sh
docker compose config --quiet
docker compose up -d --build
```

[Compose](docker-compose.yaml) waits for PostgreSQL, applies migrations, runs development seeds, then starts the backend; frontend startup depends on the backend. Seeds create/update development data and restore configured passwords/profile fields on rerun. Read the [seed guide](backend/seeds/README.md) before using an existing database.

The backend is exposed on port 8080, frontend on 5173, and PostgreSQL on 5432. The backend container uses Air for rebuilds. For backend and its database dependencies only, use `docker compose up --build backend`.

```sh
curl http://localhost:8080/health
```

`GET /health` returns `{"status":"ok"}` for process health; it does not test database readiness. Non-API demo routes `GET /` and `GET /hello` also exist.

## Backend on the host

Use a Go toolchain compatible with [backend/go.mod](backend/go.mod). Export the PostgreSQL and four JWT variables above in the process environment; the Go application does not load `.env` itself, and host TTLs do not inherit Compose defaults. Use a host-reachable database address, normally `localhost` for the published Compose database.

Apply [migrations](backend/migrations/README.md) and provision required development data/participant role through the [seed workflow](backend/seeds/README.md) before exercising registration. Then:

```sh
cd backend
go run ./cmd/api
```

The server binds all interfaces on port 8080 by default; APP_PORT overrides the host process port. Starting the API itself does not apply migrations or seeds.

## Documentation and checks

- [Swagger UI and generation](backend/docs/swagger.md): set `APP_ENV=development`,
  then open [Swagger UI](http://localhost:8080/swagger/index.html). Update with
  `make swagger`; verify with `make swagger-check`. Other environments do not expose Swagger.
- [Architecture and module navigation](AI_LEARN/index.md).
- [AI contributor instructions](AGENTS.md).
- [API response contract](backend/docs/responses.md), [auth](backend/docs/auth.md), [users](backend/docs/user.md), [teams/countries](backend/docs/teams.md), [tournaments](backend/docs/tournaments.md).
- [Backend test prerequisites and known limits](AI_LEARN/commands/topics/go.md).
- [Frontend setup, scripts and current integration status](frontend/README.md).
