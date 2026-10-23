# Teams and countries

Reviewed: 2026-10-24; targeted GetTeam, GetAllTeams and GetAllCountries HTTP tests passed against disposable PostgreSQL, and affected-package go vet passed.

## Purpose and API

Owns team/country reference-data reads, seed identities and logo URL conventions. Countries have their own model, handler, service and repository, but share team routing and this document because their current use cases are small and related.

| Public POST route | Behavior |
| --- | --- |
| `/api/teams/get_all_teams` | Optional country filter; returns data.teams |
| `/api/teams/get_team` | Required positive int64 id in JSON; returns data.team or TEAM_NOT_FOUND |
| `/api/teams/get_all_countries` | No required parameters/body; returns data.countries |

The [teams/countries API contract](../../backend/docs/teams.md) owns exact response fields and errors. Both lists use name ordering and return empty arrays rather than null. No secondary sort is specified for equal names. All three endpoints are public.

## Business and query behavior

GetTeam validates a required positive int64 ID in the JSON body. TeamService forwards it to TeamRepository.GetTeam, which uses one parameterized primary-key lookup and the same projection as the list. A missing row maps to repository.ErrNotFound and HTTP 404 TEAM_NOT_FOUND; invalid input uses INVALID_REQUEST. The response contains one Team in data.team, preserving nullable fields.

TeamHandler binds only the optional JSON body; query parameters are ignored. An empty body or `{}` lists teams without a country filter. Omitted/empty country and uppercase ALL disable the filter. Other values match the country code exactly, without trimming or case conversion. TeamService performs the ALL normalization and TeamRepository uses a bound SQL parameter. CountryService forwards the unfiltered read to CountryRepository.

Nullable team short_name and logo_url remain JSON null when absent. The country response includes created_at; the team response does not. Models serve as response projections; separate response DTOs are not implemented. The API is read-only: no team/country CRUD or roster-management workflow exists.

## Sources, models and tables

| Sources | Responsibility |
| --- | --- |
| [router/team.go](../../backend/internal/router/team.go) | All three public routes |
| [team handler](../../backend/internal/handler/team.go), [country handler](../../backend/internal/handler/country.go) | Binding and response/error mapping |
| [team service](../../backend/internal/service/team.go), [country service](../../backend/internal/service/country.go) | Filter normalization and forwarding |
| [team repository](../../backend/internal/repository/team.go), [country repository](../../backend/internal/repository/country.go) | Parameterized reads and name ordering |
| [Team](../../backend/internal/model/team.go), [Country](../../backend/internal/model/country.go), [team request DTO](../../backend/internal/dto/teams/request.go) | Response projections, optional country filter and required team ID |

[Migration 2](../../backend/migrations/000002_create_countries.up.sql) creates countries with a unique required short_name. [Migration 3](../../backend/migrations/000003_create_teams.up.sql) creates teams with a unique slug and required country FK to countries.short_name, not a numeric country ID. Countries must exist before teams. The schema permits nullable team abbreviations/logos; the three-letter uppercase rule is seed validation, not a universal API/SQL constraint.

## Seeds, assets and dependencies

[Countries](../../backend/seeds/development/countries.go) has 16 development entries; [teams](../../backend/seeds/development/teams.go) has 36. Countries upsert by short_name, teams by slug. Validation checks unique keys, required names, uppercase three-letter abbreviations/country codes and country membership in the dataset. These are explicit development datasets, not a live competition roster.

Team upserts restore name, abbreviation, country and logo_url while preserving IDs/created_at and unrelated rows. Logo URLs are `/logos/teams/<slug>.svg`, supplied on insert and overwritten on conflict. The [asset inventory](../../docs/assets/team-logos.md) is the canonical source for provenance and individual files. Vite's publicDir is `frontend/images`; see [frontend setup](../../frontend/README.md). Do not duplicate the asset list in this module.

[Users](users.md) references teams through favorite_team_id with ON DELETE SET NULL. Development profiles use literal team IDs; valid slugs alone do not make those IDs portable to an arbitrary database. Seed execution and rollback rules live in the [seed guide](../../backend/seeds/README.md).

## Verification

From `backend/`: `go test -race -mod=readonly -count=1 ./internal/handler -run 'TestGetTeam|TestGetAllTeams|TestGetAllCountries'`. Database scenarios need disposable `PREDITTO_AUTH_TEST_DSN`.

GetTeam tests cover required/body-only input, positive int64 validation, exact response fields, nullability, missing teams, unsupported methods and database-error secrecy.

[Team tests](../../backend/internal/handler/team_test.go) cover filter semantics, JSON-only filtering and ignored query parameters, invalid input, injection-like values, nullability and empty lists. [Country tests](../../backend/internal/handler/country_test.go) cover the unfiltered contract, timestamps and database-error secrecy. Seed changes require relevant seed validation/integration tests; logo changes require checking actual Vite URLs and build output as described in the asset inventory. Use the [command index](../commands/index.md) for Go and frontend checks.
