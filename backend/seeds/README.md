# Database seeds

Seed logic lives in `backend/seeds/development`; `backend/cmd/seed` is the standalone CLI. It reuses the existing configuration and PostgreSQL connection packages. No seeding framework or additional dependencies are required.

From the project root:

```sh
docker compose up -d --build
```

Startup order is PostgreSQL healthy, migrate successful, seed successful, then backend. A failed seed exits nonzero and blocks backend startup. Seeds do not run inside the HTTP application. The one-shot seed service uses the same build and environment as the backend.

To rerun seeds manually after migrations:

```sh
docker compose run --rm seed
```

Compose starts required dependencies. For a running environment that needs newly added migrations and seeds applied explicitly:

```sh
docker compose run --rm migrate up
docker compose run --rm seed
```

Restarting only the backend does not rerun seeds.

`development.Run` executes the ordered seed list in a single transaction. An error rolls back the entire run. Add future seed functions in separate files in `development` and register them in dependency order in `development/seed.go`; each function must be safe to repeat.

The registered order is `countries → teams → users → user_info → roles → user_roles → tournaments → tournament_participants → tournament_prize_rules → tournament_stages`. The users seed creates or updates 20 development accounts, matching existing users by username or email and preserving their IDs. Each password is hashed with bcrypt at its default cost before writing `password_hash`; reruns restore the configured account fields and generate new valid salted hashes. If username and email belong to different existing users, the unique constraints reject the update and the whole seed transaction rolls back.

The tournament seed requires migrations 000010 and 000012 and upserts `uefa-champions-league-2026-2027` by slug with status `active`. The seeded dates cover the league phase through the final, from 2026-09-08 through 2027-06-05, consistent with [UEFA](https://www.uefa.com/uefachampionsleague/news/02a6-20d57cfcd03e-407c22a7f465-1000--2026-27-champions-league-teams-dates-draws-format-final/) (checked 2026-10-16). These calendar dates are stored at midnight in the schema's timezone-free timestamps; they do not specify match kickoff times. Currency and tournament type use the shared `model.CurrencyUSD` (`USD`) and `model.TournamentTypeChampionsLeague` (`Champions league`) constants from `internal/model/tournament.go` and are restored on rerun. Tournament seed data uses a local `tournament` struct and slice, following the teams seed pattern. Cost is the decimal string `10.00`, inserted and restored on rerun without floating-point conversion. IDs and created_at are preserved; updated_at is refreshed on conflict. Prize rules resolve the actual tournament ID by slug and upsert places 1/2/3 to 50.00/30.00/15.00 percent. The remaining 5% is not allocated; unrelated tournaments and rules are untouched. PostgreSQL tests cover fresh and existing tournaments, repeat runs, missing/changed prize restoration, currency/type/cost restoration, and full rollback on a prize write failure.

The tournament_participants seed requires migration 000011 and resolves all tournament IDs by slug, caching each lookup for the duration of the seed function. The first tournament's fixed selection contains 11 unique users from the users seed, including korvlad21 and excluding hanna_stoma and boroda: nine active, one not_paid, and one blocked participant. It resolves user IDs by username and upserts on (tournament_id, user_id), restoring status while preserving id and joined_at. New joined_at values use the database default. A missing user or tournament fails the shared transaction; no particular numeric tournament ID is required.

Two additional active tournaments use the same upsert path. `uefa-champions-league-2026-2027-gel` has the name `Лига чемпионов УЕФА 2026/2027 (GEL)`, the same dates and type as the original, a cost of `40.00` GEL, and prizes of 50/30/15 percent. `la-liga-2026-2027` has the name `Ла Лига 2026-2027`, a cost of `20.00` USD, type `La Liga`, and prizes of 45/30/20 percent. `model.CurrencyGEL` and `model.TournamentTypeLaLiga` follow the existing model constants. La Liga starts on 2026-08-15 according to [LALIGA's first-round schedule](https://www.laliga.com/noticias/horarios-de-la-primera-jornada-de-laliga-ea-sports-2026-27) and ends on 2027-05-30 according to [RFEF's season calendar](https://rfef.es/es/noticias/la-temporada-202627-ya-tiene-establecidas-sus-fechas-clave), checked on 2026-10-18. These are calendar dates, not kickoff timestamps.

All participant groups use explicit tournamentSlug/username/status entries in the single `tournamentParticipants` slice. All accounts selected for the additional tournaments have `model.UserStatusActive` in the users seed. The GEL tournament has 14 active and one not_paid participant; La Liga has six active and one blocked participant. Account status and participation status are separate: the blocked La Liga participant still has an active user account. Each group has unique users; groups may overlap. Membership is fixed independently of the users slice order, and upserts restore statuses while preserving identities and joined_at. The first tournament includes korvlad21 as an active participant alongside its original ten users. PostgreSQL tests cover nonsequential tournament IDs, an unrelated tournament at ID 1, active accounts for the additional groups, unique membership, preserved identity/joined_at, status restoration, transaction rollback, tournament fields, and prize restoration.

The user_info seed retrieves each user's actual ID by username and upserts on the unique `user_id`. It restores names, writes SQL NULL for avatar_url, and sets updated_at with PostgreSQL CURRENT_TIMESTAMP. All 20 users have matching profiles; five have no favorite team. Favorite team IDs are used directly from the existing teams, without changing or remapping the teams seed. If a referenced team ID is absent, the foreign key rejects the profile and rolls back all seed changes.

The teams seed contains 36 entries. It validates nonempty names/slugs, unique slugs within the dataset, exactly three uppercase ASCII letters in every short_name, and country codes present in the countries seed. `ON CONFLICT (slug) DO UPDATE` uses the existing unique constraint to prevent duplicate rows and restore seeded names, abbreviations, and countries. It preserves IDs, created_at, and profile references; unrelated teams are untouched. Logo URLs are derived as `/logos/teams/<slug>.svg`, supplied on insert and restored on conflict. See the [asset inventory](../../docs/assets/team-logos.md) for static files and provenance.

Tests in `test` exercise the public runner with a test SQL driver. Tests of private validation helpers remain in `development`, since Go directories are separate packages. From `backend`, run `go test ./seeds/...`; the CLI remains `go run ./cmd/seed` after migrations with PostgreSQL configured.

To run the PostgreSQL integration tests from `backend`, set `PREDITTO_SEED_TEST_DSN` to a test database connection string and run `go test -mod=readonly -count=1 -v ./seeds/test`. Without that variable, these integration tests are skipped. Each test applies the existing migrations in a temporary schema and drops it afterward; the test database user must have permission to create schemas. Tests cover repeated runs, bcrypt verification, existing username/email matches with nonsequential user IDs, NULL fields, restored profile data, and full rollback on identity conflicts or missing favorite teams.

The roles seed upserts `admin` and `participant` by unique code, restoring their names while preserving IDs and created_at. The user_roles seed resolves users by username and roles by code. It preserves the explicit assignments (admin for korvlad21 and hanna_stoma, participant for boroda) and appends every other user from the users seed with participant at package initialization, skipping usernames already in the assignment list. All 20 development users receive a configured role; reruns do not grow the list or duplicate assignments. Existing assignments and unrelated roles are preserved. Missing users or roles fail the transaction.

Reusable persistence functions `repository.AssignUserRole` and `repository.RemoveUserRole` live in `backend/internal/repository/user_roles.go`. They accept a database or caller-owned transaction and user/role IDs. Assignment uses `ON CONFLICT (user_id, role_id) DO NOTHING`; removal deletes only that pair and succeeds if absent. Database errors, including invalid foreign keys, propagate to the caller. There is no dedicated role-management service/API or permission-enforcement middleware. AuthService already uses role assignment during registration; these helpers themselves remain persistence operations without authorization policy.

The current roles migration has no updated_at column; seeds write only code and name and also work with an additional defaulted updated_at column. PostgreSQL tests additionally cover role names, nonsequential user/role IDs, repeated assignments/removals, restoration of removed assignments, foreign-key errors, and transaction rollback.

The countries seed contains 16 entries in `development/countries.go`. Add countries to its data slice using a nonempty name and a unique three-uppercase-ASCII-letter short_name. It validates the dataset and upserts on short_name, restoring names while preserving IDs and created_at. New timestamps use the database default; unrelated countries are untouched. Migration 000002 provides the required unique constraint. Apply migrations before seeds.

## Tournament stages

[The stage seed](development/tournament_stages.go) requires migration 000013 and resolves only `uefa-champions-league-2026-2027` by slug. It upserts 17 rows on `(tournament_id, code)`, preserving IDs and created_at, refreshing updated_at and restoring all configured fields. It does not delete other stages or populate the GEL/La Liga tournaments. Errors roll back the shared seed transaction.

Official sources checked on **2026-11-17**:

- [UEFA season calendar](https://www.uefa.com/uefachampionsleague/news/02a6-20d57cfcd03e-407c22a7f465-1000--2026-27-champions-league-teams-dates-draws-format-final/): dates of all rounds.
- [UEFA league fixtures/results](https://www.uefa.com/uefachampionsleague/news/02a8-2174c9e9019d-f909a77bd77a-1000--2026-27-champions-league-all-the-phase-fixtures/) and [earlier fixture publication](https://www.uefa.com/uefachampionsleague/news/02a8-2174c9e9019d-f909a77bd77a-1000--do-not-publish-202627-champions-league-all-the-league-phase-fixtures/): league fixtures and kickoff slots; the results view omits early kickoff annotations for completed games.
- [UEFA 2026/27 regulations, Article 25](https://documents.uefa.com/r/Regulations-of-the-UEFA-Champions-League-2026/27/Article-25-Stadium-announcements-and-kick-off-times-Online): early league slot, simultaneous last matchday, knockout slots and possible exceptions.
- [UEFA final kickoff announcement](https://www.uefa.com/uefachampionsleague/news/029c-1e9620d873ad-b9d62ed347c6-1000--uefa-champions-league-final-kick-off-moves-to-18-00cet-as-of/): 18:00 from 2025/26 onward. This specific decision is used instead of Article 25's generic 21:00 final wording; recheck the 2027 match-specific announcement when published.

The dataset below shows last **kickoffs in UTC**, not final whistles. Source Central European wall-clock times use CEST (UTC+02) in September/October 2026 and April–June 2027, and CET (UTC+01) in November–March. The seed encodes explicit offsets and converts to UTC; it does not depend on host timezone or an installed timezone database. Prediction timestamps and explicitly written stage audit timestamps use UTC in the schema's timezone-free `TIMESTAMP(6)` columns.

| sequence_no | code | Last kickoff / slot (UTC) | Points (outcome / exact) |
| --- | --- | --- | --- |
| 1 | league-1 | 2026-09-10 19:00 | 1 / 3 |
| 2 | league-2 | 2026-10-14 19:00 | 1 / 3 |
| 3 | league-3 | 2026-10-21 19:00 | 1 / 3 |
| 4 | league-4 | 2026-11-04 20:00 | 1 / 3 |
| 5 | league-5 | 2026-11-25 20:00 | 1 / 3 |
| 6 | league-6 | 2026-12-09 20:00 | 1 / 3 |
| 7 | league-7 | 2027-01-20 20:00 | 1 / 3 |
| 8 | league-8 | 2027-01-27 20:00 | 1 / 3 |
| 9 | round-32-1 | 2027-02-17 20:00 | 2 / 6 |
| 10 | round-32-2 | 2027-02-24 20:00 | 2 / 6 |
| 11 | round-16-1 | 2027-03-10 20:00 | 3 / 9 |
| 12 | round-16-2 | 2027-03-17 20:00 | 3 / 9 |
| 13 | quarter-final-1 | 2027-04-07 19:00 | 4 / 12 |
| 14 | quarter-final-2 | 2027-04-14 19:00 | 4 / 12 |
| 15 | semi-final-1 | 2027-04-28 19:00 | 5 / 15 |
| 16 | semi-final-2 | 2027-05-05 19:00 | 5 / 15 |
| 17 | final | 2027-06-05 16:00 | 6 / 18 |

Stages 1–8 use `group`; 9–17 use `play-off`. The requested “1/16 final” names refer to UEFA's knockout phase play-offs, not the August qualifying play-offs. Each two-leg round has separate first/return stages; sequence numbers follow the dataset order.

**Estimation and outstanding confirmation:** knockout match assignments are not yet confirmed in the reviewed sources. Stages 9–12 provisionally use the latest permitted slot (21:00 Central European time) on the last date. Stages 13–16 use the regulation's 21:00 slot; the final uses the announced 18:00 policy. These are documented scheduling assumptions, not verified individual fixtures. Update the dataset after UEFA confirms assignments or announces exceptions/postponements. No exact wall-clock final-whistle timestamps were verified, including for completed matchdays 1–4. All stage finishes currently use **last kickoff + 3 hours**, a conservative estimate covering normal breaks, stoppage time, extra time and penalties; it cannot guarantee coverage of long interruptions or postponements. Actual completion times require reconciliation when available.

`predictions_open_at` is the previous stage’s last kickoff + 8 hours, except stage 1 opens at **2026-08-20 00:00 UTC**. Prediction locking belongs to `matches.prediction_lock_at`. For example, stage 5 opens on 2026-11-05 at 04:00 UTC. The builder validates last-kickoff timestamp syntax before any stage writes.

Status uses one execution-time clock snapshot: `finished` at or after estimated finish; the first remaining stage is `active`, including between rounds and before prediction opening; later stages are `planned`. After the final all stages are `finished`. On 2026-11-17 stages 1–4 are finished, stage 5 active and stages 6–17 planned. Rerunning refreshes statuses; there is no background updater, live UEFA fetch or match-result dependency.

[Unit tests](development/tournament_stages_test.go) cover all 17 stages, points/kinds/order, UTC and daylight-saving offsets, opening rules and status boundaries. [PostgreSQL tests](test/tournament_stages_postgres_test.go) cover actual persistence, repeat-run restoration, nonsequential IDs, unrelated rows, a non-UTC database session and complete rollback on a rejected final-stage write. They use the same disposable-schema fixture as the other seeds, now including migration 000013.
