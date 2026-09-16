# Database seeds

Seed logic lives in `backend/seeds/development`; `backend/cmd/seed` is the standalone CLI. It reuses the existing configuration and PostgreSQL connection packages. No seeding framework or additional dependencies are required.

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

`development.Run` executes the ordered seed list in a single transaction. An error rolls back the entire run. Add future seed functions in separate files in `development` and register them in dependency order in `development/seed.go`; each function must be safe to repeat.

The registered order is `teams → users → user_info → roles → user_roles`. The users seed creates or updates three development accounts, matching existing users by username or email and preserving their IDs. Each password is hashed with bcrypt at its default cost before writing `password_hash`; reruns restore the configured account fields and generate new valid salted hashes. If username and email belong to different existing users, the unique constraints reject the update and the whole seed transaction rolls back.

The user_info seed retrieves each user's actual ID by username and upserts on the unique `user_id`. It restores names, writes SQL NULL for avatar_url, and sets updated_at with PostgreSQL CURRENT_TIMESTAMP. Favorite teams are existing IDs 5 and 19; Hanna has no favorite team. These IDs are used directly as requested, without changing or remapping the teams seed. If either team ID is absent, the foreign key rejects the profile and rolls back all seed changes.

The teams seed contains 36 entries. It validates nonempty names/slugs, unique slugs within the dataset, and exactly three uppercase ASCII letters in every short_name before writing. `ON CONFLICT (slug) DO UPDATE` uses the existing unique constraint to prevent duplicate rows and correct seeded names/abbreviations. It preserves IDs, created_at, and profile references; unrelated teams are untouched. New logos are NULL, while existing logos are preserved on reruns. These checks apply to seed data; the teams schema is unchanged.

Tests in `test` exercise the public runner with a test SQL driver. Tests of private validation helpers remain in `development`, since Go directories are separate packages. From `backend`, run `go test ./seeds/...`; the CLI remains `go run ./cmd/seed` after migrations with PostgreSQL configured.

To run the PostgreSQL integration tests from `backend`, set `PREDITTO_SEED_TEST_DSN` to a test database connection string and run `go test -mod=readonly -count=1 -v ./seeds/test`. Without that variable, these integration tests are skipped. Each test applies the existing migrations in a temporary schema and drops it afterward; the test database user must have permission to create schemas. Tests cover repeated runs, bcrypt verification, existing username/email matches with nonsequential user IDs, NULL fields, restored profile data, and full rollback on identity conflicts or missing favorite teams.

The roles seed upserts `admin` and `participant` by unique code, restoring their names while preserving IDs and created_at. The user_roles seed resolves users by username and roles by code, then assigns admin to korvlad21 and hanna_stoma and participant to boroda. Existing assignments and unrelated roles are preserved. Missing users or roles fail the transaction.

Reusable persistence functions `repository.AssignUserRole` and `repository.RemoveUserRole` live in `backend/internal/repository/user_roles.go`. They accept a database or caller-owned transaction and user/role IDs. Assignment uses `ON CONFLICT (user_id, role_id) DO NOTHING`; removal deletes only that pair and succeeds if absent. Database errors, including invalid foreign keys, propagate to the caller. No service layer exists yet; these are persistence operations without application policy.

The current roles migration has no updated_at column; seeds write only code and name and also work with an additional defaulted updated_at column. PostgreSQL tests additionally cover role names, nonsequential user/role IDs, repeated assignments/removals, restoration of removed assignments, foreign-key errors, and transaction rollback.
