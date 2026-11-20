# Prediction stages

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
