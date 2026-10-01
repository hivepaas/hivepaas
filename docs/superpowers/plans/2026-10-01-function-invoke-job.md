# Function Invoke Job Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Part 5 of functions. A scheduled job type, `function-invoke`, calls a function on a schedule: its settings are the request (method, path with its query, headers, body), it runs the runtime's `invoke` in a running task of the function through the container exec the `container-command` jobs use, the response is the run's output, and a status of 400 or more fails the run.

**Architecture:**
- **The job (Task 1):** `SchedJob.FunctionInvoke` holds the request; it lives at the app scope only, in its own app, which must be a function (its kind's category).
- **Invoke's input and output (Task 2):** a package `functioninvoke` writes the request as `invoke` reads it, and reads `invoke`'s standard output as it comes - the call's log, line by line, then the result line. Test runs (part 3) use it too.
- **The run (Task 3):** the job runner gives the request to the job exec service as the command's stdin (`SchedJobExecReq.Stdin`), takes its stdout into the output reader, puts the call's log into the run's log and the response into the task's output, or, as a step of a sequence, hands the status and body on. The task API answers the response as `functionInvoke`.
- **The dashboard (Tasks 4-5):** a function's Scheduled Jobs page has **New Function Call**; its form is the request and the scheduling; the list tags a call with its request; a run's page shows the response.

**Tech Stack:** the backend as it is (Go 1.27, bun, fx, swag); the dashboard as it is (React 19, react-hook-form with zod, TanStack Query).

**Spec:** `docs/superpowers/specs/2026-10-01-functions-design.md`, section 5 ("On a schedule") and part 5 of section 9, amended by this plan's commit (its last section, "Changes after part 5's plan"). The contract of `invoke` is `CONTRACT.md` of `hivepaas/function-runtimes` v1.0.0.

## How the patches work

Every task's code was written and verified before this plan.

**Two repositories.** Tasks 1-3 are this one (the backend), on a branch in a worktree from `main` at this plan's commit, which adds only documents to `15f5ca87` ("feat(apps): an app says its category, and a list keeps the categories asked for"). Tasks 4-5 are `hivepaas-dashboard`, on a branch in a worktree from its `main` at `a4fac9f4` ("feat(functions): a function shows no preview deployments, preview settings or data files"); a dashboard worktree needs `ln -s <the dashboard's checkout>/node_modules node_modules` (git-ignored).

Patches live in this repository, in `docs/superpowers/plans/2026-10-01-function-invoke-job/`: `fi-be-NN-tests.patch` then `fi-be-NN-code.patch` for backend task NN; `fi-ui-01.patch` (Task 4) and `fi-ui-02.patch` (Task 5). With `P=/Users/tnt/go/src/github.com/hivepaas/hivepaas/docs/superpowers/plans/2026-10-01-function-invoke-job`, commands run from the worktree's root.

The branches the patches were cut from are kept: `prep/function-invoke-job` in both repositories. After the last task of each, `git diff prep/function-invoke-job` (backend: `-- . ':(exclude)docs/superpowers'`) is empty.

**What was checked, on fresh worktrees:**
- backend, of `15f5ca87`, at every task: the tests alone fail to build, as step 2 says; with the code they pass and `go build ./...` passes; the task's lint finds 0 issues; Tasks 1 and 3: `errcodelint` clean and `make gen-swag` changes nothing; Task 3: `goroutinelint` clean;
- backend, after the last task, on the prep branch, whose tree the last task equals: `golangci-lint run ./...` 0 issues, `go test ./...` all `ok` (214 packages);
- dashboard, of `a4fac9f4`, after each task: `npm run lint:ci` exits 0 and `npm run build` succeeds;
- the JSON the backend sends for a function-invoke job and its run (printed from `TransformSchedJob` and `TransformTask`) was read against the dashboard's validators: `functionInvoke {method, path, headers, body}` on the job, `functionInvoke {outcome, status, body (base64), durationMs}` on the task, with `headers` and `body` left out when empty. It was not run through them: the validators' modules load the browser's globals.

**The dashboard has no unit tests**: the type checker, the linters and the build gate its tasks; the pages are checked in the browser, after the plan (the user).

**What it needs on the machine:** Go 1.27, golangci-lint 2.13, Docker for `make gen-swag`; Node.js and the dashboard's `node_modules`.

## Global Constraints

- **Names, verbatim:** job type `function-invoke`; `SchedJob.FunctionInvoke` as `functionInvoke`, with `method`, `path`, `headers` (by name, lower case, `{"name": ["value"]}`) and `body` (text); the run's result `{outcome, status, headers, body (base64), bodyTruncated, requestId, durationMs}`, the task API's `functionInvoke`; a step's outputs `STATUS` and `BODY`.
- **What a request may be:** a method among GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS (upper case; GET when left out); a path from `/` (added when left out), with its query, without a space or a control character, at most 2048 characters; at most 50 headers, each name an HTTP token, no value with a line break; a body of at most 1 MB, and none for GET or HEAD.
- **Where a call lives:** at the app scope only, in its own app (`app` is the scope's), and that app is a function (its kind's category is `function`). An env's sequence may run it as a step, as any app job.
- **What fails a run:** `invoke` exiting without a result (2: the request could not be read, 3: the handler could not be loaded), no result line, a result over 16 MB, an outcome other than `ok`, or a status of 400 or more. The response is kept in the task even when the status fails the run.
- **What a run keeps:** the response's body cut at 64 KB (`bodyTruncated`), in the task's output and in the run's log (as text when it is UTF-8); a step hands on `STATUS` and, when the body is text, `BODY` cut at 16 KB.
- **The backend's conventions** (`docs/ARCHITECTURE.md`), its linters (`golangci-lint run ./...`, `goroutinelint`, `errcodelint`), `make gen-swag` when a DTO changes, `go test ./...` in the main checkout after merging (it has `vendor/`), tests with `assert` and fakes, no database. **The dashboard's:** its module layout, `npm run lint:ci` read by its exit code before a commit.
- **Git:** commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`; merge locally, never push; never `git stash`.
- **Deploys:** nothing is deployed to the user's Docker Desktop swarm; the live checks run on the user's Linux server.

## Review Focus

1. **A function with no running task** (its service failed, or is being redeployed): the exec service finds no task and retries as for a container command, then the run fails saying so.
   - Pinned by: nothing new - the container exec's own task search; the live check (After the plan).
2. **A call on a worker node**: `invoke` runs through that node's agent, with the request on the exec's stdin; a backup restore's command already reads its stdin this way, and a function's call is the second to.
   - Pinned by: `TestSchedJobExecGivesTheCommandItsStdin` (Task 3), in process; the live check with a function on another node.
3. **A handler that prints megabytes of log, or answers a large body**: the log goes to the run's log line by line (a line over 64 KB in pieces), the result line is held up to 16 MB and the body kept cut at 64 KB.
   - Pinned by: `TestALongLogLineIsPassedOnInPieces`, `TestAResultTooLargeInPiecesIsSaid` (Task 2), `TestAFunctionCallsLargeBodyIsCut` (Task 3).
4. **A call that outlives its function's timeout**: `invoke` ends it with outcome `timeout`, which fails the run; the job's own timeout bounds the exec beside it.
   - Pinned by: `TestAFunctionCallThatTimesOutFails` (Task 3), on a written result; the live check with a handler that sleeps past its timeout.
5. **A clone of a function's app without its deployment settings**: the clone is not a function, but a call cloned with its jobs would run `hivepaas-runtime invoke` in an app without it, and fail at its first run; saving the job again is refused.
   - Pinned by: nothing; the live check, if clones of functions are used.

## Decisions beyond the spec

- **A call lives in its function** (the app scope only), as a data backup lives in its app; the env's Scheduled Jobs list shows it among the app jobs, and an env sequence runs it as a step.
- **The request's query is in its path** (`/report?day=today`) and its body is text, as the dashboard's test panel takes them; `invoke` gets them split, with the body in base64.
- **The command is `hivepaas-runtime invoke`**, run as the container's user, in its working directory (`/app`), with the job's environment beside the function's, through `SchedJobExec`: a step of a sequence gives the handler `HIVEPAAS_SEQ_*` in its environment.
- **The call's log streams into the run's log** as `invoke` writes it; the response is written after it: `The function answered <status> in <ms> ms (outcome <outcome>)`, then the body.
- **A step hands on `STATUS` and `BODY`**, the body only when it is text and cut at 16 KB: the steps after it read it in their environment.
- **Test runs read `invoke`'s result through the same package** (`functioninvoke.ParseResult`); `functiontest.Request` is now an alias of `functioninvoke.Request`.
- **No trigger runs a call** (a trigger's job types stay container commands and sequences), and **MCP does not plan one**: `plan_create_sched_job` says a function's call is made in the dashboard.
- **The dashboard offers New Function Call only on a function's Scheduled Jobs page**; a call's form has no triggers and no command output.

---

## Task 1: A function-invoke job - the request a function is called with, in its own app (backend)

**Files:**
- Create: `hivepaas_app/entity/setting_sched_job_function_invoke.go` (`SchedJobFunctionInvoke`), `hivepaas_app/usecase/settings/schedjobuc/schedjobdto/function_invoke.go` (`SchedJobFunctionInvokeReq`, its normalization and validation, `SchedJobFunctionInvokeResp`), `hivepaas_app/usecase/settings/schedjobuc/function_invoke.go` (`checkFunctionInvoke`, `checkFunctionInvokeApp`)
- Modify: `hivepaas_app/base/sched_job.go` (`SchedJobTypeFunctionInvoke`), `hivepaas_app/entity/setting_sched_job.go` (`SchedJob.FunctionInvoke`), `hivepaas_app/usecase/settings/schedjobuc/schedjobdto/{create.go,get.go}`, `hivepaas_app/usecase/settings/schedjobuc/{create.go,update.go,sequence.go}`, `hivepaas_app/interface/mcp/tools_sched_write.go` (the field's description), `docs/openapi/swagger.json`
- Test: `hivepaas_app/usecase/settings/schedjobuc/schedjobdto/function_invoke_test.go`, `hivepaas_app/usecase/settings/schedjobuc/function_invoke_test.go`

**Interfaces:**
- Produces: `base.SchedJobTypeFunctionInvoke = "function-invoke"` (in `AllSchedJobTypes`); `entity.SchedJobFunctionInvoke{Method, Path string; Headers map[string][]string; Body string}`; `SchedJob.FunctionInvoke *SchedJobFunctionInvoke` (`functionInvoke`); `SchedJobBaseReq.FunctionInvoke *SchedJobFunctionInvokeReq` and `SchedJobResp.FunctionInvoke *SchedJobFunctionInvokeResp`, same fields; `checkFunctionInvokeApp(scope *entity.ObjectScope, job *entity.SchedJob, kind *entity.Setting) error`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/fi-be-01-tests.patch`

Tests added: `TestAFunctionCallIsItsRequest`, `TestAFunctionCallDefaultsToAGetOfTheRoot`, `TestAFunctionCallRefusesWhatInvokeCannotSend`, `TestAFunctionCallIsOnlyAFunctionInvokesAndNeedsItsApp`, `TestFunctionCallsLiveInApps`, `TestAFunctionCallIsItsFunctionsOwn`. The existing `TestEveryToolBuildsAndDescribesItsArguments` (MCP) checks the new request field is described.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas_app/usecase/settings/schedjobuc/... ./hivepaas_app/interface/mcp/`
Expected: FAIL to build: `undefined: SchedJobFunctionInvokeReq`, `undefined: base.SchedJobTypeFunctionInvoke`, `unknown field FunctionInvoke in struct literal of type SchedJobBaseReq`, `undefined: checkFunctionInvokeApp`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/fi-be-01-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/usecase/settings/schedjobuc/... ./hivepaas_app/interface/mcp/`
Expected: every package `ok`

Run: `golangci-lint run ./hivepaas_app/usecase/settings/... ./hivepaas_app/entity/... ./hivepaas_app/base/... ./hivepaas_app/interface/mcp/... && go run ./tools/errcodelint`
Expected: `0 issues.`; errcodelint prints nothing

Run: `make gen-swag && git status --short | grep -v '^[AM]  '`
Expected: no output

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(sched-jobs): a function-invoke job - the request a function is called with, in its own app

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 2: invoke's request and its output, read as they come (backend)

**Files:**
- Create: `hivepaas_app/service/functionservice/functioninvoke/functioninvoke.go`
- Modify: `hivepaas_app/service/functionservice/functiontest/{functiontest.go,run.go}` (the marker, the request and the result read through `functioninvoke`)
- Test: `hivepaas_app/service/functionservice/functioninvoke/functioninvoke_test.go`

**Interfaces:**
- Consumes: Task 1's `entity.SchedJobFunctionInvoke`.
- Produces, in `functioninvoke`: `ResultMarker = "#hivepaas-result "`; `Command = []string{"hivepaas-runtime", "invoke"}`; `Request{Method, Path string; Query, Headers map[string][]string; Body []byte}`; `RequestOf(*entity.SchedJobFunctionInvoke) *Request` (the query read out of the path); `Result{Status int; Headers map[string][]string; Body []byte; RequestID string; DurationMs float64; Outcome string}`; `ParseResult(line string) (*Result, error)`; `NewOutputWriter(onLog func(line []byte)) *OutputWriter` with `Write`, `Flush()`, `Result() []byte`, `ResultTooLarge() bool`.
- `functiontest.Request` is an alias of `functioninvoke.Request`; test runs behave as before.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/fi-be-02-tests.patch`

Tests added: `TestAJobsRequestIsWhatInvokeReads`, `TestAResultLineIsTheResponse`, `TestInvokesOutputIsTheCallsLogThenItsResult` (chunks split mid-line and mid-marker), `TestAResultWithoutItsNewlineIsStillTheResult`, `TestOnlyALineThatStartsWithTheMarkerIsTheResult`, `TestALongLogLineIsPassedOnInPieces`, `TestAResultTooLargeToReadIsSaid`, `TestAResultTooLargeInPiecesIsSaid`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas_app/service/functionservice/...`
Expected: FAIL to build: `undefined: RequestOf`, `undefined: Request`, `undefined: ParseResult`, `undefined: OutputWriter`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/fi-be-02-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/service/functionservice/...`
Expected: every package `ok`, part 3's test-run tests included

Run: `golangci-lint run ./hivepaas_app/service/functionservice/...`
Expected: `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(functions): invoke's request and its output, read as they come - shared by test runs and jobs

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 3: A function-invoke job calls its function (backend)

**Files:**
- Create: `hivepaas_app/tasks/taskschedjobexec/function_invoke.go` (`invokeFunction`, `readInvokeResult`, `describeInvokeResult`, `invokeOutputs`, `invokeExitHint`)
- Modify: `hivepaas_app/entity/setting_sched_job_function_invoke.go` (`SchedJobFunctionInvokeResult`, `Task.OutputAsFunctionInvoke`), `hivepaas_app/service/schedjobexecservice/types.go` (`SchedJobExecReq.Stdin`), `hivepaas_app/service/schedjobexecservice/schedjobexecserviceimpl/exec.go` (the stdin to the container exec), `hivepaas_app/tasks/taskschedjobexec/job_run.go` (the type's case), `hivepaas_app/usecase/taskuc/taskdto/get.go` (`TaskResp.FunctionInvoke`), `docs/openapi/swagger.json`
- Test: `hivepaas_app/tasks/taskschedjobexec/function_invoke_test.go`, `hivepaas_app/entity/setting_sched_job_function_invoke_test.go`, `hivepaas_app/service/schedjobexecservice/schedjobexecserviceimpl/exec_writer_test.go`, `hivepaas_app/usecase/taskuc/taskdto/data_backup_test.go`

**Interfaces:**
- Consumes: Task 1's `SchedJob.FunctionInvoke`; Task 2's `RequestOf`, `Command`, `NewOutputWriter`, `ParseResult`.
- Produces: `entity.SchedJobFunctionInvokeResult{Outcome string; Status int; Headers map[string][]string; Body []byte; BodyTruncated bool; RequestID string; DurationMs float64}`; `(*Task).OutputAsFunctionInvoke() (*SchedJobFunctionInvokeResult, error)` (nil for an output without an outcome); `SchedJobExecReq.Stdin io.Reader`; `TaskResp.FunctionInvoke` (`functionInvoke`).

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/fi-be-03-tests.patch`

Tests added: `TestAFunctionCallSendsItsRequestAndKeepsTheResponse` (the command, its app, its stdin, the task's output, the outputs), `TestAFunctionCallAnswering400OrMoreFails`, `TestAFunctionCallThatTimesOutFails`, `TestAFunctionThatCannotBeLoadedFailsTheRun`, `TestAFunctionCallWithoutAResultFails`, `TestAFunctionCallStepHandsOnItsResponse`, `TestAFunctionCallsLargeBodyIsCut`, `TestAFunctionCallsRunKeepsItsResponse`, `TestSchedJobExecGivesTheCommandItsStdin`, `TestTransformTaskGivesAFunctionCallsResponse`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas_app/tasks/taskschedjobexec/ ./hivepaas_app/entity/ ./hivepaas_app/usecase/taskuc/taskdto/ ./hivepaas_app/service/schedjobexecservice/...`
Expected: FAIL to build: `undefined: SchedJobFunctionInvokeResult`, `OutputAsFunctionInvoke undefined`, `req.Stdin undefined`, `unknown field Stdin`, `undefined: invokeBodyMax`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/fi-be-03-code.patch`

- [ ] **Step 4: Run them to watch them pass, and the whole repository**

Run: `go build ./... && go test -count=1 ./hivepaas_app/tasks/taskschedjobexec/ ./hivepaas_app/entity/ ./hivepaas_app/usecase/taskuc/taskdto/ ./hivepaas_app/service/schedjobexecservice/...`
Expected: every package `ok`

Run: `golangci-lint run ./... && go run ./tools/goroutinelint . && go run ./tools/errcodelint`
Expected: `0 issues.`; the two custom linters print nothing and exit 0

Run: `go test ./...`
Expected: every package `ok`

Run: `make gen-swag && git status --short | grep -v '^[AM]  '`
Expected: no output

Run: `git diff --stat prep/function-invoke-job -- . ':(exclude)docs/superpowers'`
Expected: no output

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(sched-jobs): a function-invoke job calls its function - invoke in a running task, its response the run's

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 4: A function's call - its form, its route and its list tag (dashboard)

**Files** (under `src/application/`):
- Create: `modules/projects/module-shared/components/function-invoke-form/` (`function-invoke.form.schema.ts`, `function-invoke.form-mappers.ts`, `function-invoke.form.com.tsx`, `index.ts`); `modules/projects/routes/single-project/single-app/configuration/scheduled-jobs/function-invoke/` (`app-function-invoke-create.route.com.tsx`, `app-function-invoke-form-route.com.tsx`, `index.ts`)
- Modify: the job type enum (`FunctionInvoke: "function-invoke"`); the job entity, validator and contracts (`functionInvoke`); `modules/projects/module-shared/utils/function-test.utils.ts` (`parseHeaderLines`, `headerLinesOf`, which `buildTestRequest` now uses); the list route (**New Function Call**, for a function only); the edit route (the form by `jobType`); the list's name cell (`Call · <method> <path>`); the env jobs filter (the type); `shared/constants/route.constants.ts` (`createFunctionInvoke`); `projects.router.tsx`, `projects.module.ts`, `routes/index.ts`, and the components and scheduled-jobs indexes.

**Interfaces:**
- Consumes: the job API of Task 1; `isFunctionApp` and the app's `category` (part 4).
- Produces: `FunctionInvokeForm`; `FunctionInvokeFormSchema`, `FUNCTION_INVOKE_METHODS`, `methodSendsNoBody`, `headerLinesProblem`; `createEmptyFunctionInvokeFormDefaults()`, `mapFunctionInvokeToFormInput(job)`, `mapFunctionInvokeFormToPayload(values, appId)` (no body for GET and HEAD, headers left out when none); `AppFunctionInvokeFormRoute`, `AppFunctionInvokeCreateRoute`; `ROUTE.projects.single.apps.single.configuration.scheduledJobs.createFunctionInvoke`; `parseHeaderLines(text)`, `headerLinesOf(headers)`.

The form: name; **Request**: method, path with its query, headers one per line, body (not for GET and HEAD); **Scheduling**: priority, schedule with No schedule, timeout, retry and canceling; notification.

Without the job type in the enum, a list holding a function's call fails to parse: Task 4 is needed before a call is created.

- [ ] **Step 1: Apply the task's patch**

Run: `git apply --index $P/fi-ui-01.patch`

- [ ] **Step 2: Type-check, lint, build**

Run: `npm run lint:ci; echo "exit $?"`
Expected: `exit 0`

Run: `npm run build`
Expected: `✓ built in …`

- [ ] **Step 3: Commit**

```bash
git commit -m "feat(sched-jobs): a function's call - its form, its route and its list tag

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 5: A function call's run shows its response (dashboard)

**Files** (under `src/application/modules/operations/`):
- Create: `routes/tasks/building-blocks/task-summary-card/function-invoke-response.com.tsx` (`FunctionInvokeSummary`, `FunctionInvokeBody`)
- Modify: `domain/system-task.entity.ts` (`SystemTaskFunctionInvoke`, `SystemTask.functionInvoke`), `api/services/system-tasks-services/system-tasks.api.validator.ts` (its body decoded from base64), `routes/tasks/building-blocks/task-summary-card/system-task-summary-card.com.tsx`

**Interfaces:**
- Consumes: Task 3's `functionInvoke` on the task.
- Produces: `SystemTask.functionInvoke`. The run's details show `Response: <status> in <ms> ms` (red at 400 or more, or an outcome other than ok), and its body below them, as text or as a count of bytes, with `… cut at 64 KB` when cut. The app's and the project's task pages show the same card.

- [ ] **Step 1: Apply the task's patch**

Run: `git apply --index $P/fi-ui-02.patch`

- [ ] **Step 2: Type-check, lint, build, and the tree**

Run: `npm run lint:ci; echo "exit $?"`
Expected: `exit 0`

Run: `npm run build`
Expected: `✓ built in …`

Run: `git diff --stat prep/function-invoke-job`
Expected: no output

- [ ] **Step 3: Commit**

```bash
git commit -m "feat(tasks): a function call's run shows its response

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## After the plan (the user)

- Merge each branch into its `main` locally; in the backend's main checkout run `go test ./...`.
- On the Linux server, with images built from the merged branches, for a function of each runtime:
  1. **New Function Call** with `GET /?name=Ada`, Run now: the run succeeds, its log has the call's log line and `The function answered 200`, its page shows the response and `{"hello":"Ada"}`.
  2. A handler that answers 500: the run fails, the response is still on its page, the failure notification goes out.
  3. A handler that sleeps past its timeout: the run fails with outcome `timeout`.
  4. A POST with a JSON body and a header the handler echoes.
  5. A function whose instance runs on a node other than the manager (Review Focus 2).
  6. A job sequence: the call, then a container command that prints `$HIVEPAAS_SEQ_OUTPUTS`: `STATUS` and `BODY` are there.
  7. A function scaled to 0 replicas, or failing: the run fails, saying no task was found (Review Focus 1).
- In the browser: the button appears on a function's Scheduled Jobs page only; editing a call opens its form; the list's tag; the env Scheduled Jobs filter by the type.
- The functions feature is complete after this part; the spec's "Later" list (scale to zero, asynchronous calls, other triggers, metrics) is what remains.
