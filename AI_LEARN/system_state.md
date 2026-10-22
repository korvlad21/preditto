# Active risks and next steps

Reviewed from source on 2026-10-22. No application database or deployed client was inspected in this documentation change.

| Active issue | Evidence and next step |
| --- | --- |
| Previously applied baseline may differ from current SQL | Migrations [4](../backend/migrations/000004_create_user_info.up.sql) and [10](../backend/migrations/000010_create_tournaments.up.sql) were revised for selection. Before deployment/rebuild, inspect version and actual schema; plan a data-preserving transition for shared databases. Normal up does not replay old versions. |
| Full seed race validation is timing-sensitive | [Known Go limitation](commands/topics/go.md#known-seed-race-limitation): replace the fixed timestamp window in a separately scoped test change. Targeted auth race checks do not resolve this seed limitation. |
| Client rollout is unverified | The [response contract](../backend/docs/responses.md) uses data envelopes, string errors and 200 success without data for logout/selection. Check external clients before rollout; no application API client exists in the current frontend. |
| Runtime seed/schema adoption is unknown | Current development datasets and logos are documented in the [seed guide](../backend/seeds/README.md). Inspect the intended environment before applying seeds; they restore configured fields/passwords and depend on literal favorite-team IDs. |

Implemented behavior and enduring feature limitations belong in [modules](modules/index.md), not a duplicated feature list here. Local versions, deployment observations and private operational details belong in `AI_USER_LEARN`.
