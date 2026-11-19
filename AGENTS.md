# Working in this repo

The HivePaaS Go backend, and its dashboard in [web/](web/docs/DEVELOPMENT_GUIDELINES.md).
The two are changed together whenever the wire format moves.

## Read before changing code

| You are | Read |
|---|---|
| touching anything in `hivepaas_app/` | [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) |
| adding or changing an endpoint | ARCHITECTURE.md §2 — request and response shapes are conventions here, not free choices |
| working under `services/` | ARCHITECTURE.md §6 |

## Before calling it done

- `go build ./...`
- `golangci-lint run ./...` — the **whole** repo, not just the packages you touched.
  It enforces a 120-character line limit and US spelling, and both are easy to miss.
- `go test ./...`
- `make lint-proto` if any `.proto` under `hivepaas_app/interface/agent/proto/` changed. It needs `buf`
  (the target says how to install it) and compares the agent's contract against `main`'s: an agent from the
  previous release goes on serving the new control plane for the length of a rolling update.
- `make gen-swag` if any DTO or handler swag comment changed. `docs/openapi/swagger.json` is generated and
  committed, and CI fails when it is not current.
- If the wire format changed, the matching dashboard change belongs in the same piece of work.
