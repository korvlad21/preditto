# Commands and verification

Choose one topic for the operation; routine reads/searches need no historical lookup.

| Operation | Source and prerequisites |
| --- | --- |
| Backend tests, vet, toolchain and database-test skips | [Go](topics/go.md), working directory `backend/` |
| Compose startup, migrations, seeds and schema recovery | [Docker/PostgreSQL](topics/docker.md), project root |
| Frontend development/build/lint | [Frontend README](../../frontend/README.md), working directory `frontend/` |
| Documentation-only changes | Strict UTF-8 decoding, local Markdown paths/anchors, module index coverage, source consistency and `git diff --check`; no application test run required |
| Unexpected failures or unfamiliar commands | [Command workflow](../../docs/ai/skills/command-workflow.md) |

Commands below are recipes, not evidence that the current checkout/environment was tested. Read relevant prerequisites and inspect the actual scripts before execution. Historical versions and database repairs are retained in [command history](../archive/command-history.md), outside default context.
