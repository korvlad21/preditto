# Module Research

## Purpose

Research an unfamiliar or changed feature within the scope of the current task and capture an explanation of its architecture.

## Workflow

1. Use [module context loading](module_context_loading.md) to identify the owning module and check existing notes first.
2. Choose a stable module name from actual code boundaries, responsibilities, and project terminology. Reuse an existing name when it describes the same responsibility; example business domains are not a fixed taxonomy.
3. Check AI_LEARN/modules/<module>/. If absent, create it when the code confirms the module boundary, and add a concise overview.md with verified findings. Register the overview in [the module index](../../AI_LEARN/modules/index.md).
4. Work from general to specific: purpose and entry points, then the relevant execution path, business decisions, data access, and external contracts. Follow dependencies only as far as the task requires.
5. Stop when the task's questions are answered. Record material unknowns as unverified and outside the inspected scope instead of broadening research merely to complete a template.
6. Save reusable findings using [knowledge accumulation](knowledge_accumulation.md); validate and replace stale notes using [knowledge validation](knowledge_validation.md).

## What to Explain

Cover these subjects when supported by the inspected code and relevant to future work:

- Module purpose, ownership, and boundaries.
- Main entities, relationships, business rules, and execution flow.
- Entry points and API behavior; handlers or controllers.
- Services or use cases; repositories and important interfaces.
- Models, database tables, schema or migration sources, and persistence rules.
- Dependencies and contracts with other modules or external systems.
- Constraints, project conventions, and unusual design choices, including rationale only when evidenced.

Explain how responsibilities connect rather than listing files. Cite repository-relative source paths and useful symbols for important claims. Distinguish an absent layer from one that has not been inspected; do not claim a service, repository, or table exists merely because this checklist mentions it.
