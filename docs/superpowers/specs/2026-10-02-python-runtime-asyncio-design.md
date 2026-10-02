# The Python runtime on asyncio, a process per CPU

## Why

The Python runtime answers about a tenth of what the Node.js runtime does, and
does not use a second CPU. Measured on 2026-10-02, the template's handler, the
load from another container on the same Docker network (`wrk`), every call
logged as the runtime logs it:

| | 1 CPU, 16 at once | 4 CPUs, 64 at once |
|---|---|---|
| `python313:1.0.0` (`http.server`, a thread per connection) | 3,530/s, p99 45 ms | 4,990/s, p99 1.12 s, timeouts |
| An asyncio server, stdlib only, a handler `def` | 10,400/s, p99 34 ms | 36,700/s with 4 processes, p99 16 ms |
| The same, a handler `async def` | 44,900/s | 147,000/s with 4 processes |
| uvicorn + uvloop + httptools, a handler `def` | 10,100/s, p99 33 ms | 31,400/s with 4 processes, p99 10 ms |
| `node24:1.0.0`, for comparison | 42,800/s | 42,800/s (one process) |

What costs is not reading HTTP but handing a plain `def` handler to a thread,
and `http.server`'s thread per connection under the GIL. uvicorn reads HTTP in
C and is no faster here; its packages would sit on the function's module path
beside the function's own, where a function that pins another uvicorn, or
FastAPI, would break one or the other.

So the runtime gets a server of its own on asyncio, from the standard library
only, and serves with a process per CPU.

**In scope:** `serve` of the Python runtime; the conformance suite, for every
runtime where a new check concerns them all; the release of function-runtimes
1.1.0 and HivePaaS's pins to it. **Not in scope:** the handler's API, `invoke`,
`deps`, other runtimes' servers.

## What a function sees

- **The handler is called as today**, with the same request, context and
  response. The contract stays v1; the images are `1.1.0`.
- **A handler `async def` runs on its process's event loop.** It is the fast
  way, and the asyncio way: a coroutine that blocks - `time.sleep`, a blocking
  client - holds up every call of its process, and its timeout is answered only
  when it gives the loop back. Blocking code belongs in a plain `def`, which
  runs in a thread, as today, and is answered at its timeout whatever it does.
- **Several processes serve**, each with the handler loaded: what a module
  keeps in memory - a cache, a counter, a client's pool - is its process's.
  `HP_FN_MAX_CONCURRENCY` still counts the calls of the whole instance, every
  process together.
- **`WEB_CONCURRENCY` chooses the number of processes**, as gunicorn and
  uvicorn read it; `WEB_CONCURRENCY=1` serves as one process. It is a variable
  of the function's environment, not an `HP_FN_` one: HivePaaS sets those
  itself, and removes the function's own.
- **HTTP/1.0 keep-alive is kept**: a request that asks for it with
  `Connection: keep-alive` gets it. `http.server` did not answer `ab -k`.
- **IPv6**: `serve` listens on IPv6 and IPv4; on a host without IPv6, on IPv4.

## How many processes

When `WEB_CONCURRENCY` is not set, as many as the container's CPUs, then
limited:

1. **CPUs**: the CPU quota of the container's cgroup, rounded up, when it has
   one (a function with a CPU limit); otherwise the CPUs the process may run on
   (`os.process_cpu_count()`).
2. **At most `HP_FN_MAX_CONCURRENCY`**: a process with no call to run serves
   nothing.
3. **At most one per 128 MiB of the container's memory limit**, when it has
   one: a 256 MiB function serves with 2 processes, whatever its CPUs. A
   process of the template's handler takes about 30 MiB (measured on
   `python313:1.0.0`); one that loads large libraries, several times that, and
   its author sets `WEB_CONCURRENCY`.
4. At least 1.

The cgroup is read from cgroup v2 (`cpu.max`, `memory.max`), or v1
(`cpu.cfs_quota_us` and `cpu.cfs_period_us`, `memory.limit_in_bytes`).

`WEB_CONCURRENCY` sets the number itself, still at most
`HP_FN_MAX_CONCURRENCY`. A value that is not a positive whole number stops the
runtime before it serves, as an `HP_FN_` value does.

A function without a CPU limit on a large host gets as many processes as
`HP_FN_MAX_CONCURRENCY` allows, 16 by default, unless its memory limit allows
fewer.

`serve` writes on standard error, as it starts: `hivepaas: serving with 4
processes`.

## How it serves

**One process** (`WEB_CONCURRENCY=1`, or a limit of 1): the process serves,
as `serve` does today.

**Several:** the first process opens the listening socket, then starts the
others - forked, each loading the handler itself - and serves nothing:

- each worker accepts on the one socket, so an idle worker takes the next
  connection;
- the running calls are counted in memory the workers share, per worker: a
  call takes a place when the total is under `HP_FN_MAX_CONCURRENCY`, and is
  answered 429 when it is not, whichever worker has it;
- a worker that exits while serving is started again, at most once a second,
  its places given back; stderr says `hivepaas: process 123 exited (code 1);
  starting another`. The calls it was answering are lost with it;
- `SIGTERM` (or `SIGINT`) goes to every worker; each stops as `serve` stops
  today: no new connection, `Connection: close` on what it answers, idle
  connections closed, running calls given at most one timeout. The first
  process exits 0 when they have, and kills a worker still running 5 seconds
  after the timeout - within the stop grace HivePaaS gives a function, its
  timeout and 10 seconds.

The handler is loaded in each worker after the fork, never before it: a
module that opens a connection or starts a thread when it is imported does so
in the process that uses it.

## The server

A server of the runtime's own, on `asyncio`'s streams of bytes, in each
process: one event loop, a thread pool for plain `def` handlers.

It keeps what `serve` does, as the contract and its suite say:

- any method reaches the handler; `/_hivepaas/` paths do not, and are answered
  by the runtime after their body is read;
- a body by `Content-Length` or in chunks; one over `HP_FN_MAX_BODY_SIZE` is
  answered 413 and the connection closed, unread;
- a body that does not arrive within `HP_FN_TIMEOUT_MS` is no call: the
  connection is closed, nothing is answered, nothing is logged;
- `Expect: 100-continue` is answered `100 Continue` before the body is read,
  unless the body is too large;
- the call's timeout, 504, 429 with `Retry-After: 1`, 500 for an error, 204 and
  304 and `HEAD` without a body, `Content-Length`, `X-Request-Id`, a `Date`
  header and no `Server` header;
- one call at a time per connection; pipelined requests are answered in order;
- an idle connection is kept for 5 minutes; the head of a request must arrive
  within 10 seconds of its first byte;
- a request that cannot be read is answered 400, a head over 64 KiB 431, and
  the connection closed; neither is a call;
- the invocation line and the handler's log lines, each written whole in one
  write.

Lines of more than 4 KiB that two processes write at the same moment may
interleave on standard output: a pipe keeps a write whole up to that size. An
invocation line is far shorter; a handler's own long lines can meet it.

## The contract

`CONTRACT.md` v1 gains, without a new version:

- **Python**: several processes serve, as many as the container's CPUs, within
  `HP_FN_MAX_CONCURRENCY` and one per 128 MiB of a memory limit;
  `WEB_CONCURRENCY` sets it. What a module keeps is its process's. A coroutine
  that blocks holds up its process.
- **Every runtime**: `serve` listens on IPv6 and IPv4 where the host has both;
  a connection is kept between calls, for HTTP/1.1, and for HTTP/1.0 when the
  request asks for it.

## The dashboard

- The Python template's comment says that an `async def` handler is the fast
  one, and must not block.
- The function's settings, for Python: the Concurrency field's help says how
  many processes serve and that `WEB_CONCURRENCY` sets it.

## Release

The runtime changes are released together as function-runtimes `1.1.0`, with
what main holds since `1.0.0` - Go's entrypoint check, the header sentence of
the contract, TypeScript's conformance tests:

1. merged on main locally; the user pushes main and the tag `v1.1.0`;
2. CI builds and pushes the four images, and drafts the release with their
   digests;
3. HivePaaS pins them: `functionRuntimesV1()` in `base/version.go` and
   `release.json`'s `functionRuntimes`, changed together, all four at `1.1.0`.

A function moves to 1.1.0 at its next build, as a function picks up any new
runtime image; one deployed before keeps its image until then.

## Testing

- **Unit** (Python, in `python:3.13-slim-trixie`, as CI runs them):
  - the number of processes from cgroup v2 and v1 files, the affinity,
    `WEB_CONCURRENCY` and the limits above;
  - the shared count: places taken and given back across workers, a dead
    worker's places returned;
  - the server's reading of requests: `Content-Length`, chunks, a head split
    across reads, pipelining, HTTP/1.0 with and without keep-alive, a head too
    large, a request that is not HTTP;
  - the existing tests of `runtime.py`, unchanged where the handler's calls are
    unchanged.
- **Conformance**, the whole existing suite unchanged, for Python run twice:
  with `WEB_CONCURRENCY=1` and with several processes. New checks:
  - every runtime: an HTTP/1.0 request with `Connection: keep-alive` is
    answered and the connection kept; `serve` answers at `::1` and at
    `127.0.0.1` from inside the container, where it has IPv6;
  - Python: `WEB_CONCURRENCY=3` serves from 3 processes (the fixture's `/pid`
    over new connections); with 2 processes and `HP_FN_MAX_CONCURRENCY=1` the
    second call is answered 429, whichever process has it; a worker killed is
    replaced and calls go on; stopping with several processes lets a running
    call finish and exits 0.
- **Measured** before the release with the harness of the table above, on the
  same machine: a `def` handler at least 3 times `1.0.0` on 1 CPU, and growing
  with the CPUs. Not a CI gate.

## Later

- Free-threaded Python (3.14t), where a thread per call would not share one
  GIL, and one process could do what several do here.
- A setting for the number of processes in the function's settings, in place
  of the variable.
- uvloop, if the loop rather than the handler's thread becomes what costs.
