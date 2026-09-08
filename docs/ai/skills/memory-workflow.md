# Skill: Memory Workflow

## Purpose

Manages what should be stored in knowledge, where it should be stored, and prevents an iteration from being considered complete without updating the current state.

## When to Use

- when a new confirmed project finding appears;
- when a finding about the user's local environment appears;
- when a significant iteration is completed;
- when it is necessary to decide whether to update `AI_LEARN` or `AI_USER_LEARN`.

## What It Covers

- separation of project and user-specific memory;
- normalization of knowledge instead of storing raw logs;
- synchronization of the active `system_state.md` without accumulating history;
- iteration completion criteria.

## Workflow

1. Classify the knowledge:
   - project-specific and reproducible — store it in `AI_LEARN`;
   - machine-specific or private — store it in `AI_USER_LEARN`;
   - mixed — split it without duplication.

2. Link the knowledge to a real artifact: a file, class, config, table, route, endpoint, job, or confirmed command execution.

3. Store not the raw output stream, but a short rule or fact that reduces the chance of future errors.

4. After each significant iteration, synchronize the relevant `system_state.md`: keep only the active state, risk, and next step; do not append a completed task as another historical entry.

5. If the iteration introduced new command errors or prevented previously known ones, update the corresponding `commands/*` entries as well.

6. If a topic grows too large, add a short summary and a link from the index instead of duplicating the same text across multiple places.

7. Replace `current_iteration.md` with the result of the latest iteration. Preserve the previous result in a thematic file, `AI_LEARN/archive/history/changelog.md`, or `archive/` if it has not already been recorded elsewhere.

8. If a startup file has turned into a log, first preserve the full historical snapshot in `archive/`, then rebuild a compact up-to-date version.

## New Session

In a new session, treat memory as a working navigator:

1. read the root route;
2. open only the relevant entries;
3. follow confirmed patterns instead of rediscovering already known rules;
4. if memory conflicts with the code or the latest confirmed execution, update the memory.

## Completion Criteria

An iteration is complete only if all of the following conditions are met:

- the task is completed or an external blocker is described honestly;
- changes are minimal and verified;
- the relevant `system_state.md` reflects the current active state;
- `current_iteration.md` contains only the latest result;
- new important findings are stored;
- useful `Failed` results are normalized into knowledge or explicitly marked as non-reusable.

## Do Not

- do not store secrets or local access details in `AI_LEARN`;
- do not create empty entries just to satisfy the structure;
- do not duplicate the same knowledge across the root, skills, and memory;
- do not store a hypothesis without linking it to an artifact;
- do not append completed task history to startup files;
- do not consider an iteration complete without synchronizing the state.

## Related Files

- `AI_LEARN/index.md`
- `AI_LEARN/system_state.md`
- `AI_LEARN/project_overview.md`
- `AI_LEARN/current_iteration.md`
- `AI_LEARN/research/research_backlog.md`
- `AI_LEARN/research/coverage_map.md`
- `AI_USER_LEARN/*`
