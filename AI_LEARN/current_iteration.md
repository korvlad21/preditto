# Current Iteration

Date: 2026-09-19

Follow-up: moved all 36 SVGs to the user-requested `frontend/images/logos/teams` directory. A relative `frontend/public/logos` symlink preserves the existing `/logos/teams/<slug>.svg` URLs and Vite public-asset build behavior without duplicate copies. File contents are unchanged. The production build passed again; all 36 development HTTP responses returned the original SVG bytes with the correct MIME type, and all 36 production copies matched.

Second follow-up: removed `frontend/public` entirely and configured `frontend/images` as Vite's `publicDir`. The existing favicon and SVG icon sprite moved into `frontend/images`, preserving their root-relative URLs.

Added 36 native SVG club logos to `frontend/images/logos/teams/`, matching every existing development seed slug. The seed now inserts and updates root-relative local `logo_url` values; all other team data and ordering are unchanged. Updated the existing seed test for logo paths and conflict updates.

The [asset inventory](../docs/assets/team-logos.md) records every source, local file, identity note, and verification. Strict UTF-8/XML, completeness, vector-only content, and fragment-reference checks passed. All logos rendered through Vite in Chromium on light and dark backgrounds; cleanup matched original pixels exactly. Targeted seed tests passed. No live database mutation was needed. Local download and scratch-tool issues are recorded separately in `AI_USER_LEARN`.

The production frontend build passed and contains all 36 SVGs unchanged. Temporary review pages were removed. The previous iteration is preserved in [archive](archive/2026-09-16-migration-order-iteration.md), with relative links adjusted for its new directory.
