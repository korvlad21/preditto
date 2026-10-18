# Tournaments API

`POST /api/tournaments/get_all_tournaments` requires an access token in
`Authorization: Bearer <access_token>` and returns a JSON array ordered by name,
then ID. No matches, including an empty table, return `[]` with HTTP 200.
No request body is required. GET is not registered and returns HTTP 404.

The optional query parameter `status` matches `tournaments.status` exactly.
Omitted `status`, an empty value, and `all` return every tournament, including
rows with a null status. Other values are case-sensitive and are not trimmed;
unknown values return an empty array. For example:

```text
POST /api/tournaments/get_all_tournaments?status=active
```

Each object contains exactly these fields:

| Field | JSON value |
| --- | --- |
| `id` | Tournament ID |
| `name` | Tournament name |
| `type_of_tournament` | String or null |
| `status` | String or null |
| `starts_at` | RFC 3339 timestamp or null |
| `ends_at` | RFC 3339 timestamp or null |
| `cost_amount` | Number or null |
| `amount_currency` | String or null |
| `prize_first_place` | Percentage for place 1, or null |
| `prize_second_place` | Percentage for place 2, or null |
| `prize_third_place` | Percentage for place 3, or null |
| `count_participant` | Number of distinct active user IDs, or 0 |
| `is_selected_tournament` | Whether the ID matches the authenticated user's selected tournament |

`RequireAuth` and `LoadUser` run on the shared protected API group.
`TournamentHandler` reads `middleware.CurrentUser`; no additional user query is
made. Selection is evaluated from the current request's profile snapshot, so
changes appear on the next request with the same access token. A null selection
or missing profile produces `false` for every tournament. Responses use
`Cache-Control: no-store`.

`TournamentService` normalizes `all` and sets the selection flag.
`TournamentRepository` reads the list in one parameterized query. Three prize
joins each match at most one row using the `(tournament_id, place)` primary key;
active participant counts are grouped before joining to avoid multiplying rows.
The model is the response shape, following the teams API pattern; the request
DTO only binds the query string.

Invalid authentication returns HTTP 401 (`UNAUTHORIZED`); a missing user returns
404 (`USER_NOT_FOUND`). Binding errors return 400 (`INVALID_REQUEST`), and
database failures return 500 (`INTERNAL_ERROR`) without database details.

Run targeted tests from `backend/`:

```sh
go test -mod=readonly ./internal/handler -run '^TestGetAllTournaments' -count=1
```

The PostgreSQL test requires `PREDITTO_AUTH_TEST_DSN` pointing to a disposable
PostgreSQL database. It applies the project migrations in an isolated schema
and removes that schema afterward.
