# Current Iteration

Date: 2026-09-19

Implemented the [authentication API](../../backend/docs/auth.md) using existing Gin routing, database/sql, lib/pq, config, role repository, and schema. Added DTO validation, bcrypt registration/login, distinct access/refresh JWT signing, database-backed hashed sessions, atomic one-use refresh rotation, logout, and access middleware. Registration includes user, profile, participant role, and first session in one transaction using returned user IDs. Migration 000009 adds refresh_tokens without rewriting prior migrations.

All requested scenarios are covered. `go test -mod=readonly -race -count=1 ./...` passed with both auth and seed PostgreSQL integration variables set against a disposable database; `go vet -mod=readonly ./...` passed. Integration checks also verified nonsequential IDs, rollback at four registration stages, rotation rollback, eight simultaneous refresh requests with exactly one success, and migration 9 down/up. Final token parsing and lock-order adjustments were checked with affected package race tests and vet.

API setup and examples are documented in backend/docs/auth.md. Startup requires JWT env/config; the database-only seed command remains independent of JWT. Local environment activation and the Compose no-deps startup failure are recorded in AI_USER_LEARN, without secrets in project knowledge.

Final strict UTF-8 decoding, Markdown link checks, gofmt, Compose configuration validation, and git diff checks passed.

The previous iteration is preserved in [archive](2026-09-19-team-logos-iteration.md).
