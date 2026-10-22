# Project knowledge workflow

Use for module research, documentation changes or significant handoffs. [AGENTS.md](../../../AGENTS.md) owns mandatory maintenance and the analysis-only exception.

## Load and verify

Select the module through the [module index](../../../AI_LEARN/modules/index.md). Follow its source map into the affected routes, handlers, services, repositories, models, migrations, seeds and tests. Inspect related frontend behavior when relevant. Work in focused passes; do not load the entire repository.

Use code for implementation facts, SQL for constraints and confirmed execution for runtime facts. Separate current behavior, limitations and proposed work. Record a source/schema review date without implying tests ran. Report documentation drift during analysis; correct it during authorized implementation when relevant.

## Module outline

Use one file per implemented functional module. Omit empty or inapplicable sections; link exact contracts and shared procedures instead of duplicating them.

```markdown
# Module name

Reviewed: YYYY-MM-DD; state whether source inspection or runtime verification.

## Purpose and API
Ownership, task triggers, routes/access requirements and link to exact API contract.

## Business behavior
Validation, state transitions, filters, ordering, null/empty semantics, errors,
transaction/concurrency rules and boundaries that are easy to misunderstand.

## Sources, models and persistence
Links to routes, handlers, services, repositories, DTOs/models and migrations;
owned tables, keys, important constraints and deletion behavior.

## Dependencies and limitations
Related modules, seed identity/repeat-run rules, frontend integration status,
implemented constraints versus known gaps. No invented future features.

## Verification
Existing tests and relevant commands, required fixtures/DSNs, known skips/limits.
```

The [tournaments document](../../../AI_LEARN/modules/tournaments.md) is a concrete example. Keep field-by-field payloads in API contracts and operational commands in their owning guides. Split a large module only when independent task paths need selective loading.

## Update and register

Update the owning module alongside changed facts; register newly implemented modules in the index. Update architecture only for shared boundaries, and rules/skills only for reusable procedure changes. Avoid copying the same rule across these layers.

Store project-reusable knowledge in `AI_LEARN`; local environment/private details in ignored `AI_USER_LEARN`. Summarize findings with artifact links rather than raw logs. Never promote a hypothesis to a confirmed fact.

Keep `system_state.md` for active risks/next steps. Replace `current_iteration.md` only for a significant handoff and preserve useful prior findings once, in a thematic document or archive. Do not archive every small edit or append completed tasks to startup files. Historical records must remain visibly dated and outside default context.

## Validate

Before completion, check strict UTF-8, Markdown target paths/anchors, navigation to new modules, source names, routes, schema constraints and test prerequisites. Check inbound references when moving/removing files. Review the diff for duplicated rules, stale current-state claims, secrets and accidental application changes. Report source inspection and executed tests separately.
