# Current Iteration

Date: 2026-10-16

Updated the [tournament seed](../../backend/seeds/development/tournaments.go) to insert and restore amount_currency USD and type_of_tournament Champions league using exported constants in [model/tournament.go](../../backend/internal/model/tournament.go), following the existing model constant convention. No new package or type abstraction was introduced. The user's existing dates 2026-09-08 and 2027-06-05 remain unchanged. Unknown cost is still NULL on insert and preserved on conflict. The test and README now match the user's league-phase start date rather than the earlier qualifying start date.

Targeted PostgreSQL tests passed for fresh/existing tournaments, repeat runs, currency/type restoration, preserved costs and timestamps, exact prize rules, and transactional rollback. Model/seed/CLI vet, formatting, UTF-8, and diff checks passed. The disposable test container was removed; the application database was not seeded. The preceding seed implementation is preserved in [archive](2026-10-16-tournament-seeds-iteration.md).
