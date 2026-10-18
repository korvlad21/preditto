# Current Iteration

Date: 2026-10-18

Expanded [development tournament seeds](../../backend/seeds/README.md) to three tournaments while preserving the original tournament data, prizes, and ten participants targeting ID 1. Added Champions League at 40.00 GEL and La Liga at 20.00 USD, their prize rules, and participant groups of 15 (14 active / 1 not_paid) and 7 (6 active / 1 blocked). New participants resolve tournament IDs by slug and use deterministic random permutations of the seed users. La Liga dates, 2026-08-15 through 2027-05-30, are grounded in official LALIGA and RFEF sources linked from the seed. Added model constants CurrencyGEL and TournamentTypeLaLiga.

Verification: `go test -mod=readonly -count=1 -v ./seeds/development ./seeds/test` from `backend/`, with PREDITTO_SEED_TEST_DSN pointing to disposable PostgreSQL, passed every test except the fixture timestamp assertion described below. After its correction, `go test -mod=readonly -count=1 -v ./seeds/test -run '^TestTournamentParticipantsPostgres$'` passed both seed runs. Checks cover nonsequential IDs, unique membership, requested counts/statuses, unchanged first-tournament behavior, field/prize/status restoration, preserved identities/timestamps, and transaction rollback. `go vet -mod=readonly ./seeds/development ./seeds/test ./cmd/seed`, formatting, diff whitespace, and strict UTF-8 checks passed. No migration, dependency, or application-database change was required.

Non-reusable test-fixture failure: the initial seed test command above exited 1 because the expanded participants test precreated tournament ID 1 before calling the runner but still compared joined_at with the tournament's earlier created_at. The seed correctly preserved created_at. The corrected assertion compares joined_at with updated_at, which the tournament upsert refreshes in the seed transaction; the targeted rerun passed. Rule: when a fixture precreates an upserted parent, use the timestamp refreshed in the tested transaction when checking child timestamp defaults. This one-off fixture correction is recorded here rather than adding a command recipe.

Non-reusable editing failure: apply_patch rejected a knowledge update containing both Delete File and Add File operations for current_iteration.md with `multiple operations target`. Replacing them with one Update File operation succeeded. Use a single update operation per existing path; the rejected patch made no project edits.

Next step: apply pending tournament migrations and run the normal development seeds when these examples are needed in the application database. The original requirement for tournament ID 1 remains; changing the seed users dataset or ordering may change the new random selections.

Previous iteration: [tournament API POST method](2026-10-18-tournaments-post-api-iteration.md).
