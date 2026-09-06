# Knowledge Accumulation

## Purpose

Keep compact, reusable project knowledge in AI_LEARN/modules after completing tasks or focused research.

## Selection

Save findings that reduce future investigation or prevent likely mistakes:

- Module architecture, business rules, entity relationships, and data structure.
- Significant dependencies, API contracts, database rules, and cross-module effects.
- Constraints, nonstandard technical solutions, and project conventions.
- Architectural decisions and their rationale when supported by code, documentation, or a confirmed decision.

Do not save task narration, temporary details, obvious facts isolated in one file, source-code dumps, logs, one-off values, or speculative explanations. Keep private operational context and local environment details in the location required by AGENTS.md, outside project knowledge.

## Storage

- Start with AI_LEARN/modules/<module>/overview.md: purpose and scope, relevant architecture and rules, dependencies, source evidence, and material unknowns. Include a verification date and the scope checked; a date alone is not proof of freshness.
- Split a topic into a separate file only when it has enough reusable detail to warrant selective loading. Link it from the module overview.
- Keep [the module index](../../AI_LEARN/modules/index.md) as a short map of module ownership and links. Add or rename entries as verified architecture evolves; do not create empty directories for hypothetical modules.
- Give each fact one canonical home. Other modules should link to the owning contract or note instead of copying it. Place truly shared knowledge in one relevant project-level topic and link to it.

## Updating

1. Review the task's confirmed findings and affected existing notes; if there is no reusable discovery, do not manufacture a module update.
2. Check the existing canonical entry before adding text. Merge useful detail and correct outdated claims in place using [knowledge validation](knowledge_validation.md).
3. Write new Markdown in English and all created or edited text in UTF-8. Confirm an existing file's encoding before editing it.
4. Verify source references, internal links, and consistency across affected notes.
5. Follow AGENTS.md for iteration records: keep system_state.md focused on active state, risks, and the next step; keep only the latest iteration in current_iteration.md and preserve useful previous outcomes in a topic, changelog, or archive.

Module knowledge is a maintained reference, not an append-only work log.
