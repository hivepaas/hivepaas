# The Bun Runtime Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A function can run on Bun: the runtime `bun1`, the JavaScript runtime's code run by Bun 1.4.2, with `bun install` and `bun.lock`, TypeScript whole and `index.ts` by default - released with function-runtimes `1.1.0`, taken by HivePaaS's API and build with the five images' pins, and offered as "Bun 1" in the dashboard.

**Architecture:**
- **function-runtimes (Tasks 1-2):** the JavaScript runtime's code and tests move to `runtimes/js/`, built into `node24`'s image and `bun1`'s; the fixture moves to `conformance/fixtures/js/`, shared through `runtimeDef.fixture`. What differs by engine (`engine()`: `process.versions.bun`) is `deps` (`depsCommand`: `npm` or `bun install --production [--frozen-lockfile]`) and the default entrypoint (`index.ts` on Bun). `runtimes/bun1/` holds the Dockerfile (`oven/bun:1.4.2-slim`) and `hivepaas-runtime` (`exec bun /hivepaas/runtime/main.mjs`). The suite runs `bun1` with TypeScript whole (`whole.ts`) and a stated close for HTTP/1.0 (`closesHTTP10`); CI and the release build it.
- **The release (Task 3):** the user pushes function-runtimes' `main` and the tag `v1.1.0`, which builds and pushes all five images and drafts the release with their digests.
- **HivePaaS (Task 4, after the release):** `base.FunctionRuntimeBun1` - not compiled, `index.ts`/`default`, Node.js's handler pattern and file extensions; the build's manifest `package.json`/`bun.lock` with `.npmrc` and `bunfig.toml`; the five images pinned at `1.1.0` in `functionRuntimesV1()` and `release.json`.
- **The dashboard (Task 5):** `EFunctionRuntime.Bun1` "Bun 1", its default entrypoint and TypeScript template; moving between Node.js 24 and Bun 1 says nothing of rewriting the code.

**Tech Stack:** JavaScript (Node.js 24, Bun 1.4.2), Go 1.27 (backend, conformance), Docker, GitHub Actions (the release), React (dashboard).

**Spec:** `docs/superpowers/specs/2026-10-02-bun-runtime-design.md`, amended by this plan's commit (its last section, "Changes after the plan").

## How the patches work

Tasks 1, 2, 4 and 5 were written and verified before this plan; Task 3 is the user's.

| Task | Repository | Base |
|---|---|---|
| 1, 2 | `function-runtimes` | `main` at `868d0d6` |
| 4 | `hivepaas` (this one) | `main` at this plan's commit, which adds only documents to `75d229c2` |
| 5 | `hivepaas-dashboard` | `main` at `d05d0845` |

Patches live in `docs/superpowers/plans/2026-10-02-bun-runtime/`: `bun-01-code.patch` (a move: no test of its own, the existing ones prove it), `bun-02-tests.patch` then `bun-02-code.patch`, `bun-04-tests.patch` then `bun-04-code.patch`, `bun-05-code.patch`. With `P=/Users/tnt/go/src/github.com/hivepaas/hivepaas/docs/superpowers/plans/2026-10-02-bun-runtime`, commands run from the worktree's root. Patches 01 and 02 hold renames: apply them with `git apply --index`, as written. A dashboard worktree needs `ln -s /Users/tnt/go/src/github.com/hivepaas/hivepaas-dashboard/node_modules node_modules`.

The branches the patches were cut from are kept as `prep/bun-runtime` in the three repositories. Backend's ends with a commit `TEMP: a placeholder pin for bun1`, which no patch holds: Task 4 pins the release's digests instead. After the last task, `git diff prep/bun-runtime -- . ':(exclude)docs/superpowers'` is empty for function-runtimes and the dashboard, and, for the backend, holds only the pins.

`NT` below is the shared runtime's unit tests, as CI runs them:

```bash
NT() { docker run --rm -v "$PWD:/src:ro" -w /src node:24-trixie-slim node --test 'runtimes/js/test/*.test.mjs'; }
```

**What was checked, on fresh worktrees of the bases above:**
- Task 1: `NT` 17 pass; `RUNTIMES=node24` conformance `ok`.
- Task 2, the tests alone: `NT` fails (`runtime.mjs` does not provide `depsCommand`); `RUNTIMES=bun1` conformance fails to build (`runtimes/bun1` does not exist). With the code: `NT` 19 pass; the whole conformance suite, all four runtimes, `ok`; `golangci-lint run --build-tags conformance ./...` 0 issues; shellcheck of the three `hivepaas-runtime` scripts clean.
- Task 4, the tests alone: the packages fail to build (`undefined: base.FunctionRuntimeBun1`, 10 times). With the code: every package of `base`, `appsettingsuc`, `specservice`, `functionservice` `ok` but `TestEachReleaseNamesAnImageForEveryFunctionRuntime`, which wants `bun1`'s pin; `golangci-lint run ./...` 0 issues; the custom lints clean; `make gen-swag` adds nothing beyond the patch's `swagger.json`. With a placeholder pin, on the prep branch: `go test ./...` 214 `ok`.
- Task 5: `npm run lint:ci`, `npm run build` pass.
- The templates, written out of the patches, called through `invoke`: Bun's on the `bun1` image built from Task 2's code, `{"hello":"Ada"}`.
- Bun 1.4.2 read a `package-lock.json` and wrote `bun.lock` from it.

**What it needs on the machine:** Go 1.27, golangci-lint 2.13, Docker with Docker Hub, `jq`, Node.js and the dashboard's `node_modules`; for Task 3, the user's push rights on `hivepaas/function-runtimes`.

## Global Constraints

- **`bun1`**: `oven/bun:1.4.2-slim` (Debian 13), user `hivepaas` 10001, `/app`, port 8080, the same health check; `hivepaas-runtime` is the JavaScript runtime's `main.mjs`, run by Bun.
- **Bun's `deps`**: `bun install --production`, `--frozen-lockfile` when `bun.lock` is there; the cache in `/tmp/hivepaas-deps-cache`, removed after.
- **Bun's defaults**: entrypoint `index.ts`, handler `default`; the files and names Node.js takes.
- **HTTP/1.0 on Bun** is answered `Connection: close` and closed.
- **The release is the user's**: never push, never tag. HivePaaS's pins name `1.1.0` by digest, all five images changed together in `base/version.go` and `release.json`.
- **The backend's conventions** (`docs/ARCHITECTURE.md`), the linters, `make gen-swag`, `go test ./...` in the main checkout after merging; the dashboard's `lint:ci` and `build`.
- **Git:** commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`; merge locally; never `git stash`.

## Review Focus

1. **A Node.js function moved to Bun that holds only `package-lock.json`**: its first build makes `bun.lock` from it; a test run gives `bun.lock` back to the code, where `package-lock.json` stays, unused.
   - Pinned by: the check by hand recorded above; nothing automated.
2. **A library with a native addon built for Node.js** (an old `bcrypt`): Bun cannot load it; the handler is not loaded, the health check answers 503 and the previous deployment keeps running.
   - Pinned by: nothing; the contract promises Node.js's API, not every package's.
3. **A private registry**: `.npmrc` and `bunfig.toml` reach the install step, and a build secret such as `NPM_TOKEN` reaches it as for Node.js.
   - Pinned by: `TestABunFunctionInstallsFromBunsLockFile` for the files; the secret's path is the Node.js one, `TestANodeFunctionInstallsItsLibrariesBeforeItsCode`.
4. **An HTTP/1.0 client** gets `Connection: close` from Bun after each answer.
   - Pinned by: the conformance check `an HTTP/1.0 connection asked to be kept is kept, or said to close`.
5. **A later Bun 1.x** in a later release may change what a handler sees; the release runs the whole suite on it first.
   - Pinned by: the release workflow's conformance job.

## Decisions beyond the spec

- **The fixture's `index.ts`** re-exports `index.js`'s handler, so that Bun's default entrypoint is what its suite calls.
- **`runtimeDef.closesHTTP10`**, and the check reading Go's `resp.Close`: Go takes `Connection: close` out of the headers it gives back.
- **The engine is read from `process.versions.bun`**, and `loadConfig(env, engine)` and `depsCommand(engine, locked)` take it, so that the unit tests, under Node.js, cover Bun's choices.
- **`bunfig.toml`** is copied to the install step beside `.npmrc`.
- **Backend: `javaScriptHandler` and `javaScriptFiles`** in `base`, shared by `node24` and `bun1`.

---

## Task 1: The JavaScript runtime's code and fixture are shared (function-runtimes)

**Files:**
- Move: `runtimes/node24/runtime/` → `runtimes/js/runtime/`, `runtimes/node24/test/` → `runtimes/js/test/`, `conformance/fixtures/node24/` → `conformance/fixtures/js/`
- Modify: `runtimes/node24/Dockerfile`, `.github/workflows/test.yml` (the Node.js unit tests' path), `README.md`, `conformance/conformance_test.go` (`runtimeDef.fixture`, `fixtureName`), `conformance/docker_test.go`

**Interfaces:**
- Produces: `runtimeDef.fixture string`, `(runtimeDef) fixtureName() string`.

- [ ] **Step 1: Move**

```bash
git apply --index $P/bun-01-code.patch
```

- [ ] **Step 2: Check that nothing changed**

Run: `NT; RUNTIMES=node24 go test -count=1 -tags conformance ./conformance/`
Expected: `ℹ pass 17`, `ℹ fail 0`; `ok`.

- [ ] **Step 3: Commit**

```bash
git commit -m "refactor: the JavaScript runtime's code and conformance fixture are shared - runtimes/js, conformance/fixtures/js

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

## Task 2: The runtime bun1 (function-runtimes)

**Files:**
- Create: `runtimes/bun1/Dockerfile`, `runtimes/bun1/hivepaas-runtime`, `conformance/fixtures/js/index.ts`, `conformance/fixtures/js/whole.ts`
- Modify: `runtimes/js/runtime/runtime.mjs` (`engine`, `DEPS_CACHE`, `LOCK_FILES`, `depsCommand`, `loadConfig(env, runOn)`), `runtimes/js/runtime/main.mjs` (`deps`), `runtimes/js/test/runtime.test.mjs`
- Modify: `conformance/conformance_test.go` (`bun1`, `closesHTTP10`), `conformance/typescript_test.go` (`wholeTypeScript`), `conformance/serve_test.go` (the HTTP/1.0 check), `conformance/fixtures/js/Dockerfile` (its comment)
- Modify: `.github/workflows/test.yml` (conformance matrix, shellcheck), `.github/workflows/release.yml` (`IMAGES`, the build matrix), `CONTRACT.md`, `README.md`

**Interfaces:**
- Consumes: Task 1's `runtimes/js/` and `runtimeDef.fixture`.
- Produces: `engine(): 'bun' | 'node'`; `depsCommand(runOn, locked) -> { command, args, env }`; `LOCK_FILES`; `DEPS_CACHE`; `loadConfig(env, runOn = engine())`; `typeScript` value `wholeTypeScript`; `runtimeDef.closesHTTP10`; the image `function-runtime-bun1`.

- [ ] **Step 1: Write the failing tests**

```bash
git apply --index $P/bun-02-tests.patch
```

- [ ] **Step 2: Run them to watch them fail**

Run: `NT`
Expected: FAIL - `The requested module '../runtime/runtime.mjs' does not provide an export named 'depsCommand'`.

Run: `RUNTIMES=bun1 go test -count=1 -tags conformance ./conformance/`
Expected: FAIL - the image cannot be built: `runtimes/bun1: no such file or directory`.

- [ ] **Step 3: Write the code**

```bash
git apply --index $P/bun-02-code.patch
```

- [ ] **Step 4: Run them to watch them pass, and the rest**

Run:
```bash
NT
go test -count=1 -tags conformance ./conformance/
golangci-lint run --build-tags conformance ./...
docker run --rm -v "$PWD:/src:ro" -w /src koalaman/shellcheck:stable \
  runtimes/go127/hivepaas-runtime runtimes/python313/hivepaas-runtime runtimes/bun1/hivepaas-runtime
```
Expected: `ℹ pass 19`, `ℹ fail 0`; `ok` (all four runtimes); 0 issues; shellcheck silent.

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(bun1): the Bun runtime - the JavaScript runtime's code on Bun, bun install and bun.lock, index.ts by default

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

## Task 3: Release function-runtimes 1.1.0 (the user)

What `main` holds since `1.0.0` is released together: Go's entrypoint check, the contract's header sentence, TypeScript's checks, the Python runtime on asyncio, and Bun.

- [ ] **Step 1: Merge Tasks 1-2 into function-runtimes' `main`, locally**, and run there `NT`, `go test -tags conformance ./conformance/` and `go test ./...`.

- [ ] **Step 2: The user pushes `main`**, and waits for the `Test` workflow to pass on it.

- [ ] **Step 3: The user tags and pushes `v1.1.0`**: `git tag v1.1.0 && git push origin v1.1.0`. The `Release` workflow tests, builds the five images on amd64 and arm64, pushes them, and drafts the release with `digests.txt`. Publishing the draft is optional.

- [ ] **Step 4: Read the digests**, from the public images:

```bash
for n in node24 bun1 python313 go127 go127-build; do
  echo "$n ghcr.io/hivepaas/function-runtime-$n:1.1.0@$(docker buildx imagetools inspect \
    ghcr.io/hivepaas/function-runtime-$n:1.1.0 --format '{{json .Manifest}}' | jq -r .digest)"
done
```

Expected: five lines, each `…:1.1.0@sha256:<64 hex>`, the same as the draft's `digests.txt`.

---

## Task 4: HivePaaS takes bun1, and pins 1.1.0 (backend, after Task 3)

**Files:**
- Modify: `hivepaas_app/base/function.go` (`FunctionRuntimeBun1`, `AllFunctionRuntimes`, `Compiled`, `DefaultEntrypoint`, `javaScriptHandler`, `javaScriptFiles`), `hivepaas_app/service/functionservice/functionbuild/functionbuild.go` (the manifest), `hivepaas_app/entity/app_deployment_function.go` (its comment), `docs/openapi/swagger.json`
- Modify (Step 4): `hivepaas_app/base/version.go` (`functionRuntimesV1`), `release.json`
- Test: `hivepaas_app/base/function_test.go`, `hivepaas_app/usecase/appsettingsuc/appsettingsdto/deployment_function_source_test.go`, `hivepaas_app/service/specservice/specserviceimpl/function_source_test.go`, `hivepaas_app/service/functionservice/functionbuild/functionbuild_test.go`, `hivepaas_app/service/functionservice/functionbuild/real_build_test.go`, `hivepaas_app/service/functionservice/functiontest/real_run_test.go`

**Interfaces:**
- Consumes: Task 3's five digests.
- Produces: `base.FunctionRuntimeBun1 = "bun1"`, in `AllFunctionRuntimes` after `node24`.

- [ ] **Step 1: Write the failing tests**

```bash
git apply --index $P/bun-04-tests.patch
```

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas_app/base/ ./hivepaas_app/usecase/appsettingsuc/appsettingsdto/ ./hivepaas_app/service/functionservice/functionbuild/`
Expected: FAIL - `undefined: base.FunctionRuntimeBun1`.

- [ ] **Step 3: Write the code**

```bash
git apply --index $P/bun-04-code.patch
```

Run: `go test -count=1 ./hivepaas_app/base/`
Expected: FAIL - only `TestEachReleaseNamesAnImageForEveryFunctionRuntime`: the release names no `bun1` image yet.

- [ ] **Step 4: Pin the release's images**

In `hivepaas_app/base/version.go`, `functionRuntimesV1()` names the five images of Task 3's Step 4, each as `"ghcr.io/hivepaas/function-runtime-<name>:1.1.0" + "@sha256:<digest>"`, in the order `node24`, `bun1`, `python313`, `go127`, `go127-build`; its comment says function-runtimes v1.1.0. In `release.json`, `beta.functionRuntimes` names the same five references, in the same order. Nothing else of either file changes.

- [ ] **Step 5: Run them to watch them pass, and the whole repository**

Run:
```bash
go build ./... && go test ./...
golangci-lint run ./...
go run ./tools/goroutinelint . && go run ./tools/errcodelint
make gen-swag && git status --short
```
Expected: every package `ok`; 0 issues; the custom lints clean; `git status` lists only the task's files.

- [ ] **Step 6: Commit**

```bash
git commit -m "feat(functions): the runtime bun1, and function-runtimes 1.1.0 pinned - bun1, and the Python runtime on asyncio

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

## Task 5: Bun 1 in the dashboard (dashboard, after Task 4)

**Files:**
- Modify: `src/application/modules/projects/module-shared/enums/e.function-runtime.ts` (`Bun1`, its label and default entrypoint, `FUNCTION_JAVASCRIPT_RUNTIMES`)
- Modify: `src/application/modules/projects/module-shared/constants/function-templates.constants.ts` (Bun's template)
- Modify: `src/application/modules/projects/routes/single-project/single-app/configuration/function-settings/form/function-settings.form.com.tsx` (no rewrite note between Node.js 24 and Bun 1)

**Interfaces:**
- Consumes: Task 4's API, which takes `bun1`.

- [ ] **Step 1: Write the code**

```bash
git apply --index $P/bun-05-code.patch
```

- [ ] **Step 2: Check it**

Run: `npm run lint:ci && npm run build`
Expected: both pass.

- [ ] **Step 3: Commit**

```bash
git commit -m "feat(functions): the runtime Bun 1, with its TypeScript template; code moves between Node.js and Bun as it is

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

## Finish

- Merge Task 4's and Task 5's branches into their `main`, locally; backend `main` (it has `vendor/`): `go test ./...`; dashboard `main`: `npm run lint:ci && npm run build`.
- The browser check is the user's: Create Function with Bun 1, a test run of its template, a deployment on 1.1.0's images; a Python function deployed again runs the asyncio runtime.
