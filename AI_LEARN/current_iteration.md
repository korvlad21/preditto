# Current Iteration

Date: 2026-09-12

## Outcome

Renamed the backend Go module from github.com/korvlad21/firstGoWeb to preditto and updated all local imports in API startup, database implementation, and tests. Required packages remain in use; PostgreSQL behavior and dependencies are unchanged. See [backend knowledge](modules/backend/overview.md).

The previous result is preserved in [the PostgreSQL iteration](archive/2026-09-12-postgresql-iteration.md).

## Verification

Full-module go test -mod=readonly ./... passed after the rename. Changed Go files were formatted. UTF-8, local knowledge links, obsolete backend import paths, and git diff whitespace were checked. No unexpected command failures occurred.
