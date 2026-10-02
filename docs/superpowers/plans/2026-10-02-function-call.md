# A Scheduled Call Through the Function's Server Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A `function-invoke` job, and a sequence's step calling a function, hand their request to the `serve` already running in the function's container with `hivepaas-runtime call` - same result line as `invoke`, the call's log in the app's logs, the instance's concurrency respected - falling back to `invoke` on a runtime that has no `call`; the task links its request id to the app's logs.

**Architecture:**
- **The runtimes (Tasks 1-3):** `call` reads `invoke`'s request, refuses what cannot go on an HTTP/1.1 request line or header (exit 2), sends it to `127.0.0.1:HP_FN_PORT` with `X-Hivepaas-Call: 1`, retries a 429 after `Retry-After` while `HP_FN_TIMEOUT_MS` allows, and writes the answer as `invoke`'s result (exit 3 when nothing answers). `serve` adds `X-Hivepaas-Outcome` and `X-Hivepaas-Duration-Ms` to the answer of such a request and hides the header from the handler. `call` loads nothing of the runtime: Node.js and Bun read `wire.mjs` (configuration, `invoke`'s request) and `call.mjs`, `main.mjs` importing the runtime only for the commands that run the handler; Python reads `wire.py` and speaks HTTP/1.1 on a socket (`http.client` brings the `email` package); Go's `call` is a command of the function's binary.
- **HivePaaS (Task 4):** the job runs `hivepaas-runtime call`; on exit 2 - no `call` before 1.2.0, or a request it cannot send - it runs `invoke` with the same request; exit 3 fails the run, "The function's server does not answer on its instance"; the run's log says "The call's log is in the app's logs, request <id>". No release is needed for it.
- **The dashboard (Task 5):** the task's function call shows its request id, linking to the app's Logs with `?search=<id>`, which opens the History tab searched for it.
- **The release (Task 6) and the pins (Task 7)**, as for 1.1.0.

**Tech Stack:** JavaScript (Node.js 24, Bun), Python 3.13 standard library, Go 1.27, Docker, React.

**Spec:** `docs/superpowers/specs/2026-10-02-function-call-through-serve-design.md`, amended by this plan's commit (its last section, "Changes after the plan").

## How the patches work

Tasks 1-5 were written and verified before this plan; Task 6 is the user's; Task 7 writes the release's digests.

| Task | Repository | Base |
|---|---|---|
| 1, 2, 3 | `function-runtimes` | `main` at `6bb2a21` (`v1.1.0`) |
| 4 | `hivepaas` (this one) | `main` at this plan's commit, which adds only documents to `22318e01` |
| 5 | `hivepaas-dashboard` | `main` at `b22c320a` |

Patches live in `docs/superpowers/plans/2026-10-02-function-call/`: `call-NN-tests.patch` then `call-NN-code.patch` for Tasks 1-4, `call-05-code.patch` for the dashboard. With `P=/Users/tnt/go/src/github.com/hivepaas/hivepaas/docs/superpowers/plans/2026-10-02-function-call`, commands run from the worktree's root. A dashboard worktree needs `ln -s /Users/tnt/go/src/github.com/hivepaas/hivepaas-dashboard/node_modules node_modules`.

The branches the patches were cut from are kept as `prep/function-call` in the three repositories. After Task 5, `git diff prep/function-call -- . ':(exclude)docs/superpowers'` is empty in each.

```bash
NT() { docker run --rm -v "$PWD:/src:ro" -w /src node:24-trixie-slim node --test 'runtimes/js/test/*.test.mjs'; }
UT() { docker run --rm -v "$PWD:/src:ro" -w /src -e PYTHONPATH=/src/runtimes/python313/runtime \
  -e PYTHONDONTWRITEBYTECODE=1 python:3.13-slim-trixie python -m unittest discover -s runtimes/python313/tests; }
```

**What was checked, on fresh worktrees of the bases above:**
- Task 1: the tests alone - `NT` fails (`Cannot find module …/runtime/call.mjs`), `RUNTIMES=node24` conformance fails its 5 `call` checks; with the code - `NT` 24 pass, `RUNTIMES=node24,bun1` conformance `ok`.
- Task 2: the tests alone - `UT` fails (`No module named 'hivepaas_runtime.call'`), `python313`'s `call` checks fail; with the code - `UT` OK, `RUNTIMES=python313` conformance `ok`.
- Task 3: the tests alone - `go test ./hivepaas/serve/` does not build (`undefined: Call`, `exitUnavailable`); with the code - `go test ./...` `ok`, `RUNTIMES=go127` conformance `ok`, `golangci-lint run --build-tags conformance ./...` 0 issues; shellcheck clean.
- Task 4: the tests alone - 4 tests of `taskschedjobexec` fail; with the code - `tasks/...` and `functionservice/...` `ok`, `golangci-lint run ./...` 0 issues, the custom lints clean, `make gen-swag` changes nothing.
- Task 5: `npm run lint:ci` and `npm run build` pass.
- Measured on warm instances, 20 calls each, `docker exec` alone 29 ms: Python `invoke` 181 ms, `call` 73 ms; Node.js 58 and 61 ms; Bun 48 and 48 ms.

**What it needs on the machine:** Go 1.27, golangci-lint 2.13, Docker with Docker Hub, `jq`, Node.js and the dashboard's `node_modules`; for Task 6, the user's push rights.

## Global Constraints

- **`call`'s result is `invoke`'s**: the same line, fields and exit codes (0 a result, 2 a request it cannot read or send, 3 no answer from `serve`); nothing else on its standard output.
- **`X-Hivepaas-Call: 1`** marks a call: `serve` adds `X-Hivepaas-Outcome` and `X-Hivepaas-Duration-Ms` to its answer, and the handler does not see the header; every other request is answered as before.
- **429**: `call` waits for `Retry-After` and sends again while `HP_FN_TIMEOUT_MS` from its start allows; then the 429 is its result, outcome `throttled`.
- **`call` loads nothing of the runtime**, so that it starts as fast as its language does.
- **Test runs keep `invoke`.** Contract v1; images `1.2.0`, released by the user.
- **The backend's conventions** and linters; the dashboard's `lint:ci` and `build`.
- **Git:** commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`; merge locally, never push; never `git stash`.

## Review Focus

1. **An instance busy for longer than the timeout**: `call` keeps sending until `HP_FN_TIMEOUT_MS` runs out, then the run fails `throttled`; with several instances, the `exec` lands on one task, which may be the busy one while another is idle.
   - Pinned by: `a call over the concurrency waits for a place` (conformance), and the unit tests of the retry's budget.
2. **A function with several instances**: the call goes to the task the exec picks, and its log lines are in that task's log, which the app's Logs aggregate.
   - Pinned by: nothing; the logs are the app's.
3. **A handler that reads a header named `X-Hivepaas-Call` from its own clients**: it no longer sees it, from anyone.
   - Pinned by: `serve says the outcome to a call, and keeps the header from the handler`.
4. **A function built before 1.2.0**: every scheduled call runs `call` (exit 2, its usage) then `invoke` - two `exec`s, some 30 ms more - until it is built again.
   - Pinned by: `TestAFunctionWithoutCallIsCalledThroughInvoke`.
5. **A job's run log** no longer holds the handler's own log lines; it names the request id, and the task links to the app's Logs searched for it - within the History tab's time window.
   - Pinned by: `TestAFunctionCallSaysWhereItsLogIs`; the link, by the user in the browser.

## Decisions beyond the spec

- **`wire.py` and `wire.mjs`** hold the configuration and `invoke`'s request; the runtimes import them, and `call` imports nothing else of the runtime. Python's `Config` is a `namedtuple` and its number check uses no `re`, so that `wire.py` imports little.
- **Python's `call` speaks HTTP/1.1 on a socket**, `Connection: close`, reading to the end.
- **What cannot be sent exits 2**, in every runtime: a method that is no token, a path not starting with `/` or holding a space or control character, a header name that is no token or a value with CR, LF or NUL.
- **Go: `exitUnavailable` (3)** beside `exitUsage`, and a static `errUnsendable`.
- **The fixtures answer `/headers`** with the names of the headers the handler sees.
- **HivePaaS's change is merged before the release**: older runtimes answer through `invoke`.

---

## Task 1: `call` on Node.js and Bun, and the suite's checks (function-runtimes)

**Files:**
- Create: `runtimes/js/runtime/wire.mjs`, `runtimes/js/runtime/call.mjs`, `runtimes/js/test/call.test.mjs`, `conformance/call_test.go`
- Modify: `runtimes/js/runtime/runtime.mjs` (what moved to `wire.mjs` and `call.mjs`, re-exported; the call header in `serve` and `send`), `runtimes/js/runtime/main.mjs` (`call`; the runtime imported per command), `conformance/conformance_test.go` (the `call` group), the three fixtures (`/headers`), `CONTRACT.md` (`call` in the commands, a section "Call")

**Interfaces:**
- Produces: `callServe(input, cfg, io) -> Promise<number>` in `call.mjs`; `CALL_HEADER`; `wire.mjs`'s `APP_DIR`, `RESULT_MARKER`, `ConfigError`, `loadConfig`, `engine`, `LOCK_FILES`, `DEPS_CACHE`, `depsCommand`, `invokeRequest`; conformance `checkCall(t, imgs)`, `(*container) runCall(t, input)`.

- [ ] **Step 1: Write the failing tests**

```bash
git apply --index $P/call-01-tests.patch
```

- [ ] **Step 2: Run them to watch them fail**

Run: `NT`
Expected: FAIL - `Cannot find module '/src/runtimes/js/runtime/call.mjs'`.

Run: `RUNTIMES=node24 go test -count=1 -tags conformance -run 'TestConformance/node24/call' ./conformance/`
Expected: FAIL - the five `call` checks.

- [ ] **Step 3: Write the code**

```bash
git apply --index $P/call-01-code.patch
```

- [ ] **Step 4: Run them to watch them pass**

Run: `NT; RUNTIMES=node24,bun1 go test -count=1 -tags conformance ./conformance/`
Expected: `ℹ pass 24`, `ℹ fail 0`; `ok`.

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(call): hivepaas-runtime call on Node.js and Bun - a request handed to serve, its answer written as invoke's result; wire.mjs, which call loads alone

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

## Task 2: `call` on Python (function-runtimes)

**Files:**
- Create: `runtimes/python313/runtime/hivepaas_runtime/wire.py`, `runtimes/python313/runtime/hivepaas_runtime/call.py`, `runtimes/python313/tests/test_call.py`
- Modify: `runtimes/python313/runtime/hivepaas_runtime/runtime.py` (what moved to `wire.py`, re-exported), `server.py` (the call header), `__main__.py` (`call`; imports per command), `runtimes/python313/tests/test_server.py` (`_ServerCase`, `CallHeaderTest`)

**Interfaces:**
- Consumes: Task 1's `call` checks of the suite.
- Produces: `call(data, cfg, out, err) -> int`, `CALL_HEADER`; `wire.read_invoke_request(data)`.

- [ ] **Step 1: Write the failing tests**

```bash
git apply --index $P/call-02-tests.patch
```

- [ ] **Step 2: Run them to watch them fail**

Run: `UT`
Expected: FAIL - `No module named 'hivepaas_runtime.call'`, and `CallHeaderTest`.

- [ ] **Step 3: Write the code**

```bash
git apply --index $P/call-02-code.patch
```

- [ ] **Step 4: Run them to watch them pass**

Run: `UT; RUNTIMES=python313 go test -count=1 -tags conformance ./conformance/`
Expected: `OK`; `ok`.

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(python313): hivepaas-runtime call on a socket, loading only wire.py; serve's answer to a call says its outcome and duration

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

## Task 3: `call` on Go (function-runtimes)

**Files:**
- Create: `hivepaas/serve/call.go`, `hivepaas/serve/call_test.go`
- Modify: `hivepaas/serve/main.go` (`call`, `exitUnavailable`), `hivepaas/serve/serve.go` (the call header), `hivepaas/serve/invoke.go` (`readInvokeRequest`), `runtimes/go127/hivepaas-runtime` (`call`)

**Interfaces:**
- Produces: `serve.Call(cfg, in, out, errOut) int`.

- [ ] **Step 1: Write the failing tests**

```bash
git apply --index $P/call-03-tests.patch
```

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas/serve/`
Expected: FAIL to build - `undefined: Call`, `undefined: exitUnavailable`.

- [ ] **Step 3: Write the code**

```bash
git apply --index $P/call-03-code.patch
```

- [ ] **Step 4: Run them to watch them pass, and the rest**

Run:
```bash
go test -count=1 ./...
go test -count=1 -tags conformance ./conformance/
golangci-lint run --build-tags conformance ./...
docker run --rm -v "$PWD:/src:ro" -w /src koalaman/shellcheck:stable \
  runtimes/go127/hivepaas-runtime runtimes/python313/hivepaas-runtime runtimes/bun1/hivepaas-runtime
```
Expected: `ok`; `ok` (all four runtimes); 0 issues; shellcheck silent.

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(go127): hivepaas-runtime call, and serve's answer to a call says its outcome and duration

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

## Task 4: Scheduled calls through `call` (backend)

**Files:**
- Modify: `hivepaas_app/service/functionservice/functioninvoke/functioninvoke.go` (`CallCommand`), `hivepaas_app/tasks/taskschedjobexec/function_invoke.go` (`execInRuntime`, `exitCodeOf`, the fallback, the hints, the log line)
- Test: `hivepaas_app/tasks/taskschedjobexec/function_invoke_test.go`

**Interfaces:**
- Produces: `functioninvoke.CallCommand = []string{"hivepaas-runtime", "call"}`.

- [ ] **Step 1: Write the failing tests**

```bash
git apply --index $P/call-04-tests.patch
```

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas_app/tasks/taskschedjobexec/`
Expected: FAIL - `TestAFunctionCallSendsItsRequestAndKeepsTheResponse`, `TestAFunctionCallSaysWhereItsLogIs`, `TestAFunctionWithoutCallIsCalledThroughInvoke`, `TestAFunctionWhoseServerDoesNotAnswerFailsTheRun`.

- [ ] **Step 3: Write the code**

```bash
git apply --index $P/call-04-code.patch
```

- [ ] **Step 4: Run them to watch them pass, and the linters**

Run:
```bash
go build ./... && go test -count=1 ./hivepaas_app/tasks/... ./hivepaas_app/service/functionservice/...
golangci-lint run ./...
go run ./tools/goroutinelint . && go run ./tools/errcodelint
make gen-swag && git status --short
```
Expected: every package `ok`; 0 issues; clean; only the task's files.

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(functions): a scheduled call goes through the function's serve with hivepaas-runtime call, and through invoke on a runtime before 1.2.0

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

## Task 5: The request id links to the app's logs (dashboard)

**Files:**
- Modify: `src/application/modules/operations/routes/tasks/building-blocks/task-summary-card/function-invoke-response.com.tsx` (the request id, linked), `.../task-summary-card/system-task-summary-card.com.tsx` (the link's address)
- Modify: `src/application/modules/projects/routes/single-project/single-app/tabs/logs/route/app-logs.route.com.tsx` (`?search=` opens History), `.../tabs/logs/building-blocks/app-logs-history/app-logs-history.com.tsx` (`initialSearch`)

- [ ] **Step 1: Write the code**

```bash
git apply --index $P/call-05-code.patch
```

- [ ] **Step 2: Check it**

Run: `npm run lint:ci && npm run build`
Expected: both pass.

- [ ] **Step 3: Commit**

```bash
git commit -m "feat(functions): a function call's task shows its request id, linking to the app's logs searched for it

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

## Task 6: Release function-runtimes 1.2.0 (the user)

- [ ] **Step 1:** Tasks 1-3 merged into function-runtimes' `main` locally, and there `NT`, `UT`, `go test -tags conformance ./conformance/`, `go test ./...` pass.
- [ ] **Step 2:** The user pushes `main`, waits for `Test`, then tags and pushes `v1.2.0`.
- [ ] **Step 3:** The five digests, read from GHCR:

```bash
for n in node24 bun1 python313 go127 go127-build; do
  echo "$n ghcr.io/hivepaas/function-runtime-$n:1.2.0@$(docker buildx imagetools inspect \
    ghcr.io/hivepaas/function-runtime-$n:1.2.0 --format '{{json .Manifest}}' | jq -r .digest)"
done
```

---

## Task 7: HivePaaS pins 1.2.0 (backend, after Task 6)

- [ ] **Step 1:** In `hivepaas_app/base/version.go`, `functionRuntimesV1()` names the five images of Task 6 at `1.2.0`, by digest, in the order `node24`, `bun1`, `python313`, `go127`, `go127-build`, its comment saying function-runtimes v1.2.0 and `call`; `release.json`'s `beta.functionRuntimes` names the same references.
- [ ] **Step 2:** `go build ./... && go test ./...`, `golangci-lint run ./...`, the custom lints; `go run ./tools/releasepin -check` lists no `function-runtime` image.
- [ ] **Step 3:** Commit: `feat(functions): function-runtimes 1.2.0 pinned - hivepaas-runtime call`, with the trailer.

## Finish

Merge each branch into its `main`, locally; backend `main`: `go test ./...`; dashboard `main`: `lint:ci`, `build`. The user re-signs `release.json` (`make release-sign`), and checks in the browser: a scheduled call's task, its request id, and the app's Logs it opens.
