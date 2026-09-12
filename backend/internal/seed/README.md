# Database seeds

Seed logic lives in this package; `backend/cmd/seed` is the standalone CLI. It reuses the existing configuration and PostgreSQL connection packages. No seeding framework or additional dependencies are required.

From the project root:

```sh
docker compose up -d --build
```

Startup order is PostgreSQL healthy, migrate successful, seed successful, then backend. A failed seed exits nonzero and blocks backend startup. Seeds do not run inside the HTTP application. The one-shot seed service uses the same build and environment as the backend.

To rerun seeds manually after migrations:

```sh
docker compose run --rm seed
```

Compose starts required dependencies. For a running environment that needs newly added migrations and seeds applied explicitly:

```sh
docker compose run --rm migrate up
docker compose run --rm seed
```

Restarting only the backend does not rerun seeds.

`Run` executes the ordered seed list in a single transaction. An error rolls back the entire run. Add future seed functions in separate files in this package and register them in dependency order in `seed.go`; each function must be safe to repeat.

The teams seed contains 36 entries. It validates nonempty names/slugs, unique slugs within the dataset, and exactly three uppercase ASCII letters in every short_name before writing. `ON CONFLICT (slug) DO UPDATE` uses the existing unique constraint to prevent duplicate rows and correct seeded names/abbreviations. It preserves IDs, created_at, and profile references; unrelated teams are untouched. New logos are NULL, while existing logos are preserved on reruns. These checks apply to seed data; the teams schema is unchanged.
