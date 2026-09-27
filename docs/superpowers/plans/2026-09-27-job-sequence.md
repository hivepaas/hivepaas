# Job Sequence Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A `job-sequence` scheduled job type that runs other scheduled jobs in order, at the app and the project env scope, with a dashboard to build, run and watch it.

**Architecture:** A sequence is a `sched-job` setting whose `Sequence` lists member jobs. Its run is one `task:sched-job-exec` task that runs one step per execution and asks the queue to run it again (`TaskExecData.Continue()`), keeping the run in the task's output. A `container-command` step is told the earlier steps' results through `HIVEPAAS_SEQ_*` variables and hands values on through an output file read back by a second exec in the same container. An env lists its own jobs and its apps' through a new env API; spec import writes an env's sequences with their members mapped.

**Tech Stack:** Go (bun, fx, echo), the task queue in `hivepaas_app/tasks/queue`; React 19, TanStack Query, react-hook-form and zod in `hivepaas-dashboard`.

**Spec:** `docs/superpowers/specs/2026-09-27-job-sequence-design.md`

## How the patches work

Every task's code was written and verified before this plan: the backend patches apply in order to `main` at `fd3b5a6e`, the dashboard patches to the dashboard's `main` at `d1e6e70f`. Applied in order they give a tree that builds, passes `go test ./...` and `golangci-lint run ./...` (0 issues), and, for the dashboard, `tsc`, `npm run lint` and `prettier --check`.

Patches live in `docs/superpowers/plans/2026-09-27-job-sequence/`:

- `be-NN-tests.patch`, then `be-NN-code.patch` for backend task NN: the tests first, watched failing, then the code.
- `dash-NN.patch` for dashboard task NN. The dashboard has no test runner; its gate is the type check, the lint and, in Task 12, the browser.

`P=docs/superpowers/plans/2026-09-27-job-sequence` in the commands below, run from the backend repo root. The dashboard repo is `../hivepaas-dashboard`. If a base has moved since, use `git apply --3way` and resolve.

## Global Constraints

- Backend, before a task is done: `go build ./...`, `golangci-lint run ./...` over the **whole** repo (120-character lines, US spelling), `go test ./...`.
- `make gen-swag` when a DTO changes; `docs/openapi/swagger.json` is generated and committed. Its first run can fail with `cannot find all dependencies` while the devtools container downloads modules: run it again.
- Dashboard, before a task is done: `npx tsc --noEmit`, `npm run lint` (`--max-warnings 0`), `npx prettier --check src`.
- A wire change and its dashboard change belong to the same piece of work.
- Variables a step sees are prefixed `HIVEPAAS_`, never `HP_` (the app's own configuration).
- A sequence has 1 to 50 steps; an output file is read up to 64 KB, keys `[A-Za-z_][A-Za-z0-9_]*`, at `/tmp/hivepaas-output-<task>-<step>`.
- `job-sequence` exists at the app and the project env scope only; the env scope allows no other type for now.
- No backward compatibility is needed.
- Commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`. Merge locally, never push. Never `git stash` (the stash is shared with the user's repo); never touch the user's uncommitted work.
- Nothing is deployed to the user's Docker Desktop swarm. The live checks of Task 12 run on the user's Linux server.

## Review Focus

1. **Two runs of one sequence overlap** (its schedule is shorter than a run). The second must end `done` at once, its log saying "Skipped: the previous run of this sequence is still going", and the first must go on. `otherRunGoing` is a query no unit test reaches: Task 12, check 3.
2. **The worker restarts in the middle of a step.** The step that was running must run again, not be skipped and not be followed by the next one. Task 12, check 4.
3. **A step's app runs two replicas.** Its outputs must be read from the container the step ran in, not from the other replica. Task 12, check 5.
4. **A list holds a job without a schedule, or a sequence.** The dashboard's parsing of the job list must not fail: `schedule` was a required object and `jobType` a closed enum before Task 8. Task 12, check 1 opens both lists with such jobs in them.
5. **A step's job is deleted, disabled or never imported.** The step is skipped with its reason and the run goes on as `onFailure` says. `TestMemberSkipReason` (Task 4) and `TestAnEnvSequenceWithAStepNotFoundIsWrittenPending` (Task 7) pin it.

## Decisions beyond the spec

These were decided while building the patches and are written back into the spec.

- **Disabling a member is not blocked**, only deleting one: a disabled member's step is skipped. The setting-in-use check covers deletes.
- **The run lives in the task's output**, written after every step, not in its args: the task API returns it while the run goes on (`sequenceRun` on a task).
- **The first execution only prepares the run** (overlap check, the first step's retry and timeout, which the queue reads before an execution starts), then `Continue()`.
- **Exit codes** come from the container exec, which now reports `ExitCode` and returns its response with the error of a non-zero exit, so a failed step keeps its code and outputs.
- **The output file is read by a second exec pinned to the step's container and node** (`ContainerExecReq.ContainerID`, `NodeID`).
- **Notifications:** the sequence's own, at the end of the run; a failed run's carries the step summary as its error detail. The log ends with the same summary.
- **MCP does not make sequences.** `plan_create_sched_job` describes `sequence` as not used there.
- **Env members are loaded without a scope** (`LoadRefObjectsByIDsSkipMissing` with a nil scope): an env's scope does not see its apps' settings.
- **Import** allows `sched-job` in the project env scope only (`importPolicyIn`), schedules the written env jobs in the transaction (`scheduleEnvJobs`), and does not re-check a sequence's reach: a member it cannot find is reported, the sequence written `pending`, its step empty.
- **Dashboard:** the sequence form has no status field (a sequence is created active; the lists enable and disable), no retry of its own, a step timeout for steps whose job has none, and "Allow canceling" on by default (a run is canceled through task control). A new sequence defaults to No schedule. The env page shows "Next Run" (the API has no last run), gives the env's own rows run now, enable or disable and remove, and an app's rows "Open in <app>". Env pages work in the env picked in the header, like the other env settings pages. The project's Tasks page takes `?targetId=` so "View Runs" opens a job's runs. `PriorityTabsField` moves to `module-shared/components/job-schedule-fields` so both forms use it.

---

## Task 1: The job-sequence type, its steps, and optional schedules

**Files:**
- Modify: `hivepaas_app/base/sched_job.go`, `hivepaas_app/entity/setting_sched_job.go`, `hivepaas_app/tasks/taskschedjobexec/executor.go`, `hivepaas_app/tasks/taskschedjobexec/result_notification.go`
- Create: `hivepaas_app/entity/setting_sched_job_sequence.go`, `hivepaas_app/entity/task_sched_job_seq.go`
- Test: `hivepaas_app/entity/setting_sched_job_sequence_test.go`, `hivepaas_app/tasks/taskschedjobexec/result_notification_test.go`

**Interfaces:**
- Produces: `base.SchedJobTypeJobSequence`, `base.SchedJobSeqMode` (`SchedJobSeqModeSequential`), `base.SchedJobSeqOnFailure` (`Stop`, `Continue`), `base.SchedJobSeqStepStatus` (`Pending`, `Running`, `Done`, `Failed`, `Skipped`), `base.SchedJobSeqMaxSteps = 50`; `entity.SchedJob.Sequence *SchedJobSequence`; `entity.SchedJobSequence{Mode, OnFailure, Steps}` with `MemberIDs() []string` and `StopsOnFailure() bool`; `entity.SchedJobSequenceStep{Job ObjectID, Name string}`; `entity.SchedJobSeqRun{Started, CurrentStep, Steps}`, `entity.SchedJobSeqStepResult{Job, Name, Status, ExitCode *int, Error, Attempts, StartedAt, EndedAt, Outputs}`, `entity.NewSchedJobSeqRun(*SchedJobSequence)`, `(*SchedJobSeqRun).Failed()`, `(*entity.Task).OutputAsSchedJobSeqRun() (*SchedJobSeqRun, error)`. A nil `Schedule` has no runs. `describeSchedule(sched)` in the notification.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-01-tests.patch`

Tests added: `TestSchedJobSequenceMembersAreReferences`, `TestSchedJobSequenceStopsOnFailureUnlessToldToContinue`, `TestAJobWithoutAScheduleHasNoRuns`, `TestNewSchedJobSeqRunStartsEveryStepPending`, `TestTaskOutputAsSchedJobSeqRun`, `TestDescribeSchedule`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/entity ./hivepaas_app/tasks/taskschedjobexec`
Expected: FAIL, build error `undefined: SchedJobSequence`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-01-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go test ./hivepaas_app/entity ./hivepaas_app/tasks/taskschedjobexec && go build ./... && golangci-lint run ./...`
Expected: `ok` for both packages; 0 issues

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(sched-jobs): the job-sequence type, its steps, and optional schedules

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 2: The queue's Continue()

**Files:**
- Modify: `hivepaas_app/tasks/queue/types.go`, `hivepaas_app/tasks/queue/queueimpl/queue_execution.go`, `hivepaas_app/tasks/taskworkflow/executor.go`
- Test: `hivepaas_app/tasks/queue/types_test.go`, `hivepaas_app/tasks/queue/queueimpl/queue_settle_test.go`, `hivepaas_app/tasks/taskworkflow/executor_steps_test.go`

**Interfaces:**
- Produces: `(*queue.TaskExecData).Continue()` (a sub-task's reaches its owner) and `Continued() bool`; `settleTask(task, taskData, execErr, timeNow) (rescheduleAt time.Time)` in `queueimpl`, in this order: an error fails or retries, a cancel cancels, a continued task goes back to `not-started` with `RunAt = now` and is rescheduled at once, anything else is done. `task:workflow` uses `Continue()` in place of scheduling itself after commit.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-02-tests.patch`

Tests added: `TestSettleTaskMarksASuccessDone`, `TestSettleTaskRunsAContinuedTaskAgainAtOnce`, `TestSettleTaskLetsACancelWinOverContinue`, `TestSettleTaskRetriesAFailure`, `TestSettleTaskGivesUpOnANonRetryableFailure`, `TestContinueOnASubTaskReachesItsOwner`, `TestWorkflowRunsOneStepPerExecutionAndContinues`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/tasks/queue/... ./hivepaas_app/tasks/taskworkflow`
Expected: FAIL, build error `first.Continued undefined (type *queue.TaskExecData has no field or method Continued)`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-02-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go test ./hivepaas_app/tasks/queue/... ./hivepaas_app/tasks/taskworkflow && go build ./... && golangci-lint run ./...`
Expected: `ok` for the three packages; 0 issues

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(queue): Continue() runs a task again for its next step

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 3: Creating and updating a sequence

**Files:**
- Create: `hivepaas_app/usecase/settings/schedjobuc/schedjobdto/sequence.go`, `hivepaas_app/usecase/settings/schedjobuc/sequence.go`
- Modify: `hivepaas_app/usecase/settings/schedjobuc/schedjobdto/create.go`, `…/schedjobdto/get.go`, `hivepaas_app/usecase/settings/schedjobuc/create.go`, `…/update.go`, `…/get.go`, `…/list.go`, `hivepaas_app/interface/mcp/tools_sched_write.go`
- Test: `hivepaas_app/usecase/settings/schedjobuc/schedjobdto/sequence_test.go`, `hivepaas_app/usecase/settings/schedjobuc/sequence_test.go`

**Interfaces:**
- Consumes: Task 1's entities.
- Produces: request `sequence {mode, onFailure, steps[{job{id}, name}]}` (defaults `sequential`, `stop`; 1..50 steps; a name up to 100); a sequence has no `command`, `app.id` or `commandOutput`; any job's `schedule` may be null; a `container-command` needs a command. Response `sequence {mode, onFailure, steps[{job (base setting, "missing" when gone), app {id, name}, name}]}`. `checkJobTypeInScope` (a sequence at app and env only; env allows only sequences), `checkSequenceMembers` (members exist, are in reach, are not sequences), `loadSequenceMembers` (nil scope).

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-03-tests.patch`

Tests added: `TestASequenceIsValidWithoutASchedule`, `TestASequenceNeedsBetweenOneAndFiftySteps`, `TestASequenceStepNamesAJob`, `TestASequenceRefusesAnUnknownModeOrFailurePolicy`, `TestASequenceRunsNoCommandOfItsOwn`, `TestOnlyASequenceHasSteps`, `TestAnyJobMayGoWithoutASchedule`, `TestTransformSchedJobGivesASequencesSteps`, `TestJobSequencesLiveInAppsAndEnvs`, `TestASequencesMembersAreCheckedApartFromItsOtherReferences`, `TestMemberProblem`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/usecase/settings/schedjobuc/...`
Expected: FAIL, build error `unknown field Sequence in struct literal of type SchedJobBaseReq`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-03-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go test ./hivepaas_app/usecase/settings/schedjobuc/... ./hivepaas_app/interface/mcp/... && go build ./... && golangci-lint run ./...`
Expected: `ok` for every package (the MCP test `TestEveryToolBuildsAndDescribesItsArguments` included); 0 issues

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(sched-jobs): create and update job sequences, their members checked

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 4: The runner

**Files:**
- Create: `hivepaas_app/tasks/taskschedjobexec/job_run.go`, `…/sequence.go`, `…/sequence_steps.go`
- Modify: `hivepaas_app/tasks/taskschedjobexec/executor.go` (`NewExecutor` gains `taskRepo repository.TaskRepo` after `taskLogRepo`), `hivepaas_app/usecase/taskuc/taskdto/get.go`
- Test: `hivepaas_app/tasks/taskschedjobexec/sequence_steps_test.go`, `hivepaas_app/usecase/taskuc/taskdto/sequence_run_test.go`

**Interfaces:**
- Consumes: Task 1's run entities, Task 2's `Continue()`.
- Produces: `runJob(ctx, db, *jobRun) (*jobResult, error)` (the moved job-type switch; `jobResult{skipNotification, exitCode, outputs}`), `executeSequence`, `stepConfig(runConfig, member, seq)`, `retriesLeft`, `afterStep(run, stopOnFailure) seqNext` (`seqContinue`, `seqDone`, `seqFailed`), `markCanceled`, `memberSkipReason`, `sequenceSummary`; `TaskResp.SequenceRun` (`sequenceRun` in JSON) on a `sched-job-exec` task.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-04-tests.patch`

Tests added: `TestStepConfigTakesTheMembersRetryAndTimeout`, `TestRetriesLeft`, `TestAfterStepGoesOnWhileStepsAreLeft`, `TestAfterStepEndsDoneWhenNoStepFailed`, `TestAfterStepStopsAtAFailureAndSkipsTheRest`, `TestAfterStepCountsASkipAsAFailureWhenStopping`, `TestAfterStepContinuesPastAFailureAndEndsFailed`, `TestAfterStepContinuesPastASkipAndEndsDone`, `TestMarkCanceledEndsTheRun`, `TestSequenceSummary`, `TestMemberSkipReason` (Review Focus 5), `TestTransformTaskGivesASequencesRun`, `TestTransformTaskGivesNoRunForAPlainJob`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/tasks/taskschedjobexec ./hivepaas_app/usecase/taskuc/taskdto`
Expected: FAIL, build error `resp.SequenceRun undefined (type *TaskResp has no field or method SequenceRun)`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-04-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go test ./hivepaas_app/tasks/taskschedjobexec ./hivepaas_app/usecase/taskuc/taskdto && go build ./... && golangci-lint run ./...`
Expected: `ok` for both packages; 0 issues. `go build` proves the fx wiring takes the new `NewExecutor` parameter.

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(sched-jobs): the job sequence runner, one step per execution

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 5: A step's environment, outputs and exit code

**Files:**
- Create: `hivepaas_app/service/schedjobexecservice/schedjobexecserviceimpl/exec_sequence.go`
- Modify: `hivepaas_app/service/schedjobexecservice/types.go`, `…/schedjobexecserviceimpl/exec.go`, `hivepaas_app/service/containerexecservice/types.go`, `…/containerexecserviceimpl/exec.go`, `hivepaas_app/tasks/taskschedjobexec/job_run.go`, `…/sequence.go`
- Test: `hivepaas_app/service/schedjobexecservice/schedjobexecserviceimpl/exec_sequence_test.go`

**Interfaces:**
- Consumes: Task 4's `jobRun.sequence`.
- Produces: `SchedJobExecReq.Sequence *SequenceStep{Step, Steps, Earlier, OutputFile}`; `SchedJobExecResp.ExitCode *int`, `Outputs map[string]string`; `schedjobexecservice.OutputFilePath(taskID, step)`; `sequenceEnv`, `parseOutputs` (64 KB, key rule, warnings), `readOutputs` (a second exec pinned to the step's container); `ContainerExecReq.ContainerID`, `NodeID`; `ContainerExecResp.ExitCode`, returned with the error on a non-zero exit.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-05-tests.patch`

Tests added: `TestSequenceEnvTellsAStepHowTheOnesBeforeItWent`, `TestSequenceEnvOfTheFirstStep`, `TestParseOutputs`, `TestParseOutputsKeepsTheFirst64KB`, `TestOutputFilePath`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/service/schedjobexecservice/...`
Expected: FAIL, build error `undefined: schedjobexecservice.SequenceStep`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-05-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go test ./hivepaas_app/service/schedjobexecservice/... ./hivepaas_app/service/containerexecservice/... ./hivepaas_app/tasks/taskschedjobexec && go build ./... && golangci-lint run ./...`
Expected: `ok`; 0 issues

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(sched-jobs): a sequence step's environment, outputs and exit code

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 6: An env's scheduled jobs API

**Files:**
- Create: `hivepaas_app/usecase/settings/schedjobuc/list_env.go`, `…/schedjobdto/list_env.go`, `hivepaas_app/interface/api/handler/projectenvsettingshandler/sched_job.go`
- Modify: `hivepaas_app/entity/object_scope.go`, `hivepaas_app/repository/setting_repo.go`, `hivepaas_app/interface/api/server/router_project_env.go`, `docs/openapi/swagger.json` (generated)
- Test: `hivepaas_app/repository/setting_repo_scope_test.go`, `hivepaas_app/usecase/settings/schedjobuc/schedjobdto/list_env_test.go`

**Interfaces:**
- Consumes: Task 3's DTOs and `loadSequenceMembers`.
- Produces: `ObjectScope.IncludeEnvApps`; `GET/POST /projects/{projectID}/{projectEnv}/sched-jobs`, `GET/PUT/DELETE …/:itemID`, `PUT …/:itemID/status`, `POST …/:itemID/exec`. The list takes `jobType` (comma list) and `appId`; each item adds `scope` (`project-env` or `app`) and `ownerApp {id, name}`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-06-tests.patch`

Tests added: `TestEnvScopeWithItsApps`, `TestTransformEnvSchedJobsSaysWhoseEachJobIs`, `TestListEnvSchedJobReqFiltersByJobTypeAndApp`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/repository ./hivepaas_app/usecase/settings/schedjobuc/schedjobdto`
Expected: FAIL, build error `undefined: TransformEnvSchedJobs`

- [ ] **Step 3: Write the code and regenerate the API document**

Run: `git apply --index $P/be-06-code.patch && make gen-swag && git add docs/openapi/swagger.json`
Expected: `swagger.json` gains the env `sched-jobs` paths. If `make gen-swag` fails with `cannot find all dependencies`, run it again.

- [ ] **Step 4: Run them to watch them pass**

Run: `go test ./hivepaas_app/repository ./hivepaas_app/usecase/settings/schedjobuc/... && go build ./... && golangci-lint run ./...`
Expected: `ok`; 0 issues

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(sched-jobs): an env's scheduled jobs API, with its apps' jobs

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 7: Spec import of an env's sequences

**Files:**
- Modify: `hivepaas_app/service/specservice/specserviceimpl/import_policy.go`, `…/import_diff.go`, `…/import_write.go`, `…/import_apply_apps.go`, `…/export_test.go` (the fixture gains an app job and an env sequence)
- Test: `hivepaas_app/service/specservice/specserviceimpl/import_sched_job_test.go`

**Interfaces:**
- Consumes: Task 1's `Sequence` (its `MemberIDs()` is how export and import find the references).
- Produces: `importPolicyIn(typ, scope)` over `scopedPolicies` (`sched-job` imported in a project env only), `settingsScope(node)`, `(*writer).scheduleEnvJobs(ctx)` after `PersistProjectData`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-07-tests.patch`

Tests added: `TestAnEnvSequenceIsImportedWithItsStepsMapped`, `TestAnEnvSequenceWithAStepNotFoundIsWrittenPending` (Review Focus 5), `TestAProjectSchedJobIsSkipped`, `TestImportPolicyInAScope`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/service/specservice/specserviceimpl`
Expected: FAIL, build error `undefined: importPolicyIn`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-07-code.patch`

- [ ] **Step 4: Run the whole backend**

Run: `go build ./... && go test ./... && golangci-lint run ./...`
Expected: every package `ok`; 0 issues

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(spec-import): an env's job sequences, their steps mapped and scheduled

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

Tasks 8 to 11 run in the dashboard repo, `../hivepaas-dashboard`, on a branch from its `main`. `D=../hivepaas-dashboard`.

## Task 8: Types, the env jobs API, and No schedule

**Files (dashboard, `src/application/modules/projects/`):**
- Modify: `module-shared/enums/e.app-scheduled-job-type.ts` (`JobSequence`), `e.app-scheduled-job-schedule-mode.ts` (`None`), `domain/apps/scheduled-job/app-scheduled-job.entity.ts` (`schedule: … | null`, `sequence`, `EnvScheduledJob`), `api/services/project-apps-services/scheduled-jobs/*` (schema: `schedule` and `nextRuns` nullable, `sequence`; payload: `schedule | null`, optional `app` and `command`, `sequence`; `toUpsertPayload` exported), the api context, hooks, query keys, `dialogs/create-or-edit-app-scheduled-job/form/*` (the schedule block replaced by `JobScheduleFields`), the app job form route (`mapJobScheduleFormValuesToPayload`), the jobs table (`formatJobSchedule`)
- Create: `module-shared/enums/e.sched-job-sequence.ts`, `module-shared/components/job-schedule-fields/` (`JobScheduleFields`, `job-schedule.helpers.ts`, `PriorityTabsField` moved here), `api/services/projects-services/project-env-scheduled-jobs/`, `api/hooks/project-env-scheduled-jobs/`, `data/queries/project-env-scheduled-jobs/`, `data/commands/project-env-scheduled-jobs/`

**Interfaces:**
- Consumes: Tasks 3 and 6's wire format.
- Produces: `EnvScheduledJobsApi` (`findManyPaginated` with `jobTypes`, `appId`; `findOneById`, `createOne`, `updateOne`, `updateStatus`, `deleteOne`, `runNow`), `EnvScheduledJobsQueries`, `EnvScheduledJobsCommands`; `JobScheduleFields({titleWidth, nextRuns, readOnly})` over the form fields `scheduleMode`, `scheduleInterval`, `scheduleCronExpr`, `scheduleFrom`, `scheduleTo`; `createDefaultJobScheduleFormValues`, `mapJobScheduleToFormValues`, `mapJobScheduleFormValuesToPayload` (null for No schedule), `formatJobSchedule` ("No schedule").

- [ ] **Step 1: Apply the patch**

Run: `git -C $D apply --index "$PWD/$P/dash-01.patch"`

- [ ] **Step 2: Check it**

Run: `cd $D && npx tsc --noEmit && npm run lint && npx prettier --check src`
Expected: no output from `tsc`; the lint prints no problem; "All matched files use Prettier code style!"

- [ ] **Step 3: Commit**

```bash
git -C $D commit -m "feat(sched-jobs): job sequence types, the env jobs API, and No schedule

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 9: JobSequenceForm, and sequences on an app's page

**Files (dashboard, `src/application/modules/projects/`):**
- Create: `module-shared/components/job-sequence-form/` (`JobSequenceForm`, `JobSequenceStepsField`, schema, mappers), `module-shared/definitions/tables/app-scheduled-jobs/building-blocks/name-cell.com.tsx`, `routes/single-project/single-app/configuration/scheduled-jobs/sequence/` (`AppJobSequenceFormRoute`, `AppJobSequenceCreateRoute`)
- Modify: the app jobs list (`+ New Job Sequence`), its table (the name cell's "Sequence · N steps" tag), the edit route (opens the form the job's type needs), `projects.router.tsx`, `projects.module.ts`, `routes/index.ts`, `shared/constants/route.constants.ts` (`…/sched-jobs/create-sequence`)

**Interfaces:**
- Consumes: Task 8's types, `JobScheduleFields`, `PriorityTabsField`.
- Produces: `JobSequenceForm({projectId, env, candidates, isLoadingCandidates, isPending, onSubmit, initialValues, onHasChanges, readOnly, stickyActions, onClose})`, `JobSequenceCandidate {id, name, appName, disabled}`, `mapJobSequenceFormToPayload(values)`, `ScheduledJobNameCell({job})`.

- [ ] **Step 1: Apply the patch**

Run: `git -C $D apply --index "$PWD/$P/dash-02.patch"`

- [ ] **Step 2: Check it**

Run: `cd $D && npx tsc --noEmit && npm run lint && npx prettier --check src`
Expected: as in Task 8

- [ ] **Step 3: Commit**

```bash
git -C $D commit -m "feat(sched-jobs): JobSequenceForm, and job sequences on an app's page

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 10: A project's Scheduled jobs page

**Files (dashboard):**
- Create: `src/application/modules/projects/routes/single-project/configuration/scheduled-jobs/` (the list route with its env note, `EnvScheduledJobsTable` with type and owner filters, table defs, `EnvScheduledJobMenuCell`, `EnvJobSequenceFormRoute`, create and edit routes)
- Modify: the project sidebar ("Scheduled Jobs" under Automation), `projects.router.tsx`, `projects.module.ts`, `routes/index.ts`, `routes/single-project/configuration/index.ts`, `shared/constants/route.constants.ts` (`projects/:id/integrations/sched-jobs`, `create-sequence`, `:scheduledJobId/edit`), the project Tasks route (`?targetId=`), `shared/utils/setting-usage-link.ts` (an env `sched-job` links to its edit page)

**Interfaces:**
- Consumes: Tasks 8 and 9.
- Produces: `ProjectScheduledJobsRoute`, `EnvJobSequenceCreateRoute`, `EnvJobSequenceEditRoute`.

- [ ] **Step 1: Apply the patch**

Run: `git -C $D apply --index "$PWD/$P/dash-03.patch"`

- [ ] **Step 2: Check it**

Run: `cd $D && npx tsc --noEmit && npm run lint && npx prettier --check src`
Expected: as in Task 8

- [ ] **Step 3: Commit**

```bash
git -C $D commit -m "feat(sched-jobs): a project's Scheduled jobs page, per env

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 11: A run's steps on its task page

**Files (dashboard, `src/application/modules/operations/`):**
- Create: `routes/tasks/building-blocks/sequence-run-steps/` (`SequenceRunSteps`: number, step, status, attempts, duration, exit code, error, output keys with their values in a popover)
- Modify: `domain/system-task.entity.ts` (`SystemTaskSequenceRun`, `sequenceRun`), `api/services/system-tasks-services/system-tasks.api.validator.ts`, `routes/tasks/details/route/system-task-details.route.com.tsx` (the table above the log; the task is polled while a sequence's run goes on)

**Interfaces:**
- Consumes: Task 4's `sequenceRun`.

- [ ] **Step 1: Apply the patch**

Run: `git -C $D apply --index "$PWD/$P/dash-04.patch"`

- [ ] **Step 2: Check it**

Run: `cd $D && npx tsc --noEmit && npm run lint && npx prettier --check src`
Expected: as in Task 8

- [ ] **Step 3: Commit**

```bash
git -C $D commit -m "feat(tasks): a job sequence's run shows its steps

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 12: Live checks on the Linux server, then merge

Build the images from both branches and run them on the user's Linux server (never on the Docker Desktop swarm). Two apps in one env, `a` and `b`, each with a `container-command` job; `b`'s app has two replicas.

- [ ] **Check 1 (Review Focus 4): the lists.** Give `a`'s job No schedule. Open `a`'s Scheduled Jobs: the job shows "No schedule"; create an app sequence there with `+ New Job Sequence`; the list shows it with "Sequence · 1 step". Open Project → Scheduled Jobs with "All" picked: the note shows; pick the env: the env's and both apps' jobs show, filters by type and owner work.
- [ ] **Check 2: an env sequence.** Create one of `a`'s job (`echo "VERSION=1.2.3" >> "$HIVEPAAS_OUTPUT"`) then `b`'s (`echo "$HIVEPAAS_SEQ_OUTPUT_VERSION $HIVEPAAS_SEQ_PREV_STATUS"`). Run now: the run's page shows both steps `done`, output key `VERSION` with `1.2.3`, and `b`'s log prints `1.2.3 done`. Make `a`'s command `exit 3`: with Stop, step 1 `failed` (exit code 3), step 2 `skipped`, the run failed and its notification lists the steps; with Continue, step 2 runs.
- [ ] **Check 3 (Review Focus 1): overlap.** Give `a`'s job `sleep 90`, the sequence a 1-minute interval: the second run ends `done` at once, its log saying it was skipped, while the first goes on.
- [ ] **Check 4 (Review Focus 2): restart.** Restart the HivePaaS service while `a`'s `sleep 90` step runs: after the restart the same step runs again, then step 2.
- [ ] **Check 5 (Review Focus 3): replicas.** Make `b`'s job the one writing `VERSION`, run the sequence several times: the output is read every time.
- [ ] **Check 6: in use.** Delete `a`'s job: the setting-in-use dialog names the sequence and links to it. Disable it instead: the next run skips the step with "the scheduled job is disabled".
- [ ] **Check 7: import.** Export the project, import it into a new project: the env sequence arrives with its steps naming the new apps' jobs, and runs.
- [ ] **Merge** both branches into their `main` locally once the checks pass. Never push.

---

## Self-review

- **Spec coverage:** §1 → Tasks 1, 3; §2 → Tasks 2, 4, 5; §3 → Tasks 3, 6, 7; §4 → Tasks 8-11; §5 → the tests of Tasks 1-7 and Task 12.
- **Placeholders:** none; every code step is a verified patch.
- **Type consistency:** the patches were built in order on one tree each; applied in order they reproduce it (checked by applying them to fresh worktrees and comparing).
- **Review Focus:** five lines; 5 is pinned by two tests, 1-4 by Task 12's live checks, which need a swarm this machine does not offer to tests.
