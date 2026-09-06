# Knowledge Validation

## Purpose

Use saved knowledge as a navigation aid while keeping current project code as the source of truth for implementation behavior.

## Before Relying on Knowledge

- Check the note's sources and recorded verification scope. A recent timestamp does not guarantee correctness, and an old timestamp does not by itself require a full re-audit.
- Verify claims that affect the task's implementation, business behavior, API contracts, or persistence against the smallest relevant current source or schema definition. Consult focused diffs or tests when they help resolve uncertainty.
- Missing sources, renamed symbols, changed interfaces, or contradictory behavior trigger targeted research. Do not recheck every module for an unrelated task.
- If verification is unavailable, label the claim unverified and state the gap. Do not use it as a confirmed basis for a consequential decision.

## Resolving a Mismatch

1. Follow the current implementation. If it may violate a requested requirement, report the discrepancy explicitly rather than treating the old note as proof that the code behaves differently.
2. Identify the canonical note and all affected statements within the task's scope.
3. Correct or remove obsolete content in place. Do not append a second contradictory account beside it. Move historical rationale elsewhere only if it remains useful and is clearly marked as historical.
4. Update source references, verification scope, and module-index links when paths or ownership change.
5. Follow documented or code-confirmed dependencies to check affected consumers. Refresh only related knowledge when a contract, schema, or business rule changes across modules.
6. Recheck affected links and claims against the resulting implementation. If a check cannot be completed, keep the remaining gap explicit rather than claiming full verification.

Use [module research](module_research.md) for missing architectural context and [knowledge accumulation](knowledge_accumulation.md) for canonical storage. The combined process is: identify modules, load relevant knowledge, verify and investigate narrowly, perform the task, capture reusable findings, and refresh affected knowledge.
