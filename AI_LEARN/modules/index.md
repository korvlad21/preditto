# Functional modules

Select documentation by the behavior being changed, not only by the directory name. Backend files are organized by technical layer, so a feature usually spans several directories.

| Task triggers | Module | Related context |
| --- | --- | --- |
| Register/login, passwords, JWT, refresh/logout, sessions | [Auth](auth.md) | Users for transactional account/profile/role writes |
| Account/profile storage, get_user_info, CurrentUser/LoadUser, role assignments | [Users](users.md) | Auth for identity; teams for favorites; tournaments for selection |
| Teams, countries, country filters, team logos and reference seeds | [Teams and countries](teams.md) | Users for favorite-team references; asset inventory for logos |
| Tournament lists/status, selected tournament, participant counts, prizes, stages and tournament seeds | [Tournaments](tournaments.md) | Users for profile writes/snapshot; auth for access identity |
| Match schema, stage/team references, scores, match status and schedule indexes | [Matches](matches.md) | Tournaments for stages; teams for the shared team directory |
| Prediction schema, participant/match uniqueness, predicted scores, result fields and tournament consistency | [Predictions](predictions.md) | Tournaments for participants/stages; matches for match references |
| Participant stage rankings, stage/cumulative positions, totals and leaderboard indexes | [Rankings](rankings.md) | Tournaments for participants/stages; predictions for result semantics |

For selection changes read users and tournaments; for registration changes read auth and users. For common response or middleware changes, read [architecture](../architecture.md) and affected module contracts. Frontend currently has only starter UI and static assets: use [its README](../../frontend/README.md), plus the relevant backend module when adding integration.

Countries remain in teams; participation and prize rules remain in tournaments; roles remain in users. Split them only when independent behavior and task routing justify it. Predictions currently implement persistence only; payments and scoring have no implemented modules yet.

## Adding or splitting documentation

Use one `<module>.md` per functional module and the [standard outline](../../docs/ai/skills/memory-workflow.md#module-outline). Add a row here when a new module is implemented. Introduce a directory only when substantial independent subtopics need selective loading; retain a short entry document and update inbound links. Do not create empty module stubs or copy shared architecture into each document.
