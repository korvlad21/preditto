# Current Iteration

Date: 2026-09-13

## Outcome

Moved README.md to backend/seeds, seed.go and teams.go to backend/seeds/development, and teams_test.go to backend/seeds/test. The CLI imports preditto/seeds/development and calls development.Run. Private seedTeams, validation, dataset, SQL, and transactional behavior are unchanged. Migration documentation points to the new README; Compose still runs ./cmd/seed after migrations.

Private validation tests are preserved in development/teams_validation_test.go. The separate test package exercises the existing public Run using a local SQL connector, verifying 36 valid unique teams, commit, wrapped write errors, and rollback. No production API or dependencies were added.

See [seed documentation](../backend/seeds/README.md) and [backend knowledge](modules/backend/overview.md). The preceding iteration is preserved in [archive](archive/2026-09-12-teams-seeds.md).

## Verification

- Targeted tests, go build -mod=readonly ./..., go test -mod=readonly ./..., go vet -mod=readonly ./..., and go list -mod=readonly -deps ./cmd/seed ./seeds/... passed from backend.
- Production seed files differ only in package declarations. UTF-8, formatting, current references, and whitespace were checked. No live database or full Compose startup was required for this relocation.

## Non-Reusable Tool Errors

- An orchestration script had an extra closing brace and failed JavaScript parsing before execution. Removing the brace allowed the relocation call to succeed; validate orchestration syntax before dispatch.
- A patch attempted delete and add operations on the same path, which apply_patch rejected before mutation. The corrected explicit UTF-8 file write succeeded and its tests passed; use a single update operation or an explicit validated write when replacing a file. These were editing-tool mistakes, not Go command or project failures.

- Follow-up correction: moved all entries from seeds/tests into the existing empty seeds/test directory and renamed the Go package to test. The initial rename script asserted the destination was absent; that assertion and the subsequent gofmt path check failed before edits. Inspection confirmed an empty destination, so moving entries without overwriting corrected the invocation. This one-off preflight error does not require a new command recipe; inspect destination contents before choosing a directory rename. Targeted seed tests, backend build, and seed vet passed after correction.
