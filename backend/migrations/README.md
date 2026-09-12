# PostgreSQL migrations

Run these commands from the project root with Docker Compose and the existing `.env` file.

## Automatic startup

```sh
docker compose up -d --build
```

Compose waits for the PostgreSQL healthcheck, runs the one-shot `migrate` service with `up`, then starts the backend only after migration success. A migration failure blocks backend startup. Migrations are not run by the Go application.

The service mounts this directory read-only and maps the existing `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, and `POSTGRES_SSLMODE` variables to the driver's `PG*` variables. Credentials are not embedded into a URL. The host must resolve from inside the Compose network (normally `postgres`).

## Manual commands

Apply all pending migrations:

```sh
docker compose run --rm migrate up
```

Roll back the last migration:

```sh
docker compose run --rm migrate down 1
```

Show the current version:

```sh
docker compose run --rm migrate version
```

The first rollback removes `users` and its data, its identity sequence and trigger, and the trigger function. Stop backend writers before rolling back a live database. A later `up` reapplies pending migrations. After adding migration files while the application is already running, use the manual `up` command or recreate the migration service; restarting only the backend does not run migrations.

## First migration

`000001_create_users.up.sql` creates `users` with a PostgreSQL identity primary key, unique username and email, required fields, and microsecond-precision timestamps. A `BEFORE UPDATE` trigger assigns `CURRENT_TIMESTAMP` to `updated_at` on every row update. PostgreSQL's `CURRENT_TIMESTAMP` is the transaction start time, so updates within the same transaction may have identical timestamps.

Both directions use a transaction. The down migration removes the table before its dedicated trigger function; dropping the table removes its trigger and identity sequence automatically.

References: [migrate CLI](https://github.com/golang-migrate/migrate/blob/v4.20.1/cmd/migrate/README.md), [Compose startup order](https://docs.docker.com/compose/how-tos/startup-order/), [PostgreSQL trigger functions](https://www.postgresql.org/docs/current/plpgsql-trigger.html).
