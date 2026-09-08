# Go Command Patterns

<a id="pattern-20260908-001"></a>
## [go] Query the local Go version

- ID: PATTERN-20260908-001
- When to use: A preflight check needs the locally installed Go toolchain version.
- Preflight: Confirm `go` is available. Use `backend/` as the project Go working directory and inspect `backend/go.mod` separately for module requirements. No package, test, build tags, or application environment is needed for this version query.
- Correct command / template: `GOTOOLCHAIN=local go version`
- Verification: Executed on 2026-09-08 from `backend/`; printed the Go version and exited with code `0`.
- Avoid: `go --version`; it is not a supported version-query flag.
- Project note: `GOTOOLCHAIN=local` selects the local toolchain for this inspection. Do not carry this override into builds or tests without checking module requirements. This command does not validate dependencies, compile the backend, or run tests; the current local version is not a project version requirement.
- Related failures: [FAILURE-20260908-001](go-failures.md#failure-20260908-001).
