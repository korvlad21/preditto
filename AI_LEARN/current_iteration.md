# Current iteration

Date: 2026-11-08

## Outcome

Implemented `PUT /api/user/update_user` using RequireAuth then LoadUser and
owner/admin authorization from CurrentUser. Account email and supplied profile
fields update atomically, including creation of a missing profile. Uniqueness
and favorite-team errors use the shared response contract. Other account fields
and tournament selection are preserved; successful self-updates refresh the
shared request snapshot.

The same endpoint accepts JSON or multipart `data` JSON plus an `avatar` file.
AvatarStore validates bounded JPEG/PNG content and re-encodes JPEG with a UUID v4
filename under `frontend/images/users`. SQL stores only `/users/<uuid>.jpg`.
Known database errors remove new files; uncertain commits and replaced files
retain them to avoid deleting a referenced image. The Compose upload mount and
Git ignore rule are included. No schema migration or dependency change is needed.

The [user contract](../backend/docs/user.md#update-a-user) owns request/null
semantics, multipart examples, errors, storage consistency and deployment.
The [users module](modules/users.md) links implementation and tests. Swagger
annotations and generated documentation include the new route; existing Swagger
setup and maintenance findings remain in the [Swagger guide](../backend/docs/swagger.md).

## Verification

- All `internal/...` packages and `cmd/api` passed with `-race -mod=readonly`
  against disposable PostgreSQL, including existing auth/role/selection tests.
- New update tests cover access, validation, profile creation, rollback,
  snapshots, concurrent email conflicts, actual JPEG conversion and upload failures.
- `go build -mod=readonly ./...`, `go vet -mod=readonly ./...`, `make swagger`,
  `make swagger-check` and `docker compose config --quiet` passed.
- Seed tests were not rerun; no seed or schema changes were made. Process crashes
  and ambiguous COMMIT failures were not simulated. No application database was used.

## Next steps

Persist and serve the upload directory at `/users/` in production. Recreate the
backend container to adopt its new Compose mount when deploying. The frontend's
profile-edit API adapter remains a separate integration task. Unreferenced-image
reconciliation is outside this endpoint; see the contract's consistency limits.
