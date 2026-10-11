# Current Iteration

Date: 2026-10-11

Added protected `POST /api/user/get_user_info` through `UserHandler`, `UserService`, the existing `AuthRepository` user persistence, and `model.UserInfo`. [User routes](../backend/internal/router/user.go) attach `RequireAuth` and are wired in [main.go](../backend/cmd/api/main.go). A single parameterized LEFT JOIN returns eight safe account/profile columns, preserving nullable profile fields without changing the auth model or login/registration behavior. Missing profiles return null fields; missing accounts return HTTP 404 with `USER_NOT_FOUND`. Status is returned without case conversion. The final method was corrected from GET to POST at the user's request. See [the contract](../backend/docs/user.md).

Targeted authentication/database-error tests and affected-package tests with `-race -count=1` passed, including the full handler PostgreSQL suite and existing auth tests. API build and affected-package vet passed. [User HTTP tests](../backend/internal/handler/user_test.go) verify bearer authentication, POST-only routing, exact response fields, no-store headers, NULL/empty values, missing profiles/accounts, error secrecy, and isolation between users even when request fields try to select another account. No migration or dependency changes were needed. The previous iteration is preserved in [archive](archive/2026-10-07-countries-api-iteration.md).

After correcting the HTTP method to POST, all `TestGetUserInfo` tests passed again with `-race -count=1` against isolated PostgreSQL, including rejection of GET requests. API build, diff whitespace, and strict UTF-8 checks passed.

Non-reusable discovery correction from the preceding analysis: `nl -ba backend/internal/router/teams.go` failed because the filename was guessed; `rg --files backend/internal/router` identified `team.go`, and `nl -ba backend/internal/router/team.go` succeeded. Discover paths before reading optional or uncertain filenames. Local optional-file discovery is recorded in `AI_USER_LEARN`.
