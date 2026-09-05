# firstGoWeb
first golang web-application

## Backend

The backend uses Gin and requires Go 1.22 or newer.

```sh
cd backend
go run .
```

The server listens on all interfaces on port `8080`. Set `APP_PORT` to use a
different port when running locally.

- `GET /` — Hello World HTML.
- `GET /hello` — `{"message":"Hello from Go!"}`.
- `GET /health` — HTTP 200 with `{"status":"ok"}` (process health, no database check).

From the project root, start the backend in Docker (including PostgreSQL):

```sh
docker compose up --build backend
```

Configure `POSTGRES_USER`, `POSTGRES_PASSWORD`, and `POSTGRES_DB` in `.env` first.
The development container uses Air to rebuild on Go source changes and exposes
port `8080`. Compose checks `/health` to report the backend's health.

```sh
curl http://localhost:8080/health
```
