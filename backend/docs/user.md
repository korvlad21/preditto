# User API

`POST /api/user/get_user_info` returns the authenticated user's account and profile.
No request body or parameters are required. Send the access JWT in
`Authorization: Bearer <access_token>`; refresh tokens are not accepted.

`RequireAuth` verifies the token and sets `middleware.UserIDKey` in the Gin context.
The handler uses only this ID. Query parameters and request bodies cannot select
another user. Responses use `Cache-Control: no-store`.

Example HTTP 200 response:

```json
{
  "id": 81,
  "username": "new_user",
  "email": "new_user@example.com",
  "status": "ACTIVE",
  "first_name": "Vladislav",
  "last_name": "Korobkin",
  "avatar_url": null,
  "favorite_team_id": 5
}
```

Status is returned as stored in `users`, without case conversion. The four profile
fields preserve SQL NULL as JSON `null`, and empty strings remain empty strings.
A missing `user_info` row returns HTTP 200 with all four profile fields set to
`null`. A missing user, including a deleted account with an unexpired access token,
returns HTTP 404 with `USER_NOT_FOUND`.

Missing, invalid, or expired access tokens return HTTP 401 with `UNAUTHORIZED`.
Database errors return HTTP 500 with `INTERNAL_ERROR` and no database details.
Errors use the existing `{"error":{"code":"...","message":"..."}}` envelope.
The middleware's existing behavior is retained: account blocking or logout does
not immediately invalidate an issued access token; no new role checks are added.

`router.RegisterUserRoutes` attaches authentication in `internal/router/user.go`
and is wired in `cmd/api/main.go`. `UserHandler` and `UserService` reuse the existing
`AuthRepository` user persistence in `internal/repository/user.go`. A single
parameterized `LEFT JOIN` selects exactly the eight response columns; it never
reads password hashes. `model.UserInfo` is the response shape because the existing
auth `model.User` does not preserve nullable names or include avatars.

Run HTTP tests from `backend/`:

```sh
go test -mod=readonly ./internal/handler -run '^TestGetUserInfo' -count=1
```

The PostgreSQL test reuses the isolated-schema fixture and requires
`PREDITTO_AUTH_TEST_DSN` pointing to a PostgreSQL test database. It covers response
fields, secrecy, NULL and empty values, missing profiles and users, and isolation
between authenticated users. Authentication and database-error tests run without
a live database.
