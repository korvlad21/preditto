# Current Iteration

Date: 2026-09-15

## Outcome

The four existing RBAC tables were empty and their migration files were still untracked. Application writers were stopped; the existing database was rolled back from version 8 through migration 4 to version 3, preserving three users. Migration 4 was revised to create roles without updated_at, its trigger, or its function. Its down now drops only roles. The no-longer-needed migration 8 pair was removed. Migrations 4-7 were reapplied and the Compose stack restarted.

The existing database reports version 7 with dirty=false, roles without updated_at, and three users. Migration SQL for roles and the actual database schema now agree.

## Verification

Fresh isolated PostgreSQL applied migrations 1-7, then rolled back 7-4 with down 4 and reapplied them successfully. SQL assertions confirmed absence of roles.updated_at, its trigger, and its function in both isolated and existing databases. Normal docker compose up -d --build completed; migrate and seed exited 0, backend and frontend run, and /health returned status ok.

An initial isolated down 5 from version 7 also rolled back migration 3; its current down leaves its trigger function and reapplying 3 failed. This controlled failure is recorded in [docker command knowledge](commands/topics/docker-failures.md#failure-20260915-003). The existing database did not roll back migration 3.

The previous iteration is preserved in [archive](archive/2026-09-15-remove-updated-at-iteration.md).
