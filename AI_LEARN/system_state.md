# System State

## Active State

The [backend](modules/backend/overview.md) uses Docker Compose to run PostgreSQL migrations, then development seeds, then the API. The migrations now cover users, teams, user_info, and RBAC through version 7. In this local untracked RBAC baseline, [migration 4](../backend/migrations/000004_create_roles.up.sql) creates `roles` without `updated_at` or a trigger. The empty RBAC tables were rolled back to version 3 and reapplied to version 7; existing users remained. A fresh isolated database passed up, down 4, and reapply. Normal Compose startup and the API health endpoint passed.

## Active Gaps and Risks

- Language guidance conflicts: `project.md` requires English project memory, while `AGENTS.md` requires English instructions and `knowledge_accumulation.md` requires English Markdown. The user's explicit English-language direction is retained for these command instructions.
- User-memory paths differ: `AGENTS.md` and `project.md` specify `AI_USER_LEARN`, while command and memory skills use `AI_USER_LEARN`.
- `module_context_loading.md` describes itself as extending itself. Research backlog and coverage files referenced by the research skill are not yet available.
- Favorite team IDs 5 and 19 must exist; the seed intentionally fails and rolls back if these literal IDs are absent. Broader business-module boundaries remain unresearched.
- Applied migration files should remain consistent with the version already recorded by `schema_migrations`; use a new version for later schema changes.
- [Migration 3 down](../backend/migrations/000003_create_user_info.down.sql) leaves `set_user_info_updated_at()` behind; reapplying migration 3 after its rollback currently fails. The RBAC rollback must stop at version 3 until that separate defect is fixed.

## Next Step

Use [the seed README](../backend/seeds/README.md) for seed work. Use the next migration version for future schema changes. Align the remaining guidance conflicts in a separate instruction-maintenance task.
