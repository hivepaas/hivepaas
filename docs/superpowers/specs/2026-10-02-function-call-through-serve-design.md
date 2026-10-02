# A scheduled call through the function's server

## Why

A `function-invoke` job, and a sequence's step that calls a function, run
`hivepaas-runtime invoke` in a running task of the function: a new process that
starts the runtime and loads the handler for the one call. Measured on a warm
instance, 1 CPU, ten calls each:

| Runtime | `docker exec` of nothing | `invoke` in all | The handler itself |
|---|---|---|---|
| Node.js 24 | ~50 ms | ~74 ms | 0.4 ms |
| Python 3.13 | ~49 ms | ~247 ms | 0.2 ms |
| Go 1.27 | ~51 ms | ~51 ms | 0.04 ms |

What `invoke` adds to `exec` is the runtime starting and the handler loading:
some 25 ms on Node.js, some 200 ms on Python. And the call shares nothing the
instance holds - a database pool, a cache - and is not counted among its running
calls.

A new command, `hivepaas-runtime call`, hands the same request to the `serve`
already running in the container, and writes the same result.

**In scope:** `call` in every runtime; scheduled calls (jobs and sequence steps)
through it; where the call's log is. **Not in scope:** test runs, which run code
not deployed and have no server to call, and keep `invoke`; calling the server
without `exec` (from the agent, over the project's network).

## The runtime: `call`

`hivepaas-runtime call` reads on standard input the request `invoke` reads, and:

- sends it to `serve` on this container, `127.0.0.1:HP_FN_PORT`, over HTTP/1.1:
  its method, its path and query, its headers, its body, and the header
  `X-Hivepaas-Call: 1`;
- writes the answer as `invoke` writes its result, the last line of standard
  output after `#hivepaas-result `: `status`; `headers`, by name in lower case,
  with `x-request-id`, without `content-length` and the runtime's own
  `x-hivepaas-*`; `body` in base64; `requestId`, the answer's `x-request-id`;
  `durationMs` and `outcome`, from the answer's `x-hivepaas-duration-ms` and
  `x-hivepaas-outcome`;
- writes nothing else on standard output: the call's log lines are `serve`'s,
  in the container's log, and carry its request id;
- on 429 waits for `Retry-After` and sends the request again, for as long as
  `HP_FN_TIMEOUT_MS` from its start allows; then the 429 is its result, outcome
  `throttled`;
- waits for an answer at most `HP_FN_TIMEOUT_MS` and ten seconds: `serve`
  answers 504 at the timeout.

It exits 0 when it wrote a result, whatever the status; 2 when the request on
its input cannot be read; 3 when `serve` does not answer - nothing listens, or
the connection ends before an answer. Exits 2 and 3 write no result.

`serve`, to a request with `X-Hivepaas-Call: 1`, adds `X-Hivepaas-Outcome` and
`X-Hivepaas-Duration-Ms` to its answer, and does not give the header to the
handler. Any other request is answered as today.

A path under `/_hivepaas/` is answered by the runtime, as `serve` answers it;
through `invoke` it reached the handler.

Node.js and Bun share `call` with the rest of the JavaScript runtime's code;
Python's is in its runtime; Go's is a command of the function's binary, which
`hivepaas-runtime call` runs as it runs `serve` and `invoke`.

## HivePaaS: scheduled calls through `call`

- A `function-invoke` job, and a sequence's step that calls a function, run
  `hivepaas-runtime call` in a running task of the function, with the request
  they send today.
- **An image without `call`** - a function built before the runtimes' `1.2.0` -
  exits 2 with its usage. On exit 2 the run sends the same request to
  `hivepaas-runtime invoke`, as today: a request that cannot be read is refused
  by `invoke` too, and an older image answers through it. A function built
  again runs on `1.2.0` and through `call`.
- **Exit 3** fails the run, and says that the function's server does not answer
  on that instance.
- The run's log says where the call's log is: `the call's log is in the app's
  logs, request <id>`. The task keeps the response as today, its request id with
  it.

## The dashboard

- A function call's task shows its request id, beside its status and duration,
  linking to the app's **Logs** filtered by it.
- The app's Logs take `?search=<text>` in their address, filling their search
  with it.

## The contract

`CONTRACT.md` v1 gains `call`, without a new version: the command in the image's
table, a section beside "Invoke", and `X-Hivepaas-Call` in "Paths of the
runtime"'s neighbourhood - what `serve` adds to the answer of such a request.

## Release

function-runtimes `1.2.0`, all five images, tagged and pushed by the user; then
HivePaaS pins them as for `1.1.0`, in `functionRuntimesV1()` and `release.json`.
A function moves to `call` at its next build.

## Testing

- **Conformance**, every runtime, `call` by `docker exec` in a serving
  container:
  - a request answered as `invoke` answers it - status, headers, body, the
    request id given or made - and the result's `outcome` and `durationMs` are
    `serve`'s;
  - a failed call and a call past its timeout are results, outcomes `error` and
    `timeout`;
  - with `HP_FN_MAX_CONCURRENCY=1` and a slow call running, `call` waits and is
    answered once it ends;
  - the call's invocation line and log line are in the container's log, its
    request id on them; nothing but the result on `call`'s output;
  - a request that cannot be read exits 2; no server - `call` in a container
    that does not serve - exits 3;
  - a request with `X-Hivepaas-Call: 1` from outside gets the two headers, and
    the handler does not see it.
- **Unit**, in each runtime where it has its own tests: reading the answer into
  the result; the retry's budget.
- **HivePaaS**: the job runs `call`; on exit 2 it runs `invoke` with the same
  request; on exit 3 it fails, saying why; the run's log names the request id.
- **The dashboard**: `lint:ci`, `build`; the link, in the browser, by the user.

## Later

- Calling the server from the agent over the project's network, without `exec`.
- The call's log lines copied into the run's log, by request id, from the logs
  backend.

## Changes after the plan

Writing and measuring the plan's code
(`docs/superpowers/plans/2026-10-02-function-call.md`) settled what follows:

- **`call` loads nothing of the runtime.** As first written, Python's `call`
  took 220 ms against `invoke`'s 214: the interpreter imported the runtime
  package - asyncio, the server - and `http.client`, which alone brings the
  `email` package, some 60 ms. What `call` reads now lives apart, in `wire.py`
  and `wire.mjs` - the configuration from the environment, `invoke`'s request
  and result marker - and Python's `call` speaks HTTP/1.1 on a socket.
- **What it does, measured** on warm instances, 20 calls each, `docker exec`
  alone 29 ms: Python `invoke` 181 ms, `call` 73 ms; Node.js 58 and 61 ms; Bun
  48 and 48 ms. On Node.js and Bun the process starting is all there is to a
  trivial handler; what `call` saves there is a handler's own loading - its
  libraries, its connections - which the measure's handler does not have.
- **A request `call` cannot send** - a method that is no token, a path that does
  not start with `/` or holds a space or a control character, a header that is
  no token or whose value holds CR, LF or NUL - is one it cannot read: exit 2,
  and the run sends it to `invoke`, which reads it as before.
- **HivePaaS's part needs no release**: on a runtime without `call` the job runs
  `invoke`, so the backend's change is merged before `1.2.0`; only the pins wait
  for it.
