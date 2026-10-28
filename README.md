# Preditto

Go/Gin/PostgreSQL backend with a React/TypeScript/Vite frontend. The API implements authentication, current-user profiles, teams/countries and tournament listing/selection. The frontend currently contains routing, a home page, login/registration screens with mocked auth calls, a favorite-team picker loading teams from the API, and static team logos.

## Development with Docker Compose

From the project root, copy [.env.example](.env.example) to an untracked `.env` and configure all values. PostgreSQL requires POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB and POSTGRES_SSLMODE. Inside Compose, the database host is normally `postgres`.

Set two independent JWT_ACCESS_SECRET and JWT_REFRESH_SECRET values, each at least 32 bytes. For example, run `openssl rand -hex 32` separately for each secret. Keep secrets out of commits and command output shared in documentation. JWT_ACCESS_TOKEN_TTL must be whole seconds between 1s and 1h; JWT_REFRESH_TOKEN_TTL must be whole seconds and longer. JWT_REFRESH_TOKEN_SHORT_TTL defaults to 12h and must be longer than access TTL but shorter than JWT_REFRESH_TOKEN_TTL. The example uses 15m access, 12h short refresh and 720h remembered/legacy refresh. See [Remember Me](backend/docs/auth.md#remember-me) for optional request behavior.

```sh
docker compose config --quiet
docker compose up -d --build
```

[Compose](docker-compose.yaml) waits for PostgreSQL, applies migrations, runs development seeds, then starts the backend; frontend startup depends on the backend. Seeds create/update development data and restore configured passwords/profile fields on rerun. Read the [seed guide](backend/seeds/README.md) before using an existing database.

The backend is exposed on port 8080, frontend on 5173, and PostgreSQL on 5432. The backend container uses Air for rebuilds. For backend and its database dependencies only, use `docker compose up --build backend`.

```sh
curl http://localhost:8080/health
```

`GET /health` returns `{"status":"ok"}` for process health; it does not test database readiness. Non-API demo routes `GET /` and `GET /hello` also exist.

## Backend on the host

Use a Go toolchain compatible with [backend/go.mod](backend/go.mod). Export the PostgreSQL and four JWT variables above in the process environment; the Go application does not load `.env` itself, and host TTLs do not inherit Compose defaults. Use a host-reachable database address, normally `localhost` for the published Compose database.

Apply [migrations](backend/migrations/README.md) and provision required development data/participant role through the [seed workflow](backend/seeds/README.md) before exercising registration. Then:

```sh
cd backend
go run ./cmd/api
```

The server binds all interfaces on port 8080 by default; APP_PORT overrides the host process port. Starting the API itself does not apply migrations or seeds.

## SMTP email

The [mailer module](AI_LEARN/modules/email.md) sends plain-text messages with an
optional HTML alternative. Sending is disabled by default. There is no public
email endpoint and registration/authentication do not send messages.

Add the mail settings from [.env.example](.env.example) to your existing ignored
`.env` (do not overwrite existing database/JWT values). For Mail.ru, including an
`@inbox.ru` mailbox, use:

```dotenv
MAIL_ENABLED=true
MAIL_FROM_ADDRESS=your-mailbox@inbox.ru
MAIL_FROM_NAME=Preditto
SMTP_HOST=smtp.mail.ru
SMTP_PORT=465
SMTP_USERNAME=your-mailbox@inbox.ru
SMTP_PASSWORD='replace-with-an-application-password'
SMTP_TLS_MODE=tls
SMTP_AUTH_TYPE=plain
SMTP_TIMEOUT=10s
```

Create a dedicated external application password in Mail.ru's security settings
and enable SMTP access; see the [Mail.ru instructions](https://help.mail.ru/mail/login/mailer/).
Use the mailbox address for both `MAIL_FROM_ADDRESS` and `SMTP_USERNAME`. Never
commit `.env`, use the main mailbox password, print credentials, or pass them as
CLI arguments. Single-quote Compose `.env` values containing `$` to keep them
literal. Production should inject the same settings from deployment secrets.

From the project root, send one explicit test message to an address you control:

```sh
docker compose --profile tools run --rm --build --no-deps mailtest -to recipient@example.com
```

The tool reads the root `.env` through Compose, uses only mail configuration, and
does not start the API, PostgreSQL, migrations or seeds. It exits nonzero on an
invalid recipient/configuration, disabled mail or SMTP failure. To customize the
message, append `-subject 'SMTP test' -text 'Plain text' -html '<p>HTML alternative</p>'`.
The success message means SMTP accepted the email, not that it reached the inbox;
check the recipient's inbox/spam folder. This command sends a real email.

For host execution, export these variables through your environment manager,
then run `go run ./cmd/mailtest -to recipient@example.com` from `backend/`. Go does
not parse `.env` itself. After changing `.env`, recreate an already running API
container to refresh its environment; restarting alone keeps the old values.

`SMTP_TLS_MODE=tls` uses TLS from the first byte; `starttls` requires a successful
STARTTLS upgrade before authentication (typically port 587). Both verify server
certificates with TLS 1.2 or newer. Plaintext and opportunistic TLS are not
supported. `SMTP_AUTH_TYPE` supports `plain` and `login`, both only over TLS.
`SMTP_TIMEOUT` is a positive duration no greater than one minute and bounds the
whole send, including greeting/authentication and message transfer. With
`MAIL_ENABLED=false` (or unset), credentials are optional and sending returns
`mailer.ErrDisabled`; enabled but incomplete settings fail configuration loading.

To change provider, update host, port, TLS/auth modes, credentials and the
provider-authorized sender. The application-facing message format remains the
same. Future service integration can construct one sender with
`mailer.New(cfg.Mail)` at the composition root and call
`sender.Send(ctx, mailer.Message{To: address, Subject: subject, Text: text, HTML: html})`.
`To` is one mailbox; `Subject` and `Text` must be nonempty. HTML is optional and
is sent as supplied; use escaped templates when inserting user data. Pass the
caller's context. No connection is opened by the constructor.

Sending is synchronous with no queue or automatic retries. A lost final SMTP
reply can make acceptance uncertain; blindly retrying may duplicate a message.
Errors expose a safe operation/status code, with the original cause available
via `errors.Is`/`errors.As`; do not log unwrapped SMTP replies or entire config
objects. SMTP secrets are passed only to backend and the explicit mailtest tool,
not to seed/migrate/frontend. Future transactional use cases should send after
database commit and define their own retry/recovery policy.

## Documentation and checks

- [Swagger UI and generation](backend/docs/swagger.md): set `APP_ENV=development`,
  then open [Swagger UI](http://localhost:8080/swagger/index.html). Update with
  `make swagger`; verify with `make swagger-check`. Other environments do not expose Swagger.
- [Architecture and module navigation](AI_LEARN/index.md).
- [AI contributor instructions](AGENTS.md).
- [API response contract](backend/docs/responses.md), [auth](backend/docs/auth.md), [users](backend/docs/user.md), [teams/countries](backend/docs/teams.md), [tournaments](backend/docs/tournaments.md).
- [Backend test prerequisites and known limits](AI_LEARN/commands/topics/go.md).
- [Frontend setup, scripts and current integration status](frontend/README.md).
