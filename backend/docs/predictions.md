# Predictions

## List stages

`POST /api/predictions/get_stages` requires a Bearer access token and takes no path, query or body parameters. The API reads the tournament from the user's profile and counts predictions for the user's active participant record in that tournament.

HTTP 200 returns all stages, ordered by ascending sequence number and then stage ID. An empty tournament returns `{"success":true,"data":[]}`. Each stage includes the total number of its matches and the number of those matches with a prediction from this participant. Both counts include all match statuses. A stage with no matches has two zero counts. The opening timestamp can be null.

```json
{
  "success": true,
  "data": [{
    "id": 1,
    "tournament_id": 1,
    "code": "league_phase_1",
    "name": "Matchday 1",
    "stage_kind": "group",
    "sequence_no": 1,
    "status": "finished",
    "predictions_open_at": "2026-09-01T00:00:00Z",
    "outcome_points": 1,
    "exact_score_points": 3,
    "count_matches": 18,
    "count_predictions": 15
  }]
}
```

| HTTP | Error code | Condition |
| --- | --- | --- |
| 401 | UNAUTHORIZED | Missing or invalid access token |
| 404 | USER_NOT_FOUND | The authenticated user's account is missing |
| 404 | TOURNAMENT_NOT_SELECTED | The user's profile has no tournament selection |
| 404 | TOURNAMENT_PARTICIPANT_NOT_FOUND | No active participant record for this user and tournament |
| 500 | INTERNAL_ERROR | Internal lookup or processing failure |

## Get one stage

`POST /api/predictions/get_stage` requires a Bearer access token and a JSON body such as `{"id":1}`. The `id` field is a required positive signed 64-bit integer. The API uses the user's selected tournament and active participant record in that tournament. Query parameters are ignored.

HTTP 200 returns the stage object directly in `data`. It has the same fields and count semantics as an item in the list. A stage without matches has zero matches and zero predictions. Responses that reach the user loader include `Cache-Control: no-store`.

| HTTP | Error code | Condition |
| --- | --- | --- |
| 400 | INVALID_REQUEST | Invalid JSON body or missing, null, nonpositive or invalid `id` |
| 401 | UNAUTHORIZED | Missing or invalid access token |
| 404 | TOURNAMENT_PARTICIPANT_NOT_FOUND | No active participant record for this user and tournament |
| 404 | USER_NOT_FOUND | The authenticated user's account is missing |
| 404 | TOURNAMENT_NOT_SELECTED | The user's profile has no tournament selection |
| 404 | TOURNAMENT_STAGE_NOT_FOUND | No stage with this ID in the selected tournament |
| 500 | INTERNAL_ERROR | Internal lookup or processing failure |

## Get matches with my predictions

`POST /api/predictions/get_user_predictions` requires a Bearer access token and a JSON body such as `{"id":1}`. The required positive int64 `id` identifies a stage in the selected tournament. The API uses the active participant from the authenticated request. Client-supplied tournament and participant IDs are ignored.

On HTTP 200, `data` contains every match of the stage, ordered by `scheduled_start_at` ascending and then by match `id` ascending. An existing stage without matches returns `{"success":true,"data":[]}`.

| Field in each item | Contents |
| --- | --- |
| `match` | `id`, nullable `home_score` and `away_score`, `scheduled_start_at`, `prediction_lock_at`, nullable `actual_start_at` and `finished_at`, and `status`. Timestamps use RFC 3339. |
| `home_team`, `away_team` | Each contains `id`, `name`, nullable `short_name` and nullable `logo_url`. |
| `prediction` | `null` if this participant has no prediction for the match. Otherwise contains `id`, `home_score`, `away_score`, nullable `result_type` and nullable `awarded_points`. |

Validation, error codes and cache headers follow [Get one stage](#get-one-stage). Both an unknown stage and a stage from another tournament return HTTP 404 with `TOURNAMENT_STAGE_NOT_FOUND`.

## Save a prediction

`POST /api/predictions/save_prediction` requires a Bearer access token. Send a JSON object with all three integer fields.

```json
{"match_id":1,"home_score":2,"away_score":0}
```

`match_id` is a positive signed 64-bit integer. Both scores accept 0 through 32767 inclusive. Missing or null fields are invalid. The API uses the authenticated user's selected tournament and active participant record. It derives the match's tournament from its stage.

Submission opens at the stage's `predictions_open_at` (inclusive) and closes at the match's `prediction_lock_at` (exclusive), using the database clock and UTC fixture times. A null opening time keeps the stage closed. Live and finished matches are closed. Scheduled, postponed and cancelled matches use the same time window. A calculated prediction is final.

HTTP 200 returns `{"success":true}` after creating or updating the prediction. Repeated saves replace both scores and refresh the update timestamp. Result calculation owns `result_type`, `awarded_points` and `calculated_at`.

| HTTP | Error code | Condition |
| --- | --- | --- |
| 400 | INVALID_REQUEST | Invalid request body, field type, presence or range |
| 401 | UNAUTHORIZED | Missing or invalid access token |
| 404 | USER_NOT_FOUND / TOURNAMENT_NOT_SELECTED | Missing user or tournament selection |
| 404 | TOURNAMENT_PARTICIPANT_NOT_FOUND | Active participation is missing or has been revoked |
| 404 | MATCH_NOT_FOUND | Match is missing or belongs to another tournament |
| 409 | PREDICTIONS_NOT_OPEN | Stage opening time is missing or in the future |
| 409 | PREDICTION_LOCKED | Match is live or finished, or its deadline has arrived |
| 409 | PREDICTION_ALREADY_CALCULATED | The prediction has a calculation timestamp |
| 500 | INTERNAL_ERROR | Internal lookup or save failure |
