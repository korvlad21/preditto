# Project knowledge

Start with the [module index](modules/index.md) and read only the module(s) matching the task. The root [AGENTS.md](../AGENTS.md) owns workflow requirements; this directory stores project facts.

| Need | Read |
| --- | --- |
| Functional behavior and source/test entry points | [Modules](modules/index.md) |
| Layering, shared request flow, persistence boundaries | [Architecture](architecture.md) |
| Build/test, Docker or database commands | [Command index](commands/index.md) |
| Swagger generation, UI access and API documentation checks | [Swagger guide](../backend/docs/swagger.md) |
| Known active risks or incomplete verification | [System state](system_state.md) |
| Resume the latest significant change | [Current iteration](current_iteration.md) |
| Explain a past decision or failure | Search `archive/` only for that question |

Exact HTTP contracts live in [backend/docs](../backend/docs); setup lives in the [root README](../README.md) and [frontend README](../frontend/README.md). Avoid duplicating their payload examples and operational recipes in modules. Source wins over stale descriptions; runtime claims require execution evidence.

## API documentation maintenance

Whenever any API is created, changed or deleted, updating Swagger documentation
is mandatory in the same change. Keep handler annotations and referenced DTO/model
schemas accurate for request parameters, required fields and validation, response
envelopes and nullability, HTTP statuses, error codes and authorization. Remove
obsolete operations. Run `make swagger` and `make swagger-check`, and commit the
generated output. Update the owning module and Markdown API contract when their
documented behavior changes. See the [Swagger guide](../backend/docs/swagger.md)
for source ownership and the automated drift/route-coverage checks.
