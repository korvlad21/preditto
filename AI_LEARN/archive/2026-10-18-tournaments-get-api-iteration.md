# Current Iteration

Date: 2026-10-18

Implemented protected GET /api/tournaments/get_all_tournaments with all 13 requested fields, optional status filtering, nullable prize percentages, distinct active participant counts, and selection from middleware.CurrentUser. The repository performs one list query without multiplying tournament rows; the service reuses the loaded selection. The architecture follows the teams API. See [the contract](../../backend/docs/tournaments.md) and [backend knowledge](../modules/backend/overview.md#tournaments-api).

Targeted HTTP tests passed against disposable PostgreSQL with race detection, including authentication, all filter modes, SQL injection input, nullable fields, duplicate participant user IDs, zero counts, selected tournament changes with the same token, user isolation, empty results, and error responses. Affected-package and API entry-point vet passed. Formatting and strict UTF-8 checks passed. No migrations, dependencies, or existing business rules changed.

Non-reusable navigation failure: `rg --files AI_LEARN/archive/history` exited 2 because the optional history directory does not exist. The incorrect assumption was that this optional archive destination was present. Corrected discovery used `Path('AI_LEARN/archive').iterdir()` successfully; previous iteration records were preserved in the existing archive directory. Rule: inspect the existing archive parent before selecting an optional history path.

Non-reusable editing failure: apply_patch rejected a patch that deleted and added the same iteration path in one call. The patch did not apply. Replacing each delete/add pair with one Update File operation succeeded. Use one update operation per existing path when replacing its contents.

Next step: integrate the documented GET endpoint into the client. Existing environments still require the current tournament tables and selected_tournament_id column.

Previous iteration: [shared user context](2026-10-17-shared-user-context-iteration.md).
