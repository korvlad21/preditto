# Current Iteration

Date: 2026-10-17

Added migration 000013 for nullable user_info.selected_tournament_id, referencing tournaments(id) with ON DELETE SET NULL. The down migration removes only this column. Existing migrations, Go code, and the application database are unchanged.

An isolated PostgreSQL 18.6 container verified the SQL migration sequence, existing-profile preservation, nullability, valid selection without membership, rejection of missing tournaments, ON DELETE SET NULL, rollback, reapplication, and full reverse-order rollback. This check used psql, not migrate version tracking. The disposable container was removed. UTF-8 and diff checks passed; no unexpected commands failed.

Next step: apply the migration in the intended environment when requested; model, repository, and API integration remain separate work.

Previous iteration: [inline auth literals](2026-10-17-inline-auth-literals-iteration.md).
