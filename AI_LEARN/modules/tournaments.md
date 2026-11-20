# Tournaments

Reviewed: 2026-10-22, source/schema inspection and targeted tournament/user-info HTTP tests with race detection against disposable PostgreSQL; affected-package/API vet passed.

## Purpose and API

Owns tournament listing, user selection coordination, participant/prize projections and stage persistence. Read this for filters, selection, counts, prizes, stages or tournament seeds.

| Protected POST route | Successful result |
| --- | --- |
| `/api/tournaments/get_all_tournaments` | 200, data.tournaments; optional status and is_participate JSON filters, no required body |
| `/api/tournaments/select_tournament` | 200, success without data; JSON id required |

Both routes require RequireAuth followed by LoadUser. The [tournament API contract](../../backend/docs/tournaments.md) owns exact fields, error codes and examples, using the [shared response contract](../../backend/docs/responses.md).

Tournament handler Swagger annotations reuse the tournament request DTOs and Tournament projection. See the [Swagger guide](../../backend/docs/swagger.md) for generation and shared envelope ownership.

## Execution and business rules

Listing passes CurrentUser.ID and CurrentUser.SelectedTournamentID from handler to service. TournamentService normalizes omitted/empty status or lowercase all to no status filter and marks selected rows without another user read. Other status values match exactly; unknown values return an empty list. TournamentRepository uses one parameterized query, ordered by created_at DESC then id DESC.

The optional is_participate boolean combines with the tournament-status filter using AND. True requires an active tournament_participants row for the current user; false requires its absence; omission imposes no participation restriction. Not-paid and blocked rows are not participation. The repository uses correlated EXISTS/NOT EXISTS in the listing query, with no per-tournament reads and no changes to counts or selection semantics. Filters are read only from the JSON body; query parameters are ignored. An empty body or `{}` applies no filters. Request fields cannot override the authenticated user ID.

Prize joins for places 1/2/3 each match at most one row by the composite key. Active participant counts are grouped before joining and count distinct user IDs, preventing multiplied rows. The query filters participation status, not account status. Missing prize rules stay null and absent active participants yield zero. A null selection, including a missing profile, makes all selection flags false.

Selection accepts only a positive non-null int64 JSON ID and uses the authenticated user's ID. The service checks tournament existence, then UserRepository updates the existing profile. No profile is created implicitly. Missing profiles and tournaments have different public errors. A foreign-key violation maps deletion between the existence check and write to TournamentNotFound. Successful persistence precedes updating the shared CurrentUser pointer.

Selection is not membership or payment and imposes no tournament-status restriction. The repository accepts nil to clear selection, but the HTTP API does not expose reset. Dates do not automatically change tournament status.

The optional [LoadParticipant middleware](../../backend/internal/middleware/tournament_participant.go) runs after RequireAuth and LoadUser. It reads the user ID and selected tournament ID from CurrentUser. The [participant service](../../backend/internal/service/tournament_participant.go) and [repository](../../backend/internal/repository/tournament_participant.go) look up that pair once and require active participation. CurrentParticipant returns the shared request snapshot containing only the participant ID and joined_at. The [prediction stage endpoint](../../backend/docs/predictions.md) uses this middleware.

LoadParticipant returns 401 for a missing user snapshot, 404 TOURNAMENT_NOT_SELECTED for a nil selection, 404 TOURNAMENT_PARTICIPANT_NOT_FOUND when no active participant row matches, and 500 INTERNAL_ERROR for a lookup failure.

## Sources, models and persistence

| Sources | Responsibility |
| --- | --- |
| [router](../../backend/internal/router/tournament.go), [handler](../../backend/internal/handler/tournament.go), [request DTOs](../../backend/internal/dto/tournaments/request.go) | Protected endpoints, input binding, HTTP mapping and snapshot update |
| [service](../../backend/internal/service/tournament.go), [repository](../../backend/internal/repository/tournament.go) | Selection orchestration, existence checks, listing and aggregates |
| [UserRepository](../../backend/internal/repository/user.go), [user middleware](../../backend/internal/middleware/user.go) | Profile selection persistence and shared request snapshot |
| [Tournament model](../../backend/internal/model/tournament.go), [participant model](../../backend/internal/model/tournament_participant.go) | Tournament projection, currencies, types and statuses, participation statuses and ParticipantInfo |

| Table | Important constraints |
| --- | --- |
| [tournaments, migration 10](../../backend/migrations/000010_create_tournaments.up.sql) | Unique slug; nullable status, dates, cost and currency; allowed non-null statuses are planned/active/finished/cancelled |
| [tournament_participants, migration 11](../../backend/migrations/000011_create_tournament_participants.up.sql) | Unique tournament/user pair; required active/not_paid/blocked status; indexes by user and tournament/status |
| [tournament_prize_rules, migration 12](../../backend/migrations/000012_create_tournament_prize_rules.up.sql) | Primary key tournament/place; each percentage between 0 and 100 |
| [tournament_stages, migration 13](../../backend/migrations/000013_create_tournament_stages.up.sql) | Required tournament FK without cascading deletion; unique tournament/code; positive sequence_no; nonnegative outcome/exact-score points; nullable prediction opening timestamp; index by tournament/sequence |
| [user_info, migration 4](../../backend/migrations/000004_create_user_info.up.sql) | Nullable selected_tournament_id; migration 10 adds its FK with ON DELETE SET NULL |

Participant and prize-rule references do not declare cascading deletion. The selection FK alone does not make tournament deletion unrestricted. Unlike users/user_info, tournaments has no automatic updated_at trigger; the seed upsert sets it explicitly. The schema does not require a positive prize place, a total percentage of 95/100, ordered dates or nonnegative cost; do not claim these constraints exist.

Migration 13 stores stage status and stage_kind as required VARCHAR(20) fields with CHECK constraints: active/planned/finished/cancelled and group/round/play-off respectively. It creates no PostgreSQL enum types. Stage timestamps use TIMESTAMP(6); created_at/updated_at default to CURRENT_TIMESTAMP without an update trigger. Stage sequence numbers need not be unique within a tournament. Rollback drops only the stage table, without CASCADE. The [stage projection](../../backend/internal/model/tournament_stage.go) includes match and prediction counts for the prediction stage list.

## Seeds, dependencies and limits

The [stage seed](../../backend/seeds/development/tournament_stages.go) adds 17 stages only for `uefa-champions-league-2026-2027`, resolving its ID by slug inside the existing development transaction. Upserts use `(tournament_id, code)`, preserve IDs/created_at and restore all configured fields. Sequence 1–8 is `group`; 9–17 is `play-off`. Outcome/exact-score points progress through 1/3, 2/6, 3/9, 4/12, 5/15 and 6/18. Stage timestamps written by this seed use UTC even with a non-UTC database session.

The [stage calendar and sources](../../backend/seeds/README.md#tournament-stages) own dates, codes and outstanding scheduling assumptions, checked 2026-11-17. Finishes are estimated as last kickoff + 3 hours; subsequent prediction windows open 8 hours after the previous stage’s last kickoff. The first window opens 2026-08-20 00:00 UTC. Status is recomputed once per seed execution: finished at/after estimated end, first unfinished stage active, subsequent stages planned. There is no automatic status refresh between runs. Knockout kickoff assignments and actual final-whistle times require confirmation; the seed uses documented regulation slots and estimates, not live match state. Migration 13 omits the former stage closing timestamp and its check. The seed stores only the last kickoff per stage, which drives opening times and estimated completion statuses. Prediction locking belongs to matches.prediction_lock_at; submission enforcement is not implemented.

[Matches](matches.md) references tournament_stages without cascading deletion. A stage with matches cannot be deleted until those references are removed.

[Rankings](rankings.md) references participants and stages without cascading deletion. Referenced parents cannot be deleted until their ranking rows are removed; the rankings module documents the tournament-consistency limitation.

The [tournament](../../backend/seeds/development/tournaments.go), [participant](../../backend/seeds/development/tournament_participants.go) and [prize](../../backend/seeds/development/tournament_prize_rules.go) seeds share the development transaction. Tournaments upsert by slug; participants resolve tournament IDs by slug and user IDs by username. Membership is explicit, independent of user slice order, and unique within each group. Reruns preserve identities/joined_at while restoring configured participation status. Account and participation statuses are independent.

The first tournament (`uefa-champions-league-2026-2027`) includes `korvlad21` as an active participant. The [seed guide](../../backend/seeds/README.md) owns the three configured tournament datasets, dates, costs and counts. Each currently allocates 95% in prize rules; this is dataset configuration, not an enforced general rule. SQL stores DECIMAL, seeds pass decimal strings, and the response model reads cost/percentages into *float64. Exact monetary calculations are not implemented.

[Users](users.md) owns the selection column and snapshot lifecycle; [auth](auth.md) establishes identity. [Predictions](predictions.md) references participants without cascading deletion and documents the current tournament-consistency limitation. The frontend `/tournaments` page lists tournaments and selects the current one through both routes; see the [frontend README](../../frontend/README.md#tournament-selection). Join/leave, payment, prediction submission, scoring and payouts are not implemented.

## Verification

The [middleware tests](../../backend/internal/middleware/tournament_participant_test.go) cover participant lookup, error responses and the two-field snapshot. TestLoadParticipantPostgres in the [tournament tests](../../backend/internal/handler/tournament_test.go) checks isolation by both IDs, successful active lookup, and 404 for blocked and unpaid participants.

Baseline simplification verified on 2026-11-17: first-kickoff data was removed from the seed, development unit tests and vet passed in a separate container, and the seed ran successfully after PostgreSQL restart and application-service recreation. The already reconciled local schema was retained and migration metadata aligned from 17 to 16 without rolling back tables. All 17 stage rows and their remaining fields (excluding updated_at) matched the pre-change checksum; the obsolete column stayed absent.

The stage baseline is migration 13 without a stage closing timestamp; the migration set ends at version 16. Editing an applied migration does not change existing schemas. For older installations, inspect and reconcile the obsolete column/check before using this revised baseline; do not reset a database merely to replay migration 13.

Stage seed verified on 2026-11-17: all `./seeds/...` tests passed with disposable PostgreSQL 18.6, including integration cases without skips; `go vet -mod=readonly ./seeds/... ./cmd/seed` passed. No application database was seeded in this verification.

Stage [unit tests](../../backend/seeds/development/tournament_stages_test.go) check all 17 stages, points, UTC opening times, seasonal offsets and before/at/after completion statuses. Stage [PostgreSQL tests](../../backend/seeds/test/tournament_stages_postgres_test.go) check idempotence, restored fields, preserved identities, unrelated data, non-UTC session behavior and full transaction rollback. Run `go test -mod=readonly -count=1 ./seeds/...` with `PREDITTO_SEED_TEST_DSN` pointing to disposable PostgreSQL, plus `go vet -mod=readonly ./seeds/... ./cmd/seed`.

Stage schema verified on 2026-10-24 in disposable PostgreSQL 18.6: all thirteen migration pairs applied, rolled back and reapplied; stage columns/defaults, allowed status/kind values, indexes, FK, unique code and check constraints exercised. Migration 13 rollback preserved tournament data.

From `backend/`: `go test -race -mod=readonly -count=1 ./internal/handler -run 'TestGetAllTournaments|TestSelectTournament|TestGetUserInfo'`, with disposable `PREDITTO_AUTH_TEST_DSN` for database scenarios.

[List tests](../../backend/internal/handler/tournament_test.go) cover combined status/participation filtering, omitted and invalid JSON filters, ignored query parameters, active versus not-paid/blocked membership, ordering, aggregation, nullable fields, selection and identity isolation. [Selection tests](../../backend/internal/handler/select_tournament_test.go) cover invalid IDs, missing profiles/tournaments, nil repository writes, repeated selection, forced write failure and snapshot consistency. Seed changes require the relevant [tournament](../../backend/seeds/test/tournaments_postgres_test.go) and [participant tests](../../backend/seeds/test/tournament_participants_postgres_test.go) with `PREDITTO_SEED_TEST_DSN`; see [Go limitations](../commands/topics/go.md).
