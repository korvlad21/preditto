# Current Iteration

Date: 2026-10-18

Implemented [POST /api/tournaments/select_tournament](../../backend/docs/tournaments.md#select-a-tournament) through the existing DTO, router, handler, service, and repository layers. It rejects missing/null/nonpositive/non-integer IDs, checks tournament existence, updates only the current user profile, and refreshes the shared Gin snapshot after success. The repository accepts nil to clear the selection. Foreign-key failures and missing profiles have explicit errors; all selection failures, including middleware failures, use success:false through the shared response helper. Other endpoint error envelopes are unchanged. No schema or dependency changes were needed.

Verification: targeted selection and user-info PostgreSQL integration tests and authentication/user-context regression tests passed with race detection; affected-package/API vet passed. Tests cover nullable repository writes, repeated selection, user isolation, same-request context consistency, later listing with the same token, missing profiles/tournaments, forced write failure, and middleware failures. Router was compiled in the filtered test invocation but had no matching tests. UTF-8, formatting, and diff checks passed. No unexpected command failures occurred.

Next step: integrate the documented selection contract in the frontend.

Existing limitation found by source inspection: the older list test still expects name/ID order, while the repository orders by creation time descending. The API documentation was corrected to match code; list ordering and its legacy test were not changed. Later resolution: the [2026-10-21 contract update](2026-10-21-json-api-contract-iteration.md) corrected this legacy test without changing the SQL ordering.

Previous iteration: [positional participant literals](2026-10-18-positional-participants-iteration.md).
