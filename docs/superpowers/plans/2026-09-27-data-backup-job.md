# Data Backup Job Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A scheduled job type, `data-backup`, takes a snapshot of an app's data - a command's output, such as `pg_dump`, or the app's own directory of a volume it mounts - into a backup repository.

**Architecture:** `SchedJob.DataBackup` says the source and the repository. The job runner hands a `data-backup` job to a new `databackupservice`. For a command, it runs the command in the app through the scheduled job exec service, with the command's stdout written into a pipe. `kopia snapshot create --stdin-file` reads that pipe in the backend. For a volume, it resolves the app's directory on the volume's host and runs `kopia snapshot create <dir>` through the agent on the volume's node. `backupreposervice` builds the engine (storage, password, where its commands run) and gains `BackupStream`, `BackupDirectory` and `DeleteSnapshot`. After a snapshot, the repository's list is synced, and the snapshot is recorded in the task's output.

**Tech Stack:** Go (bun, fx), kopia through `services/backup`; React 19, react-hook-form and zod in `hivepaas-dashboard`.

**Spec:** `docs/superpowers/specs/2026-09-27-data-backup-job-design.md`

## How the patches work

Every task's code was written and verified before this plan.

**Bases.** The backend patches apply in order to `main` at `6d2cb06b`; this plan's commit changes only `docs/`. The dashboard patches apply to the dashboard's `main` at `52eb2915`.

**What was checked on fresh worktrees:**

- Backend, at every task:
  - the tests alone fail to build;
  - with the code, the task's packages and the fx wiring test pass, and `golangci-lint run ./...` finds 0 issues.
- Backend, after the last task: `go test ./...` passes, `kopia` integration tests included (kopia 0.23.1 is installed).
- Dashboard, after both patches: `tsc`, `npm run lint` and `prettier --check src` pass.
- The JSON the backend sends was parsed with the dashboard's validators: a data backup job of each source, one with a deleted repository, and a run's snapshot.

Patches live in `docs/superpowers/plans/2026-09-27-data-backup-job/`:

- `be-NN-tests.patch`, then `be-NN-code.patch`, for backend task NN: the tests first, watched failing, then the code.
- `dash-NN.patch` for dashboard task NN. Its gate is the type check, the lint and, in Task 8, the browser.

The commands below run from the backend repo root, with `P=docs/superpowers/plans/2026-09-27-data-backup-job` and `D=../hivepaas-dashboard`. If a base has moved since, use `git apply --3way` and resolve.

**Your uncommitted work.** Task 1 includes your uncommitted changes to `hivepaas_app/base/sched_job.go` and `hivepaas_app/entity/setting_sched_job.go`. It also replaces your `hivepaas_app/entity/setting_sched_job_backup.go` with `setting_sched_job_data_backup.go`, which has the same `SchedJobDataBackup` and more fields. Drop your copies of those three files before the branch is merged into your working tree: git refuses a merge over local changes to the files it touches.

## Global Constraints

- **Backend, before a task is done:**
  - `go build ./...`;
  - `golangci-lint run ./...` over the **whole** repo (120-character lines, US spelling);
  - `go test ./...`.
  - `go test ./hivepaas_app/cmd/...` includes the fx wiring test, which catches a dependency cycle.
- **`make gen-swag` when a DTO changes** (Tasks 4 and 5). `docs/openapi/swagger.json` is generated and committed.
  - A first run that fails with `cannot find all dependencies` is run again.
  - Then `go mod verify` must say `all modules verified`: an interrupted run once left modules half-extracted in `~/go/pkg/mod`.
- **Dashboard, before a task is done:** `npx tsc --noEmit`, `npm run lint` (`--max-warnings 0`), `npx prettier --check src`.
- **Names, verbatim:**
  - job type `data-backup`;
  - `SchedJob.DataBackup` as `dataBackup`, with `source` (`command` | `volume`), `sourceCommand`, `sourceFileName`, `sourceVolume`, `sourceVolumeSubpath`, `targetRepository` and `tags`;
  - the result `{snapshotId, sizeBytes}`.
- **Snapshot tags:**
  - `hivepaas.job:<job id>`, `hivepaas.app:<app id>`, `hivepaas.run:<task id>`, then the job's own, sorted;
  - a user's tag key may not start with `hivepaas.`.
- **Snapshot source and description:**
  - source `hivepaas@data-backup:/<job id>` for every snapshot of a job;
  - description `<job name> (run <task id>)`.
- **What is not refused when a job is saved:** a command into a repository on a volume, and a volume on another node than the repository's volume. The repository server of the next spec makes them work, and their runs fail, saying why, until then. **Do not add a check for them.**
- **Scope:** a data backup lives at the app scope only, and runs in its own app.
- **No backward compatibility is needed.**
- **Git:**
  - commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`;
  - merge locally, never push;
  - never `git stash`, and never touch the user's uncommitted work.
- **Deploys:** nothing is deployed to the user's Docker Desktop swarm. The live checks of Task 8 run on the user's Linux server.

## Review Focus

1. **A command that prints gigabytes.**
   - Its output passes through `io.Pipe`, from the agent's exec stream to kopia's stdin, and nothing on the way may buffer it whole.
   - A reasonable person expects the backend's memory to stay flat during a large `pg_dump`.
   - Pinned by: Task 8, check 1, with a database of a few GB, watching the backend's memory.
2. **A run stopped half way.**
   - When the job's timeout passes, or the run is canceled, during a stream:
     - the command is stopped;
     - kopia is killed through the canceled context;
     - `deleteRunSnapshots` runs on a context without the cancel.
   - A reasonable person expects no partial snapshot to be left behind.
   - Pinned by: Task 8, check 4, with a timeout shorter than the dump.
3. **A volume on another node than the backend, into a cloud repository.**
   - kopia runs through that node's agent and connects there with the repository's own config file.
   - A second run must find the connection in place, or make it again.
   - Pinned by: Task 8, check 2, on a node other than the manager.
4. **A command that exits 0 and prints nothing** (a wrong database name, or a script that swallows its errors).
   - The run succeeds with a snapshot of size 0.
   - The run's page shows `0 B` next to the snapshot, which is how a person would notice.
   - Pinned by: Task 8, check 7.
5. **Two runs of one job at once** (run-now while a scheduled run goes on).
   - Each run tags its snapshots with its own task ID, and a failed run deletes only its own.
   - Pinned by: `TestBackupOfAFailedCommandDeletesTheRunsSnapshots` (Task 3), which lists by `hivepaas.run:t1`, and by Task 8, check 5.

## Decisions beyond the spec

These were taken while building the patches, and written back into the spec.

- **One kopia source a job: `hivepaas@data-backup:/<job id>`.**
  - kopia's retention goes by source.
  - The hostnames of the backend and an agent change when they are redeployed, which would leave each old source keeping its own latest N forever.
  - The repository's retention now keeps its count per job.
- **The engine had three gaps, fixed in Task 2:**
  - a stream needs a source path, and `kopia` answered `no snapshot sources` without one;
  - `snapshot create --json` gives no `stats`, so a new snapshot's size is taken from its root entry;
  - `BackupOptions` gains `Description` and `Source` (`--override-source`), and `backupmodel.Snapshot` gains `Description`.
- **The command starts once kopia is connected** (`BackupStreamReq.OnConnected`).
  - Building the engine reads the database through the run's transaction, which the command's environment reads too.
  - A pgx transaction is not safe to use from two goroutines.
- **kopia stopping first fails the run with kopia's error.**
  - Otherwise the run would fail with the command's broken pipe.
  - A stream that ends before kopia is connected runs no command at all.
- **A failed command's snapshot:**
  - the snapshot kopia named is deleted;
  - when kopia named none, the snapshots tagged `hivepaas.run:<task id>` are listed and deleted.
- **The job exec service runs a given command into a given writer** (`SchedJobExecReq.Command`, `StdoutWriter`), with no TTY. A data backup uses the command runner of a container job this way.
- **A volume source is the app's own directory of the volume.**
  - It is found in the app's swarm service mounts through `DescribeAppMounts`, as the storage screen finds it.
  - A volume mounted whole carries no volume ID there, and is refused at save.
  - A volume pinned to no node fails the run.
- **The storage settings API gives each own-directory mount its `volumeId`,** which the form's volume picker lists. There is no new endpoint.
- **A step of a job sequence:**
  - leaves the task's output alone, because that output is the sequence's run;
  - hands on `SNAPSHOT_ID` and `SNAPSHOT_SIZE_BYTES` as step outputs.
- **The task API gives `sequenceRun` only for a run that started.** A data backup's output reads as an empty sequence run.
- **A sync failure after a snapshot is logged, not fatal.** The snapshot is taken, and the repository's next sync finds it.
- **`TransformSchedJob`'s command output part moved into `transformCommandOutput`,** to keep the function under the linter's complexity limit.
- **The dashboard's command section still shows a TTY checkbox.** The payload always sends `tty: false`, and the server forces it too.

---

## Task 1: A data backup on a scheduled job

**Files:**
- Modify: `hivepaas_app/base/sched_job.go`, `hivepaas_app/entity/setting_sched_job.go`, `hivepaas_app/tasks/taskschedjobexec/job_run.go` (an interim case, replaced in Task 3)
- Create: `hivepaas_app/entity/setting_sched_job_data_backup.go`
- Test: `hivepaas_app/entity/setting_sched_job_data_backup_test.go`

**Interfaces:**
- Produces:
  - `base.SchedJobTypeDataBackup`, which is in `AllSchedJobTypes`;
  - `base.SchedJobDataBackupSource`, with `SchedJobDataBackupSourceCommand` and `SchedJobDataBackupSourceVolume`, and `base.AllSchedJobDataBackupSources`;
  - `entity.SchedJob.DataBackup *SchedJobDataBackup`;
  - `entity.SchedJobDataBackup{Source, SourceCommand *CommandTemplate, SourceFileName, SourceVolume ObjectID, SourceVolumeSubpath, TargetRepository ObjectID, Tags map[string]string}`, whose volume and repository are in `GetRefObjectIDs().RefSettingIDs`;
  - `(*SchedJobDataBackup).SnapshotTags(jobID, appID string) []string`;
  - `entity.DataBackupTagJob = "hivepaas.job"` and `entity.DataBackupTagApp = "hivepaas.app"`;
  - `entity.SchedJobDataBackupResult{SnapshotID, SizeBytes}`, and `(*Task).OutputAsDataBackup() (*SchedJobDataBackupResult, error)`, which returns nil for a task whose output has no snapshot ID.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-01-tests.patch`

Tests added: `TestSchedJobDataBackupReferences`, `TestSchedJobDataBackupOfACommandReferencesTheRepositoryOnly`, `TestSchedJobDataBackupSnapshotTags`, `TestTaskOutputAsDataBackup`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/entity`
Expected: FAIL, build error `undefined: base.SchedJobTypeDataBackup`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-01-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test ./hivepaas_app/entity ./hivepaas_app/tasks/... && golangci-lint run ./...`
Expected: PASS; `0 issues.` The interim `data-backup` case in `runJob` answers `ERR_UNSUPPORTED` until Task 3. Without it, the linter's exhaustive switch check fails.

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(sched-jobs): a data backup on a scheduled job

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 2: The backup repository takes a stream, a directory, and deletes a snapshot

**Files:**
- Modify:
  - `services/backup/backupmodel/types.go` (`BackupOptions.Description`, `.Source`; `Snapshot.Description`);
  - `services/backup/kopia/backup.go` (a stream's source path, `--description`, `--override-source`, the root entry's size);
  - `services/backup/kopia/get.go`;
  - `hivepaas_app/service/backupreposervice/service.go`, `…/types.go`.
- Create: `hivepaas_app/service/backupreposervice/backupreposerviceimpl/backup.go`
- Test:
  - `…/backupreposerviceimpl/backup_test.go`;
  - `services/backup/kopia/backup_stream_integration_test.go`, which runs the real `kopia`;
  - `services/backup/kopia/get_test.go`.

**Interfaces:**
- Consumes: nothing from Task 1.
- Produces, on `backupreposervice.Service`:
  - `BackupStream(ctx, db database.IDB, *BackupStreamReq) (*BackupResp, error)`;
  - `BackupDirectory(ctx, db, *BackupDirectoryReq) (*BackupResp, error)`;
  - `DeleteSnapshot(ctx, db, *DeleteSnapshotReq) error`;
  - `VolumeHostDir(ctx, volume *entity.Setting) (*VolumeHostDir, error)`.
- Produces, in `backupreposervice`:
  - `RepoTarget{Scope, RepoSetting, RefObjects}`;
  - `BackupStreamReq{RepoTarget; Stdin io.Reader; FileName; Tags []string}`;
  - `BackupDirectoryReq{RepoTarget; HostDir, NodeID, NodeLabel string; Tags}`;
  - `BackupResp{Snapshot *RepoSnapshot}`;
  - `DeleteSnapshotReq{RepoTarget; SnapshotID}`;
  - `VolumeHostDir{Dir, NodeID, NodeLabel}`.
- Produces, in `backup.BackupOptions`: `Description` and `Source`.
- Task 3 adds `Source`, `Description` and `OnConnected` to `BackupStreamReq`, and `Source` and `Description` to `BackupDirectoryReq`.
- `BackupDirectory` runs kopia through the agent on the node it is given. The directory is taken under `volumeservice.HostPathPrefix`. A repository on a volume of another node fails with `ERR_NOT_IMPLEMENTED`, saying why.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-02-tests.patch`

Tests added: `TestOnNodeExecutorRunsWhereAsked`, `TestCheckRepoReachableFrom`, `TestToRepoSnapshot`, `TestIntegration_BackupStream_TakesTheStreamAsAFile`, `TestIntegration_BackupDirectory_UnderAGivenSource`, `TestToStandardSnapshot_SizeFromTheRootEntry`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./services/backup/... ./hivepaas_app/service/backupreposervice/...`
Expected: FAIL, build errors `undefined: onNodeExecutor` and `unknown field Source in struct literal of type backupmodel.BackupOptions`.

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-02-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test ./services/backup/... ./hivepaas_app/service/backupreposervice/... ./hivepaas_app/cmd/... && golangci-lint run ./...`
Expected: PASS; `0 issues.` The integration tests skip without a `kopia` binary. Run them where one is installed: they are the proof of the stream fix.

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(backup): a repository takes a stream or a directory, and deletes a snapshot

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 3: The data backup runner

**Files:**
- Create:
  - `hivepaas_app/service/databackupservice/service.go`;
  - `…/databackupserviceimpl/service.go`, `backup.go`, `volume.go`;
  - `hivepaas_app/tasks/taskschedjobexec/data_backup.go`.
- Modify:
  - `hivepaas_app/service/schedjobexecservice/types.go` (`SchedJobExecReq.Command`, `.StdoutWriter`);
  - `…/schedjobexecserviceimpl/exec.go`, `exec_cmd.go`, `exec_env.go`;
  - `hivepaas_app/service/backupreposervice/types.go`, `…/backupreposerviceimpl/backup.go` (`OnConnected`, `Source`, `Description`);
  - `hivepaas_app/tasks/taskschedjobexec/job_run.go`, `executor.go`;
  - `hivepaas_app/registry/provides.go`.
- Test:
  - `…/databackupserviceimpl/backup_test.go`;
  - `…/schedjobexecserviceimpl/exec_writer_test.go`;
  - `hivepaas_app/tasks/taskschedjobexec/data_backup_test.go`.

**Interfaces:**
- Consumes: Task 1's entity, and Task 2's `backupreposervice` methods.
- Produces, on `databackupservice.Service`:
  - `Backup(ctx, db database.Tx, *BackupReq) (*BackupResp, error)`;
  - `FindAppVolume(ctx, db database.IDB, app *entity.App, volumeID string) (*AppVolume, error)`.
  - Task 4 adds `CheckAppVolume` to it.
- Produces, in `databackupservice`:
  - `BackupReq{*queue.TaskExecData; JobSetting *entity.Setting; App *entity.App; RefObjects *entity.RefObjects; Sequence *schedjobexecservice.SequenceStep}`;
  - `BackupResp{Result *entity.SchedJobDataBackupResult}`;
  - `AppVolume{HostDir, NodeID, NodeLabel}`.
- Produces, in `schedjobexecservice.SchedJobExecReq`: `Command *entity.CommandTemplate` runs in place of the job's own command, and `StdoutWriter io.Writer` receives its stdout, with no TTY.
- Produces, in `backupreposervice.BackupStreamReq`: `Source`, `Description` and `OnConnected func()`. `BackupStream` calls `OnConnected` once the engine is connected, and from then on reads no more from the database.
- `runJob` runs `data-backup` through `dataBackupService.Backup`. As a sequence step, it gives the outputs `SNAPSHOT_ID` and `SNAPSHOT_SIZE_BYTES`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-03-tests.patch`

Tests added:
- the command source:
  - `TestBackupOfACommand` (streamed, synced, the task's output, the source and the description);
  - `TestBackupOfAFailedCommandLeavesNoSnapshot`;
  - `TestBackupOfAFailedCommandDeletesTheRunsSnapshots`;
  - `TestBackupOfACommandIntoAnUnreachableRepository` (the command never runs);
  - `TestBackupOfACommandWhenKopiaStops` (kopia's error, not the broken pipe);
- the volume source: `TestBackupOfAVolume`, `TestAppVolumeDir`, `TestJoinSubpath`, `TestPickAppVolumeMount`;
- the run as a whole:
  - `TestBackupAsASequenceStepLeavesTheTaskOutputAlone`;
  - `TestBackupFailsWithoutAnActiveRepository`;
  - `TestBackupFailsWithoutItsApp`;
- the exec service and the runner: `TestSchedJobExecRunsAGivenCommandIntoAGivenWriter`, `TestDataBackupStepOutputs`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/service/databackupservice/... ./hivepaas_app/service/schedjobexecservice/... ./hivepaas_app/tasks/taskschedjobexec/`
Expected: FAIL, build errors; the first is that no module provides the package `…/service/databackupservice`.

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-03-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -race ./hivepaas_app/service/databackupservice/... && go test ./hivepaas_app/service/schedjobexecservice/... ./hivepaas_app/service/backupreposervice/... ./hivepaas_app/tasks/taskschedjobexec/ ./hivepaas_app/cmd/... && golangci-lint run ./...`
Expected: PASS, the runner's tests under `-race` too; `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(sched-jobs): the data backup runner - a command's output or an app's volume into a repository

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 4: A data backup in the job API

**Files:**
- Create:
  - `hivepaas_app/usecase/settings/schedjobuc/schedjobdto/data_backup.go`;
  - `hivepaas_app/usecase/settings/schedjobuc/data_backup.go`.
- Modify:
  - `…/schedjobdto/create.go` (`SchedJobBaseReq.DataBackup`);
  - `…/schedjobdto/get.go` (`SchedJobResp.DataBackup`, `transformCommandOutput`);
  - `…/schedjobuc/sequence.go` (`checkJobTypeInScope`), `create.go`, `update.go`, `uc.go`;
  - `hivepaas_app/service/databackupservice/service.go`, `…/databackupserviceimpl/volume.go` (`CheckAppVolume`);
  - `hivepaas_app/usecase/appsettingsuc/appsettingsdto/storage_settings_get.go` (`Mount.VolumeID`);
  - `hivepaas_app/interface/mcp/tools_sched_write.go`;
  - `docs/openapi/swagger.json`.
- Test:
  - `…/schedjobdto/data_backup_test.go`;
  - `…/schedjobuc/data_backup_test.go`;
  - `…/appsettingsdto/storage_volume_id_test.go`.

**Interfaces:**
- Consumes: Task 3's `databackupservice.Service`, which gains `CheckAppVolume(ctx, db, app *entity.App, volumeID string) error`.
- Produces the request, `dataBackup: {source, sourceCommand, sourceFileName, sourceVolume {id}, sourceVolumeSubpath, targetRepository {id}, tags}`. Its validation:
  - a file name matching `^[A-Za-z0-9._-]{1,100}$`;
  - a relative subpath, with no leading `/` and never above its start;
  - at most 20 tags, each with a key matching `^[A-Za-z0-9._-]{1,50}$` that does not start with `hivepaas.`, and a value matching `^[A-Za-z0-9._:/@+=-]{1,100}$`;
  - a data-backup job has an `app` and no `command` or `commandOutput`, and no other type has a `dataBackup`.
- Produces the response, `dataBackup` with `sourceVolume` and `targetRepository` as named settings (`"missing"` once deleted), and `sourceCommand`.
- Produces `volumeId` on each storage mount of the app's own directory.
- The use case checks:
  - `data-backup` is at the app scope only;
  - the job's app is the scope's app;
  - a volume source is mounted by the app as its own directory.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-04-tests.patch`

Tests added:
- the request:
  - `TestADataBackupOfACommand`, `TestADataBackupOfACommandNeedsTheCommandAndAFileName`;
  - `TestADataBackupOfAVolume`, `TestADataBackupSubpathStaysInsideTheVolume`;
  - `TestADataBackupNeedsARepositoryAndASource`, `TestADataBackupsTags`;
  - `TestADataBackupRunsNoCommandOfItsOwn`;
- the use case: `TestDataBackupsLiveInApps`, `TestADataBackupIsItsAppsOwn`;
- the responses: `TestTransformSchedJobNamesADataBackupsVolumeAndRepository`, `TestAStorageMountNamesTheVolumeOfTheAppsOwnDirectory`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/usecase/settings/schedjobuc/... ./hivepaas_app/usecase/appsettingsuc/...`
Expected: FAIL, build errors such as `undefined: SchedJobDataBackupReq` and `undefined: checkDataBackupApp`.

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-04-code.patch`

The patch carries the regenerated `swagger.json`. Run `make gen-swag` anyway: it must leave `docs/openapi/swagger.json` unchanged. Then run `go mod verify`.

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test ./hivepaas_app/usecase/... ./hivepaas_app/interface/... ./hivepaas_app/cmd/... && golangci-lint run ./...`
Expected: PASS; `0 issues.` `TestEveryToolBuildsAndDescribesItsArguments` (MCP) passes with `dataBackup` described.

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(sched-jobs): a data backup in the job API, its app's own

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 5: A run's snapshot in the task API

**Files:**
- Modify: `hivepaas_app/usecase/taskuc/taskdto/get.go`, `docs/openapi/swagger.json`
- Test: `hivepaas_app/usecase/taskuc/taskdto/data_backup_test.go`

**Interfaces:**
- Consumes: `(*entity.Task).OutputAsDataBackup()` (Task 1).
- Produces:
  - `TaskResp.DataBackup *entity.SchedJobDataBackupResult` (`dataBackup {snapshotId, sizeBytes}`);
  - `sequenceRun` only for a run that started.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-05-tests.patch`

Tests added: `TestTransformTaskGivesADataBackupsSnapshot`, `TestTransformTaskGivesNoSnapshotForASequencesRun`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/usecase/taskuc/...`
Expected: FAIL, build error `resp.DataBackup undefined (type *TaskResp has no field or method DataBackup)`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-05-code.patch`. Then run `make gen-swag`, which must leave `swagger.json` unchanged, and `go mod verify`.

- [ ] **Step 4: Run the whole suite**

Run: `go build ./... && go test ./... && golangci-lint run ./...`
Expected: PASS; `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(tasks): a data backup's run gives its snapshot

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 6: The data backup form, its route and its list tag (dashboard)

**Files** (under `$D/src/application/`):
- Create:
  - `modules/projects/module-shared/enums/e.sched-job-data-backup.ts`;
  - `modules/projects/module-shared/components/data-backup-form/` (`data-backup.form.schema.ts`, `data-backup.form-mappers.ts`, `data-backup.form.com.tsx`, `index.ts`);
  - `modules/projects/routes/single-project/single-app/configuration/scheduled-jobs/data-backup/` (`app-data-backup-create.route.com.tsx`, `app-data-backup-form-route.com.tsx`, `index.ts`).
- Modify:
  - the job type enum (`DataBackup: "data-backup"`);
  - the job entity, validator and contracts (`dataBackup`);
  - the storage settings entity and validator (`volumeId`);
  - the list route (`+ New Data Backup`);
  - the edit route (the form by `jobType`);
  - the list's name cell (`Backup · <repository>`, or `deleted repository`);
  - the env jobs filter (the type);
  - `shared/constants/route.constants.ts` (`createDataBackup`);
  - `projects.router.tsx`, `projects.module.ts`, `routes/index.ts`, and the components and enums indexes.

**Interfaces:**
- Consumes: the job API of Task 4, and `volumeId` on storage mounts.
- Produces:
  - `DataBackupForm`;
  - `mapDataBackupFormToPayload(values, appId)`, which always sends `tty: false`;
  - `mapDataBackupToFormInput(job)`;
  - `AppDataBackupFormRoute`, `AppDataBackupCreateRoute`;
  - `ROUTE.projects.single.apps.single.configuration.scheduledJobs.createDataBackup`.

The form:
- name;
- **Source** tabs:
  - **Command**: the command section of the job form, under `sourceCommand`, with templates and arg groups, and a file name;
  - **Volume**: the app's own-directory mounts, by target path, and a path, with the note on reading a volume that is being written;
- **Repository**: the backup repositories of the app's env and the scopes above it, with a link to configure them;
- tags;
- **Scheduling**: priority, schedule with No schedule, timeout, retry (max and delay) and canceling;
- triggers;
- notification.

Without the job type in the enum, a list holding a data backup fails to parse. Task 6 is needed before a data backup job is created.

- [ ] **Step 1: Apply the patch**

Run: `git -C $D apply --index $(pwd)/$P/dash-01.patch`

- [ ] **Step 2: Check it**

Run: `(cd $D && npx tsc --noEmit && npm run lint && npx prettier --check src)`
Expected: no type errors, no lint output, `All matched files use Prettier code style!`

- [ ] **Step 3: Commit (dashboard repo)**

```bash
git -C $D commit -m "feat(sched-jobs): a data backup's form, its route and its list tag

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 7: A run's snapshot on its page (dashboard)

**Files:** in `$D/src/application/modules/operations/`, modify `api/services/system-tasks-services/system-tasks.api.validator.ts`, `domain/system-task.entity.ts` and `routes/tasks/building-blocks/task-summary-card/system-task-summary-card.com.tsx`.

**Interfaces:**
- Consumes: `dataBackup {snapshotId, sizeBytes}` of Task 5.
- Produces: `SystemTask.dataBackup`. The run's details show `Snapshot: <id> (<size>)`, with the size from `formatDataSizeCompact`.

- [ ] **Step 1: Apply the patch**

Run: `git -C $D apply --index $(pwd)/$P/dash-02.patch`

- [ ] **Step 2: Check it**

Run: `(cd $D && npx tsc --noEmit && npm run lint && npx prettier --check src)`
Expected: as in Task 6.

- [ ] **Step 3: Commit (dashboard repo)**

```bash
git -C $D commit -m "feat(tasks): a data backup's run shows its snapshot

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 8: Live checks on the Linux server, and the browser

These cannot run here: the Docker Desktop swarm is the user's, and nothing is deployed to it. The user runs them, or has them run, on the Linux server, with images built from the merged branches.

- [ ] **Check 1 - `pg_dump` into an S3 repository.**
  - A Postgres app, and a data backup with source Command `pg_dump -U postgres app`, file `db.sql`, and an S3 repository.
  - Run now. Expected:
    - the run is `done`, and its page shows the snapshot and a size near the dump's;
    - the repository's snapshot list has it, with the tags;
    - the backend's memory stays flat on a database of a few GB (Review Focus 1).
- [ ] **Check 2 - a volume into an S3 repository, from a node other than the manager.**
  - An app whose own directory of a cluster volume is pinned to a worker.
  - Run now twice. Expected: both runs `done`, and the snapshots' paths read `/<job id>` (Review Focus 3).
- [ ] **Check 3 - a volume into a volume repository on the same node.** Expected: `done`.
- [ ] **Check 4 - a failing command leaves no snapshot.**
  - `sh -c 'pg_dump -U postgres app; exit 1'` gives `failed` and no new snapshot in the list, `kopia snapshot list --all` included.
  - A timeout shorter than a large dump gives the same (Review Focus 2).
- [ ] **Check 5 - two runs at once.** Run now twice quickly on a slow dump, one of them failing (edit the command in between). Expected: only the failed run's snapshot is gone.
- [ ] **Check 6 - a snapshot opened with the kopia CLI.** `kopia snapshot list --all` lists it under `hivepaas@data-backup:/<job id>`, and `kopia snapshot restore` gives back `db.sql`.
- [ ] **Check 7 - a command that prints nothing.** It gives `done`, and the run's page shows `0 B` (Review Focus 4).
- [ ] **Check 8 - on `pre-deploy`, the deploy waiting for it.** A deploy waits for the backup and goes on when it ends.
- [ ] **Check 9 - the combinations not built yet.** A command into a volume repository, and a volume on another node than a volume repository, each save, and their runs fail saying why (`ERR_NOT_IMPLEMENTED`).
- [ ] **Check 10 - browser:**
  - create and edit a data backup of each source;
  - the Volume picker lists the app's own mounts only;
  - a wrong file name, a `../` path, and a `hivepaas.x` tag are refused on the form;
  - the `Backup · <repository>` tag is in the list;
  - a run's snapshot is on its page;
  - an env's job list filters by the type.
