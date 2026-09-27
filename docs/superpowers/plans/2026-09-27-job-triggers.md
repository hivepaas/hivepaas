# Job Triggers Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Scheduled jobs gain `triggers`: events of an app - `pre-deploy`, `post-deploy`, `deploy-failed`, `health-down`, `health-up`, `app-enabled`, `app-disabled` - that run them, a `pre-deploy` trigger able to hold the deploy until its run ends.

**Architecture:** A job's `Triggers` list events and, for an env's job, the apps they come from. A new `schedjobtriggerservice` finds the jobs listening to an event of an app and schedules a `sched-job-exec` run of each, in a transaction of its own, recording the cause in the task's args. The deploy calls it before applying the new version and waits for the runs that hold it; the deploy's after-commit hook, the health check and the status use cases call it for the other events.

**Tech Stack:** Go (bun, fx), the task queue in `hivepaas_app/tasks/queue`; React 19, react-hook-form and zod in `hivepaas-dashboard`.

**Spec:** `docs/superpowers/specs/2026-09-27-job-triggers-design.md`

## How the patches work

Every task's code was written and verified before this plan: the backend patches apply in order to `main` at `fb4e9f38` (this plan's commit changes only `docs/`), the dashboard patches to the dashboard's `main` at `7fe4fbc9`. Applied in order they give a tree that builds, passes `go test ./...` and `golangci-lint run ./...` (0 issues), and, for the dashboard, `tsc`, `npm run lint` and `prettier --check`. The JSON the backend sends for a job's triggers and a run's cause was parsed with the dashboard's validators.

Patches live in `docs/superpowers/plans/2026-09-27-job-triggers/`:

- `be-NN-tests.patch`, then `be-NN-code.patch` for backend task NN: the tests first, watched failing, then the code.
- `dash-NN.patch` for dashboard task NN; its gate is the type check, the lint and, in Task 9, the browser.

`P=docs/superpowers/plans/2026-09-27-job-triggers` in the commands below, run from the backend repo root; `D=../hivepaas-dashboard`. If a base has moved since, use `git apply --3way` and resolve.

## Global Constraints

- Backend, before a task is done: `go build ./...`, `golangci-lint run ./...` over the **whole** repo (120-character lines, US spelling), `go test ./...`. `go test ./hivepaas_app/cmd/...` includes the fx wiring test, which catches a dependency cycle.
- `make gen-swag` when a DTO changes (Tasks 2 and 6); `docs/openapi/swagger.json` is generated and committed. A first run failing with `cannot find all dependencies` is run again; then `go mod verify` must say `all modules verified` - an interrupted run once left modules half-extracted in `~/go/pkg/mod`.
- Dashboard, before a task is done: `npx tsc --noEmit`, `npm run lint` (`--max-warnings 0`), `npx prettier --check src`.
- Events, verbatim: `pre-deploy`, `post-deploy`, `deploy-failed`, `health-down`, `health-up`, `app-enabled`, `app-disabled`. At most 10 triggers a job. `wait` for `pre-deploy` only.
- Configuration: `HP_TASKS_TRIGGERS_WAIT_POLL_INTERVAL` default `3s`, `HP_TASKS_TRIGGERS_WAIT_TIMEOUT` default `30m`.
- A run is told `HIVEPAAS_TRIGGER_EVENT`, `HIVEPAAS_TRIGGER_APP`, `HIVEPAAS_TRIGGER_DEPLOYMENT`; the prefix is `HIVEPAAS_`, never `HP_`.
- `preDeploymentCommand` and `postDeploymentCommand` stay as they are.
- No backward compatibility is needed.
- Commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`. Merge locally, never push. Never `git stash`; never touch the user's uncommitted work.
- Nothing is deployed to the user's Docker Desktop swarm; the live checks of Task 9 run on the user's Linux server.

## Review Focus

1. **A deploy waiting for its pre-deploy jobs** holds a worker of the task queue and the deploy's transaction, with its deployment row locked, for up to the wait's timeout. Another deploy of the same app queued meanwhile must wait its turn and then run, not fail or deadlock. Task 9, check 3.
2. **A waited job of the app itself** runs in the app's running container - the old version - while the deploy waits, and the new version is applied only after it succeeds. Task 9, check 2.
3. **A restart of HivePaaS** must fire no `health-up`: the last state lives in the cache for a day and the first check after a restart finds it. Task 9, check 5.
4. **A project disabled** disables its envs' apps one by one, and each app fires its own `app-disabled`, after its transaction commits. Task 9, check 6.
5. **A job disabled, or an app with its scheduled jobs feature off,** runs nothing by trigger, while an env's job naming that app still runs. `TestBuildRunsPicksTheJobsListening` and `TestBuildRunsLeavesOutTheAppsJobsWhenTheFeatureIsOff` (Task 3) pin it.

## Decisions beyond the spec

Taken while building the patches, and written back into the spec.

- **Status events are fired by the use cases,** through `FireAppStatusEvents` of the trigger service, after their transactions: the app service cannot hold the trigger service (trigger service → task queue → scheduled job service → env var service → app service is a cycle). `SetAppStatus` and `SetProjectEnvStatus` return the apps whose status changed.
- **Failing to fire `pre-deploy` fails the deploy:** it cannot tell whether a job would have held it.
- **A waited run past its timeout is left running;** the deploy fails. Stopping a migration half way is worse than letting it end.
- **Trigger apps travel as app IDs,** not bundle paths: import maps them with the bundle's app IDs, as it maps every reference to an app.
- **A run's cause** is in the task's args (`TaskSchedJobExecArgs.Trigger`), in the task API (`trigger {event, app {id, name}, deploymentId}`) and in the run's details on its page; there is no link to the deployment, whose route needs the app's env.
- **MCP** describes `triggers` for the app jobs it plans.
- **`post-deploy` and `deploy-failed` fire from the deploy's after-commit hook** (`onPostTx`), so the jobs see the deployment committed.
- **The trigger apps of a job are checked by the use case,** not the setting base, which would look for them in the job's own scope (`verifyingRefIDs` leaves them out).

---

## Task 1: Triggers on a scheduled job

**Files:**
- Modify: `hivepaas_app/base/sched_job.go`, `hivepaas_app/entity/setting_sched_job.go`
- Create: `hivepaas_app/entity/setting_sched_job_trigger.go`
- Test: `hivepaas_app/entity/setting_sched_job_trigger_test.go`, `hivepaas_app/service/specservice/specserviceimpl/import_sched_job_test.go`, `…/export_test.go` (the fixture's env sequence gains a trigger)

**Interfaces:**
- Produces: `base.SchedJobTriggerEvent` and its seven constants (`SchedJobTriggerPreDeploy`, `…PostDeploy`, `…DeployFailed`, `…HealthDown`, `…HealthUp`, `…AppEnabled`, `…AppDisabled`), `base.AllSchedJobTriggerEvents`, `base.SchedJobMaxTriggers = 10`; `entity.SchedJob.Triggers []*SchedJobTrigger`; `entity.SchedJobTrigger{Event, Apps []ObjectID, Wait}`; `(*SchedJob).ListensTo(event, appID string, ownApp bool) (listens, wait bool)`; `(*SchedJob).TriggerAppIDs() []string`, included in `GetRefObjectIDs().RefAppIDs`; `entity.SchedJobTriggerCause{Event, AppID, DeploymentID}`, `entity.TaskSchedJobExecArgs{Trigger}`, `(*Task).ArgsAsSchedJobExec()`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-01-tests.patch`

Tests added: `TestSchedJobListensTo`, `TestSchedJobListensToSaysWhetherToWait`, `TestSchedJobTriggerAppsAreReferences`, `TestTaskArgsAsSchedJobExec`, `TestAnEnvJobsTriggerAppsAreMappedOnImport`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/entity ./hivepaas_app/service/specservice/specserviceimpl`
Expected: FAIL, build error `unknown field Triggers in struct literal of type SchedJob`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-01-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go test ./hivepaas_app/entity ./hivepaas_app/service/specservice/specserviceimpl && go build ./... && golangci-lint run ./...`
Expected: `ok` for both; 0 issues

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(sched-jobs): triggers on a scheduled job

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 2: A job's triggers in its API

**Files:**
- Create: `hivepaas_app/usecase/settings/schedjobuc/schedjobdto/trigger.go`, `hivepaas_app/usecase/settings/schedjobuc/trigger.go`
- Modify: `…/schedjobdto/create.go`, `…/schedjobdto/get.go`, `…/schedjobuc/create.go`, `update.go`, `get.go`, `list.go`, `list_env.go`, `sequence.go` (`verifyingRefIDs`), `hivepaas_app/interface/mcp/tools_sched_write.go`, `docs/openapi/swagger.json` (generated)
- Test: `…/schedjobdto/trigger_test.go`, `…/schedjobuc/trigger_test.go`

**Interfaces:**
- Consumes: Task 1's entities.
- Produces: request `triggers: [{event, apps: [{id}], wait}]`, validated: a known event; at most 10; no two alike (same event and apps); `wait` with `pre-deploy` only; only `container-command` and `job-sequence` have triggers. `checkTriggerApps` (an app's job names no app; an env's names apps of its env); `loadTriggerApps` (nil scope); response `triggers: [{event, apps: [{id, name}], wait}]` (`TransformSchedJobTriggers`).

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-02-tests.patch`

Tests added: `TestTriggersAreKeptOnTheJob`, `TestATriggerNeedsAKnownEvent`, `TestOnlyAPreDeployTriggerWaits`, `TestAJobHasAtMostTenTriggersNoTwoAlike`, `TestASystemJobHasNoTriggers`, `TestTransformSchedJobTriggersNamesTheApps`, `TestTriggerAppsAreCheckedApartFromTheJobsOtherReferences`, `TestTriggerAppProblem`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/usecase/settings/schedjobuc/...`
Expected: FAIL, build error `undefined: SchedJobTriggerReq`

- [ ] **Step 3: Write the code and regenerate the API document**

Run: `git apply --index $P/be-02-code.patch && make gen-swag && go mod verify && git add docs/openapi/swagger.json`
Expected: `swagger.json` gains the trigger request and response; `all modules verified`

- [ ] **Step 4: Run them to watch them pass**

Run: `go test ./hivepaas_app/usecase/settings/schedjobuc/... ./hivepaas_app/interface/mcp/... && go build ./... && golangci-lint run ./...`
Expected: `ok` for every package (the MCP `TestEveryToolBuildsAndDescribesItsArguments` included); 0 issues

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(sched-jobs): a job's triggers in its API, their apps checked

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 3: The trigger service

**Files:**
- Create: `hivepaas_app/service/schedjobtriggerservice/service.go`, `types.go`, `schedjobtriggerserviceimpl/service.go`, `fire.go`; `hivepaas_app/service/schedjobexecservice/schedjobexecserviceimpl/exec_trigger.go`
- Modify: `hivepaas_app/config/tasks.go`, `…/schedjobexecserviceimpl/exec.go`, `hivepaas_app/registry/provides.go`
- Test: `…/schedjobtriggerserviceimpl/fire_test.go`, `…/schedjobexecserviceimpl/exec_trigger_test.go`, `hivepaas_app/config/config_test.go`

**Interfaces:**
- Consumes: Task 1's `ListensTo` and args.
- Produces: `schedjobtriggerservice.Service.Fire(ctx, event, app *entity.App, info *TriggerInfo) (*FireResult, error)`; `TriggerInfo{DeploymentID}`; `FireResult{Runs []*Run}` with `WaitedRuns()`; `Run{Task, JobName, Wait, Timeout}`; `buildRuns` (the jobs listening: the app's own active ones when its scheduled jobs feature is on, the env's naming it); `config.TaskTriggers{WaitPollInterval (3s), WaitTimeout (30m)}` in `config.Tasks.Triggers`; `triggerEnv(task)`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-03-tests.patch`

Tests added: `TestBuildRunsPicksTheJobsListening`, `TestBuildRunsRecordsTheCause`, `TestBuildRunsLeavesOutTheAppsJobsWhenTheFeatureIsOff` (Review Focus 5), `TestTriggerEnv`, `TestTaskTriggersDefaults`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/config ./hivepaas_app/service/schedjobexecservice/... ./hivepaas_app/service/schedjobtriggerservice/...`
Expected: FAIL, `no required module provides package …/service/schedjobtriggerservice`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-03-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go test ./hivepaas_app/config ./hivepaas_app/service/schedjobexecservice/... ./hivepaas_app/service/schedjobtriggerservice/... ./hivepaas_app/cmd/... && go build ./... && golangci-lint run ./...`
Expected: `ok`; 0 issues

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(sched-jobs): the trigger service runs the jobs an event fires

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 4: Pre-deploy jobs hold the deploy; the deploy's end fires

**Files:**
- Create: `hivepaas_app/service/schedjobtriggerservice/schedjobtriggerserviceimpl/wait.go`, `hivepaas_app/service/appdeploymentservice/appdeploymentserviceimpl/deployment_trigger.go`
- Modify: `…/schedjobtriggerservice/service.go` (`WaitForRuns`), `…/schedjobtriggerserviceimpl/service.go` (`taskService`, `cancelRun`), `…/appdeploymentserviceimpl/service.go`, `image_deploy.go`, `repo_deploy.go`, `deployment.go` (`onPostTx`)
- Test: `…/schedjobtriggerserviceimpl/wait_test.go`, `…/appdeploymentserviceimpl/deployment_trigger_test.go`

**Interfaces:**
- Consumes: Task 3's `Fire`, `FireResult.WaitedRuns()`, `config.Tasks.Triggers`.
- Produces: `WaitForRuns(ctx, runs []*Run, logStore *tasklog.Store) error` - a run failed for good or canceled, or past its timeout, fails it, logged in words; ctx ending cancels the runs still going. `deployStepPreDeployJobs` (after the pre-deployment command, both deploy methods) and `fireDeployEndedEvent` (from `onPostTx`: `done` → `post-deploy`, `failed` → `deploy-failed`, else nothing).

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-04-tests.patch`

Tests added: `TestWaitForRunsGoesOnWhenEveryRunIsDone`, `TestWaitForRunsFailsWithARunThatFailed`, `TestWaitForRunsFailsWithARunCanceled`, `TestWaitForRunsFailsPastARunsTimeout`, `TestWaitForRunsCancelsTheRunsWhenTheDeployIsCanceled`, `TestWaitedRunsLeavesOutRunsNotWaitedFor`, `TestPreDeployJobsHoldTheDeployForTheRunsThatWait`, `TestPreDeployJobsGoOnWithNoJobListening`, `TestFireDeployEndedEvent`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/service/schedjobtriggerservice/... ./hivepaas_app/service/appdeploymentservice/...`
Expected: FAIL, build error `svc.cancelRun undefined`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-04-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go test ./hivepaas_app/service/schedjobtriggerservice/... ./hivepaas_app/service/appdeploymentservice/... ./hivepaas_app/cmd/... && go build ./... && golangci-lint run ./...`
Expected: `ok`; 0 issues; the wiring test proves the deploy service takes the trigger service without a cycle

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(deploy): pre-deploy jobs hold the deploy; its end fires post-deploy or deploy-failed

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 5: Health checks and status changes fire

**Files:**
- Create: `hivepaas_app/service/healthcheckservice/healthcheckserviceimpl/healthcheck_trigger.go`, `hivepaas_app/service/schedjobtriggerservice/schedjobtriggerserviceimpl/app_status.go`
- Modify: `…/healthcheckserviceimpl/healthcheck.go`, `service.go`; `hivepaas_app/service/appservice/service.go`, `appserviceimpl/update_status.go`; `hivepaas_app/service/projectservice/service.go`, `projectserviceimpl/update_status.go`; `…/schedjobtriggerservice/service.go` (`FireAppStatusEvents`); `hivepaas_app/usecase/appuc/uc.go`, `status_update.go`; `hivepaas_app/usecase/projectuc/uc.go`, `status_update.go`; `hivepaas_app/usecase/projectenvuc/uc.go`, `status_update.go`
- Test: `…/healthcheckserviceimpl/healthcheck_trigger_test.go`, `…/schedjobtriggerserviceimpl/app_status_test.go`

**Interfaces:**
- Consumes: Task 3's `Fire`.
- Produces: `healthEvent(last, status)` - a change from a known state only; `SetAppStatus(...) (changed []*entity.App, err error)`; `SetProjectEnvStatus(...) (changedApps []*entity.App, err error)`; `schedjobtriggerservice.Service.FireAppStatusEvents(ctx, apps)`, called by the three status use cases after their transactions.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-05-tests.patch`

Tests added: `TestHealthEvent`, `TestAppStatusEvent`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/service/healthcheckservice/... ./hivepaas_app/service/schedjobtriggerservice/...`
Expected: FAIL, build error `undefined: appStatusEvent`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-05-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go test ./hivepaas_app/service/healthcheckservice/... ./hivepaas_app/service/schedjobtriggerservice/... ./hivepaas_app/cmd/... && go build ./... && go vet ./... && golangci-lint run ./...`
Expected: `ok`; 0 issues

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(sched-jobs): health checks and app status changes fire their events

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 6: A run says what fired it

**Files:**
- Modify: `hivepaas_app/entity/task.go` (`GetRefObjectIDs`), `hivepaas_app/usecase/taskuc/taskdto/get.go`, `docs/openapi/swagger.json` (generated)
- Test: `hivepaas_app/entity/setting_sched_job_trigger_test.go`, `hivepaas_app/usecase/taskuc/taskdto/trigger_test.go`

**Interfaces:**
- Consumes: Task 1's `TaskSchedJobExecArgs`.
- Produces: `TaskResp.Trigger *TaskTriggerResp{Event, App *basedto.NamedObjectResp, DeploymentID}` (`trigger` in JSON) for a `sched-job-exec` task a trigger fired; the task's references include the trigger's app.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-06-tests.patch`

Tests added: `TestATriggeredRunReferencesItsApp`, `TestTransformTaskSaysWhatTriggeredARun`, `TestTransformTaskGivesNoTriggerForARunByHand`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/entity ./hivepaas_app/usecase/taskuc/...`
Expected: FAIL, build error `undefined: TaskTriggerResp`

- [ ] **Step 3: Write the code and regenerate the API document**

Run: `git apply --index $P/be-06-code.patch && make gen-swag && go mod verify && git add docs/openapi/swagger.json`

- [ ] **Step 4: Run the whole backend**

Run: `go build ./... && go test ./... && golangci-lint run ./...`
Expected: every package `ok`; 0 issues

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(tasks): a run says what fired it

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

Tasks 7 and 8 run in the dashboard repo, on a branch from its `main`.

## Task 7: A job's triggers, in its form and its list

**Files (dashboard, `src/application/modules/projects/`):**
- Create: `module-shared/enums/e.sched-job-trigger.ts` (`ESchedJobTriggerEvent`, `SCHED_JOB_TRIGGER_EVENT_OPTIONS`), `module-shared/components/job-triggers-field/` (`JobTriggersField`, helpers)
- Modify: the job entity and validator (`triggers`; an event the dashboard does not know is left out, not a failed list), the upsert payload (`triggers?`), the job form's schema, mappers and form (a Triggers block under Scheduling), the app job form route's payload, `JobSequenceForm` (schema, mappers, `triggerApps` prop, Triggers block), the env sequence form route (passes the env's apps), `ScheduledJobNameCell` (a tag per trigger)

**Interfaces:**
- Consumes: Task 2's wire format.
- Produces: `JobTriggersField({apps?, readOnly})` over the form value `triggers: {event, appIds, wait}[]` - `apps` given for an env's job, left out for an app's; `mapJobTriggersToFormValues`, `mapJobTriggersFormValuesToPayload`, `formatJobTrigger` (`post-deploy`, `pre-deploy · waits`).

- [ ] **Step 1: Apply the patch**

Run: `git -C $D apply --index "$PWD/$P/dash-01.patch"`

- [ ] **Step 2: Check it**

Run: `cd $D && npx tsc --noEmit && npm run lint && npx prettier --check src`
Expected: no output from `tsc`; the lint prints no problem; "All matched files use Prettier code style!"

- [ ] **Step 3: Commit**

```bash
git -C $D commit -m "feat(sched-jobs): a job's triggers, in its form and its list

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 8: A run says which trigger fired it

**Files (dashboard, `src/application/modules/operations/`):**
- Modify: `domain/system-task.entity.ts` (`SystemTaskTrigger`, `trigger`), `api/services/system-tasks-services/system-tasks.api.validator.ts`, `routes/tasks/building-blocks/task-summary-card/system-task-summary-card.com.tsx` ("Triggered By" and "Deployment" in the run's details)

**Interfaces:**
- Consumes: Task 6's `trigger`.

- [ ] **Step 1: Apply the patch**

Run: `git -C $D apply --index "$PWD/$P/dash-02.patch"`

- [ ] **Step 2: Check it**

Run: `cd $D && npx tsc --noEmit && npm run lint && npx prettier --check src`
Expected: as in Task 7

- [ ] **Step 3: Commit**

```bash
git -C $D commit -m "feat(tasks): a run says which trigger fired it

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 9: Live checks on the Linux server, then merge

Build the images from both branches and run them on the user's Linux server. Two apps in one env, `a` and `b`, each with a health check; a job of `a` running `sleep 60; echo migrated`.

- [ ] **Check 1: post-deploy.** Give the job a `post-deploy` trigger; deploy `a`: the deploy's log names the job started; its run page says "Triggered By: post-deploy of a" and the deployment; its log has `HIVEPAAS_TRIGGER_*` (`env | grep HIVEPAAS_TRIGGER`).
- [ ] **Check 2 (Review Focus 2): pre-deploy with wait.** Make it `pre-deploy`, "Deploy waits for this job"; deploy `a`: the log waits about a minute, says the job is done, then applies the new version. Make the command `exit 1`: the deploy fails, its log saying the job failed, and the old version keeps running.
- [ ] **Check 3 (Review Focus 1): a second deploy while one waits.** Start a deploy of `a` and, while it waits, another: the second runs after the first ends. Cancel a deploy while it waits: the job's run is canceled.
- [ ] **Check 4: an env job.** An env sequence with `deploy-failed` of `b`; make `b`'s deploy fail: the sequence runs; `a`'s deploys do not run it.
- [ ] **Check 5 (Review Focus 3): health.** A job on `health-down` and one on `health-up` of `a`; make `a`'s health check fail, then pass: each runs once. Restart HivePaaS: neither runs.
- [ ] **Check 6 (Review Focus 4): status.** Jobs of an env on `app-disabled` / `app-enabled` of `a` and `b`; disable then enable the project: each fires once per app.
- [ ] **Check 7: the dashboard.** Add, edit and remove triggers on an app job and an env sequence; the list tags (`pre-deploy · waits`); a job with triggers and no schedule shows "No schedule".
- [ ] **Merge** both branches into their `main` locally once the checks pass. Never push.

---

## Self-review

- **Spec coverage:** §1 → Tasks 1, 2; §2 → Tasks 3, 4, 5; §3 → Tasks 2, 6, 7, 8 (export and import in Task 1's test); §4 → the tests of Tasks 1-6 and Task 9.
- **Placeholders:** none; every code step is a verified patch.
- **Type consistency:** built in order on one tree each; the patches applied to fresh worktrees reproduce both trees, and `go test ./...`, the lint and swagger matched.
- **Review Focus:** five lines; 5 is pinned by two tests, 1-4 by Task 9's live checks, which need a swarm and a real deploy.
