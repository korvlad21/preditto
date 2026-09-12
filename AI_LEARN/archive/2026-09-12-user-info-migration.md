# Current Iteration

Date: 2026-09-12

## Outcome

Added the 000003_create_user_info SQL migration pair without changing existing migrations or application code. The profile references users with a unique required user_id and ON DELETE CASCADE, and teams with nullable favorite_team_id and ON DELETE SET NULL. updated_at has a default only and is not automatically updated. The contract is recorded in [backend knowledge](../modules/backend/overview.md). The previous result is preserved in [archive](2026-09-12-teams-migration.md).

## Verification

An isolated temporary Docker project applied versions 1 through 3 using the existing migrate configuration. PostgreSQL assertions verified column definitions, primary key and identity, uniqueness, missing-reference rejection, nullability, length limits, timestamp default, absence of application triggers, unchanged updated_at on update, user deletion cascade, and team deletion setting the reference to null. Down 1 returned to clean version 2 and removed user_info and its identity sequence while preserving users and teams. Temporary containers, network, and test files were removed. UTF-8, knowledge links, whitespace, and existing application files were checked.

## Diagnostic

A documentation apply_patch attempted to match a sentence as a whole line, although it was part of a longer paragraph, and was rejected before making changes. The corrected patch used the observed heading as its context and succeeded. This one-off patch construction mistake is classified as non-reusable project knowledge; use exact complete-line context when constructing patches.
