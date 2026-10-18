# Current Iteration

Date: 2026-10-18

Changed /api/tournaments/get_all_tournaments from GET to POST at the user's request. Updated the route, HTTP test helper, rejected-method assertion, API contract, and active knowledge references. Status remains an optional query parameter; no body is required. See [the contract](../../backend/docs/tournaments.md).

Targeted TestGetAllTournamentsRequiresAuth and TestGetAllTournamentsMissingUserContext tests passed: POST reaches the protected handler and GET returns 404. The existing PostgreSQL test now sends POST; it was not rerun because only the HTTP method changed. SQL, service, and response behavior retain the previous PostgreSQL verification. No unexpected command failures occurred.

Next step: use POST in Postman and client integration, with the existing access token and optional status query parameter.

Previous iteration: [tournament API implementation](2026-10-18-tournaments-get-api-iteration.md).
