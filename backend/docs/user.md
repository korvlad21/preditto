# User API

All responses follow the [shared API contract and client migration guide](responses.md).

`POST /api/user/get_user_info` returns the authenticated user's account and profile in `data`.
No request body or parameters are required. Send the access JWT in
`Authorization: Bearer <access_token>`; refresh tokens are not accepted.

`RequireAuth` verifies the token and sets `middleware.UserIDKey` in the Gin context.
`LoadUser` then calls `UserService.GetUserInfo` once, using that ID, and stores
a `*model.UserInfo` snapshot in Gin Context. The handler reads it through
`middleware.CurrentUser(c)` without querying the database. Query parameters and
request bodies cannot select another user. Responses use `Cache-Control: no-store`.

Example HTTP 200 response:

```json
{
  "success": true,
  "data": {
    "id": 81,
    "username": "new_user",
    "email": "new_user@example.com",
    "status": "ACTIVE",
    "created_at": "2026-01-02T03:04:05.123456Z",
    "first_name": "Vladislav",
    "last_name": "Korobkin",
    "avatar_url": null,
    "favorite_team_id": 5,
    "selected_tournament_id": null
  }
}
```

`created_at` is the account creation timestamp from `users.created_at`, serialized
as an RFC 3339 string with fractional seconds when present. It is loaded in the
existing query and remains available when the profile is missing.

Status is returned as stored in `users`, without case conversion. The five profile
fields preserve SQL NULL as JSON `null`, and empty strings remain empty strings.
A missing `user_info` row returns HTTP 200 with all five profile fields set to
`null`. A missing user, including a deleted account with an unexpired access token,
returns HTTP 404 with `USER_NOT_FOUND`.

Missing, invalid, or expired access tokens return HTTP 401 with `UNAUTHORIZED`.
Database errors return HTTP 500 with `INTERNAL_ERROR` and no database details.
Errors use the shared `{"success":false,"error":"CODE"}` envelope.
The middleware's existing behavior is retained: account blocking or logout does
not immediately invalidate an issued access token; no new role checks are added.

`cmd/api/main.go` creates a shared middleware chain with `RequireAuth` followed by
`LoadUser`. `router.RegisterUserRoutes` receives a protected `/api` group;
`RegisterTournamentRoutes` receives the same protected group. Public auth and team routes
remain on their public groups. `LoadUser` reuses `UserService` and the `UserRepository`
in `internal/repository/user.go`. A single parameterized `LEFT JOIN`
selects the ten response columns; it never
reads password hashes. `model.UserInfo` is the response shape because the existing
auth `model.User` does not preserve nullable names or include avatars.

The snapshot is loaded afresh for every authenticated request; selection is not
stored in JWT or cached between requests. A tournament change committed before
the query is visible even with the same token. Concurrent changes after the query
do not alter the current request snapshot. A handler that changes selection must
update `CurrentUser(c).SelectedTournamentID` after a successful database commit
so subsequent consumers in the same request see its change. Context is not an
authorization check for tournament membership. The [tournament selection
endpoint](tournaments.md#select-a-tournament) updates this snapshot after saving.

Run HTTP tests from `backend/`:

```sh
go test -mod=readonly ./internal/handler -run '^TestGetUserInfo' -count=1
```

The PostgreSQL test reuses the isolated-schema fixture and requires
`PREDITTO_AUTH_TEST_DSN` pointing to a PostgreSQL test database. It covers response
fields, secrecy, NULL and empty values, missing profiles and users, and isolation
between authenticated users. Authentication and database-error tests run without
a live database.
