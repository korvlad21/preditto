# Current iteration

Date: 2026-10-25

## Outcome

Registration now always selects the short refresh mode on the backend; its DTO
has no remember_me field and ignores client-supplied values. Login retains
optional Remember Me. Explicit false
selects configurable short refresh lifetime (12h default); true selects the existing
long lifetime (720h in the environment example). Omission/null preserves legacy
behavior. The signed refresh claim carries the mode through the existing atomic
rotation path; access tokens and public response envelopes are unchanged. Existing
refresh tokens remain supported, with no database migration or frontend changes.

The [auth module](modules/auth.md) and [API contract](../backend/docs/auth.md#remember-me)
own behavior, configuration, compatibility details and source/test references.

## Verification

Passed targeted registration, login, Remember Me, refresh/logout, rollback,
concurrent rotation and invalid-request tests in auth, auth DTOs and handlers
against disposable PostgreSQL 18.6. The rollback scenarios also passed with
race detection after updating their failure-injection constraint to allow the
revoked registration session created before login. The initial race run exposed
this test-fixture mismatch; production transaction behavior did not change.
Affected-package go vet, UTF-8, local documentation references, formatting and
git diff whitespace checks passed. No deployed API or application database was
used to validate the feature.

## Next step

Clients can opt in by sending a boolean remember_me on login. Registration
always issues a short session, including when the client sends remember_me=true.
Configure JWT_REFRESH_TOKEN_SHORT_TTL and JWT_REFRESH_TOKEN_TTL for deployment;
short TTL must exceed access TTL and be less than long TTL. Client token storage
is outside this backend change. Existing unrelated risks remain in
[system state](system_state.md).

The previous documentation reorganization is retained in AGENTS.md, the module
and command indexes, and the conditional workflows; no additional archive is needed.
