# The PHP runtime

Status: approved and built 2026-10-16 - the runtime `php85`, Go's server in
front of PHP processes, a handler `handler`, Composer, the extensions below.
function-runtimes v1.3.0 released it; HivePaaS pins the six 1.3.0 images, Go
1.27.2 in Go's.

## Why

PHP is what a large part of the web is written in, and a function is the
smallest thing to run some of it: a webhook, a form's handler, a page. Every
other runtime HivePaaS has is a language a PHP team does not use.

## The runtime

- **`php85`**, image `ghcr.io/hivepaas/function-runtime-php85`: PHP 8.5, the
  newest stable. A later PHP is another runtime, as `node24` is Node.js 24.
- **The server is Go's**: the runtime's `hivepaas-runtime` is a Go command built
  on `hivepaas/serve`, the Go runtime's own server, which already keeps the
  whole contract - HTTP, 413/429/504, the invocation lines, the health check,
  `invoke`, `call`, keep-alive, IPv6, a stop that lets calls finish. It hands
  each call to a **PHP process that keeps the handler loaded** and runs one call
  at a time - the model of RoadRunner, Laravel Octane and Bref's function
  runtime. Not php-fpm, whose share-nothing calls do not fit the contract's
  output and need a FastCGI client; not FrankenPHP, whose threads need a ZTS
  PHP and cgo for both architectures.
- **The server and a process** talk over two pipes, the process's file
  descriptors 3 (calls) and 4 (replies): a line of JSON, then the body's bytes.
  The process's standard output and error are the function's, passed on by the
  server line by line, whole - lines never interleave, unlike Python's - and a
  call's lines before the call is answered: after a call the process writes a
  sync line with the call's number, which the server takes out.
- **Processes**: one is kept, more are started as calls need them, up to
  `HP_FN_MAX_CONCURRENCY` and one per 32 MiB of the container's memory limit;
  every 30 seconds those above the most calls that ran at once are stopped.
  `WEB_CONCURRENCY` sets a number, all started (as Python's does). A PHP process
  waits on its I/O - a database, an API - holding its call, so the count follows
  the calls, not the CPUs. Idle processes take calls in turn.
- **The timeout**: PHP cannot be interrupted. The call is answered 504 at once
  and its process is stopped - with what it started, its process group - and
  replaced. The other runtimes let the handler run on, holding its place.
- **A process that ends** - `exit`, a fatal error such as `memory_limit` - fails
  its call, 500, with PHP's message on standard error, and is replaced, at most
  once a second. A handler that cannot be loaded - a parse error, no such
  function - is not loaded: health 503, calls 500 `not_loaded`, the reason
  once, no process started again; `invoke` exits 3.

## The handler

```php
<?php

function handler(Hivepaas\Request $req, Hivepaas\Context $ctx): array
{
    return ['status' => 200, 'body' => ['hello' => $req->query['name'] ?? 'world']];
}
```

- `HP_FN_ENTRYPOINT`, `index.php` by default; `HP_FN_HANDLER`, `handler`: a
  function's name, with its namespace (`App\handle`), or a static method
  (`Handler::handle`).
- The entrypoint is loaded once by each process, after Composer's autoloader,
  at the top level: its variables are global, as when PHP runs the file. What a
  variable keeps is its process's, from one call to the next.
- `Hivepaas\Request`: `method`, `path`, `query` (first values), `queryAll`,
  `headers` (lower case, joined by `, `), `body` (a string), `text()`, `json()`
  (objects as arrays). `Hivepaas\Context`: `requestId`, `deadline` (seconds,
  as `microtime(true)`), `log(...)`.
- The response: `null` (204), or an array with any of `status` (an `int`),
  `headers`, `body`. A string body is text, `Hivepaas\Bytes` bytes, anything
  else JSON. A list is not a response.
- `echo` goes to the log, never to the response; `header()` does nothing.

## Libraries, settings, extensions

- **Composer 2**: `composer.json`, `composer install --no-dev
  --optimize-autoloader`, from `composer.lock`, which `deps` writes when there
  is none - a test run gives it back to the code. `auth.json` beside it has the
  credentials of private repositories; a `path` repository is the function's own
  directory, copied in as Node's `file:` libraries are.
- **php.ini**: the runtime's settings are production's - errors on standard
  error, one line each (`PHP Warning:  …`), `memory_limit` 128M, UTC. A
  `php.ini` in the function's code is read after them.
- **Extensions**: the official image's, and `bcmath`, `exif`, `gd`, `gmp`,
  `intl`, `mysqli`, `opcache`, `pdo_mysql`, `pdo_pgsql`, `pgsql`, `redis`,
  `sockets`, `zip`. Composer cannot add one.
- **OPcache is in the image, off.** A process compiles its code once, so
  OPcache's cache saves nothing; its JIT gave 4-6 % on the template, for 3 MiB
  more a process (below). `opcache.enable_cli` in the function's `php.ini` turns
  both on. (The proposal said on; measured, it is not worth its memory.)
- **The image** is built in the official `php:8.5-cli-trixie`, which keeps the
  compilers its extensions are built with (644 MB), and is PHP alone, with the
  packages of the libraries it loads, on `debian:trixie-slim`: 228 MB, Python's
  190 MB. Its build checks that every extension loads without a word.

## Measured

2026-10-16, Docker Desktop on an Apple M-series, the template's handler (`/json`)
called by `hey` from another container on the same network, 512 MiB:

| | `php85` | `python313` |
|---|---|---|
| 1 CPU, 16 at once | 10,900/s, p99 63 ms (16 processes) | 6,300/s, p99 19 ms |
| 1 CPU, 16 at once, `WEB_CONCURRENCY=1` | 7,400/s, p99 27 ms | - |
| 4 CPUs, 64 at once | 54,000/s, p99 6 ms (16 processes) | 25,800/s, p99 9 ms |
| Memory, at rest | 30 MiB (one process) | 25 MiB |
| Memory, 16 processes | 190 MiB | - |

A process takes 10 MiB of its own (13 MiB PSS); the server 11 MiB; a process
starts and loads the handler in about 20 ms, which is what a call waits for when
none is idle. On one CPU, more processes than 4 add no throughput and lengthen
the tail - the CPU's quota is shared by more of them - but the count follows the
calls, and a call that waits on I/O needs its own process.

## function-runtimes

- `runtimes/php85/`: the Go command (`main.go`, `pool.go`, `worker.go`,
  `size.go`, `deps.go`) and its tests, a fake PHP in the test binary;
  `runtime/`, the PHP side (`worker.php`, `Hivepaas\Request`, `Context`,
  `Bytes`, `Runtime\Worker`, `Runtime\Response`) and `tests/run.php`; `php.ini`;
  the Dockerfile.
- `hivepaas/serve` gains what a runtime that loads its handler needs:
  `SetLoaded` (health 503 until), `ErrNotLoaded` and `ErrResponseTooLarge` (a
  call's outcome, from another process), `Output` and `ErrorOutput` (a line
  passed on whole, none after invoke's result). The Go runtime's behaviour is
  unchanged, its conformance as before.
- `CONTRACT.md` v1 gains `php85` without a new version: the image, the
  extensions, `deps`, the defaults, the handler, a PHP column in the request's
  and the context's tables, the response, the timeout, the logs.
- Conformance runs `php85`, with several processes (`processes: true`), on a
  fixture whose library is a Composer `path` repository, so that it installs
  without a network; the whole suite passes. CI: the PHP side's tests in the
  official image, `php85` in the conformance matrix; the release builds and
  pushes `function-runtime-php85`.

## HivePaaS

Once function-runtimes is released with `php85` (a new tag rebuilds every
image: every function's runtime is then said to be outdated, as with Bun's):

- **The API**: `base.FunctionRuntimePhp85`, not compiled; default entrypoint
  `index.php`, handler `handler`; an entrypoint is a `.php` file; a handler is a
  function's name with its namespace, or `Class::method`. Creating, saving, test
  runs, scheduled calls and spec import take it as any runtime.
- **The build**: the libraries' manifest is `composer.json`, its lock
  `composer.lock`, its settings `auth.json`; a `path` repository is a local
  library, copied into the libraries' stage. A test run that wrote
  `composer.lock` gives it back.
- **The pins**: `php85` in `release.json`'s `functionRuntimes`, by digest; the
  test that asks every runtime for an image holds HivePaaS until then.
- **MCP**: the runtime in `tools_function.go`'s lists and `docs/functions.md`.
- **The dashboard**: the runtime **PHP 8.5** in Create Function and the
  settings, its template (`index.php`, `composer.json`), PHP's highlighting
  (Prism's `php`, after `markup-templating`), and the settings' note on its
  processes, as Python's has.
- **The CLI**: `--runtime php85`, its template, `vendor/` left out of a PHP
  function's code.
- **Docs**: the functions page's runtimes, handler, libraries; the CLI's and
  the MCP server's lists.
- **e2e**: a PHP function created from its template, run, deployed and called.

## Building it, in order

1. function-runtimes: the runtime, its tests, conformance, contract, CI. Done
   2026-10-16.
2. A function-runtimes release (`v1.3.0`): the user's, on GitHub. Done.
3. HivePaaS: the API, the build, the pins, MCP, swagger. Done.
4. The dashboard, the CLI, the docs, the e2e (2.132). Done; the dashboard waits
   for its review.

## The API level

`php85` in the runtime's enum changed the fingerprint of two write requests -
creating a function, saving its deployment settings - which raised the API
level to 3: every CLI released would have been refused its writes. A CLI writes
back the value it read, one it does not know as well, and loses nothing to a
value added; the level now leaves enums out (decided 2026-10-16), and stays 2.
The fingerprints were recorded again once, at level 2, as the change of the
rule.

## Later

- A slimmer worker start - `fork` from a process that loaded the code, as
  php-fpm does - if 20 ms per new process matters.
- Stopping a process after so many calls, as php-fpm's `pm.max_requests`, for
  code that leaks.
