# Current Iteration

Date: 2026-10-04

Added public `POST /api/teams/get_all_teams` through the existing DTO/model, repository, service, handler, and router layers, wired in [main.go](../backend/cmd/api/main.go). The [teams contract](../backend/docs/teams.md) defines optional JSON/query `country`, exact filtering, all-country sentinels, the six response fields, nullable values, and empty arrays. SQL uses a bound country parameter and deterministic ID ordering; no schema or dependency changes were needed.

Targeted handler/router/repository/service/model/teams-DTO package checks and API/affected-package vet passed. The new HTTP tests passed with `-race -count=1` against isolated PostgreSQL 18.6 schemas, covering missing/empty/ALL country, exact matching, query/JSON precedence, injection input, nullable fields, empty tables, malformed requests, and database-error secrecy. Local test infrastructure was disposable. The previous iteration is preserved in [archive](archive/2026-09-27-user-repository-iteration.md).
