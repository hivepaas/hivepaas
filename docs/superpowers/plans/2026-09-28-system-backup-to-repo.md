# System Backup Into a Backup Repository Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** The system backup takes HivePaaS's database, the spec of the whole installation, or both, as one snapshot of a global backup repository. Any file of any snapshot can be downloaded.

**Architecture:**
- **Backend:**
  - The backup repository service gains `BackupLocalDirectory`: kopia, in the backend, snapshots a directory of the backend's own. It writes into cloud storage directly, and into a volume repository through its server.
  - The system backup service is rewritten:
    - `pg_dump -Fc` into `db.pg_dump`;
    - the spec export at the global scope into `spec.tar.gz[.age]`;
    - one snapshot, then the repository's records are synced.
  - The configuration is replaced.
  - The old files, their API, file kind and path, and the system cleanup's retention of them go.
  - A download endpoint streams a snapshot's file at the four scopes, recorded in the audit log.
- **Dashboard:**
  - a new configuration form;
  - no Backup Files page;
  - a Files browser with Download in the snapshot details;
  - *System* as a source;
  - a system run's card links its snapshot.

**Tech Stack:** Go (bun, gin, fx), kopia 0.23, pg_dump; React 19, TanStack Query, react-hook-form, zod.

**Spec:** `docs/superpowers/specs/2026-09-28-system-backup-to-repo-design.md`

## How the patches work

Every task's code was written and verified before this plan.

**Bases:**
- The backend patches apply in order to `main` at `a87588ea`. This plan's commit changes only `docs/`.
- The dashboard patches apply to the dashboard's `main` at `ac091d76`.

**What was checked:**
- **Backend, on a fresh worktree, at every task:**
  - the tests alone fail;
  - with the code, the task's packages pass and `golangci-lint run ./...` finds 0 issues.
  - Task 3 removes code and has no test of its own: its gate is the build and the suite.
- **Backend, after the last task:** `go test ./...` passes, and `make gen-swag` leaves `swagger.json` unchanged.
- **Dashboard, after each patch:** `tsc`. After the last one: `npm run lint` and `prettier --check src`.

Patches live in `docs/superpowers/plans/2026-09-28-system-backup-to-repo/`:
- `be-NN-tests.patch` (none for Task 3), then `be-NN-code.patch`, for backend task NN;
- `dash-NN.patch` for dashboard task NN.

The commands below run from the backend repo root, with `P=docs/superpowers/plans/2026-09-28-system-backup-to-repo` and `D=../hivepaas-dashboard`.

## Global Constraints

- **Backend, before a task is done:**
  - `go build ./...`;
  - `golangci-lint run ./...` over the whole repo;
  - `go test ./...`.
  - `go test ./hivepaas_app/cmd/...` includes the fx wiring test.
- **`make gen-swag`** in Task 4 (it covers Tasks 2 to 4), then `go mod verify`. A first run that fails with `cannot find all dependencies` is run again.
- **Dashboard:** `npx tsc --noEmit`, `npm run lint`, `npx prettier --check src`.
- **Main checkout:** after the merge, run `go test ./...` there too. It has a local `vendor/`. This plan imports nothing new.
- **Names:**
  - source `hivepaas@system-backup:/system`;
  - tags `hivepaas.source:system-backup` and `hivepaas.run:<task id>`;
  - files `db.pg_dump` and `spec.tar.gz`, or `spec.tar.gz.age` when encrypted;
  - audit type `backup-download`.
- **Configuration JSON:** `{status, schedule, includeDB, includeSpec, specSecrets: encrypted|omit|plaintext, specPassphrase, targetRepository:{id}, notification}`.
- **Download:** `GET …/backup-snapshots/:itemID/download?path=<file>` at the four scopes of `backup-snapshots`.
- **Git:**
  - commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`;
  - merge locally, never push;
  - never `git stash`, and never touch the user's uncommitted work.

## Review Focus

1. **A system backup into a repository on a volume.**
   - kopia runs in the backend and reaches it through the repository server, a path that was broken on `main` until the restore plan fixed it.
   - Pinned by: `TestBackupLocalDirectoryIntoAVolumeRepositoryGoesThroughAServer` (Task 1), and Task 6, check 1, live.
2. **An encrypted spec, then imported again.**
   - The file is `spec.tar.gz.age`, and the Import page must take it with the passphrase.
   - Pinned by: `TestSystemBackupTakesTheDatabaseAndTheSpecInOneSnapshot` (Task 2) for the name, and Task 6, check 2.
3. **Downloading a file of several hundred megabytes.** The blob is held in the browser's memory.
   - Pinned by: Task 6, check 3, which measures it live. A short-lived link is a later spec.
4. **A configuration saved by a person without the reveal capability,** with encrypted or plaintext secrets: it must be refused, and the refusal recorded.
   - Pinned by: Task 6, check 4, live. The gate is the permission manager's own, tested there.
5. **The `pg_dump` in the backend image** must speak the server's major version, or `-Fc` fails.
   - Pinned by: Task 6, check 1: `pg_restore --list` on the downloaded dump.

## Decisions beyond the spec

These were taken while building the patches, and written back into the spec where they change it.

- **An encrypted spec is `spec.tar.gz.age`.** The export encrypts the whole bundle with age and names it so; the snapshot keeps the name.
- **The secrets gate at save time is the capability alone** (`AuthorizeSecretMount`), recorded as a reveal. The backup puts the secrets in a repository, as a mount puts one in a container, and hands the caller nothing in the clear. So the operator's flag on what the API returns does not apply, which would otherwise refuse the default mode.
- **A disabled configuration may name no repository and take nothing.** The defaults are disabled, so they save as they are.
- **The defaults:** disabled, the database only, secrets encrypted, no repository and no passphrase.
- **The system cleanup's retention of backup files goes,** from its settings, run, API and form. The repository's retention keeps the snapshots now.
- **The run's output** is `{snapshotId, sizeBytes, includes}`. Those are a data backup's names, so the task API reads it with the same code, and the task card shows it.
- **A download is recorded** as `backup-download`: against the app for an app's snapshot, else against the repository's scope, before anything is read.
- **The work directory** is a temporary directory under the app path, removed after the run.
- **The refmap tests** used `SystemBackupCloudStorage` as their example of a bespoke reference. They now use `SchedJobCommandOutputFileStorage`.

---

## Task 1: Back up a directory of the backend's own

**Files:**
- Modify:
  - `hivepaas_app/service/backupreposervice/service.go`, `types.go` (`BackupLocalDirectoryReq`);
  - `…/backupreposerviceimpl/backup.go`.
- Test: `…/backupreposerviceimpl/restore_test.go`

**Interfaces:**
- Produces: `BackupLocalDirectory(ctx, db, *BackupLocalDirectoryReq{RepoTarget, Dir, Source, Description, Tags, Progress}) (*BackupResp, error)`.

- [ ] **Step 1: Write the failing test**

Run: `git apply --index $P/be-01-tests.patch` (`TestBackupLocalDirectoryIntoAVolumeRepositoryGoesThroughAServer`)

- [ ] **Step 2: Run it to watch it fail**

Run: `go test ./hivepaas_app/service/backupreposervice/...`
Expected: FAIL, build error `undefined: backupreposervice.BackupLocalDirectoryReq`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-01-code.patch`

- [ ] **Step 4: Run it to watch it pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/service/backupreposervice/... && golangci-lint run ./...`
Expected: PASS; `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(backup): back up a directory of the backend's own into a repository

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 2: The new configuration and the run

**Files:**
- Modify:
  - `hivepaas_app/entity/setting_system_backup.go`, `task_system_backup.go`, `setting_refmap.go` (a comment);
  - `hivepaas_app/usecase/systemsettings/systembackupuc/update.go`, `systembackupdto/update.go`, `systembackupdto/get.go`;
  - `hivepaas_app/service/settinginitservice/settinginitserviceimpl/init_defaults_sys_backup.go`;
  - `hivepaas_app/service/sysbackupservice/sysbackupserviceimpl/service.go`, `sys_backup.go`, `sys_backup_db.go`.
- Delete: `…/sysbackupserviceimpl/sys_backup_files.go`, `sys_backup_save_in_local.go`, `sys_backup_save_in_storage.go`, `sys_backup_files_test.go`.
- Test:
  - `hivepaas_app/entity/setting_system_backup_test.go`, `setting_refmap_test.go`;
  - `…/systembackupdto/update_test.go`, `…/systembackupuc/update_test.go`;
  - `…/sysbackupserviceimpl/sys_backup_test.go`.

**Interfaces:**
- Consumes: Task 1's `BackupLocalDirectory`, the spec service's `Export`.
- Produces:
  - `entity.SystemBackup{Schedule, IncludeDB, IncludeSpec, SpecSecrets string, SpecPassphrase EncryptedField, TargetRepository ObjectID, Notification}` and `Includes() []string`;
  - `entity.TaskSystemBackupOutput{SnapshotID, SizeBytes, Includes}`;
  - `sysbackupserviceimpl.New(settingRepo, backupRepoService, scopeService, specService)`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-02-tests.patch`

Tests added:
- `TestSystemBackupReferencesItsRepository`, `TestSystemBackupIncludes`;
- the refmap tests on another bespoke reference;
- `TestASystemBackupOfTheDatabaseAndTheSpec`, `TestASystemBackupTakesSomethingIntoARepository` (a disabled one excepted), `TestASystemBackupsSpecSecrets`, `TestASystemBackupKeepsAMaskedPassphrase`;
- `TestCheckTargetRepository`;
- `TestSystemBackupTakesTheDatabaseAndTheSpecInOneSnapshot`, `TestSystemBackupOfTheDatabaseOnly`, `TestSystemBackupOfTheSpecOnly`, `TestSystemBackupThatCannotDumpTakesNothing`, `TestSystemBackupDeletesAHalfMadeSnapshot`, `TestSystemBackupOfNothingIsRefused`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/entity/ ./hivepaas_app/service/sysbackupservice/... ./hivepaas_app/usecase/systemsettings/systembackupuc/...`
Expected: FAIL, build error `unknown field IncludeDB in struct literal of type SystemBackup`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-02-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/entity/ ./hivepaas_app/service/sysbackupservice/... ./hivepaas_app/usecase/systemsettings/systembackupuc/... ./hivepaas_app/cmd/... && golangci-lint run ./...`
Expected: PASS, the fx wiring test included; `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(backup): the system backup takes the database, the spec, or both, into a backup repository

The configuration picks what to take, how the spec holds secrets, and a
repository of the global scope; compression, encryption, cloud storage and
local files go. A run dumps the database (pg_dump -Fc), exports the spec at the
global scope, and takes the two as one snapshot.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 3: The old files, their API and their cleanup go

**Files:**
- Delete: `hivepaas_app/interface/api/handler/systemsettingshandler/sys_backup_files.go`, `hivepaas_app/service/syscleanupservice/syscleanupserviceimpl/sys_cleanup_backups.go`.
- Modify:
  - `hivepaas_app/interface/api/server/router_system.go`;
  - `hivepaas_app/base/file.go`;
  - `hivepaas_app/config/path.go`;
  - `hivepaas_app/entity/setting_system_cleanup.go`, `task_system_cleanup.go`;
  - `hivepaas_app/service/syscleanupservice/types.go`, `…/syscleanupserviceimpl/sys_cleanup.go`;
  - `…/settinginitserviceimpl/init_defaults_sys_cleanup.go`;
  - `hivepaas_app/usecase/systemsettings/systemcleanupuc/systemcleanupdto/get.go`, `update.go`.

- [ ] **Step 1: Write the code**

Run: `git apply --index $P/be-03-code.patch`

- [ ] **Step 2: Check it**

Run: `go build ./... && go test -count=1 ./hivepaas_app/entity/ ./hivepaas_app/service/syscleanupservice/... ./hivepaas_app/usecase/systemsettings/... ./hivepaas_app/interface/... && golangci-lint run ./...`
Expected: PASS; `0 issues.` `grep -rn "FileKindSystemBackup\|BackupCleanup\b" hivepaas_app` finds nothing but `BackupRepoCleanup`.

- [ ] **Step 3: Commit**

```bash
git commit -m "feat(backup): the system backup's files, their API and their cleanup go

A system backup is a snapshot of a repository now, which the repository's
retention keeps: the Backup Files API, the system-backup file kind, the local
backup directory and the system cleanup's retention of backup files go.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 4: Download a file of a snapshot

**Files:**
- Create: `hivepaas_app/usecase/settings/backupsnapshotuc/download.go`, `…/backupsnapshotdto/download.go`
- Modify:
  - `hivepaas_app/base/audit.go` (`AuditLogTypeBackupDownload`);
  - `…/basesettinghandler/xtra_api_backup_snapshot.go`;
  - the four scopes' `backup_snapshot.go` handlers;
  - the four routers;
  - `docs/openapi/swagger.json`.
- Test: `…/backupsnapshotdto/download_test.go`, `…/backupsnapshotuc/download_test.go`, `…/basesettinghandler/backup_snapshot_restore_request_test.go`

**Interfaces:**
- Consumes: the restore plan's `ListEntries` and `RestoreStream`.
- Produces: `UC.DownloadBackupSnapshotFile(ctx, auth, *DownloadBackupSnapshotFileReq{Scope, ID, Path "path"}) (*DownloadBackupSnapshotFileResp{FileName, SizeBytes, Write func(ctx, io.Writer) error}, error)`.
- The handler asks `write` on the scope; the reach is taken for `write` too.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-04-tests.patch`

Tests added:
- `TestDownloadRequestNamesAFileInside`, `TestDownloadableEntry`, `TestDownloadAuditScope`;
- `TestBackupSnapshotDownloadRequestParses`, `TestDownloadHeaders`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/usecase/settings/backupsnapshotuc/... ./hivepaas_app/interface/api/handler/basesettinghandler/`
Expected: FAIL, build error `undefined: NewDownloadBackupSnapshotFileReq`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-04-code.patch`. Then run `make gen-swag`, which must leave `swagger.json` unchanged, and `go mod verify`.

- [ ] **Step 4: Check it**

Run: `go build ./... && go test ./... && golangci-lint run ./...`
Expected: PASS; `0 issues.`
- `grep -c backup-snapshots docs/openapi/swagger.json` gives `20`;
- `grep -c backup/files docs/openapi/swagger.json` gives `0`.

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(backup): download a file of a backup snapshot, at the four scopes

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 5: The configuration form, and no Backup Files (dashboard)

**Files** (under `$D/src/application/`):
- Delete:
  - `modules/settings/routes/data-backup/backup-files/`;
  - the system backup files' services, hook, queries, commands, domain, schema and table definitions;
  - the old compression, encryption and file enums.
- Modify:
  - the system backup's domain, schema and contract (`includeDB`, `includeSpec`, `specSecrets`, `specPassphrase`, `targetRepository`; `SYSTEM_BACKUP_SPEC_SECRETS`);
  - the configuration form, its schema, mappers and route;
  - the Data Backup layout, router, module and route constants;
  - the Data Cleanup form, schema, mappers, route, domain, schema and contract (no `backupCleanup`);
  - the audit types (`BackupDownload`).

- [ ] **Step 1: Apply the patch**

Run: `git -C $D apply --index $(pwd)/$P/dash-01.patch`

- [ ] **Step 2: Check it**

Run: `(cd $D && npx tsc --noEmit && npm run lint && npx prettier --check src)`
Expected: no type errors, no lint output, `All matched files use Prettier code style!`

- [ ] **Step 3: Commit (dashboard repo)**

```bash
git -C $D commit -m "feat(backup): the system backup's configuration takes the database, the spec, and a repository

The Backup Files page goes, with its data layer, and the system cleanup's
retention of backup files.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 6: Files with Download, and the checks (dashboard, then live)

**Files** (under `$D/src/application/modules/`):
- Create: `settings/module-shared/components/backup-snapshot-table/backup-snapshot-files.com.tsx`
- Modify:
  - the snapshot API, contracts and hook (`downloadFile`);
  - the details drawer (a `scope` prop, the Files row);
  - the table;
  - the helpers (*System*);
  - the task card: a system run's snapshot links to Settings › Backup Snapshots filtered on its run.

- [ ] **Step 1: Apply the patch**

Run: `git -C $D apply --index $(pwd)/$P/dash-02.patch`, then the checks of Task 5, Step 2.

- [ ] **Step 2: Commit (dashboard repo)**

```bash
git -C $D commit -m "feat(backup): download a snapshot's files from its details, and see a system backup's snapshots

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

- [ ] **Step 3: Live and browser checks** (the user, on the Linux server)
  - **Check 1:** configure a system backup of both into an S3 repository, run it, and again into a volume repository. Download `db.pg_dump` and run `pg_restore --list` on it (Review Focus 1 and 5).
  - **Check 2:** download `spec.tar.gz.age` and import it through the Import page with its passphrase (Review Focus 2).
  - **Check 3:** download a large file and watch the browser's memory (Review Focus 3).
  - **Check 4:** as a user without the reveal capability, save encrypted or plaintext: it is refused, and a reveal entry is recorded (Review Focus 4).
  - **Check 5:**
    - the form's checks: none chosen, no repository, no passphrase;
    - saving it disabled;
    - no Backup Files tab;
    - the Data Cleanup form without backup files;
    - a run's snapshot link.
