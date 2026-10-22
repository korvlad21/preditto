# Command workflow

Use for build/test/database/container operations, unfamiliar invocations or unexpected failures. Routine reads, searches and instruction bootstrap do not require reading command history.

## Preflight and execution

1. Establish purpose, working directory and affected scope. Use the [command index](../../../AI_LEARN/commands/index.md) and the relevant topic only; do not read every prior failure.
2. Verify prerequisites from existing context: actual scripts/package paths, toolchain, flags, required configuration, dependency state and service readiness. Inspect only necessary environment values; never print secrets.
3. For Go, work from `backend/`, inspect `go.mod`, and select existing packages/tests. Use `-mod=readonly` for validation; do not run dependency updates as incidental repair. A `-run` filter still compiles its packages; confirm it selected tests.
4. For Docker/SQL, confirm context, Compose project, service, mounts and target database. Host and container paths/tools differ. Database-changing verification belongs in a disposable environment.
5. Start with the narrowest useful check and expand for affected behavior or a new concern. Do not repeat successful suites without a reason. Do not present skipped tests as passed integration coverage.

## Unexpected failure

Distinguish expected probe/negative-test results from unexpected failures. Capture the redacted command, relevant context, symptom and cause. Form a correction hypothesis before retrying. Verify it with the smallest useful check; label an externally blocked correction unverified.

During authorized implementation, merge reusable lessons into the relevant command topic. Keep one concise record: context, failed invocation, cause, corrected invocation, verification and prevention rule. Historical detail may go in the archive when it explains a decision. Local/private cases belong in `AI_USER_LEARN`. Non-reusable failures need only a brief report; analysis-only work must defer file edits.

A prior successful command is evidence of syntax and behavior in that context, not permission to run it against another environment. Trust current artifacts and the latest applicable confirmed result over older recipes.
