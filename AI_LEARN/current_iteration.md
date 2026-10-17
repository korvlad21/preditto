# Current Iteration

Date: 2026-10-17

Restored JWT types, the Bearer scheme, API error codes, and error messages to their original string literals at the user's request. Removed the corresponding constant declarations and references from auth, response, middleware, and handlers. Domain/model constants and PostgreSQL violation constants remain unchanged; see [constant scope](modules/backend/overview.md#shared-constants).

All seven reverted Go files match their original Git HEAD versions exactly. Targeted `go test -mod=readonly -count=1 ./internal/auth ./internal/middleware ./internal/handler ./internal/response` passed. PostgreSQL-dependent tests were skipped because no test DSN was supplied; the database suite was not rerun for this literal restoration. UTF-8, remaining-reference, and diff checks passed.

One non-reusable editing failure occurred: an initial `apply_patch` used the nonexistent `MessageInvalidError` instead of `MessageInvalidRequest` as context in `handler/team.go`. Patch verification rejected it without changing files. A fresh source search confirmed the actual identifier; reversing the reviewed seven-file diff with a generated exact patch succeeded. Rule: derive reversal context from the current diff instead of retyping identifiers. No application or test command failed in this iteration.

Previous iteration: [backend constant extraction](archive/2026-10-17-backend-constants-iteration.md).
