# AGENTS.md

## Role

The agent works as a careful project engineer: first understand, then plan, then make minimal changes, then verify, and only then update knowledge.

## Global Invariants

- All instructions must be written in English, including `AGENTS.md`, rules, skills, and procedural guidance in documentation and knowledge files.
- All created and edited text files must use UTF-8. If the encoding has not been confirmed, the file is unsafe to edit.
- `AI_LEARN` stores only project knowledge; `AI_USER_LEARN` stores only the local environment and private operational context.
- An unexpectedly failed command must not go unaccounted for: its cause, context, corrected invocation, and rule must be stored in knowledge.
- If code conflicts with memory, trust the code. If a confirmed execution conflicts with command memory, trust the latest confirmed result.
- Minimize context, diff, and the number of executions. Do not repeat a failed approach without a new hypothesis.

## Minimum Context

Do not load the entire project or all of `AI_LEARN`.

Default reading order:

1. `AGENTS.md`
2. `.aiassistant/rules/project.md`
3. `AI_LEARN/index.md`
4. `AI_LEARN/project_overview.md`
5. then only the files needed

If the task uses commands, also read the following before the first `Run command`:

1. `AI_LEARN/commands/index.md`
2. only relevant entries from `AI_LEARN/commands/patterns.md`
3. only related entries from `AI_LEARN/commands/failures.md`

If the task depends on the user's local environment, also read the relevant files from `AI_USER_LEARN`.

If the target subdirectory contains a local `AGENTS.md`, read it after the root file and before working in that subdirectory.

## Skill Routing

Open only the skills actually needed for the scenario:

- `docs/ai/skills/encoding-safe-editing.md` — creating or editing text files, especially with non-ASCII text, uncertain encoding, or shell-based edits.
- `docs/ai/skills/command-workflow.md` — any `Run command`, choosing a safe invocation, analyzing `Failed` results, and updating `AI_LEARN/commands/*`.
- `docs/ai/skills/memory-workflow.md` — updating `AI_LEARN` or `AI_USER_LEARN`, completing an iteration, and normalizing knowledge.
- `docs/ai/skills/research-mode.md` — research mode, mapping an unfamiliar project area, and updating backlog and coverage.

## Constraints

- Do not overload context or analyze the entire project in a single pass.
- Do not rewrite large portions of code unnecessarily or perform bulk refactoring.
- Do not invent project structure or store unconfirmed conclusions as knowledge.
- Do not duplicate the same instructions across the root, skills, and memory without a useful reason.
- Do not make a series of similar attempts hoping one will work.

## Iteration Completion

An iteration is complete only if:

- the task is completed or has honestly reached an external limitation;
- changes are minimal and verified;
- the relevant `system_state.md` in `AI_LEARN` and/or `AI_USER_LEARN` reflects only the active state, risk, and next step, rather than receiving another historical entry;
- `current_iteration.md` contains only the latest iteration; the previous result is preserved in a thematic file, `changelog.md`, or `archive/`;
- useful new findings have been stored in knowledge;
- unexpected `Failed` results have either been normalized into knowledge or explicitly classified as non-reusable.
