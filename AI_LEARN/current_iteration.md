# Current Iteration

Date: 2026-09-13

## Outcome

Added development/users.go and development/user_info.go to the existing seed runner in the order teams, users, user_info. All steps use its existing single transaction. Users are matched by username or email, existing IDs are preserved, and password_hash receives per-account bcrypt hashes at default cost. The existing x/crypto version is now a direct dependency; go.sum is unchanged.

Profiles retrieve actual user IDs by username and upsert by user_id. Avatar values are SQL NULL, favorite teams are the requested existing IDs 5 and 19 (or NULL), and updated_at uses PostgreSQL CURRENT_TIMESTAMP. Boroda's profile is Nikita Borodin. The teams implementation, migrations, schema, CLI and Compose launch paths are unchanged.

See [seed documentation](../backend/seeds/README.md) and [backend knowledge](modules/backend/overview.md). The previous iteration is preserved in [archive](archive/2026-09-13-seed-layout.md).

## Verification

- Targeted seed tests, go build -mod=readonly ./..., go test -mod=readonly ./..., go vet -mod=readonly ./..., and go list -mod=readonly ./... passed from backend.
- PostgreSQL integration tests ran with PREDITTO_SEED_TEST_DSN against a temporary isolated PostgreSQL container using existing migrations in disposable schemas. Two runs verified stable user and profile IDs, exact counts, bcrypt passwords, username-only/email-only matches, nonsequential IDs, names/statuses, NULL avatars/favorites, and current timestamps.
- Expected negative tests verified full rollback for split username/email conflicts and missing favorite team IDs. Temporary schemas and the container were removed. No existing application database was used.
- Strict UTF-8, formatting, whitespace, and unchanged teams/migration files were verified. No unexpected command failures occurred.
