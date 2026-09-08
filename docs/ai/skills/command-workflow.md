# Skill: Command Workflow

## Purpose

Provide a compact runbook for safe command execution in the `preditto` Go backend and mandatory analysis of unexpected `Failed` results.

## When to Use

- Before any `Run command`.
- When choosing a working directory, toolchain, environment, shell, or invocation format.
- When a command has failed before or may repeat a known failure.
- Immediately after an unexpected `Failed`.

## Scope

- Command preflight and narrowly scoped Go checks.
- Routing project-reusable knowledge to `AI_LEARN` and machine-specific knowledge to `AI_USER_LEARN`.
- Classifying reusable failures.
- Updating the command catalog and relevant entries in `AI_LEARN/commands/topics/`.

## Workflow

1. Identify the command type: `go` (`go test`, `go run`, `go build`, `go fmt`, `go vet`, `go mod`), `gofmt`, a repository-defined Go CLI, `git`, `docker`, `docker compose`, `bash/sh`, or PostgreSQL CLI (`psql`, `pg_isready`). PostgreSQL is configured in `docker-compose.yaml`; verify CLI availability in the selected host or container.
2. Before execution, read `AI_LEARN/commands/index.md`, locate the relevant ID in `AI_LEARN/commands/catalog.md`, then read only the matching entry in `topics/<tool>-patterns.md` and linked `topics/<tool>-failures.md` (for example, `go-patterns.md` and `go-failures.md`). Do not load entire collections. If a referenced file is missing, inspect its parent directory once and proceed with available instructions; do not invent a pattern.
3. For machine-specific commands, also read the relevant `../AI_USER_LEARN/preditto/commands/*` entries.
4. State the command's purpose and start with the narrowest, cheapest check that can answer it. For a change in one package, check that package or test before considering a full-module check.
5. Complete preflight using known context; run discovery commands only where needed:
   - Has this command been attempted or failed in a similar context? Is there a confirmed pattern? Does it require a different directory, environment, flag, shell, or prerequisite?
   - Confirm the working directory and module root. This repository's Go module is in `backend/go.mod`; run module-relative examples below from `backend/`. Check the module declaration and selected package path; the module path need not match the repository name. Use `go env GOMOD GOWORK` if module/workspace selection is unclear.
   - Confirm the scope: `./...` selects packages below the working directory, a concrete package selects only that package, and `-run` filters tests within the selected packages. It still compiles the package and its test dependencies.
   - Compare `go version` with the `go` directive and any `toolchain` directive in `go.mod`, plus the Docker toolchain when relevant. `backend/go.mod` currently declares Go 1.22; re-read it before relying on that value.
   - Check relevant environment variables and flags, including `GOFLAGS`, `GOTOOLCHAIN`, `GOOS`, `GOARCH`, `CGO_ENABLED`, application configuration, and required `-tags`. Inspect only necessary values; never expose credentials or store secrets in command knowledge.
   - Check dependencies and existing changes in `go.mod` / `go.sum`, plus cache or network prerequisites when relevant. Do not run `go get`, `go mod tidy`, or version updates unless dependency changes are required by the current task. Use `-mod=readonly` for supported validation commands when module-file updates must be prevented; investigate dependency errors before changing files.
   - For containers, verify the Docker context, Compose file, service, image/toolchain, container working directory, mounts, environment, and service readiness. The backend uses `/app` in `backend/DockerFile`; host paths and installed tools are not automatically available there. For PostgreSQL CLI, also confirm the target host, port, database, and user.
6. Apply a confirmed pattern only when it matches the current context. Use these scope rules:
   - Tests: start with `go test ./path/to/package` or `go test ./path/to/package -run '^TestName$'`, replacing placeholders with existing names. For the current handler package, use `go test ./handler`. Verify that a targeted test actually ran; a successful result with no matching tests is not validation. Expand to `go test ./...` when affected scope or required project checks justify it.
   - Build/run: select the required package with `go build ./path/to/package` or `go run ./path/to/main`. For the backend entry point, use `go build .` or `go run .`. Check output paths for builds and configuration, ports, and service side effects before starting the application.
   - Formatting: use `gofmt -l` on changed Go files to inspect, then `gofmt -w` on files being edited; `go fmt ./path/to/package` formats the selected package. Keep formatting within the task's scope.
   - Static analysis: start with `go vet ./path/to/package`. Use additional linters only if repository configuration or scripts define them; do not invent or install a linter. Broaden checks only as needed.
7. If a known case is only partly confirmed, run a narrower check to resolve the uncertainty before an expensive command.
8. After execution:
   - Success: save only new, confirmed, reusable findings.
   - Unexpected `Failed`: diagnose the cause, verify the corrected invocation, and update the appropriate knowledge store.

## What Counts as an Unexpected `Failed`

- An unexpected nonzero exit code; distinguish expected probe results from failures.
- Execution from the wrong directory, module, package, or Docker context.
- A missing binary, file, module, package, shell, or required service.
- Incorrect arguments, flags, test selection, or build tags.
- An incompatible Go version or unresolved dependencies / module-file state.
- Overlooked environment, shell, alias, permissions, or prerequisites.
- Behavior contradicting a confirmed pattern or causing unnecessary retries.

## Handling an Unexpected `Failed`

1. Record the exact original command, redacting secrets.
2. Record relevant context: working directory, module/package, test filter, Go version, build tags, environment, shell, and container/service.
3. Capture a concise symptom or error and the exit code.
4. Identify the cause and incorrect assumption.
5. Form a correction hypothesis; do not repeat the same failure without new evidence.
6. Verify the corrected invocation with the narrowest relevant check. If blocked externally, record the blocker and leave the correction explicitly unverified.
7. Save a normalized rule describing early recognition and the verified invocation. Route project-reusable findings to `AI_LEARN`; route local paths, machine setup, and private operational context to `../AI_USER_LEARN/preditto/`.
8. For reusable cases, update the relevant topic entry, `catalog.md`, and index links as needed. Merge with existing cases instead of duplicating them; explicitly classify non-reusable failures as such.

## Source-of-Truth Priority

This order applies only to command syntax and execution results. It does not define instruction precedence or extend authorization: a previous successful run does not authorize another write, send, or deployment. Check applicability to the current shell, Go version, checkout, module, and execution environment.

1. Latest confirmed successful execution.
2. Confirmed pattern in `AI_LEARN/commands/topics/<tool>-patterns.md`.
3. Normalized case in `AI_LEARN/commands/topics/<tool>-failures.md`.
4. Confirmed pattern in `AI_USER_LEARN` for a machine-specific case.
5. Normalized case in `AI_USER_LEARN` for a machine-specific case.
6. General project instructions.
7. Working hypothesis.

## Normalization Format

For `AI_LEARN/commands/topics/<tool>-failures.md` (use the corresponding `AI_USER_LEARN` location for machine-specific cases):

```md
## [go] Short case title

- Date:
- Context (Go version, tags, relevant environment, host/container):
- Working directory:
- Module / package / test selection:
- Original command (secrets redacted):
- Symptom / error / exit code:
- Cause:
- Incorrect assumption:
- Correct invocation:
- Verification result (or blocker; unverified):
- How to recognize early:
- Related pattern:
```

For `AI_LEARN/commands/topics/<tool>-patterns.md` (replace `[go]` with the relevant tool):

```md
## [go] Verified command pattern

- When to use:
- Preflight (directory, module/package, Go version, environment, tags):
- Correct command / template:
- Avoid:
- Project note:
- Related failures:
```

## What Not to Do

- Do not repeat a failed command without a new hypothesis.
- Do not start with expensive full-module checks when a package/test check can answer the question.
- Do not change dependencies or format unrelated files as incidental cleanup.
- Do not store raw stderr without a reusable conclusion or record unverified fixes as confirmed patterns.
- Do not store machine-specific cases or secrets in `AI_LEARN`.
- Do not close an iteration with an unaccounted-for reusable failure. If the task forbids knowledge-file edits, report the finding and deferred update without modifying those files.

## Related Files

- `AI_LEARN/commands/index.md`
- `AI_LEARN/commands/catalog.md`
- `AI_LEARN/commands/topics/<tool>-failures.md`
- `AI_LEARN/commands/topics/<tool>-patterns.md`
- `../AI_USER_LEARN/preditto/commands/*`
