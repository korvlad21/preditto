# Current Iteration

Date: 2026-10-16

Added 17 accounts and matching profiles to [users.go](../backend/seeds/development/users.go) and [user_info.go](../backend/seeds/development/user_info.go), bringing both slices to 20 records. The new accounts use unique usernames/emails, 14 ACTIVE and 3 BLOCKED statuses, realistic names, and four nullable favorite teams. The original three records are unchanged. Every literal favorite team ID was verified against the existing PostgreSQL teams table without writing application data.

Updated the existing [PostgreSQL user/profile test](../backend/seeds/test/users_postgres_test.go) to expect 20 records and reject missing/incomplete profiles or invalid favorite-team references on both seed runs. Updated the seed README counts. All development and seed tests passed against disposable PostgreSQL, including repeat runs and rollback cases; formatting, strict UTF-8, identity matching, original-record preservation, and diff checks passed. The application database was not seeded.

Keep usernames synchronized across the two data slices and verify literal favorite team IDs before adding profiles; teams are upserted by slug and IDs may differ across databases. Local Docker readiness and Go-cache sandbox failures were diagnosed and corrected; details remain in AI_USER_LEARN/commands/countries-api-verification.md. The previous iteration is preserved in [archive](archive/2026-10-16-tournament-cost-fix-iteration.md).
