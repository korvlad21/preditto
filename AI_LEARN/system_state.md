# System State

## Active State

The [backend](modules/backend/overview.md) runs migrations, development seeds, then the API through Compose. Migrations cover users, teams, user_info, and RBAC through version 7; roles has no updated_at column. Seeds now include roles and user_roles in the existing single transaction. Reusable assignment/removal functions are in internal/repository. Targeted Go tests and vet pass; PostgreSQL integration tests remain unverified for this change.

## Active Gaps and Risks

- Language guidance conflicts: `project.md` requires English project memory, while `AGENTS.md` requires English instructions and `knowledge_accumulation.md` requires English Markdown. The user's explicit English-language direction is retained for these command instructions.
- User-memory paths differ: `AGENTS.md` and `project.md` specify `AI_USER_LEARN`, while command and memory skills use `AI_USER_LEARN`.
- `module_context_loading.md` describes itself as extending itself. Research backlog and coverage files referenced by the research skill are not yet available.
- Favorite team IDs 5 and 19 must exist; the seed intentionally fails and rolls back if these literal IDs are absent. Broader business-module boundaries remain unresearched.
- Applied migration files should remain consistent with the version already recorded by `schema_migrations`; use a new version for later schema changes.
- [Migration 3 down](../backend/migrations/000003_create_user_info.down.sql) leaves `set_user_info_updated_at()` behind; reapplying migration 3 after its rollback currently fails. The RBAC rollback must stop at version 3 until that separate defect is fixed.

## Next Step

Run the PostgreSQL integration tests using [the seed README](../backend/seeds/README.md) when a test database is available. Use the next migration version for future schema changes. Align the remaining guidance conflicts in a separate instruction-maintenance task.
