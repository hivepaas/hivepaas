# Functions

A function is code that answers a request: a handler, not a server. Its author
writes the handler, picks a language, and gets an HTTP endpoint and a job that
can run on a schedule, without a Dockerfile, a port or a health check.

HivePaaS already runs apps. A function is one: it is built, deployed, routed,
configured and watched the way an app is. What a function adds is the part an
app's author writes themselves - a server around the code - and the things that
only make sense when the unit is one call rather than one process: a timeout per
call, a log line per call, a test run of one call.

This design covers the first version. It is built so that what comes after it -
scale to zero, asynchronous calls, other triggers, other languages - is added
next to it rather than in place of it.

---

## Decisions

1. **A function is an app.** Its kind settings say category `function`; its
   deployment method is `function`. Everything an app has, a function has.
2. **HivePaaS writes its own runtimes**, in the repository
   `hivepaas/function-runtimes`: one per language, on Debian slim (Go compiles
   in the full `golang` image, and runs on slim).
3. **The contract between a handler and its runtime is versioned.** A function
   records its runtime (`node24`) and its contract (`v1`). A later contract runs
   beside this one; nothing moves a function to it but its author.
4. **The first runtimes are Node.js 24, Python 3.13 and Go 1.27**: the current
   long-term or stable line of each.
5. **Code comes from the dashboard's editor or from a repository.** Libraries
   are declared the language's own way (`package.json`, `requirements.txt`,
   `go.mod`). Debian packages can be added.
6. **A deployment is a build through the existing pipeline**: HivePaaS writes
   the Dockerfile, and the build runs where any build runs, on another node
   included, with build secrets as BuildKit secrets.
7. **A test run calls the handler once**, with code not yet saved, in a
   throwaway container on a build node, with the function's own variables and
   secrets.
8. **Calls are synchronous.** HTTP through the app's routing; a schedule through
   a new kind of scheduled job.
9. **A function always runs at least one instance.** Scaling to zero is for
   later.

## 1. The model

### A function is an app

- **Identity:** the app's kind settings (`AppKindSettings`) gain the category
  `function`. It is set when the function is created and cannot be changed: a
  function does not become a web app, or the reverse.
- **What is deployed:** the app's deployment settings gain the method
  `function` and its source:

```go
type DeploymentFunctionSource struct {
    // Runtime is the language and its line: node24, python313, go127.
    Runtime string `json:"runtime"`
    // Contract is the version of the handler contract: v1.
    Contract string `json:"contract"`
    // Entrypoint is the file and the name of the handler in it.
    Entrypoint FunctionEntrypoint `json:"entrypoint"`
    // Code is inline (the dashboard's editor) or a repository.
    Code FunctionCode `json:"code"`
    // SystemPackages are Debian packages installed before the libraries:
    // "libpq-dev", or "ffmpeg=7:5.1.6-0+deb12u1".
    SystemPackages []string `json:"systemPackages,omitempty"`
    // Limits of one call.
    Timeout        timeutil.Duration `json:"timeout"`        // default 30s, at most 15m
    MaxConcurrency int               `json:"maxConcurrency"` // calls at once per instance, default 16
    MaxBodySize    unit.DataSize     `json:"maxBodySize"`    // request and response, default 6 MB
}

type FunctionCode struct {
    Inline *FunctionInlineCode   `json:"inline,omitempty"`
    Repo   *DeploymentRepoSource `json:"repo,omitempty"`
    // Dir is where the function is in the repository, for a repository holding
    // several.
    Dir string `json:"dir,omitempty"`
}

type FunctionInlineCode struct {
    Files []FunctionFile `json:"files"` // path relative to the function's root, and content
}
```

- **Inline code is kept in the setting**, so a deployment's snapshot of its
  settings holds the exact code it ran, and redeploying an old deployment rolls
  the code back with it. Its limits: 1 MB in all, 100 files, paths relative and
  without `..`.
- **Everything about a function is in its deployment settings**, limits
  included: they reach the running function as configuration of its container,
  so they take effect on a deployment and are recorded with it.

### What is fixed for a function

- **The container port** is the runtime's (8080). The routing settings keep
  their domains; the dashboard does not ask for a port.
- **The command** is the runtime's. The deployment settings' command and
  working directory do not apply.
- **The health check** is the image's own (`HEALTHCHECK` in the runtime image),
  which Swarm already waits for during a rolling update.
- **Stopping:** on `SIGTERM` the runtime lets running calls finish, for at most
  one timeout; the service's stop grace period is set above the function's
  timeout so that Swarm does not kill them first.

## 2. The runtime contract, v1

The repository `function-runtimes` holds the contract (`CONTRACT.md`) as its
reference. In short:

### The image

- `ghcr.io/hivepaas/function-runtime-<runtime>:<version>`, for example
  `function-runtime-node24:1.0.0`. The version's major is the contract's: every
  `1.x.y` speaks v1.
- Debian slim, the runtime in `/hivepaas`, the function in `/app` with its
  libraries (`node_modules`, the virtual environment `.venv`), a user of its own
  (`hivepaas`, uid 10001): the function never runs as root.
- Go has two images: `function-runtime-go127-build` compiles a function with the
  runtime into one binary, and `function-runtime-go127`, without the toolchain,
  runs it.

### Two ways to run it

- **Serve:** `hivepaas-runtime serve` answers HTTP on port 8080. Paths under
  `/_hivepaas/` are the runtime's: `/_hivepaas/health` for the health check.
- **Invoke:** `hivepaas-runtime invoke` reads one request, as JSON, on its
  standard input, calls the handler once, writes the call's log and then the
  result, as JSON on a last line behind a marker, on its standard output, and
  exits; errors go to standard error. A test run and a scheduled call use this;
  neither needs a port or a network.
- **Deps:** `hivepaas-runtime deps` installs the function's libraries from its
  manifest, making the lock file when there is none; **build** (Go) compiles.

### The handler

```js
// Node.js: index.js
export default async function handler(req, ctx) {
  return { status: 200, body: { hello: req.query.name } }
}
```

```python
# Python: main.py (def or async def)
def handler(req, ctx):
    return {"status": 200, "body": {"hello": req.query.get("name")}}
```

```go
// Go: package function
func Handle(ctx context.Context, req *hivepaas.Request) (*hivepaas.Response, error) {
    return &hivepaas.Response{Status: 200, Body: []byte("hello")}, nil
}
```

The Go types are the package `github.com/hivepaas/function-runtimes/hivepaas`,
of the module at the repository's root, so that a tag is both the images'
version and the module's.

- **The request:** method, path, query (a name to its values), headers, the body
  as bytes, with helpers to read it as text or JSON.
- **The context:** the request's id, its deadline, a logger whose lines carry
  the id.
- **The response:** status (200 when not given), headers, body. A body that is
  an object or a list is sent as JSON.
- **What the runtime answers on its own:**
  - the handler raised an error: `500`, with the request's id and nothing of the
    error, which is in the log;
  - the call outlived its timeout: `504`;
  - all the instance's concurrent calls are taken: `429`, with `Retry-After`;
  - the request or the response is larger than the limit: `413`, or `500`.
- **Configuration**, from the container's environment: the entrypoint, the
  handler's name, the timeout, the concurrency and the body size, as
  `HP_FN_*` variables HivePaaS sets from the deployment settings. The function's
  own variables are in the same environment.
- **A log line per call**, on standard output, as JSON: the request's id,
  method, path, status, duration, and whether it timed out or failed.

## 3. Building a function

1. **The source** is written into the build's checkout directory: the inline
   files, or the repository checked out as for an app (its directory, if set).
2. **HivePaaS writes the Dockerfile**, given to the build as a manual one:

```dockerfile
FROM ghcr.io/hivepaas/function-runtime-node24:1.0.0@sha256:… AS base
# only when the function names Debian packages; the image runs as hivepaas
USER root
RUN apt-get update && apt-get install -y --no-install-recommends libpq-dev \
    && rm -rf /var/lib/apt/lists/*
USER hivepaas

FROM base AS deps
COPY --chown=hivepaas:hivepaas package.json package-lock.json /app/
RUN --mount=type=secret,id=NPM_TOKEN,env=NPM_TOKEN hivepaas-runtime deps

FROM deps
COPY --chown=hivepaas:hivepaas . /app/
```

   - Each layer is cached by what it is built from: the runtime image, then the
     package list, then the manifest and its lock file. A change to the code
     alone rebuilds only the last layer.
   - **The install is the runtime's `deps`**: `npm ci`, or `npm install` that
     makes the lock file; `uv pip compile` then `uv pip sync` into the virtual
     environment `/app/.venv`; `go mod download`.
   - **Go** has two more stages: the build image runs `hivepaas-runtime build`,
     which tidies `go.mod` and `go.sum` and compiles the function with the
     runtime into one binary; the last stage is the slim `function-runtime-go127`
     with the Debian packages and `/app` copied from the build.
   - **Every build variable that uses a secret is mounted into the install
     step**, so a private npm or Python registry works, the way the build-secret
     change made every build read secrets.
   - **A Debian package name is checked** before it reaches the Dockerfile:
     letters, digits, `+`, `-`, `.`, and an optional `=version`. Anything else is
     refused.
3. **The lock file:** a build that finds a manifest without its lock file
   resolves the versions and makes one (`package-lock.json`,
   `requirements.lock`; for Go, `go.mod` and `go.sum`, which `build` tidies from
   the code's imports). For inline code, HivePaaS saves it with the
   function's files, so the next build installs the same versions. For a
   repository, the build log says that the lock file is missing.
4. **Then the build and the deployment of any app**: buildx, on the build node,
   the push to the registry if one is set, the service updated.

### Which runtime image

The runtime images are released by HivePaaS, like its system images: they are
built in `function-runtimes`, and `release.json` names each by its digest, in a
map of their own:

```json
"functionRuntimes": {
  "node24": "ghcr.io/hivepaas/function-runtime-node24:1.0.0@sha256:…",
  "python313": "ghcr.io/hivepaas/function-runtime-python313:1.0.0@sha256:…",
  "go127": "ghcr.io/hivepaas/function-runtime-go127:1.0.0@sha256:…",
  "go127-build": "ghcr.io/hivepaas/function-runtime-go127-build:1.0.0@sha256:…"
}
```

`make release-pin` pins them with the other images. A function built under a
release uses the runtime image that release names; a running function keeps the
image it was built with until it is deployed again.

## 4. A test run

```
dashboard ── code (not saved) and a request ──▶ app
app: the function's variables and secrets, resolved as for a build
app ── code and request ──▶ agent of a build node
agent: the libraries' image, built once per manifest and lock file
       a throwaway container from it: code copied in, `invoke`, the request on stdin
       the response and the logs; the container removed
app ──▶ dashboard: status, headers, body, logs, duration, a lock file if one was made
```

- **The API:** a call on the function, taking files and a request (method, path,
  query, headers, body). Only someone who can change the function can run it:
  it runs the code with the function's secrets.
- **Where:** a build node, chosen as for a build; the code travels in the call,
  as a build's source does.
- **The libraries' image** is built on that node with buildx and kept,
  named after a hash of the runtime image, the Debian packages, the manifest and
  the lock file. A test run with the same libraries does not install them again.
  Node cleanup removes these images with the build cache.
- **The container** is created through Docker's API, not its command line, so
  the variables are not on a command line: no published port, the function's
  memory and CPU limits, stopped when the call outlives its timeout and removed
  in every case.
- **What comes back:** the response, its body cut at 1 MB for display; the
  handler's logs; the duration; a lock file the libraries' install made, which
  the editor adds to the code for the author to save.
- **How long it takes** (estimated from a container measured on a laptop, not
  yet on a server): about 0.5 to 1.5 seconds for Node.js or Python when the
  libraries are already installed; for Go, a container that tidies, compiles and
  calls took about 0.4 seconds, the standard library and the runtime being
  compiled in the build image already. The first run after the libraries change
  adds their install.
- **The code copied into the container belongs to uid 10001**: `go mod tidy` and
  `npm` write into `/app`.

## 5. Calling a function

- **HTTP:** the app's routing, its domains, its Basic Auth and Key Auth: a
  request reaches the runtime's server through Traefik like any app.
- **On a schedule:** a new scheduled job type, `function-invoke`. Its settings
  are the request: method, path, headers, body. It runs `hivepaas-runtime invoke`
  in a running task of the function, through the container exec the
  `container-command` jobs use; the response is the job's output, and a status
  of 400 or more fails the job.
- **From another app of the project:** by the service's name on the project's
  network, as any app reaches another.

## 6. The dashboard

- **Create a function:** a name, a runtime, a starting template per runtime, and
  the code's source: the editor or a repository.
- **The function's page:**
  - **Code:** the editor with several files (the handler, the manifest, other
    files), with syntax for JavaScript, Python, Go and JSON, beside the bash and
    Dockerfile it has.
  - **Test:** a request (method, path, query, headers, body), a Run button, and
    the response, the logs and the duration.
  - **Settings:** runtime, entrypoint, timeout, concurrency, body size, Debian
    packages.
  - The app's own pages: deployments, logs, variables and secrets, domains,
    scheduled jobs.
- **App lists** show a function with a `function` badge, and can filter on it.

## 7. Room to grow

Each of these is added next to what is here, with a default that keeps today's
behaviour:

| Later | How it fits |
|---|---|
| Another language or line (Python 3.14, Bun) | a new runtime id and image; existing functions keep theirs |
| A contract v2 | a new major of the runtime images; a function moves to it only when its author does |
| Scale to zero | `MinInstances`, defaulting to 1 |
| Asynchronous calls | an invocation mode, defaulting to synchronous |
| Other triggers (an event, a webhook with a queue) | new trigger types beside the scheduled job |
| Metrics (calls, errors, p95) | read from the log line every call already writes |
| TypeScript | Node.js 24 strips types itself; the runtime can load `.ts` files |

## 8. What goes where

- **`hivepaas/function-runtimes`** (Apache-2.0, public):
  - `CONTRACT.md`, the contract, by version;
  - one directory per runtime: its server and invoke modes, its Dockerfile, its
    tests;
  - the Go module of the types a Go handler uses;
  - a conformance suite: the same cases run against every runtime image, through
    `serve` and through `invoke`;
  - CI: the suite on every pull request; on a tag, the images built for amd64
    and arm64 and pushed to GHCR.
- **`hivepaas`**: the model, the generated Dockerfile, the test run (API and
  agent), the scheduled job type, the runtime images in `release.json`.
- **`hivepaas-dashboard`**: creating a function, the editor, the test panel, the
  settings, the badges.

## 9. In five parts

Each part has its own plan and its own review:

1. **The runtimes**, in `function-runtimes`: the contract, three runtimes, the
   conformance suite, the images. Usable with `docker run` before HivePaaS knows
   about them.
2. **Functions in the backend:** the model, the generated Dockerfile, the
   deployment, the routing.
3. **Test runs:** the API, the agent's call, the libraries' image.
4. **The dashboard.**
5. **Scheduled calls:** the `function-invoke` job type.

## 10. Testing

- **The runtimes:** the conformance suite against each image, in both modes: a
  JSON and a text response, a raised error, a timeout, the concurrency limit, a
  body over the limit, the health check, the per-call log line, a library used
  by a handler. A Debian package is installed by the generated Dockerfile and is
  checked with it, in part 2.
- **The backend:** the generated Dockerfile for each runtime, with and without
  packages, libraries, a lock file and build secrets; package names refused;
  inline code limits; a function's fixed port and command; the test run's
  container settings (no published port, limits, removal) against a fake Docker
  API; the job type's request and its failure on a 400.
- **Live, on the Linux server:** a function per runtime created in the editor,
  test-run, deployed, called over HTTP and on a schedule; one from a repository
  with a private library; one with a Debian package.

## Later

- Scale to zero, after measuring how long Swarm takes to start a service from
  nothing on the server.
- Asynchronous calls with a queue and retries.
- Triggers other than HTTP and schedules.
- Metrics in the dashboard.
- Runtimes a user brings.

## Changes after part 1's plan

Writing the runtimes (`docs/superpowers/plans/2026-10-01-function-runtimes.md`)
settled what the sections above now say:

- the Go module is the repository's root, not `go/hivepaas`;
- Go has a build image and a slim run image, and `release.json` names both;
- every runtime has `deps`, which the generated Dockerfile and the test run's
  libraries image call instead of the package manager;
- invoke writes its result on standard output as a marked last line, after the
  call's log, instead of keeping the log on standard error;
- the images run as `hivepaas`, so the Dockerfile switches to root for Debian
  packages;
- a stopping instance finishes its calls, so the stop grace period follows the
  timeout;
- a Go test run takes well under a second, not 2 to 5;
- the details of the contract are in `CONTRACT.md` of `function-runtimes`.

## Changes after part 2's plan

Writing the backend's part (`docs/superpowers/plans/2026-10-01-functions-backend.md`)
settled what follows; the sections above are read with it:

- the lock file of inline code is not saved back after a deployment's build:
  part 3's test run returns it and the editor adds it to the code; until then a
  build without one says so in its log;
- `FunctionCode.Repo` is a repository of its own type - type, URL, ref, commit,
  options, credentials - without a repository source's Dockerfile and registry;
- the function's source names the registry its image is pushed to, which a
  cluster of several nodes needs, as a repository's does;
- the `HP_FN_*` limits are the image's `ENV`, written by the Dockerfile;
- a function is created by `POST /projects/{project}/{env}/apps/function`;
- the container settings keep what is fixed for a function, as a deployment
  does; a clone that takes a function's deployment settings is a function;
- a function is not part of a spec yet: export leaves it out, and a template or
  an import refuses it.

## Changes after part 3's plan

Writing the test runs (`docs/superpowers/plans/2026-10-01-function-test-runs.md`)
settled what follows:

- a test run takes the code, as files, and a request; the runtime, entrypoint,
  Debian packages and limits are the function's saved settings;
- its environment, CPU and memory limits are the function's service's, and its
  container joins the project's network, so it reaches the project's apps;
- the call is synchronous, bounded to 15 minutes; the install's log comes back
  with the answer rather than as it happens;
- the libraries' image is removed by the cluster cleanup's image prune, as any
  image no container uses;
- a test run does not wait for a build slot: with every build node busy it
  fails;
- outcomes beyond the runtime's: the install failed, the handler could not be
  loaded, the request could not be read, the container was killed past the
  timeout, or it ended without a result.

## Changes after part 4's plan

Writing the dashboard (`docs/superpowers/plans/2026-10-01-function-dashboard.md`)
settled what follows:

- an app answers its category, and the app list filters on it
  (`category=function`, or `category=webapp,database`; an app without a kind is
  a webapp): the badge and the filter need no other call;
- the test panel is inside the Code tab, beside the editor, and runs the
  editor's files; the Code tab is a function's first;
- the function's settings take the place of its deployment settings, under the
  name "Function", with the code's place (the editor or a repository) and the
  registry its image is pushed to; "App Kind" is not shown for a function;
- creating a function asks a name, an env, a runtime and the code's source (the
  runtime's template or a repository); its entrypoint and limits are the
  defaults until its settings change them;
- the lock file a test run made is added to the editor's files with one button,
  and saved with the code.

## Changes after part 5's plan

Writing the scheduled calls (`docs/superpowers/plans/2026-10-01-function-invoke-job.md`)
settled what follows:

- a function's call lives in the function, at the app scope only, and only a
  function has one; an env's job sequence runs it as a step, as any app job;
- its request is a method, a path with its query, headers and a text body; a
  GET or HEAD sends no body;
- the call's log streams into the run's log; the response is kept in the run,
  its body cut at 64 KB, even when its status fails the run;
- an exit of invoke without a result, an outcome other than ok, or a status of
  400 or more fails the run;
- as a step, it hands on `STATUS` and, when the body is text, `BODY` cut at
  16 KB;
- no trigger runs a call yet, and MCP does not plan one.
