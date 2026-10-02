# TypeScript Functions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A Node.js 24 function can be written in TypeScript: its entrypoint may be a `.ts`, `.mts` or `.cts` file, Create Function offers TypeScript with its template, the editor highlights it, and the contract and its conformance suite say and check what Node.js runs.

**Architecture:**
- **The API (Task 1, backend):** `base.FunctionEntrypointExts[node24]` gains `.ts`, `.mts`, `.cts`. Creating, saving and spec import check the entrypoint through that one list, so nothing else changes.
- **The contract (Task 2, function-runtimes):** `CONTRACT.md` says what TypeScript runs on Node.js; the conformance suite's `node24` fixture gains `typed.ts` (with a file and a type it imports) and `enum.ts`, and a `typescript` group checks them through `serve` and `invoke`. `runtimeDef.typeScript` says how a runtime runs TypeScript (`typesRemoved` for `node24`), so the Bun runtime's plan can add its own. No image is released: the released `node24:1.0.0` (Node.js 24.21.0) already runs TypeScript.
- **The dashboard (Task 3):** `EFunctionLanguage` and `functionTemplateOf(runtime, language)`; Create Function shows **Language** for Node.js 24 and sends the TypeScript template and the entrypoint `index.ts`; `languageOfPath` gives `typescript`, highlighted with Prism's grammar.

**Tech Stack:** Go 1.27 (backend, conformance), Docker (conformance), React + TypeScript + zod (dashboard), Prism.

**Spec:** `docs/superpowers/specs/2026-10-02-typescript-functions-design.md`.

## How the patches work

Every task's code was written and verified before this plan.

**Repositories and bases** (each task works on a branch in a worktree of its repository):

| Task | Repository | Base |
|---|---|---|
| 1 | `hivepaas` (this one) | `main` at this plan's commit, which adds only documents to `2733f521` |
| 2 | `function-runtimes` | `main` at `5a2964f` |
| 3 | `hivepaas-dashboard` | `main` at `298712ab` |

Patches live in `docs/superpowers/plans/2026-10-02-typescript-functions/`: `ts-NN-tests.patch` then `ts-NN-code.patch` for tasks 1 and 2; `ts-03-code.patch` alone for the dashboard, which has no unit test runner. With `P=/Users/tnt/go/src/github.com/hivepaas/hivepaas/docs/superpowers/plans/2026-10-02-typescript-functions`, commands run from the worktree's root. A dashboard worktree needs `ln -s /Users/tnt/go/src/github.com/hivepaas/hivepaas-dashboard/node_modules node_modules`.

The branches the patches were cut from are kept as `prep/typescript-functions` in each of the three repositories. After the last task, `git diff prep/typescript-functions -- . ':(exclude)docs/superpowers'` is empty in each.

**What was checked, on fresh worktrees of the bases above:**
- Task 1: the tests alone fail (`TestANodeJSHandlerMayBeTypeScript`, `TestAnImportedTypeScriptFunctionIsValid`); with the code the packages pass, `go build ./...` passes, `golangci-lint run ./...` finds 0 issues, `goroutinelint` and `errcodelint` are clean, `make gen-swag` changes nothing.
- Task 2: the tests alone fail, the three checks of `TestConformance/node24/typescript`; with the code `RUNTIMES=node24 go test -tags conformance ./conformance/` passes whole, `golangci-lint run --build-tags conformance ./...` finds 0 issues.
- Task 3: `npm run lint:ci` and `npm run build` pass.
- The TypeScript template, written out of `ts-03-code.patch`, called with `function-runtime-node24:1.0.0@sha256:a4560e…` through `invoke` with `HP_FN_ENTRYPOINT=index.ts`: `{"hello":"Ada"}` for `?name=Ada`.

**What it needs on the machine:** Go 1.27, golangci-lint 2.13, Docker (conformance builds images and needs Docker Hub; `make gen-swag`), Node.js and the dashboard's `node_modules`.

## Global Constraints

- **Node.js runs TypeScript by removing types** (Node.js 24, strip-only): no `enum`, no `namespace` holding values, no parameter properties, no `import x = require(...)`, no decorators; imports name the `.ts` file; a type from another file is imported with `import type`; `tsconfig.json` is not read; types are not checked.
- **No new runtime and no image release:** a TypeScript function is a `node24` function whose entrypoint is `.ts`, `.mts` or `.cts`. The default entrypoint stays `index.js`.
- **Create Function:** Language (JavaScript default, TypeScript) only for Node.js 24; TypeScript sends `package.json` + `index.ts` and the entrypoint `index.ts`, from a repository too; the choice is not stored.
- **The backend's conventions** (`docs/ARCHITECTURE.md`), the linters (`golangci-lint run ./...`, `goroutinelint`, `errcodelint`), `go test ./...` in the main checkout after merging (it has `vendor/`).
- **The dashboard:** `npm run lint:ci`, `npm run build`; the browser check is the user's.
- **Git:** commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`; merge locally, never push; never `git stash`.

## Review Focus

1. **A repository function created as TypeScript whose handler is not `index.ts`** (`src/handler.ts`): it is created with `index.ts`, its first deployment cannot load it and answers 503 to its health check until the entrypoint is changed in its settings - as any repository function whose entrypoint is not the default.
   - Pinned by: nothing; a decision of the spec.
2. **A JavaScript function renamed to TypeScript in the Code tab** (`index.js` to `index.ts`): its entrypoint still says `index.js` until its settings change it; the call answers 500 `not_loaded`.
   - Pinned by: nothing new; an entrypoint names a file, as for any rename.
3. **An import without its extension** (`./lib/util`): the module cannot load; 500 `not_loaded`, 503 health, the error on stderr names the import.
   - Pinned by: the contract's text; not tested.
4. **A TypeScript function moved to Python in its settings:** its entrypoint `index.ts` is refused for `python313` (`ERR_VLD_FUNCTION_ENTRYPOINT_INVALID`) until it is changed; the dashboard keeps a non-default entrypoint when the runtime changes.
   - Pinned by: `TestANodeJSHandlerMayBeTypeScript`, its Python case (Task 1).
5. **A spec bundle of a TypeScript function imported where HivePaaS is older:** that HivePaaS refuses the entrypoint, `FUNCTION_SOURCE_INVALID`.
   - Pinned by: the parity test's `typescript` cases (Task 1), for this HivePaaS.

## Decisions beyond the spec

- **`runtimeDef.typeScript`** (`noTypeScript`, `typesRemoved`) in the conformance suite, so that the Bun runtime, which runs TypeScript whole, adds its own value and checks rather than a second group.
- **The enum check reads Node.js's error code** (`ERR_UNSUPPORTED_TYPESCRIPT_SYNTAX`) on stderr besides "the handler could not be loaded": a missing file says the same first line, and the check must fail for the right reason.
- **`functionTemplateOf(runtime, language)`** in the dashboard's constants, one place for a new function's files and entrypoint; Bun's template goes there.

---

## Task 1: A Node.js function's entrypoint may be TypeScript (backend)

**Files:**
- Modify: `hivepaas_app/base/function.go` (`FunctionEntrypointExts`)
- Test: `hivepaas_app/usecase/appsettingsuc/appsettingsdto/deployment_function_source_test.go`, `hivepaas_app/service/specservice/specserviceimpl/function_source_test.go`

**Interfaces:**
- Consumes: nothing new.
- Produces: `base.FunctionEntrypointExts[base.FunctionRuntimeNode24] == {".js", ".mjs", ".cjs", ".ts", ".mts", ".cts"}`.

- [ ] **Step 1: Write the failing tests**

```bash
git apply --index $P/ts-01-tests.patch
```

`TestANodeJSHandlerMayBeTypeScript` takes `index.ts`, `src/handler.mts`, `main.cts` for `node24` and refuses `index.ts` for `python313`; the parity cases gain `typescript` and `typescript in python`; `TestAnImportedTypeScriptFunctionIsValid` checks the spec import's own rules.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas_app/usecase/appsettingsuc/appsettingsdto/ ./hivepaas_app/service/specservice/specserviceimpl/`
Expected: FAIL - `TestANodeJSHandlerMayBeTypeScript` and `TestAnImportedTypeScriptFunctionIsValid`.

- [ ] **Step 3: Write the code**

```bash
git apply --index $P/ts-01-code.patch
```

- [ ] **Step 4: Run them to watch them pass, and the linters**

Run:
```bash
go test -count=1 ./hivepaas_app/usecase/appsettingsuc/... ./hivepaas_app/service/specservice/... ./hivepaas_app/base/
go build ./... && golangci-lint run ./...
go run ./tools/goroutinelint . && go run ./tools/errcodelint
make gen-swag && git status --short
```
Expected: every package `ok`; 0 issues; the custom lints clean; `git status` lists only the task's staged files.

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(functions): a Node.js function's entrypoint may be TypeScript - .ts, .mts, .cts

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

## Task 2: TypeScript in the contract and its conformance suite (function-runtimes)

**Files:**
- Create: `conformance/typescript_test.go`, `conformance/fixtures/node24/typed.ts`, `conformance/fixtures/node24/typed-lib/greet.ts`, `conformance/fixtures/node24/typed-lib/types.ts`, `conformance/fixtures/node24/enum.ts`
- Modify: `conformance/conformance_test.go` (`runtimeDef.typeScript`, the `typescript` group), `CONTRACT.md` (`HP_FN_ENTRYPOINT`, a TypeScript paragraph under "The handler"), `README.md` (the runtimes' table)

**Interfaces:**
- Consumes: the `node24` image as released.
- Produces: `type typeScript int` with `noTypeScript`, `typesRemoved`; `runtimeDef.typeScript`; `checkTypeScript(t, rt, imgs)`.

- [ ] **Step 1: Write the failing tests**

```bash
git apply --index $P/ts-02-tests.patch
```

- [ ] **Step 2: Run them to watch them fail**

Run: `RUNTIMES=node24 go test -count=1 -tags conformance -run 'TestConformance/node24/typescript' ./conformance/`
Expected: FAIL - `a TypeScript handler, served`, `a TypeScript handler, invoked`, `an enum cannot be loaded: the types are only removed` (the fixture has no `typed.ts` nor `enum.ts` yet).

- [ ] **Step 3: Write the code**

```bash
git apply --index $P/ts-02-code.patch
```

- [ ] **Step 4: Run them to watch them pass, and the linters**

Run:
```bash
RUNTIMES=node24 go test -count=1 -tags conformance ./conformance/
golangci-lint run --build-tags conformance ./...
go test -count=1 ./...
```
Expected: `ok` for the whole `node24` suite; 0 issues; the Go packages `ok`.

- [ ] **Step 5: Commit**

```bash
git commit -m "docs(contract): TypeScript on Node.js - what runs, what cannot be loaded; the conformance suite checks it

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

## Task 3: TypeScript in Create Function and the editor (dashboard)

**Files:**
- Modify: `src/application/modules/projects/module-shared/enums/e.function-runtime.ts` (`EFunctionLanguage`, `FUNCTION_LANGUAGE_LABELS`, `FUNCTION_RUNTIMES_WITH_LANGUAGES`)
- Modify: `src/application/modules/projects/module-shared/constants/function-templates.constants.ts` (`FUNCTION_TYPESCRIPT_TEMPLATE`, `functionTemplateOf`)
- Modify: `src/application/modules/projects/module-shared/utils/function-source.utils.ts` (`CodeLanguage`, `languageOfPath`)
- Modify: `src/application/modules/projects/module-shared/components/function-code-editor/function-code-editor.com.tsx` (Prism's `typescript`)
- Modify: `src/application/modules/projects/dialogs/create-function/schemas/create-function.form.schema.ts` (`language`, `createFunctionSource`)
- Modify: `src/application/modules/projects/dialogs/create-function/form/create-function.form.com.tsx` (the Language field, "Starts with")

**Interfaces:**
- Consumes: Task 1's API, which takes the entrypoint `index.ts` for `node24`.
- Produces: `functionTemplateOf(runtime: EFunctionRuntime, language: EFunctionLanguage): { files: FunctionFile[]; entrypoint: string }` - an empty entrypoint is the runtime's default.

- [ ] **Step 1: Write the code**

```bash
git apply --index $P/ts-03-code.patch
```

- [ ] **Step 2: Check it**

Run: `npm run lint:ci && npm run build`
Expected: both pass.

- [ ] **Step 3: Commit**

```bash
git commit -m "feat(functions): Create Function offers TypeScript for Node.js 24, with its template; .ts files are highlighted as TypeScript

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

## Finish

- Merge each task's branch into its repository's `main`, locally.
- Backend `main` (it has `vendor/`): `go build ./... && go test ./...`, all `ok`.
- function-runtimes `main`: `RUNTIMES=node24 go test -tags conformance ./conformance/`.
- Dashboard `main`: `npm run lint:ci && npm run build`.
- The browser check is the user's: Create Function with Node.js 24 and TypeScript, the Code tab's highlighting, a test run of the template.
