# Current iteration

Date: 2026-10-25

## Outcome

Swagger UI is available at `/swagger/index.html` only with explicit
`APP_ENV=development`. The same gate protects the JSON specification and UI assets;
Gin debug mode alone cannot enable them. All eleven existing API POST routes and
three system/demo GET routes are documented, with Bearer authorization on the four
protected operations. Existing routes, middleware order and business contracts
are preserved.

Handler annotations reuse actual request DTOs and response models; only anonymous
envelopes/containers have documentation-only types. A single generated Go artifact
serves the JSON specification. `make swagger` regenerates it; `make swagger-check`
compares temporary output and checks actual registered routes. The new GitHub
Actions workflow invokes that check on pushes and pull requests.

The [Swagger guide](../backend/docs/swagger.md) owns startup, authorization,
generation and source ownership. The [mandatory maintenance rule](index.md#api-documentation-maintenance)
requires updating Swagger for every API creation, modification or deletion.

## Changed files

Created:

- [Makefile](../Makefile) and [Swagger CI workflow](../.github/workflows/swagger.yml).
- [Swagger route registration](../backend/internal/router/swagger.go).
- [Envelope schemas](../backend/internal/apidocs/models.go).
- [Generated specification](../backend/docs/swagger/docs.go) and [Swagger guide](../backend/docs/swagger.md).
- [Application Swagger tests](../backend/cmd/api/swagger_test.go).

Updated:

- [API composition](../backend/cmd/api/main.go), [configuration](../backend/internal/config/config.go)
  and [configuration tests](../backend/internal/config/config_test.go).
- Handler annotations: [auth](../backend/internal/handler/auth.go), [user](../backend/internal/handler/user.go),
  [team](../backend/internal/handler/team.go), [country](../backend/internal/handler/country.go),
  [tournament](../backend/internal/handler/tournament.go), [hello](../backend/internal/handler/hello.go).
- DTO metadata: [auth requests](../backend/internal/dto/auth/request.go),
  [auth responses](../backend/internal/dto/auth/response.go), [teams](../backend/internal/dto/teams/request.go),
  [tournaments](../backend/internal/dto/tournaments/request.go).
- Model metadata: [user](../backend/internal/model/user.go), [team](../backend/internal/model/team.go),
  [country](../backend/internal/model/country.go), [tournament](../backend/internal/model/tournament.go).
- [go.mod](../backend/go.mod), [go.sum](../backend/go.sum), [.env.example](../.env.example),
  [Compose](../docker-compose.yaml), [README](../README.md).
- [AI conventions](../.aiassistant/rules/project.md), [knowledge index](index.md),
  [architecture](architecture.md), [Go command guide](commands/topics/go.md), this handoff,
  and module ownership notes for [auth](modules/auth.md), [users](modules/users.md),
  [teams](modules/teams.md), [tournaments](modules/tournaments.md).

## Verification

- Generation and freshness checks passed; a deliberate generated-file change was
  rejected, then restored and the positive check rerun successfully.
- Full `go test -mod=readonly -count=1 ./...` passed against disposable PostgreSQL
  with both integration DSNs: 338 test/subtest pass events, no failures or skips.
- Race-enabled checks passed for cmd/api, auth, config, auth DTOs, middleware,
  router and handlers against the disposable database. The full seed race suite
  was not rerun; its existing timing limitation remains documented separately.
- `go build -mod=readonly ./...` and `go vet -mod=readonly ./...` passed.
- A real headless Chrome session executed all fourteen operations through Swagger
  UI against a seeded disposable database. Registration/login, Authorize, protected
  reads, selection, refresh and logout succeeded. Missing access credentials and
  refresh JWTs in Authorization were rejected; authorization cleared on reload.
  No browser page errors occurred. Browser tooling was temporary, outside the repo.
- Frontend build/lint and `docker compose config --quiet` passed. No application
  database was used; no hosted CI run or full Compose deployment is claimed.
- UTF-8, local Markdown references, Go formatting and diff whitespace checked.

## Next steps

Set APP_ENV explicitly in the intended development process or Compose environment.
Keep it unset or non-development elsewhere. Follow the Swagger maintenance rule
for subsequent API work. Existing unrelated risks remain in [system state](system_state.md).
The previous registration/session findings remain in the auth module and contract;
no additional historical archive is needed.
