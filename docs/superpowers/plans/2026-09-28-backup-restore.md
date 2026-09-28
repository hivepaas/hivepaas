# Backup Restore Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restore a backup snapshot into the app that took it, or into another app, from the Backup Snapshots views: a command snapshot through a command run in the app, a volume snapshot into the app's directory of a volume, replaced or overwritten.

**Architecture:**
- **Engine:** kopia lists a snapshot's directories (`kopia ls -l`), streams one file (`kopia show` on the file's object), and restores a snapshot or one directory of it (`kopia snapshot restore <id>/<path>`).
- **Repository service:** `RestoreStream`, `RestoreDirectory` and `ListEntries` reach the repository as backups do: directly, on the data's node, or through a repository server as `hivepaas@restore`.
- **Data backup service:** `Restore` runs the steps.
  - A command snapshot is streamed into a command run in the app, with the new `RunCommand`.
  - A volume snapshot: stop the app, move the directory aside, restore, and put it back on failure or cancel. The app is started again whatever happened.
- **A task, `task:backup-restore`,** runs the restore with a log. The snapshots use case records it after its checks, at the four scopes' `…/backup-snapshots/:itemID/restore`. An `…/entries` endpoint lists a snapshot for the dialog.
- **Dashboard:**
  - the data backup job form gains a Restore Command;
  - the Backup Snapshots views gain a Restore drawer, from the row menu and the details drawer;
  - a restore's task card links its snapshot.

**Tech Stack:** Go (bun, gin, fx), kopia 0.23; React 19, TanStack Query, react-hook-form, zod.

**Spec:** `docs/superpowers/specs/2026-09-28-backup-restore-design.md`

## How the patches work

Every task's code was written and verified before this plan.

**Bases:**
- The backend patches apply in order to `main` at `d7709fbd`. This plan's commit changes only `docs/`.
- The dashboard patches apply to the dashboard's `main` at `2efb2faa`.

**What was checked:**
- **Backend, on a fresh worktree, at every task:**
  - the tests alone fail (to build, or on their asserts);
  - with the code, the task's packages pass and `golangci-lint run ./...` finds 0 issues.
- **Backend, after the last task:** `go test ./...` passes, and `make gen-swag` leaves `swagger.json` unchanged.
- **The kopia steps ran against real kopia 0.23:** `ls -l` output, restoring `<id>/<dir>`, and `show` refusing a snapshot ID where it wants an object ID.
- **Dashboard, after each patch:** `tsc`. After the last one: `npm run lint` and `prettier --check src`.
- **The new JSON** - a snapshot's job with its restore command, and entries - was compared with the dashboard's zod schemas.

Patches live in `docs/superpowers/plans/2026-09-28-backup-restore/`:
- `be-NN-tests.patch`, then `be-NN-code.patch`, for backend task NN;
- `dash-NN.patch` for dashboard task NN.

The commands below run from the backend repo root, with `P=docs/superpowers/plans/2026-09-28-backup-restore` and `D=../hivepaas-dashboard`.

## Global Constraints

- **Backend, before a task is done:**
  - `go build ./...`;
  - `golangci-lint run ./...` over the whole repo;
  - `go test ./...`.
  - `go test ./hivepaas_app/cmd/...` includes the fx wiring test.
- **`make gen-swag`** in Task 9, then `go mod verify`. A first run that fails with `cannot find all dependencies` is run again.
- **Dashboard:** `npx tsc --noEmit`, `npm run lint`, `npx prettier --check src`.
- **Main checkout:** after the merge, run `go test ./...` in the main checkout too. It has a local `vendor/` (git-ignored), so it builds with `-mod=vendor`. This plan imports nothing new.
- **Routes:**

  | Scope | Base |
  |---|---|
  | Global | `/settings/backup-snapshots` |
  | Project | `/projects/:projectID/backup-snapshots` |
  | Env | `/projects/:projectID/:projectEnv/backup-snapshots` |
  | App | `/projects/:projectID/:projectEnv/apps/:appID/backup-snapshots` |

  Each gains `GET /:itemID/entries?path=` and `POST /:itemID/restore`.
- **Restore request:** `{targetApp:{id}, command?, volume?:{id}, subpath?, snapshotPath?, stopApp?, mode?: "replace"|"overwrite"}`.
  - The answer is `{data:{task:{id}}}`.
  - `mode: replace` requires `stopApp: true`.
- **The task:**
  - type `task:backup-restore`;
  - scope app, object the target app, target the snapshot record;
  - the kopia server user `hivepaas@restore`;
  - a directory moved aside is named `<dir>.before-restore-YYYYMMDD-HHMMSS` (UTC).
- **Git:**
  - commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`;
  - merge locally, never push;
  - never `git stash`, and never touch the user's uncommitted work.

## Review Focus

1. **`psql` loading a dump into an app on another node than the backend.**
   - The stream goes into a remote exec, and the command must see its input end and finish.
   - Pinned by: `TestForwardExecInputClosesStdinWhenTheClientIsDone` (Task 5), and Task 12, check 1, live.
2. **A command snapshot in a repository on a volume.**
   - It is read here through a repository server. Before this plan, every backup through a server failed (Task 3's fix).
   - Pinned by: `TestNewEngineTakesEveryStorage` (Task 3), `TestRestoreStreamOutOfAVolumeRepositoryGoesThroughAServer` (Task 4), and Task 12, check 2.
3. **Replace failing or canceled half way:** the directory is back as it was, and the app runs again.
   - Pinned by: `TestRestoreVolumeThatFailsPutsTheDirectoryBack` and `TestRestoreVolumeCanceledPutsTheDirectoryBack` (Task 6), and Task 12, check 3, live with a cancel.
4. **A snapshot taken outside HivePaaS, or whose job is gone:** its kind and file come from what it holds.
   - Pinned by: `TestSnapshotKind` (Task 8), and Task 12, check 4.
5. **Two restores into one app asked for at once.**
   - The in-flight check reads the tasks table without a lock, so two requests in the same instant can both pass.
   - Pinned by: `TestRestoreInFlightQuery` (Task 8) for the query only. The race is accepted, and named under Decisions.

## Decisions beyond the spec

These were taken while building the patches, and written back into the spec where they change it.

- **Two bugs on `main` found on the way, fixed here:**
  - `backup.NewEngine` refused a storage that is only a repository server. Every backup through a server failed with `ERR_BACKUP_STORAGE_TYPE_REQUIRED` (Task 3).
  - The agent never closed a remote exec's stdin, so a command reading to the end of its input never finished (Task 5).
- **`kopia show` reads an object, not a snapshot:** the file's object ID is found with `kopia ls -l <snapshot id>`.
- **Entries come back directories first, then by name.** kopia's own order is not stable.
- **Where a part goes:** a directory of the snapshot, `snapshotPath`, is restored to the same place under the target: `<app dir in volume>/<subpath>/<snapshotPath>`. The target directory is created (`mkdir -p`) before the restore, in both modes.
- **The task never retries by itself.** Its timeout is the long default: a restore is bounded by the snapshot's size.
- **A restore is audited** as an app update with the section `backup-restore`: the snapshot, the repository, the mode.
- **The in-flight check has no lock.** Two requests in one instant can both be recorded. The cost is two restores racing on one app, which a person asked for twice.
- **What a restore needs from the job** comes with every snapshot row:
  - `job.fileName`, `job.restoreCommand`, `job.sourceVolumeId`, `job.sourceVolumeSubpath`;
  - `app.projectId`.

  A restore command's script comes through only when it is inline, which is how the job form writes it.
- **The task API** gains `backupRestore` (snapshot, repository, path, mode) for the task card's link.
- **A data backup's source command script** is now a reference of the job too, as the restore command's is.
- **Dashboard:**
  - **The global view's target picker** is project, then app: the apps carry their env as a badge, instead of a separate env step.
  - **The shared directory warning** names the apps given a directory of the target app's storage (`borrowedBy`). That is close to "the apps that mount the target directory" without a new API.
  - **Restore is offered** to those with write on the Project module; the server checks write on the target app.

---

## Task 1: The restore command on a data backup job

**Files:**
- Modify: `hivepaas_app/entity/setting_sched_job_data_backup.go`, `hivepaas_app/usecase/settings/schedjobuc/schedjobdto/data_backup.go`
- Test: `hivepaas_app/entity/setting_sched_job_data_backup_test.go`, `hivepaas_app/usecase/settings/schedjobuc/schedjobdto/data_backup_test.go`

**Interfaces:**
- Produces:
  - `entity.SchedJobDataBackup.RestoreCommand *CommandTemplate`, whose script is a reference, like the source command's;
  - on the request, `restoreCommand`: a command source's only, forced without a TTY;
  - on the response, `restoreCommand`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-01-tests.patch`

Tests added:
- `TestSchedJobDataBackupReferencesItsCommandsScripts`;
- `TestADataBackupOfACommandMayHaveARestoreCommand`;
- `TestADataBackupOfAVolumeHasNoRestoreCommand`;
- `TestTransformSchedJobGivesADataBackupsRestoreCommand`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/entity/ ./hivepaas_app/usecase/settings/schedjobuc/...`
Expected: FAIL, build error `unknown field RestoreCommand`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-01-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/entity/ ./hivepaas_app/usecase/settings/schedjobuc/... && golangci-lint run ./...`
Expected: PASS; `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(sched-jobs): a data backup of a command says how to load it back

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 2: The engine lists a snapshot and restores a part of it

**Files:**
- Create: `services/backup/kopia/entries.go`
- Modify: `services/backup/kopia/restore.go`, `services/backup/backupmodel/engine.go`, `services/backup/backupmodel/types.go`, `services/backup/engine.go`
- Test: `services/backup/kopia/entries_test.go`, `services/backup/kopia/restore_integration_test.go` (skips without the kopia binary)

**Interfaces:**
- Produces:
  - `Engine.ListEntries(ctx, snapshotID, path string) ([]SnapshotEntry, error)`;
  - `SnapshotEntry{Name, Dir, SizeBytes}`;
  - `RestoreOptions{Path}`;
  - `RestoreStream` finds the file's object with `ls -l`, then `show`s it.
- Errors keep kopia's words, as every command's do.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-02-tests.patch`

Tests added:
- `TestParseEntries`, `TestParseEntriesRefusesWhatItCannotRead`, `TestSnapshotObjectPath`;
- `TestIntegration_RestoreStream_GivesBackTheFile`, `TestIntegration_ListEntriesThenRestoreOneDirectory`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./services/backup/kopia/`
Expected: FAIL, build error `undefined: backupmodel.SnapshotEntry`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-02-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./services/backup/... && golangci-lint run ./...`
Expected: PASS (the integration tests run where kopia is installed); `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(backup): the engine lists a snapshot's directories, and restores one of them or a stream

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 3: An engine takes a repository server's storage

**Files:**
- Modify: `services/backup/factory.go`
- Test: `services/backup/factory_test.go`

- [ ] **Step 1: Write the failing test**

Run: `git apply --index $P/be-03-tests.patch` (`TestNewEngineTakesEveryStorage`)

- [ ] **Step 2: Run it to watch it fail**

Run: `go test ./services/backup/`
Expected: FAIL, `ERR_BACKUP_STORAGE_TYPE_REQUIRED` for `server`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-03-code.patch`

- [ ] **Step 4: Run it to watch it pass**

Run: `go test -count=1 ./services/backup/ && golangci-lint run ./...`
Expected: PASS; `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "fix(backup): an engine takes a repository server's storage

NewEngine refused a storage that is only a repository server, so every backup
through a server failed with ERR_BACKUP_STORAGE_TYPE_REQUIRED.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 4: Restore and list, routed as backups are

**Files:**
- Create: `hivepaas_app/service/backupreposervice/backupreposerviceimpl/restore.go`
- Modify: `…/backupreposerviceimpl/backup.go` (`runWith` and `throughServer` shared with backups), `hivepaas_app/service/backupreposervice/service.go`, `types.go`
- Test: `…/backupreposerviceimpl/restore_test.go`

**Interfaces:**
- Consumes: Task 2's engine methods, Task 3's fix.
- Produces, on `backupreposervice.Service`:
  - `RestoreStream(ctx, db, *RestoreStreamReq{RepoTarget, SnapshotID, FileName, Stdout, Progress, OnConnected}) error`;
  - `RestoreDirectory(ctx, db, *RestoreDirectoryReq{RepoTarget, SnapshotID, Path, HostDir, NodeID, NodeLabel, Progress}) error`;
  - `ListEntries(ctx, db, *ListEntriesReq{RepoTarget, SnapshotID, Path}) ([]backup.SnapshotEntry, error)`.
  - The server user is `restoreServerUser = "hivepaas@restore"`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-04-tests.patch`

Tests added:
- `TestRestoreDirectoryOnTheRepositorysNode`;
- `TestRestoreDirectoryOnAnotherNodeGoesThroughAServer`;
- `TestRestoreStreamOutOfAVolumeRepositoryGoesThroughAServer`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/service/backupreposervice/...`
Expected: FAIL, build error `undefined: backupreposervice.RestoreDirectoryReq`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-04-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/service/backupreposervice/... && golangci-lint run ./...`
Expected: PASS; `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(backup): restore a stream or a directory, and list a snapshot, routed as backups are

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 5: A command outside a job, reading a stream to its end

**Files:**
- Create: `hivepaas_app/service/schedjobexecservice/schedjobexecserviceimpl/run_command.go`
- Modify: `hivepaas_app/service/schedjobexecservice/service.go`, `types.go`, `hivepaas_app/usecaseagent/containeragentuc/command_exec.go`
- Test: `…/schedjobexecserviceimpl/run_command_test.go`, `…/exec_writer_test.go` (its fake records stdin), `hivepaas_app/usecaseagent/containeragentuc/command_exec_test.go`

**Interfaces:**
- Produces:
  - `schedjobexecservice.Service.RunCommand(ctx, db, *RunCommandReq{TaskExecData, App, Command, Stdin}) (*RunCommandResp{ExitCode}, error)`: no TTY, output to the task's log;
  - on the agent, `forwardExecInput(recv, stdin, resize)`, which closes the exec's stdin on the client's EOF.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-05-tests.patch`

Tests added:
- `TestRunCommandGivesTheCommandItsStdin`, `TestRunCommandRefusesAnEmptyCommand`;
- `TestForwardExecInputClosesStdinWhenTheClientIsDone`, `TestForwardExecInputLeavesStdinOpenWhenTheStreamBreaks`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/service/schedjobexecservice/... ./hivepaas_app/usecaseagent/containeragentuc/`
Expected: FAIL, build errors `undefined: schedjobexecservice.RunCommandReq` and `undefined: forwardExecInput`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-05-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/service/schedjobexecservice/... ./hivepaas_app/usecaseagent/containeragentuc/ && golangci-lint run ./...`
Expected: PASS; `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(sched-jobs): run a command in an app outside a job, reading a stream to its end

The agent now closes an exec's stdin when the client has sent all of it: a
command reading to its end, on another node, finished never.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 6: The restore's steps

**Files:**
- Create: `hivepaas_app/service/databackupservice/databackupserviceimpl/restore.go`
- Modify: `…/databackupserviceimpl/service.go` (new constructor parameters `docker.Manager`, `appservice.Service`, `nodeexecservice.Service`; `swarmAppControl`), `hivepaas_app/service/databackupservice/service.go`, `hivepaas_app/base/sched_job.go` (`BackupRestoreMode`)
- Test: `…/databackupserviceimpl/restore_test.go`

**Interfaces:**
- Consumes: Task 4's `RestoreStream` and `RestoreDirectory`, Task 5's `RunCommand`.
- Produces:
  - `databackupservice.Service.Restore(ctx, db, *RestoreReq{TaskExecData, Target, SnapshotID, App, Command *RestoreCommand{Command, FileName}, Volume *RestoreVolume{VolumeID, Subpath, SnapshotPath, StopApp, Mode}}) error`;
  - `base.BackupRestoreModeReplace`, `base.BackupRestoreModeOverwrite`.
- Host commands run through the node's agent, without a shell, on `/host` paths:
  - `test -e <dir>`, then `mv -- <dir> <aside>`, then `mkdir -p -- <dir>`;
  - on failure, `rm -rf -- <dir>`, then `mv -- <aside> <dir>`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-06-tests.patch`

Tests added:
- `TestRestoreVolumeReplacesTheDirectory`, `…ReplacesADirectoryThatIsNotThere`, `…ThatFailsPutsTheDirectoryBack`, `…CanceledPutsTheDirectoryBack`;
- `…OverwritesWithoutStopping`, `…LeavesAStoppedAppStopped`, `…ReplaceRefusesARunningApp`;
- `TestRestoreCommandStreamsTheFileIntoTheCommand`, `…FailsWhenTheStreamBreaks`, `…FailsWhenTheCommandFails`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/service/databackupservice/...`
Expected: FAIL, build error `unknown field apps in struct literal of type service`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-06-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/service/databackupservice/... ./hivepaas_app/cmd/... && golangci-lint run ./...`
Expected: PASS, the fx wiring test included; `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(backup): restore a snapshot into an app - a stream into a command, a directory replaced or overwritten

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 7: The restore task

**Files:**
- Create: `hivepaas_app/tasks/taskbackuprestore/executor.go`, `hivepaas_app/entity/task_backup_restore.go`
- Modify:
  - `hivepaas_app/base/task.go`: `TaskTypeBackupRestore`, in `AllTaskTypes` and `AllProjectTaskTypes`;
  - `hivepaas_app/tasks/queue/queueimpl/task_timeout.go`;
  - `hivepaas_app/tasks/initializer/initializer.go`, `hivepaas_app/registry/provides.go`;
  - `hivepaas_app/usecase/taskuc/taskdto/get.go` (`backupRestore`).
- Test: `hivepaas_app/tasks/taskbackuprestore/executor_test.go`, `hivepaas_app/entity/setting_sched_job_data_backup_test.go` (`TestTaskBackupRestoreArgs`), `hivepaas_app/usecase/taskuc/taskdto/backup_restore_test.go`

**Interfaces:**
- Consumes: Task 6's `Restore`.
- Produces:
  - `entity.TaskBackupRestoreArgs{ProjectID, AppID, RepoID, SnapshotID, Command, FileName, Volume, Subpath, SnapshotPath, StopApp, Mode}` and `Task.ArgsAsBackupRestore()`;
  - `taskdto.TaskBackupRestoreResp{SnapshotRecordID, RepoID, SnapshotID, SnapshotPath, FileName, Mode, StopApp}`.
- The executor:
  - keeps its log as the job runs do (`task:<id>:log`, then `task_logs`);
  - fails when the snapshot's record is gone;
  - loads the command's references.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-07-tests.patch` (`TestRestoreReqOf`, `TestTaskBackupRestoreArgs`, `TestTransformTaskGivesARestoresSnapshot`)

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/entity/ ./hivepaas_app/tasks/taskbackuprestore/ ./hivepaas_app/usecase/taskuc/...`
Expected: FAIL, build error `undefined: base.TaskTypeBackupRestore`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-07-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/entity/ ./hivepaas_app/tasks/... ./hivepaas_app/usecase/taskuc/... ./hivepaas_app/cmd/... && golangci-lint run ./...`
Expected: PASS; `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(backup): a backup restore task

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 8: Asking for a restore, and listing a snapshot

**Files:**
- Create: `hivepaas_app/usecase/settings/backupsnapshotuc/restore.go`, `entries.go`, `…/backupsnapshotdto/restore.go`, `entries.go`
- Modify: `…/backupsnapshotuc/uc.go` (new parameters `repository.TaskRepo`, `queue.TaskQueue`, `databackupservice.Service`), `…/backupsnapshotdto/transform.go`
- Test: `…/backupsnapshotuc/restore_test.go`, `…/backupsnapshotdto/restore_test.go`, `…/backupsnapshotdto/transform_test.go`

**Interfaces:**
- Consumes: Task 4's `ListEntries`, Task 6's `CheckAppVolume` (existing), Task 7's task type and arguments.
- Produces:
  - `UC.RestoreBackupSnapshot(ctx, auth, *RestoreBackupSnapshotReq) (*RestoreBackupSnapshotResp, error)`;
  - `UC.ListBackupSnapshotEntries(ctx, auth, *ListBackupSnapshotEntriesReq{Scope, ID, Path "path"})`;
  - `SnapshotJobResp` gains `fileName`, `restoreCommand`, `sourceVolumeId`, `sourceVolumeSubpath`, and `SnapshotAppResp` gains `projectId`.
- The checks, in order:
  - the snapshot in the view's reach, read;
  - an active repository;
  - the kind - its tag, its job, or one file at its root;
  - a request of that kind;
  - the target app: write on its env, the volume mounted as its own directory, no restore in flight;
  - the audit entry;
  - the task.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-08-tests.patch`

Tests added:
- `TestSnapshotKind`, `TestRestoreFitsTheSnapshot`, `TestRestoreInFlightQuery`, `TestRestoreTask`;
- `TestRestoreRequestOfACommand`, `…OfAVolume`, `…IsACommandsOrAVolumes`;
- `TestTransformBackupSnapshotGivesItsJobsRestore`, `…GivesItsJobsVolume`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/usecase/settings/backupsnapshotuc/...`
Expected: FAIL, build errors `undefined: restoreKind` and `undefined: NewRestoreBackupSnapshotReq`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-08-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/usecase/settings/backupsnapshotuc/... ./hivepaas_app/cmd/... && golangci-lint run ./...`
Expected: PASS; `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(backup): ask for a snapshot's restore, and list what a snapshot holds

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 9: The four scopes' endpoints

**Files:**
- Modify:
  - `hivepaas_app/interface/api/handler/basesettinghandler/xtra_api_backup_snapshot.go`;
  - `…/settinghandler/backup_snapshot.go`, `…/projectsettingshandler/backup_snapshot.go`, `…/projectenvsettingshandler/backup_snapshot.go`, `…/appsettingshandler/backup_snapshot.go`;
  - `hivepaas_app/interface/api/server/router_settings.go`, `router_projects.go`, `router_project_env.go`, `router_apps.go`;
  - `docs/openapi/swagger.json`.
- Test: `…/basesettinghandler/backup_snapshot_restore_request_test.go`

**Interfaces:**
- Consumes: Task 8's use case.
- Both endpoints read the snapshot as the scope's viewer (`read`). The use case checks write on the target app.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-09-tests.patch` (`TestBackupSnapshotEntriesRequestParses`, `TestBackupSnapshotRestoreRequestParses`)

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/interface/api/handler/basesettinghandler/`
Expected: FAIL, build error `(*Handler).RestoreBackupSnapshot undefined`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-09-code.patch`. Then run `make gen-swag`, which must leave `swagger.json` unchanged, and `go mod verify`.

- [ ] **Step 4: Check it**

Run: `go build ./... && go test ./... && golangci-lint run ./...`
Expected: PASS, the fx wiring test included; `0 issues.` `grep -c backup-snapshots docs/openapi/swagger.json` gives `16`.

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(backup): the restore and entries endpoints at the global, project, env and app scopes

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 10: The restore command in the job form (dashboard)

**Files** (under `$D/src/application/modules/projects/`):
- Modify: `domain/apps/scheduled-job/app-scheduled-job.entity.ts`, `api/services/project-apps-services/scheduled-jobs/app-scheduled-jobs.api.contracts.ts`, `…validator.ts`, and `module-shared/components/data-backup-form/*`.
- Modify: `module-shared/components/command-config-section/command-config-section.com.tsx` (its label takes a node).

- [ ] **Step 1: Apply the patch**

Run: `git -C $D apply --index $(pwd)/$P/dash-01.patch`

- [ ] **Step 2: Check it**

Run: `(cd $D && npx tsc --noEmit && npm run lint && npx prettier --check src)`
Expected: no type errors, no lint output, `All matched files use Prettier code style!`

- [ ] **Step 3: Commit (dashboard repo)**

```bash
git -C $D commit -m "feat(sched-jobs): a data backup's form takes the command that restores it

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 11: A snapshot's restore and entries in the data layer (dashboard)

**Files** (under `$D/src/application/modules/settings/`): `domain/backup-snapshot.entity.ts`, `api/services/backup-snapshot-services/*`, `api/hooks/use-backup-snapshot.api.ts`, `data/queries/backup-snapshot.queries.ts`, `data/commands/backup-snapshot.commands.ts`, `data/constants/settings.query-keys.ts`.

**Interfaces:**
- Produces:
  - `BackupSnapshotQueries.useFindEntries({scope, id, path})`;
  - `BackupSnapshotCommands.useRestore()` taking `{scope, id, payload: BackupSnapshot_Restore_Payload}`;
  - `BACKUP_RESTORE_MODE`, `BackupSnapshotEntry`, `BackupSnapshotJob`.

- [ ] **Step 1: Apply the patch**

Run: `git -C $D apply --index $(pwd)/$P/dash-02.patch`, then the checks of Task 10, Step 2.

- [ ] **Step 2: Commit (dashboard repo)**

```bash
git -C $D commit -m "feat(backup): a snapshot's restore and entries in the dashboard's data layer

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 12: The Restore drawer, and the checks (dashboard, then live)

**Files** (under `$D/src/application/modules/`):
- Create, under `settings/module-shared/components/backup-snapshot-table/`:
  - `backup-snapshot-restore-drawer.com.tsx`;
  - `backup-snapshot-restore-target.com.tsx`;
  - `backup-snapshot-path-picker.com.tsx`;
  - `backup-snapshot-restore.helpers.ts`.
- Modify:
  - the table, its row menu, its details drawer and its helpers;
  - `settings/domain` and the snapshot validator (`app.projectId`);
  - `operations/domain/system-task.entity.ts`, the task validator and the task card (`backupRestore`);
  - the tasks filter's type list;
  - the data backup form's exports.

- [ ] **Step 1: Apply the patch**

Run: `git -C $D apply --index $(pwd)/$P/dash-03.patch`, then the checks of Task 10, Step 2.

- [ ] **Step 2: Commit (dashboard repo)**

```bash
git -C $D commit -m "feat(backup): restore a snapshot from the Backup Snapshots views

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

- [ ] **Step 3: Live and browser checks** (the user, on the Linux server)
  - **Check 1:** restore a `db.sql` into a Postgres app whose container is on another node than the backend. The task ends, and the data is there (Review Focus 1).
  - **Check 2:** the same from a repository on a volume, through the repository server. Also take a data backup of a command into a volume repository, which Task 3 fixes (Review Focus 2).
  - **Check 3:** replace a directory:
    - look at `….before-restore-…`;
    - fail one restore on purpose (a snapshot path that is not there);
    - cancel another half way.

    The directory is back each time and the app runs (Review Focus 3).
  - **Check 4:** a snapshot synced in from outside HivePaaS: one file restores through a command, a directory into a volume (Review Focus 4).
  - **Check 5:** the drawer at the four views:
    - the target picker, another app's warning;
    - the snapshot tree;
    - Stop locked under Replace;
    - the confirmation;
    - the task page, and its snapshot link.
  - **Check 6:** the job form's Restore Command saves, and pre-fills the drawer.
