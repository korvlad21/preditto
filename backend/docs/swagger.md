# Swagger UI

Swagger documents all 11 existing `/api` POST operations plus `GET /`, `/hello`
and `/health`. Matches, predictions, rankings and tournament stages currently
have no HTTP endpoints to document. The [response contract](responses.md) and
module API guides explain behavior beyond the interactive schemas.

## Run and authorize

Configure PostgreSQL and JWT variables using the [setup guide](../../README.md).
Set `APP_ENV=development` explicitly, then start the backend:

```sh
cd backend
APP_ENV=development go run ./cmd/api
```

Open [Swagger UI](http://localhost:8080/swagger/index.html). With Compose, set
`APP_ENV=development` in `.env` and recreate/start the backend through the normal
development workflow. The application does not load `.env` when run on the host.
Use the backend origin directly; the frontend's `/api` proxy does not proxy Swagger.

Only the exact environment value `development` registers `/swagger/*any`.
Unset, empty, production, staging and other values leave both UI assets and
`/swagger/doc.json` unavailable (404), regardless of Gin debug/release mode.
There is no separate switch that enables Swagger in production.

1. Expand register or login, click **Try it out**, edit the JSON, and **Execute**.
2. Copy `data.access_token` from the successful response.
3. Click **Authorize**, enter `Bearer <access_token>` (including the prefix), and confirm.
4. Execute a protected user, favorite-team or tournament operation. Public routes
   do not require authorization. Missing/invalid/expired access JWTs return 401.
5. Refresh uses `refresh_token` in its JSON body. After rotation, replace the pair
   and authorize again with the new access token. Logout also takes a refresh token.

Requests use the UI's current host and scheme; no localhost server URL is baked
into the specification. Authorization is not persisted across page reloads.
Swagger 2.0 represents Bearer authentication as an API-key header, so the UI
does not add `Bearer ` automatically. Refresh tokens are not access credentials.

## Generate and check

From the repository root:

```sh
make swagger
make swagger-check
```

Go and Make are required; the first run downloads the pinned generator and its
dependencies. No global `swag` install is needed. `SWAG_VERSION` in the
[Makefile](../../Makefile) matches the runtime swag version in
[go.mod](../go.mod). Update them together deliberately.

`make swagger` generates only [swagger/docs.go](swagger/docs.go). Commit it with
the source changes. The imported package serves the JSON specification at
`/swagger/doc.json`; separate checked-in JSON/YAML copies are unnecessary.
Generation has no timestamp. Never edit generated output by hand.

`make swagger-check` generates into a temporary directory, compares the output
without modifying the checkout, then tests route coverage, JWT metadata and the
development gate against the actual application composition. It detects both
missing and obsolete routes, including service/demo routes. The
[GitHub Actions workflow](../../.github/workflows/swagger.yml) runs this on pushes
and pull requests. Semantic changes still require updating annotations: byte
comparison cannot infer changed business behavior from a handler implementation.

## Source ownership

- [cmd/api/main.go](../cmd/api/main.go): API metadata, Bearer definition, health annotation
  and route composition. `newRouter` is shared with coverage tests.
- [Handlers](../internal/handler): operation annotations beside the implementations,
  including body requirements, successful responses, errors and security.
- [Request DTOs](../internal/dto): existing validation tags plus documentation-only
  constraints, descriptions and examples. Do not change binding rules to describe an API.
- [Models](../internal/model) and auth response DTOs: actual payload definitions,
  timestamp/int64 formats and `x-nullable` metadata for nullable fields.
- [apidocs/models.go](../internal/apidocs/models.go): documentation-only descriptions
  of anonymous envelopes/list containers, reusing actual payload types. A separate
  empty-success schema preserves the absence of `data`; favorite-team data permits null.
- [router/swagger.go](../internal/router/swagger.go): development gate and embedded UI assets.

Follow the mandatory [API documentation maintenance rule](../../AI_LEARN/index.md#api-documentation-maintenance)
for every API creation, change or deletion. Keep Markdown contracts, annotations
and generated output consistent without duplicating business models.

The integration uses [swag](https://github.com/swaggo/swag),
[gin-swagger](https://github.com/swaggo/gin-swagger) and its embedded
`swaggo/files` UI assets. Versions are pinned; no CDN is needed to render the UI.
