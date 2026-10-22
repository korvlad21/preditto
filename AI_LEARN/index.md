# Project knowledge

Start with the [module index](modules/index.md) and read only the module(s) matching the task. The root [AGENTS.md](../AGENTS.md) owns workflow requirements; this directory stores project facts.

| Need | Read |
| --- | --- |
| Functional behavior and source/test entry points | [Modules](modules/index.md) |
| Layering, shared request flow, persistence boundaries | [Architecture](architecture.md) |
| Build/test, Docker or database commands | [Command index](commands/index.md) |
| Known active risks or incomplete verification | [System state](system_state.md) |
| Resume the latest significant change | [Current iteration](current_iteration.md) |
| Explain a past decision or failure | Search `archive/` only for that question |

Exact HTTP contracts live in [backend/docs](../backend/docs); setup lives in the [root README](../README.md) and [frontend README](../frontend/README.md). Avoid duplicating their payload examples and operational recipes in modules. Source wins over stale descriptions; runtime claims require execution evidence.
