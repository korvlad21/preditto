# Current Iteration

Date: 2026-09-27

Moved all four authentication route declarations and their shared request limit and cache headers from `handler.AuthHandler.RegisterRoutes` to [internal/router/auth.go](../../backend/internal/router/auth.go). `cmd/api/main.go` and the HTTP integration test setup now call `router.RegisterAuthRoutes`; handler methods remain responsible for binding and responses. No URL, HTTP method, or auth flow changed. The [API contract](../../backend/docs/auth.md) describes the ownership boundary.

The router test verifies the exact route list and common headers. Targeted router/handler/API package tests and vet passed. Existing auth HTTP integration tests passed against PostgreSQL using disposable schemas. The previous authentication implementation iteration is preserved in [archive](2026-09-19-auth-api-iteration.md).
