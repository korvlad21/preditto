# Authentication API

All responses follow the [shared API contract and client migration guide](responses.md).

The existing Gin entry point wires `handler.AuthHandler` to `service.AuthService`,
`repository.AuthRepository`, `repository.UserRepository`, and the existing
`database/sql` PostgreSQL pool.
`router.RegisterAuthRoutes` owns all four auth route declarations and their
shared request limit and cache headers. The handler owns HTTP binding and responses.
`UserRepository` in `repository/user.go` owns user/profile writes and user lookups;
its `AssignRole` method lives in `repository/user_roles.go`.
`AuthRepository` in `repository/auth.go` owns refresh-session persistence and the
transaction wrapper. `AuthService` receives both repositories. `InTransaction`
passes transaction-scoped auth and user repositories backed by the same `sql.Tx`;
registration, login, and refresh use those instances for all transactional work.
Registration therefore commits or rolls back the user, profile, role assignment,
and first refresh session together.
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
JWT_REFRESH_TOKEN_SHORT_TTL=12h
```

Generate each secret independently, for example with `openssl rand -hex 32`.
Never commit actual secrets. `config.Load` validates the API configuration at
startup: secrets must differ; TTLs must be positive whole seconds; access TTL
cannot exceed one hour; short refresh TTL must exceed access TTL and be shorter
than the long refresh TTL. The short TTL defaults to `12h` when unset or empty.
The existing `JWT_REFRESH_TOKEN_TTL` controls the long and legacy lifetime;
Compose and the environment example use `720h` (30 days). The seed CLI uses
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
| POST | `/api/auth/logout` | 200, `{"success":true}` |

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

## Remember Me

Login accepts optional JSON `remember_me`:

| Value | Refresh lifetime |
| --- | --- |
| `true` | `JWT_REFRESH_TOKEN_TTL` (configured example: 720h / 30 days) |
| `false` | `JWT_REFRESH_TOKEN_SHORT_TTL` (default: 12h) |
| Omitted or `null` | Existing `JWT_REFRESH_TOKEN_TTL` behavior for compatibility |

For example, login accepts
`{"login":"korvlad21","password":"password","remember_me":false}`.
Strings, numbers, arrays and objects for this field return 400 `INVALID_REQUEST`.
Access lifetime, response envelopes and `expires_in` (access lifetime in seconds)
are unchanged. No frontend or cookie storage is implemented by this API.

Registration always selects `remember_me = false` on the backend and issues a
short refresh token. The registration DTO has no `remember_me` field; if a client
sends it, it is ignored like other unknown fields. This intentionally changes
registration without the field from a long session to a short session.

The selected mode is embedded in the signed refresh JWT, never trusted from the
refresh request body. Rotation copies the verified mode to the new token and
uses the currently configured TTL for that mode, starting from issuance time.
There is no absolute session lifetime cap. Existing refresh JWTs without the
claim remain valid and retain legacy behavior during rotation. No schema change
is needed: the existing stored hash binds the entire signed token, including
its mode, to the session. Logout and single-use rotation retain their existing
transactional behavior. Deploying this feature does not change the expiration
of tokens already issued.

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
  "success": true,
  "data": {
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

After the session commits, the login handler attempts a synchronous email
notification to the account email when mail is enabled. SMTP failure or disabled
mail does not change the HTTP 200 response or token pair. A send may delay the
response up to the configured SMTP timeout (or request cancellation). Rejected
login and registration do not send notifications.

The repository looks up username or email using parameterized SQL. The service
uses `bcrypt.CompareHashAndPassword`; unknown logins also run a bcrypt comparison
against an unused hash to reduce the timing difference. Wrong passwords and
unknown logins both return exactly:

```json
{
  "success": false,
  "error": "INVALID_CREDENTIALS"
}
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
  "success": true,
  "data": {
    "access_token": "<new access JWT>",
    "refresh_token": "<new refresh JWT>",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

The TokenManager requires HS256, a valid signature using the refresh secret,
expiration, issued-at time, positive user subject, nonempty JTI, and
`token_type = refresh`. JWTs contain `sub`, `iat`, `exp`, `jti`, and
`token_type`. Refresh JWTs additionally contain `remember_me`: false for registration, or the
explicitly selected login mode; access JWTs do not. Both JTIs are generated
independently with `crypto/rand`.

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
under the same row lock. It returns 200 with `{"success":true}`; repeating logout for that
unexpired revoked token returns the same response. Invalid, missing, or expired tokens
return 401. A blocked user may still log out. Other login sessions are unaffected.
The access token remains valid until its short expiration; there is no blacklist.

The protected `POST /api/user/get_user_info` endpoint returns the authenticated
user's account and nullable profile; see [the user API contract](user.md).
Routes requiring the shared current-user snapshot use the same chain as the API entry point:

```go
protected := router.Group("/api", middleware.RequireAuth(tokens), middleware.LoadUser(userService))
```

Here `userService` implements UserInfoLoader. RequireAuth alone establishes token identity;
LoadUser then loads the profile snapshot for handlers using CurrentUser. See the
[users module](../../AI_LEARN/modules/users.md#current-user-snapshot).

Clients send `Authorization: Bearer <access JWT>`. Middleware validates the access
key, HS256, required claims, expiration, and `token_type = access`, then stores an
`int64` under `middleware.UserIDKey` (`"user_id"`) in the Gin context. It does not
query PostgreSQL or apply roles. A status change therefore prevents new login and
refresh but does not immediately invalidate existing access JWTs.

## Errors

All auth and middleware errors use `{"success":false,"error":"CODE"}`.
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

From `backend`, select the affected auth packages:

```sh
go test -mod=readonly -race -count=1 ./internal/auth ./internal/config ./internal/repository ./internal/service ./internal/handler ./internal/middleware ./internal/router
go vet -mod=readonly ./internal/auth ./internal/config ./internal/repository ./internal/service ./internal/handler ./internal/middleware ./internal/router
```

Set `PREDITTO_AUTH_TEST_DSN` to a disposable PostgreSQL database URL for HTTP
integration coverage. Auth tests create isolated schemas, apply migrations and
remove them afterward; the database account needs schema-creation privileges.
Each pooled connection uses the same isolated search path for concurrent requests.
Without the DSN, database scenarios explicitly skip.

Seed integration checks separately require `PREDITTO_SEED_TEST_DSN`. See the
[Go runbook](../../AI_LEARN/commands/topics/go.md) for full-suite prerequisites
and the known seed timestamp assertion failure under race instrumentation.

Coverage includes the requested register/login/refresh/logout/middleware cases,
minimal JWT claims, wrong algorithms and signatures, wrong token types signed
with the correct key, missing/expired/tampered sessions, nonsequential IDs,
registration rollback at multiple stages, rotation rollback, concurrent rotation,
and migration 000009 down/up.

JWT parsing options follow the [golang-jwt documentation](https://golang-jwt.github.io/jwt/usage/parse/).
