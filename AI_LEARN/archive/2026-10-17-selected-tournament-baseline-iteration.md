# Current Iteration

Date: 2026-10-17

Moved selected_tournament_id into migration 000004 at the user's request. Migration 000010 adds the foreign key after creating tournaments and removes it before dropping tournaments. Removed migration 000013; Go code and the application database are unchanged.

Isolated PostgreSQL 18.6 SQL checks passed for fresh migration application, nullable BIGINT, selection without membership, invalid-reference rejection, ON DELETE SET NULL, partial rollback/reapplication, and full rollback/reapplication. The disposable container was removed. UTF-8 and diff checks passed; no unexpected commands failed.

Next step: deliberately rebuild the intended local database when authorized; normal up cannot update already applied migrations. Model and API integration remain separate work.

Previous iteration: [separate selected-tournament migration](2026-10-17-selected-tournament-migration-iteration.md).
