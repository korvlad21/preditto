# Current Iteration

Date: 2026-09-16

Fixed the migration ordering defect introduced by adding a countries foreign key to migration 2: migration 2 again creates teams without that later relation, and migration 9 adds the required teams.country column and foreign key after countries exists. Migration 9 backfills country codes for the 36 known seeded team slugs, then enforces NOT NULL. The team seed now supplies and restores country codes.

The local database had dirty version 2 with no teams or countries and zero users. After isolated migration 9 up/down/up and targeted Go checks passed, the version marker was set to 1 and migrations applied through version 9. Seeds completed. The database reports version 9 with dirty=false, 16 countries, 36 teams, three users, and no invalid team country references. PostgreSQL seed integration tests pass. See [normalized recovery](commands/topics/docker-failures.md#failure-20260916-002).

The previous iteration is preserved in [archive](archive/2026-09-16-dirty-user-info-recovery.md).
