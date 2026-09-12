# Command Pattern Catalog

Search for the ID or symptom; open the specified file and only the relevant section. The recipe date limits the evidence: versions, environment, and applicability must be verified before execution.

Only linked records describe verified commands; empty inherited headings are placeholders.

## Go Toolchain

- [FAILURE-20260912-001](topics/go-failures.md#failure-20260912-001) — `lib/pq` configuration validation is deferred by `sql.Open`; explicit connector initialization is verified.

- [PATTERN-20260908-001](topics/go-patterns.md#pattern-20260908-001) — query the local Go version with `GOTOOLCHAIN=local go version`.
- [FAILURE-20260908-001](topics/go-failures.md#failure-20260908-001) — unsupported `--version` flag; `flag provided but not defined: -version` (controlled exercise).

## Database and Doctrine

- [PATTERN-20260912-001](topics/docker-patterns.md#pattern-20260912-001) — PostgreSQL migrations through Docker Compose, including manual up and down.
- [FAILURE-20260912-001](topics/docker-failures.md#failure-20260912-001) — use `postgres:///` with PG environment variables; `postgres://` loses URL syntax during normalization.

## Frontend and Build

## MCP and AI

## Queues and Worker Processes

## Search, Git, and Editing

## Integrations and Data

## Merged Entries

## Local Entries
