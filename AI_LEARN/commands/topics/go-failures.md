# Go Command Failures

<a id="failure-20260908-001"></a>
## [go] Unsupported version flag

- ID: FAILURE-20260908-001
- Date: 2026-09-08
- Classification: Intentional, user-requested workflow exercise; not an accidental failure or application defect. The CLI syntax correction is reusable.
- Context: Host shell invocation with `GOTOOLCHAIN=local`; no container, build tags, package execution, or dependency changes.
- Working directory: `backend/`, relative to the project root.
- Module / package / test selection: `backend/go.mod` declares `github.com/korvlad21/firstGoWeb`; no package or test was selected.
- Original command: `GOTOOLCHAIN=local go --version`
- Symptom / error / exit code: `flag provided but not defined: -version`; exit code `2`. The usage output lists `version` as a subcommand.
- Cause: `--version` was passed as a top-level flag instead of using the `version` subcommand.
- Incorrect assumption: Go accepts the generic `--version` convention used by some other CLIs.
- Correct invocation: `GOTOOLCHAIN=local go version`
- Verification result: The corrected invocation printed the local Go version and exited with code `0` in the same directory. This verifies the invocation, not the backend build or tests.
- How to recognize early: For a Go version check, use the documented subcommand. Do not assume flags are shared between tools or retry the unsupported flag.
- Related pattern: [PATTERN-20260908-001](go-patterns.md#pattern-20260908-001).
