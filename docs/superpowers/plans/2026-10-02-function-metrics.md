# Function Metrics Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A function's Metrics tab shows its calls, failures and durations over 1 h, 6 h, 24 h or 7 d (24 h by default), counted by VictoriaLogs from the invocation line its runtime writes for every call.

**Architecture:**
- **The backend client (Task 1):** `loggingmodel.InvocationStatsReq/Resp` and `Backend.InvocationStats`; `victorialogs.BuildInvocationStatsQueries` writes three LogsQL queries - by step, in all, by outcome - by `BuildQuery`'s rules (values quoted, unpacked fields under `app.`); `Client.InvocationStats` runs them through `post`, now shared with `Query`, and reads the rows - buckets sorted, quantiles that are no number left nil.
- **The service and the API (Task 2):** `loggingservice.FunctionMetrics` scopes the request to the app and fills the steps without a call; `appuc.GetFunctionMetrics` refuses an app that is not a function, answers the History tab's reason when the logs cannot be read, and computes the window - whole steps, ending with the step now is in, started at the logs' retention when the range is longer; `GET /projects/{id}/{env}/apps/{appId}/function-metrics?range=`.
- **The dashboard (Task 3):** the API's contract, validator, service, hook and query; the route `…/apps/:appId/metrics` and its tab beside Code; range buttons remembered in the browser; totals, a chart of calls (succeeded and failed, stacked), a chart of p50/p95/p99, the counts by outcome; `recharts` 3.10.1.

**Tech Stack:** Go 1.27, VictoriaLogs v1.52.0 (LogsQL `stats`), React, `recharts`, yarn.

**Spec:** `docs/superpowers/specs/2026-10-02-function-metrics-design.md`, amended by this plan's commit (its last section, "Changes after the plan").

## How the patches work

All three tasks were written and verified before this plan.

| Task | Repository | Base |
|---|---|---|
| 1, 2 | `hivepaas` (this one) | `main` at this plan's commit, which adds only documents to `812d54ce` |
| 3 | `hivepaas-dashboard` | `main` at `6013005e` |

Patches live in `docs/superpowers/plans/2026-10-02-function-metrics/`: `metrics-NN-tests.patch` then `metrics-NN-code.patch` for Tasks 1 and 2, `metrics-03-code.patch` for the dashboard. With `P=/Users/tnt/go/src/github.com/hivepaas/hivepaas/docs/superpowers/plans/2026-10-02-function-metrics`, commands run from the worktree's root.

**The dashboard worktree installs its own dependencies** - `yarn install --frozen-lockfile` after the patch, which adds `recharts` - rather than linking the main checkout's `node_modules`, so that nothing is installed there before the merge. After the merge, the main checkout needs `yarn install` once: the user's to run, or to allow.

The branches the patches were cut from are kept as `prep/function-metrics` in both repositories. After the last task, `git diff prep/function-metrics -- . ':(exclude)docs/superpowers'` is empty in each.

**What was checked, on fresh worktrees of the bases above:**
- Task 1: the tests alone do not build (`undefined: loggingmodel.InvocationStatsReq`); with the code `services/logging/...` passes, `go vet ./...` is clean (the service's test fake embeds the interface).
- Task 2: the tests alone do not build (`s.FunctionMetrics undefined`, `GetFunctionMetricsReq`, `metricsWindow`); with the code the logging service, `appuc` and `services/logging` pass, `golangci-lint run ./...` finds 0 issues, the custom lints are clean, `make gen-swag` changes only what the patch holds. On the prep branch, `go test ./...` 215 `ok`.
- The live tests (`TestLiveQueryCannotLeaveItsScope`, `TestLiveInvocationStatsCountOnlyTheAppsInvocationLines`), against VictoriaLogs v1.52.0 in a container: pass.
- Task 3: `yarn install --frozen-lockfile`, `npm run lint:ci`, `npm run build` pass.

**What it needs on the machine:** Go 1.27, golangci-lint 2.13, Docker (`make gen-swag`; the live tests, optionally), Node.js and yarn 1.22.

## Global Constraints

- **A call** is an invocation line (`"hp":"invocation"`) in the app's logs; **failed** is an outcome other than `ok`; 5xx answers are counted apart; durations are `durationMs`'s p50, p95, p99, close rather than exact.
- **Ranges and steps:** `1h` by 1 min, `6h` by 5 min, `24h` by 15 min, `7d` by 1 h; `24h` by default; nothing else accepted.
- **Every LogsQL query** is built in `services/logging/victorialogs`, values in Go-quoted literals, unpacked fields under `app.`; the app's id scopes it and nothing in the request widens it.
- **When the logs cannot be read**, the answer is the History tab's reason, not an error.
- **Who may read the app's logs may read its metrics**; the app must be a function.
- **Dashboard:** yarn for dependencies; `lint:ci` and `build`.
- **Git:** commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`; merge locally, never push; never `git stash`.

## Review Focus

1. **A function with hundreds of thousands of calls a day over 7 d**: three queries over some 2 million lines; measured at 260 ms for 300,000 lines over 24 h, so a second or two; the query timeout is the client's 30 s.
   - Pinned by: nothing; the spike's measure, in the spec.
2. **A handler that prints a line like an invocation line**: counted, for its own function only; a line naming another app does not reach that app's metrics.
   - Pinned by: `TestLiveInvocationStatsCountOnlyTheAppsInvocationLines` (live).
3. **A step boundary and VictoriaLogs' buckets**: both are aligned to the epoch for 1 min, 5 min, 15 min and 1 h, so a filled step and a backend bucket meet; a step that did not divide a day would not.
   - Pinned by: `TestFunctionMetricsAreTheFunctionsAndEveryStepIsAPoint`, `TestAMetricsWindowEndsWithTheStepNowIsIn`.
4. **Logs kept for less than the range**: the window starts at the first whole step the logs hold, and the tab says the charts start there.
   - Pinned by: `TestAMetricsWindowStopsWhereTheLogsDo`.
5. **A function on runtimes before 1.2.0**: its scheduled calls go through `invoke` and are not counted; its HTTP calls are.
   - Pinned by: nothing; the spec says so.

## Decisions beyond the spec

- **`recharts` used directly**, not through shadcn/ui's `chart` wrapper, which the dashboard does not have.
- **yarn, not npm**, for the new dependency: `yarn.lock` is the lock file kept up to date.
- **`LOG_HISTORY_UNAVAILABLE_TEXT`** moves to `log-history-unavailable.constants.ts`, shared by History and Metrics (a component file that exports a constant breaks fast refresh).
- **`Client.post`** is shared by `Query` and `InvocationStats`; its handling of errors is `Query`'s.
- **The window ends with the step now is in**, so the last point is the current, partial step.

---

## Task 1: Invocation stats in the logging backend (backend)

**Files:**
- Create: `services/logging/victorialogs/stats.go`, `services/logging/victorialogs/stats_test.go`, `services/logging/victorialogs/stats_live_test.go`
- Modify: `services/logging/loggingmodel/types.go` (`InvocationStatsReq`, `InvocationCounts`, `InvocationBucket`, `InvocationStatsResp`), `services/logging/loggingmodel/service.go` (`Backend.InvocationStats`), `services/logging/logging.go` (their aliases), `services/logging/victorialogs/client.go` (`post`)

**Interfaces:**
- Produces: `victorialogs.BuildInvocationStatsQueries(req) (*InvocationStatsQueries, error)` with `Series`, `Totals`, `Outcomes`; `(*Client) InvocationStats(ctx, req) (*InvocationStatsResp, error)`.

- [ ] **Step 1: Write the failing tests**

```bash
git apply --index $P/metrics-01-tests.patch
```

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./services/logging/victorialogs/`
Expected: FAIL to build - `undefined: loggingmodel.InvocationStatsReq`.

- [ ] **Step 3: Write the code**

```bash
git apply --index $P/metrics-01-code.patch
```

- [ ] **Step 4: Run them to watch them pass**

Run: `go test -count=1 ./services/logging/... && go vet ./...`
Expected: `ok`, four packages; vet silent.

Optionally, live:
```bash
docker run -d --rm --name vl-live -p 127.0.0.1:19428:9428 victoriametrics/victoria-logs:v1.52.0
HP_TEST_VICTORIALOGS_URL=http://127.0.0.1:19428 go test -count=1 -run Live ./services/logging/victorialogs/
docker rm -f vl-live
```
Expected: `ok`.

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(logging): invocation stats - the LogsQL that counts a function's calls, failures and durations, and the client that reads them

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

## Task 2: A function's metrics through the API (backend)

**Files:**
- Create: `hivepaas_app/service/loggingservice/loggingserviceimpl/metrics.go`, `hivepaas_app/usecase/appuc/function_metrics.go`, `hivepaas_app/usecase/appuc/appdto/function_metrics.go`, and their tests
- Modify: `hivepaas_app/service/loggingservice/service.go` (`FunctionMetrics`), `types.go` (`FunctionMetricsQuery`), `loggingserviceimpl/query.go` (`queryBackend`), `hivepaas_app/interface/api/handler/apphandler/logs.go` (`GetFunctionMetrics`), `hivepaas_app/interface/api/server/router_apps.go`, `docs/openapi/swagger.json`

**Interfaces:**
- Consumes: Task 1's `Backend.InvocationStats`.
- Produces: `GET /projects/{projectID}/{projectEnv}/apps/{appID}/function-metrics?range=` answering `appdto.GetFunctionMetricsResp`.

- [ ] **Step 1: Write the failing tests**

```bash
git apply --index $P/metrics-02-tests.patch
```

- [ ] **Step 2: Run them to watch them fail**

Run: `go vet ./hivepaas_app/usecase/appuc/... ./hivepaas_app/service/loggingservice/...`
Expected: FAIL - `s.FunctionMetrics undefined`, `undefined: GetFunctionMetricsReq`, `undefined: metricsWindow`.

- [ ] **Step 3: Write the code**

```bash
git apply --index $P/metrics-02-code.patch
```

- [ ] **Step 4: Run them to watch them pass, and the linters**

Run:
```bash
go build ./... && go test -count=1 ./hivepaas_app/service/loggingservice/... ./hivepaas_app/usecase/appuc/... ./services/logging/...
golangci-lint run ./...
go run ./tools/goroutinelint . && go run ./tools/errcodelint
make gen-swag && git status --short
```
Expected: `ok`; 0 issues; clean; only the task's files.

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(functions): GET .../function-metrics - a function's calls, failures and durations over 1h, 6h, 24h or 7d, from its logs

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

## Task 3: The Metrics tab (dashboard)

**Files:**
- Create: `src/application/modules/projects/routes/single-project/single-app/tabs/metrics/` (`route/app-metrics.route.com.tsx`, `building-blocks/metrics-charts.com.tsx`, `building-blocks/metrics-range.ts`), `.../tabs/logs/building-blocks/log-history-unavailable.constants.ts`
- Modify: the logs API (`app-logs.api.contracts.ts`, `app-logs.api.ts`, `app-logs.api.validator.ts`), `use-app-logs.api.ts`, `app-logs.queries.ts`, `projects.query-keys.ts`, `route.constants.ts` (`metrics`), `projects.router.tsx`, `projects.module.ts`, `routes/index.ts`, `tabs/index.ts`, `single-app-header.com.tsx` (the Metrics link for functions), `app-logs-history.com.tsx`, `package.json`, `yarn.lock`

**Interfaces:**
- Consumes: Task 2's API.
- Produces: `AppLogsQueries.useGetFunctionMetrics({ projectID, env, appID, range })`; `ROUTE.projects.single.apps.single.metrics`.

- [ ] **Step 1: Write the code, and install what it adds**

```bash
git apply --index $P/metrics-03-code.patch
yarn install --frozen-lockfile
```

- [ ] **Step 2: Check it**

Run: `npm run lint:ci && npm run build`
Expected: both pass.

- [ ] **Step 3: Commit**

```bash
git commit -m "feat(functions): a function's Metrics tab - calls, failures and durations over 1h, 6h, 24h or 7d, charted with recharts

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

## Finish

Merge both branches into their `main`, locally; backend `main` (it has `vendor/`): `go test ./...`; dashboard `main`: `yarn install` (the user's, or with the user's leave: it installs `recharts` in the main checkout's `node_modules`), then `lint:ci`, `build`. The browser check is the user's: a function's Metrics tab, its ranges, and the reason when logging is off.
