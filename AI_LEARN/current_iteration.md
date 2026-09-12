# Current Iteration

Date: 2026-09-12

## Outcome

Added a standalone Go seed command in backend/cmd/seed and an ordered transactional runner in backend/internal/seed, reusing the existing config and database packages. The teams seed contains exactly the 36 requested triples, validates three uppercase ASCII letters per abbreviation, and upserts on slug without changing existing IDs, timestamps, logos, or profile references. New logos are NULL. Compose now starts PostgreSQL, migrate, seed, then backend. No dependencies or SQL migrations were changed.

See [seed documentation](../backend/internal/seed/README.md) and [backend knowledge](modules/backend/overview.md). The preceding result is preserved in [archive](archive/2026-09-12-user-info-migration.md).

## Verification

- Targeted seed tests and go test -mod=readonly ./... passed; go vet -mod=readonly ./cmd/seed ./internal/seed passed.
- All 36 dataset entries were independently compared with the user-provided list.
- Compose configuration validated. The isolated PostgreSQL run used the actual migration files, actual seed command/source mounted read-only, the existing backend Docker image, and a lightweight downstream backend probe.
- Initial run produced exactly 36 rows with valid abbreviations and NULL logos. Rerun restored modified names/abbreviations while preserving row count, IDs, created_at, an existing logo, and a profile reference.
- An intentional database exception on the final team caused a nonzero seed/Compose exit, kept the downstream backend unstarted, and rolled back earlier team changes. Removing that test-only trigger restored successful startup and the exact dataset. This was an expected negative test, not an unexpected command failure.
- Temporary containers, network, and test files were removed. UTF-8, documentation references, formatting, and whitespace were checked. The full application image was not rebuilt and the actual HTTP backend was not started.
