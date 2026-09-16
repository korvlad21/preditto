# Current Iteration

Date: 2026-09-16

Repaired the local dirty migration version 3. The user_info table existed but was empty; its trigger and function were absent. Removed that partial table and optional dependent objects, forced version 2, and ran migration up through version 8. Moved migration 3 COMMIT to the end of its up file so table, function, and trigger are created atomically. The previously fixed down removes all three objects.

Verification: live schema_migrations reports version 8 and dirty=false. Compose migrate and seed exited 0; backend is healthy and frontend runs. The database contains three users, 36 teams, three user_info records, and 16 countries. The current migration 3 files passed up/down/up/down in a disposable PostgreSQL schema. See [normalized recovery](commands/topics/docker-failures.md#failure-20260916-001).

The previous iteration is preserved in [archive](archive/2026-09-16-user-info-down-fix.md).
