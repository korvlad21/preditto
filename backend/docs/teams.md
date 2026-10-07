# Teams API

`POST /api/teams/get_all_teams` is public and returns a JSON array of teams,
ordered by `name`. The optional `country` parameter can be supplied in a JSON body
or the query string. A JSON `country` value takes precedence when both are supplied.
An empty body or `{}` is valid.

Omitted `country`, `""`, and `"ALL"` return every team. Other values match
`teams.country` exactly, with no case conversion or whitespace trimming.
For example, `{"country":"GBR"}` filters by the country code, not a numeric ID.

Each object contains `id`, `name`, `short_name`, `slug`, `country`, and `logo_url`.
Nullable `short_name` and `logo_url` are returned as JSON `null` when absent in
PostgreSQL. No matches, including an empty table, return `[]` with HTTP 200.

Example response:

```json
[
  {
    "id": 11,
    "name": "British team",
    "short_name": "BRI",
    "slug": "british-team",
    "country": "GBR",
    "logo_url": "/logos/teams/british-team.svg"
  }
]
```

Invalid JSON or an incompatible `country` type returns HTTP 400 with the existing
`INVALID_REQUEST` error envelope. Database errors return HTTP 500 with
`INTERNAL_ERROR`; database details are logged through Gin and omitted from responses.

The route is registered in `internal/router/team.go` and wired in `cmd/api/main.go`.
`handler.TeamHandler` binds the request DTO, `service.TeamService` handles `ALL`,
and `repository.TeamRepository` reads the six columns with a parameterized filter.
`model.Team` is also the response shape; a separate response DTO is unnecessary.

Run request validation tests from `backend/` with:

```sh
go test -mod=readonly ./internal/handler -run '^TestGetAllTeamsInvalidRequest$'
```

PostgreSQL HTTP integration tests reuse the existing isolated-schema fixture and
require `PREDITTO_AUTH_TEST_DSN` pointing to a PostgreSQL test database:

```sh
go test -mod=readonly ./internal/handler -run '^TestGetAllTeamsPostgres$' -count=1
```

## Countries

`POST /api/teams/get_all_countries` is public and requires no request parameters
or body. It returns every row from `countries`, ordered by `name`, as a JSON array.
Each object contains `id`, `name`, `short_name`, and `created_at`. The timestamp
uses Go's `time.Time` JSON representation (RFC 3339 with fractional seconds).
An empty table returns `[]` with HTTP 200. Database errors use the same HTTP 500
`INTERNAL_ERROR` envelope as teams, without exposing database details.

`model.Country` is the response shape. `CountryRepository`, `CountryService`, and
`CountryHandler` follow the teams dependency pattern; no request or response DTO
is needed. The endpoint shares `internal/router/team.go` and is wired in
`cmd/api/main.go`.

Country HTTP tests reuse the same isolated PostgreSQL schema fixture:

```sh
go test -mod=readonly ./internal/handler -run '^TestGetAllCountries' -count=1
```
