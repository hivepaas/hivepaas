# Job Sequence: Scheduled Jobs That Run Other Jobs in Order

A user who needs several jobs run one after another - migrate, then reindex,
then purge a cache, across the apps of an env - has no way to say so: each
scheduled job runs on its own schedule, alone. This design adds a scheduled job
type, **job sequence**, that runs a list of other scheduled jobs in order,
stops or goes on when one fails, and hands each step what the steps before it
did.

HivePaaS itself has no use for it; it is for its users.

---

## Decisions

1. **A job type, not a new setting.** `SchedJobType = "job-sequence"` on the
   existing `sched-job` setting: the schedule, run-now, priority, timeout,
   notifications, task history and logs of scheduled jobs are reused.
2. **Members are existing scheduled jobs, referenced.** A member keeps its own
   schedule, retry and timeout; the sequence only adds an order.
3. **Scopes: app and project env.** An env sequence orders jobs of the apps in
   that env; an app sequence, jobs of that app. Not project (it spans envs, and
   a chain across them reaches prod by accident), not global (not needed yet).
4. **Sequential now, parallel later.** `mode` exists with one value,
   `sequential`; the run's per-step results are shaped so a parallel mode can
   be added without changing them.
5. **One task, one step per run, progress kept in the task.** A run is one
   `task:sched-job-exec` task that runs one step each time it executes, in that
   execution's own transaction, saves its progress in its args and asks the
   queue to run it again for the next step. A worker restart resumes at the
   step it was on. The existing `task:workflow` has the same idea but is unused
   and does not advance past its first step (the queue marks a task done when
   it returns, and `ScheduleTask` skips done tasks); the queue change that
   fixes this is generic.
6. **Steps talk through the environment.** A `container-command` step is told
   how the earlier steps went and what they output: GitHub Actions' model, an
   output file of `KEY=value` lines read after the step.
7. **A separate form.** `+ New Job Sequence` next to `+ New Job`, with a form of
   its own shared by the app and the env screens; the lists stay one.

## 1. Data model

```go
// base
SchedJobTypeJobSequence SchedJobType = "job-sequence"
SchedJobSeqModeSequential SchedJobSeqMode = "sequential"
SchedJobSeqOnFailureStop     SchedJobSeqOnFailure = "stop"     // default
SchedJobSeqOnFailureContinue SchedJobSeqOnFailure = "continue"

// entity.SchedJob gains
Sequence *SchedJobSequence `json:"sequence,omitempty"` // job-sequence only

type SchedJobSequence struct {
    Mode      base.SchedJobSeqMode      `json:"mode"`
    OnFailure base.SchedJobSeqOnFailure `json:"onFailure"`
    Steps     []*SchedJobSequenceStep   `json:"steps"`
}

type SchedJobSequenceStep struct {
    Job  ObjectID `json:"job"`            // another scheduled job
    Name string   `json:"name,omitempty"` // a label for logs and results
}
```

**Rules**, checked when a sequence is created or updated:

- 1 to 50 steps; a job may appear more than once.
- Each member exists and is in the sequence's reach: an app sequence's members
  are scheduled jobs of that app; an env sequence's members are scheduled jobs
  of an app in that env, not deleted. A member is never a `job-sequence`: no
  nesting in this version.
- A member may be disabled; it is skipped when the run reaches it (§2).
- `job-sequence` is allowed at the app and project env scopes only. At the env
  scope it is, for now, the only job type allowed.
- `Command`, `App`, `TargetSetting` and `CommandOutput` are empty for a
  `job-sequence`.
- The sequence's own `MaxRetry` and `RetryDelay*` are not used: retry belongs
  to each step. Its `Timeout` is the default for a step whose job sets none.
  `Priority` applies to the run.
- `GetRefObjectIDs()` includes the members, so deleting or disabling a member
  goes through the existing setting-in-use check, whatever the scopes of the
  two.

**A schedule becomes optional, for every job type.** `Schedule` may be nil: the
job never runs by itself - `createTasksForJobs` skips it - and runs only when
run by hand or as a step of a sequence. The dashboard shows "No schedule".

**A run's results**, kept in the task's args while it runs and in its output
when it ends:

```go
type SchedJobSeqRun struct {
    CurrentStep int                      `json:"currentStep"`
    Steps       []*SchedJobSeqStepResult `json:"steps"`
}

type SchedJobSeqStepResult struct {
    Job       ObjectID          `json:"job"`
    Name      string            `json:"name,omitempty"`
    Status    string            `json:"status"` // pending|running|done|failed|skipped
    ExitCode  *int              `json:"exitCode,omitempty"`
    Error     string            `json:"error,omitempty"`
    Attempts  int               `json:"attempts"`
    StartedAt time.Time         `json:"startedAt,omitzero"`
    EndedAt   time.Time         `json:"endedAt,omitzero"`
    Outputs   map[string]string `json:"outputs,omitempty"`
}
```

## 2. Running a sequence

A run is a `task:sched-job-exec` task whose `TargetID` is the sequence, created
by the schedule or by run-now, like any scheduled job's.

**The queue gains `execData.Continue()`.** When `execute` returns without error
and it was called, the queue does not mark the task done: it sets
`status = not-started`, `RunAt = now`, saves the task with its new args and
schedules it at once. Any task type may use it; `task:workflow` is changed to
use it too.

**Each execution runs one step:**

1. Read `SchedJobSeqRun` from the task's args (initialised from the sequence
   on the first execution), take the current step, load its job.
2. A job gone or disabled makes the step `skipped`, with the reason, and the
   run goes on as `onFailure` says (a skip counts as a failure for `stop`).
3. Run the job with the code that runs each job type - the `switch` of today's
   executor, moved into a function both paths call - on a child `TaskExecData`
   whose logs go to the sequence task's log, under a
   `── Step 2/5: <name> ──` header.
4. **Retry and timeout per step.** Before a step first runs, the task's
   `Config` takes the member's retry and timeout (the sequence's timeout when
   the member has none) and `Retry` is set back to 0. A step that fails with
   retries left returns its error: the queue's own retry runs the same step
   again.
5. A step that fails with none left is `failed`. With `continue` the run goes
   on; with `stop` the remaining steps are `skipped` and the task ends `failed`,
   not retried.
6. A step that succeeds is `done`. If steps remain, `Continue()`; otherwise the
   task ends - `done` when no step failed, `failed` when one did.

**The environment of a `container-command` step:**

| Variable | Value |
|---|---|
| `HIVEPAAS_SEQ_STEP`, `HIVEPAAS_SEQ_STEPS` | this step's number, 1-based, and the count |
| `HIVEPAAS_SEQ_PREV_STATUS` | the previous step's status; empty for the first |
| `HIVEPAAS_SEQ_RESULTS` | JSON: each earlier step's job, name, status, exit code and duration |
| `HIVEPAAS_SEQ_OUTPUTS` | JSON: each earlier step's outputs |
| `HIVEPAAS_SEQ_OUTPUT_<KEY>` | the previous step's outputs, one variable each |
| `HIVEPAAS_OUTPUT` | the file this step writes its outputs to |

The prefix is `HIVEPAAS_`, not `HP_`, which is the app's own configuration.

**Outputs.** A step writes `KEY=value` lines to `$HIVEPAAS_OUTPUT`
(`/tmp/hivepaas-output-<task>-<step>` in its container). After the step,
HivePaaS reads it with a second exec and deletes it: at most 64 KB, keys
`[A-Za-z_][A-Za-z0-9_]*`, upper-cased in variable names, one-line values.
Lines that break these rules are skipped with a warning in the log. A file that
cannot be read - no `cat` in the container - is a warning and no outputs, never
a failed step. Large data does not belong here: `saveToFile` puts a job's
output in storage, for a later step to read from there.

System job types (backup, cleanup, SSL renewal, backup repo cleanup) take no
input and give no outputs; their status counts like any step's.

**No overlapping runs.** A run that starts its first step while another run of
the same sequence has not ended - a task with the same `TargetID`, started,
neither done nor canceled nor failed for good - ends `done` at once, noted
"skipped: the previous run is still going".

**A worker that restarts mid-step.** Progress is saved when a step ends, so the
step that was running runs again when the task is picked up: steps run at
least once. The form and this spec say so; jobs in a sequence should be safe to
run twice.

**Cancel** stops the run at its current step; the steps left are `skipped`.

**Notifications.** A member run as a step sends none of its own. The sequence
sends one when the run ends, through its `Notification` settings: each step,
its status and duration.

## 3. API

**Project env: `/projects/{projectID}/{projectEnv}/sched-jobs`**, with the
project env settings handler and `GetAuthInEnv`, as the other env settings:

| Method | Path | Does | Access on the env |
|---|---|---|---|
| GET | `` | the env's jobs **and** those of every app in it; each item has `scope` (`project-env`, `app`) and `app {id, name}`; filters `jobType`, `appId`; search, paging, sort | read |
| GET | `/:itemID` | an env-scope job | read |
| POST | `` | create (this version: `job-sequence` only) | write |
| PUT | `/:itemID`, `/:itemID/status` | update; enable or disable | write |
| DELETE | `/:itemID` | delete | delete |
| POST | `/:itemID/exec` | run now | execute |

The list is a repository filter: `setting.object_id = <env>` or
`setting.object_id IN (the env's apps, not deleted)`. `applyProjectEnvFilter`
and its callers are unchanged. The list does not assume the env holds only
sequences: other env job types will appear in it when they exist.

**App: `/…/apps/{appID}/sched-jobs`** is unchanged, and takes
`jobType: job-sequence`.

Access to the env covers its apps: a sequence's members are not checked app by
app.

**Runs** are tasks: the existing task API lists a sequence's runs
(`TargetID`), and a run's task output carries `SchedJobSeqRun`.

**Spec export and import** carry a sequence as the `sched-job` it is.

- An app sequence travels with its app: import writes an app's scheduled jobs
  and schedules their tasks (`import_apply_apps.go`), sequences among them.
- An env sequence is exported with the env's settings. Import skips every
  `sched-job` outside an app today (`import_policy.go`, `reasonSchedulesTasks`);
  this design lifts that for the project env scope: the env's scheduled jobs
  are written, and their tasks scheduled once written, as for an app's.
- A sequence's members are references to scheduled jobs of the env's apps, and
  import maps them to the jobs it writes, as it maps the settings' other
  references. The planning confirms that the reference handling reaches an
  `ObjectID` inside `sequence.steps`, and a test pins it: an env sequence
  exported and imported into a new env runs the jobs of that env's apps. A
  member the import does not write - a job it skipped, an app it did not
  import - is reported, and the step is skipped when the sequence runs (§2).

## 4. Dashboard

- **App scheduled jobs:** `+ New Job Sequence` next to `+ New Job`, to
  `…/apps/:appId/sched-jobs/create-sequence`. The edit route opens the form the
  job's type needs. The regular form's job types do not include job sequence.
  In the list, a sequence has a "Sequence" tag and its step count.
- **Project → Scheduled jobs (new):** the project page with the env picker of
  the other env settings (`ProjectProviderSettingsScopeHeader`). An env has to
  be picked; "All" shows a note. The list is the env list API: name, type,
  owner ("Env" or the app), schedule or "No schedule", status, last run;
  filters by app and type. `+ New Job Sequence`; per row run now, enable or
  disable, edit, delete, runs. An app's job opens its app's page to edit.
- **`JobSequenceForm`**, one component for both scopes: name, status; steps -
  added from a picker of the scope's jobs (by app at the env scope, without
  sequences), reordered with up and down, an optional label each, a job may be
  added twice; `On failure` (Stop, Continue); `Mode` (Sequential; Parallel
  shown disabled, "coming"); schedule with **No schedule**; timeout, priority,
  notification. It says that a step may run twice after a restart, and
  explains `HIVEPAAS_OUTPUT` and the `HIVEPAAS_SEQ_*` variables. The parts it
  shares with the job form - schedule, timeout, priority, notification - are
  components both use.
- **Every job form** gains **No schedule** (manual, or run by a sequence).
- **A run's page** gains, for a sequence, a table of steps: number, name or job,
  status, attempts, duration, error, output keys (their values on click),
  refreshed while the run goes on. The log viewer is the existing one; the step
  headers mark the steps.
- Deleting or disabling a job a sequence uses opens the existing setting-in-use
  dialog, with a link to the sequence.

## 5. Testing

- **Go:** the rules of §1 (scopes, reach, no nesting, empty fields, 1-50
  steps); the runner of §2 - `stop` and `continue`, retry per step and its
  counter reset, a member gone or disabled, cancel, overlapping runs, the step
  that runs again after a restart; `Continue()` in the queue, and
  `task:workflow` advancing through its steps; the environment and the output
  file (parsing, the 64 KB limit, bad keys, no file); a job without a schedule
  never scheduled; setting usages across the env and app scopes; the env list
  API with jobs of the env and of its apps; export and import of an app and an
  env sequence, their members mapped to the imported jobs.
- **Locally, against the running swarm:** an env sequence of two or three
  `container-command` jobs of two apps, a later step reading an earlier one's
  output, with `stop` and with `continue`; the log, the step table, the
  summary notification.
- **Dashboard:** typecheck and lint; in a browser: create and edit a sequence
  at the app and the env scope, run it now, open a run, and delete a job a
  sequence uses.

## Not in this design

- **Parallel mode**, and a sequence as a step of another.
- **Global and project sequences.**
- **Multi-line output values** (GitHub's `KEY<<EOF`), and passing a step's
  stdout to the next one's stdin.
- **A member's own run history** showing its runs inside sequences: they are in
  the sequence's run.
