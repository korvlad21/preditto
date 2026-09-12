# System State

## Active State

The first [backend migration](modules/backend/overview.md) creates users and an updated_at trigger. Compose gates backend startup on successful migrate completion after PostgreSQL health. Isolated real PostgreSQL checks passed for constraints, trigger behavior, up/down/reapply, and dependency ordering. [Command knowledge](commands/index.md) records the tested migration commands and URL correction.

## Active Gaps and Risks

- Language guidance conflicts: `project.md` requires English project memory, while `AGENTS.md` requires English instructions and `knowledge_accumulation.md` requires English Markdown. The user's explicit English-language direction is retained for these command instructions.
- User-memory paths differ: `AGENTS.md` and `project.md` specify `AI_USER_LEARN`, while command and memory skills use `AI_USER_LEARN`.
- `module_context_loading.md` describes itself as extending itself. Research backlog and coverage files referenced by the research skill are not yet available.
- Full application startup was not tested in this migration iteration; the Docker ordering check used a lightweight backend probe. Broader business-module boundaries remain unresearched.

## Next Step

Start the full application with `docker compose up -d --build` when needed. Use the migration README for manual apply or rollback. Align the remaining guidance conflicts in a separate instruction-maintenance task.
