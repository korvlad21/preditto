# Command Pattern Catalog

Search for the ID or symptom; open the specified file and only the relevant section. The recipe date limits the evidence: versions, environment, and applicability must be verified before execution.

Only linked records describe verified commands; empty inherited headings are placeholders.

## Go Toolchain

- [FAILURE-20260912-001](topics/go-failures.md#failure-20260912-001) — `lib/pq` configuration validation is deferred by `sql.Open`; explicit connector initialization is verified.

- [PATTERN-20260908-001](topics/go-patterns.md#pattern-20260908-001) — query the local Go version with `GOTOOLCHAIN=local go version`.
- [FAILURE-20260908-001](topics/go-failures.md#failure-20260908-001) — unsupported `--version` flag; `flag provided but not defined: -version` (controlled exercise).

## Database and Doctrine

- [FAILURE-20260919-001](topics/docker-failures.md#failure-20260919-001) — duplicate version 8 remained after countries moved to migration 2; remove the obsolete one-way file.
- [PATTERN-20260912-001](topics/docker-patterns.md#pattern-20260912-001) — PostgreSQL migrations through Docker Compose, including manual up and down.
- [FAILURE-20260912-001](topics/docker-failures.md#failure-20260912-001) — use `postgres:///` with PG environment variables; `postgres://` loses URL syntax during normalization.
- [PATTERN-20260915-001](topics/docker-patterns.md#pattern-20260915-001) — change an already applied schema through a new migration version.
- [FAILURE-20260915-001](topics/docker-failures.md#failure-20260915-001) — audit and repair a failed migration before clearing its dirty marker.
- [FAILURE-20260915-002](topics/docker-failures.md#failure-20260915-002) — a leading Cyrillic character in seed SQL blocks Compose startup after migrations succeed.
- [FAILURE-20260915-003](topics/docker-failures.md#failure-20260915-003) — historical migration 3 rollback left its trigger function; the down migration now cleans it up.
- [FAILURE-20260916-001](topics/docker-failures.md#failure-20260916-001) — dirty version 3 with a partially committed empty user_info table; inspect objects before repairing version state.
- [FAILURE-20260916-002](topics/docker-failures.md#failure-20260916-002) — teams referenced countries before creation; the migration files were reordered and a local database was rebuilt from a verified backup.

## Frontend and Build

## MCP and AI

## Queues and Worker Processes

## Search, Git, and Editing

## Integrations and Data

## Merged Entries

## Local Entries
