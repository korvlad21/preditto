# Preditto frontend

React and TypeScript application built with Vite. [src/main.tsx](src/main.tsx) mounts [src/App.tsx](src/App.tsx) inside a React Router `BrowserRouter`. App wraps the routes in `CurrentUserProvider` and uses two layout routes:

- `AuthLayout` (full-screen stadium background, no header): `/login` (inside `GuestOnly`: a signed-in user is redirected to `/`), `/register` (inside `RequireAdmin`: only users whose `get_user_info` `roles` contain `admin`; guests go to `/login`, others to `/`), `/forgot-password` (placeholder).
- [AppLayout](src/components/AppLayout/AppLayout.tsx) (`.app-shell`: fixed navy background with glow gradients, content max-width 1360px, shared [AppHeader](src/components/AppHeader/AppHeader.tsx) with [UserMenu](src/components/UserMenu/UserMenu.tsx)): `/` (home), and behind [RequireAuth](src/auth/RequireAuth.tsx) `/profile` ([ProfilePage](src/pages/ProfilePage/ProfilePage.tsx)), `/tournaments` ([TournamentsPage](src/pages/TournamentsPage/TournamentsPage.tsx)) and `/profile/edit` ([ProfileEditPage](src/pages/ProfileEditPage/ProfileEditPage.tsx)).

`/home` and unknown paths redirect to `/`. The app background is `images/app-background.jpg` (served as `/app-background.jpg`, not in the repository yet; the gradients show until it is added). At ≤ 1024px header links show icons only; at ≤ 640px the navigation becomes a fixed bottom tab bar and the header keeps the logo and the user menu.

Layout: each page lives in its own folder, `src/pages/<Page>/<Page>.tsx` with its `<Page>.css` next to it when it has page-specific styles. Reusable components follow the same pattern in `src/components/<Component>/`. Shared design tokens (`--ui-*` on `:root`) and primitives (`.ui-panel` glass panel, `.ui-alert`, `.ui-button-outline`) are in [src/styles/ui.css](src/styles/ui.css). Styles shared by all auth pages (stadium background, card, form, footer) are in [src/styles/auth.css](src/styles/auth.css); its `--auth-*` tokens are aliases of `--ui-*` plus auth-only input/error values; the `visually-hidden` utility is in `src/index.css`. CSS is global (not scoped), so keep class names page-prefixed or shared deliberately.

[src/api/client.ts](src/api/client.ts) posts JSON to relative `/api` paths, unwraps the response envelope and throws `ApiError` (HTTP status + error code). [src/api/teams.ts](src/api/teams.ts) loads teams and countries from the real API (one shared request per page load). Backend shapes are typed in `src/types`; Russian country labels for codes from the seed are in [src/data/countries.ts](src/data/countries.ts).

The login and registration forms apply backend `LoginRequest`/`RegisterRequest` limits client-side (username 3–50 `[a-zA-Z0-9_.-]`, names ≤ 100, password 8 characters to 72 UTF-8 bytes). `login()` and `register()` in [src/api/auth.ts](src/api/auth.ts) call the real `POST /api/auth/login` (with `remember_me` from the "Запомнить меня" checkbox) and `POST /api/auth/register` (snake_case body). Login maps `INVALID_CREDENTIALS`, `USER_BLOCKED` and `INVALID_REQUEST` to Russian messages; registration maps 409 `USERNAME_ALREADY_EXISTS`/`EMAIL_ALREADY_EXISTS` to field errors and `FAVORITE_TEAM_NOT_FOUND` to a general error that clears the chosen team. On success, login reloads the current user and returns to the page `RequireAuth` redirected from (otherwise `/`). Registration is admin-only: `register()` sends the admin access token through `authPost`, maps 403 `FORBIDDEN`/401 `UNAUTHORIZED` to general errors, and on success shows a confirmation and clears the form for the next account (tokens of the created account are not stored). The login page has no registration link; admins open `/register` from the user menu. The avatar is chosen and previewed client-side only (PNG/JPEG/WebP, ≤ 5 MB): the backend has no upload endpoint. The favorite team is picked in `TeamPickerModal` (portal with focus trap, search and country filter); team logos come from `logo_url` with an initials fallback. There is no tournament UI yet.

### Tokens and authorized requests

- [src/api/tokens.ts](src/api/tokens.ts) stores the login pair under `preditto.auth` with the access expiry computed from `expires_in`: in `localStorage` when "Запомнить меня" is checked (backend long session), otherwise in `sessionStorage` (cleared with the tab; backend short session). Rotation keeps the storage chosen at login. Storage access is wrapped in try/catch; blocked storage means a signed-out user.
- [src/api/session.ts](src/api/session.ts) `authPost()` sends `Authorization: Bearer <access>`; it refreshes proactively when the access token expires within 30 s and, on a 401, refreshes once and retries. Refresh tokens are single-use, so refreshes are serialized: one shared in-flight request per tab and a Web Lock (`navigator.locks`) across tabs; inside the lock the pair is re-read and reused if another tab already rotated it. A 4xx refresh response (`INVALID_TOKEN`, `USER_BLOCKED`, `INVALID_REQUEST`) clears the pair and `authPost()` throws `ApiError(401, 'UNAUTHORIZED')`; network/5xx errors keep the pair.
- `logout()` revokes the refresh session via `/api/auth/logout` under the same lock and always clears local tokens; the access token stays valid server-side until expiry.
- [src/api/user.ts](src/api/user.ts) `getUserInfo()` calls `POST /api/user/get_user_info` (concurrent callers share one request). `getCurrentUser()` maps it to the camelCase UI model `User` (`src/types/user.ts`) and resolves `favoriteTeam` (name, logo) from the teams directory; `favoriteTeam` is `null` when none is chosen and `undefined` when the teams request failed.
- [src/auth/CurrentUserProvider.tsx](src/auth/CurrentUserProvider.tsx) is the single source of the signed-in user (`useCurrentUser()` in `src/auth/currentUser.ts`): status `loading` / `guest` / `ready` / `error`, `reload()` (after login, or "Повторить") and `signOut()` (logout + guest). The header menu, home page and profile read it; `RequireAuth` redirects guests to `/login` with `state.from`.
- `UserAvatar` shows `avatarUrl`, falling back to a silhouette when it is null or fails to load.

### Profile editing (no API yet)

- [ProfileEditPage](src/pages/ProfileEditPage/ProfileEditPage.tsx) edits first/last name and email (registration limits), the avatar (`AvatarUpload` with `initialUrl`) and the favorite team (`TeamPickerModal`, or "Убрать"); the login is read-only. The card carries `.auth-theme` so the shared auth form components keep their tokens. "Сохранить" is enabled only after a change.
- The backend has no profile-update or avatar-upload endpoint. `updateProfile()` in [src/api/user.ts](src/api/user.ts) sends nothing and rejects with `ApiError(501, 'NOT_IMPLEMENTED')`; the page then shows an info notice that nothing was saved. Its TODO holds a draft, unconfirmed request shape. When the endpoint exists, a successful save reloads the current user and returns to `/profile`.

### Tournament selection

- [src/types/tournament.ts](src/types/tournament.ts): `TournamentDto` and the request types mirror the Swagger schemas 1:1; the UI type `Tournament` is built by `toTournament()` in [src/api/tournaments.ts](src/api/tournaments.ts). API status `planned` → `upcoming`, `cancelled` → `finished`; a null/unknown status is derived from the dates. `participantsLimit` and `prizePool` are not provided by the API and stay undefined (their card cells are hidden); `prizes` are the per-place percentages; `isCurrent` is `is_selected_tournament`.
- Filters live in the URL: `?status=active|upcoming|finished` → API `status` (`upcoming` sent as `planned`), `?participation=mine|not-mine` → API `is_participate` true/false, `participation=current` is filtered on the client. Unknown values fall back to "all".
- "Выбрать турнир" calls `select_tournament` (`{ id }`) with all buttons disabled meanwhile: the new current tournament is shown optimistically, then the list is re-read; success shows the "Турнир выбран" toast and focuses the new "Текущий турнир" badge. Errors roll the selection back and show a Russian message for the API code inside the card; 401 reloads the current user so `RequireAuth` redirects to `/login`. There is no confirmation dialog: selection is not membership or payment.
- Clearing the selection is not part of the product: the current tournament's card shows only the "Текущий турнир" badge, without an action button. The selection changes only by choosing another tournament.

## Local development and checks

From `frontend/`, use Node/npm compatible with [package.json](package.json) and the lockfile. The repository's container toolchain is defined in [DockerFile](DockerFile). Bind mounts from Windows hosts do not deliver file events into the container and file-watch polling is not configured, so the Compose dev server may keep serving stale modules after edits; restart the frontend container (`docker compose restart frontend`) to pick them up.

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
The proxy only forwards headers; token storage and refresh live in the client
(see [Tokens and authorized requests](#tokens-and-authorized-requests)), so use
`authPost()` from `src/api/session.ts` for protected endpoints.

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
