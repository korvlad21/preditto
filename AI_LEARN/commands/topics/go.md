# Go commands

Run package commands from `backend/`, whose [go.mod](../../../backend/go.mod) declares module `preditto`. Check its Go directive and the [Docker toolchain](../../../backend/DockerFile) when choosing an environment. Do not store the user's installed version as a project requirement.

<a id="pattern-20260908-001"></a>
<a id="failure-20260908-001"></a>
## Toolchain and targeted checks

`GOTOOLCHAIN=local go version` inspects the installed toolchain without an automatic selection/download; `go --version` is invalid. Do not carry the local override into builds/tests without checking compatibility.

Choose existing packages and test names based on the changed behavior. For example:

```sh
go test -mod=readonly -count=1 ./internal/handler
go vet -mod=readonly ./internal/handler
```

Use `-race` for affected concurrent/auth behavior when supported. Check verbose output or the existing test declarations to confirm a `-run` filter matches; a successful result with no tests is not verification. Expand to `go test -mod=readonly -count=1 ./...` and `go vet -mod=readonly ./...` for a broad change when appropriate. Avoid incidental `go get` or `go mod tidy`.

For formatting, inspect changed Go files with `gofmt -l`; write only files in scope. `go run ./cmd/api` starts the configured API; `go run ./cmd/seed` writes development data. Follow the [setup guide](../../../README.md) and [seed guide](../../../backend/seeds/README.md) before starting them.

## PostgreSQL test prerequisites

- HTTP integration tests require `PREDITTO_AUTH_TEST_DSN`; seed integration tests require `PREDITTO_SEED_TEST_DSN`.
- Set only the needed DSN(s) to a disposable database with schema-creation privileges. Do not print/store connection secrets.
- Tests create isolated schemas, apply migrations and clean up. Without the respective variable, database scenarios skip; package success alone does not prove integration coverage.
- Full-suite integration verification requires both variables. Module documents link the relevant test files and narrower commands.

<a id="failure-20261017-001"></a>
## Known seed race limitation

The [user seed integration test](../../../backend/seeds/test/users_postgres_test.go) checks profile timestamps against a fixed ten-second window. PostgreSQL CURRENT_TIMESTAMP uses the transaction start, before bcrypt work for all seeded users. Race instrumentation exceeded the window in the recorded 2026-10-17 run, producing `recent=false` despite correct fields.

The ordinary full suite with both DSNs passed in that historical run. Removing `-race` avoids that recorded timing failure but does not repair the test or establish full race coverage. Use targeted auth/HTTP race checks for their behavior; reliable seed race coverage requires separately changing the time assertion to bounds tied to the seed execution. [Historical evidence](../../archive/command-history.md#go-failure-20261017-001).

<a id="failure-20260912-001"></a>
## Driver initialization lesson

Do not assume `sql.Open` validates all driver configuration or connects. Inspect the selected driver and [database initialization](../../../backend/internal/database/postgres.go), including connector validation, PingContext and connection timeout propagation. This historical failure was repaired; it is not a current startup defect. [Original evidence](../../archive/command-history.md#go-failure-20260912-001).

## Swagger generation and drift checks

From the repository root, run `make swagger` after API annotation/schema changes
and `make swagger-check` before completion. The latter compares temporary output
and tests actual route coverage, security metadata and environment gating without
PostgreSQL. See the [Swagger guide](../../../backend/docs/swagger.md).

Verified generator lessons: scan actual Go package directories discovered with
`go list`, rather than the package-less `internal/` directory; otherwise aliased
DTO imports can fail with `cannot find type definition`. Pin `--packageName swagger`
when generating in a temporary directory: swag otherwise derives the package
name from the random directory, causing drift (and invalid Go for dotted names).
The Makefile applies both corrections; validate them with `make swagger-check`.
