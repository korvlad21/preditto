# Current iteration

2026-11-24

## Single-match predictions

The [save endpoint](../backend/docs/predictions.md#save-a-prediction) creates and updates predictions using the authenticated selection and active participant. The [predictions module](modules/predictions.md) covers the write window, locking and ownership of calculation fields.

The full Go suite passed against disposable PostgreSQL 18.6. The [HTTP integration tests](../backend/internal/handler/save_prediction_test.go) also passed with the race detector in a non-UTC database session. They exercise parallel first saves, calculation and access revocation during a save, and expiration while waiting for a row lock. Swagger generation and route coverage passed.
