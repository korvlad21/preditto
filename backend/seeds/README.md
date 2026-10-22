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

The registered order is `countries → teams → users → user_info → roles → user_roles → tournaments → tournament_participants → tournament_prize_rules`. The users seed creates or updates 20 development accounts, matching existing users by username or email and preserving their IDs. Each password is hashed with bcrypt at its default cost before writing `password_hash`; reruns restore the configured account fields and generate new valid salted hashes. If username and email belong to different existing users, the unique constraints reject the update and the whole seed transaction rolls back.

The tournament seed requires migrations 000010 and 000012 and upserts `uefa-champions-league-2026-2027` by slug with status `active`. The seeded dates cover the league phase through the final, from 2026-09-08 through 2027-06-05, consistent with [UEFA](https://www.uefa.com/uefachampionsleague/news/02a6-20d57cfcd03e-407c22a7f465-1000--2026-27-champions-league-teams-dates-draws-format-final/) (checked 2026-10-16). These calendar dates are stored at midnight in the schema's timezone-free timestamps; they do not specify match kickoff times. Currency and tournament type use the shared `model.CurrencyUSD` (`USD`) and `model.TournamentTypeChampionsLeague` (`Champions league`) constants from `internal/model/tournament.go` and are restored on rerun. Tournament seed data uses a local `tournament` struct and slice, following the teams seed pattern. Cost is the decimal string `10.00`, inserted and restored on rerun without floating-point conversion. IDs and created_at are preserved; updated_at is refreshed on conflict. Prize rules resolve the actual tournament ID by slug and upsert places 1/2/3 to 50.00/30.00/15.00 percent. The remaining 5% is not allocated; unrelated tournaments and rules are untouched. PostgreSQL tests cover fresh and existing tournaments, repeat runs, missing/changed prize restoration, currency/type/cost restoration, and full rollback on a prize write failure.

The tournament_participants seed requires migration 000011 and resolves all tournament IDs by slug, caching each lookup for the duration of the seed function. The first tournament's fixed selection contains 10 unique users from the users seed, excluding korvlad21, hanna_stoma, and boroda: eight active, one not_paid, and one blocked participant. It resolves user IDs by username and upserts on (tournament_id, user_id), restoring status while preserving id and joined_at. New joined_at values use the database default. A missing user or tournament fails the shared transaction; no particular numeric tournament ID is required.

Two additional active tournaments use the same upsert path. `uefa-champions-league-2026-2027-gel` has the name `Лига чемпионов УЕФА 2026/2027 (GEL)`, the same dates and type as the original, a cost of `40.00` GEL, and prizes of 50/30/15 percent. `la-liga-2026-2027` has the name `Ла Лига 2026-2027`, a cost of `20.00` USD, type `La Liga`, and prizes of 45/30/20 percent. `model.CurrencyGEL` and `model.TournamentTypeLaLiga` follow the existing model constants. La Liga starts on 2026-08-15 according to [LALIGA's first-round schedule](https://www.laliga.com/noticias/horarios-de-la-primera-jornada-de-laliga-ea-sports-2026-27) and ends on 2027-05-30 according to [RFEF's season calendar](https://rfef.es/es/noticias/la-temporada-202627-ya-tiene-establecidas-sus-fechas-clave), checked on 2026-10-18. These are calendar dates, not kickoff timestamps.

All participant groups use explicit tournamentSlug/username/status entries in the single `tournamentParticipants` slice. All accounts selected for the additional tournaments have `model.UserStatusActive` in the users seed. The GEL tournament has 14 active and one not_paid participant; La Liga has six active and one blocked participant. Account status and participation status are separate: the blocked La Liga participant still has an active user account. Each group has unique users; groups may overlap. Membership is fixed independently of the users slice order, and upserts restore statuses while preserving identities and joined_at. The original ten usernames and participation statuses remain unchanged. PostgreSQL tests cover nonsequential tournament IDs, an unrelated tournament at ID 1, active accounts for the additional groups, unique membership, preserved identity/joined_at, status restoration, transaction rollback, tournament fields, and prize restoration.

The user_info seed retrieves each user's actual ID by username and upserts on the unique `user_id`. It restores names, writes SQL NULL for avatar_url, and sets updated_at with PostgreSQL CURRENT_TIMESTAMP. All 20 users have matching profiles; five have no favorite team. Favorite team IDs are used directly from the existing teams, without changing or remapping the teams seed. If a referenced team ID is absent, the foreign key rejects the profile and rolls back all seed changes.

The teams seed contains 36 entries. It validates nonempty names/slugs, unique slugs within the dataset, exactly three uppercase ASCII letters in every short_name, and country codes present in the countries seed. `ON CONFLICT (slug) DO UPDATE` uses the existing unique constraint to prevent duplicate rows and restore seeded names, abbreviations, and countries. It preserves IDs, created_at, and profile references; unrelated teams are untouched. Logo URLs are derived as `/logos/teams/<slug>.svg`, supplied on insert and restored on conflict. See the [asset inventory](../../docs/assets/team-logos.md) for static files and provenance.

Tests in `test` exercise the public runner with a test SQL driver. Tests of private validation helpers remain in `development`, since Go directories are separate packages. From `backend`, run `go test ./seeds/...`; the CLI remains `go run ./cmd/seed` after migrations with PostgreSQL configured.

To run the PostgreSQL integration tests from `backend`, set `PREDITTO_SEED_TEST_DSN` to a test database connection string and run `go test -mod=readonly -count=1 -v ./seeds/test`. Without that variable, these integration tests are skipped. Each test applies the existing migrations in a temporary schema and drops it afterward; the test database user must have permission to create schemas. Tests cover repeated runs, bcrypt verification, existing username/email matches with nonsequential user IDs, NULL fields, restored profile data, and full rollback on identity conflicts or missing favorite teams.

The roles seed upserts `admin` and `participant` by unique code, restoring their names while preserving IDs and created_at. The user_roles seed resolves users by username and roles by code, then assigns admin to korvlad21 and hanna_stoma and participant to boroda. Existing assignments and unrelated roles are preserved. Missing users or roles fail the transaction.

Reusable persistence functions `repository.AssignUserRole` and `repository.RemoveUserRole` live in `backend/internal/repository/user_roles.go`. They accept a database or caller-owned transaction and user/role IDs. Assignment uses `ON CONFLICT (user_id, role_id) DO NOTHING`; removal deletes only that pair and succeeds if absent. Database errors, including invalid foreign keys, propagate to the caller. There is no dedicated role-management service/API or permission-enforcement middleware. AuthService already uses role assignment during registration; these helpers themselves remain persistence operations without authorization policy.

The current roles migration has no updated_at column; seeds write only code and name and also work with an additional defaulted updated_at column. PostgreSQL tests additionally cover role names, nonsequential user/role IDs, repeated assignments/removals, restoration of removed assignments, foreign-key errors, and transaction rollback.

The countries seed contains 16 entries in `development/countries.go`. Add countries to its data slice using a nonempty name and a unique three-uppercase-ASCII-letter short_name. It validates the dataset and upserts on short_name, restoring names while preserving IDs and created_at. New timestamps use the database default; unrelated countries are untouched. Migration 000002 provides the required unique constraint. Apply migrations before seeds.
