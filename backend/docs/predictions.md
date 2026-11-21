# Prediction stages

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
