# Function Runtimes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Part 1 of functions. The new repository `hivepaas/function-runtimes` holds the v1 contract, three runtimes that turn a handler into a server or into one call (Node.js 24, Python 3.13, Go 1.27), a conformance suite every image passes, and the CI that publishes the images to GHCR. HivePaaS does not use them yet; `docker run` does.

**Architecture:**
- **The contract** (`CONTRACT.md`) says what a handler is given and what an image promises: user 10001, port 8080, `/app`, the `HP_FN_*` limits, the answers the runtime makes itself, a JSON log line per call, `invoke`'s marked result line, the exit codes.
- **Go:** the package `hivepaas` (what a handler is written against) and `hivepaas/serve` (the runtime). The image `function-runtime-go127-build` compiles a function with the runtime into one binary; the slim `function-runtime-go127` runs that binary.
- **Node.js and Python:** the runtime is code in the image, under `/hivepaas/runtime`, that loads the handler from `/app` when it starts. `deps` installs the function's libraries: npm into `/app/node_modules`, uv into the virtual environment `/app/.venv`.
- **Conformance:** a Go test suite behind the `conformance` build tag. For each runtime it builds the image and a fixture function on it, then calls the fixture over HTTP and through `invoke`.
- **CI:** on pull requests, the unit tests and the suite on amd64 and arm64. On a tag, per-architecture images pushed to GHCR, merged into multi-architecture ones, their digests in a draft GitHub Release.

**Tech Stack:** Go 1.27 (standard library only), Node.js 24 (`node:http`, `node:test`), Python 3.13 (`http.server`, `asyncio`, `unittest`), uv 0.9, Docker BuildKit, GitHub Actions.

**Spec:** `docs/superpowers/specs/2026-10-01-functions-design.md` (backend), part 1 of section 9, amended by this plan's commit (its last section lists what changed). The contract in full is `CONTRACT.md`, written in Task 1.

## How the patches work

Every task's code was written and verified before this plan.

**Repository:** `hivepaas/function-runtimes`, checked out next to the backend at `/Users/tnt/go/src/github.com/hivepaas/function-runtimes`. **Base:** its `main` at `caebb7e` ("Initial commit": LICENSE and README). The tasks commit on a branch of that repository; the backend changes only by this plan's commit.

Patches live in the backend, in `docs/superpowers/plans/2026-10-01-function-runtimes/`: `rt-NN-tests.patch` then `rt-NN-code.patch` for task NN. Task 6 has only `rt-06-code.patch`. The commands below run from the function-runtimes root (or a worktree of it), with `P=/Users/tnt/go/src/github.com/hivepaas/hivepaas/docs/superpowers/plans/2026-10-01-function-runtimes`.

The branch the patches were cut from is kept as `prep/runtimes-v1`. After the last task, `git diff prep/runtimes-v1` is empty.

**What was checked, on a fresh worktree of `caebb7e`, at every task:**
- the tests alone fail, as each task's step 2 says;
- with the code, they pass;
- `golangci-lint run` finds 0 issues, and shellcheck passes on the shell commands.

**After the last task:** the whole conformance suite passes (94 checks over the three runtimes, about a minute once the base images are pulled); actionlint finds nothing in the workflows; the CI's own lint command finds nothing; the tree equals `prep/runtimes-v1`.

**What it needs on the machine:**
- Go 1.27, golangci-lint 2.13, and Docker with its default builder: an image `docker build` makes must be one `docker run` can start.
- Node.js and Python are not needed on the host: their tests run in their base images.
- The first run pulls `golang:1.27-trixie` (1.3 GB), `debian:trixie-slim`, `node:24-trixie-slim`, `python:3.13-slim-trixie` and `ghcr.io/astral-sh/uv:0.9`; the checks pull `koalaman/shellcheck:stable`, `rhysd/actionlint:latest` and `golangci/golangci-lint:v2.13.0`.
- The suite removes the `hp-conformance/*` images it builds (`KEEP_IMAGES=1` keeps them); `docker builder prune` frees the build cache it leaves.

## Global Constraints

- **The images:** `ghcr.io/hivepaas/function-runtime-<runtime>:<version>`, runtimes `node24`, `python313`, `go127`, and `go127-build` for Go's compiler. The version's major is the contract's: every `1.x.y` keeps v1.
- **In every image:** Debian 13 (trixie); the user `hivepaas`, uid and gid 10001, as the image's `USER`; the function in `/app`, the working directory; the runtime in `/hivepaas`; the command `hivepaas-runtime` on the `PATH`; `EXPOSE 8080`; `HEALTHCHECK --interval=10s --timeout=3s --start-period=30s --start-interval=1s --retries=3 CMD ["hivepaas-runtime", "health"]`; `CMD ["hivepaas-runtime", "serve"]`.
- **Configuration:** `HP_FN_ENTRYPOINT` (`index.js`, `main.py`, `.`), `HP_FN_HANDLER` (`default`, `handler`, `Handle`), `HP_FN_TIMEOUT_MS` 30000, `HP_FN_MAX_CONCURRENCY` 16, `HP_FN_MAX_BODY_SIZE` 6291456, `HP_FN_PORT` 8080. A limit that is not a whole number from 1 to 2147483647 stops the runtime, exit 2.
- **What the runtime answers itself:** 500 `internal error`, 504 `timeout`, 429 `too many requests` with `Retry-After: 1`, 413 `request too large`, each as `{"error": …, "requestId": …}`; `/_hivepaas/health` 200 `{"status":"ok"}`, or 503 `{"status":"unavailable"}` when the handler could not be loaded; another `/_hivepaas/` path 404 `{"error":"not found"}`.
- **The call's id:** the request's `X-Request-Id` when it matches `^[A-Za-z0-9._-]{1,128}$`, else 32 hexadecimal characters; every response carries it.
- **Logs:** compact JSON lines on standard output, `{"hp":"invocation",…}` per call and `{"hp":"log",…}` per line the handler logs; errors on standard error, starting `hivepaas: call <id> failed: ` or `hivepaas: the handler could not be loaded: `.
- **Invoke:** the result is the last line of standard output, `#hivepaas-result {…}`. Exit 0 with a result; 2 when the request cannot be read; 3 when the handler cannot be loaded or a Go function cannot be built.
- **Idle connections** are kept 5 minutes.
- **The Go module** is `github.com/hivepaas/function-runtimes`, `go 1.27`, standard library only. The linter configuration mirrors the backend's: 120 columns, US spelling.
- **The fixtures build without a network** (`--network none`): what they install is in them.
- **Git:** commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`; merge locally, never push; never `git stash`.

## Review Focus

Inputs the tests do not reach, or reach only in part:

1. **A Go function that imports a module from the internet.** `build` tidies `go.mod` with the network, then compiles in a workspace that replaces the runtime module with the image's copy. The suite has no network.
   - Pinned by: the fixture's `example.com/fixturelib`, through a `replace` of the function's own (Task 3). Checked by hand with `github.com/google/uuid` while preparing this plan; part 3's live check runs it on the server.
2. **A large request body over the limit.** The runtime answers 413 and closes the connection without reading the rest. A client still sending a body much larger than the limit may see the connection reset instead of the 413, as with Go's own server past 256 KiB.
   - Pinned by: `bodies over the limit` and `a body sent in chunks` (Task 3), for bodies of a few KiB.
3. **A Node.js handler that blocks the event loop**, with a loop that never awaits: its timeout cannot fire, and the instance's other calls wait. That is Node.js; nothing tests it.
4. **Handlers that keep running past their timeout.** Each keeps its place; with `HP_FN_MAX_CONCURRENCY` of them the instance answers 429 until one returns.
   - Pinned by: `TestACallThatTimesOutIsAnsweredAndStillCounted` (Task 2), `a call that times out keeps its place until the handler returns` (Task 4), `test_a_function_that_times_out_keeps_its_place_until_it_returns` (Task 5).
5. **Code a test run copies into a container** (part 3) must belong to uid 10001, since `go mod tidy` and `npm` write into `/app`; and a tar made on macOS carries extended attributes Docker refuses.
   - Pinned by: nothing here. Part 3's agent writes the tar.

## Decisions beyond the spec

- **The Go module is the repository's root** (`github.com/hivepaas/function-runtimes`, packages `hivepaas` and `hivepaas/serve`), not `go/hivepaas`: one tag is then the images' version and the module's.
- **Go has two images.** `function-runtime-go127-build` is the full `golang` image with the runtime module and a build cache warmed by one build; `function-runtime-go127` is Debian slim with CA certificates and time zones. A Go function's Dockerfile (part 2) compiles in the first and runs in the second; `release.json` names both.
- **`hivepaas-runtime deps`** installs the libraries in every runtime, and makes the lock file when there is none: `npm ci`, or `npm install`; `uv pip compile` then `uv pip sync`; `go mod download`. Part 2's Dockerfile and part 3's libraries image call it instead of the package manager.
- **Go's `build` tidies `go.mod` and `go.sum`:** the code's imports are added, and the runtime module is required at the image's version. Those two files are a Go function's lock files.
- **Invoke writes the call's log on standard output, before the result** (the spec said standard error): a handler's own prints go there anyway, so the result is a marked last line, and nothing is written after it.
- **The request:** the path as sent, still percent-encoded; the method in upper case; `host` among the headers. A status must be from 200 to 599. A request whose body does not arrive is not a call: the connection is closed, nothing is logged. HEAD, 204 and 304 send no body.
- **Python** takes a function or a coroutine. Coroutines share one event loop and are cancelled at the timeout; functions run in a pool of `HP_FN_MAX_CONCURRENCY` threads and cannot be stopped. Nagle's algorithm is off: with it, a warm call took 40 ms.
- **Node.js:** `ctx.signal` aborts at the timeout. A module may be ES or CommonJS, transpiled `exports.default` included. A `Headers` object is taken as the response's headers.
- **The images run as `hivepaas`:** part 2's Dockerfile switches to root for Debian packages, then back.
- **`serve` lets running calls finish on SIGTERM, for at most one timeout:** part 2 sets the service's stop grace period above the function's timeout.
- **Measured here:** a Go test run (a container that tidies, compiles and calls) took about 0.4 s on a laptop, against the spec's estimate of 2 to 5 s.
- **Not in this part:** a Debian package used by a handler is checked with part 2's Dockerfile, which installs it.

---

## Task 1: The contract, and the Go package a handler is written against

**Files:**
- Create: `go.mod`, `.golangci.yaml`, `.gitignore`, `CONTRACT.md`, `hivepaas/hivepaas.go`
- Modify: `README.md`
- Test: `hivepaas/hivepaas_test.go`

**Interfaces:**
- Produces, in package `github.com/hivepaas/function-runtimes/hivepaas`:
  - `type Request struct { Method, Path string; Query url.Values; Headers http.Header; Body []byte }`, with `Text() string` and `JSON(v any) error`;
  - `type Response struct { Status int; Headers http.Header; Body []byte }`: a zero status is 200, a nil response is 204;
  - `func JSON(status int, v any) (*Response, error)`, `func Text(status int, text string) *Response`;
  - `type HandlerFunc func(ctx context.Context, req *Request) (*Response, error)`;
  - `func WithInvocation(ctx context.Context, requestID string, logFunc func(msg string)) context.Context`, `func RequestID(ctx context.Context) string`, `func Log(ctx context.Context, format string, args ...any)`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/rt-01-tests.patch`

The patch also creates `go.mod`. Tests added: `TestARequestBodyReadsAsTextAndAsJSON`, `TestABodyThatIsNotJSONSaysSo`, `TestQueryAndHeadersAreTheStandardLibrarysOwn`, `TestAJSONResponseCarriesItsContentType`, `TestAValueThatCannotBeJSONIsAnError`, `TestATextResponseIsUTF8PlainText`, `TestTheCallsIDAndLoggerTravelInItsContext`, `TestLoggingOutsideACallDoesNotPanic`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas/`
Expected: FAIL, `no non-test Go files in …/hivepaas`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/rt-01-code.patch`

It brings `CONTRACT.md`, the contract this plan's runtimes keep, and the README.

- [ ] **Step 4: Run them to watch them pass**

Run: `go test -count=1 ./hivepaas/ && golangci-lint run ./...`
Expected: `ok`; `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat: the v1 contract, and the Go package a handler is written against

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 2: The Go runtime: serve and invoke under the contract's limits

**Files:**
- Create: `hivepaas/serve/config.go`, `hivepaas/serve/serve.go`, `hivepaas/serve/invoke.go`, `hivepaas/serve/main.go`
- Test: `hivepaas/serve/config_test.go`, `hivepaas/serve/serve_test.go`, `hivepaas/serve/invoke_test.go`

**Interfaces:**
- Consumes: Task 1's package `hivepaas`.
- Produces, in package `github.com/hivepaas/function-runtimes/hivepaas/serve`:
  - `type Config struct { Port int; Timeout time.Duration; MaxConcurrency int; MaxBodySize int64 }` and `func LoadConfig(getenv func(string) string) (Config, error)`;
  - `func New(h hivepaas.HandlerFunc, cfg Config, out, errOut io.Writer) *Runtime`, with `ServeHTTP`, `Serve(ctx context.Context, l net.Listener) error` and `Invoke(in io.Reader) int`;
  - `func Health(cfg Config, errOut io.Writer) int`;
  - `func Main(h hivepaas.HandlerFunc)`: `serve`, `invoke` or `health` from the command line, then exits. Task 3's build compiles a `main` that calls it.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/rt-02-tests.patch`

Tests added:
- configuration: `TestTheDefaultsAreTheContracts`, `TestTheEnvironmentSetsTheLimits`, `TestALimitThatIsNotAPositiveNumberIsRefused`;
- serve: `TestTheHandlerSeesTheRequestAsSent`, `TestAResponseIsSentAsMadeWithTheCallsID`, `TestARequestIDIsMadeWhenTheRequestHasNoUsableOne`, `TestNoResponseIsNoContent`, `TestABodyWithoutAContentTypeIsOctetStream`, `TestAHandlersHeadersAreSentWhateverTheirCase`, `TestAStatusThatCarriesNoBodySendsNone`, `TestAHandlerErrorIsAnInternalErrorThatSaysNothing` (an error, a panic, three statuses out of range), `TestACallThatTimesOutIsAnsweredAndStillCounted`, `TestTheHandlersContextEndsAtTheTimeout`, `TestCallsBeyondTheConcurrencyLimitAreTurnedAway`, `TestARequestBodyOverTheLimitIsRefused` (with and without a length), `TestTheWholeBodyReachesTheHandler`, `TestARequestWhoseBodyDoesNotArriveIsDropped`, `TestAConnectionWhoseBodyStallsIsClosedAtTheTimeout`, `TestAResponseBodyOverTheLimitIsAnInternalError`, `TestTheRuntimesPathsNeverReachTheHandler`, `TestAHandlersLogLineCarriesTheCallsID`, `TestStoppingLetsRunningCallsFinish`;
- invoke and the command: `TestInvokeCallsTheHandlerOnceWithTheRequestGiven`, `TestAnEmptyRequestIsAGetOfTheRoot`, `TestAFailedCallIsAResultNotAFailedInvoke`, `TestInvokeTimesOutLikeServe`, `TestARequestThatCannotBeReadExits2`, `TestInvokeRefusesABodyOverTheLimit`, `TestHealthAsksTheServerOnThePort`, `TestAnUnknownCommandIsAUsageError`, `TestALimitThatIsNotAPositiveNumberStopsTheRuntime`, `TestInvokeIsACommand`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas/serve/`
Expected: FAIL, build errors starting `undefined: LoadConfig`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/rt-02-code.patch`

A body that stalls is dropped by hijacking the connection and closing it: a panic with `http.ErrAbortHandler` alone would let the server read the rest of the body first, with no deadline (`TestAConnectionWhoseBodyStallsIsClosedAtTheTimeout` is that case).

- [ ] **Step 4: Run them to watch them pass**

Run: `go test -race -count=1 ./... && golangci-lint run ./...`
Expected: `ok` for `hivepaas` and `hivepaas/serve`; `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat: the Go runtime: serve and invoke under the contract's limits

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 3: The go127 images, and the conformance suite they pass

**Files:**
- Create: `runtimes/go127/Dockerfile` (targets `build` and `run`), `runtimes/go127/hivepaas-runtime`, `runtimes/go127/warmup/go.mod`, `runtimes/go127/warmup/warmup.go`
- Test:
  - the suite: `conformance/conformance_test.go`, `conformance/docker_test.go`, `conformance/serve_test.go`, `conformance/invoke_test.go`;
  - the fixture: `conformance/fixtures/go127/Dockerfile`, `go.mod`, `fixture.go`, `fixturelib/go.mod`, `fixturelib/fixturelib.go`.

**Interfaces:**
- Consumes: Task 2's `serve.Main`.
- Produces:
  - `hivepaas-runtime` for Go: `deps`, `build`, and `serve`, `invoke`, `health` run `/app/.hivepaas/function`, building it first in the build image;
  - the suite's list `runtimes []runtimeDef{name string; compiled bool}`, which Tasks 4 and 5 extend by one line each;
  - what every fixture answers, by path: `/` nothing, `/json?name=`, `/text` ("hello"), `/bytes` (0 1 2 255), `/echo` (method, path, `query` and `queryAll` of `a`, header `X-Custom`, body, requestId), `/status?code=`, `/error`, `/slow?ms=` (past the timeout on purpose), `/big?bytes=`, `/invalid` (status 42), `/env` (`GREETING`), `/deadline` (`remainingMs`), `/log`, `/whoami` (uid, gid), `/lib` ("from lib", from a library), anything else 404;
  - a fixture's Dockerfile takes `ARG RUNTIME` (and `BUILD_RUNTIME`, with a stage `source` holding the function not yet built, for a compiled runtime).

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/rt-03-tests.patch`

The suite, per runtime:
- `image`: user, working directory, port, command and health check of every image;
- `serve`: `a body is sent with the type it has`, `the handler sees the request as sent`, `a status of the handler's own, or no content`, `an error says nothing of itself`, `a status that is not one is an error`, `a call is answered at its timeout`, `bodies over the limit`, `a body sent in chunks`, `a request whose body does not arrive is dropped`, `the function's environment and user`, `every call has an id and a log line`, `a handler's log line carries its call's id`, `the runtime's paths`, `docker sees it healthy`, `calls over the concurrency are turned away`, `stopping lets the running call finish`, and for a runtime that loads its handler, `a handler that cannot be loaded`;
- `invoke`: `the request reaches the handler, the result is the last line`, `an empty request is a GET of the root`, `a failed call is a result`, `a call is answered at its timeout`, `a handler's log line comes before the result`, `a body over the limit`, `a request that cannot be read`, `deps again, from the lock file the build made`, `a handler that cannot be loaded`, and for Go, `a function not built yet is built first`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go vet -tags conformance ./conformance/ && RUNTIMES=go127 go test -tags conformance -count=1 ./conformance/`
Expected: vet passes; the suite FAILS at `docker build … runtimes/go127/Dockerfile`: `lstat …/runtimes/go127: no such file or directory`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/rt-03-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `docker run --rm -v "$PWD:/src:ro" -w /src koalaman/shellcheck:stable runtimes/go127/hivepaas-runtime`
Expected: no output

Run: `RUNTIMES=go127 go test -tags conformance -count=1 ./conformance/ && golangci-lint run --build-tags conformance ./...`
Expected: `ok` (about 30 s once the base images are pulled); `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat: the go127 images, and the conformance suite they pass

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 4: The node24 runtime

**Files:**
- Create: `runtimes/node24/Dockerfile`, `runtimes/node24/runtime/main.mjs`, `runtimes/node24/runtime/runtime.mjs`
- Modify: `conformance/conformance_test.go` (the runtime `node24`)
- Test: `runtimes/node24/test/runtime.test.mjs`; the fixture `conformance/fixtures/node24/Dockerfile`, `package.json`, `index.js`, `fixture-lib/package.json`, `fixture-lib/index.js`

**Interfaces:**
- Consumes: Task 3's suite and fixture paths.
- Produces, in `runtimes/node24/runtime/runtime.mjs`: `loadConfig(env)`, `loadHandler(appDir, cfg)`, `requestOf(method, rawUrl, rawHeaders, body)`, `toAnswer(value, maxBodySize)`, `class Runtime` (`load`, `call`, `listener`, `invoke`), `ConfigError`. `main.mjs` is the command, linked as `/usr/local/bin/hivepaas-runtime`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/rt-04-tests.patch`

Unit tests added: the configuration's defaults, the environment, limits and entrypoints refused; ES module and CommonJS handlers, transpiled ones, missing ones; every kind of body; no content; headers and the content type; bodiless statuses; what is not a response; a body over the limit; the request as sent; a timeout that aborts `ctx.signal`; a timed-out call that keeps its place; invoke's base64.

- [ ] **Step 2: Run them to watch them fail**

Run: `docker run --rm -v "$PWD:/src:ro" -w /src node:24-trixie-slim node --test 'runtimes/node24/test/*.test.mjs'`
Expected: FAIL, `ERR_MODULE_NOT_FOUND` for `runtimes/node24/runtime/runtime.mjs`

Run: `RUNTIMES=node24 go test -tags conformance -count=1 ./conformance/`
Expected: FAIL at `docker build`: the runtime has no Dockerfile

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/rt-04-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `docker run --rm -v "$PWD:/src:ro" -w /src node:24-trixie-slim node --test 'runtimes/node24/test/*.test.mjs'`
Expected: `pass 17`, `fail 0`

Run: `RUNTIMES=node24 go test -tags conformance -count=1 ./conformance/`
Expected: `ok`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat: the node24 runtime

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 5: The python313 runtime

**Files:**
- Create: `runtimes/python313/Dockerfile`, `runtimes/python313/hivepaas-runtime`, `runtimes/python313/runtime/hivepaas_runtime/__init__.py`, `__main__.py`, `runtime.py`, `server.py`
- Modify: `.gitignore`; `conformance/conformance_test.go` (the runtime `python313`); `conformance/serve_test.go` (the check `a warm call is answered without delay`, for every runtime)
- Test: `runtimes/python313/tests/test_runtime.py`; the fixture `conformance/fixtures/python313/Dockerfile`, `requirements.txt`, `main.py`, `lib-src/make_wheel.py`, `lib-src/fixture_lib/__init__.py`

**Interfaces:**
- Consumes: Task 3's suite and fixture paths.
- Produces, in package `hivepaas_runtime.runtime`: `Config`, `load_config(env)`, `load_handler(app_dir, cfg)`, `Request`, `Context`, `request_of(method, raw_path, header_pairs, body)`, `Answer`, `to_answer(value, max_body_size)`, `class Runtime` (`load`, `try_reserve`, `call`, `own_path`, `invoke`, `close`); `hivepaas_runtime.server.serve(runtime)`. The command runs `python -m hivepaas_runtime` with the virtual environment's Python.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/rt-05-tests.patch`

Unit tests added: the configuration; handlers by name, `/app` first on the module path, missing handlers and failing modules; every kind of body; headers; bodiless statuses; what is not a response; a body over the limit; the request as sent; an async handler awaited, cancelled at its timeout, an async callable object too; a function that returns a coroutine; a timed-out function that keeps its place; an error logged with its traceback; invoke's base64.

The fixture packs its library into a wheel before `deps`, standing for one from an index, so that the build needs no network.

- [ ] **Step 2: Run them to watch them fail**

Run: `docker run --rm -v "$PWD:/src:ro" -w /src -e PYTHONPATH=/src/runtimes/python313/runtime -e PYTHONDONTWRITEBYTECODE=1 python:3.13-slim-trixie python -m unittest discover -s runtimes/python313/tests`
Expected: FAIL, `ModuleNotFoundError: No module named 'hivepaas_runtime'`

Run: `RUNTIMES=python313 go test -tags conformance -count=1 ./conformance/`
Expected: FAIL at `docker build`: the runtime has no Dockerfile

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/rt-05-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `docker run --rm -v "$PWD:/src:ro" -w /src koalaman/shellcheck:stable runtimes/python313/hivepaas-runtime`
Expected: no output

Run the unit tests of step 2 again.
Expected: `Ran 21 tests`, `OK`

Run: `RUNTIMES=python313 go test -tags conformance -count=1 ./conformance/ && golangci-lint run --build-tags conformance ./...`
Expected: `ok`; `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat: the python313 runtime

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 6: CI: tests on pull requests, images on tags

**Files:**
- Create: `.github/workflows/test.yml`, `.github/workflows/release.yml`
- Modify: `README.md` (the unit tests, releasing)

**Interfaces:**
- Produces: on a tag `vX.Y.Z`, the images `ghcr.io/hivepaas/function-runtime-{node24,python313,go127,go127-build}:X.Y.Z` (and `:X` for a release that is not a prerelease), and a draft GitHub Release with `digests.txt`, one line per image, `go127=ghcr.io/…:X.Y.Z@sha256:…`, for part 2's `release.json`.

- [ ] **Step 1: Write the workflows**

Run: `git apply --index $P/rt-06-code.patch`

`test.yml`: the Go job (the linter in its image, `go test -race`, shellcheck), the Node.js and Python unit tests in their base images, and the suite per runtime on `ubuntu-24.04` and `ubuntu-24.04-arm`. `release.yml` follows the backend's: the tag checked (and `CONTRACT.md` has its major), the tests and the suite, one build per image and architecture, the manifests, the digests, the draft. Actions are pinned by digest, as in the backend.

- [ ] **Step 2: Check them**

Run: `docker run --rm -v "$PWD:/repo" -w /repo rhysd/actionlint:latest -no-color`
Expected: no output

Run: `docker run --rm -v "$PWD:/src" -w /src golangci/golangci-lint:v2.13.0 golangci-lint run --build-tags conformance ./...`
Expected: `0 issues.`

- [ ] **Step 3: The whole suite, and the tree**

Run: `go test -tags conformance -count=1 ./conformance/`
Expected: `ok` (about a minute)

Run: `git diff --stat prep/runtimes-v1`
Expected: no output

- [ ] **Step 4: Commit**

```bash
git commit -m "ci: tests on pull requests, multi-arch images to GHCR on tags

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## After the plan (the user)

- Merge into `main` of `function-runtimes` and push; the pull request's first CI run checks the workflows on GitHub itself, arm64 included.
- Tag `v1.0.0` when ready: the Release workflow builds and pushes the images and drafts the release with the digests.
- On GHCR, make the four `function-runtime-*` packages public: HivePaaS servers pull them without credentials.
- Part 2 starts from the digests: `functionRuntimes` in `release.json`, the generated Dockerfile with `deps` and Go's two stages, the stop grace period.
