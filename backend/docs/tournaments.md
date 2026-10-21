# Tournaments API

All responses follow the [shared API contract and client migration guide](responses.md).

`POST /api/tournaments/get_all_tournaments` requires an access token in
`Authorization: Bearer <access_token>` and returns an array in `data.tournaments` ordered by
`created_at DESC, id DESC`. No matches, including an empty table, return `data.tournaments: []` with HTTP 200.
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

`RequireAuth` and `LoadUser` protect both routes using the shared middleware chain.
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

## Select a tournament

`POST /api/tournaments/select_tournament` uses the same access-token middleware.
Send `Content-Type: application/json` and a body such as `{"id":201}`. The ID must
be a positive signed 64-bit JSON integer; missing IDs and `null` are rejected.
The authenticated user comes exclusively from `middleware.CurrentUser(c)`.
No membership or tournament-status restriction is imposed.

HTTP 200 returns exactly `{"success":true,"data":null}`, including when selecting the same
tournament again. Errors include `success:false` and the shared error
code, for example:

```json
{
  "success": false,
  "error": "TOURNAMENT_NOT_FOUND"
}
```

| HTTP status | Code | Meaning |
| --- | --- | --- |
| 401 | `UNAUTHORIZED` | Missing, invalid, or expired access token |
| 400 | `TOURNAMENT_ID_REQUIRED` | Missing or null ID |
| 400 | `INVALID_TOURNAMENT_ID` | Nonpositive ID, incorrect JSON type, or integer overflow |
| 400 | `INVALID_REQUEST` | Empty, malformed, or incompatible JSON body |
| 404 | `TOURNAMENT_NOT_FOUND` | Tournament does not exist |
| 404 | `USER_INFO_NOT_FOUND` | The current user's profile row is missing |
| 404 | `USER_NOT_FOUND` | The authenticated user no longer exists |
| 500 | `TOURNAMENT_UPDATE_FAILED` | Selection could not be saved |
| 500 | `INTERNAL_ERROR` | User loading or tournament lookup failed, or user context is unavailable |

The router receives the same protected group as the user routes, with
`RequireAuth` before `LoadUser`. All handler and middleware errors use
`response.AbortError`. Internal database details are never exposed.

`TournamentService.SelectTournament` checks existence before saving.
`AuthRepository.UpdateSelectedTournament(ctx, userID, tournamentID *int64)` in
`repository/user.go` updates only the existing `user_info` row; `nil` writes SQL
NULL for future reset operations. Zero affected rows return `ErrNotFound`.
The selected-tournament foreign-key violation maps to `ErrTournamentNotFound`,
covering deletion between the existence check and update. No profile is created
implicitly. The handler updates the shared user snapshot only after the write
succeeds; subsequent requests load the selection afresh with the same token.

Run the selection tests with the same disposable database configuration:

```sh
go test -race -mod=readonly ./internal/handler -run '^TestSelectTournament' -count=1
```
