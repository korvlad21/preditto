# Project Overview

## Verified Structure

- backend/ contains the Go HTTP application. [backend/cmd/api/main.go](../backend/cmd/api/main.go) creates a Gin router, connects the hello handler, and defines a process-health endpoint.
- frontend/ contains the React application with TypeScript and Vite, as declared in [frontend/package.json](../frontend/package.json).
- .aiassistant/rules/ is the existing project-rule directory, including the leading dot.
- AI_LEARN/modules/ is the entry point for gradually collected module knowledge. Concrete business-module boundaries have not been researched in this iteration.

Verified on 2026-09-06 through the root directory layout, backend/cmd/api/main.go, and frontend/package.json. This is orientation, not a complete architecture inventory.

## Documentation Discovery

Verified on 2026-09-08: `.aiassistant/rules/project.md`, all four skills referenced by `AGENTS.md`, and the command knowledge navigation files exist. [Command knowledge](commands/index.md) now links to a verified Go pattern and a controlled failure record in `commands/topics/`. The current `AGENTS.md` does not reference a lifecycle skill. See [system_state.md](system_state.md) for active guidance gaps; historical discovery failures are preserved in [the archived iteration](archive/2026-09-06-current_iteration.md).
