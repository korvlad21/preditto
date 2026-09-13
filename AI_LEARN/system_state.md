# System State

## Active State

The [backend](modules/backend/overview.md) has users, teams, and user_info migrations plus a standalone Go seed command for 36 teams, three bcrypt-hashed development users, and their profiles in dependency order in one transaction. Seed implementation lives in `backend/seeds/development`; public runner tests live in `backend/seeds/test`, and private validation tests remain alongside the implementation. The seed runner passes backend build, all tests, vet, and import resolution. Isolated PostgreSQL tests confirm idempotent user/profile writes with real IDs, matching by username or email, and rollback on identity conflicts or absent favorite team IDs. Compose gates startup on PostgreSQL health, migration success, then seed success. Unit tests and isolated PostgreSQL checks verified seed data, repeated upserts, stable IDs and references, full rollback on an injected failure, blocked backend startup, and successful recovery. Existing SQL migrations are unchanged.

## Active Gaps and Risks

- Language guidance conflicts: `project.md` requires English project memory, while `AGENTS.md` requires English instructions and `knowledge_accumulation.md` requires English Markdown. The user's explicit English-language direction is retained for these command instructions.
- User-memory paths differ: `AGENTS.md` and `project.md` specify `AI_USER_LEARN`, while command and memory skills use `AI_USER_LEARN`.
- `module_context_loading.md` describes itself as extending itself. Research backlog and coverage files referenced by the research skill are not yet available.
- Favorite team IDs 5 and 19 must exist; the seed intentionally fails and rolls back if these literal IDs are absent. Full application startup was not tested in this seed iteration. The seed ran with the existing backend image and actual source; its downstream dependency was a lightweight probe. Broader business-module boundaries remain unresearched.

## Next Step

Start the full application with `docker compose up -d --build` when needed. Use [the seed README](../backend/seeds/README.md) to rerun seeds or register the next seed. Align the remaining guidance conflicts in a separate instruction-maintenance task.
