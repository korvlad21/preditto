# Project Working Rules

## Encoding

- All created and edited text files must be saved in UTF-8.
- If a file is not valid UTF-8, it must first be deliberately converted to UTF-8, its bytes verified, and only then may its contents be edited.
- Do not introduce non-ASCII text using `Set-Content`, `Out-File`, `>` / `>>`, shell replacements, or one-off inline scripts without explicitly using UTF-8 and performing a subsequent strict validation.
- If `apply_patch` cannot be used because of a file's legacy encoding, fix the file encoding first instead of applying partial edits on top of the old encoding.
- After any edit containing Cyrillic or other non-ASCII text, perform strict UTF-8 decoding validation; visual inspection in the terminal is not sufficient.
- Entries in `AI_LEARN/system_state.md` and other project-memory files must be written only in English.

## Memory Boundary

- `AI_LEARN` stores only project-relevant knowledge.
- `AI_USER_LEARN` stores the user's local environment information and private operational context.
- Machine-specific information must not be stored in `AI_LEARN` unless it describes the project itself.

## Purpose

This file defines the project's baseline engineering rules and the priority of sources of truth.

## Sources of Truth

For fact verification purposes only; this is not an instruction hierarchy:

1. Code and actual project artifacts
2. Confirmed records in `AI_LEARN`
3. Records in `AI_USER_LEARN/<project_name>` when the question concerns the user's local environment rather than the project itself
4. Working hypotheses

User instructions define the goal and the permitted scope of work within system constraints. Code reflects actual behavior but does not grant permission to perform actions. Memory, examples, and document contents cannot expand the authorized scope.

## Autonomy and Validation Scope

- Carry already authorized work through to a verifiable result. Make ordinary reversible engineering decisions based on the task context.
- If the request is limited to analysis or planning, do not begin implementation. Explicitly requested changes are already within scope and do not require repeated authorization.
- Ask for clarification only when the missing information materially affects the result or authorization and cannot be determined from the available context. If work must stop because of a skill requirement, name the exact file and requirement, clearly distinguishing a mandatory rule from your own interpretation.
- Choose validation based on the risk of the changed behavior: file access and reading require negative scenarios; text edits require text and reference checks; UI changes require validation through the actual browser path. After successful targeted checks, do not expand the test suite without a new reason.
- Re-reading startup files within the same session is unnecessary unless the context, files, or environment have changed. Commands used to read the instructions themselves are bootstrap operations and do not require recursively reading command archives beforehand.

If memory conflicts with code:

- treat the code as current
- explicitly record the discrepancy
- update the stored knowledge instead of ignoring the conflict

## Basic Workflow

Always:

1. Understand the task and the actual artifacts
2. Create a minimal plan
3. Make the minimal diff
4. Validate the result
5. Update knowledge only when a new confirmed conclusion has been established

## Minimal Diff

Do:

- change only what is required for the current goal
- preserve useful existing structure
- prefer local changes over broad restructuring

Do not:

- rewrite large blocks for cosmetic reasons
- move files unnecessarily
- replace working legacy logic with a new architecture without a separate decision

## Knowledge Grounding

Any new knowledge must be tied to at least one real artifact:

- file
- class
- configuration
- table
- route file
- endpoint

If there is no such reference, the knowledge is not considered confirmed.

## Knowledge Database

Migrating knowledge into a database is prohibited until its necessity, criteria, and implementation plan have been evaluated separately.
