# Users

Reviewed: 2026-11-08; profile updates, avatar persistence, rollback and concurrent email uniqueness verified with race-enabled HTTP tests against disposable PostgreSQL. Existing role-loading and RequireRole coverage is retained.

## Purpose and API

Owns account/profile persistence, the request's current-user snapshot, selected tournament storage and role assignment persistence. Read this for user fields, identity isolation, nullable profiles, roles or user repository changes.

`POST /api/user/get_user_info` returns the authenticated account/profile with HTTP 200. It takes no user selector; request/query fields cannot select another account. Exact fields, errors and examples are in the [user contract](../../backend/docs/user.md). Registration/login and token policy belong to [auth](auth.md); tournament selection coordination belongs to [tournaments](tournaments.md).

The response includes the account creation timestamp from `users.created_at` through `UserInfo.CreatedAt`. The existing joined query loads it into the shared snapshot without additional database queries, including when no profile exists.

`UserInfo.Roles` returns all assigned role codes ordered by code, or `[]` when none are assigned. GetUserInfo loads them through an `ARRAY` subquery joining `user_roles` to `roles` within the existing account/profile query. No extra database round trips or row multiplication are introduced, and roles remain available without a profile. The frontend UserInfo type mirrors this additive field; auth User responses are unchanged.

The user handler Swagger annotation reuses UserInfo, including its nullable profile fields. See the [Swagger guide](../../backend/docs/swagger.md) for generation and shared envelope ownership.

`PUT /api/user/update_user` updates the target's email and supplied profile fields for the owner or an admin. The handler checks the loaded CurrentUser ID/roles. UserRepository owns the transaction/locks, creates a missing profile, relies on named database constraints for email/favorite-team races, and preserves credentials, status, roles and tournament selection. Successful self-updates refresh the shared snapshot; an admin editing another account keeps their own snapshot. Exact optional-field/null semantics, validation, statuses and multipart examples are in the [update contract](../../backend/docs/user.md#update-a-user).

UserService coordinates database writes with [AvatarStore](../../backend/internal/service/avatar.go), which decodes bounded JPEG/PNG input and writes real JPEG files with UUID v4 names under `frontend/images/users`; SQL stores `/users/<uuid>.jpg`. JSON may retain only the target's stored URL or clear it. Known database failures remove newly written files; uncertain commit outcomes and replaced/cleared files retain them to avoid dangling references. Crash atomicity and unreferenced-file reconciliation remain filesystem/SQL limitations. Compose shares uploads with the frontend; production must persist and serve `/users/`. See the contract for deployment details.

## Current-user snapshot

`RequireAuth` establishes UserIDKey; `LoadUser` then calls UserService.GetUserInfo once and stores a shared `*model.UserInfo` in Gin context. `UserHandler` reads `middleware.CurrentUser` without another database query. The chain is attached in [main.go](../../backend/cmd/api/main.go) and shared with tournament routes.

A missing account produces USER_NOT_FOUND. A missing `user_info` row is allowed for reads: the LEFT JOIN returns null profile fields. Empty strings stay distinct from null. Snapshots are loaded on each request, so committed selection and role changes are visible with the same access token on a subsequent request. Concurrent writes after loading do not change the current snapshot. After a successful profile write, handlers must update the shared snapshot if later consumers need the changed value. A failed write must not update it.

The snapshot is not itself a tournament-membership or role authorization check. [RequireRole](../../backend/internal/middleware/role.go) runs after RequireAuth and LoadUser, checks an exact role code in the shared snapshot without database queries, returns 403 FORBIDDEN for a missing role, and fails closed with 401 UNAUTHORIZED if the snapshot is absent. Registration uses it with RoleCodeAdmin. LoadUser does not reject BLOCKED status. Access-token lifetime and blocking policy are documented in auth.

## Models and persistence

| Artifact | Meaning and constraints |
| --- | --- |
| [User / UserInfo](../../backend/internal/model/user.go) | User is the auth response shape and excludes PasswordHash from JSON; UserInfo preserves nullable profile fields and includes avatar/selection |
| [users, migration 1](../../backend/migrations/000001_create_users.up.sql) | Unique username/email, required password hash/status, identity ID and update trigger; status strings are constants, not a database CHECK |
| [user_info, migration 4](../../backend/migrations/000004_create_user_info.up.sql) | Unique required user_id with cascade deletion; nullable names/avatar; favorite team FK uses SET NULL; update trigger |
| [selection FK, migration 10](../../backend/migrations/000010_create_tournaments.up.sql) | Nullable selected_tournament_id references tournaments with ON DELETE SET NULL |
| [roles](../../backend/migrations/000005_create_roles.up.sql), [permissions](../../backend/migrations/000006_create_permissions.up.sql) | Unique codes; roles has no updated_at column |
| [user_roles](../../backend/migrations/000007_create_user_roles.up.sql), [role_permissions](../../backend/migrations/000008_create_role_permissions.up.sql) | Composite primary keys, cascading parent references and reverse-lookup indexes |

UserRepository.CreateUser performs both account and profile inserts and must run inside the auth transaction callback. UpdateUser owns an independent account/profile transaction. FindUserByLogin is used before credential verification; LockUser obtains a shared account-row lock for session issuance. GetUserInfo never reads the password hash.

UpdateSelectedTournament changes only an existing profile. A nil ID clears selection at repository level; zero affected rows return ErrNotFound. Uniqueness and named FK errors are mapped in userWriteError, including ErrTournamentNotFound for selection races. The HTTP selection endpoint does not expose nil/reset behavior.

## Roles and seeds

[AssignUserRole / RemoveUserRole](../../backend/internal/repository/user_roles.go) accept a database or caller-owned transaction. Assignment is idempotent on the user/role pair; removing an absent pair succeeds. UserRepository.AssignRole resolves a role code before insertion. Auth registration assigns participant; this does not create tournament participation.

[Development users](../../backend/seeds/development/users.go) and [profiles](../../backend/seeds/development/user_info.go) have 20 matching entries. The [role assignments](../../backend/seeds/development/user_roles.go) preserve the explicit admin roles for korvlad21 and hanna_stoma and participant for boroda, then append every remaining seeded username with participant once at package initialization. Username deduplication keeps the explicit assignments authoritative and covers all development users without growing the list on reruns. Existing database assignments are preserved. Repeated seeds restore configured account/profile fields and passwords; favorite teams use literal IDs, which must exist. See the [seed guide](../../backend/seeds/README.md) for exact repeat-run behavior. There is no role-management service/API or permission-enforcement middleware.

## Source map and dependencies

- [User routes](../../backend/internal/router/user.go), [handler](../../backend/internal/handler/user.go), [service](../../backend/internal/service/user.go), [repository](../../backend/internal/repository/user.go).
- [Update handler](../../backend/internal/handler/update_user.go), [request DTO](../../backend/internal/dto/users/request.go), [avatar storage](../../backend/internal/service/avatar.go) and [update HTTP tests](../../backend/internal/handler/update_user_test.go).
- [User middleware](../../backend/internal/middleware/user.go) defines UserInfoLoader, LoadUser and CurrentUser; this small consumer interface differs from the otherwise concrete service dependencies.
- [Auth](auth.md) owns credential/token policy and the shared transaction wrapper.
- [Teams](teams.md) supplies favorite-team references; [tournaments](tournaments.md) owns the selection use case.
- Tournament participant FKs do not cascade user deletion; profile/session cascade rules alone do not guarantee unrestricted account deletion.

## Verification

Profile-update tests use `go test -race -mod=readonly -count=1 ./internal/handler -run '^TestUpdateUser'` with disposable `PREDITTO_AUTH_TEST_DSN`. They cover owner/admin isolation, validation, unchanged account fields, missing-profile insertion and rollback, failed-write snapshot preservation, error secrecy, concurrent duplicate emails, real JPEG conversion, generated URLs, invalid/truncated/excessive images, filesystem failures and removal of new files on SQL errors. They do not simulate process termination or ambiguous network failure during COMMIT.

From `backend/`: `go test -race -mod=readonly -count=1 ./internal/middleware ./internal/handler -run 'TestLoadUser|TestUserContextMissing|TestGetUserInfo|TestSelectTournament'`.

[User HTTP tests](../../backend/internal/handler/user_test.go) require disposable `PREDITTO_AUTH_TEST_DSN` for database scenarios. They cover identity isolation, null/empty fields, missing users/profiles, error secrecy, selection refresh, empty/single/multiple roles, code ordering and role changes with the same token. [Middleware tests](../../backend/internal/middleware/user_test.go) cover one load, shared-pointer behavior and role serialization from the snapshot. User creation/role changes also require auth rollback checks; seed changes use `PREDITTO_SEED_TEST_DSN` and the relevant [seed tests](../../backend/seeds/test/users_postgres_test.go). See [Go prerequisites and the seed race limit](../commands/topics/go.md).
