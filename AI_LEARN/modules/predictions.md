# Predictions

Prediction saving added on 2026-11-24. Stage and match read APIs reviewed against source on 2026-11-23. Prediction schema reviewed on 2026-11-20.

## Purpose and scope

Owns prediction stage lookups, the stage match list with the current participant's predictions, single-match prediction submission, predicted scores and stored calculation results. See the [API contract](../../backend/docs/predictions.md) and [prediction schema](../../backend/migrations/000015_create_predictions.up.sql).

All [prediction routes](../../backend/internal/router/prediction.go) require authentication, a loaded user and active participation in the selected tournament. The [handler](../../backend/internal/handler/prediction.go) passes the selected tournament ID and active participant ID through the [prediction service](../../backend/internal/service/prediction.go). Missing selection or active participation returns 404.

The [stage repository](../../backend/internal/repository/tournament_stage.go) counts matches and participant predictions for `get_stages` and `get_stage`. Its left joins retain empty stages. The list returns an ordered array and the single-stage lookup returns the same projection as an object.

`get_user_predictions` uses the [prediction repository](../../backend/internal/repository/predictions.go) to fetch a stage's matches, both teams and optional participant predictions in one query. The join on participant and match IDs yields at most one prediction per match. Matches are ordered by scheduled start and ID. An empty stage returns an empty array. Both single-stage routes return 404 for an unknown stage or one outside the selected tournament. The [HTTP integration test](../../backend/internal/handler/prediction_test.go) covers the three routes against PostgreSQL.

## Persistence and constraints

`predictions` has a BIGINT identity primary key and required FKs to tournament_participants.id and matches.id, without cascading deletion. UNIQUE (participant_id, match_id) allows one prediction per participant/match pair. Its index also supports participant lookups and FK checks; the separate match_id index supports match lookups and that FK. No duplicate participant_id index is needed.

Required SMALLINT home_score and away_score must be nonnegative. Nullable VARCHAR(20) result_type is CHECK-constrained to exact_score (exact score), outcome (correct outcome but not exact score) and missed (incorrect prediction). Nullable SMALLINT awarded_points must be nonnegative when present. result_type, awarded_points and calculated_at remain NULL when omitted, representing an uncalculated prediction. The schema permits these nullable fields independently; it does not calculate results or enforce their semantic consistency.

All timestamps use TIMESTAMP(6) without time zone. created_at and updated_at are required and default to CURRENT_TIMESTAMP; there is no automatic updated_at trigger. Both migration directions are transactional. Rollback removes only predictions, including its constraints, indexes and identity sequence.

## Dependencies and tournament consistency

[Tournaments](tournaments.md) owns participants and stages; [matches](matches.md) owns match references. A participant has tournament_participants.tournament_id; a match reaches its tournament through matches.stage_id -> tournament_stages.tournament_id. The two independent FKs in predictions verify existence, not equality of those tournament IDs. Cross-tournament predictions are therefore possible through direct SQL.

The [save endpoint](../../backend/docs/predictions.md#save-a-prediction) checks the active participant's user and tournament, the match's stage and tournament, and the UTC submission window in one SQL statement. A row lock on the participant serializes submissions even for a new prediction. Shared locks keep the match and stage stable while allowing other participants to save. The UPSERT checks the wall-clock deadline again after locking a conflicting prediction and updates its scores only while `calculated_at IS NULL`. Its `RETURNING` result determines success. A later reassignment of a match, stage or participant can change the tournament relationship after this statement commits.

## Verification

Use disposable PostgreSQL for migration up/down/reapply, valid and invalid values, FK/unique/check constraints, NULL defaults and parent-row preservation. Run the [save HTTP integration tests](../../backend/internal/handler/save_prediction_test.go) against disposable PostgreSQL for validation, access, timing, UPSERT and concurrent scoring/submission. See the [database command workflow](../commands/topics/docker.md).

Verified all fifteen migrations up/down/up; result types, NULL/defaults, timestamp precision and index definitions; 17 negative FK/unique/check/NOT NULL cases. Migration 15 rollback removed prediction objects while preserving participants and matches, and reapplied successfully. The proposed conditional insert accepted a same-tournament pair and rejected a cross-tournament pair; direct SQL still accepted a cross-tournament prediction, confirming the documented limitation. These checks used a disposable database, not the application database.
