# Current Iteration

Date: 2026-10-17

Added LoadUser after RequireAuth on the shared protected API group. It reuses UserService.GetUserInfo and stores a shared UserInfo pointer in Gin Context. The repository query includes selected_tournament_id; get_user_info reads the context without another database call. The documented snapshot is fresh per request; successful writes must update the shared object for later consumers in that request.

Targeted middleware and PostgreSQL user HTTP tests passed with race detection, including invalid authentication, missing users/profiles, database-error secrecy, one load per request, shared-object updates, and tournament changes/deletion with the same JWT. The complete affected handler, middleware, and router suites also passed with race detection against PostgreSQL, including auth regression tests; affected-package vet passed. No migration, dependency, selection-write endpoint, or application database changes were made. No unexpected command failures occurred.

Next step: implement tournament selection persistence separately if requested. Existing environments must have the selected_tournament_id column before running this query.

Previous iteration: [selected-tournament baseline](archive/2026-10-17-selected-tournament-baseline-iteration.md).
