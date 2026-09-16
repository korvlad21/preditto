# System State

## Active State

The [backend](modules/backend/overview.md) runs migrations, development seeds, then the API through Compose. The current migration order is users 1, countries 2, teams 3, user_info 4, and RBAC 5-8. Teams.country is required and references countries.short_name. The local database was rebuilt from this order and reports version 8 with dirty=false. Seed data contains 16 countries, 36 teams, and three users with no invalid country links. PostgreSQL seed integration tests pass; migrate and seed exited 0, backend is healthy, and frontend runs.

## Active Gaps and Risks

- Language guidance conflicts: `project.md` requires English project memory, while `AGENTS.md` requires English instructions and `knowledge_accumulation.md` requires English Markdown. The user's explicit English-language direction is retained for these command instructions.
- User-memory paths differ: `AGENTS.md` and `project.md` specify `AI_USER_LEARN`, while command and memory skills use `AI_USER_LEARN`.
- `module_context_loading.md` describes itself as extending itself. Research backlog and coverage files referenced by the research skill are not yet available.
- Favorite team IDs 5 and 19 must exist; the seed intentionally fails and rolls back if these literal IDs are absent. Broader business-module boundaries remain unresearched.
- Applied migration files should remain consistent with the version already recorded by `schema_migrations`; use a new version for later schema changes.

## Next Step

Use the next migration version for future schema changes in databases where the earlier version is already applied. The current renumbering was verified through a local development database rebuild and must not be applied as a routine in-place migration on populated external databases. Align the remaining guidance conflicts in a separate instruction-maintenance task.
