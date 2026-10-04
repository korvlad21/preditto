# System State

## Active State

The public [teams API](../backend/docs/teams.md) is registered at `POST /api/teams/get_all_teams`. The implementation and PostgreSQL HTTP tests verify optional country filtering, all-country sentinels, all six fields, nullability, and empty arrays. This addition requires no migration or new dependency.

Vite uses `frontend/images` as its `publicDir`, so `frontend/images/logos/teams` is served at `/logos/teams/`. The `frontend/public` directory is no longer used; seed URLs remain `/logos/teams/<slug>.svg`.

The 36 development teams have local SVG assets in `frontend/images/logos/teams/`. The seed derives `/logos/teams/<slug>.svg` and updates `logo_url` on conflict. Sources, identity notes, and verification are recorded in [the asset inventory](../docs/assets/team-logos.md). These seed changes have not been applied to a running database in this iteration.

The [backend](modules/backend/overview.md) runs migrations, development seeds, then the API through Compose. Migration order is users 1, countries 2, teams 3, user_info 4, RBAC 5-8, and refresh_tokens 9. Teams.country references countries.short_name. The [authentication API](../backend/docs/auth.md) provides register/login/refresh/logout and access middleware with bcrypt, separate JWT secrets, hashed refresh sessions, and atomic rotation. [Auth route declarations](../backend/internal/router/auth.go) live in `internal/router`; [user persistence](../backend/internal/repository/user.go) lives in `repository/user.go`. Registration assigns participant using the existing role repository. Prior full backend race tests and vet passed; the route and repository moves each passed targeted tests, vet, and PostgreSQL HTTP integration tests.

## Active Gaps and Risks

- Language guidance conflicts: `project.md` requires English project memory, while `AGENTS.md` requires English instructions and `knowledge_accumulation.md` requires English Markdown. The user's explicit English-language direction is retained for these command instructions.
- User-memory paths differ: `AGENTS.md` and `project.md` specify `AI_USER_LEARN`, while command and memory skills use `AI_USER_LEARN`.
- `module_context_loading.md` describes itself as extending itself. Research backlog and coverage files referenced by the research skill are not yet available.
- Favorite team IDs 5 and 19 must exist; the seed intentionally fails and rolls back if these literal IDs are absent. Broader business-module boundaries remain unresearched.
- Applied migration files should remain consistent with the version already recorded by `schema_migrations`; use a new version for later schema changes.
- API startup requires distinct JWT secrets and valid TTLs. Registration requires the participant role. Access tokens remain usable until expiry after logout or blocking; no blacklist or refresh-session cleanup scheduler is implemented.

## Next Step

Run the normal development seeds to populate local logo paths in an existing database. Keep future seed slugs and the [asset inventory](../docs/assets/team-logos.md) in sync.

Use migration 10 or later for future schema changes. Environments adopting authentication must apply migration 9, provision participant, and configure JWT environment variables before API startup. The earlier renumbering must not be applied as a routine in-place migration on populated external databases. Align the remaining guidance conflicts in a separate instruction-maintenance task.
