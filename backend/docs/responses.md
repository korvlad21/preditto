# JSON API Response Contract

This is a breaking change for existing clients of `/api/*`. Deploy compatible
clients together with the backend. There is no legacy-envelope negotiation.

## Responses

Success with data:

```json
{"success":true,"data":{"teams":[]}}
```

Failure:

```json
{"success":false,"error":"INVALID_REQUEST"}
```

Successful operation without data:

```json
{"success":true,"data":null}
```

Success has exactly `success` and `data`; failure has exactly `success` and
`error`. There is no public `message` or nested error object. Public error codes
are centralized in `internal/response/error.go`. Existing string values are
preserved; `NOT_FOUND` identifies unmatched API paths and unsupported methods.
Clients translate these stable codes into Russian, English, or another locale.
Unknown codes should use a generic localized error fallback.

| POST route | Success status | `data` |
| --- | --- | --- |
| `/api/auth/register` | 201 | Existing user and token fields |
| `/api/auth/login` | 200 | Existing user and token fields |
| `/api/auth/refresh` | 200 | Existing token fields |
| `/api/auth/logout` | 200 | `null` |
| `/api/user/get_user_info` | 200 | Existing account and profile fields |
| `/api/teams/get_all_teams` | 200 | `{"teams":[...]}` |
| `/api/teams/get_all_countries` | 200 | `{"countries":[...]}` |
| `/api/tournaments/get_all_tournaments` | 200 | `{"tournaments":[...]}` |
| `/api/tournaments/select_tournament` | 200 | `null` |

Empty collections remain `[]`. Nullable model fields remain `null`; empty strings,
zeroes, and `false` retain their existing meaning. Business fields are unchanged.
Status codes remain unchanged except logout, which changes from 204 to 200.
Responses retain `Content-Type: application/json; charset=utf-8`, existing
`Cache-Control` / `Pragma` policies, and the `WWW-Authenticate: Bearer` challenge
for access authentication failures. Endpoint guides list their error codes.

## Routing and internal errors

`router.New` installs Gin logging and panic recovery, an unanswered-error fallback,
and API-specific `NoRoute` handling. Route registration must use this engine.
Unknown `/api/*` paths return HTTP 404 with `NOT_FOUND`. Unsupported methods on
known paths also retain HTTP 404 with `NOT_FOUND`; automatic 405 handling is
disabled. OPTIONS has no implicit CORS success response. HEAD receives the same
404 status and JSON content type, but the HTTP server suppresses its body.
Gin's existing trailing-slash redirects remain unchanged.

Unhandled `c.Error` entries or pending 5xx statuses produce HTTP 500 with
`INTERNAL_ERROR` when no response has been sent. Recovered panics use the same
contract. Gin retains error and panic diagnostics server-side. Already committed
headers/bodies are never rewritten or appended to by this fallback; a disconnected
client cannot receive a replacement response. This is not a JSON interception
middleware: handlers create their responses with `response.Success` and
`response.AbortError` directly.

Only paths starting with `/api/` receive these routing/error fallbacks. `/`,
`/hello`, `/health`, static files, and other non-API routes keep their existing
behavior. `/health` still returns `{"status":"ok"}`. The bare `/api` path is
outside this prefix. Protected routes retain `RequireAuth` then `LoadUser`, with
the same shared `CurrentUser` pointer and one user load per request.

## Client migration

1. Read successful payloads from `body.data`: for example, tokens from
   `body.data.access_token`, profile fields from `body.data`, and teams from
   `body.data.teams` instead of the former top-level array.
2. Read `body.error` as a string instead of `body.error.code`. Replace use of
   `body.error.message` with the frontend translation dictionary.
3. Accept HTTP 200 and parse `data: null` for logout. Tournament selection also
   now includes `data: null`; do not assume it returns only `success`.
4. Keep HTTP-status handling and token-refresh rules. A successful JSON parse
   does not imply a successful operation: check `success` and the HTTP status.
5. Handle `NOT_FOUND` for unsupported methods and unknown paths; handle HEAD
   without attempting JSON parsing. Do not apply this API decoder to health,
   HTML, static resources, or other non-API responses.

## Verification

From `backend/`, run:

```sh
go test -race -mod=readonly -count=1 ./internal/response ./internal/middleware ./internal/router ./internal/handler
go vet -mod=readonly ./internal/response ./internal/middleware ./internal/router ./internal/handler ./cmd/api
```

Set `PREDITTO_AUTH_TEST_DSN` to a disposable PostgreSQL URL to run the actual
handler integration tests. Without it, database tests skip explicitly. Tests
cover data envelopes, empty lists, nullable values, logout, middleware headers,
shared user state, routing failures, panic/error diagnostics, committed responses,
and unchanged non-API behavior.
