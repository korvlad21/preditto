# Current Iteration

Date: 2026-09-15

## Outcome

Restored migration 000004 to the SQL shape already applied in PostgreSQL: roles has updated_at and its dedicated update trigger/function. Added migration 000008 with an up direction that removes that column and both update objects, and a down direction that restores them. The seed SQL leading-character error found during normal Compose startup was corrected in backend/seeds/development/user_info.go.

The existing PostgreSQL database had version 7 and retained roles.updated_at after migration 000004 was edited. Normal Compose startup applied version 8; schema_migrations now reports version 8 with dirty=false. No PostgreSQL volume or existing user table was dropped.

## Verification

Fresh isolated PostgreSQL applied migrations 1-8, rolled back migration 8, and reapplied it. SQL assertions verified column, trigger, and function absence after up and presence after down. Normal docker compose up -d --build completed with migrate and seed exit code 0, backend and frontend running, and /health returning status ok. The live PostgreSQL assertion confirmed version 8 with dirty=false and no roles.updated_at, roles trigger, or roles function.

The previous seed iteration is preserved in [archive](archive/2026-09-13-seed-iteration.md).
