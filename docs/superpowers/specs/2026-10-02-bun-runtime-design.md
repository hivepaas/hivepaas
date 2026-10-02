# The Bun runtime

## Why

Bun runs JavaScript and TypeScript faster than Node.js, in less memory, and
runs TypeScript whole - `enum`, `namespace`, imports without an extension -
where Node.js only removes types. It implements Node.js's modules, `node:http`
included, so the Node.js runtime's own code runs on it.

Measured on 2026-10-02 (Bun 1.4.2, `oven/bun:1.4.2-slim`), the template's
handler, the load from another container on the same Docker network (`wrk`):

| | `node24` | The Node.js runtime's code on Bun | A server on `Bun.serve` |
|---|---|---|---|
| 1 CPU, 16 at once | 39,900/s, p99 19 ms | 55,600/s, p99 5 ms | 60,800/s |
| 4 CPUs, 64 at once | 41,500/s | 64,600/s | 67,900/s |
| Memory of an instance | 62 MiB | 38 MiB | - |

With only `deps` changed to `bun install`, the Node.js runtime's code on Bun
passed the whole conformance suite, its 32 checks. A server of Bun's own
(`Bun.serve`) would be some 6 % faster, for a second server to keep to the
contract; it is not written.

**In scope:** the runtime `bun1` - its image, its contract, its conformance,
HivePaaS's API, build and dashboard for it. **Not in scope:** a server on
`Bun.serve`; Bun's own APIs as part of the contract (a handler may use them,
the contract does not promise them).

## The runtime

- **`bun1`**, image `ghcr.io/hivepaas/function-runtime-bun1`. The id is
  Bun's major, as `node24` is Node.js's; each runtime release pins one Bun 1.x
  in its Dockerfile, raised on purpose, its conformance run on it.
- **Base**: `oven/bun:<version>-slim`, Debian 13 (trixie), as the contract
  says of every image; user `hivepaas`, uid and gid 10001; `/app`; port 8080;
  the same health check.
- **The same code as `node24`.** The runtime's code moves from
  `runtimes/node24/runtime/` to `runtimes/js/runtime/` and is built into both
  images; its unit tests move with it and run under Node.js as today.
  `hivepaas-runtime` is `exec bun /hivepaas/runtime/main.mjs` in Bun's image.
  What differs, it decides by `process.versions.bun`:
  - `deps`: `bun install --production`, with `--frozen-lockfile` when `bun.lock`
    is there; without one, `bun install` writes it. A function that holds only
    `package-lock.json` has it read by Bun, which writes `bun.lock` from it;
  - the default entrypoint: `index.ts`.
- **The handler's API is Node.js's**: the request, the context, the response,
  the errors, the logs. Code moves between `node24` and `bun1` unchanged, but
  for what it imports of one engine only.

## TypeScript on Bun

Bun transpiles TypeScript: everything TypeScript writes runs, `enum`,
`namespace` and parameter properties included; an import may leave out
`.ts`; `tsconfig.json`'s `paths` apply. Types are not checked. The template is
TypeScript.

## The contract

`CONTRACT.md` v1 gains `bun1` without a new version:

- the runtimes, the image, its base and its libraries (`/app/node_modules`,
  installed from `package.json` and `bun.lock`);
- `HP_FN_ENTRYPOINT`: a `.js`, `.mjs`, `.cjs`, `.ts`, `.mts` or `.cts` file,
  `index.ts` by default; `HP_FN_HANDLER`: `default`;
- the handler, the request, the context and the response: Bun shares Node.js's
  column of each table ("Node.js and Bun");
- TypeScript: on Bun, whole, as above.

## HivePaaS

**The API**: `bun1` is a runtime (`base.FunctionRuntimeBun1`): not compiled;
the default entrypoint `index.ts` and handler `default`; a handler's name and
an entrypoint's extensions as Node.js's (`.js`, `.mjs`, `.cjs`, `.ts`, `.mts`,
`.cts`). Creating, saving, test runs, scheduled calls and spec import take it
as any runtime.

**The build**: the libraries' manifest is `package.json`, their lock file
`bun.lock`, kept as `package-lock.json` is: a test run that writes it gives it
back to the code.

**The pins**: `bun1`'s image is pinned with the others, in
`functionRuntimesV1()` and `release.json`. A test asks every runtime to have
one, so HivePaaS takes `bun1` only once its image is released: with the
function-runtimes `1.1.0` release, beside the Python runtime's.

**The dashboard**: the runtime **Bun 1** in Create Function and in a function's
settings, with its template:

- `package.json` (`"type": "module"`);
- `index.ts`, the handler of the Node.js TypeScript template, its types in the
  file - TypeScript whole on Bun, so nothing it says of `enum` applies.

Moving a function between Node.js 24 and Bun 1 in its settings keeps its code,
and says nothing of rewriting it; to or from another runtime the settings say
what they say today.

## function-runtimes

- `runtimes/js/` (the shared code and its tests), `runtimes/node24/Dockerfile`,
  `runtimes/bun1/Dockerfile` and `runtimes/bun1/hivepaas-runtime`.
- The conformance suite runs `bun1`, with a fixture of its own
  (`conformance/fixtures/bun1/`, the Node.js fixture's paths, its library from
  `file:`); and `bun1`-only: a handler with an `enum` and an import without an
  extension answers. The checks the Python runtime's spec adds for every runtime
  - HTTP/1.0 keep-alive, IPv6 and IPv4 - run on it too.
- CI: `bun1` in the conformance matrix, on amd64 and arm64; the release builds
  and pushes `function-runtime-bun1`, its digest in the drafted release.
- `README.md`: `bun1` in the table, and how to try it.

## Testing

- The shared runtime's unit tests, under Node.js, with `deps`' choice of
  command and the default entrypoint by engine.
- Conformance, `bun1`, on both architectures, as above.
- The API: `bun1` is accepted with `index.ts`, `src/main.mjs`; refused with
  `main.py`; its default entrypoint is `index.ts`; the spec import's parity test
  covers it.
- The build: a `bun1` function's Dockerfile installs from `bun.lock`; a test run
  that wrote `bun.lock` gives it back.
- The template is called with the `bun1` image through `invoke`, as the others
  were.
- The dashboard: `lint:ci`, `build`, and the user in the browser.

## Later

- A server on `Bun.serve`, if the 6 % becomes worth a second server.
- Bun's own APIs (`Bun.sql`, `Bun.redis`) said in the contract, for what a
  function may count on.

## Changes after the plan

Writing the plan (`docs/superpowers/plans/2026-10-02-bun-runtime.md`) settled
what follows:

- **Bun closes an HTTP/1.0 connection after each answer**, even one the request
  asks to keep, and says `Connection: close`, so an HTTP/1.0 client does not wait
  for more. Node.js keeps it; Bun's `node:http` does not, with or without the
  header set by hand. The contract says so, and the conformance check takes a
  stated close for Bun (`runtimeDef.closesHTTP10`). A proxy in front speaks
  HTTP/1.1, so a function never meets it there.
- **One fixture for both JavaScript runtimes**, `conformance/fixtures/js`, which
  gains `index.ts` - Bun's default entrypoint - re-exporting `index.js`'s
  handler, and `whole.ts`, an `enum` and an import without its extension, which
  Bun runs.
- **Bun reads `package-lock.json`** to write `bun.lock` when there is none,
  checked with a `file:` library; `deps` downloads into
  `/tmp/hivepaas-deps-cache`, for both engines, and removes it.
- **A Bun function's install step** copies `bunfig.toml` beside `.npmrc`.
- **HivePaaS takes `bun1` with its pin**: a test asks every runtime for one, so
  the backend's task runs after the release, with the five images' digests.
