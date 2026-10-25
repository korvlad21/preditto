# Preditto frontend

React and TypeScript application built with Vite. [src/main.tsx](src/main.tsx) mounts [src/App.tsx](src/App.tsx) inside a React Router `BrowserRouter`. App renders the routes (there is no shared header): `/` (empty home page), `/login`, `/register` and the `/forgot-password` placeholder; unknown paths redirect to `/`.

Layout: each page lives in its own folder, `src/pages/<Page>/<Page>.tsx` with its `<Page>.css` next to it when it has page-specific styles. Reusable components follow the same pattern in `src/components/<Component>/`. Styles shared by all auth pages (design tokens, stadium background, card, form, footer) are in [src/styles/auth.css](src/styles/auth.css); the `visually-hidden` utility is in `src/index.css`. CSS is global (not scoped), so keep class names page-prefixed or shared deliberately.

The login and registration forms mirror backend `LoginRequest`/`RegisterRequest` field names and limits with client-side validation only. Login calls `login()` from [src/api/auth.ts](src/api/auth.ts), which currently resolves a mock `AuthResponse`; registration submission is a no-op. There is no real API client, token storage or tournament UI yet.

## Local development and checks

From `frontend/`, use Node/npm compatible with [package.json](package.json) and the lockfile. The repository's container toolchain is defined in [DockerFile](DockerFile). Compose sets `VITE_USE_POLLING=true`, which enables file-watch polling in [vite.config.ts](vite.config.ts) because bind mounts from Windows hosts do not deliver file events into the container.

```sh
npm ci
npm run dev
```

Defined scripts:

| Command | Purpose |
| --- | --- |
| `npm run dev` | Vite development server |
| `npm run build` | TypeScript project build followed by Vite production build into dist |
| `npm run lint` | Existing ESLint configuration |
| `npm run preview` | Preview the built output locally |

There is no test script configured. Validate UI changes in the browser as well as with relevant build/lint checks. Project-wide Compose setup is in the [root README](../README.md).

## Assets and API integration

[vite.config.ts](vite.config.ts) sets publicDir to `images`. Files under `images/logos/teams` are served as `/logos/teams/<slug>.svg`, and copied to that public path in the build. Use the [logo inventory](../docs/assets/team-logos.md) for provenance and maintenance; no duplicate asset list is needed here.

The Vite development server proxies `/api` requests to the Go backend, preserving
the path (including `/api`), query, method, body and Authorization header. Use
relative URLs in React so requests go through the frontend origin:

```ts
const response = await fetch('/api/user/get_user_info', {
  method: 'POST',
  headers: { Authorization: `Bearer ${accessToken}` },
})
const result = await response.json()
```

The actual route is `/api/user/get_user_info`, not `/api/users/get_user_info`.
Use the [shared API contract](../backend/docs/responses.md) and relevant
[module](../AI_LEARN/modules/index.md) for other endpoints and response handling.
Token storage and automatic refresh are not implemented by the proxy.

Local `npm run dev` uses `http://localhost:8080` by default. Override it with
`API_PROXY_TARGET` in the process environment or `frontend/.env.local`.
Compose supplies `API_PROXY_TARGET=http://backend:8080`, using the backend
service name inside the Docker network. This variable is read only by the Vite
configuration and is not exposed as a `VITE_*` browser variable. Restart the dev
server after environment changes; recreate the frontend container when changing
its Compose environment.

These same-origin browser requests do not require backend CORS. Direct browser
requests to a separate backend origin still do. The proxy is enabled only for
the development server, not builds or `vite preview`. Production hosting must
route `/api/...` to Go separately, for example through Nginx on the same domain.
