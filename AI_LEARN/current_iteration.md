# Current Iteration

Date: 2026-10-16

Fixed the [tournament seed](../backend/seeds/development/tournaments.go) after adding cost 10: the SQL referenced nonexistent cost_currency, omitted cost_amount from INSERT, and had seven placeholders for eight arguments. INSERT and conflict updates now use cost_amount with eight matching parameters. A local tournament struct and named-field data slice follow the teams seed pattern; the exact decimal string 10.00 avoids floating-point conversion. Dates and shared currency/type constants are unchanged.

Targeted PostgreSQL tests passed for new/existing tournaments, repeat runs restoring cost/currency/type, prize rules, and full rollback. Seed/CLI vet, UTF-8, formatting, and diff checks passed. The disposable database container was removed; the application database was not seeded. No command failed during this iteration; the reported SQL defect was diagnosed from the user's edited source and verified after correction.

The previous rename to championsLeague20262027Slug remains documented in system_state.md and the seed source. Earlier constant changes are preserved in [archive](archive/2026-10-16-tournament-constants-iteration.md). Keep INSERT columns, positional placeholders, arguments, and conflict-update column names aligned when extending a seed.
