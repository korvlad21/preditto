# Go Command Failures

<a id="failure-20261017-001"></a>
## [go] Seed profile freshness window under race instrumentation

- ID: FAILURE-20261017-001
- Date: 2026-10-17
- Context: Go 1.23.3, module `preditto`, no build tags, both PostgreSQL test DSNs set to a disposable database; working directory `backend/`.
- Original command (test URLs omitted): `PREDITTO_AUTH_TEST_DSN='<disposable PostgreSQL URL>' PREDITTO_SEED_TEST_DSN='<same URL>' go test -race -mod=readonly -count=1 ./...`.
- Symptom / exit code: `TestUsersAndUserInfoPostgres` reported correct identity, status, and profile fields but `recent=false`; exit 1. The individual test took 30.36 seconds. Other tested packages passed, including HTTP integration tests.
- Cause: [the test](../../../backend/seeds/test/users_postgres_test.go) requires `updated_at` within the last ten seconds. The [seed runner](../../../backend/seeds/development/seed.go) starts one transaction before [bcrypt hashing for 20 users](../../../backend/seeds/development/users.go); PostgreSQL `CURRENT_TIMESTAMP` remains the transaction-start time. Race instrumentation made the work exceed the assertion's window.
- Incorrect assumption: a full race run would fit the existing fixed freshness window. Literal substitution confirmed the constant extraction preserved seed values and execution logic; the test was unchanged.
- Correct validation invocation (same disposable test DSNs): `go test -mod=readonly -count=1 ./...`.
- Verification: the ordinary full backend suite passed, including PostgreSQL seed and HTTP tests; full-module vet passed. Removing race instrumentation is an alternative validation scope, not a repair of the timing-sensitive test.
- Rule: diagnose `recent=false` against the SQL interval and transaction duration before changing seed data or status constants. Reliable full race coverage needs timestamp bounds tied to the seed invocation in a separate test-maintenance change; that correction is not implemented or verified here.

<a id="failure-20260912-001"></a>
## [go] Driver configuration validation is deferred by sql.Open

- ID: FAILURE-20260912-001
- Date: 2026-09-12
- Context: Go 1.23.3, `lib/pq v1.12.3`, no build tags or special Go flags; module `github.com/korvlad21/firstGoWeb`, working directory `backend/`.
- Original command: `go test -mod=readonly ./internal/config ./internal/database`.
- Symptom / exit code: `TestNewPostgresReturnsOpenError` expected `open PostgreSQL:` for invalid SSL mode but received `ping PostgreSQL:`; exit code 1.
- Cause / incorrect assumption: `sql.Open` was assumed to validate this driver's settings immediately. This driver implements `Driver.Open`, so parsing was deferred until connection establishment.
- Correction: Parse with `pq.NewConfig`, create a connector with `pq.NewConnectorConfig`, then use `sql.OpenDB` and `PingContext`. Propagate the caller deadline into `ConnectTimeout` to bound the handshake too.
- Correct invocation: The same targeted test command after correcting initialization.
- Verification: Targeted tests, `go test -mod=readonly ./...`, and `go vet -mod=readonly ./...` pass.
- Rule: Verify the selected driver's initialization and context behavior rather than assuming `sql.Open` establishes or validates a connection. See [implementation](../../../backend/internal/database/postgres.go) and [tests](../../../backend/internal/database/postgres_test.go).

<a id="failure-20260908-001"></a>
## [go] Unsupported version flag

- ID: FAILURE-20260908-001
- Date: 2026-09-08
- Classification: Intentional, user-requested workflow exercise; not an accidental failure or application defect. The CLI syntax correction is reusable.
- Context: Host shell invocation with `GOTOOLCHAIN=local`; no container, build tags, package execution, or dependency changes.
- Working directory: `backend/`, relative to the project root.
- Module / package / test selection: `backend/go.mod` declares `github.com/korvlad21/firstGoWeb`; no package or test was selected.
- Original command: `GOTOOLCHAIN=local go --version`
- Symptom / error / exit code: `flag provided but not defined: -version`; exit code `2`. The usage output lists `version` as a subcommand.
- Cause: `--version` was passed as a top-level flag instead of using the `version` subcommand.
- Incorrect assumption: Go accepts the generic `--version` convention used by some other CLIs.
- Correct invocation: `GOTOOLCHAIN=local go version`
- Verification result: The corrected invocation printed the local Go version and exited with code `0` in the same directory. This verifies the invocation, not the backend build or tests.
- How to recognize early: For a Go version check, use the documented subcommand. Do not assume flags are shared between tools or retry the unsupported flag.
- Related pattern: [PATTERN-20260908-001](go-patterns.md#pattern-20260908-001).
