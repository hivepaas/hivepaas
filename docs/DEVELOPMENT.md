# Development

How to get HivePaaS running on your own machine, and which of the three ways to
run the backend you want for the change you are making.

## At a glance

| You want | Run                                         | Where it answers |
|---|---------------------------------------------|---|
| The supporting cluster (once) | `make local-deploy`                         | postgres, redis, traefik, agent |
| The backend, fast loop | `make local-app-run`                        | http://localhost:10000 |
| The dashboard, bundled | `make local-build-dashboard`                       | served by the backend above |
| The dashboard, hot reload | `yarn dev` in `../hivepaas-dashboard`       | http://localhost:4321 |
| The backend as a real swarm service | `make local-app-up` + `make local-agent-up` | https://localhost |
| Extra swarm nodes | `make local-node-up` | `docker node ls` |

Sign in with `admin` / `abc123`.

## Prerequisites

- **Go 1.27+** (`go.mod` pins the toolchain).
- **Docker Engine 24+ with swarm mode.** Docker Desktop is fine; the install
  script runs `docker swarm init` for you.
- **Node and yarn**, only if you are touching the dashboard.
- **The dashboard repo checked out beside this one**, at `../hivepaas-dashboard`.
  Override with `HP_FE_DIR` if it lives somewhere else.
- **`make init`** once. It builds the `hivepaas-devtools` image, which is what
  `make lint`, `make gen-go` and every `make migrate-*` target actually run
  inside - you do not need `sql-migrate`, `psql` or `golangci-lint` on your host.

## 1. The local cluster

```bash
make local-deploy
```

This wraps `deployment/local/install.sh`, which:

1. creates `.appdata/hivepaas/` and a self-signed cert for `*.swarm.localhost`,
2. copies traefik's dynamic config into place,
3. runs `docker swarm init` and labels this node `hivepaas.role=control-plane`
   (the stack's placement constraints require that label),
4. creates the `hivepaas_net` overlay network,
5. deploys the `hivepaas` stack,
6. runs `make seed-data-with-clear` - migrations plus seed data.

Step 6 drops and recreates the seeded rows, so treat `make local-deploy` as
"start over", not "refresh".

Run it from the repository root and nothing else is needed. If you ever deploy
the stack by hand, do it from `.appdata/` - the file names host paths two ways,
and only that directory makes them agree:

| Written as | Resolved against | Example |
|---|---|---|
| `./hivepaas/traefik/...` (traefik) | the stack file's own directory | `.appdata/hivepaas/traefik/...` |
| `${PWD}/hivepaas` (app, worker, updater) | the shell running the deploy | `<wherever you are>/hivepaas` |

Deploying from the repo root leaves the app bound to an empty `<repo>/hivepaas`
while traefik reads the config that was just written, and docker reports no
problem at all. The install script asserts against exactly this before it goes on;
if it stops with `hivepaas_app is bound to ...`, that is what happened.

### What you get

| Service | Port | Notes |
|---|---|---|
| traefik | 80, 443 | dashboard on http://localhost:38081 |
| postgres | 35432 | `hivepaas` / `hivepaas` / `abc123` |
| redis | 36379 | |
| adminer | 38080 | also https://adminer.localhost |
| redisinsight | 35540 | |
| agent | - | global mode, gRPC on 10001 inside the overlay |
| app, worker, updater | - | deployed at **0 replicas** |

Those last three are deliberately not running. That empty slot is where your own
build goes, either beside the stack (§2) or inside it (§4).

## 2. Running the backend directly

This is the loop to use for almost everything.

```bash
make local-app-run
```

Then open http://localhost:10000.

`make build` produces a `hivepaas` binary instead, and `make build-agent` the
agent. Your IDE can run `hivepaas_app/cmd/app` directly; it only needs
`HP_CONFIG_FILE` in its environment.

**There is no `.env` loading in the Go process.** `HP_CONFIG_FILE` has to be in
the environment before the process starts. Without it - and without a
`config.toml` under `HP_APP_PATH` - startup fails with `HP_CONFIG_FILE must be
defined`.

`config/config.local.toml` is the one that matches the local stack: `platform =
"local"`, postgres on `localhost:35432`, redis on `localhost:36379`, and
`app_path = ".appdata/hivepaas"`, the same directory the stack's services bind.
Any scalar in it can be overridden by an `HP_`-prefixed environment variable
(`HP_HTTP_SERVER_PORT`, `HP_DB_HOST`, `HP_RUN_MODE`, …).

## 3. The dashboard

### Bundled into the backend

```bash
make local-build-dashboard
```

Runs `git pull`, `yarn install` and `yarn build` in `$HP_FE_DIR` (default
`../hivepaas-dashboard`), then moves the output to `dist-dashboard/`, which the Go
binary serves. Do this at least once, or the dashboard is a blank page.

It pulls that repo, so commit or stash your dashboard work first.

### Hot reload, for dashboard work

Run the backend as in §2, then in the dashboard repo:

```bash
yarn dev
```

Vite serves on `PORT` (4321 in its `.env.development`) and forwards `/_` to
`http://localhost:10000`, websockets included.

Leave `VITE_HP_DASHBOARD_BASE_URL` **empty** for this. The app then calls `/_/…`
on its own origin and the proxy carries the request; point it at an absolute
backend URL and the browser stops treating the call as same-origin, so the
httpOnly refresh cookie is not attached and you get bounced to the sign-in page
as soon as the access token expires.

## 4. Running as swarm services

Closest to production: your build runs as the `hivepaas_app` and `hivepaas_agent`
tasks, under swarm, with the real placement constraints, healthchecks, mounts and
the docker socket. Use it for anything that reads the app's own service, the
updater paths, the agent's view of its node, or code gated on `platform =
"remote"`.

```bash
make local-app-up      # build, load into hivepaas_app, wait for the task
make local-app-logs
make local-app-down    # back to the stack's image, env and 0 replicas

make local-agent-up
make local-agent-logs
make local-agent-down
```

It is slower than §2, though not as slow as it sounds: about 17s to build the
image, and 25-70s for the whole `up` depending on whether swarm has an old task to
stop first. The first run is the exception - it pulls the published dev image that
every build layers onto, which took 1m47s here.

Four things to know:

- **The app service runs `app+worker`.** If your `make run` backend is also up,
  both are pulling from the same task queue against the same database. Stop one
  of them when that matters.
- **The URL is `https://localhost`**, not the `app.dev.localhost` in the stack
  file. The app rewrites its own traefik labels from the domains in the database,
  so the stack's static labels are replaced the first time it runs.
  `docker service inspect hivepaas_app` shows what is actually in force.
- **`make local-agent-down` swaps the published image back** rather than scaling
  to zero. The agent is a global service and nothing else answers the app's gRPC
  calls for a node, so it has to keep running something.
- **Single-node only.** The images are tagged `:local` and exist on one daemon; on
  a multi-node swarm the other nodes have nothing to pull.

### Extra nodes

Anything that reasons about more than one node - global services, placement
constraints, node labels, a volume pinned to nowhere - needs a second node to mean
anything.

```bash
make local-node-up        # add one, ~1m20s: 15s for its daemon, the rest is the agent image
make local-node-down      # remove the one that was added last
make local-node-down-all  # remove all of them
```

Each node is a privileged `docker:dind` container running its own dockerd, joined
to the swarm as a worker. `local-node-up` names it for the first free number -
`hivepaas-node2`, then `hivepaas-node3` - so running it again adds another rather
than failing. The manager counts as node 1.

`hivepaas_agent` is a global service, so swarm puts a task on each node as it
joins: two extra nodes and the service reads 3/3.

`local-node-down` takes the highest-numbered one, undoing the last add. To remove
a particular one, name it: `make local-node-down LOCAL_NODE=hivepaas-node3`.

The numbering counts containers as well as swarm nodes, so a node that was only
half removed - a container left behind, or a node showing Down with no container -
keeps its number reserved and is cleaned up by the next teardown rather than
colliding with a new one.

Three things it deliberately does not do:

- **No node labels.** `app`, `db` and traefik are pinned to
  `node.labels.hivepaas.role == control-plane` and have to stay on the Desktop
  node, whose host paths their volumes point at.
- **No local images.** The `:local` tags that `local-app-up` and `local-agent-up`
  build exist on the Desktop daemon only; this node pulls published images.
- **Nothing persists.** The node's images and data live in the container's
  writable layer. Use real VMs - colima, lima, multipass - for anything that has
  to survive a node restart.

The one thing that is not obvious: Docker Desktop rewrites the stack's
`/var/run/docker.sock` bind to `/run/host-services/docker.proxy.sock`, a path that
exists only inside the Desktop VM. The agent is scheduled onto the new node too,
and swarm rejects the task with `bind source path does not exist` - over and over,
every few seconds. `local-node-up` creates that path on the node as a link to the
node's own docker socket, which is what the agent there should be talking to
anyway.

## 5. Database

Migrations and psql run inside the devtools image, so `make init` is the only
setup.

```bash
make migrate-status
make migrate-up
make migrate-new NAME=add_something     # creates the file in hivepaas_app/db/
make seed-data                          # migrate, then load seed data
make seed-data-with-clear               # wipe first - this is destructive
```

## 6. Before you push

```bash
make fmt                  # gofmt + import grouping, in place
golangci-lint run ./...   # or `make lint` to run it in the devtools image
make test
make gen-swag             # only if a DTO changed
```

Run the linter over the **whole** repo, not the packages you touched. It enforces
a 120-character line limit and US spelling, and both are easy to miss in a file
you never opened.

`make fmt` is a convenience, not a separate gate: the linter reports an
unformatted file as an issue of its own, so CI fails on one whether or not you
remembered to run it. Both apply the formatters configured in `.golangci.yaml`,
so they cannot disagree about what formatted means.

`docs/openapi/swagger.json` is generated and committed, so a DTO change that skips
`make gen-swag` leaves the committed contract wrong.

If the wire format changed, the matching dashboard change belongs in the same
piece of work.

## See also

- [ARCHITECTURE.md](ARCHITECTURE.md) - what belongs in which layer, and the
  request and response conventions.
- [recovery.md](recovery.md) - what to do when a configuration change makes the
  dashboard unreachable.
