# Teams API

All responses follow the [shared API contract and client migration guide](responses.md).

`POST /api/teams/get_all_teams` is public and returns a team array in `data.teams`,
ordered by `name`. The optional `country` parameter is read only from the JSON body.
Query parameters are ignored. Send `Content-Type: application/json`.
An empty body or `{}` is valid.

Omitted `country`, `""`, and `"ALL"` return every team. Other values match
`teams.country` exactly, with no case conversion or whitespace trimming.
For example, `{"country":"GBR"}` filters by the country code, not a numeric ID.

Each object contains `id`, `name`, `short_name`, `slug`, `country`, and `logo_url`.
Nullable `short_name` and `logo_url` are returned as JSON `null` when absent in
PostgreSQL. No matches, including an empty table, return `data.teams: []` with HTTP 200.

Example response:

```json
{
  "success": true,
  "data": {
    "teams": [
      {
        "id": 11,
        "name": "British team",
        "short_name": "BRI",
        "slug": "british-team",
        "country": "GBR",
        "logo_url": "/logos/teams/british-team.svg"
      }
    ]
  }
}
```

Invalid JSON or an incompatible `country` type returns HTTP 400 with the shared
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

## Single team

`POST /api/teams/get_team` is public and requires a JSON body containing `id`,
a positive signed 64-bit integer. Query parameters do not supply or override it.

Request:

```json
{"id":11}
```

HTTP 200 response:

```json
{
  "success": true,
  "data": {
    "team": {
      "id": 11,
      "name": "British team",
      "short_name": "BRI",
      "slug": "british-team",
      "country": "GBR",
      "logo_url": "/logos/teams/british-team.svg"
    }
  }
}
```

The response reuses `model.Team`, including nullable `short_name` and `logo_url`.
Missing, null, non-positive, out-of-range or incorrectly typed IDs and malformed
JSON return HTTP 400 with `{"success":false,"error":"INVALID_REQUEST"}`.
A missing team returns HTTP 404 with `{"success":false,"error":"TEAM_NOT_FOUND"}`.
Database failures return HTTP 500 with `INTERNAL_ERROR`, without database details.

`TeamHandler.GetTeam` validates `GetTeamRequest`, then calls `TeamService.GetTeam`
and `TeamRepository.GetTeam`. The repository performs one parameterized query
with `WHERE id = $1` and maps `sql.ErrNoRows` to the shared `repository.ErrNotFound`.
It shares the six-column projection with the list query.

Run `go test -mod=readonly ./internal/handler -run '^TestGetTeam' -count=1`
from `backend/`. PostgreSQL cases require `PREDITTO_AUTH_TEST_DSN`; they cover
exact fields, nullability, full int64 IDs, JSON-body selection and missing teams.
Validation, method and database-error checks also run without PostgreSQL.

## Favorite team

`POST /api/teams/get_favorite_team` requires an access JWT in
`Authorization: Bearer <access_token>`. No body or parameters are required;
request fields cannot choose a user or team.

`RequireAuth` followed by `LoadUser` loads the current account/profile once.
The handler reads `middleware.CurrentUser(c).FavoriteTeamID`, already selected
from `user_info.favorite_team_id`, without querying the profile again.
It shares team retrieval, error mapping and response formatting with `GetTeam`
through the existing `TeamService.GetTeam` and `TeamRepository.GetTeam` methods.
There is one team lookup when a favorite ID is present, and none when it is null.

HTTP 200 uses exactly the same `data.team` shape as `get_team`, including nullable
`short_name` and `logo_url`. Responses use `Cache-Control: no-store`.
When `favorite_team_id` is null (including a missing profile), HTTP 200 returns:

```json
{"success":true,"data":{"team":null}}
```

| HTTP status | Error | Meaning |
| --- | --- | --- |
| 401 | `UNAUTHORIZED` | Missing, invalid or expired access token |
| 404 | `USER_NOT_FOUND` | The authenticated account no longer exists |
| 404 | `TEAM_NOT_FOUND` | The team referenced by the loaded snapshot no longer exists |
| 500 | `INTERNAL_ERROR` | Database failure or missing current-user context |

`RegisterProtectedTeamRoutes` is wired to the shared protected group in
`cmd/api/main.go`. The other team routes remain public.
Run `go test -mod=readonly ./internal/handler -run '^TestGetFavoriteTeam' -count=1`
from `backend/`, with `PREDITTO_AUTH_TEST_DSN` for PostgreSQL scenarios.

## Countries

`POST /api/teams/get_all_countries` is public and requires no request parameters
or body. It returns every row from `countries`, ordered by `name`, as an array in `data.countries`.
Each object contains `id`, `name`, `short_name`, and `created_at`. The timestamp
uses Go's `time.Time` JSON representation (RFC 3339 with fractional seconds).
An empty table returns `data.countries: []` with HTTP 200. Database errors use the same HTTP 500
`INTERNAL_ERROR` envelope as teams, without exposing database details.

`model.Country` is the response shape. `CountryRepository`, `CountryService`, and
`CountryHandler` follow the teams dependency pattern; no request or response DTO
is needed. The endpoint shares `internal/router/team.go` and is wired in
`cmd/api/main.go`.

Country HTTP tests reuse the same isolated PostgreSQL schema fixture:

```sh
go test -mod=readonly ./internal/handler -run '^TestGetAllCountries' -count=1
```
