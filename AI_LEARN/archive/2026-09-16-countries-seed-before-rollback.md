# Current Iteration

Date: 2026-09-16

Added the 16-country seed before teams in the existing transaction. It validates nonempty names and unique uppercase three-letter codes, upserts names by short_name, and relies on database defaults for IDs and created_at. Migration 000009 adds the required unique constraint with a matching down; migration 000008 remains unchanged. See [seed documentation](../backend/seeds/README.md).

Verification: targeted seed tests and vet including cmd/seed passed. PostgreSQL tests for repeat runs, existing IDs/timestamps, unrelated rows, and full rollback were added but skipped because PREDITTO_SEED_TEST_DSN is unset. Migration execution remains unverified.

Command note (non-reusable discovery result): `git status --short && rg --files -g AGENTS.md backend && rg --files backend/migrations backend/seeds` exited 1 because no nested AGENTS.md exists; the final listing was not executed. Running `rg --files backend/migrations backend/seeds` separately succeeded. Treat a no-match search as a probe result and do not gate independent discovery on it.

The previous iteration is preserved in [archive](archive/2026-09-16-role-seeds-iteration.md).
