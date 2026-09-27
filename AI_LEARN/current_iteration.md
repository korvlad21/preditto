# Current Iteration

Date: 2026-09-27

Moved user/profile creation, user lookup and row locking, and user write-error mapping from `repository/auth.go` to [repository/user.go](../backend/internal/repository/user.go). The role lookup/assignment wrapper now lives in [repository/user_roles.go](../backend/internal/repository/user_roles.go); `repository/auth.go` retains the auth transaction wrapper and refresh-session persistence. The exported repository API, SQL, transaction scope, and HTTP behavior did not change. The [API contract](../backend/docs/auth.md) reflects the file ownership.

Targeted repository and handler package tests, service/handler/repository vet, and the auth HTTP integration tests against PostgreSQL all passed. The tests used disposable schemas and covered registration rollback, login, refresh rotation, and logout. The previous route extraction iteration is preserved in [archive](archive/2026-09-27-auth-router-iteration.md).
