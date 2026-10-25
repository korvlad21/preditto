# Architecture

Reviewed against source on 2026-10-22. This describes the checkout, not a running environment or a new test execution.

## System map

Preditto has a Go HTTP backend and a React/TypeScript/Vite frontend. The backend is organized by technical layer inside one `preditto` Go module, not by separate business-module packages. Functional documentation connects the relevant files across layers.

| Area | Entry point and responsibility |
| --- | --- |
| API composition | [cmd/api/main.go](../backend/cmd/api/main.go) loads configuration, opens PostgreSQL, constructs concrete dependencies, and registers routes |
| Routing | [router.New](../backend/internal/router/router.go) installs logging, error fallback and recovery; individual router files declare endpoints |
| HTTP handlers | Bind requests, call services, translate errors and use the [shared response contract](../backend/docs/responses.md) |
| Services | Coordinate use cases and repository calls using `context.Context` |
| Repositories | Parameterized `database/sql` operations through `lib/pq`; user and auth operations can share a transaction |
| Models and DTOs | Models include response projections and constants; request DTOs carry binding rules. They are not a complete domain-object or ORM layer |
| Database startup | [Compose](../docker-compose.yaml) waits for PostgreSQL health, migration success, seed success, then starts the API |
| Seed CLI | [cmd/seed](../backend/cmd/seed/main.go) uses database-only configuration and the transactional [development runner](../backend/seeds/development/seed.go) |
| Frontend | [main.tsx](../frontend/src/main.tsx) mounts [App.tsx](../frontend/src/App.tsx), currently the Vite starter UI; see [frontend setup](../frontend/README.md) |

The API process does not run migrations or seeds. `/health` checks process responsiveness, not database readiness. Host setup and environment requirements are in the [root README](../README.md); migration and seed procedures live in their respective [migration guide](../backend/migrations/README.md) and [seed guide](../backend/seeds/README.md).

During development, Vite forwards relative `/api` requests to Go without path
rewriting. The default target is localhost:8080; Compose supplies the internal
backend:8080 address through API_PROXY_TARGET. This development-only routing is
disabled for build and preview and adds no backend CORS policy. Production must
provide its own same-origin `/api` routing. Configuration and client examples
live in the [frontend guide](../frontend/README.md#assets-and-api-integration).

## Functional boundaries

Use the [module index](modules/index.md) to select sources and tests.

| Module | Owns | Depends on |
| --- | --- | --- |
| [Auth](modules/auth.md) | Credentials, token issuance and refresh-session lifecycle | User persistence and participant-role assignment |
| [Users](modules/users.md) | Accounts, profiles, shared current-user snapshot and role persistence | Team and tournament references; auth establishes identity |
| [Teams](modules/teams.md) | Team/country reference data and logo paths | Country FK and frontend static assets |
| [Tournaments](modules/tournaments.md) | Listing, selection coordination, participant/prize projections | User profile selection and authenticated snapshot |
| [Matches](modules/matches.md) | Match schema, scores, status and schedule storage | Tournament stages and shared team references |
| [Predictions](modules/predictions.md) | Predicted scores and stored calculation results | Tournament participants and matches |
| [Rankings](modules/rankings.md) | Stored stage and cumulative participant rankings | Tournament participants and stages |

Countries share the teams document. Participation and prize rules belong to tournaments. Role tables and assignment belong to users; an authorization policy layer is not implemented. Prediction submission, match scoring, payments, payouts, tournament join/leave endpoints and application frontend screens are not implemented. Predictions currently have schema only. Do not infer these features from names, costs or seed data.

## Shared request and transaction behavior

All nine current `/api/*` endpoints use POST. Public auth/team routes and protected user/tournament routes are wired separately. Protected routes requiring `CurrentUser` use `RequireAuth` then `LoadUser`; the latter loads one shared profile snapshot per request. See [users](modules/users.md#current-user-snapshot) for its lifecycle and [responses](../backend/docs/responses.md) for envelopes, status codes, fallback errors and non-API exceptions.

`AuthRepository.InTransaction` supplies auth and user repository instances backed by one `sql.Tx`. AuthService uses the callback instances for atomic registration and session issuance/rotation; see [auth transaction rules](modules/auth.md#transactions-and-concurrency). Ordinary profile reads and selection writes use UserRepository directly. No generic transaction manager is implemented.

## Shared constants

Domain constants live in the relevant [user](../backend/internal/model/user.go), [role](../backend/internal/model/role.go), [tournament](../backend/internal/model/tournament.go) and [participation](../backend/internal/model/tournament_participant.go) model files. Account statuses are uppercase; tournament and participation statuses are lowercase. Constants alone do not imply database or application validation.

Public error-code strings are centralized in [response/error.go](../backend/internal/response/error.go). Service/repository errors remain in their owning packages. PostgreSQL violation codes are local to user persistence; JWT types and Bearer literals remain in auth/middleware. SQL and independent test assertions retain their contract literals.

## Persistence and verification boundaries

The current schema has sixteen migration pairs and sixteen application tables. Rankings currently have schema only; position calculation and points awarding are not implemented. Module documents explain important relationships; SQL remains authoritative for exact types and constraints. Existing baseline migrations 4 and 10 were revised for selected tournament storage: an already applied older baseline will not be replayed by normal `up`. See [active constraints](system_state.md).

Source review establishes implementation behavior, not deployed state. PostgreSQL HTTP tests require `PREDITTO_AUTH_TEST_DSN`; seed tests require `PREDITTO_SEED_TEST_DSN`. Without them, integration tests skip. Use the [command index](commands/index.md) for checks and the known seed race limitation. Local toolchain, database versions, credentials and runtime observations belong only in ignored `AI_USER_LEARN`.
