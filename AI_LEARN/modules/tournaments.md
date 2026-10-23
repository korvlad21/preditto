# Tournaments

Reviewed: 2026-10-22, source/schema inspection and targeted tournament/user-info HTTP tests with race detection against disposable PostgreSQL; affected-package/API vet passed.

## Purpose and API

Owns tournament listing, user selection coordination and participant/prize projections. Read this for filters, selection, counts, prizes or tournament seeds.

| Protected POST route | Successful result |
| --- | --- |
| `/api/tournaments/get_all_tournaments` | 200, data.tournaments; optional status and is_participate JSON filters, no required body |
| `/api/tournaments/select_tournament` | 200, success without data; JSON id required |

Both routes require RequireAuth followed by LoadUser. The [tournament API contract](../../backend/docs/tournaments.md) owns exact fields, error codes and examples, using the [shared response contract](../../backend/docs/responses.md).

## Execution and business rules

Listing passes CurrentUser.ID and CurrentUser.SelectedTournamentID from handler to service. TournamentService normalizes omitted/empty status or lowercase all to no status filter and marks selected rows without another user read. Other status values match exactly; unknown values return an empty list. TournamentRepository uses one parameterized query, ordered by created_at DESC then id DESC.

The optional is_participate boolean combines with the tournament-status filter using AND. True requires an active tournament_participants row for the current user; false requires its absence; omission imposes no participation restriction. Not-paid and blocked rows are not participation. The repository uses correlated EXISTS/NOT EXISTS in the listing query, with no per-tournament reads and no changes to counts or selection semantics. Filters are read only from the JSON body; query parameters are ignored. An empty body or `{}` applies no filters. Request fields cannot override the authenticated user ID.

Prize joins for places 1/2/3 each match at most one row by the composite key. Active participant counts are grouped before joining and count distinct user IDs, preventing multiplied rows. The query filters participation status, not account status. Missing prize rules stay null and absent active participants yield zero. A null selection, including a missing profile, makes all selection flags false.

Selection accepts only a positive non-null int64 JSON ID and uses the authenticated user's ID. The service checks tournament existence, then UserRepository updates the existing profile. No profile is created implicitly. Missing profiles and tournaments have different public errors. A foreign-key violation maps deletion between the existence check and write to TournamentNotFound. Successful persistence precedes updating the shared CurrentUser pointer.

Selection is not membership or payment and imposes no tournament-status restriction. The repository accepts nil to clear selection, but the HTTP API does not expose reset. Dates do not automatically change tournament status.

## Sources, models and persistence

| Sources | Responsibility |
| --- | --- |
| [router](../../backend/internal/router/tournament.go), [handler](../../backend/internal/handler/tournament.go), [request DTOs](../../backend/internal/dto/tournaments/request.go) | Protected endpoints, input binding, HTTP mapping and snapshot update |
| [service](../../backend/internal/service/tournament.go), [repository](../../backend/internal/repository/tournament.go) | Selection orchestration, existence checks, listing and aggregates |
| [UserRepository](../../backend/internal/repository/user.go), [user middleware](../../backend/internal/middleware/user.go) | Profile selection persistence and shared request snapshot |
| [Tournament model](../../backend/internal/model/tournament.go), [participation constants](../../backend/internal/model/tournament_participant.go) | Response projection, currencies/types/statuses; the participant model file currently contains constants only |

| Table | Important constraints |
| --- | --- |
| [tournaments, migration 10](../../backend/migrations/000010_create_tournaments.up.sql) | Unique slug; nullable status, dates, cost and currency; allowed non-null statuses are planned/active/finished/cancelled |
| [tournament_participants, migration 11](../../backend/migrations/000011_create_tournament_participants.up.sql) | Unique tournament/user pair; required active/not_paid/blocked status; indexes by user and tournament/status |
| [tournament_prize_rules, migration 12](../../backend/migrations/000012_create_tournament_prize_rules.up.sql) | Primary key tournament/place; each percentage between 0 and 100 |
| [user_info, migration 4](../../backend/migrations/000004_create_user_info.up.sql) | Nullable selected_tournament_id; migration 10 adds its FK with ON DELETE SET NULL |

Participant and prize-rule references do not declare cascading deletion. The selection FK alone does not make tournament deletion unrestricted. Unlike users/user_info, tournaments has no automatic updated_at trigger; the seed upsert sets it explicitly. The schema does not require a positive prize place, a total percentage of 95/100, ordered dates or nonnegative cost; do not claim these constraints exist.

## Seeds, dependencies and limits

The [tournament](../../backend/seeds/development/tournaments.go), [participant](../../backend/seeds/development/tournament_participants.go) and [prize](../../backend/seeds/development/tournament_prize_rules.go) seeds share the development transaction. Tournaments upsert by slug; participants resolve tournament IDs by slug and user IDs by username. Membership is explicit, independent of user slice order, and unique within each group. Reruns preserve identities/joined_at while restoring configured participation status. Account and participation statuses are independent.

The first tournament (`uefa-champions-league-2026-2027`) includes `korvlad21` as an active participant. The [seed guide](../../backend/seeds/README.md) owns the three configured tournament datasets, dates, costs and counts. Each currently allocates 95% in prize rules; this is dataset configuration, not an enforced general rule. SQL stores DECIMAL, seeds pass decimal strings, and the response model reads cost/percentages into *float64. Exact monetary calculations are not implemented.

[Users](users.md) owns the selection column and snapshot lifecycle; [auth](auth.md) establishes identity. Join/leave, payment, predictions, scoring, payouts and application frontend integration are not implemented.

## Verification

From `backend/`: `go test -race -mod=readonly -count=1 ./internal/handler -run 'TestGetAllTournaments|TestSelectTournament|TestGetUserInfo'`, with disposable `PREDITTO_AUTH_TEST_DSN` for database scenarios.

[List tests](../../backend/internal/handler/tournament_test.go) cover combined status/participation filtering, omitted and invalid JSON filters, ignored query parameters, active versus not-paid/blocked membership, ordering, aggregation, nullable fields, selection and identity isolation. [Selection tests](../../backend/internal/handler/select_tournament_test.go) cover invalid IDs, missing profiles/tournaments, nil repository writes, repeated selection, forced write failure and snapshot consistency. Seed changes require the relevant [tournament](../../backend/seeds/test/tournaments_postgres_test.go) and [participant tests](../../backend/seeds/test/tournament_participants_postgres_test.go) with `PREDITTO_SEED_TEST_DSN`; see [Go limitations](../commands/topics/go.md).
