# Current Iteration

Date: 2026-09-16

Fixed migration 000003 rollback: it now drops the user_info update trigger before the table and drops its function afterward, with IF EXISTS for both optional objects. Removed an extra trailing COMMIT from the current up migration so applying it produces no transaction warning. The user_info update trigger remains part of the up migration.

Verified the actual up/down/up/down cycle in an isolated PostgreSQL schema; the trigger and function were removed and the schema was dropped. Also verified down against a temporary table with no trigger or function. The live application database was not rolled back.

The previous iteration is preserved in [archive](archive/2026-09-16-countries-migration-reapply.md).
