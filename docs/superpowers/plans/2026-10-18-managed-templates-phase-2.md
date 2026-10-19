# Managed Templates Phase 2 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Take managed PostgreSQL from "safe" to "safe to run a team on":
- the archive holds from a database's first minute;
- every deploy can leave a way back;
- dumps are kept for months;
- a restore can be switched in;
- a database can have read replicas, back up from one, and be promoted onto one;
- volumes that every node reaches can be used;
- the WAL-G binary is scanned like an image.

**Architecture:**
- **Scripts in the repo (Task 0).** The WAL-G scripts' sources and their generator move
  into `app-templates`. Today only the generated template is committed.
- **Small fixes first (Tasks 1-2).**
  - The entrypoint checkpoints once the server is up, so `archive_timeout` holds from
    the start.
  - A deploy can mark a restore point on a managed database first.
- **Kept longer (Task 3).** The engine owns a `data-backup` job that streams
  `pg_dump -Fc` into a Kopia repository. This is the existing job type, with no
  new mechanism.
- **Switch (Task 4).** A restored app takes its source's consumers. Their
  `${<key>.VAR}` references are rewritten and the consumers are redeployed. Names stay
  as they are: a key never changes.
- **Members (Tasks 5-7).**
  - Replicas are logical children the engine creates, as it already creates the
    verifier. Each member is pinned by its own volume.
  - The nightly backup can run on a replica.
  - Promotion is manual and fenced, and its stable address is a multi-host connection
    string.
- **Later in the phase (Tasks 8-9).** Unpinned volumes, and scanning the artifact.

**Spec:** `docs/superpowers/specs/2026-10-17-managed-templates-design.md`. The phase-2 row
of §10 is this plan's scope. §9 Clusters gives the shape of Tasks 5-7. Where this plan
decides something §9 leaves open, the task says so. The spec is updated in the task's
commit.

## How this plan works

As in phase 1:
- **No patches were written in advance.** Each task is written test-first when it is
  executed, and is done when its checks pass.
- **File references come from reading the code at this plan's commit** (2026-10-18,
  after phase 1 and its live checks). Where the code has moved since, follow the code.

**Repositories:**
- this one (the backend);
- `../app-templates`, Tasks 0, 1, 5;
- `../hivepaas-dashboard`, one task per screen, done with the backend task it belongs to;
- `../hivepaas-cli`, whenever the API level moves.

Work on a branch per repository, in a worktree.

**Global constraints, from phase 1:**
- **Secrets:** "S3 keys and the libsodium key reach the container as Swarm secrets,
  never as env and never on argv. Logs name the step, not the value." A replication
  password is held to the same rule.
- **Git:**
  - Commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.
  - Merge locally, never push.
  - Never `git stash`.
  - Never touch the user's uncommitted work.
- **Live checks** run on the dashboard's e2e dind stack (`e2e/env/up.sh`, with
  `HP_TEMPLATES_SRC` pointing at an app-templates checkout on the branch under test).
  It is isolated from the machine's own swarm. Nothing is deployed to the user's swarm.
- **The API level** moves when a request body gains a field (`make gen-swag`). The CLI is
  regenerated in the same piece of work.

**Order:** 0, then 1 and 2, then 3, then 4. 5-7 come in that order. 8 and 9 come last
and can be dropped from the phase without touching the others.

## Task 0 (`app-templates`): the WAL-G scripts' sources in the repository

Phase 1 built `templates/postgres-managed.yaml` from script sources and a generator
that were never committed. Only the generated file is in the repository. Every later
task edits a script, so the sources come first.

**Files:**
- Create `scripts/walg/`:
  - `template-head.yaml`, the template without its config files;
  - one file per script: `walg-lib`, `walg-run`, `walg-fetch`, `walg-nightly`,
    `walg-health`, `walg-mark`, `walg-restore-test`, `walg-entrypoint`.
  These are taken from the committed template, so the generator reproduces it byte for
  byte.
- Create `scripts/walg/generate.sh`. It writes `templates/postgres-managed.yaml`, and
  `-check` exits 1 when the committed file differs.
- Modify the CI workflow that runs `lint` and `index -check` to run
  `scripts/walg/generate.sh -check` too.
- Modify the README's managed section to say where the scripts live and how the
  template is made from them.

**Tests:**
- `generate.sh` against the committed template gives no diff.
- `scripts/test/walg/run.sh` passes on 17 and 18, unchanged.

- [x] Commit: `walg: the scripts' sources, and the template generated from them`.

**As built** (`../app-templates`, commit 02c4eae, merged as ad53dda):
- **The sources.** The header and the eight scripts were cut from the committed
  template as text, de-indented, rather than parsed as YAML. They are identical to the
  sources phase 1 generated it from.
- **`generate.sh`** is bash, so CI needs nothing more. It writes the template from them,
  byte for byte. `-check` exits 1 and shows the start of the diff when the template
  differs.
- **Checks.** Lint and `index -check` pass. shellcheck is clean on `generate.sh`.
  - The scripts themselves have warnings for variables `walg-lib` defines. shellcheck
    cannot follow the source path, so CI does not run it on them.
- **Not rerun:** the harness. It reads the scripts from the template, which did not
  change.

## Task 1 (`app-templates`): archive_timeout holds from the start

The live checks showed that PostgreSQL forces no segment switch after a start until its
first checkpoint, and that plain images behave the same. A write in the first minutes
can therefore wait up to checkpoint_timeout + archive_timeout, about 6 minutes, to be
archived.

**Files:**
- `scripts/walg/walg-entrypoint`, on the normal start path, before
  `exec docker-entrypoint.sh`: a background subshell waits for `pg_isready`, runs
  `CHECKPOINT` once as the superuser over the socket, and exits. Its failure is logged
  and never stops the server.
- `scripts/test/walg/run.sh`: a check that a write in the first minute after a start
  is archived within archive_timeout + 30 s, on 17 and 18.
- `README.md`: the managed section's RPO sentence.

**Tests:**
- The new harness check fails before the change and passes after it.
- The rest of the harness passes.
- On the e2e stack, the same check runs against a freshly created managed app.

**Notes:**
- `engineApi` stays 1: no contract changes.
- A restored app is not affected. Its promotion ends recovery with a checkpoint.

- [x] Commit: `walg-entrypoint: a checkpoint once the server is up`.

**As built** (app-templates 6d653f4, merged as 70fe28b; backend 11cf983e, merged as
4be5c071):
- **The backend changed too.** Only a restored database started through
  `walg-entrypoint`; a new one started as the image does. `setServerArguments` now
  sets the entrypoint for every managed database. Its non-restore path only execs
  `docker-entrypoint.sh`, so nothing else changes. Task 5's standby mode needs this
  anyway.
- **`serve` in the entrypoint**, used by both of its `exec docker-entrypoint.sh` paths
  (a normal start, and a restored database's later starts):
  - when the command is `postgres`, it starts `checkpoint_once` detached twice. The
    verifier's `sleep infinity` gets none.
  - `checkpoint_once` waits up to 10 minutes for a server whose
    `listen_addresses` is not empty. That skips the socket-only server the image runs
    while it initializes a data directory.
  - It runs `CHECKPOINT` once, logs either way, and always exits 0. Without an init,
    the postmaster reaps it, and would take another exit for a backend's crash.
  - HivePaaS runs containers with docker's init unless the image has one of its own
    (`deployment_container_init.go`), so in practice PID 1 reaps it.
- **Harness:**
  - the archive database starts through `walg-entrypoint` with `--init`, as HivePaaS
    does;
  - a new check, `archived_unasked`, requires a write in the first minute to be
    archived within archive_timeout + 30s, with no switch asked for. It failed before
    the change. All 114 checks pass on 17 and 18 after it.
- **Template.** The change to the template moves `index.json`: run `apptemplate index`
  after `generate.sh`.
- **README.** No change: the template's description already says a failure loses about
  a minute of writes, which now holds from the first minute.
- **Live check, on the e2e stack:**
  - PG18 started 15:15:54, a write followed at 15:15:55, and it was archived at
    15:16:54, archive_timeout after the start.
  - PG17 was archived at 15:17:09, also archive_timeout after its start.
  - The log says `checkpointed once the server was up`.
- **To ship:** the template change reaches installations only through a new pin in
  release.json, signed again, after the phase's other template tasks.

## Task 2: a restore point before every deploy

An app chooses a managed database in its environment, and each of its deployments
marks a restore point there before its pre-deploy command runs.

**Decided here (the spec says only "an app that links to a managed database"):**
- **The database is chosen, not inferred.** Links are not stored: the link dialog only
  writes `${<key>.VAR}` text. The setting names the database by id, and the dashboard
  offers the managed databases of the environment.
- **The point's name** is `deploy-<app key>-<yyyymmddhhmmss>`, with the key shortened
  so that the name fits 63 bytes.
  - Deployments have no number, only a ULID.
  - A name that already exists counts as done, not as a failure.
- **A failed mark fails the deployment**, before anything changes: the point is what the
  option asks for. The error says the database and the reason, so the person can turn
  the option off and deploy.
- **Points made by deploys are removed by the engine** once no base backup reaches
  them. A point a person made is never removed for them.
  - Each point gets `Origin` (`person` or `deploy`).
  - The nightly run prunes `deploy` points whose `restorableUntil` has passed.
  - Without this, one point per deploy would grow the list forever.

**Files (backend):**
- **Move the mark into a service.**
  - `createRestorePoint`, `runMark` and the managed load move from
    `usecase/appmanageduc/restore_point.go` into `managedservice` as
    `MarkRestorePoint(ctx, req)`. That method runs its own transaction, so the
    database's row is never locked inside a deploy's long transaction.
  - ARCHITECTURE.md §1: a service cannot import a usecase.
  - `appmanageduc` calls the service.
- **The setting.**
  - `entity/app_deployment_settings.go`: `RestorePointBeforeDeploy *ObjectID`, the
    database. It is copied into the deployment's snapshot as the whole settings are.
  - Request and response DTOs in `usecase/appsettingsuc/appsettingsdto/deployment_settings_*.go`.
    A field left out of the request is erased, since `ToEntity` builds a new entity.
  - Validation: the database is in the same environment, is managed, and its engine is
    postgres.
- **The deploy step.**
  - `service/appdeploymentservice/appdeploymentserviceimpl`: a
    `deployStepRestorePoint` after `lockDockerServiceForDeployment` and before
    `deployStepExecCmd(..., true)`, in the repo, image and function flows.
  - `appdeploymentserviceimpl/service.go` gains a `managedService` dependency.
  - While in that code, fix the step label: `deployment_cmd_exec.go:37-46` labels a
    pre-deploy command `stepPostDeployCmd`.
- **Clearing.** Clone (`appcloneserviceimpl/clone_2_settings.go`) and the verifier copy
  (`appmanageduc/verifier.go`) clear the field, as they clear the pre-deploy command.
- **Origin and pruning.**
  - `entity.ManagedRestorePoint.Origin`.
  - Pruning runs in `tasks/taskschedjobexec/managed_nightly.go` after retention: it
    deletes `deploy` points whose `restorableUntil` has passed, and audits it.
- **Response fix.** The create-point response computes `restorableUntil`, as Status
  does. It was `null`.
- `make gen-swag`: the API level moves to 5.

**Files (dashboard):**
- **Deployment settings, run configuration:** a select, "Mark a restore point before
  each deploy on", listing the environment's managed databases. It sits beside the
  pre-deploy command.
- **Restore points list:** the origin, with deploy points tagged.

**CLI:** regenerate.

**Tests:**
- The step runs only when set, and before the pre-deploy command.
- A failed mark fails the deployment and leaves the service untouched.
- An existing name counts as done.
- The name is shortened to 63 bytes.
- Pruning removes only `deploy` points, and only past `restorableUntil`.
- Settings round-trip, and clone clears the field.
- Live: an app linked to a managed database deploys, the point is listed, and
  restoring to it gives the rows from before the deploy.

- [x] Commit per repository.

**As built** (backend fe17440b + 50ff1dd7, merged as 81e1653c; dashboard 4196f0b7,
merged as a1053a34; CLI 711f734, merged as 8cacdad):
- **The service.** `managedservice.MarkRestorePoint(ctx, *MarkPointReq)` runs in its own
  transaction and locks the database's app row `FOR UPDATE`. The request carries
  `Origin`, `ExistingOK`, and a `Record` callback, which the person's endpoint uses for
  its audit row. `PruneDeployPoints` sits beside it.
  `appmanageduc` no longer needs `schedJobExecService`.
- **The setting** is `RestorePointBeforeDeploy entity.ObjectID` (`omitzero`), not a
  pointer. Validation (`appsettingsuc/deployment_restore_point.go`) refuses:
  - the app itself;
  - an id that is not an app of the project;
  - an app of another environment;
  - one with no active managed setting, or whose engine is not postgres.

  Each is `ERR_ARGUMENT_INVALID` on `restorePointBeforeDeploy`, with the reason as
  extra detail.
- **The name** comes from the deployment's `CreatedAt` (UTC), not from now. A
  deployment run again asks for the same name, and `ExistingOK` answers it as made
  without marking.
- **The step** is `restore-point`. It logs `Marking restore point <name> on the app's
  database...` and `Restore point <name> is on <db>, at <LSN>.`. A failure is
  `ERR_APP_DEPLOY_RESTORE_POINT_FAILED`, which names the database and the reason's
  words, without its code. A database whose service runs no container gives
  `ERR_ACTIVE_CONTAINER_NOT_FOUND`, not a generic failure.
- **Pruning differs from the plan in two ways.**
  - `Engine.StaleDeployPoints` drops `deploy` points that have no kept base backup
    before them. It reads the `BACKUPS` the retain step printed, which is what
    `restorableUntil` comes from. With no listing, it prunes nothing.
  - The nightly logs the names it removed in its run (`No base backup kept reaches
    these deployments' restore points any more; they are removed: ...`). It writes no
    audit rows: deployments' points are not audited when they are made either. A
    failed prune is logged and never fails the nightly.
- **API level** stays 4. Level 4 is unreleased, so `gen-swag`'s move to 5 was undone,
  keeping the new fingerprint.
- **Dashboard.**
  - The select is labelled "Restore point before deploy" and sits under the
    pre-deploy command.
  - It lists the environment's `database` apps whose managed engine is postgres
    (`useManagedDatabases`). It asks for each engine with the Database tab's query key.
  - A saved database the list no longer has stays shown as saved.
  - The restore points table tags a deployment's point with a "Deploy" badge.
- **CLI.** `carry.DeploymentSettings` copies the response into the request through
  JSON, so `hivepaas deploy` keeps the setting as read.
- **Also fixed:** `deployment_cmd_exec.go` labelled the pre-deploy command step as
  post-deploy.
- **Known limits.**
  - A spec export carries the database's app id as is. After an import into another
    project, the deploy fails with the error above until the setting is chosen again.
  - The other steps' `defer s.addStepEndLog(ctx, data, start, err)` evaluates `err`
    when deferred, so their end logs never carry the error. The new step uses a
    closure. The rest is left as it was.
- **Tests:**
  - unit tests for the step, the name, the mark and its refusals, pruning, the
    setting's validation and round trip, and clone clearing;
  - the e2e level-0 spec checks the "Deploy" badge, and all 4 tests pass.
- **Live check, on the e2e stack (`lc/check8.sh`):**
  - the setting refuses the app itself and an unknown id, and round-trips pg18;
  - the deploy marks `deploy-lc-web-20261018155801` at `0/8002388`, listed with
    origin `deploy` and a `restorableUntil`;
  - a restore to it gives exactly the 110 rows from before the deploy, not the 10
    written after;
  - with pg18 scaled to 0, the deploy fails with `...could not be made: No active
    container found for app 'pg18'`, and the web service still runs `sleep 1d`;
  - with both points' `at` moved before every base, the nightly removes the
    deployment's point and keeps the person's;
  - the dashboard's select shows pg18 chosen among None, pg18, pg17 and the restored
    app.

## Task 3: a nightly `pg_dump` into a Kopia repository

**Decided here:**
- **The engine owns a `data-backup` job.** It does not use a new mechanism:
  - source `command`, `pg_dump -Fc`;
  - file `<db name>.dump`;
  - restore command `pg_restore --clean --if-exists --no-owner -d <db>`;
  - the repository chosen.
  It streams to Kopia, keeps per-job retention through the repository's policy, and
  restores through the snapshots view, all as it does today.
- **The repository** is the archive's backup repository when the archive is one.
  Otherwise the person chooses one. Dumps are off until then.
- **The schedule is the person's.** It is a cron expression read in the installation's
  timezone, as the base backup's is.
  - It can be any frequency, not only nightly: every 6 hours is `0 */6 * * *`.
  - It is set in the create form, and changed later in the Database tab's Settings or
    with `PUT /managed` (`dumpSchedule`).
  - Clearing the repository turns dumps off.
  - The default, `0 3 * * *`, runs after the base backup's `0 2 * * *`.
  - A schedule that fires within the same minute as the base backup's is accepted, with
    a note that both read the whole database at once.
  - The job is managed, so the generic jobs screen shows it read-only, as it shows the
    others. "Run now" stays.
- **The dump runs in the database's container,** or in the verifier when the restore
  test runs there, to spare the database. The verifier holds a copy only during a
  test, so this would mean dumping from the archive.
  - That last is left out of this task: the job runs in the database's container, and
    this is noted in the spec.

**Files (backend):**
- **Entity.** `ManagedPostgres.Dump {Repo ObjectID, Schedule string}` and
  `ManagedPostgresJobs.Dump`. `GetRefObjectIDs` covers the repository.
- **Engine.**
  - `postgres/configure.go` creates the job with `ManagedBy`, `DataBackup` set and
    `Command` nil. The `job()` helper always sets `Command`, so it is extended.
  - `postgres/reconfigure.go` adds, changes or removes the job.
  - `postgres/status.go`: the last dump's time and size come from the job's
    `SNAPSHOT_ID` and `SNAPSHOT_SIZE_BYTES` outputs.
- **Where job roles are listed:**
  - `appmanageddto/settings.go` (`jobs.dump`, `dump`);
  - `appmanageduc/restore.go` copies the schedule into a restored app;
  - `appmanageduc/unmanage.go` removes the job;
  - `usecase/apptemplateuc` create request: `managed.dump`.
- **Status DTO:** `lastDump`.
- `make gen-swag`.

**Files (dashboard):**
- **Create form:** "Daily dump" with a repository select.
- **Database tab:**
  - the last dump in the status cards;
  - Settings: the repository and the schedule;
  - a link to the app's backup snapshots.

**Tests:**
- Configure and reconfigure create, change and remove the job.
- The job's command, file name and restore command.
- Status reads the outputs.
- Live: the job makes a snapshot on the e2e stack's volume repository, and restoring it
  into a fresh `postgres` app gives the rows.

- [x] Commit per repository.

**As built** (backend c9af051d, merged as e9ed32c7; dashboard f9ea8ae9, merged as
00d56438; CLI 94a3e06, merged as c1a8cec):
- **The job** is "Database dump", a `data-backup` job with source `command`, made by
  `postgres/dump.go`'s `newDumpJob`. Configure and Reconfigure share `newJob`.
- **The commands run through `sh -c`.** A job's plain command is split into words, not
  given to a shell, so the image's variables need one:
  - dump: `pg_dump --format=custom --username="${POSTGRES_USER:-postgres}"
    --dbname="${POSTGRES_DB:-…}"`;
  - restore: `pg_restore --clean --if-exists --no-owner --single-transaction`, same
    user and database.

  `--single-transaction` is added to the plan's flags: a restore that fails leaves the
  database as it was. Neither command needs a template change, so databases made
  before this have dumps too.
- **The file name is `<app key>.dump`, not `<db name>.dump`.** The database's name is in
  the kind setting, which the engine does not read when Reconfigure turns dumps on later.
  The restore command reads `$POSTGRES_DB` anyway.
- **The repository.** The create request's `managed.dump` is `{repo: {id} | null,
  schedule}`.
  - Left out, a database archiving into a backup repository is dumped into it, and one
    archiving into cloud storage is not dumped.
  - `repo: null` turns dumps off.
  - A repository is checked to be one the environment sees, and active.
  - The dashboard always sends `dump` and works out the default itself.
- **`PUT /managed`** takes `dumpRepo`, where `{"id": ""}` turns dumps off, and
  `dumpSchedule`. The schedule is kept in `ManagedPostgres.Dump` while dumps are off, and
  the settings show it. A database made before has no schedule kept, so the engine's
  `0 3 * * *` applies.
- **Reconfigure can now make and remove jobs.**
  - `ReconfigureReq` carries `ManagedID` and `Now`; `ReconfigureResp` adds `Created` and
    `Deleted`.
  - `appmanageduc.storeReconfigured` inserts and schedules what was made.
  - It marks removed jobs deleted and reschedules them, which drops their coming runs.
  - Changing only the repository changes the job without moving its schedule.
- **A restore copies the dump settings.** A dump job made while the database is being
  restored waits, disabled, like the other jobs. The finish step turns on every job the
  engine made. Unmanage removes the dump job with the others.
- **Status** has `lastDump` (succeeded, ended, snapshot id, bytes), read from the run's
  `SNAPSHOT_ID` and `SNAPSHOT_SIZE_BYTES`.
- **A schedule in the same minute as the base backup's is not refused.**
  - The status suggests `dump-apart` when the two schedules fire in the same minute in
    the coming week.
  - The settings form also notes when the two expressions are the same.
- **Also fixed:** for a managed job, `keepManagedOutputs` replaced the run's output with
  its outputs, which would have dropped the snapshot the run's page links. A managed data
  backup's run now keeps both: `{snapshotId, sizeBytes, outputs}`.
- **API level** stays 4. After resetting `api-level.json`, run `make gen-swag` again so
  that `swagger.json`'s `x-api-level` is 4 too. The CLI reads its level from there, and
  the first regeneration came out as 5 before this was fixed.
- **Dashboard.**
  - The create form has "Dump into" (the archive's repository, no dumps, or another
    repository) and "Dump schedule".
  - The Database tab's settings have the same two fields.
  - The tab has a "Last dump" row under the status cards, with a link to the app's
    Backup Snapshots, and the wording of `dump-apart`.
- **Not done:** dumping from the verifier (as decided above), and a template default for
  the dump schedule. The engine's default is used.
- **Tests:**
  - engine: the job, its commands and file, turning dumps on, moving them and turning
    them off, the kept schedule, a bad schedule, a restoring database, the last dump,
    and `dump-apart`;
  - usecase: turning dumps on and off through `reconfigureManaged` and the settings
    view, where the create request dumps, refused repositories, and the restore copying
    the settings;
  - the managed output;
  - e2e level-0: the create request's default dump, the "Last dump" row and its link,
    and saving a moved or stopped dump. All 6 pass.
- **Live check, on the e2e stack (`lc/check9.sh`):**
  - an unknown repository is refused (404);
  - dumps turned on make the job, managed, with file `pg18.dump`;
  - a run took snapshot `3ea53dbd…` (2300 bytes), which the status, the run's page and
    the app's snapshots list all name;
  - restoring it into a fresh `postgres` app (18, debian) gave the 139 rows dumped and
    not the 5 written after;
  - a schedule of `0 2 * * *` brought `dump-apart`, and `30 3 * * *` cleared it;
  - dumps turned off removed the job (404) and kept the schedule.

  A volume repository needs a volume with `bindOptions.directory`, as the e2e helpers
  make; a plain local volume is refused with `ERR_BACKUP_REPO_VOLUME_PATH_UNRESOLVED`.

## Task 4: switch consumers to a restored database

§4 step 6 left switching to the person ("Swapping names is Later"). This task does the
part that can be done without renaming.

**Decided here:**
- **A key never changes,** so a swap of names is not attempted. A Swarm service cannot
  be renamed, and the key is the DNS alias and every `${<key>.VAR}`.
- **Switching repoints the consumers.**
  - In every app of the environment whose app-scope env vars reference
    `${<source key>.`, it rewrites those references to the restored app's key. This is
    the replacer the preview database clone already uses
    (`apppreviewserviceimpl/clone_db_apps.go`).
  - It redeploys the consumers it changed, if the person asks.
  - Preview `AppsToClone` entries that name the source by id are repointed too.
- **The source is left running.** Its jobs are paused with a note on the Database tab,
  so two databases do not both back up as "the" database. The person stops or deletes
  it once the switch is checked.
- **Only a restored app can switch,** and only from its own source (`Restore.SourceApp`),
  once it is ready.

**Files (backend):**
- `usecase/appmanageduc/switch.go`, at `POST /managed/switch` with body
  `{redeploy bool, consumers []id}`:
  - it lists consumers (`GET /managed/switch` answers the candidates);
  - it rewrites the env settings in one transaction;
  - it queues redeploys;
  - it pauses the source's managed jobs;
  - it is audited.
- A service helper for the reference rewrite, shared with the preview clone.
- `make gen-swag`.

**Files (dashboard):** on a ready restored app's Database tab, "Switch consumers to this
database". It lists the consumers found, with a checkbox for each and "redeploy them".
The confirmation names the source, which keeps running.

**Tests:**
- Consumers are found by key reference, and only in the same environment.
- The rewrite replaces only `${<source key>.` prefixes.
- Redeploys are queued.
- The source's jobs are paused.
- A database that was not restored, or not ready, is refused.
- Live: an app reading the source switches to the restored database and reads its rows
  after the redeploy.

- [x] Commit per repository.

**As built** (backend 4af73af4, merged as 51392d5f; dashboard 0af7e4b4, merged as
d9b41b09; CLI 6e9bd1f, merged as 3661a52):
- **Endpoints.**
  - `GET /managed/switch` answers `{source {id, name, key}, consumers [{app, envVars,
    previewClone, deployRestorePoint}], switchedAt}`.
  - `POST /managed/switch` takes `{consumers: [{id}], redeploy}` and answers the apps
    switched and the deployments queued. Problems applying the variables come back as
    `meta.warning`.
  - Both refuse a database that was not restored (`ERR_APP_MANAGED_SWITCH_NOT_RESTORED`),
    or not ready (`…_NOT_READY`), with 412.
  - Both refuse a source that is no longer an app of the environment
    (`…_SOURCE_GONE`): the key its consumers name cannot be known.
  - A consumer not found is refused with 400.
- **Consumers** are the apps of the environment that name the source in any of three ways:
  - a variable, not literal, with `${<source key>.`;
  - its id among their previews' `AppsToClone`;
  - their deployments' `restorePointBeforeDeploy` (Task 2's setting). The plan did not
    list this one.

  Not consumers: the source, the restored database, and previews (`ParentID`). Previews
  keep the variables they were made with until they are made again. Only app-scope
  variables are read; a project's or environment's variables that name an app's key are
  not.
- **The rewrite** is `envvarhelper.RefReplacer` / `RewriteRefs` / `RefersTo`, matching
  `${<key>.` whole. The preview's database clone now uses it too, in place of its own
  replacer.
- **Variables are always applied; redeploy adds a deployment.** The live check found
  that a deployment does not apply runtime variables, because saving them does.
  - So the switch builds the consumers' variables in the transaction. This checks that
    the references resolve.
  - After commit, it applies them to the running services, as the variables screen does.
  - With `redeploy`, each consumer whose variables changed is also deployed. This is for
    what reads the database during the build or in the pre-deploy command. The
    variables are applied before the deployments are queued.
- **The source** keeps running, and its archiving goes on: `archive_command` is the
  server's, not a job.
  - Its active managed jobs are disabled, and their coming runs dropped.
  - Its setting records `switchedTo {app, at}`. The restored database records
  `restore.switchedAt`.
  - Its Database tab shows a banner saying the jobs are paused. Archiving there reads
    "Behind" once the paused health check goes stale.
  - Nothing resumes the jobs: the way on is to stop, unmanage or delete the source.
- **Switching again** is allowed. The list shows whatever still names the source.
- **Audited** on the restored app (section `managed`: switchedFrom, consumers, redeploy).
- **Dashboard.** A ready restored database has a "Switch consumers" section:
  - every consumer is checked, with what it uses;
  - there is a "Redeploy them" option;
  - the PopConfirm names the source and says it keeps running, its jobs paused.
- **Tests:**
  - envvarhelper: matching the key whole, rewriting only the key given, literals left
    alone;
  - usecase: finding consumers by each of the three ways and leaving out the
    source, target, previews and literals; repointing each setting; refusing a database
    that was not restored, or not ready; pausing only the engine's active jobs;
  - e2e level-0: the switch section, its payload, and the source's banner. All 8 pass.
- **Live check, on the e2e stack (`lc/check10.sh`, `check10b.sh`):**
  - two readers used `PGURL` naming pg18 (144 rows), and lc-web marked deploy points on
    pg18;
  - the switch on pg18 itself was refused (412), and so was a non-consumer (400);
  - the restored database listed all three;
  - switching lc-reader without redeploy rewrote `PGURL` to the restored key, and it read
    the restored 110 rows, while lc-reader-2 still read pg18;
  - pg18's nightly, health, restoreTest and dump jobs were disabled, and both databases
    recorded the switch;
  - switching lc-reader-2 and lc-web with redeploy deployed lc-reader-2 only, since
    lc-web's variables did not change. lc-reader-2 then read 110 rows, lc-web's deploy
    points moved to the restored database, and no consumer was left.

## Task 5: asynchronous read replicas

**Decided here (§9 gives the shape):**
- **A replica is a member the engine creates,** a logical child of the primary, as the
  verifier is (`appmanageduc/verifier_provision.go`):
  - it copies settings with new ids;
  - it has its own pinned volume, so it lands on its own node;
  - artifacts are ensured on that node;
  - its key is `<primary key>-r<n>`.
- **It is seeded from the archive, not from the primary.**
  - `walg-entrypoint` gains a `standby` mode: `backup-fetch` from the primary's prefix,
    `standby.signal`, `primary_conninfo` and `restore_command` reading the archive as a
    fallback.
  - This keeps the load off the primary, as §9 prefers.
- **Replication user.**
  - The engine creates a `replication` role with a generated password, held as a
    Swarm secret file. It is never in env and never in `primary_conninfo` on disk: the
    replica writes a `passfile` from the secret.
  - The primary's entrypoint appends `host replication replication all scram-sha-256`
    to `pg_hba.conf` once, and reloads.
- **No slot by default.** The archive is the fallback for recycled WAL. A slot with
  `max_slot_wal_keep_size` is an option.
- **Archiving.** Only the primary archives: a standby with `archive_mode=on` does not.
- **The archive must be off-site or shared.** A same-node volume archive cannot reach
  other nodes, which the verifier already requires.
- **Lag in Status,** from the primary's `pg_stat_replication`, read by the health job.

**Files:**
- **`app-templates`:** `walg-entrypoint` standby mode, the `pg_hba` line and the
  passfile; harness checks for a standby seeded from the archive that follows the
  primary.
- **Backend:**
  - **Entity:** `ManagedPostgres.Members []{App ObjectID, Role, Node}` and the
    replication secret.
  - **Use case:** `appmanageduc/replica.go`, at `POST /managed/replicas` with
    `{dataVolume}` and `DELETE /managed/replicas/:id`.
  - **Engine:** `postgres/replica.go` holds the member's doc and env, and the primary's
    replication settings (`max_wal_senders`, `hot_standby`).
  - **Status:** members, with lag.
  - **Locks:** a member's fields are locked as the primary's are.
  - `make gen-swag`.
- **Dashboard:** the Database tab lists members, with their node and lag, and has "Add
  replica" (a volume picker) and "Remove".

**Tests:**
- The member's doc, its secrets and env, and its pin.
- The primary's settings.
- Status with lag.
- Removal deletes the member.
- Live: a replica on a second volume reads rows written on the primary, with lag in
  Status.

- [x] Commit per repository.

**As built** (app-templates b3f1647, merged as 2d5b489; backend 70e963fe, merged as
69e6fb5f; dashboard 6ed84be6, merged as 6c524f22; CLI cc0f625, merged as b9d6a45):
- **Template.**
  - **`walg-entrypoint` has a standby mode**, entered when `HIVEPAAS_STANDBY_HOST` is
    set:
    - an empty (or half-fetched) data directory is seeded with `backup-fetch … LATEST`
      from the archive in the variables, which is the primary's; it never runs
      `initdb`, and a replica it cannot seed idles;
    - the passfile is written to `/dev/shm` at each start, from the file
      `HIVEPAAS_STANDBY_PASSWORD_FILE` names;
    - `primary_conninfo` (user `hivepaas_replica`, `application_name` =
      `HIVEPAAS_STANDBY_NAME`), `restore_command` (`walg-fetch`),
      `recovery_target_timeline = 'latest'` and `standby.signal` are written again at
      each start. Task 7 can therefore repoint a replica by changing its variable.
    - The container health check passes while it is seeding.
  - **`walg-replication` replaces the plan's "the primary's entrypoint appends the
    line".** It runs on the primary, through `RunCommand`, with the password on stdin.
    It makes or updates the role, adds the `pg_hba` line once (`show hba_file`), and
    reloads. The role and the line live in the data directory, so a restore of the
    primary carries them. Nothing on the primary holds the password as a file.
  - **`walg-health`** prints `REPLICAS=name@state@bytes@seconds …` from
    `pg_stat_replication`.
  - `replication_user` is in `walg-lib`.
  - **Harness** `test_standby`, on 17 and 18:
    - `walg-replication` refuses a bad password, is idempotent, adds its line once,
      and leaves the password out of every `/proc/*/cmdline`;
    - the standby is seeded, is in recovery, streams, follows a write within 30s,
      archives nothing, and keeps no password on the data volume;
    - after a restart it is still a standby and follows again;
    - a replica with nothing to seed from idles without an `initdb`.

    150 checks pass on 17 and 18. `archiving catches up` (stopping S3) failed twice
    on 18 when that major ran alone, and passed when the full run was repeated, and
    passes on main: it is timing.
- **Backend.**
  - `ManagedPostgres.Members []{App, Key, Role, Volume, Node, At}` and
    `ReplicationPassword` (encrypted, made with the first replica).
  - `AppManagedSettings.MemberOf {App, Role}`.
  - **A replica has an app-managed setting of its own**, with `MemberOf` and its own
    `Locks`: the primary's image, command and env keys, its data mount, and the ids
    of its copies of the primary's secrets, scripts and mounts, plus the password
    secret and its mount. The existing lock rules therefore apply to it unchanged.
    Its `Postgres` is nil: its state is its primary's.
  - **`engine.Replica(req)`** returns the standby's env, the secret's name and file,
    the data mount target (from `Locks.Mounts`, the archive's volume excluded) and
    the allow command. It refuses:
    - a member (`ERR_APP_MANAGED_IS_MEMBER`);
    - a primary whose setting mounts include no `walg-replication`, meaning its
      scripts predate replicas (`ERR_APP_MANAGED_REPLICA_UNSUPPORTED`): its entrypoint
      would `initdb` a replica and archive it into the primary's prefix;
    - one still restoring;
    - an archive not off-site (`…_REPLICA_NEEDS_OBJECT_STORAGE`). **Shared volume
      archives are left out**, as the verifier's are: the replica would need the
      archive's volume mounted, which the copied settings do not hold.
  - **`managedService.AllowReplicas`** runs the allow command in the primary's task,
    with the password on stdin.
  - **`appmanageduc/replica.go`:**
    - `POST /managed/replicas {dataVolume}` checks the feature, plans the volume pin
      and the artifacts before the transaction, then locks the primary and checks
      again;
    - it makes or reads the password, lets replicas into the primary first, and
      provisions a logical child (key `<key>-r<n>`, the first free; the volume
      mounted with its own directory; the copies as the verifier's, keeping the
      server command, since a standby does not archive);
    - it records the member and audits;
    - `DELETE /managed/replicas/:memberID` removes the member and deletes the app last.
  - **Unmanage refuses** a member (`ERR_APP_MANAGED_IS_MEMBER`) and a primary with
    members (`ERR_APP_MANAGED_HAS_MEMBERS`). Deleting the primary deletes its
    replicas, which are logical children.
  - **Status `members`:** state is `pg_stat_replication`'s, `disconnected` when the
    list omits a member, and `unknown` without a health run or with an older
    `walg-health` that prints no list. Lag is given in bytes and in seconds; the
    seconds are null while nothing is written.
  - The primary's `max_wal_senders` and `hot_standby` are left at Postgres's defaults
    (10, on): nothing to set.
- **Dashboard.**
  - The Database tab's "Replicas" section shows each member with a link, its node,
    state, and lag ("caught up" at 0), and a Remove button behind a PopConfirm.
  - "Add replica" offers a select of pinned volumes, shown only for an off-site
    archive.
  - A replica's tab shows only a banner naming its primary.
- **Not done:**
  - no replication slot;
  - no shared-volume archive;
  - refreshing the primary's credentials does not refresh its replicas' copies,
    whose archive fallback keeps the old key;
  - a restored primary's replicas are not restored (as §9 says), and the restored
    data keeps the role with its old password until a replica is added.
- **Tests:**
  - engine: the spec, its four refusals, and the members' status;
  - service: the password on stdin, a failed allow;
  - usecase: the replica's settings (copies, variables, password secret and its mount
    at mode 0400, the member setting and its locks, no routing), its key, the
    password's shape, and the unmanage refusals;
  - e2e level-0: the replicas section, its add and remove requests, and the member's
    tab. All 10 pass.
- **Live check, on the e2e stack (`lc/check11.sh`), with the new template copied into
  dind:**
  - pg18, made before, is refused (412 `…_REPLICA_UNSUPPORTED`);
  - pg18b, made from the new template, had a base backup taken, then `pg18b-r1` added
    on a second volume;
  - the replica was seeded from the archive, is a standby, read 60 rows after 10 were
    written on the primary, streams, and archives nothing;
  - after a health run, Status showed it `streaming`, 0 bytes behind, and it named
    its primary;
  - changing its command was refused (`…_FIELD_LOCKED`), and neither database could be
    unmanaged;
  - removing it emptied the members, and its app is gone (404).

## Task 6: base backups from a replica

**Files:**
- **Engine:**
  - `ManagedPostgres.BackupFrom` holds `primary`, or a member's id.
  - The nightly job's `App` follows it. `backup-push` runs on a standby, which WAL-G
    supports.
- **The rest of the nightly sequence still runs on the primary:** retention, and
  pruning deploy points. Retention needs only the archive, and only the primary's
  credentials.
- **Settings:** a select.
- **Dashboard:** a select in Settings.

**Tests:**
- Reconfigure moves the job.
- Live: the nightly run on a replica makes a base backup that a restore uses.

- [x] Commit per repository.

**As built** (backend 72525471, merged as 662a50c5; dashboard f5de14e5, merged as
3b1157a1; CLI f2e289d, merged as d876935; app-templates harness 4704921, merged as
96657ff):
- **No script changed.** `walg-nightly backup` (`wal-g backup-push`) takes a base
  backup of a standby as it does of a primary, into the primary's archive.
  - The standby still archives nothing.
  - The backup ends in WAL that only the primary archives. A restore made in the
    archive_timeout after the backup waits for that WAL, as it does for any backup.
  - The harness checks this on 17 and 18: the backup is listed among the primary's,
    and after the primary archives past it, a restore to that backup with
    `recovery_target=immediate` holds the rows it was taken with. 166 checks pass.
- **`ManagedPostgres.BackupFrom`** is an `ObjectID`: empty for the primary, or a
  member's app.
  - `PUT /managed` takes `backupFrom`; `{"id": ""}` sends the backups home.
  - Anything that is not one of the database's members is refused with 400.
  - The settings answer `backupFrom`.
  - `postgres.moveBaseBackups` points the nightly job's `App` at the member or back
    at the primary. Its schedule stays, so the job is changed but not rescheduled.
- **`runNightly`** takes the backup on the app the job names. Retention and the
  pruning of deploy points run on the app the job belongs to, which is the primary:
  the job setting's object, from the scope's ref.
  - If the named replica is gone from the refs, the backup is taken on the primary,
    with a warning in the run's log.
- **Removing the member the backups are taken on** has the engine move them home first,
  through `moveBaseBackupsHome`, in the same transaction.
- **Dashboard.**
  - Settings has a "Base backups taken on" select ("This database" or each replica),
    shown only when the database has replicas.
  - The last base backup's card shows "Runs on <replica>".
- **Not done:** no automatic fallback to the primary when the replica is up but
  unhealthy. A failed backup on it fails the run and notifies, as any failed backup
  does.
- **Tests:**
  - engine: moving the job to a member, nothing to move, moving home, a stranger
    refused;
  - executor: the backup on the replica, retention and pruning on the primary, and the
    primary taking over from a replica that is gone;
  - usecase: the move through `reconfigureManaged` and the settings view;
  - e2e level-0: the select's payload and the "Runs on" line. All 11 pass.
- **Live check, on the e2e stack (`lc/check12.sh`):**
  - a stranger as backup host is refused (400);
  - with pg18b's backups moved to `pg18b-r1`, the nightly job's app is the replica,
    and its run is done;
  - WAL-G records `base_…0006` as taken on host `pg18b-r1`, and the replica archived
    nothing;
  - a restore of pg18b to that backup holds 70 rows, as at the backup, not the 5
    written after;
  - removing the replica set `backupFrom` back to null and the nightly job's app back
    to pg18b.

## Task 7: manual promotion and the stable address

**Decided here (§9):**
- **Promotion is fenced first.** It is one audited action, in order:
  1. scale the old primary's service to 0;
  2. `pg_promote()` on the chosen replica;
  3. make it the member that archives, and the nightly's target;
  4. repoint the other replicas' `primary_conninfo`, and restart them;
  5. swap roles in `Members`.
  A failure stops at the step it is in, with the database left fenced rather than with
  two primaries.
- **The stable address is a multi-host connection string,** shared as new system
  variables of the cluster's primary app:
  - `HIVEPAAS_DATABASE_URL_RW` uses `target_session_attrs=read-write`;
  - `HIVEPAAS_DATABASE_URL_RO` uses `target_session_attrs=prefer-standby`.
  The link dialog offers it for Postgres clusters. PgBouncer is left out of this task.
- **The archive.** The promoted member pushes to the cluster's prefix on its new
  timeline. `wal-verify timeline` in the health job catches two members pushing.

**Files:**
- **Backend:**
  - `appmanageduc/promote.go`, at `POST /managed/promote` with `{member}`;
  - the system vars in `envvarserviceimpl/app_system_env.go`;
  - the link recipe in `envlink/recipes.go`;
  - `make gen-swag`.
- **Dashboard:**
  - "Promote" on a member, with a confirmation naming what stops;
  - the URLs shown on the tab.

**Tests:**
- The step order, and a failure at each step.
- The system vars for 1 and 3 members.
- Live: promote a replica. The old primary is stopped, an app using the RW URL writes
  to the new primary, and the remaining replica follows it.

- [x] Commit per repository.

**As built** (app-templates 38dc260, merged as b0235e5; backend fad42d26, merged as
97b56b9e; dashboard cf1db0c5, merged as ce1999a0; CLI 9bc62b5, merged as 6a48fd5):
- **The cluster's app stays the cluster's.** Nothing moves between apps: its setting,
  its jobs and the addresses consumers name all stay on the app.
  - `ManagedPostgres.Writer` is the member promoted, and `FencedAt` is when the app was
    fenced. A member promoted over is given role `fenced`.
  - Everything that runs "on the database" runs on the writer: marks (person's and
    deployments'), letting replicas in, the nightly's retention (the executor looks up
    the writer, loading it if the run's refs lack it), and the jobs' `App`.
  - New replicas stream from the writer.
  - `engine.Reconfigure` with `Writer` moves the jobs: health, dump and restore test
    (container mode) to the writer, and the nightly to `BackupFrom` or the writer. In
    verifier mode the restore test's source host follows. Schedules do not move.
- **Template.**
  - `walg-promote` runs `pg_promote(true, 60)` and leaves a marker beside the data. It
    is idempotent.
  - Its exit codes: 0 when the server is the primary, 3 when it certainly is not (the
    server does not answer, or is still in recovery), 1 when that is not known.
  - In `walg-entrypoint`:
    - a standby with the marker starts as the primary whatever its standby variables
      say, so the promoted member is never restarted by the promotion;
    - `HIVEPAAS_FENCED` makes it idle, never starting the server, with the health
      check passing so Swarm does not restart it.
- **The steps, `runPromotion` over `promotionSteps`:**
  1. fence: `HIVEPAAS_FENCED` is written to the writer's variables and locked;
  2. stop: `SetAppRunning(false)`, then `ServiceWaitUntilStopped` for up to 90s;
  3. applyFence: the variables are applied to the stopped service;
  4. promote;
  5. allow: `walg-replication` on the new primary. **The plan did not have this.** The
     live run showed a replica's `pg_hba.conf` is its base backup's, which may predate
     the replicas' line;
  6. record (one transaction): writer, roles and jobs; the promoted member's standby
     variables are removed from its settings only; the other replicas'
     `HIVEPAAS_STANDBY_HOST` is set;
  7. restart: the other replicas' variables are applied.

  The outcome of each failure:
  - the fence not written: nothing happened;
  - not seen stopped, or certainly not promoted: the fence is lifted (variables removed
    and applied, app started), reporting "… runs again";
  - not known whether promoted, or promoted but not recorded: the fence stays, and
    promoting again finishes. A fence already written is skipped, and `walg-promote`
    answers 0 for a primary;
  - allow and restart only add warnings.

  `ServiceWaitUntilStopped` does not finish for a task on a node that does not answer.
  So a primary whose node is lost is never promoted over: that is §9's line.
- **The stable address.**
  - `HIVEPAAS_DATABASE_URL_RW/RO` are shared system variables of a managed postgres
    app that is no member's, and secret like the password. They are built by
    `clusterEnvVars`, from the kind's credentials.
  - Hosts: the writer first, then the replicas. The app is listed only while it is
    not fenced, and fenced members not at all. The port is the routing port, or 5432.
  - `sslmode`, `connect_timeout=5` (a stopped member's VIP would hold a client), and
    `target_session_attrs=read-write` or `prefer-standby`.
  - The link dialog recommends a "Cluster address" group, `DATABASE_URL` /
    `DATABASE_READ_URL`, for a Postgres target that shares them.
  - A consumer's resolved string is refreshed only when its variables are next
    applied, which a promotion does not do. It still reaches the new primary: it was
    one of the replicas listed, and libpq passes over the stopped one after its
    timeout.
- **Guards:**
  - a member is not deleted as an app (`ERR_APP_MANAGED_IS_MEMBER`); it is removed
    from the primary's tab;
  - the app a replica was promoted over is not deleted while it holds the cluster
    (`ERR_APP_MANAGED_CLUSTER_HELD`);
  - the writer member is not removed (`ERR_APP_MANAGED_IS_WRITER`);
  - promotion needs a primary with `walg-promote` (`ERR_APP_MANAGED_PROMOTE_UNSUPPORTED`)
    and a member with role `replica`.
- **Dashboard.**
  - "Promote" on a replica's row, with a PopConfirm saying what stops and that apps
    using the old host lose the database.
  - The primary member shows as `primary`, can be neither promoted nor removed, and
    has no lag.
  - The app promoted over shows a fenced banner naming the writer.
  - The Replicas section shows the two `${key.HIVEPAAS_DATABASE_URL_RW/RO}`
    references.
- **Not done:**
  - rejoining a fenced database as a replica (it would need re-seeding, or
    `pg_rewind`);
  - the health job's `wal-verify timeline` (the nightly's retain still runs it);
  - PgBouncer;
  - refreshing consumers' resolved addresses on a promotion;
  - a way to retire the fenced app other than keeping it: it holds the cluster.
- **Tests:**
  - engine: the spec, the refusals, first and second promotions (roles, the fence,
    the jobs, `BackupFrom`), replicas of a writer, and the members' states;
  - usecase: the step order, and each failure's steps and reason; `changedEnv`;
  - executor: retention on the writer;
  - env: the addresses with 1 and 3 members, before and after a promotion, and none
    for a member;
  - link: the cluster group first;
  - delete guards;
  - e2e level-0: 13, all pass;
  - harness: 198 checks on 17 and 18.
- **Live check, on the e2e stack (`lc/check13.sh`), with the template copied into
  dind:**
  - pg18b, from the previous template, is refused (`…_PROMOTE_UNSUPPORTED`);
  - pg18c got two replicas and an app writing through
    `PGURL=${pg18c.HIVEPAAS_DATABASE_URL_RW}`, which resolved to
    `pg18c:5432,pg18c-r1:5432,pg18c-r2:5432`;
  - promoting r1 answered writer r1, fenced pg18c, restarted [r2]. pg18c's service is
    at 0, runs nothing, and carries the fence;
  - r1 is out of recovery. The app's next write, through the unchanged address, landed
    on r1, and r2 has it and streams from r1;
  - the nightly, health and restore test jobs run on r1. The status shows r1 `primary`
    and r2 `streaming`. A nightly run on r1 took `base_00000002…` on timeline 2, and a
    restore point was marked on r1;
  - deleting pg18c is refused (`…_CLUSTER_HELD`). Started again, it logs its fence and
    runs no server.

## Task 8: unpinned volumes

**Decided here:**
- **Allowed when:**
  - the volume is reachable from every node, with `NodeID` and `NodeLabel` empty;
  - the archive is off-site;
  - the eligible nodes share one architecture. `sameBuildEverywhere` already refuses
    mixed architectures.
- **Artifacts are kept on every eligible node by a reconcile, not only at deploy.**
  - A cluster job runs every 10 minutes and on node join. For every managed app it calls
    `MountsFor` against the service's current spec.
  - `eligibleNodes` already evaluates constraints.
  - A task placed on a node that lacks the file fails, and Swarm retries elsewhere,
    until the reconcile has run.
- **The spec's "Which nodes" section is updated:** the helper it said was missing
  exists.

**Files:**
- `usecase/apptemplateuc/app_create_managed.go` (`dataVolumePin`) and
  `appmanageduc/verifier_provision.go`: accept an unpinned volume under these
  conditions.
- The reconcile job in `clustercleanupservice` or a new system job.
- Error messages.

**Tests:**
- Acceptance and refusal cases.
- The reconcile ensures files on nodes that lack them.
- Live: needs a second node; the dind stack has one. This task is checked on the Linux
  server or a two-node dind.

- [x] Commit per repository.

**As built** (backend f7e91542, merged as 3f6e21f6; dashboard aec27247, merged as
e4ea929d; no template change, and the API is unchanged, at level 4, so no CLI change):
- **Accepted:**
  - An unpinned data volume is accepted, as the volume's own claim of shared storage.
    `dataVolumePin` returns the pin the volumes share, or none.
  - With no pin, `volumeArchive` refuses a volume archive with
    `ERR_APP_MANAGED_SHARED_VOLUME_NEEDS_OBJECT_STORAGE`. An archive on a shared volume
    is refused too, as the plan said: only off-site.
  - `memberVolumePin` replaces `verifierVolumePin`, for the verifier and replicas, and
    accepts an unpinned volume. Those already require off-site. A replica on one
    records node `""`, and the dashboard shows "On any node".
  - `placementservice.PinnedSpec(pin)` is the spec `MountsFor` is given: no placement
    for no pin, so every eligible node.
- **Added beyond the plan:**
  - **tmpfs is refused** (`volumeservice.DriverOptsInMemory`, driver option
    `type=tmpfs`) with `ERR_APP_MANAGED_VOLUME_IN_MEMORY`. This applies to the data
    volume, pinned or not, and a replica's. A verifier's scratch may be in memory.
  - **One task at most.** `checkManagedMode` in the service settings refuses another
    mode than replicated, or more than 1 replica, as field `mode` / `replicas`. Two
    tasks would be two servers on one data directory, which shared storage lets start
    on two nodes.
  - `ERR_APP_MANAGED_VOLUME_NOT_PINNED` now means no volume, or volumes pinned to
    different nodes. Its message changed and it takes no parameter.
- **The reconcile:**
  - `artifactService.ReconcileAll` lists the active app-artifacts settings (the
    database's, its replicas', its verifier's) and loads their apps. For each it
    inspects the service and calls `MountsFor` with its spec as it stands.
  - A service that is gone is skipped. A failure (two architectures, no Ready node, a
    download) is recorded per app, and the others go on. It changes no service.
  - The global job `managed-artifacts` (`base.SchedJobTypeManagedArtifacts`, a
    variable as the others are) is made on every start by `initDefaultManagedArtifacts`
    if missing. It runs every 10 minutes from that start, with no target, no retry and
    no notification.
  - Its run logs "The artifacts of N app(s) are on every node they may start on", and
    one line per failed app with its reason. It fails when any app did.
  - `JoinNode` (over SSH) then creates a task of that job, run two minutes later
    (`RunAt`), for the node to be Ready and its agent up. A failure to schedule it
    does not fail the join. A node joined by the join command alone waits for the next
    turn.
- **Not done:**
  - **A volume that claims to be shared but is not.** It gives an empty directory on
    another node, where the database initializes afresh.
    - `WALG_PREVENT_WAL_OVERWRITE` makes its archiving fail on the segments the archive
      has. Its nightly base backup would still join the archive's.
    - Refusing an empty data directory over an archive with backups, in
      `walg-entrypoint`, is a follow-up. The spec says so.
  - **No two-node live check.** That is for the Linux server.
- **Dashboard:**
  - The verifier, replica and restore volume pickers list every volume with where it
    is: `node …`, `nodes labeled …` or `every node`.
  - Replicas and restore leave out or disable tmpfs ("kept in memory").
  - Restore disables a shared volume unless the source archives off-site.
  - The create form's note says a shared data volume archives to object storage.
- **Tests:**
  - create plan: shared accepted and unpinned; refused with a volume archive, and for
    tmpfs pinned or not;
  - `memberVolumePin`;
  - `PinnedSpec`;
  - the mode lock;
  - `ReconcileAll` (pinned, shared, mixed architectures fails alone, service gone,
    none);
  - the job's log;
  - the job's init, once;
  - the join's scheduled run;
  - e2e level-0: 13, all pass. The restore test now expects the shared volume enabled
    and a tmpfs one disabled.
- **Live check, on the e2e stack (`lc/check14.sh`):**
  - At start: one `managed-artifacts` job, every 10m.
  - Refusals:
    - pg18s on an unpinned bind volume with the dumps' volume repository is refused
      (`…_SHARED_VOLUME_NEEDS_OBJECT_STORAGE`);
    - a tmpfs data volume is refused (`…_VOLUME_IN_MEMORY`).
  - pg18s with S3:
    - answers; its service has no node constraint and binds the WAL-G artifact;
    - its nightly is done.
  - A replica on another unpinned volume has node `""`, and streams.
  - Service settings:
    - 2 replicas is refused (`…_FIELD_LOCKED`, replicas);
    - 1 is saved.
  - Reconcile: the cached artifact file was removed from the node. The job's next run
    (7.9s) was done, logged 10 apps, and the file was back.

## Task 9: scanning the artifact

**Files:**
- **`imagescanservice/imagescanrun`:**
  - a `rootfs` mode, `trivy rootfs --scanners vuln --format json <file>`, with the
    cached file bind-mounted read-only;
  - `RunReq` gains the mode and the host path;
  - the agent's proto and server copy them;
  - `validate` accepts an artifact's `sha256:<hex>`.
- **`imagescanserviceimpl`:**
  - an app's managed artifacts are scanned with its images;
  - results are keyed by the artifact's sha256 in `image_scans`;
  - `app_image_scans.image_ids` includes them.
- **Dashboard:** the Security tab names an artifact as such.

**Tests:**
- The script for the mode.
- Report parsing for a Go binary.
- The app's image list includes the artifact.
- Live: the WAL-G binary is scanned, and its findings are shown.

- [x] Commit per repository.

**As built** (backend 67c29677, merged as dfcb47e6; dashboard 6489604b, merged as
f8639fdf; CLI 09157b5, merged as dd7fa4c; no template change):
- **The runner (`imagescanrun`):**
  - `RunReq.Mode` is `""` for an image and `ModeRootfs` for an artifact's file. The
    script then runs `trivy rootfs … /scan/artifact`, and an SBOM is
    `trivy rootfs --format cyclonedx`.
  - The source is `base.ImageScanSourceArtifact` (`"artifact"`). The node's Docker is
    not asked for the image, no socket is mounted, and no credentials are copied.
  - **Changed from the plan: `RunReq` takes no host path.** The runner binds
    `CacheHostDir()/<hex of ImageID>` read-only. The agent then never mounts a path a
    call names, only a file of its cache by a validated `sha256:<64 hex>`, which
    `imageIDPattern` already checks.
  - `validate` refuses another mode, a `RemoteRef` in rootfs mode, and a node with no
    host directory.
  - The proto gains `mode = 15`. It was regenerated with protoc 35.1, and the agent's
    client and server copy it.
- **The service (`imagescanserviceimpl`):**
  - `scanApp` adds `appArtifacts` to `appImages`. For each node the app's tasks run
    on, the node's architecture comes from `NodeInspect`, and each artifact gives the
    file for that build: once per sha256, ID `sha256:<hex>`, Ref `<name> <version>`,
    Platform `linux/<arch>`. Failing to read the artifacts is a warning, and the
    images are scanned all the same.
  - Before the scan, `ensureArtifact` has the file on the scan node.
    `EnsureOnNodeReq.Arch` is new: it picks the task's build whatever that node's
    architecture. A file of another sha256 is refused.
  - `runReqOf` builds an image's or a file's request, shared with the SBOM. The SBOM
    finds artifacts as the scan does.
  - The scan is kept as an image's: `image_scans` by the sha256, with source
    `artifact` and the build's platform (a failed first scan too).
    `app_image_scans.image_ids` lists it. The hour's freshness, ignores, counts,
    notifications and the daily scan treat it as an image.
- **The API:** `ImageScanSource` gains `"artifact"`, in responses only, so the level
  stays 4. The CLI client was regenerated for the swagger change.
- **Dashboard:**
  - `ScannedImages` shows an artifact with a file icon, a "Managed artifact" badge and
    "read from the node's artifacts cache".
  - Its SBOM button says "the artifact's modules".
- **Found on WAL-G 3.0.9:**
  - 35 findings in its Go modules, 15 of them HIGH.
  - One is a false positive: CVE-2021-38599 (fixed in 1.1) on WAL-G's own module,
    because the binary carries the pseudo-version `v0.0.0-…+dirty`. An ignore sets it
    aside.
- **Tests:**
  - the script and the request's validation in rootfs mode;
  - the run binds the file and nothing of Docker;
  - parsing a Go binary's report (`testdata/walg-gobinary.json`, trimmed from WAL-G's
    real one);
  - `artifactTargets`, `appArtifacts`, `runReqOf`;
  - `readScan`'s platform;
  - `ensureArtifact` (the build asked for; another sha256 refused);
  - `EnsureOnNode` with `Arch`;
  - an app's scan lists the artifact among its image IDs;
  - e2e level-0: 14, all pass. The new test is the Security tab naming the artifact.
- **Live check, on the e2e stack (`lc/check15.sh`):**
  - Scanning was switched on with a pinned volume, then off again after the check.
  - pg18's scan task was done. It scanned `postgres:18.6-trixie`, then
    `wal-g 3.0.9 (965a71478524)`: 0 CRITICAL, 15 HIGH, 34 with a fix.
  - The image list has it as `source: artifact`, `linux/amd64`, `done`, with 35
    `gobinary` findings.
  - Its SBOM is CycloneDX with 126 components, `golang.org/x/net` among them.
  - `image_scans` keeps it by its sha256, and the app's `image_ids` lists it.
  - The real Security tab shows the badge (screenshot checked).

## After the plan

- Release the CLI for the API level this phase ends at, as phase 1 left it: a hivepaas
  tag, then `make update-spec REF=<tag>`.
- Live checks for each task are in the task. Task 8 needs two nodes.

**Done after the plan** (2026-10-19):
- **The archive prefix carries the app's ID** (backend 2bf6b640, merged as 9a692ca0).
  - Found while writing the guard below: an app key is unique within its environment
    only, and HivePaaS deletes no archive. So `walg/<app key>` was shared:
    - by two environments' or projects' databases of one name on one storage;
    - by a database made again by the name of a deleted one.
  - A new database archives under `walg/<app key>/<app ID>`, as the user chose.
    Databases made before keep the prefix they recorded.
- **`walg-entrypoint` never starts afresh over an archive with history**
  (app-templates 2440068, merged as 1f6be93; spec ac5196f9).
  - An empty data directory is initialized only when `wal-g st ls` lists no
    `basebackups_005/` or `wal_005/`.
  - Over history it idles. It writes why to `hivepaas-refused` beside the data:
    `health` passes, so Swarm leaves it, and `walg-health` fails, saying why.
  - An archive it cannot read within `HIVEPAAS_ARCHIVE_CHECK_TIMEOUT` (60s) fails the
    start. WAL-G retries an unreachable S3 for minutes, so the check is bounded.
  - Harness: 12 checks a major (222 in all, 17 and 18): history refused, the archive
    untouched, unreadable fails, empty starts.
  - Live (`lc/check16.sh`): pg18g archived under `walg/pg18g/<ID>` and took a base
    backup. Its `PGDATA` was removed and its service forced. It then idled with the
    refusal, its container healthy, its health job failing with why, and its backup
    kept.
- **`release.json` pins app-templates at 1f6be93** (6aa3e30b, merged as 647a6060). Its
  `indexSha256` was checked against `index.json` there. It needs app-templates main
  pushed, then `release.signed.json` signed again.
