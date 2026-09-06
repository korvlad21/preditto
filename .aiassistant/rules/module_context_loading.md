# Module Context Loading

## Purpose

Load only the knowledge and source files needed for the current task. This rule extends [context_loading.md](context_loading.md); keep its minimum-context requirements.

## Workflow

1. Identify the affected behavior and likely module from the request and the project overview. Use [the module index](../../AI_LEARN/modules/index.md) to locate existing knowledge without reading every module.
2. Read only relevant notes in AI_LEARN/modules/<module>/. For a task spanning modules, load the owning module first and follow only dependencies needed for the task.
3. Before relying on a consequential claim, apply [knowledge validation](knowledge_validation.md) to its supporting implementation. Existing knowledge guides navigation; it does not replace checking code that will be changed.
4. Inspect the smallest relevant source area: an entry point, symbol, contract, or call path. Do not reread unrelated modules or repeat research supported by still-current evidence.
5. If knowledge is missing or insufficient, use [module research](module_research.md). Expand scope only to answer a specific unresolved question or follow an actual dependency.

## Boundaries

- Do not preload the repository, all of AI_LEARN, or every rule and module note.
- If the module is unclear, use a targeted path or symbol search before opening source files. Do not infer architecture from example module names.
- Missing documentation is a knowledge gap: check the parent directory once, use the available rules and relevant code, and record the gap. Do not repeatedly retry the missing path or invent its contents.
- After completing the task, apply [knowledge accumulation](knowledge_accumulation.md) to reusable discoveries.
