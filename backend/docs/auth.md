# Authentication API

The existing Gin entry point wires `handler.AuthHandler` to `service.AuthService`,
`repository.AuthRepository`, and the existing `database/sql` PostgreSQL pool.
`router.RegisterAuthRoutes` owns all four auth route declarations and their
shared request limit and cache headers. The handler owns HTTP binding and responses.
`repository/user.go` owns user/profile writes and user lookups;
`repository/auth.go` owns refresh-session persistence and the transaction wrapper.
`auth.TokenManager` owns JWT signing, parsing, and refresh-token hashing.

```text
HTTP -> Gin handler -> Auth service -> Repository -> PostgreSQL
                           |
                           +-> TokenManager -> JWT (golang-jwt/jwt/v5)
```

The existing `repository.AssignUserRole` is reused. Authentication does not embed
roles or permissions in JWTs and does not implement RBAC checks.

## Configuration and startup

Set these environment variables alongside the existing PostgreSQL configuration:

```dotenv
JWT_ACCESS_SECRET=<independent random secret, at least 32 bytes>
JWT_REFRESH_SECRET=<different random secret, at least 32 bytes>
JWT_ACCESS_TOKEN_TTL=15m
JWT_REFRESH_TOKEN_TTL=720h
```

Generate each secret independently, for example with `openssl rand -hex 32`.
Never commit actual secrets. `config.Load` validates the API configuration at
startup: secrets must differ; TTLs must be positive whole seconds; access TTL
cannot exceed one hour; refresh TTL must exceed access TTL. The seed CLI uses
`config.LoadPostgres` and does not require JWT secrets.

Apply migration 000009 before running the API. Registration requires the existing
`participant` role created by the development role seed. The API resolves its ID
by code; neither user IDs nor role IDs are assumed. Environments using a different
bootstrap process must provision this role before allowing registration.

From the project root, after configuring `.env`:

```sh
docker compose run --rm migrate up
docker compose up -d --build
```

Compose passes the JWT variables to the API; its normal dependency chain applies
migrations and development seeds before API startup. JWT secrets have no default.

## Routes and validation

| Method | Route | Success |
| --- | --- | --- |
| POST | `/api/auth/register` | 201, user and token pair |
| POST | `/api/auth/login` | 200, user and token pair |
| POST | `/api/auth/refresh` | 200, token pair |
| POST | `/api/auth/logout` | 204, empty body |

Send `Content-Type: application/json`. Auth responses use `Cache-Control: no-store`.
Request bodies are limited to 16 KiB. DTO validation and malformed JSON return 400.

- Username: 3-50 ASCII letters, digits, underscores, dots, or hyphens. Excluding
  `@` keeps username and email login namespaces separate for new registrations.
- First and last name: required, not whitespace-only, at most 100 characters each.
- Email: required, valid format, at most 255 characters.
- Password: at least 8 characters and at most 72 UTF-8 bytes, matching bcrypt's
  input limit. Confirmation must match exactly. Passwords are never trimmed.
- Favorite team: optional or null; when supplied, a positive existing team ID.
- Username and email uniqueness follow the existing case-sensitive PostgreSQL
  constraints. Inputs are not silently lowercased or normalized.

## Registration

```http
POST /api/auth/register
Content-Type: application/json

{
  "username": "korvlad21",
  "first_name": "Vladislav",
  "last_name": "Korobkin",
  "email": "example@example.com",
  "password": "password",
  "password_confirmation": "password",
  "favorite_team_id": 5
}
```

Example response (IDs and JWT placeholders are illustrative):

```json
{
  "access_token": "<access JWT>",
  "refresh_token": "<refresh JWT>",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": 81,
    "username": "korvlad21",
    "email": "example@example.com",
    "status": "ACTIVE",
    "first_name": "Vladislav",
    "last_name": "Korobkin",
    "favorite_team_id": 5
  }
}
```

The service validates the DTO and calls `bcrypt.GenerateFromPassword` at default
cost. One PostgreSQL transaction inserts `users` with `status = ACTIVE` and obtains
its ID with `RETURNING id`, inserts `user_info`, assigns `participant`, and stores
the first refresh session. Profile defaults supply `updated_at` and a null avatar.
Any error, including token/session creation or role assignment, rolls back all
four records. Tokens are returned only after commit.

Database unique constraints handle competing registrations safely. The favorite
team foreign key validates existence in the same transaction, including deletion
races; an invalid team causes the user insert to roll back.

## Login

```http
POST /api/auth/login
Content-Type: application/json

{"login":"korvlad21","password":"password"}
```

`login` may also be `example@example.com`. The 200 response has the same shape as
registration, including the user and a newly persisted token pair. Each successful
login creates an independent session.

The repository looks up username or email using parameterized SQL. The service
uses `bcrypt.CompareHashAndPassword`; unknown logins also run a bcrypt comparison
against an unused hash to reduce the timing difference. Wrong passwords and
unknown logins both return exactly:

```json
{"error":{"code":"INVALID_CREDENTIALS","message":"Invalid login or password"}}
```

Only `ACTIVE` users may log in or refresh. A `BLOCKED` user with a correct password
receives 403; an incorrect password still receives the generic 401. The service
rechecks the user under a shared row lock before issuing a session, protecting
against concurrent status or password changes.

## Refresh rotation

```http
POST /api/auth/refresh
Content-Type: application/json

{"refresh_token":"<current refresh JWT>"}
```

```json
{
  "access_token": "<new access JWT>",
  "refresh_token": "<new refresh JWT>",
  "token_type": "Bearer",
  "expires_in": 900
}
```

The TokenManager requires HS256, a valid signature using the refresh secret,
expiration, issued-at time, positive user subject, nonempty JTI, and
`token_type = refresh`. JWTs contain only `sub`, `iat`, `exp`, `jti`, and
`token_type`. Both JTIs are generated independently with `crypto/rand`.

Within one transaction the service locks and checks the user, then locks the
refresh session with `FOR UPDATE` and checks its user, database and JWT expiration,
revoked state, and SHA-256 hash using constant-time comparison. Expiration is
checked after lock acquisition, including any wait. Locking the user first follows
the same parent-first order as cascading user deletion. The service revokes the
old session, creates a new pair, stores the new hash, and commits. A failed insert or
commit returns no tokens and rolls back revocation. Concurrent refreshes of the
same token serialize: exactly one succeeds; later requests receive 401.

The client must replace its stored pair after success and serialize refresh
requests. Old tokens cannot be retried after a committed rotation, even if the
response was lost. Reuse rejects that token; this implementation does not revoke
other sessions or implement token-family compromise detection.

`refresh_tokens.id` is the refresh JWT's random JTI. The database stores only the
32-byte SHA-256 hash of the complete refresh JWT, never the bearer credential.
`user_id` has a cascading foreign key and an index for future session management;
`expires_at` has an index for future cleanup. Revoked records remain available;
no cleanup scheduler is added by this change.

## Logout and protected endpoints

```http
POST /api/auth/logout
Content-Type: application/json

{"refresh_token":"<current refresh JWT>"}
```

Logout validates the refresh JWT and stored session/hash, then revokes the session
under the same row lock. It returns 204 with no body; repeating logout for that
unexpired revoked token also returns 204. Invalid, missing, or expired tokens
return 401. A blocked user may still log out. Other login sessions are unaffected.
The access token remains valid until its short expiration; there is no blacklist.

Attach middleware to future protected routes:

```go
protected := router.Group("/api", middleware.RequireAuth(tokens))
protected.GET("/me", meHandler) // Supply the endpoint handler when implementing /me.
```

Clients send `Authorization: Bearer <access JWT>`. Middleware validates the access
key, HS256, required claims, expiration, and `token_type = access`, then stores an
`int64` under `middleware.UserIDKey` (`"user_id"`) in the Gin context. It does not
query PostgreSQL or apply roles. A status change therefore prevents new login and
refresh but does not immediately invalidate existing access JWTs.

## Errors

All auth and middleware errors use `{"error":{"code":"...","message":"..."}}`.
PostgreSQL details, JWT library errors, and password/token hashes are never sent
in responses. PasswordHash also has a defensive `json:"-"` model tag.

| Status | Code |
| --- | --- |
| 400 | `INVALID_REQUEST`, `FAVORITE_TEAM_NOT_FOUND` |
| 401 | `INVALID_CREDENTIALS`, `INVALID_TOKEN`, `UNAUTHORIZED` |
| 403 | `USER_BLOCKED` |
| 409 | `USERNAME_ALREADY_EXISTS`, `EMAIL_ALREADY_EXISTS` |
| 500 | `INTERNAL_ERROR` |

## Verification

From `backend`, run:

```sh
go test -mod=readonly ./...
go vet -mod=readonly ./...
```

For actual PostgreSQL integration tests, set `PREDITTO_AUTH_TEST_DSN` and
`PREDITTO_SEED_TEST_DSN` to a disposable test database URL, then run:

```sh
go test -mod=readonly -race -count=1 ./...
```

Without their respective environment variables, PostgreSQL tests skip explicitly.
Auth tests create and remove isolated schemas and apply all migrations. They use
a connection-wide search path in each pooled connection so concurrent requests
operate on the same isolated schema. The database account needs schema creation
privileges. Existing seed tests keep their established test setup.

Coverage includes the requested register/login/refresh/logout/middleware cases,
minimal JWT claims, wrong algorithms and signatures, wrong token types signed
with the correct key, missing/expired/tampered sessions, nonsequential IDs,
registration rollback at multiple stages, rotation rollback, concurrent rotation,
and migration 000009 down/up.

JWT parsing options follow the [golang-jwt documentation](https://golang-jwt.github.io/jwt/usage/parse/).
