# System State

## Active State

[Migration 000004](../backend/migrations/000004_create_user_info.up.sql) now creates nullable selected_tournament_id; migration 000010 adds its tournament foreign key with ON DELETE SET NULL and removes that key on rollback. Migration 000013 was removed at the user's request for a fresh local rebuild. Isolated PostgreSQL checks passed. An existing database requires a deliberate rebuild to adopt this revised baseline; runtime application state has not been inspected in this iteration.

Backend domain statuses, role codes, and PostgreSQL violation codes reuse [constants in their owning packages](modules/backend/overview.md#shared-constants). JWT types, the Bearer scheme, API error codes, and messages use their original inline literals at the user's request. The seven restored Go files match Git HEAD; targeted auth, middleware, handler, and response checks passed without PostgreSQL test DSNs. The earlier full PostgreSQL suite and vet results are preserved in [the previous iteration](archive/2026-10-17-backend-constants-iteration.md).

The [tournament participants seed](../backend/seeds/development/tournament_participants.go) runs after tournaments and users. It resolves 10 selected usernames to actual user IDs and targets literal tournament ID 1, with statuses 8 active / 1 not_paid / 1 blocked. It excludes the original three users and upserts status by the unique tournament/user pair, preserving joined_at on rerun. Migration 000011 and an existing tournament ID 1 are required; a missing user or tournament rolls back the shared seed transaction. All seed tests passed against disposable PostgreSQL, including repeated-run status restoration, preserved joined_at, and rollback when tournament ID 1 is absent. These seed changes have not been applied to the application database.

The development [users](../backend/seeds/development/users.go) and [profiles](../backend/seeds/development/user_info.go) contain 20 matching records, including 17 new accounts (14 ACTIVE, 3 BLOCKED); the original three records are unchanged. Five profiles have no favorite team. All referenced team IDs were confirmed in the existing PostgreSQL teams table on 2026-10-16. The application database has not been seeded with the new users. The existing PostgreSQL user/profile test now expects 20 records and checks profile completeness and favorite-team references on repeat runs.

The [development seeds](../backend/seeds/README.md) now include the UEFA Champions League 2026/2027 tournament and three prize rules totaling 95%. Both run in the existing transaction; the prize rules resolve the actual tournament ID. The tournament uses `championsLeague20262027Slug` for its seed key and shared model constants for USD and Champions league; its user-selected dates remain 2026-09-08 through 2027-06-05. Tournament data uses a local struct and slice, matching the teams seed pattern. Cost is 10.00 USD and is restored on rerun; targeted PostgreSQL checks verified cost/currency/type restoration and rollback. PostgreSQL seed integration tests with race detection and seed/CLI vet passed on 2026-10-16. These seeds have not been applied to the application database.

Tournament migrations [000010](../backend/migrations/000010_create_tournaments.up.sql), [000011](../backend/migrations/000011_create_tournament_participants.up.sql), and [000012](../backend/migrations/000012_create_tournament_prize_rules.up.sql) are ready. Isolated PostgreSQL checks passed for application, constraints, rollback to version 9 with original schema and users preserved, reapplication, and full rollback. Existing migrations and the application database were not changed. See [backend knowledge](modules/backend/overview.md) for the schema constraints.

The protected [user API](../backend/docs/user.md) now uses RequireAuth followed by LoadUser on the shared protected group in cmd/api/main.go. LoadUser calls UserService.GetUserInfo once per request and stores a shared UserInfo pointer, including nullable selected_tournament_id. The handler reads middleware.CurrentUser without another query. Missing users retain HTTP 404; missing profiles retain null fields. Targeted PostgreSQL race tests and affected-package vet passed.

The public [teams API](../backend/docs/teams.md) exposes `POST /api/teams/get_all_teams` with optional country filtering and `POST /api/teams/get_all_countries` without required parameters or body. Both return arrays ordered by name. Country responses include `id`, `name`, `short_name`, and `created_at`; empty tables return `[]`. Targeted HTTP tests passed against PostgreSQL with race detection, and affected-package/API vet passed. No migration or new dependency is required.

Vite uses `frontend/images` as its `publicDir`, so `frontend/images/logos/teams` is served at `/logos/teams/`. The `frontend/public` directory is no longer used; seed URLs remain `/logos/teams/<slug>.svg`.

The 36 development teams have local SVG assets in `frontend/images/logos/teams/`. The seed derives `/logos/teams/<slug>.svg` and updates `logo_url` on conflict. Sources, identity notes, and verification are recorded in [the asset inventory](../docs/assets/team-logos.md). These seed changes have not been applied to a running database in this iteration.

The [backend](modules/backend/overview.md) runs migrations, development seeds, then the API through Compose. Migration order is users 1, countries 2, teams 3, user_info 4, RBAC 5-8, refresh_tokens 9, tournaments 10, tournament_participants 11, and tournament_prize_rules 12. Teams.country references countries.short_name. The [authentication API](../backend/docs/auth.md) provides register/login/refresh/logout and access middleware with bcrypt, separate JWT secrets, hashed refresh sessions, and atomic rotation. [Auth route declarations](../backend/internal/router/auth.go) live in `internal/router`; [user persistence](../backend/internal/repository/user.go) lives in `repository/user.go`. Registration assigns participant using the existing role repository. Prior full backend race tests and vet passed; the route and repository moves each passed targeted tests, vet, and PostgreSQL HTTP integration tests.

## Active Gaps and Risks

- The [user seed integration test](../backend/seeds/test/users_postgres_test.go) checks profile timestamps against a fixed ten-second window. Full race instrumentation exceeded that window during bcrypt work; the ordinary suite passed. See the [normalized failure](commands/topics/go-failures.md#failure-20261017-001). Reliable full seed race validation needs a separate change to the elapsed-time assertion.
- Language guidance conflicts: `project.md` requires English project memory, while `AGENTS.md` requires English instructions and `knowledge_accumulation.md` requires English Markdown. The user's explicit English-language direction is retained for these command instructions.
- User-memory paths differ: `AGENTS.md` and `project.md` specify `AI_USER_LEARN`, while command and memory skills use `AI_USER_LEARN`.
- `module_context_loading.md` describes itself as extending itself. Research backlog and coverage files referenced by the research skill are not yet available.
- All literal favorite team IDs in `user_info.go` must exist; the seed intentionally fails and rolls back if a referenced ID is absent. Broader business-module boundaries remain unresearched.
- Applied migration files should remain consistent with the version already recorded by `schema_migrations`; use a new version for later schema changes.
- API startup requires distinct JWT secrets and valid TTLs. Registration requires the participant role. Access tokens remain usable until expiry after logout or blocking; no blacklist or refresh-session cleanup scheduler is implemented.

## Next Step

For future race validation of the expanded user seeds, replace the fixed freshness window with timestamp bounds tied to the seed invocation; keep this test maintenance separate from constant extraction.

Use `Authorization: Bearer <access_token>` when integrating the user-info endpoint into the frontend; no user ID or body is needed.

Run the normal development seeds to populate local logo paths in an existing database. Keep future seed slugs and the [asset inventory](../docs/assets/team-logos.md) in sync.

Apply pending tournament migrations through `docker compose run --rm migrate up`, then run `docker compose run --rm seed` to populate the tournament and prize rules in the intended environment. Use migration 13 or later for future schema changes. Environments adopting authentication must apply migration 9, provision participant, and configure JWT environment variables before API startup. The earlier renumbering must not be applied as a routine in-place migration on populated external databases. Align the remaining guidance conflicts in a separate instruction-maintenance task.
