# Matches

Reviewed: 2026-10-24; source/schema inspection and disposable PostgreSQL 18.6 verification of migration 14.

## Purpose and scope

Owns match persistence: stage/team references, scores, scheduling, prediction lock time and status. The implemented surface is [migration 14 up](../../backend/migrations/000014_create_matches.up.sql) and [down](../../backend/migrations/000014_create_matches.down.sql). No match model, repository, service, API or seeds are implemented.

## Persistence and constraints

`matches` has a BIGINT identity primary key and required references to `tournament_stages.id` and `teams.id`, without cascading deletion. Home and away teams must differ. Nullable SMALLINT scores accept zero and positive values; NULL represents an unknown score. Each score can be NULL independently.

Required VARCHAR(20) status defaults to scheduled when omitted and is CHECK-constrained to scheduled/live/finished/postponed/cancelled. Explicit NULL remains invalid. The schema does not enforce status transitions or require scores/timestamps for particular statuses.

All timestamps use TIMESTAMP(6) without time zone. scheduled_start_at and prediction_lock_at are required; actual_start_at and finished_at are nullable. created_at and updated_at default to CURRENT_TIMESTAMP; there is no automatic updated_at trigger or enforced ordering between match timestamps.

Composite B-tree indexes pair stage_id, home_tournament_team_id, away_tournament_team_id and status individually with scheduled_start_at. They support filtered schedules and date ranges; the leading FK columns also support reference checks. A separate scheduled_start_at index supports the global schedule without requiring a stage, team or status filter. The home/away indexes allow either side to participate in team-match searches. There are no redundant standalone indexes on their leading columns.

Both migration directions are transactional. Rollback drops only matches; its indexes, constraints and identity sequence are removed with the table.

## Dependencies and limitations

[Tournaments](tournaments.md) owns stages; [teams](teams.md) owns the shared team directory. Despite their names, home_tournament_team_id and away_tournament_team_id reference teams directly. No tournament_teams table exists, and these FKs do not establish membership in the stage's tournament. The column names intentionally retain the supplied schema contract.

Match storage does not implement predictions, scoring rules or tournament roster management.

## Verification

Use disposable PostgreSQL for migration application, constraint checks, rollback and reapplication; see the [database command workflow](../commands/topics/docker.md). Verify invalid FKs, same-team matches, negative scores, invalid/null statuses and required timestamps are rejected; unknown scores and actual times must remain NULL.

Verified all fourteen migrations up/down/up, all five match statuses, 19 negative constraint cases, nullable scores/times, defaults, timestamp precision and index definitions. Migration 14 rollback removed the table, indexes and identity sequence while preserving tournament, stage and team rows; reapplication succeeded. These checks used a disposable database, not the application database.
