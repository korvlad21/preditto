# Current iteration

Date: 2026-10-25

## Outcome

Implemented backend-only Remember Me for registration and login. Explicit false
selects configurable short refresh lifetime (12h default); true selects the existing
long lifetime (720h in the environment example). Omission/null preserves legacy
behavior. The signed refresh claim carries the mode through the existing atomic
rotation path; access tokens and public response envelopes are unchanged. Existing
refresh tokens remain supported, with no database migration or frontend changes.

The [auth module](modules/auth.md) and [API contract](../backend/docs/auth.md#remember-me)
own behavior, configuration, compatibility details and source/test references.

## Verification

Passed targeted race tests for auth, config, auth DTOs, handlers, middleware and
router with PREDITTO_AUTH_TEST_DSN set to disposable PostgreSQL 18.6. Coverage
includes both modes and legacy requests, expiration boundaries, invalid types,
signed-mode tampering, repeated rotation, persisted expiration, ignored body
overrides, logout, transaction rollback and concurrent refresh requests.
Backend-wide go vet, Compose configuration validation, UTF-8, local documentation
references, formatting and git diff whitespace checks passed. No deployed API or
application database was used to validate the feature.

## Next step

Clients can opt in by sending a boolean remember_me on registration/login.
Configure JWT_REFRESH_TOKEN_SHORT_TTL and JWT_REFRESH_TOKEN_TTL for deployment;
short TTL must exceed access TTL and be less than long TTL. Client token storage
is outside this backend change. Existing unrelated risks remain in
[system state](system_state.md).

The previous documentation reorganization is retained in AGENTS.md, the module
and command indexes, and the conditional workflows; no additional archive is needed.
