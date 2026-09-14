# Working in this repo

The HivePaaS Go backend. The dashboard is a separate repo at `../hivepaas-dashboard`,
and the two are changed together whenever the wire format moves.

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
- `make gen-swag` if any DTO changed. `docs/openapi/swagger.json` is generated and committed.
- If the wire format changed, the matching dashboard change belongs in the same piece of work.
