# The Python Runtime on asyncio Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** The Python runtime serves on a server of its own on asyncio, with a process per CPU: a `def` handler runs in a thread as today, an `async def` handler on its process's loop; `WEB_CONCURRENCY` sets the processes, `HP_FN_MAX_CONCURRENCY` counts every process's calls together; HTTP/1.0 keep-alive and IPv6 work, for every runtime as the conformance suite now checks.

**Architecture:**
- **`server.py`** (rewritten): `listen(port)` opens a dual-stack socket (IPv4 alone without IPv6). `Server` serves it with `loop.create_server`: each `_Connection` (an `asyncio.Protocol`) reads into a buffer and takes requests from it - head, `Content-Length` or chunked body, `100 Continue` - and answers what the runtime answers itself at once and a handler's call in a task, one request at a time per connection. Deadlines (idle 300 s, head 10 s, body `HP_FN_TIMEOUT_MS`) are looked at by one sweep, four times a second. `Server.stop()` closes the listener and the idle connections and waits for the busy ones, at most one timeout.
- **`runtime.py`**: `Runtime.call_async` awaits a coroutine handler on the running loop, or a function in the thread pool, under `asyncio.timeout`; `call` (blocking, for `invoke`) stays. Places among the running calls come from `places` (`take(limit)`, `give()`): `LocalPlaces` by default. The pool and `call`'s own loop are made when first needed.
- **`processes.py`** (new): `process_count(env, cfg)` - `WEB_CONCURRENCY`, else the cgroup's CPU quota or the CPUs, within `HP_FN_MAX_CONCURRENCY` and one per 128 MiB of a memory limit; `SharedPlaces` - each worker's count in anonymous shared memory, no lock; `serve(cfg, count)` - one process serves itself, several are forked workers of a first process that restarts one that dies (at most once a second) and stops them on `SIGTERM`.
- **The conformance suite**: every runtime's `serve` checks gain HTTP/1.0 keep-alive (the answer's `Connection: keep-alive`) and IPv4/IPv6 from inside the container; Python's `serve` checks run twice (several processes, and `WEB_CONCURRENCY=1`), and a `processes` group checks the count, the shared limit, a killed worker's replacement and stopping.
- **The dashboard**: the Python template says `async def` is the fastest and must not block; the Concurrency help says 429, and, for Python, the processes and `WEB_CONCURRENCY`.

**Tech Stack:** Python 3.13 standard library (`asyncio`, `mmap`, `os.fork`), Go 1.27 and Docker (conformance), React (dashboard).

**Spec:** `docs/superpowers/specs/2026-10-02-python-runtime-asyncio-design.md`, amended by this plan's commit (its last section, "Changes after the plan").

## How the patches work

Every task's code was written and verified before this plan.

| Task | Repository | Base |
|---|---|---|
| 1 | `function-runtimes` | `main` at `0b63d97` |
| 2 | `hivepaas-dashboard` | `main` at `5da1bbcf` |

Patches live in `docs/superpowers/plans/2026-10-02-python-runtime/` of the backend: `py-01-tests.patch` then `py-01-code.patch`; `py-02-code.patch` for the dashboard, which has no unit test runner. With `P=/Users/tnt/go/src/github.com/hivepaas/hivepaas/docs/superpowers/plans/2026-10-02-python-runtime`, commands run from the worktree's root. A dashboard worktree needs `ln -s /Users/tnt/go/src/github.com/hivepaas/hivepaas-dashboard/node_modules node_modules`.

The branches the patches were cut from are kept as `prep/python-runtime` in both repositories; the Bun runtime's plan builds on function-runtimes' one. After the last task, `git diff prep/python-runtime -- . ':(exclude)docs/superpowers'` is empty in each.

`UT` below is the Python unit tests, as CI runs them:

```bash
UT() { docker run --rm -v "$PWD:/src:ro" -w /src -e PYTHONPATH=/src/runtimes/python313/runtime \
  -e PYTHONDONTWRITEBYTECODE=1 python:3.13-slim-trixie python -m unittest discover -s runtimes/python313/tests; }
```

**What was checked, on fresh worktrees of the bases above:**
- Task 1, the tests alone: `UT` fails with 7 errors (`hivepaas_runtime.server` has no `Server`, there is no `hivepaas_runtime.processes`, `Runtime` has no `call_async`); `RUNTIMES=python313` conformance fails `serve` and `serve, one process` on the HTTP/1.0 and IPv4/IPv6 checks, and `processes` on `WEB_CONCURRENCY sets how many processes serve` and `a process that dies is replaced`. On the same tests, `go127` and `node24` pass both new `serve` checks.
- Task 1, with the code: `UT` 49 tests OK; the whole conformance suite, every runtime, `ok`; `golangci-lint run --build-tags conformance ./...` 0 issues; `go test ./...` `ok`. Python's `serve` and `processes` groups passed three more runs in a row.
- Task 2: `npm run lint:ci` and `npm run build` pass; the Python template, written out of the patch, called with the new image through `invoke`, answers `{"hello":"Ada"}`.
- Measured with the spec's harness (`wrk` in a container on the same Docker network, `HP_FN_MAX_CONCURRENCY=64`): `def` 9,000/s on 1 CPU, 32,000/s on 4; `async def` 35,700/s and 106,000/s.

**What it needs on the machine:** Go 1.27, golangci-lint 2.13, Docker with Docker Hub (the conformance suite builds its images; the unit tests run in `python:3.13-slim-trixie`), Node.js and the dashboard's `node_modules`.

## Global Constraints

- **The handler's API and contract v1 do not change**; the images become `1.1.0`. `invoke` and `deps` are as before.
- **Processes:** `WEB_CONCURRENCY` (a positive whole number, else exit 2), else the cgroup's CPU quota rounded up (v2 `cpu.max`, v1 `cpu.cfs_quota_us`/`cpu.cfs_period_us`) or `os.process_cpu_count()`; at most `HP_FN_MAX_CONCURRENCY`; at most one per 128 MiB of a memory limit (v2 `memory.max`, v1 `memory.limit_in_bytes`); at least 1. stderr: `hivepaas: serving with N processes` (`1 process`).
- **The handler is loaded in each worker after the fork**, never in the first process.
- **The server keeps** every behaviour of `CONTRACT.md` and its suite; the head within 10 s of its first byte and at most 64 KiB (431), a request that cannot be read 400, both closing the connection and neither logged.
- **Python standard library only** in the runtime.
- **Git:** commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`; merge locally, never push - the release (the tag `v1.1.0`) is the user's, after the Bun runtime's plan; never `git stash`.

## Review Focus

1. **A handler module that connects to a database when it is imported**: each worker connects, after the fork; a function with 8 processes and a pool of 10 holds 80 connections, which a small database's limit may refuse.
   - Pinned by: nothing; the contract says what a module keeps is its process's.
2. **A memory limit with large libraries** (pandas, a model): the 128 MiB rule is a floor, not a measure; a process of 300 MiB in a 512 MiB function serving with 4 processes is killed by the kernel.
   - Pinned by: `test_at_most_one_per_128_mib_of_a_memory_limit` for the rule; `WEB_CONCURRENCY` is the way out, in the contract and the dashboard's help.
3. **A worker that dies at once, again and again** (a C extension crashing on import): it is started again once a second for as long as it dies; the others serve meanwhile; when all die, the health check fails and the orchestrator restarts the container.
   - Pinned by: `a process that dies is replaced` (conformance), for one death.
4. **An `async def` handler that blocks**: it holds up its process's other calls, and its 504 waits for it to give the loop back.
   - Pinned by: the contract's text; `test_a_coroutine_runs_on_the_callers_loop`.
5. **Two workers taking the last place at the same instant** may both have it: the count has no lock, so that a worker killed in the middle of taking one leaves nothing held.
   - Pinned by: nothing; a decision of this plan (below).

## Decisions beyond the spec

- **An `asyncio.Protocol` server**, not streams, and **coroutine handlers awaited under the timeout**, not as tasks of their own - measured, in the spec's last section.
- **No lock on the shared count** (Review Focus 5).
- **`serve` exits with `os._exit`** after the server stops, as `invoke` does: a `def` handler still running past its timeout in a thread does not hold the exit.
- **`serving with 1 process`** is said too, so the line is always there.
- **`Runtime.running`, `answering`, `begin_answer`, `end_answer` are gone**: the server counts what it answers; `runtime.positive_whole_number` is shared with `WEB_CONCURRENCY`.
- **The conformance suite**: `images.env` and `images.serve(t, env...)` give `WEB_CONCURRENCY=1` to every `serve` check of the second run; the Python fixture answers `/pid`; a worker is killed with the shell's `kill`, which slim images have no binary of.
- **The dashboard's Concurrency help** said the calls over it wait; it now says 429.

---

## Task 1: The Python runtime on asyncio, with a process per CPU (function-runtimes)

**Files:**
- Create: `runtimes/python313/runtime/hivepaas_runtime/processes.py`, `runtimes/python313/tests/test_server.py`, `runtimes/python313/tests/test_processes.py`, `conformance/processes_test.go`
- Rewrite: `runtimes/python313/runtime/hivepaas_runtime/server.py`
- Modify: `runtimes/python313/runtime/hivepaas_runtime/runtime.py` (`LocalPlaces`, `places`, `call_async`, `_settled`, `positive_whole_number`, the lazy pool and loop), `runtimes/python313/runtime/hivepaas_runtime/__main__.py` (`serve` through `processes`, `os._exit`), `runtimes/python313/tests/test_runtime.py` (`CallAsyncTest`)
- Modify: `conformance/conformance_test.go` (`runtimeDef.processes`, the second `serve` run, the `processes` group), `conformance/docker_test.go` (`images.env`, `images.serve`), `conformance/serve_test.go` (HTTP/1.0 keep-alive, IPv4/IPv6), `conformance/typescript_test.go` (`imgs.serve`), `conformance/fixtures/python313/main.py` (`/pid`)
- Modify: `CONTRACT.md` (`WEB_CONCURRENCY`, Python's processes, connections and addresses, interleaving lines), `README.md`

**Interfaces:**
- Consumes: nothing new.
- Produces:
  - `server.listen(port: int) -> socket.socket`; `server.Server(runtime)` with `async start(sock)`, `async stop()`; `server.serve(runtime, sock) -> int`
  - `processes.process_count(env, cfg, cgroup="/sys/fs/cgroup", cpu_count=os.process_cpu_count) -> int`; `processes.SharedPlaces(workers)` with `of(index)`, `reset(index)`, `total()`; `processes.serve(cfg, count) -> int`
  - `Runtime(handler, cfg, out=None, err=None, places=None)`, `Runtime.load(cfg, out=None, err=None, places=None)`, `async Runtime.call_async(request, request_id) -> Answer`
  - conformance: `images.env []string`, `(images) serve(t, env...) *container`, `runtimeDef.processes bool`, `checkProcesses(t, imgs)`, `(*container) fresh(path) response`, `(*container) pids(t, want) map[int]bool`

- [ ] **Step 1: Write the failing tests**

```bash
git apply --index $P/py-01-tests.patch
```

- [ ] **Step 2: Run them to watch them fail**

Run: `UT`
Expected: FAILED (errors=7) - `cannot import name 'Server' from 'hivepaas_runtime.server'`, `Failed to import test module: test_processes`, `'Runtime' object has no attribute 'call_async'` (4).

Run: `RUNTIMES=python313 go test -count=1 -v -tags conformance ./conformance/ | grep -- '--- FAIL'`
Expected: FAIL - in `serve` and `serve, one process`: `an HTTP/1.0 connection asked to be kept is kept`, `it answers on IPv4, and on IPv6 where its container has it`; in `processes`: `WEB_CONCURRENCY sets how many processes serve`, `a process that dies is replaced`.

- [ ] **Step 3: Write the code**

```bash
git apply --index $P/py-01-code.patch
```

- [ ] **Step 4: Run them to watch them pass, and the rest**

Run:
```bash
UT
go test -count=1 -tags conformance ./conformance/
golangci-lint run --build-tags conformance ./...
go test -count=1 ./...
```
Expected: `Ran 49 tests ... OK`; `ok` (every runtime); 0 issues; `ok`.

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(python313): serve on asyncio with a process per CPU - WEB_CONCURRENCY, a shared count of running calls, HTTP/1.0 keep-alive, IPv6

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

## Task 2: What the dashboard says of Python's processes (dashboard)

**Files:**
- Modify: `src/application/modules/projects/module-shared/constants/function-templates.constants.ts` (the Python template's comment)
- Modify: `src/application/modules/projects/routes/single-project/single-app/configuration/function-settings/form/function-settings.form.com.tsx` (the Concurrency help; the Python line)

**Interfaces:**
- Consumes: Task 1's `WEB_CONCURRENCY` and process rule, as the contract says them.
- Produces: nothing other tasks use.

- [ ] **Step 1: Write the code**

```bash
git apply --index $P/py-02-code.patch
```

- [ ] **Step 2: Check it**

Run: `npm run lint:ci && npm run build`
Expected: both pass.

- [ ] **Step 3: Commit**

```bash
git commit -m "feat(functions): the Python template says async def is the fastest; the Concurrency help says 429, and, for Python, the processes and WEB_CONCURRENCY

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

## Finish

- Merge both branches into their repositories' `main`, locally.
- function-runtimes `main`: `UT`, `go test -tags conformance ./conformance/`, `go test ./...`.
- Dashboard `main`: `npm run lint:ci && npm run build`.
- The release - the tag `v1.1.0`, pushed by the user - and HivePaaS's pins come with the Bun runtime's plan, whose last task they are.
