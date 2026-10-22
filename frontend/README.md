# Preditto frontend

React and TypeScript application built with Vite. [src/main.tsx](src/main.tsx) mounts [src/App.tsx](src/App.tsx), which currently contains the starter screen/counter. There is no API client, application router, authentication flow or tournament UI yet.

## Local development and checks

From `frontend/`, use Node/npm compatible with [package.json](package.json) and the lockfile. The repository's container toolchain is defined in [DockerFile](DockerFile).

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
