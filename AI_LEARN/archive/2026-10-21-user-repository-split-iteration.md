# Current Iteration

Date: 2026-10-21

[UserRepository](../../backend/internal/repository/user.go) now owns CreateUser, FindUserByLogin, GetUserInfo, LockUser, UpdateSelectedTournament, and AssignRole (in user_roles.go). UserService and TournamentService depend on UserRepository; AuthService receives it alongside AuthRepository. The API entry point shares a UserRepository backed by the same database pool. AuthRepository retains refresh-session persistence and InTransaction, whose callback receives auth and user repositories backed by one sql.Tx. Registration remains atomic across user, profile, role assignment, and first session; login and refresh retain user locking and transactional session writes. SQL, API contracts, schemas, and dependencies are unchanged. API documentation and backend knowledge reflect the split.

Verification: affected repository/service/handler/middleware/router/API packages passed `go test -race -mod=readonly -count=1` and `go vet -mod=readonly`. The complete handler suite also passed with race detection against disposable PostgreSQL, including registration rollback at four stages, login, refresh/logout, rotation rollback and concurrency, user profiles, and tournament selection. Formatting, diff checks, strict UTF-8, and documentation links passed. No unexpected command failures occurred.

Next step: keep transactional user operations on the callback-scoped UserRepository; future profile operations belong to UserRepository.

Previous iteration: [omit absent success data](2026-10-21-omit-empty-success-data-iteration.md).
