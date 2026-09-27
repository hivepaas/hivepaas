# Job Triggers: Scheduled Jobs That Run When Something Happens

A scheduled job runs on its schedule, by hand, or as a step of a job sequence.
A user who needs one to run when something happens to an app - migrate before
a deploy, warm a cache after one, page someone when a health check goes down -
has no way to say so. This design lets a job name the events that run it.

HivePaaS itself has no use for it; it is for its users.

---

## Decisions

1. **`triggers` on the job, not a new setting.** A scheduled job gains a list
   of triggers, as a GitHub Actions workflow has `on:`. One place says when a
   job runs: its schedule, its triggers, or both. Several jobs in order are a
   job sequence with a trigger.
2. **Events, first version:** `pre-deploy`, `post-deploy`, `deploy-failed`,
   `health-down`, `health-up`, `app-enabled`, `app-disabled`.
3. **Whose events.** An app's job listens to its own app. An env job names the
   apps of its env it listens to; there is no "every app".
4. **`pre-deploy` may wait.** A trigger with `wait` holds the deploy until its
   run ends, and a failed run fails the deploy. Without `wait` the job is only
   scheduled, as for every other event.
5. **Direct calls, not the event bus.** Where an event happens, the code calls
   the trigger service, which schedules the jobs. The system event bus is
   fire-and-forget pub/sub: a restarting worker would lose events.
6. **A waited-for job is a task of its own.** The deploy schedules it and
   waits for it, so the run has its history, retries and cancel as any run, and
   a sequence uses its runner unchanged.
7. **The deployment commands stay.** `preDeploymentCommand` and
   `postDeploymentCommand` remain the quick way to run one line; triggers are
   for jobs.

## 1. Data model

```go
// base
type SchedJobTriggerEvent string

const (
    SchedJobTriggerPreDeploy   SchedJobTriggerEvent = "pre-deploy"
    SchedJobTriggerPostDeploy  SchedJobTriggerEvent = "post-deploy"
    SchedJobTriggerDeployFailed SchedJobTriggerEvent = "deploy-failed"
    SchedJobTriggerHealthDown  SchedJobTriggerEvent = "health-down"
    SchedJobTriggerHealthUp    SchedJobTriggerEvent = "health-up"
    SchedJobTriggerAppEnabled  SchedJobTriggerEvent = "app-enabled"
    SchedJobTriggerAppDisabled SchedJobTriggerEvent = "app-disabled"
)

// entity.SchedJob gains
Triggers []*SchedJobTrigger `json:"triggers,omitempty"`

type SchedJobTrigger struct {
    Event base.SchedJobTriggerEvent `json:"event"`
    // Apps whose events run the job: empty for an app's job, which listens to
    // its own app; one or more apps of the env for an env job.
    Apps []ObjectID `json:"apps,omitempty"`
    // Wait holds the deploy until the run ends. pre-deploy only.
    Wait bool `json:"wait,omitempty"`
}
```

**Rules**, checked when a job is created or updated:

- Triggers are for the jobs that run in an app or an env: `container-command`
  and `job-sequence`. The system job types have none.
- At most 10 triggers; no two with the same event and the same apps.
- `wait` only with `pre-deploy`.
- An app's job: `apps` empty. An env job: `apps` one or more apps of that env,
  not deleted.
- `GetRefObjectIDs()` includes the apps of the triggers, so export and import
  map them as the job's other references.
- A job may have a schedule, triggers, both, or neither.
- A disabled job is not run by a trigger, as it is not run by its schedule.
  An app whose scheduled jobs feature is off runs none of its jobs by trigger.

**The events:**

- `pre-deploy`: in the deploy, after `preDeploymentCommand` and before the new
  version is applied; deploys from an image and from a repository alike.
- `post-deploy`: the deploy ended `done`. `deploy-failed`: it ended `failed`.
  A canceled deploy fires neither.
- `health-down`, `health-up`: any health check of the app changed state, from
  success to failure or back. The first check, and a check whose last state
  was lost from the cache, fire nothing: the code today counts them as a
  transition to save the task, and a restart would otherwise fire a
  `health-up` for every app. Choosing one health check is not in this version.
- `app-enabled`, `app-disabled`: the app's status changed between `active` and
  `disabled`, by itself or through its project's or env's status
  (`SetAppStatus`). The event fires once the change is committed.
  Each health check fires on its own: two checks going down run a job twice.

## 2. Firing and running

**`schedjobtriggerservice`:**

```go
Fire(ctx, event base.SchedJobTriggerEvent, app *entity.App, info *TriggerInfo) (*FireResult, error)
// TriggerInfo: what the event is about beyond the app, e.g. the deployment ID.
// FireResult: the tasks created, and which of them are waited for.
```

1. **The jobs that listen:** the app's own jobs with a trigger for the event and
   no apps; the jobs of the app's env with a trigger for the event naming the
   app; active jobs only, of an app with the scheduled jobs feature on. An app
   and its env hold few jobs: they are listed and filtered in code, with no
   new index.
2. **A run per job:** a `sched-job-exec` task, made as run-now makes one
   (`CreateSchedJobTask`), inserted in a transaction of its own and scheduled
   at once. Its task records what fired it - event, app, deployment - for the
   run's page, and a `container-command` step or job is told it in
   `HIVEPAAS_TRIGGER_EVENT`, `HIVEPAAS_TRIGGER_APP`,
   `HIVEPAAS_TRIGGER_DEPLOYMENT`.
3. **A failure to fire** is logged, and never fails what fired it, but for
   `pre-deploy`: a deploy that cannot list its jobs cannot tell whether one
   would hold it, and fails.

**Where `Fire` is called:** the deploy's `pre-deploy` step; the deploy's
after-commit hook, once its status is settled (`post-deploy`,
`deploy-failed`); the health check, after `calculateStateTransition`, on a
real transition; the status changes of an app, an env and a project
(`app-enabled`, `app-disabled`), through `FireAppStatusEvents` of the trigger
service, once their transactions have committed. `SetAppStatus` and
`SetProjectEnvStatus` say which apps changed; the use cases fire. The app
service cannot hold the trigger service itself: the trigger service needs the
task queue, which needs the app service, through the scheduled job and env var
services.

The transaction of its own matters: a task inserted in the deploy's
transaction is not seen by another worker until the deploy commits, and a
deploy waiting for it would wait forever.

**`pre-deploy` with `wait`.** The deploy schedules every matching job, then
waits for the runs of the triggers with `wait`: it reads their status from the
database every `WaitPollInterval`, on a connection of its own, and logs in the
deploy's log which job it waits for and how it ended.

- Every waited run ends `done`: the deploy goes on.
- One ends `failed` or `canceled`: the deploy fails, its error naming the job.
- The wait passes the job's timeout, or `WaitTimeout` for a job without one
  and for a sequence (whose timeout is each step's): the deploy fails; the run
  goes on - stopping a migration half way is worse than letting it end.
- The deploy is canceled while it waits: the waited runs are canceled.
- Runs of triggers without `wait` are scheduled and not waited for.

A waiting deploy holds a worker of the task queue (10 by default); the
timeout bounds how long.

**Configuration** (`config/tasks.go`):

```go
type TaskTriggers struct {
    WaitPollInterval time.Duration `toml:"wait_poll_interval" env:"HP_TASKS_TRIGGERS_WAIT_POLL_INTERVAL" default:"3s"`
    // For a job without a timeout of its own.
    WaitTimeout time.Duration `toml:"wait_timeout" env:"HP_TASKS_TRIGGERS_WAIT_TIMEOUT" default:"30m"`
}
// Tasks gains Triggers TaskTriggers `toml:"triggers"`
```

**Edge cases:**

- An app's job that listens to `app-disabled` of its own app has no container
  to run in and fails; the form says so. An env job runs it in another app.
- One event may run several jobs; they run side by side. A sequence runs them
  in order.
- A sequence's overlap check applies to its triggered runs. A plain job's runs
  may overlap, as with run-now.

## 3. API and dashboard

**API.** No new endpoint. A job's request and response, at the app and the
env scope, carry `triggers: [{event, apps: [{id}], wait}]`; the response names
the apps (`apps: [{id, name}]`). The rules of §1 are the request's validation.
`make gen-swag`.

**Dashboard:**

- **The job form and `JobSequenceForm`** gain a **Triggers** block under
  Scheduling, one component for both: per row an event, the apps (env scope
  only, several), "Deploy waits for this job" (for `pre-deploy`), remove. It
  warns about an app's job that listens to its own app's `app-disabled`.
- **The job lists**, app and env: the triggers as tags, e.g. `post-deploy`,
  `pre-deploy · waits`. A job with triggers and no schedule still shows
  "No schedule".
- **A run's page:** in the run's details, "Triggered By: post-deploy of a1"
  and the deployment's ID. The task API gives a `sched-job-exec` task a
  `trigger {event, app {id, name}, deploymentId}`; a run's references include
  the app, for its name. No link to the deployment: the task does not know the
  app's env, which the route needs.
- **The deploy's log** says which jobs it waits for and how they ended; no new
  view.

**Export and import.** Triggers are part of the job's data and travel with it;
their apps are references to apps, which a bundle writes as the IDs they had
where it was made; import maps each to the app with that bundle ID here, as
it does an app's other references to apps.

**MCP.** `plan_create_sched_job` describes `triggers`: an app's job listens to
its own app's events.

## 4. Testing

- **Go:**
  - the rules of §1;
  - the jobs that listen: an app's own, an env's naming the app; a disabled
    job, an app with the feature off, another event, an app not named - none;
  - the runs: the task records the event, app and deployment; made in a
    transaction of its own and scheduled; a failure to fire is logged and does
    not fail what fired it;
  - the wait, with a fake task repository and a short poll interval: all done
    goes on; one failed or canceled fails the deploy naming the job; the
    timeout fails it; a canceled deploy cancels the waited runs; runs without
    `wait` are not waited for;
  - the events: a deploy `done` fires `post-deploy`, `failed` fires
    `deploy-failed`, `canceled` fires nothing; a health check fires only on a
    transition from a known state; an app's status change fires
    `app-enabled` / `app-disabled`;
  - `HIVEPAAS_TRIGGER_*` in a step's environment;
  - the configuration defaults (3s, 30m);
  - an env job's trigger apps imported as this installation's apps.
- **Live, on the Linux server:** a `post-deploy` job after a deploy;
  `pre-deploy` with `wait` - a job that exits 1 fails the deploy, one that
  exits 0 lets it go on; a deploy canceled while it waits; a health check that
  fails then recovers - one `health-down`, one `health-up`; an app disabled
  then enabled; a restart fires no `health-up`.
- **Dashboard:** typecheck, lint, prettier; in a browser: add, edit and remove
  triggers at the app and the env scope, the tags in the lists, "Triggered by"
  on a run's page.

## Not in this version

- An event of every app of an env.
- Choosing one health check for `health-down` / `health-up`.
- More events (app created or deleted, a backup done, a certificate renewed).
- Waiting on events other than `pre-deploy`.
