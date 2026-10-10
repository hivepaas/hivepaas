# End-to-end tests

Playwright tests that drive the dashboard in a browser, against a running
HivePaaS. This is a package of its own, so its dependencies stay out of the
app's.

## Running them

Once:

```bash
cd e2e
yarn install
npx playwright install chromium
```

### On a throwaway HivePaaS (level 1 and up)

Level 1 makes and removes data; run it on an installation that is there to go:

```bash
yarn env:up     # builds the dashboard and the backend, then starts them: http://localhost:10100
HP_E2E_BASE_URL=http://localhost:10100 yarn test
yarn env:down
```

Without `HP_E2E_BASE_URL` the tests go to the local backend, and there levels 1
and 2 are left out - the run says so: they never make, deploy or remove
anything on a developer's own installation. The dashboard's dev server,
`http://localhost:4000`, counts as that installation too: it hands the API
calls to the local backend.

`env/up.sh` runs postgres and redis as containers, and the backend with its
agent inside a dind container that is a swarm of its own - never on this
machine's swarm, which it checks before it says the stack is up. It needs the
backend repo beside this one (or `HP_BACKEND_DIR`), Go, and the backend's
`hivepaas-devtools` image (`make init` there, once). `HP_E2E_PORT` moves it off
10100; `HP_E2E_SKIP_BUILD=1` reuses the last build. The template catalog is the
app-templates repo beside this one (or `HP_TEMPLATES_SRC`); without it, the
template test finds none. dind lists two GPUs of no hardware, `NVIDIA-GPU=GPU-e2e0` and
`AMD_GPU=0xe2e1`, for the apps that reserve one. dind is also given kopia, the backup engine, from
its pinned image, and Git repositories to build apps from, on dind's loopback,
made at fixed dates, so of fixed hashes: `git://127.0.0.1/e2e/shop.git` (two
commits on main, a branch, a pull request's head), `site.git` (a page of HTML)
and `args.git` (a Dockerfile of build arguments), by git's own daemon; and a
private one, over HTTPS with a token (`https://127.0.0.1:8443/cgi-bin/git/e2e/private.git`)
and over SSH with a key (`git@127.0.0.1:e2e/private.git`, the key left in
`env/.build/git-ssh-key`). The seed's notification channels - a mail account, a Slack, a
Discord and a Telegram, real ones - are removed: nothing the tests do is sent
to anyone.

The stack's proxy is the release's Traefik, inside dind; the apps' domains
answer on this machine at 10180 (HTTP) and 10443 (HTTPS), moved by
`HP_E2E_HTTP_PORT` and `HP_E2E_HTTPS_PORT`. A test reaches an app at
`https://<name>.localhost/`, the address a person would use: the browser is told
to connect for `*.localhost` where the proxy answers (`HP_E2E_INGRESS_HTTP` and
`HP_E2E_INGRESS_HTTPS` below), and the address stays as it is - a redirect
writes it whole. The backend in dind reaches the same names through Docker
Desktop's DNS, which answers `*.localhost` with loopback; a host whose DNS does
not fails the health check of an app's domain.

### On the local backend (level 0)

Against the local backend - see the backend repo's `docs/DEVELOPMENT.md`:

```bash
make local-deploy           # once: the cluster, and the seeded admin / abc123
make local-build-dashboard  # the dashboard the backend serves
make local-app-run          # http://localhost:10000
```

Then `yarn test` here, or `npm run e2e` from the dashboard's root: level 0
alone runs there. With the dashboard's dev server running (`yarn dev`),
`HP_E2E_BASE_URL=http://localhost:4000 yarn test` tests the code as it is now,
without building it into the backend first; two tests run at a time there. `yarn test:ui` opens Playwright's UI mode; `yarn report` shows
the last run, with a trace of each test that failed.

To run against another installation:

| Variable               | Default                          |
| ---------------------- | -------------------------------- |
| `HP_E2E_BASE_URL`      | `http://localhost:10000`         |
| `HP_E2E_USERNAME`      | `admin`                          |
| `HP_E2E_PASSWORD`      | `abc123`                         |
| `HP_E2E_RUN_ID`        | when the run started, in base 36 |
| `HP_E2E_INGRESS_HTTP`  | `127.0.0.1:10180`                |
| `HP_E2E_INGRESS_HTTPS` | `127.0.0.1:10443`                |

Point them only at an installation whose data does not matter: the tests make
what they need there, and remove it. A read-only account, as on the demo
servers, is enough for the sign-in tests alone.

### In CI

`.github/workflows/e2e.yml` checks the backend and app-templates out beside the
dashboard, makes the throwaway HivePaaS with `env/up.sh` on the runner, and runs
levels 0 and 1 on every push and pull request, every level every night and by
hand (Run workflow). A failed run keeps its report, traces and videos as the
`playwright-report` artifact, and shows the end of the backend's log.

## Writing tests

- **Tests start signed in**, with the session `tests/auth.setup.ts` saved. One
  that must start signed out says
  `test.use({ storageState: { cookies: [], origins: [] } })`.
- **Make what a test needs through the API.** The `api` fixture in
  `support/fixtures.ts` is signed in as the tests' user. Name what you make with
  `e2eName("...")`, and hand its removal to `cleanup(...)`. The screens are for
  the flow under test.
- **Find elements as a person does:** `getByRole`, `getByLabel`, `getByText`, by
  their accessible names. A button with only an icon needs an `aria-label` in the
  app, which helps screen readers too. Use `data-testid` only where no name
  tells two elements apart.
- **Never sign in to a real account with a wrong password:** failed attempts
  block it for 15 minutes. Use an unknown username.
- **Tests that deploy** - apps, functions, templates - create swarm services.
  Run them only on a cluster that is there to be thrown away.

## With an AI assistant

Playwright MCP lets Claude Code drive a browser by its accessibility tree, to
explore a screen and to write or fix a test:

```bash
claude mcp add playwright -- npx @playwright/mcp@latest --isolated --allowed-origins "http://localhost:10000"
```

`npx playwright init-agents --loop=claude`, run here, adds Playwright's planner,
generator and healer agents. Review every assertion the healer changes: it can
make a test pass by agreeing with a bug.
