# Current iteration

Date: 2026-10-22

## Outcome

Reorganized AI documentation around [AGENTS.md](../AGENTS.md), one coding-conventions file, three conditional workflows, [architecture](architecture.md) and four [functional modules](modules/index.md). Module maintenance is mandatory for implementation changes that alter documented facts, including automatic documentation/index registration for new modules. Analysis-only tasks explicitly leave files unchanged.

Consolidated overlapping research/memory rules and command navigators, preserving dated command evidence in [the archive](archive/command-history.md). Updated setup, frontend, seed, migration and auth guidance against current source. Historical reference targets were repaired; current state now contains only active risks and next steps. No application code, migrations, seeds, dependencies or runtime configuration changed.

## Verification

Source/schema and existing test-scenario inspection covered all four modules and the shared backend/frontend boundaries. Strict UTF-8, local Markdown paths/anchors, module navigation, route/test-reference checks and git diff whitespace checks passed. The pre-edit SHA-256 baseline of 165 tracked non-Markdown files remained unchanged. Application tests, services, migrations and seeds were not executed for this documentation-only change; no deployed-state claims are made. External URLs were retained, not revalidated.

## Next step

Use the module routing and maintenance rules on subsequent implementation tasks; update only facts affected by the task. Existing runtime/test limitations remain in [system state](system_state.md). This change introduces instruction-based maintenance, not a CI gate or proof of identical behavior across assistant products.

Previous significant iteration: [UserRepository split](archive/2026-10-21-user-repository-split-iteration.md).
