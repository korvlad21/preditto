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

Compose supplies VITE_API_URL, but current frontend code does not read it. No Vite proxy or backend CORS middleware is configured; unmatched API OPTIONS requests receive the normal API 404. When adding browser integration, explicitly choose a same-origin proxy or appropriate CORS handling and use the [shared API contract](../backend/docs/responses.md) plus the relevant [module](../AI_LEARN/modules/index.md). Do not assume supplying an environment variable wires up requests.
