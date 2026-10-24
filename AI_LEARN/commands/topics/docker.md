# Docker and PostgreSQL

Run Compose commands from the project root. Confirm the selected Docker context/project and intended database before operations. Read [setup](../../../README.md), [.env.example](../../../.env.example) and [Compose](../../../docker-compose.yaml); use `docker compose config --quiet` to validate configuration without dumping secrets.

<a id="pattern-20260912-001"></a>
## Startup and database commands

The dependency chain is PostgreSQL healthy -> migrate successful -> seed successful -> backend. The frontend depends on the backend. Seeds restore configured development fields and passwords; startup is not a read-only validation.

Use the canonical [migration guide](../../../backend/migrations/README.md) for up/version/down and [seed guide](../../../backend/seeds/README.md) for repeat-run behavior. Restarting the backend alone does not rerun migrations or seeds. If startup stops, inspect the failed dependency's logs before changing migration metadata.

The migration service uses `postgres:///` with PG* environment variables. Keep the trailing slash: historical `postgres://` was normalized into invalid connection syntax. Do not construct a credential-bearing URL just to work around this.

<a id="pattern-20260915-001"></a>
## Applied schema and recovery

For a fresh disposable PostgreSQL container, probe `pg_isready -h 127.0.0.1` inside the container and use the same TCP host for `psql`. A Unix-socket readiness check can succeed against the image's temporary initialization server, then an immediate migration connection can fail when that server stops (`No such file or directory` for the socket). TCP readiness avoids this initialization race; verified with the full migration 1-14 up/down/up cycle on PostgreSQL 18.6.

Normal `up` only applies pending versions. Use a new numbered up/down pair for changes to an applied/shared schema. Every version must have exactly one up and one down file; referenced tables must exist before FKs, and rollback must remove dependents first.

Read the current version and inspect affected objects/data before diagnosing a dirty schema. `force` changes version metadata only: it neither repairs SQL nor restores data. Historical force/drop/renumbering commands in the [archive](../../archive/command-history.md) describe specific old states, not reusable recovery recipes.

Calculate rollback counts from the current version. At current version 12, `down 1` removes prize rules; rolling back version 1 removes users and their data. Stop writers before an authorized rollback. Use disposable PostgreSQL for up/down/reapply verification, never a live database solely to test migrations.

Baseline migrations 4 and 10 were revised for selection storage. An older applied baseline needs an explicit data-preserving transition or deliberate local rebuild; normal `up` cannot reconcile it. Inspect the actual environment before choosing a recovery, and do not infer its state from dated memory.
