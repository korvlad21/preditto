# Current Iteration

Date: 2026-09-16

Added idempotent roles and user-role seeds in dependency order within the existing transaction. Assignment/removal SQL is reusable through internal/repository with caller-owned DB/Tx and actual IDs. The existing roles migration has no updated_at; no schema changes were required. See [backend knowledge](modules/backend/overview.md) and [seed documentation](../backend/seeds/README.md).

Validation: targeted seed/repository Go tests and vet including cmd/seed passed; PostgreSQL tests were added for repeated runs, nonsequential IDs, deletion, foreign keys, and rollback, but skipped without a test DSN. Runtime integration verification remains outstanding.

The previous iteration is preserved in [archive](archive/2026-09-15-rbac-baseline-iteration.md).
