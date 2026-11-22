# Matches

Reviewed on 2026-11-19 against the development seed and disposable PostgreSQL 18.6.

## Purpose and scope

Owns match persistence and development fixtures, including stage and team references, scores, scheduling, prediction locks and status. The [schema](../../backend/migrations/000014_create_matches.up.sql), [rollback](../../backend/migrations/000014_create_matches.down.sql) and [match seed](../../backend/seeds/development/matches.go) implement this module.

## Persistence and constraints

`matches` has a BIGINT identity primary key and required references to `tournament_stages.id` and `teams.id`, without cascading deletion. Home and away teams must differ. Nullable SMALLINT scores accept zero and positive values; NULL represents an unknown score. Each score can be NULL independently.

Required VARCHAR(20) status defaults to scheduled when omitted and is CHECK-constrained to scheduled/live/finished/postponed/cancelled. Explicit NULL remains invalid. The schema does not enforce status transitions or require scores/timestamps for particular statuses.

All timestamps use TIMESTAMP(6) without time zone. scheduled_start_at and prediction_lock_at are required; actual_start_at and finished_at are nullable. created_at and updated_at default to CURRENT_TIMESTAMP; there is no automatic updated_at trigger or enforced ordering between match timestamps.

Composite B-tree indexes pair stage_id, home_team_id, away_team_id and status individually with scheduled_start_at. They support filtered schedules and date ranges; the leading FK columns also support reference checks. A separate scheduled_start_at index supports the global schedule without requiring a stage, team or status filter. The home/away indexes allow either side to participate in team-match searches. There are no redundant standalone indexes on their leading columns.

Both migration directions are transactional. Rollback drops only matches; its indexes, constraints and identity sequence are removed with the table.

## Dependencies and limitations

[Tournaments](tournaments.md) owns stages; [teams](teams.md) owns the shared team directory. Both `home_team_id` and `away_team_id` reference teams directly. No tournament_teams table exists, and these FKs do not establish membership in the stage's tournament.

[Predictions](predictions.md) references matches without cascading deletion; a referenced match cannot be deleted until its predictions are removed. Match storage does not implement prediction submission, scoring rules or tournament roster management.

## Development seed

The development runner seeds the 2026/27 Champions League after the stages. It resolves each fixture by tournament slug, stage code and the home and away team slugs. The match seed locks the table for writes while checking those references and merging fixtures. A missing reference or duplicate fixture aborts the shared transaction. Existing matches keep their IDs and creation times. Identical rows keep their update times. A scheduled snapshot entry preserves a match already marked live, finished, postponed or cancelled. The [seed guide](../../backend/seeds/README.md#league-phase-matches) records the source data, timestamp rules and snapshot date.

## Verification

The [PostgreSQL seed tests](../../backend/seeds/test/matches_postgres_test.go) check the eight matchdays, identity on rerun, other-tournament isolation, NULL fields, UTC wall times in a non-UTC session and rollback on a missing club.

Use disposable PostgreSQL for migration application, constraint checks, rollback and reapplication; see the [database command workflow](../commands/topics/docker.md). Verify invalid FKs, same-team matches, negative scores, invalid/null statuses and required timestamps are rejected; unknown scores and actual times must remain NULL.

Verified all fourteen migrations up/down/up, all five match statuses, 19 negative constraint cases, nullable scores/times, defaults, timestamp precision and index definitions. Migration 14 rollback removed the table, indexes and identity sequence while preserving tournament, stage and team rows; reapplication succeeded. These checks used a disposable database, not the application database.
