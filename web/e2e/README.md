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

`env/up.sh` runs postgres and redis as containers, and the backend with its
agent inside a dind container that is a swarm of its own - never on this
machine's swarm, which it checks before it says the stack is up. It needs the
backend repo beside this one (or `HP_BACKEND_DIR`), Go, and the backend's
`hivepaas-devtools` image (`make init` there, once). `HP_E2E_PORT` moves it off
10100; `HP_E2E_SKIP_BUILD=1` reuses the last build. The template catalog is the
app-templates repo beside this one (or `HP_TEMPLATES_SRC`); without it, the
template test finds none. dind is also given kopia, the backup engine, from
its pinned image.

### On the local backend (level 0)

Against the local backend - see the backend repo's `docs/DEVELOPMENT.md`:

```bash
make local-deploy           # once: the cluster, and the seeded admin / abc123
make local-build-dashboard  # the dashboard the backend serves
make local-app-run          # http://localhost:10000
```

Then `yarn test` here, or `npm run e2e` from the dashboard's root. `yarn test:ui`
opens Playwright's UI mode; `yarn report` shows the last run, with a trace of
each test that failed.

To run against another installation:

| Variable          | Default                          |
| ----------------- | -------------------------------- |
| `HP_E2E_BASE_URL` | `http://localhost:10000`         |
| `HP_E2E_USERNAME` | `admin`                          |
| `HP_E2E_PASSWORD` | `abc123`                         |
| `HP_E2E_RUN_ID`   | when the run started, in base 36 |

Point them only at an installation whose data does not matter: the tests make
what they need there, and remove it. A read-only account, as on the demo
servers, is enough for the sign-in tests alone.

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
