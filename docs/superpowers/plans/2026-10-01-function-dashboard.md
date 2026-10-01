# Function Dashboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Part 4 of functions. In the dashboard, a function is created with a **New Function** button, its page opens on a **Code** tab - its files in an editor beside a test panel, saved and deployed together - its settings take the place of its deployment settings, and the app list marks functions and filters on them. The backend adds what the list needs: an app's category, and a filter on it.

**Architecture:**
- **Backend (Task 1):** an app answers its `category` (its kind's, empty when it declares none); `GET …/apps` takes `category=function` or `category=webapp,database` (an app without a kind is a webapp); the MCP `list_apps` tool describes the parameter.
- **Dashboard, API (Task 2):** the function deployment method and its source in the deployment settings (read and saved), `POST …/apps/function`, `POST …/apps/{app}/function/test-run`, an app's `category` and the list's `category` filter.
- **Dashboard, pages (Tasks 3-7):** the Code tab (files, editor, Save & Deploy), the test panel inside it, the New Function dialog, the function's settings, the list's badge and filter.

**Tech Stack:** the backend as it is (Go 1.27, bun, swag); the dashboard as it is (React 19, TanStack Query, react-hook-form with zod, oxide.ts, rxjs, the shadcn components of `@/components/ui`, `react-simple-code-editor` with `prismjs`, both already dependencies).

**Spec:** `docs/superpowers/specs/2026-10-01-functions-design.md`, section 6 and part 4 of section 9, amended by this plan's commit (its last section, "Changes after part 4's plan"). Part 2's plan (`2026-10-01-functions-backend.md`) made `POST …/apps/function` and the function's deployment settings; part 3's (`2026-10-01-function-test-runs.md`) made the test run API.

## How the patches work

Every task's code was written and verified before this plan.

**Two repositories.** Task 1 is this one (the backend), on a branch in a worktree from `main` at this plan's commit, which adds only documents to `e5c6560c` ("feat(functions): the test run API"). Tasks 2-7 are `hivepaas-dashboard`, on a branch in a worktree from its `main` at `38639748` ("feat: Allow PR Comments in an app's App Preview settings"); a dashboard worktree needs `ln -s <the dashboard's checkout>/node_modules node_modules` (git-ignored).

Patches live in this repository, in `docs/superpowers/plans/2026-10-01-function-dashboard/`: `fd-be-01-tests.patch` then `fd-be-01-code.patch` for Task 1; `fd-ui-0N.patch` for dashboard task N+1 (`fd-ui-01.patch` is Task 2). With `P=/Users/tnt/go/src/github.com/hivepaas/hivepaas/docs/superpowers/plans/2026-10-01-function-dashboard`, commands run from the worktree's root.

The branches the patches were cut from are kept: `prep/function-dashboard-api` in the backend, `prep/function-dashboard` in the dashboard. After the last task of each, `git diff <prep branch>` (backend: `-- . ':(exclude)docs/superpowers'`) is empty.

**The dashboard has no unit tests**: no test runner is set up in it. Its tasks are checked by the type checker and the linters (`npm run lint:ci`) and by the build (`npm run build`); what they show is checked in the browser, after the plan (the user). Pure logic the pages rely on is in `module-shared/utils` and the forms' schema files, where a test runner would find it.

**What was checked, on fresh worktrees:**
- backend, of `e5c6560c`: the tests alone fail to build, as step 2 says; with the code they pass, `go build ./...` passes, `golangci-lint run ./...` finds 0 issues, `goroutinelint` and `errcodelint` are clean, `make gen-swag` changes nothing, and `go test ./...` is all `ok` (213 packages);
- dashboard, of `38639748`: after each task, `npm run lint:ci` exits 0 and `npm run build` succeeds.
- the starting templates (Task 5) were built with each runtime's image (v1.0.0) and called with `?name=Ada`: each answered 200 `{"hello":"Ada"}`.

**What it needs on the machine:** Go 1.27, golangci-lint 2.13, Docker for `make gen-swag`; Node.js and the dashboard's `node_modules`.

## Global Constraints

- **The API's names:** a function's source is `functionSource` in the deployment settings (`activeMethod: "function"`), with `code.repo.repoURL` on the wire (`repoUrl` in the dashboard's types); `POST /projects/{project}/{env}/apps/function` takes `name`, `note`, `tags`, `status`, `source` and answers `id`, `deploymentId`, `taskId`; `POST …/apps/{app}/function/test-run` takes `code.files` and `request` (`method`, `path`, `query`, `headers`, `body`) and answers the body in base64.
- **The runtimes:** `node24` (Node.js 24, `index.js` / `default`), `python313` (Python 3.13, `main.py` / `handler`), `go127` (Go 1.27, the package's directory `.` / `Handle`). The backend fills what a source leaves out: contract `v1`, the entrypoint, timeout `30s`, concurrency `16`, body size `6mb`.
- **The limits the forms check before the backend does:** a file's path is relative, plain (`[A-Za-z0-9._/-]`), inside the function and not under `.hivepaas`; concurrency 1-1000; a Debian package matches `^[a-z0-9][a-z0-9+.-]+(=[A-Za-z0-9.+~:-]+)?$`. Timeout (1s-15m) and body size (1kb-100mb) are checked by the backend, its errors shown on their fields.
- **Saving deploys:** Save & Deploy on the Code tab and on the settings sends the deployment settings back, which deploys the function.
- **The backend's conventions** (`docs/ARCHITECTURE.md`), its linters (`golangci-lint run ./...`, `goroutinelint`, `errcodelint`), `make gen-swag` when a DTO changes, `go test ./...` in the main checkout after merging (it has `vendor/`). **The dashboard's:** its module layout (`api/services` contracts, validator and api; `api/hooks`; `data/commands` and `data/queries`; `domain`; `module-shared`; `dialogs` with a zustand state; `routes`), `npm run lint:ci` read by its exit code before a commit.
- **Git:** commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`; merge locally, never push; never `git stash`.

## Review Focus

The dashboard's tests are the browser. What its pages meet that the type checker does not:

1. **Changing a function's runtime in its settings while its code is in the editor:** the code stays as it was, so the next deployment cannot load it; the form says so under the runtime.
   - Pinned by: nothing - the note is in `function-settings.form.com.tsx` (Task 6); the browser check.
2. **Moving a function's code between the editor and a repository:** to a repository, saving drops the code written in the editor (the form warns); back to the editor, the code starts from the runtime's template.
   - Pinned by: `functionSettingsToSource` (Task 6); the browser check.
3. **Two saves of the same function** - the Code tab and the settings, or two people: the second is refused by the deployment settings' version; the API's hook notifies it and the edits stay in the page.
   - Pinned by: nothing; the browser check.
4. **A large file in the editor** (inline code may reach 1 MB): `react-simple-code-editor` highlights on every keystroke.
   - Pinned by: nothing; the browser check with a file of a few hundred KB.
5. **Leaving the Code tab with unsaved edits** loses them: there is no prompt yet ("Changes not saved" shows above the editor).
   - Pinned by: nothing; a decision of this plan (below).

## Decisions beyond the spec

- **The app list's filter** is "All kinds / Functions / Apps": Apps are the other categories, an app without a kind among them. The backend answers each app's category, so a badge needs no second call.
- **The Code tab comes first** on a function's page, and **the test panel is inside it**, beside the editor (the user's choice): a test runs the files as the editor has them, not as saved. A function whose code is in a repository shows where, with a link to its settings.
- **The function's settings replace "Deployment Settings"**, under the name "Function", and "App Kind" is not shown for a function. They hold the runtime, the entrypoint, the Debian packages, the limits of a call, the code's place (the editor or a repository: URL, ref, directory, credentials) and the registry its image is pushed to. A function's run commands and deployment notifications are kept as saved, not shown.
- **New Function** sits beside New App and New From Template. Its dialog asks a name, an env, a runtime and the code's source: the runtime's starting template, or a repository (URL, ref, directory, credentials). The rest - entrypoint, limits - is the backend's default, edited in the settings afterwards. On success the function's Code tab opens while its first deployment runs.
- **A starting template per runtime** answers `?name=Ada` with `{"hello":"Ada"}` and logs the call: Node.js `package.json` and `index.js`, Python `main.py`, Go `go.mod` and `handler.go`.
- **The editor highlights** JavaScript (`.js`, `.mjs`, `.cjs`, `.ts`), Python, Go (`.go`, `go.mod`, `go.sum`) and JSON; other files are plain text.
- **The lock file a test run made** is added to the editor's files with one button ("Add to the code"), replacing a file of its name; it is saved with the next Save & Deploy.
- **The git credentials and registry pickers** of the deployment settings become shared components (`GitCredentialCombobox`, `PushToRegistryCombobox`), used by the old form, the dialog and the function's settings.
- **Not yet:** a prompt before leaving unsaved code; a function in a spec, a template or MCP's create tools; metrics.

---

## Task 1: An app says its category, and a list keeps the categories asked for (backend)

**Files:**
- Modify: `hivepaas_app/usecase/appuc/appdto/get.go` (`AppResp.Category`, `TransformAppCategory`), `hivepaas_app/usecase/appuc/appdto/list.go` (`ListAppReq.Category`, its validation), `hivepaas_app/usecase/appuc/list.go` (`appCategoryFilter`), `hivepaas_app/usecase/appuc/get.go` (the kind setting loaded with the routing settings), `hivepaas_app/interface/api/handler/apphandler/{get_in_env.go,get_in_project.go}` (the `category` parameter), `hivepaas_app/interface/mcp/tools_endpoints.go` (`list_apps` describes it), `docs/openapi/swagger.json` (`make gen-swag`)
- Test: `hivepaas_app/usecase/appuc/appdto/app_category_test.go`, `hivepaas_app/usecase/appuc/list_category_test.go`

**Interfaces:**
- Produces: `AppResp.Category base.AppCategory` (`json:"category,omitempty"`); `func TransformAppCategory(app *entity.App) base.AppCategory`; `ListAppReq.Category []base.AppCategory` (`mapstructure:"category"`, values in `base.AllAppCategories`); `appCategoryFilter(categories []base.AppCategory) bunex.SelectQueryOption`: `EXISTS` a kind setting of these categories, `OR NOT EXISTS` one when `webapp` is asked.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/fd-be-01-tests.patch`

Tests added: `TestAnAppSaysWhatItIs`, `TestAListIsFilteredByKnownCategories`, `TestAListOfFunctionsIsTheAppsWhoseKindSaysSo` (the SQL the filter renders), `TestAnAppWithoutAKindIsAWebApp`. The existing `TestEveryToolBuildsAndDescribesItsArguments` (MCP) checks that `list_apps` describes the new parameter.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas_app/usecase/appuc/... ./hivepaas_app/interface/mcp/`
Expected: FAIL to build: `function.Category undefined (type *AppResp has no field or method Category)`, `unknown field Category in struct literal of type ListAppReq`, `undefined: appCategoryFilter`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/fd-be-01-code.patch`

- [ ] **Step 4: Run them to watch them pass, and the whole repository**

Run: `go build ./... && go test -count=1 ./hivepaas_app/usecase/appuc/... ./hivepaas_app/interface/mcp/`
Expected: every package `ok`

Run: `golangci-lint run ./... && go run ./tools/goroutinelint . && go run ./tools/errcodelint`
Expected: `0 issues.`; the two custom linters print nothing and exit 0

Run: `go test ./...`
Expected: every package `ok`

Run: `make gen-swag && git status --short | grep -v '^[AM]  '`
Expected: no output

Run: `git diff --stat prep/function-dashboard-api -- . ':(exclude)docs/superpowers'`
Expected: no output

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(apps): an app says its category, and a list keeps the categories asked for

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 2: The API of functions (dashboard)

**Files** (under `src/application/modules/projects/`):
- Create: `module-shared/enums/e.function-runtime.ts`; `domain/apps/function/{app-function.entity.ts,index.ts}`; `api/services/project-apps-services/function/{app-function.api.contracts.ts,app-function.api.validator.ts,app-function.api.ts,index.ts}`; `api/hooks/project-apps/use-app-function.api.ts`; `data/commands/project-apps/app-function.commands.ts`
- Modify: `module-shared/enums/{e.app-deployment-method.ts,index.ts}`; `domain/apps/{index.ts,base/app.details.entity.ts,deployment-settings/app-deployment-settings.entity.ts}`; `api/services/project-apps-services/deployment-settings/app-deployment-settings.api.{contracts,validator}.ts` and `.api.ts`; `api/services/project-apps-services/project-apps/project-apps.api.{contracts,schemas,validator}.ts` and `.api.ts`; `api/services/project-apps-services/index.ts`; `api/api-context/projects.api.context.ts`; `api/hooks/project-apps/{index.ts,use-project-apps.api.ts}`; `data/commands/project-apps/{index.ts,project-apps.commands.ts}`; `routes/single-project/single-app/configuration/deployment-settings/form/app-config-deployment-settings.form.com.tsx` and `…/route/app-config-deployment-settings.route.com.tsx` (the repository and image form takes only their settings)

**Interfaces:**
- Produces:
  - `EAppDeploymentMethod.Function = "function"`; `EFunctionRuntime` (`Node24`, `Python313`, `Go127`), `ALL_FUNCTION_RUNTIMES`, `FUNCTION_RUNTIME_LABELS`, `FUNCTION_RUNTIME_DEFAULT_ENTRYPOINT: Record<EFunctionRuntime, {file, handler}>`;
  - in `~/projects/domain`: `APP_CATEGORY_FUNCTION = "function"`, `FunctionFile {path, content}`, `FunctionRepoCode`, `FunctionSource` (`runtime`, `contract`, `entrypoint`, `code {inline: {files} | null, repo: FunctionRepoCode | null, dir}`, `systemPackages`, `timeout: string`, `maxConcurrency: number`, `maxBodySize: string`, `pushToRegistry`), `FunctionTestRequest`, `FunctionTestOutcome`, `FunctionTestRunResult` (`body: Uint8Array`, `lockFiles: FunctionFile[]`, …); `FunctionMethod = BaseDeploymentSettings & {activeMethod: "function"; functionSource: FunctionSource}`; `ProjectAppDetails.category: string`;
  - `FunctionSourcePayload` (contracts), `functionSourceToJson(source)` (`repoUrl` → `repoURL`), `FunctionSourceSchema`;
  - `ProjectApps_FindManyPaginated_Req.category?: string[]` (sent joined by `,`); `ProjectApps_CreateFunction_Req` (`projectID`, `name`, `env`, `note`, `tags`, `source`) and `_Res` (`id`, `deploymentId`, `taskId`); `ProjectAppsCommands.useCreateFunction`;
  - `AppFunctionApi.testRun`, `AppFunctionCommands.useTestRun()` (`{projectID, env, appID, files, request}`).

- [ ] **Step 1: Apply the task's patch**

Run: `git apply --index $P/fd-ui-01.patch`

- [ ] **Step 2: Type-check, lint, build**

Run: `npm run lint:ci; echo "exit $?"`
Expected: `exit 0`

Run: `npm run build`
Expected: `✓ built in …`

- [ ] **Step 3: Commit**

```bash
git commit -m "feat(functions): the API of functions - creating one, its source, its test runs, an app's category

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 3: A function's Code tab - its files in an editor, saved and deployed together (dashboard)

**Files** (under `src/application/modules/projects/` unless written in full):
- Create: `module-shared/utils/function-source.utils.ts`; `module-shared/components/function-code-editor/{function-code-editor.com.tsx,index.ts}`; `routes/single-project/single-app/tabs/code/` (`index.ts`, `route/{app-code.route.com.tsx,index.ts}`, `building-blocks/{function-files.com.tsx,function-code-workspace.com.tsx,index.ts}`)
- Modify: `module-shared/utils/index.ts`, `module-shared/components/index.ts`, `routes/index.ts`, `routes/single-project/single-app/tabs/index.ts`, `projects.module.ts`, `projects.router.tsx`, `layouts/single-app/header/single-app-header.com.tsx` (the Code tab, first, for a function), `src/application/shared/constants/route.constants.ts` (`ROUTE.projects.single.apps.single.code`)

**Interfaces:**
- Consumes: Task 2's `FunctionMethod`, `FunctionSource`, `FunctionSourcePayload`, `APP_CATEGORY_FUNCTION`.
- Produces: `isFunctionApp(app: {category: string}): boolean`; `FUNCTION_RESERVED_DIR = ".hivepaas"`; `functionFilePathProblem(path, takenPaths): string | null`; `languageOfPath(path): "javascript" | "python" | "go" | "json" | "plain"`; `functionSourceToPayload(source, files?)` (with files: inline code of those files, no repository, no directory); `functionSettingsToPayload(settings: FunctionMethod, source: FunctionSourcePayload)`; `FunctionCodeEditor({path, value, onChange, readOnly})`; `ROUTE.projects.single.apps.single.code.$route(id, env, appId)`; `FunctionCodeWorkspace` with a render-prop child `(files, setFiles) => ReactNode`.

- [ ] **Step 1: Apply the task's patch**

Run: `git apply --index $P/fd-ui-02.patch`

- [ ] **Step 2: Type-check, lint, build**

Run: `npm run lint:ci; echo "exit $?"`
Expected: `exit 0`

Run: `npm run build`
Expected: `✓ built in …`

- [ ] **Step 3: Commit**

```bash
git commit -m "feat(functions): a function's Code tab - its files in an editor, saved and deployed together

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 4: A test panel beside the code (dashboard)

**Files** (under `src/application/modules/projects/`):
- Create: `module-shared/utils/function-test.utils.ts`; `routes/single-project/single-app/tabs/code/building-blocks/function-test-panel.com.tsx`
- Modify: `module-shared/utils/index.ts`; `routes/single-project/single-app/tabs/code/building-blocks/index.ts`; `routes/single-project/single-app/tabs/code/route/app-code.route.com.tsx` (the panel as the workspace's child)

**Interfaces:**
- Consumes: Task 2's `AppFunctionCommands.useTestRun`, `FunctionTestRunResult`; Task 3's `FunctionCodeWorkspace` render prop.
- Produces: `buildTestRequest(method, pathWithQuery, headerLines, body): FunctionTestRequest` (the query is everything after the first `?`; headers one per line as `name: value`, names lower-cased; no body for GET and HEAD); `describeBody(body, headers): {text, isJson}` (JSON indented, text as is, other bytes counted); `withLockFiles(files, lockFiles)` (each added, replacing the file of its name).

- [ ] **Step 1: Apply the task's patch**

Run: `git apply --index $P/fd-ui-03.patch`

- [ ] **Step 2: Type-check, lint, build**

Run: `npm run lint:ci; echo "exit $?"`
Expected: `exit 0`

Run: `npm run build`
Expected: `✓ built in …`

- [ ] **Step 3: Commit**

```bash
git commit -m "feat(functions): a test panel beside the code - a request, the answer, the logs, the lock file

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 5: A New Function button (dashboard)

**Files** (under `src/application/modules/projects/`):
- Create: `module-shared/constants/function-templates.constants.ts`; `module-shared/components/git-credential-combobox/{git-credential-combobox.com.tsx,index.ts}`; `dialogs/create-function/` (`index.ts`; `types/`, `hooks/` (zustand state and hook), `schemas/`, `form/`, `dialog/`, each with its `index.ts`)
- Modify: `module-shared/constants/index.ts`, `module-shared/components/index.ts`; `dialogs-container/dialogs-container.com.tsx`; `routes/single-project/apps/building-blocks/project-apps-table/project-apps-table.com.tsx` (the button, guarded like New App); `routes/single-project/single-app/configuration/deployment-settings/form-components/git-credential-select/git-credential-select.com.tsx` (now the shared combobox in its form field)

**Interfaces:**
- Consumes: Task 2's `ProjectAppsCommands.useCreateFunction`, `FunctionSourcePayload`, the runtime enums; Task 3's `functionFilePathProblem`, `ROUTE…code`.
- Produces: `FUNCTION_TEMPLATES: Record<EFunctionRuntime, FunctionFile[]>`; `GitCredentialCombobox({projectId, env, value, onChange, readOnly?, invalid?, className?})`, `GitCredentialLinks({projectId})`, `type GitCredentialOption`; `useCreateFunctionDialog(props).actions.open(projectId)`; `EFunctionCodeSource` (`template`, `repository`); `createFunctionSource(values): FunctionSourcePayload` (the template's files, or the repository; entrypoint and limits left to the backend).

- [ ] **Step 1: Apply the task's patch**

Run: `git apply --index $P/fd-ui-04.patch`

- [ ] **Step 2: Type-check, lint, build**

Run: `npm run lint:ci; echo "exit $?"`
Expected: `exit 0`

Run: `npm run build`
Expected: `✓ built in …`

- [ ] **Step 3: Commit**

```bash
git commit -m "feat(functions): a New Function button - a name, an env, a runtime, and its template or a repository

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 6: A function's settings in place of its deployment settings (dashboard)

**Files** (under `src/application/modules/projects/`):
- Create: `module-shared/components/push-to-registry-combobox/{push-to-registry-combobox.com.tsx,index.ts}`; `routes/single-project/single-app/configuration/function-settings/` (`index.ts`, `function-settings.com.tsx`, `form/{function-settings.form.com.tsx,index.ts}`, `schemas/{function-settings.form.schema.ts,index.ts}`)
- Modify: `module-shared/components/index.ts`; `routes/single-project/single-app/configuration/index.ts`; `routes/single-project/single-app/configuration/deployment-settings/route/app-config-deployment-settings.route.com.tsx` (a function's settings are `FunctionSettings`); `…/deployment-settings/form-components/push-to-registry-select/push-to-registry-select.com.tsx` (the shared combobox); `layouts/single-app/configuration-layout/single-app-configuration-layout.com.tsx` ("Function" in place of "Deployment Settings", no "App Kind", for a function)

**Interfaces:**
- Consumes: Task 2's `FunctionMethod`, `FunctionSource`, `AppDeploymentSettingsCommands.useUpdateOne`; Task 3's `functionSourceToPayload`, `functionSettingsToPayload`, `functionFilePathProblem`, `isFunctionApp`; Task 5's `FUNCTION_TEMPLATES`, `GitCredentialCombobox`.
- Produces: `PushToRegistryCombobox({projectId, env, value, onChange, …})`, `RegistryCredentialsLink`; `FunctionSettings({projectId, env, appId, settings})`; `functionSettingsDefaultValues(source)`, `functionSettingsToSource(values, source)` (the code stays where it is unless the form moves it; a repository's commit is kept only while its URL and ref are), `functionSettingsErrorField(path)` (a backend error's `functionSource.…` path to the form's field).

- [ ] **Step 1: Apply the task's patch**

Run: `git apply --index $P/fd-ui-05.patch`

- [ ] **Step 2: Type-check, lint, build**

Run: `npm run lint:ci; echo "exit $?"`
Expected: `exit 0`

Run: `npm run build`
Expected: `✓ built in …`

- [ ] **Step 3: Commit**

```bash
git commit -m "feat(functions): a function's settings in place of its deployment settings - runtime, entrypoint, limits, packages, code's place

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 7: A function badge in the app list, and a filter on the kind of app (dashboard)

**Files** (under `src/application/modules/projects/`):
- Modify: `module-shared/definitions/tables/project-apps/project-apps-table.defs.tsx` (the badge); `routes/single-project/apps/building-blocks/project-apps-table/project-apps-table.com.tsx` (the filter beside the search)

**Interfaces:**
- Consumes: Task 1's `category` parameter; Task 2's `category` request field and `ProjectAppDetails.category`; Task 3's `isFunctionApp`; `ALL_APP_CATEGORIES` (the kind settings' categories).

- [ ] **Step 1: Apply the task's patch**

Run: `git apply --index $P/fd-ui-06.patch`

- [ ] **Step 2: Type-check, lint, build, and the tree**

Run: `npm run lint:ci; echo "exit $?"`
Expected: `exit 0`

Run: `npm run build`
Expected: `✓ built in …`

Run: `git diff --stat prep/function-dashboard`
Expected: no output

- [ ] **Step 3: Commit**

```bash
git commit -m "feat(functions): a function badge in the app list, and a filter on the kind of app

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## After the plan (the user)

- Merge each branch into its `main` locally; in the backend's main checkout run `go test ./...`.
- In the browser, against a backend with this plan's Task 1:
  - New Function of each runtime from its template: the Code tab opens; the first deployment succeeds; a test run with `?name=Ada` answers 200 `{"hello":"Ada"}` with its log line; the function's route answers the same.
  - A Node.js function with a dependency in `package.json`: the first test run installs it and offers `package-lock.json`; "Add to the code", Save & Deploy, and the build uses it.
  - New Function from a private repository with credentials, with a directory; the Code tab shows where the code is.
  - The settings: change the timeout, add a Debian package, a wrong value (a timeout of `20m`) shown on its field; move the code to a repository and back.
  - The app list: the `function` badge, the filter's three choices, with an app of no kind among the Apps.
  - The Review Focus items above.
- Part 5 (scheduled calls, the `function-invoke` job type) starts from here.
