# Current Iteration

Date: 2026-10-16

Added [tournament](../../backend/seeds/development/tournaments.go) and [prize rule](../../backend/seeds/development/tournament_prize_rules.go) seeds to the existing transactional runner, after its prior steps and in dependency order. The UEFA Champions League 2026/2027 entry uses the requested name, slug, and active status. Official full-season dates, midnight timestamp semantics, nullable unknown fields, and repeat-run preservation rules are documented in the [seed README](../../backend/seeds/README.md). Prize rules resolve the actual ID by slug and upsert 50.00/30.00/15.00 percent; no rule allocates the remaining 5%. Existing migrations and production seed behavior for other tables were not edited.

All seed tests passed with race detection against isolated PostgreSQL 18.6, including fresh/nonsequential existing tournament IDs, repeat-run uniqueness, exact percentages, missing/changed rule restoration, unknown-field and created_at preservation, unrelated-row preservation, and full transaction rollback after a prize failure. Seed package/CLI vet, gofmt, strict UTF-8 validation, and diff checks passed. The application database was not seeded. Local Docker startup/sandbox corrections are recorded only in AI_USER_LEARN.

The previous migration iteration is preserved in [archive](2026-10-11-tournament-migrations-iteration.md). Apply pending migrations before running the normal seed CLI in the intended environment.
