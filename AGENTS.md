# AGENTS.md

## Scope and invariants

Work as a careful project engineer: understand the request, inspect relevant code, make minimal authorized changes, and verify the result.

- Write instructions and project knowledge in English. Confirm existing text is valid UTF-8 before editing; write and validate UTF-8 afterward.
- Keep project facts in `AI_LEARN`; keep private/local environment details in ignored `AI_USER_LEARN`. Never store secrets in project knowledge.
- Code and current artifacts establish implementation facts; confirmed execution establishes runtime facts. Neither grants authorization. Respect the user's task scope.
- Analysis-only tasks must not change code or documentation. Report discrepancies and recommendations instead.
- Do not invent features, broaden a refactor, or repeat a failed approach without a new hypothesis.

## Load only relevant context

1. Read this file and [coding conventions](.aiassistant/rules/project.md).
2. Use the [knowledge index](AI_LEARN/index.md) and [module index](AI_LEARN/modules/index.md) to select the owning module(s).
3. Read their documentation and relevant source/tests before development. Read [architecture](AI_LEARN/architecture.md) for unfamiliar or cross-cutting work.
4. Read any local `AGENTS.md` before working in its subtree. Do not reread unchanged instructions within the same session.

Do not load all knowledge or archives by default. Routine read/search commands need no command-history preflight. For builds, tests, database or container work, use the [command index](AI_LEARN/commands/index.md) and only the relevant topic. Consult `AI_USER_LEARN` only when local environment context is necessary.

## Mandatory module maintenance

- Before implementation, identify the functional owner through the module index and verify its documented behavior against the affected code.
- When an authorized change alters business rules, API behavior, models/schema, dependencies, constraints or documented source ownership, independently update the corresponding `AI_LEARN/modules/<module>.md` in the same change. Keep detailed API contracts in `backend/docs` and link to them.
- When implementing a new functional module, automatically create its grounded module document and register its purpose, task triggers and dependencies in `AI_LEARN/modules/index.md`. Use the [module outline](docs/ai/skills/memory-workflow.md#module-outline); do not create placeholders for planned features.
- Update shared architecture or instructions only when boundaries or conventions change. Link shared rules instead of copying them into modules.
- Do not create extra documents or update memory merely because a task occurred. Internal changes that leave documented facts unchanged need no documentation edits. The analysis-only restriction above also applies to discovered documentation drift.

## Conditional workflows

- Text edits: [encoding-safe editing](docs/ai/skills/encoding-safe-editing.md).
- Nontrivial command execution or unexpected failures: [command workflow](docs/ai/skills/command-workflow.md).
- Research, module documentation or knowledge updates: [memory workflow](docs/ai/skills/memory-workflow.md).

Diagnose unexpected failures. Record reusable cause, context, correction and verification once in the relevant command topic when documentation edits are authorized; merge existing lessons. Report non-reusable or deferred findings without creating repetitive logs. Never label an unexecuted correction as verified.

## Completion

Verify affected behavior with appropriate targeted checks; documentation-only edits require UTF-8, reference and source-consistency checks. A skipped integration test is not evidence of database behavior. Report checks and material limits accurately.

Keep `system_state.md` limited to active risks and next steps, updating it only when they change. Update `current_iteration.md` for a significant handoff; preserve useful prior findings in their owning topic or archive when not already retained. Completion does not require a new historical file for every task.
