# Current Iteration

Date: 2026-10-16

Added [tournament_participants.go](../../backend/seeds/development/tournament_participants.go) with a fixed selection of 10 unique seeded usernames, excluding korvlad21, hanna_stoma, and boroda. Every participant targets tournament ID 1; statuses are eight active, one not_paid, and one blocked. The seed resolves user IDs by username, omits joined_at to use its database default, and updates only status on a tournament/user conflict. It runs after tournaments within the existing seed transaction.

Updated the PostgreSQL fixture to apply migration 000011 and the test driver to accept participant inserts. Existing tournament tests retain nonsequential tournament IDs while providing tournament ID 1 for participants. Added integration coverage for nonsequential user IDs, status counts, excluded users, default timestamps, repeated-run status restoration with preserved IDs/joined_at, and rollback when tournament ID 1 is absent. The prize failure test now includes participants in its rollback check.

All seed tests passed against disposable PostgreSQL with `go test -mod=readonly -count=1 -v ./seeds/...`. Dataset, formatting, strict UTF-8, and diff checks passed. No command failed. The application database was not seeded. Migration 000011 and tournament ID 1 must exist before applying this seed; tournament IDs are deliberately not resolved by slug.

Previous iteration: [expanded user seeds](2026-10-16-expanded-users-iteration.md).
