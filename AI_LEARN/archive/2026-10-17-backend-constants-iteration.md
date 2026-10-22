# Current Iteration

Date: 2026-10-17

Extracted participant, tournament, and user statuses plus role codes into their owning [model files](../architecture.md#shared-constants). Development seeds and registration reuse those constants. JWT types and the Bearer scheme belong to auth; API error codes and repeated messages belong to response; PostgreSQL violation codes remain local to user persistence.

Literal substitution verified equivalence of all 16 edited Go files; two new model files contain only constants. Values, case, SQL, and business logic are preserved. Independent literal assertions and fixtures in existing tests remain unchanged, as do SQL migrations and seed reference data.

Full backend `go test -mod=readonly -count=1 ./...` passed with both PostgreSQL test DSNs pointing to a disposable database; `go vet -mod=readonly ./...`, formatting, strict UTF-8, and diff checks passed. The full race run passed every other tested package but failed the existing seed profile freshness assertion: its ten-second timestamp window was exceeded during bcrypt work. The [diagnosis and verified non-race invocation](../commands/topics/go.md#failure-20261017-001) preserve this limitation. The disposable database was removed; the application database was not seeded.

Previous iteration: [tournament participants seed](2026-10-16-tournament-participants-iteration.md).
