# Rankings

Reviewed: 2026-10-24; source/schema inspection and disposable PostgreSQL 18.6 verification of migration 16.

## Purpose and scope

Owns stored participant rankings after each tournament stage. The implemented surface is [migration 16 up](../../backend/migrations/000016_create_participant_stage_rankings.up.sql) and [down](../../backend/migrations/000016_create_participant_stage_rankings.down.sql). No ranking model, repository, service, API, seeds, position calculation or points-awarding logic is implemented.

## Persistence and constraints

`participant_stage_rankings` references tournament_participants.id and tournament_stages.id without cascading deletion. The named composite primary key `pk_participant_stage_rankings` (participant_id, stage_id) allows one row per participant/stage. There is no separate id or identity sequence.

All eleven columns are required and have no defaults. INT stage_position and cumulative_position must be positive. INT stage_points, cumulative_points, stage_exact_scores, cumulative_exact_scores, stage_outcomes and cumulative_outcomes must be nonnegative. calculated_at is TIMESTAMP(6) without time zone and must be explicitly supplied as the latest calculation time.

Stage fields store results for that stage; cumulative fields describe the tournament through that stage, inclusive. Exact-score counts represent exact predictions; outcome counts exclude exact scores. These meanings are not calculated or cross-validated by the migration. It does not require cumulative values to exceed stage values or enforce uniqueness of positions, so tied positions are permitted.

The two nonunique B-tree indexes are `idx_rankings_stage_position` (stage_id, stage_position) and `idx_rankings_cumulative_position` (stage_id, cumulative_position). They support the stage and cumulative leaderboards respectively. Their stage_id prefix and the primary key's participant_id prefix also support FK lookups; no standalone duplicate indexes are needed.

Both migration directions are transactional. Rollback drops only participant_stage_rankings, automatically removing its indexes and constraints.

## Dependencies and tournament consistency

[Tournaments](tournaments.md) owns participants and stages. Independent FKs do not guarantee that participant.tournament_id equals stage.tournament_id; direct SQL can store cross-tournament rows.

For future persistence, validate both IDs in the write itself using INSERT ... SELECT with a join on tournament_id, and apply the same validation when updating existing rows. This is a proposed write pattern, not implemented ranking behavior. Missing or mismatched parents must produce no write and be handled by the caller. As with the [prediction write proposal](predictions.md#dependencies-and-tournament-consistency), it validates the statement's snapshot only. Reassignment of parent tournament IDs must also be prevented or coordinated while dependent rows exist. A full database invariant requires a broader composite-FK or trigger design covering parent changes, outside this migration's scope.

## Verification

Use disposable PostgreSQL to check the named primary key, both index definitions, all positive/nonnegative constraints, required explicit values, timestamp precision, duplicate pairs, tied positions, FKs, rollback and reapplication. Test the proposed tournament join separately from the direct-SQL limitation. See the [database command workflow](../commands/topics/docker.md).

Verified all sixteen migrations up/down/up; exact PK/index definitions, types, no defaults, zero totals, tied positions and timestamp precision; 38 negative FK/PK/NOT NULL/CHECK cases. Migration 16 rollback preserved participant and stage rows, and reapplication succeeded. A conditional insert using the proposed tournament join rejected a mismatched pair and accepted a matching pair; direct SQL accepted a mismatched pair, confirming the limitation. All runtime checks used a disposable database, not the application database.
