# Auth

Reviewed: 2026-10-25, forced-short registration and targeted auth/HTTP tests against disposable PostgreSQL; rollback scenarios passed with race detection and affected-package vet passed.

## Purpose and API

Owns registration, credential verification, JWT issuance, refresh-session rotation and logout. Read this for password rules, tokens, sessions or authentication failures. Account/profile storage belongs to [users](users.md).

| POST route | Successful result |
| --- | --- |
| `/api/auth/register` | 201, user and token pair |
| `/api/auth/login` | 200, user and token pair |
| `/api/auth/refresh` | 200, replacement token pair |
| `/api/auth/logout` | 200, success without data |

Exact fields, validation bounds, errors and examples live in the [auth contract](../../backend/docs/auth.md), which uses the [shared response contract](../../backend/docs/responses.md). Routes are public, but refresh/logout validate the submitted refresh credential. Auth requests are limited to 16 KiB and responses have no-store/no-cache headers.

Auth handler Swagger annotations reuse the auth request/response DTOs and User model. See the [Swagger guide](../../backend/docs/swagger.md) for generation and shared envelope ownership.

## Business behavior

- Registration validates the DTO, hashes the password with bcrypt, creates an ACTIVE account and profile, assigns the existing participant role by code, and persists the first refresh session. Database uniqueness and favorite-team constraints remain authoritative; IDs are obtained from the database.
- Login accepts username or email with case-sensitive lookup. Unknown users still trigger a bcrypt comparison; wrong-password and unknown-user responses are identical. Only ACTIVE accounts may log in or refresh.
- Access and refresh JWTs use HS256 with different configured secrets, independent random JTIs, and distinct token types. Parsing requires the expected type/signature, expiration, issued-at time, positive user subject and nonempty JTI.
- Registration always issues a short refresh session with signed `remember_me = false`; its DTO has no mode field and client-supplied `remember_me` is ignored. Login accepts optional `remember_me`: true uses configured long refresh TTL (720h in the example), false uses short TTL (12h default), and omission/null preserves legacy behavior. A signed refresh claim carries the mode through rotation; refresh body fields cannot override it. Access TTL and response fields are unchanged. Existing JWTs without the claim remain supported; no schema migration is needed.
- Refresh is single-use rotation: revoke the old session and store a replacement atomically. The client must serialize refresh requests and replace its pair after success; a committed rotation cannot be retried with the old credential.
- Logout revokes only the submitted session. Repeating logout for a still-unexpired revoked session succeeds. A blocked account may log out. Access tokens remain valid until expiry; logout or blocking does not immediately revoke them.

## Transactions and concurrency

[AuthService](../../backend/internal/service/auth.go) receives both AuthRepository and UserRepository. `AuthRepository.InTransaction` passes callback-scoped instances backed by the same `sql.Tx`. All transactional user/session operations must use these instances, not the service's pool-backed UserRepository.

Registration commits account, profile, role assignment and refresh session together, and returns tokens only after commit. Login rechecks password hash and account status under `LockUser` before issuing a session. Refresh locks the user first (`FOR SHARE`), then the session (`FOR UPDATE`), matching parent-first deletion order. Session expiration and hash are checked after locks have been acquired. Concurrent rotations serialize and only one succeeds. A failure creating the replacement rolls back revocation. Logout locks the session before revoking it.

## Sources, models and tables

| Source | Responsibility |
| --- | --- |
| [router/auth.go](../../backend/internal/router/auth.go), [handler/auth.go](../../backend/internal/handler/auth.go) | Route declarations, request binding and HTTP error mapping |
| [service/auth.go](../../backend/internal/service/auth.go) | Use cases and transaction orchestration |
| [repository/auth.go](../../backend/internal/repository/auth.go) | Refresh-session SQL and shared transaction wrapper |
| [repository/user.go](../../backend/internal/repository/user.go), [user_roles.go](../../backend/internal/repository/user_roles.go) | User creation, lookup, locking and role assignment |
| [auth/token.go](../../backend/internal/auth/token.go) | Signing, parsing and refresh hashing |
| [request DTOs](../../backend/internal/dto/auth/request.go), [response DTOs](../../backend/internal/dto/auth/response.go) | Register/Login/RefreshRequest and TokenResponse/AuthResponse |
| [User](../../backend/internal/model/user.go), [RefreshToken](../../backend/internal/model/refresh_token.go) | Auth response account and persisted session data; hashes are excluded from JSON |

[Migration 9](../../backend/migrations/000009_create_refresh_tokens.up.sql) defines `refresh_tokens`: JTI primary key, required user FK with ON DELETE CASCADE, 32-byte token hash, expiration/creation/revocation timestamps, and user/expiration indexes. Only the SHA-256 hash of the high-entropy refresh JWT is persisted. Passwords use bcrypt, not this token hash. Accounts, profiles and role relationships are documented in [users](users.md#models-and-persistence).

## Dependencies and limits

[Config](../../backend/internal/config/config.go) requires distinct secrets of at least 32 bytes and valid access/short-refresh/long-refresh TTLs; the [database-only seed CLI](../../backend/cmd/seed/main.go) does not require JWT settings. Registration needs a provisioned participant role. There is no access blacklist, token-family compromise detection, refresh cleanup scheduler, or role/permission authorization middleware. Existing access-token authentication does not enforce account status on every request. See users for the separate database snapshot middleware.

## Verification

From `backend/`, run `go test -race -mod=readonly -count=1 ./internal/auth ./internal/dto/auth ./internal/middleware ./internal/router ./internal/handler` for changes spanning auth behavior. Narrow the package/filter when the affected scope permits; use the [Go guide](../commands/topics/go.md) for prerequisites.

[HTTP integration tests](../../backend/internal/handler/auth_postgres_test.go) require a disposable `PREDITTO_AUTH_TEST_DSN`. They cover validation, uniqueness, nonsequential IDs, registration rollback at profile/role/session stages, login secrecy and blocking, refresh rejection/rollback, concurrent rotation, logout and migration 9 rollback. Remember Me tests cover configured expiration boundaries, signed-mode tampering, invalid JSON types, forced-short registration (including ignored client overrides) and login mode selection, repeated rotation, ignored refresh-body overrides and persisted expiration. Rollback and concurrent rotation tests exercise legacy, short and remembered modes. Token and middleware tests cover algorithm, signature, type, claim and header validation. Passing a run without the DSN does not verify database behavior.
