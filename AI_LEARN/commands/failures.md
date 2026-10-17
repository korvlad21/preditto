# Failure Analyses: Choose a Tool

- [FAILURE-20261017-001: Seed timestamp window under race instrumentation](topics/go-failures.md#failure-20261017-001) — distinguish a fixed freshness assertion from incorrect seed data; ordinary full-module validation passed.

- [FAILURE-20260912-001: Empty PostgreSQL URL](topics/docker-failures.md#failure-20260912-001) — preserve the slash in `postgres:///` for environment-only connection settings.

- [FAILURE-20260908-001: Unsupported Go version flag](topics/go-failures.md#failure-20260908-001) — `go --version` fails; verified through a user-requested controlled exercise.
