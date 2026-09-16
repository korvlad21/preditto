# Current Iteration

Date: 2026-09-16

Reordered the eight migrations so users is 1, countries is 2, teams with the required country foreign key is 3, user_info is 4, and RBAC is 5-8. The team seed now supplies a valid country code for each of its 36 teams. Updated seed integration fixtures and migration documentation.

Before rebuilding the local database, confirmed that its rows matched development seed counts and identifiers and saved a verified custom-format dump. Fresh migration 1-8 up/down passed in an isolated schema, and PostgreSQL seed integration tests passed. Rebuilt the local public schema, applied migrations 1-8, and ran the normal Compose startup. The database reports version 8 with dirty=false; 16 countries, 36 teams, three users, and no invalid team country references are present. Migrate and seed exited 0, backend is healthy, and frontend runs. See [normalized migration-order failure](commands/topics/docker-failures.md#failure-20260916-002).

The previous iteration is preserved in [archive](archive/2026-09-16-country-fk-order-fix.md).
