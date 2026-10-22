# Coding conventions

Scope, context routing and documentation maintenance are defined once in [AGENTS.md](../../AGENTS.md). Current implementation boundaries are in [architecture](../../AI_LEARN/architecture.md).

## Backend

- Keep HTTP binding and response/error mapping in handlers, use-case orchestration in services, and parameterized SQL in repositories. Preserve `context.Context` propagation.
- Use existing composition in `cmd/api/main.go`; add abstractions only for a concrete need. The project mainly uses concrete dependencies, with small interfaces at selected consumers.
- User/profile/role persistence belongs to UserRepository; refresh-session persistence belongs to AuthRepository. Within auth transactions, use the callback-scoped repositories so all writes share the same transaction.
- Protected endpoints needing CurrentUser must run RequireAuth before LoadUser. Reuse the request snapshot and update it only after a successful write.
- Follow the [response contract](../../backend/docs/responses.md): shared envelopes/error constants, omitted nil success data, empty list arrays, and preserved nullable fields. Do not expose database errors or credentials.
- Reuse domain constants from the relevant model files. Check actual SQL constraints before treating a constant as enforced policy.
- Preserve migration up/down symmetry and dependency order. Use a new migration for changes to an applied/shared schema; do not assume editing old SQL changes existing databases.
- Register seed steps in dependency order in the existing transaction. Preserve repeat-run identity and rollback behavior; do not confuse seed examples with business constraints.

## Frontend and validation

- Follow the existing React/TypeScript, Vite and ESLint configuration. Inspect current UI and scripts before proposing new client, routing or state-management conventions.
- Static team logos use the configured public directory; follow [asset provenance and paths](../../docs/assets/team-logos.md).
- Test changed behavior and relevant negative cases, not incidental implementation details. Use existing test fixtures and disposable databases; follow the [command index](../../AI_LEARN/commands/index.md).
- Keep dependency changes, formatting and generated output within task scope. Documentation-only work does not require starting services or rerunning application suites.
