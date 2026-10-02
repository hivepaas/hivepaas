# Functions in a Spec Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A function travels through spec export and import: export writes it as any app, its inline code as files of the bundle under its own directory; reading puts the code back inline; import creates and updates a function through what creating one checks, and an app keeps its kind.

**Architecture:**
- **The code as files (Task 1):** the env documents are built as today, the code inline. Only when an env document is written does `writeFunctionCode` take each inline function's files out into `projects/<p>/envs/<env>/functions/<app>/<path>`, its source keeping `code.inline.filesFrom`. `parseBundleFiles` - for an upload and for the target's own state alike - calls `readFunctionCode`, which puts the files back inline, so every later step (report, references, the comparison of a bundle with what exists) sees code inline. Export stops skipping functions.
- **The checks (Task 2):** `checkFunctions` runs with the plan's other checks. For each app the import writes, it reads its kind and deployment source as they will be - the bundle's for what the import writes, the app's own otherwise - and blocks `APP_KIND_CHANGED`, `FUNCTION_KIND_MISMATCH`, `FUNCTION_SOURCE_INVALID` and `FUNCTION_BUILD_SOURCE`. The writer normalizes a function's source and pins a function's routing to its runtime's port. A service does not import a DTO, so the source's checks are written for the entity in `specserviceimpl` and held to the API's by a test.

**Tech Stack:** Go 1.27, the backend's spec service (yaml.v3 documents, tar/age bundles).

**Spec:** `docs/superpowers/specs/2026-10-02-functions-in-spec-design.md`, amended by this plan's commit (its last section, "Changes after the plan").

## How the patches work

Every task's code was written and verified before this plan.

**Repository:** this one (the backend). **Base:** `main` at this plan's commit, which adds only documents to `b3b6de9e` ("docs(specs): functions in a spec…"). Work on a branch in a worktree.

Patches live in `docs/superpowers/plans/2026-10-02-functions-in-spec/`: `fs-NN-tests.patch` then `fs-NN-code.patch` for task NN. With `P=/Users/tnt/go/src/github.com/hivepaas/hivepaas/docs/superpowers/plans/2026-10-02-functions-in-spec`, commands run from the worktree's root.

The branch the patches were cut from is kept as `prep/functions-in-spec`. After the last task, `git diff prep/functions-in-spec -- . ':(exclude)docs/superpowers'` is empty.

**What was checked, on a fresh worktree of `b3b6de9e`, at every task:** the tests alone fail to build, as step 2 says; with the code the task's packages pass, `go build ./...` passes, `golangci-lint run ./...` finds 0 issues, `goroutinelint` and `errcodelint` are clean, and `make gen-swag` changes nothing. **After the last task**, on the prep branch, whose tree the last task equals: `go test ./...` all `ok` (214 packages).

**What it needs on the machine:** Go 1.27, golangci-lint 2.13, Docker for `make gen-swag`, `tar` (the export tests read the archive).

## Global Constraints

- **Bundle layout:** a function's inline code is at `projects/<project>/envs/<env>/functions/<app>/<path>`, `<app>` the app's key in the env document; its source holds `code.inline.filesFrom: projects/<project>/envs/<env>/functions/<app>` instead of `code.inline.files`, relative to the bundle's root. A repository function's source is written as it is.
- **A bundle is unreadable** (`ErrSpecBundleInvalid`, nothing planned) when `filesFrom` is not the function's own directory, the directory holds no file, a file's path is not a function's file path (`base.FunctionPathOK(path, false)` and not under `.hivepaas`), or a file is not UTF-8. A file under `functions/` that no source names is ignored.
- **Blocked codes:** `FUNCTION_SOURCE_INVALID` (detail `problems`, one line per field), `FUNCTION_KIND_MISMATCH`, `APP_KIND_CHANGED`, `FUNCTION_BUILD_SOURCE`. `FUNCTION_SKIPPED` is gone.
- **The backend's conventions** (`docs/ARCHITECTURE.md`; **a service does not import a DTO**), the linters (`golangci-lint run ./...`, `goroutinelint`, `errcodelint`), `make gen-swag` when a DTO changes, `go test ./...` in the main checkout after merging (it has `vendor/`), tests with `assert` and fakes.
- **Git:** commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`; merge locally, never push; never `git stash`.

## Review Focus

1. **An env with several large functions:** each function's code is files of its own, so the env document stays small; the bundle's limits are 16 MiB per decompressed file and 128 MiB in all, and a function's code is at most 1 MB.
   - Pinned by: `TestExportWritesAFunctionWithItsCodeAsFiles` (Task 1) for the layout; the sizes are the existing reader's.
2. **A bundle edited by hand in git:** a file renamed, added or removed under a function's directory reads back as the function's code; a stray directory is ignored; a file whose path no function may have refuses the bundle.
   - Pinned by: the four `TestABundle…IsRefused` tests and `TestAFunctionsDirectoryNobodyNamesIsIgnored` (Task 1).
3. **A function from a repository whose credentials the import creates too:** the repository is not reached at import, as for any repository app; its first build says if it cannot be.
   - Pinned by: nothing; a decision of this plan (below).
4. **Importing onto an installation of several nodes** where the function names a registry the bundle carries as an external reference: the check sees a name, not an id, and passes; the reference itself is resolved, or asked about, by the import's existing reference checks.
   - Pinned by: `TestImportBlocksAFunctionWithoutARegistryOnSeveralNodes` (Task 2) for the refusal.
5. **An env sequence whose step is a function's call**, the function excluded from the import's selection: the step names an app the import does not write, as for any app job, and the existing reference checks report it.
   - Pinned by: nothing new.

## Decisions beyond the spec

- **The repository is not reached at import.** Creating a function checks that its repository answers at its ref with its credentials; an import may create those credentials in the same run, so at plan time there is nothing to reach the repository with. As for a repository app imported today, the first build says when it cannot be reached. `FUNCTION_BUILD_SOURCE` covers the registry a cluster of several nodes needs.
- **A function's checks block the whole import**, as the spec says, rather than skipping the one app: the bundle is what has to change.
- **The source's checks are written twice**, for the API's DTO and for the entity in `specserviceimpl`, because a service does not import a DTO; `TestAnImportedFunctionSourceIsCheckedAsTheAPIChecksIt` runs one set of sources through both and fails when they disagree. The path, handler and entrypoint rules they share moved to `base` (`FunctionPathOK`, `FunctionPathReserved`, `FunctionHandlerPatterns`, `FunctionEntrypointExts`), and the routing pin to `entity` (`(*AppRoutingSettings).PinToFunctionPort`).
- **A function's routing is pinned to its runtime's port when written**, as saving its routing pins it, whatever the bundle says.

---

## Task 1: A function is exported with its code as files of the bundle, and read back inline

**Files:**
- Create: `hivepaas_app/service/specservice/specserviceimpl/function_code.go` (`functionDir`, `inlineCodeOf`, `writeFunctionCode`, `readFunctionCode`)
- Modify: `hivepaas_app/service/specservice/specserviceimpl/walk.go` (no skip; the code taken out when the env document is written), `hivepaas_app/service/specservice/specserviceimpl/bundle_read.go` (the code read back), `hivepaas_app/service/specservice/specmodel/issue.go` (`CodeFunctionSkipped` gone), `hivepaas_app/base/function.go` (`FunctionPathOK`, `FunctionPathReserved`, `FunctionPathMaxLen`, `FunctionHandlerPatterns`, `FunctionEntrypointExts`), `hivepaas_app/usecase/appsettingsuc/appsettingsdto/deployment_function_source.go` (uses them)
- Test: `hivepaas_app/service/specservice/specserviceimpl/function_code_test.go`, `hivepaas_app/service/specservice/specserviceimpl/export_test.go` (the fixture's function has two files; the skip test becomes the files test)

**Interfaces:**
- Produces: `filesFromKey = "filesFrom"`; `functionDir(envPath, appKey string) string`; `inlineCodeOf(doc *specmodel.AppDoc) map[string]any`; `writeFunctionCode(bundle *specmodel.Bundle, envPath string, apps map[string]*specmodel.AppDoc)`; `readFunctionCode(bundle *specmodel.ImportBundle, files map[string][]byte) error`; `base.FunctionPathOK(p string, dir bool) bool`, `base.FunctionPathReserved(p string) bool`, `base.FunctionHandlerPatterns`, `base.FunctionEntrypointExts`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/fs-01-tests.patch`

Tests added or changed: `TestExportWritesAFunctionWithItsCodeAsFiles` (in place of `TestExportSkipsAFunction`), `TestAFunctionsCodeIsReadBackFromItsFiles`, `TestABundleWhoseFunctionNamesAnotherDirectoryIsRefused`, `TestABundleWhoseFunctionHasNoFileIsRefused`, `TestABundleWhoseFunctionHasAFileNoFunctionMayHaveIsRefused`, `TestABundleWhoseFunctionHasAFileThatIsNotTextIsRefused`, `TestAFunctionsDirectoryNobodyNamesIsIgnored`, `TestAnExportedFunctionImportedWhereItCameFromChangesNothing`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas_app/service/specservice/... ./hivepaas_app/entity/ ./hivepaas_app/usecase/appsettingsuc/...`
Expected: FAIL to build: `undefined: inlineCodeOf`, `undefined: filesFromKey`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/fs-01-code.patch`

- [ ] **Step 4: Run them to watch them pass, and the linters**

Run: `go build ./... && go test -count=1 ./hivepaas_app/service/specservice/... ./hivepaas_app/entity/ ./hivepaas_app/usecase/appsettingsuc/...`
Expected: every package `ok`

Run: `golangci-lint run ./... && go run ./tools/goroutinelint . && go run ./tools/errcodelint`
Expected: `0 issues.`; the two custom linters print nothing and exit 0

Run: `make gen-swag && git status --short | grep -v '^[AM]  '`
Expected: no output

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(spec): a function is exported with its code as files of the bundle, and read back inline

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 2: An imported function goes through what creating one checks

**Files:**
- Create: `hivepaas_app/service/specservice/specserviceimpl/function_source.go` (`normalizeFunctionSource`, `functionSourceProblems`), `hivepaas_app/service/specservice/specserviceimpl/import_functions.go` (`checkFunctions`, `checkFunction`, `natureOf`)
- Modify: `hivepaas_app/service/specservice/specmodel/importplan.go` (the four codes), `hivepaas_app/service/specservice/specserviceimpl/import_plan.go` (`checkFunctions` in `plan`, `planner.functionApps`), `hivepaas_app/service/specservice/specserviceimpl/import_write.go` (a function's source normalized, its routing pinned), `hivepaas_app/entity/app_routing_settings.go` (`PinToFunctionPort`), `hivepaas_app/usecase/appsettingsuc/function_rules.go` (uses it)
- Test: `hivepaas_app/service/specservice/specserviceimpl/import_functions_test.go`, `hivepaas_app/service/specservice/specserviceimpl/function_source_test.go`, `hivepaas_app/service/specservice/specserviceimpl/export_test.go` (the fixture's function has a `function-invoke` job; the cluster fake answers `IsMultiNode`)

**Interfaces:**
- Consumes: Task 1's `base.FunctionPathOK`, `base.FunctionPathReserved`, `base.FunctionHandlerPatterns`, `base.FunctionEntrypointExts`.
- Produces: `specmodel.CodeFunctionSourceInvalid`, `CodeFunctionKindMismatch`, `CodeAppKindChanged`, `CodeFunctionBuildSource`; `normalizeFunctionSource(*entity.DeploymentFunctionSource)`; `functionSourceProblems(*entity.DeploymentFunctionSource) []string`; `(*entity.AppRoutingSettings).PinToFunctionPort()`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/fs-02-tests.patch`

Tests added: `TestImportCreatesAFunctionAsCreatingOneDoes` (kind, normalized source, routing at 8080, first deployment), `TestImportBlocksAFunctionWhoseSourceIsNotValid`, `TestImportBlocksAKindThatDisagreesWithItsSource` (three cases), `TestImportBlocksAnAppChangingKind` (both ways), `TestImportBlocksAFunctionWithoutARegistryOnSeveralNodes`, `TestApplyRefusesAFunctionTheChecksBlock` (the write path refused, nothing provisioned), `TestAnImportedFunctionSourceIsCheckedAsTheAPIChecksIt` (18 sources through the DTO and the entity's checks); `TestExportWritesAFunctionWithItsCodeAsFiles` also checks the function's call travels with it.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas_app/service/specservice/... ./hivepaas_app/entity/ ./hivepaas_app/usecase/appsettingsuc/...`
Expected: FAIL to build: `undefined: normalizeFunctionSource`, `undefined: functionSourceProblems`, `undefined: specmodel.CodeFunctionSourceInvalid`, `undefined: specmodel.CodeFunctionKindMismatch`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/fs-02-code.patch`

- [ ] **Step 4: Run them to watch them pass, and the whole repository**

Run: `go build ./... && go test -count=1 ./hivepaas_app/service/specservice/... ./hivepaas_app/entity/ ./hivepaas_app/usecase/appsettingsuc/...`
Expected: every package `ok`

Run: `golangci-lint run ./... && go run ./tools/goroutinelint . && go run ./tools/errcodelint`
Expected: `0 issues.`; the two custom linters print nothing and exit 0

Run: `go test ./...`
Expected: every package `ok`

Run: `make gen-swag && git status --short | grep -v '^[AM]  '`
Expected: no output

Run: `git diff --stat prep/functions-in-spec -- . ':(exclude)docs/superpowers'`
Expected: no output

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(spec): an imported function goes through what creating one checks - kind, source, registry - and keeps its kind

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## After the plan (the user)

- Merge into `main` locally; run `go test ./...` in the main checkout.
- On the Linux server: export a project with an inline function, a repository function and a function's call; unpack the bundle and read a function's directory; import it into another env or installation: the functions are created, deployed and answer, their calls run. Import it again: no change planned. Edit a file under a function's directory and import: the function's source changes and it is deployed again when the options ask.
- The dashboard needs no change: the export's report no longer lists `FUNCTION_SKIPPED`, and the import's plan shows the new blocked issues as it shows the others.
