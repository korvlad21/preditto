# Current Iteration

Date: 2026-10-21

[response.Success](../backend/internal/response/success.go) now omits data when called with nil, returning {"success":true}. Logout and tournament selection keep HTTP 200 and use this compact body. Non-nil payloads, empty collections, nullable fields inside payloads, false, zero, and empty strings retain their values. Existing HTTP expectations and [contract documentation](../backend/docs/responses.md) were updated. The implementation uses omitempty on the existing interface field without new abstractions.

Verification: `go test -race -mod=readonly -count=1 ./internal/response ./internal/router ./internal/handler` passed. PostgreSQL-dependent tests were not rerun and skipped without their DSN; their no-data body assertions were updated and compiled. Diff, formatting, strict UTF-8, documentation JSON examples, and local links passed. No unexpected command failures occurred.

Next step: clients must accept omitted data on successful operations without payloads.

Previous iteration: [unified JSON API contract](archive/2026-10-21-json-api-contract-iteration.md).
