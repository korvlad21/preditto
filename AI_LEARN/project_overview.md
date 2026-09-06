# Project Overview

## Verified Structure

- backend/ contains the Go HTTP application. [backend/main.go](../backend/main.go) creates a Gin router, connects the hello handler, and defines a process-health endpoint.
- frontend/ contains the React application with TypeScript and Vite, as declared in [frontend/package.json](../frontend/package.json).
- .aiassistant/rules/ is the existing project-rule directory, including the leading dot.
- AI_LEARN/modules/ is the entry point for gradually collected module knowledge. Concrete business-module boundaries have not been researched in this iteration.

Verified on 2026-09-06 through the root directory layout, backend/main.go, and frontend/package.json. This is orientation, not a complete architecture inventory.

## Documentation Discovery

AGENTS.md references .aiassistant/rules/project.md, AI_LEARN/commands/*, .agents/skills/origami-task-lifecycle/SKILL.md, and docs/ai/skills/*. These resources were absent during setup; a targeted local guide search found no relocated copies. Do not infer their contents. Inspect parent directories before retrying a missing reference, and use available instructions with narrowly scoped source inspection. See [system_state.md](system_state.md) for the active gap.
