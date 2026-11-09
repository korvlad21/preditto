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
    "selected_tournament_id": null,
    "roles": ["participant"]
  }
}
```

`created_at` is the account creation timestamp from `users.created_at`, serialized
as an RFC 3339 string with fractional seconds when present. It is loaded in the
existing query and remains available when the profile is missing.

`roles` contains all assigned role codes from `user_roles` joined to `roles`,
ordered by code. It is always an array: users without assignments receive `[]`.
Roles are loaded in the same SQL query as the account/profile, including when
the profile is missing. Returning roles does not add authorization checks or
change the registration/login response shapes.

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
`RegisterTournamentRoutes` receives the same protected group. Registration uses
the same authentication/loading chain plus RequireRole(admin); login, refresh,
logout and team-directory routes remain public. `LoadUser` reuses `UserService` and the `UserRepository`
in `internal/repository/user.go`. A single parameterized query uses a `LEFT JOIN`
for the profile and an `ARRAY` subquery for role codes; it never
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
fields, secrecy, NULL and empty values, missing profiles and users, no/single/multiple
roles, role ordering, role changes with the same token, and isolation between
authenticated users. Authentication and database-error tests run without
a live database.

## Update a user

`PUT /api/user/update_user` uses `RequireAuth -> LoadUser`. The handler checks
`CurrentUser(c).ID` and its role codes: the caller must be the target's owner or
have `model.RoleCodeAdmin`. The body selects the target, never the caller's
identity. Query selectors are ignored. No additional status restriction applies.

Send `Authorization: Bearer <access_token>` and `Content-Type: application/json`:

```json
{
  "user_id": 81,
  "email": "updated@example.com",
  "favorite_team_id": 5,
  "first_name": "Alex",
  "last_name": "Smith",
  "avatar_url": ""
}
```

| Field | Validation and update semantics |
| --- | --- |
| `user_id` | Required positive signed 64-bit integer; account must exist |
| `email` | Required valid email, at most 255 characters; case-sensitive uniqueness among other accounts; retaining one's own email succeeds |
| `favorite_team_id` | Optional integer; positive IDs must reference a team, zero clears the reference to SQL NULL |
| `first_name`, `last_name` | Optional strings, at most 100 characters; empty strings are preserved |
| `avatar_url` | Optional string, at most 500 characters; empty clears to SQL NULL; nonempty must exactly match the target's currently stored URL |

Omitted or JSON-null optional fields are unchanged. Unknown fields, invalid
types, NUL characters, invalid UTF-8, multiple JSON values and oversized bodies
are rejected. JSON bodies are limited to 16 KiB. Usernames, passwords, status,
roles, account creation time and selected tournament are preserved.

UserRepository locks the target account/profile and updates `users` and
`user_info` in one transaction, creating a missing profile. Database uniqueness
and favorite-team foreign keys provide race-safe validation; a profile failure
rolls back the email and any inserted profile. Concurrent edits of a target are
serialized. HTTP 200 returns updated `UserInfo` in the shared `success/data`
envelope with `Cache-Control: no-store`. After commit, self-updates refresh the
existing shared snapshot. An admin editing another account keeps their own snapshot.

### Upload an avatar

Use the same PUT route with `multipart/form-data`: exactly one text field `data`
containing the JSON request and exactly one file field `avatar`. Omit
`avatar_url` or set it empty when uploading. Let the client generate the boundary;
do not manually set Content-Type with browser FormData.

```sh
curl -X PUT http://localhost:8080/api/user/update_user \
  -H 'Authorization: Bearer <access_token>' \
  --form-string 'data={"user_id":81,"email":"updated@example.com","first_name":"Alex"}' \
  --form 'avatar=@/absolute/path/avatar.png'
```

Limits: complete multipart body 5 MiB + 20 KiB; JSON field 16 KiB; image file
5 MiB. Only decoded JPEG/PNG content is accepted, at most 4096 pixels per
dimension. Configuration is checked before allocating pixels, then the image
is fully decoded, transparency flattened on white and JPEG encoded at quality
85. Client filenames, extensions and MIME declarations do not select the format
or path. Re-encoding removes source metadata and appended content.

Files receive a cryptographically random UUID v4 plus `.jpg` under
`frontend/images/users/`; only `/users/<uuid>.jpg` is stored in the database.
No client-selected filesystem path is read/written or URL fetched. JSON can
retain the target's existing URL or clear it; new paths require an upload.

Run the backend from `backend/`: it uses `../frontend/images/users`. Compose
mounts that host directory at `/frontend/images/users` for the backend; the
frontend's existing mount and Vite's `images` public directory serve `/users/`.
Production hosting must persist this directory and map `/users/` to it; a static
build does not include later uploads. The backend adds no static route.
Uploaded files are ignored by Git.

The image is written and closed before the SQL transaction. A write failure
prevents database changes; a known database failure removes the new file.
Cleanup failures return an internal error and retain server-side details.
An uncertain COMMIT outcome retains the new file to avoid deleting an image
the database may reference. Replaced/cleared files are also retained to avoid
concurrent-request deletion races. Filesystem and SQL cannot share a transaction:
process termination can leave an unreferenced file. Reconciliation is outside
this endpoint.

| HTTP status | Error code |
| --- | --- |
| 400 | `INVALID_REQUEST` (validation, image or avatar path); `FAVORITE_TEAM_NOT_FOUND` |
| 401 | `UNAUTHORIZED` (missing/invalid/expired access token or missing current-user context) |
| 403 | `FORBIDDEN` (neither owner nor admin) |
| 404 | `USER_NOT_FOUND` (missing caller or target account) |
| 409 | `EMAIL_ALREADY_EXISTS` |
| 413 | `INVALID_REQUEST` (request, JSON field or image too large) |
| 415 | `INVALID_REQUEST` (unsupported/missing request Content-Type) |
| 500 | `INTERNAL_ERROR` (database/filesystem error without internal details) |

[Update HTTP tests](../internal/handler/update_user_test.go) cover validation,
authorization, missing profiles, rollback, snapshot updates, concurrent email
uniqueness and avatar handling. Run from `backend/` with
`go test -race -mod=readonly -count=1 ./internal/handler -run '^TestUpdateUser'`.
Database cases require disposable `PREDITTO_AUTH_TEST_DSN`; validation/access
tests run without PostgreSQL.
