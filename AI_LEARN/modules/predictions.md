# Predictions

Reviewed: 2026-10-24; source/schema inspection and disposable PostgreSQL 18.6 verification of migration 15.

## Purpose and scope

Owns predicted match scores and stored calculation results. The implemented surface is [migration 15 up](../../backend/migrations/000015_create_predictions.up.sql) and [down](../../backend/migrations/000015_create_predictions.down.sql). No prediction model, repository, service, API, seeds or scoring process are implemented.

## Persistence and constraints

`predictions` has a BIGINT identity primary key and required FKs to tournament_participants.id and matches.id, without cascading deletion. UNIQUE (participant_id, match_id) allows one prediction per participant/match pair. Its index also supports participant lookups and FK checks; the separate match_id index supports match lookups and that FK. No duplicate participant_id index is needed.

Required SMALLINT home_score and away_score must be nonnegative. Nullable VARCHAR(20) result_type is CHECK-constrained to exact_score (exact score), outcome (correct outcome but not exact score) and missed (incorrect prediction). Nullable SMALLINT awarded_points must be nonnegative when present. result_type, awarded_points and calculated_at remain NULL when omitted, representing an uncalculated prediction. The schema permits these nullable fields independently; it does not calculate results or enforce their semantic consistency.

All timestamps use TIMESTAMP(6) without time zone. created_at and updated_at are required and default to CURRENT_TIMESTAMP; there is no automatic updated_at trigger. Both migration directions are transactional. Rollback removes only predictions, including its constraints, indexes and identity sequence.

## Dependencies and tournament consistency

[Tournaments](tournaments.md) owns participants and stages; [matches](matches.md) owns match references. A participant has tournament_participants.tournament_id; a match reaches its tournament through matches.stage_id -> tournament_stages.tournament_id. The two independent FKs in predictions verify existence, not equality of those tournament IDs. Cross-tournament predictions are therefore possible through direct SQL.

For a future repository write, the minimal proposed validation is an atomic conditional insert (not implemented application behavior):

```sql
INSERT INTO predictions (participant_id, match_id, home_score, away_score)
SELECT p.id, m.id, $3, $4
FROM tournament_participants p
JOIN matches m ON m.id = $2
JOIN tournament_stages s ON s.id = m.stage_id
WHERE p.id = $1
  AND p.tournament_id = s.tournament_id
RETURNING id;
```

No returned row means the participant/match was absent or their tournaments differed. This validates the statement's snapshot; it is not a database-wide invariant. Future workflows must also prevent or coordinate reassignment of participant.tournament_id, match.stage_id and stage.tournament_id while predictions reference them. A complete database-enforced rule would require a broader schema or trigger design covering those parent updates, beyond this migration.

The migration does not enforce participant status, prediction deadlines, match status or scoring rules.

## Verification

Use disposable PostgreSQL for migration up/down/reapply, valid and invalid values, FK/unique/check constraints, NULL defaults and parent-row preservation. Test the proposed conditional insert for both matching and different tournaments, and distinguish that query's behavior from the weaker direct-SQL guarantees. See the [database command workflow](../commands/topics/docker.md).

Verified all fifteen migrations up/down/up; result types, NULL/defaults, timestamp precision and index definitions; 17 negative FK/unique/check/NOT NULL cases. Migration 15 rollback removed prediction objects while preserving participants and matches, and reapplied successfully. The proposed conditional insert accepted a same-tournament pair and rejected a cross-tournament pair; direct SQL still accepted a cross-tournament prediction, confirming the documented limitation. These checks used a disposable database, not the application database.
