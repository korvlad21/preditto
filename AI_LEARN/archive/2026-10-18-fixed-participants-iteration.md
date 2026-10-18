# Current Iteration

Date: 2026-10-18

Replaced random participant generation with explicit username/status lists in [tournament_participants.go](../../backend/seeds/development/tournament_participants.go). The two additional groups contain only accounts marked UserStatusActive in the users seed: 15 participants for Champions League GEL (14 active / 1 not_paid) and 7 for La Liga (6 active / 1 blocked). Account status and participation status remain separate. The original ten entries and tournament ID 1 are preserved; new tournament IDs still resolve by slug. Membership no longer depends on users slice ordering or a random seed.

Verification: targeted TestTournamentParticipants PostgreSQL tests passed, including active account checks, uniqueness, counts, reruns, preserved timestamps, and rollback. Seed package/test go vet, formatting, strict UTF-8, and diff whitespace checks passed. No unexpected failures occurred. The application database was not changed.

Next step: run development seeds when these examples are needed in the intended database.

Previous iteration: [three tournament seeds](2026-10-18-three-tournament-seeds-iteration.md).
