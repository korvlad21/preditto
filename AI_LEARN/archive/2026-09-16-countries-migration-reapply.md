# Current Iteration

Date: 2026-09-16

Moved countries.short_name uniqueness into migration 000008 and removed the unneeded 000009 pair, as requested. Before rollback, the existing database was at version 8 with dirty=false and countries had zero rows. Manual `docker compose run --rm migrate down 1` followed by `up` succeeded. PostgreSQL now reports version 8, dirty=false, and the countries_short_name_key constraint exists; countries remains empty.

Go seed tests pass on the host, and PostgreSQL seed integration tests pass inside the backend container using disposable schemas. Strict UTF-8 and git diff checks pass.

The previous country seed iteration is preserved in [archive](archive/2026-09-16-countries-seed-before-rollback.md).
