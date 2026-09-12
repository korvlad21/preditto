# System State

## Active State

The [backend](modules/backend/overview.md) uses the local module name `preditto`, loads PostgreSQL configuration from environment variables, and verifies its connection before starting HTTP. Full-module tests pass after the module rename. Static analysis and Compose configuration validation passed during PostgreSQL setup. [Command knowledge](commands/index.md) records the verified driver initialization correction.

## Active Gaps and Risks

- Language guidance conflicts: `project.md` requires English project memory, while `AGENTS.md` requires English instructions and `knowledge_accumulation.md` requires English Markdown. The user's explicit English-language direction is retained for these command instructions.
- User-memory paths differ: `AGENTS.md` and `project.md` specify `AI_USER_LEARN`, while command and memory skills use `AI_USER_LEARN`.
- `module_context_loading.md` describes itself as extending itself. Research backlog and coverage files referenced by the research skill are not yet available.
- A successful real PostgreSQL connection remains unverified. Backend knowledge covers configuration and startup; business-module boundaries remain unresearched.

## Next Step

Verify API startup against a running PostgreSQL service using the configured environment. Align the remaining guidance conflicts in a separate instruction-maintenance task.
