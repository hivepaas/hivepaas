# Managed MariaDB Phase 1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A managed MariaDB with what managed PostgreSQL had after its phase 1:
- continuous archiving;
- nightly base backups with retention;
- health and a restore test;
- restore into a new app to the latest, a time, a point or a backup;
- restore points, deploy points and a dump job.

**Architecture:**
- **First, a refactor with no change in behavior (Task 1).** The engine's state becomes
  engine-neutral (`Database`, `Position`, `PendingBytes`). Engines declare their
  capabilities.
- **Then MariaDB (Tasks 2-6).** Its scripts and harness come first in `app-templates`
  (Task 0). Then the engine, alongside `postgres`.

**Spec:** `docs/superpowers/specs/2026-10-19-managed-mariadb-design.md`. The spike is
`2026-10-19-mariadb-spike.md`.

## How this plan works

As in the Postgres phases:
- **No patches were written in advance.** Each task is written test-first when it is
  executed, and is done when its checks pass.
- **File references come from the code at this plan's commit.** The inventory of
  Postgres-specific code outside the engine was taken on 2026-10-19; Task 1 works from
  it. Where the code has moved since, follow the code.
- **The user approves each task before it starts.**

**Repositories:**
- this one (the backend);
- `../app-templates`, Tasks 0, 2, 4, 5, 6;
- `../hivepaas-dashboard`, Tasks 1, 2, 6;
- `../hivepaas-cli`, whenever the spec changes.

Work on a branch per repository, in a worktree.

**Global constraints:**
- **Secrets:**
  - S3 keys and the libsodium key reach the container as Swarm secrets, never as env
    and never on argv.
  - Logs name the step, not the value.
  - Database credentials reach `mariadb-backup` and the replay through a tmpfs option
    file.
- **Git:**
  - Commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.
  - Merge locally, never push.
  - Never `git stash`.
  - Never touch the user's uncommitted work.
- **Live checks:**
  - on the dashboard's e2e dind stack;
  - for Task 6, on the dev cluster. The user deploys it, and it is reseeded on each
    deploy.
- **The API level** stays 4 unless a request changes. Reset it after `make gen-swag`.

**Order:** 0 and 1 can go in either order, and both before 2. Then 2, 3, 4, 5. 6 is last.

## Task 0 (`app-templates`): the scripts, their harness and the generated template

**Files:**
- `scripts/walg-mariadb/`:
  - `template-head.yaml`;
  - `walg-lib`, `walg-run`, `walg-entrypoint`, `walg-archive`, `walg-replay`;
  - `walg-health`, `walg-nightly`, `walg-mark`, `walg-restore-test`;
  - `generate.sh` (with `-check`), as `scripts/walg/` has them.
- `scripts/test/walg-mariadb/run.sh`, for 11.8 and 11.4. It starts from the spike's
  restore script, with SeaweedFS and plain containers.
- `templates/mariadb-managed.yaml`, generated.
- CI: `generate.sh -check` for the new directory.

**The harness checks:**
- **Archiving:**
  - A write is pushed within the archive timeout plus a margin. The flush happens only
    when the log has events.
  - An idle database adds no binary log files.
  - Purged logs on the server are only ones already pushed.
- **Base backup:** `backup-push` through `walg-run`, encrypted. `xtrabackup_checkpoints`
  is warned about, and that is all.
- **Restores, each into a new volume through `walg-entrypoint`'s restore mode:**
  - the latest;
  - a time, rows after it absent;
  - a point by position;
  - a base backup with no replay.

  Also a base backup taken while a log rotates: the replay still starts at the right
  offset.
- **Retention:** `walg-nightly retain` keeps N, and logs before the oldest go.
- **The guard:**
  - an empty data directory over an archive with history idles, and health and
    `walg-health` say why;
  - an unreadable archive fails the start.
- **The restore test:** it passes, and fails on a check query that fails.
- **Credentials:** no secret appears in `/proc/*/cmdline` during a backup or a replay.
  The harness also answers open question 1: root on the socket against WAL-G's driver.

**Not yet:** `apptemplate index`. The backend's lint does not know engine `mariadb`
until Task 2.

- [ ] Commit (app-templates).

## Task 1 (backend, dashboard, CLI): engine-neutral state and capabilities

Postgres keeps working exactly as before: the Postgres harness, the e2e level-0 spec and
a dind live check are rerun.

**Backend:**
- **entity:**
  - `ManagedPostgres` becomes `ManagedDatabase`, and `AppManagedSettings.Postgres`
    becomes `.Database` (`json:"database"`). A legacy `json:"postgres,omitempty"` field
    is read only by `Migrate`.
  - `ManagedRestorePoint.LSN` and `ManagedRestoreTarget.LSN` become `Position`.
  - `ManagedPostgresJobs` becomes `ManagedJobs`.
  - `CurrentAppManagedVersion` is 2. `Migrate` moves `postgres` to `database`, and
    `lsn` to `position`.
- **The deref sites move from `.Postgres` to `.Database`**, about 40 in about 20 files.
  The inventory lists them:
  - appmanageduc;
  - managedserviceimpl (`lockManaged`, `WriterOf`, mark and prune);
  - entity (`GetRefObjectIDs`, `Decrypt`);
  - envvarserviceimpl (`clusterEnvVars`);
  - appsettingsuc (the deploy point check);
  - appuc (the delete guard);
  - taskschedjobexec (`writerOf`);
  - apptemplateuc (`managedOptions`).
- **`managedmodel`:**
  - `ArchivingStatus.PgWalBytes` becomes `PendingBytes`.
  - `Capabilities{Replicas, VerifierApp}` and `PointNameMax()` are added to `Engine`.
    Postgres answers true, true and 63.
- **Use cases:**
  - Replicas, promote and backup-from refuse an engine without `Replicas`, with
    `ERR_APP_MANAGED_UNSUPPORTED_BY_ENGINE`.
  - A verifier is refused without `VerifierApp`.
  - The deploy point check asks for any managed database, not a Postgres one.
- **`clusterEnvVars` and envlink:** only for an engine with `Replicas`.
- **DTOs:**
  - `lsn` becomes `position`, and `pgWalBytes` becomes `pendingBytes`.
  - The settings response gains `capabilities`.
  - The point-name limit comes from the engine.
  - Then gen-swag, at API level 4.
- **templatemodel:** `ManagedPostgresDefaults` becomes `ManagedDefaults`, with
  `Managed.Defaults()`.

**Dashboard:**
- One engine table: name, archive noun ("WAL" or "binary logs"), dump tool, example
  check query. It replaces `engineName`, `managedEngineName` and the hard-coded copy the
  inventory lists.
- `use-managed-databases` drops its `postgres` filter.
- The Replicas section and "Base backups taken on" show only with `capabilities.replicas`.
- The types follow: `position`, `pendingBytes`, `capabilities`.

**CLI:** the client is regenerated.

**Tests:**
- the migration, with a v1 row read and written as v2;
- capabilities refuse;
- the existing suites, renamed;
- e2e level-0 15/15.

- [ ] Commit per repository.

## Task 2 (backend, app-templates): the template contract and the engine's skeleton

**Files:**
- `templatemodel/managed.go`:
  - `ManagedEngineMariaDB`;
  - `ManagedEngines[mariadb] = {MinAPI 1, MaxAPI 1, Artifacts [wal-g]}`;
  - `Managed.MariaDB *ManagedDefaults` (`yaml:"mariadb"`);
  - validation: only for engine mariadb, and requires `v000003`.
- `base/version.go`: `VersionCodeV3`, and `CurrentVersion = VersionCodeV3`.
- `service/managedservice/mariadb/`:
  - the engine registered;
  - `Name`, `APIRange`, `Capabilities{false, true}`, `PointNameMax 63`;
  - every other method `ErrUnsupported` until its task.
- **app-templates:** `apptemplate index` with the Task 0 template.
- **Dashboard:** the store shows `mariadb-managed` with its engine name. The create form
  works as for Postgres, with the engine's copy.

**Tests:**
- the contract, including a v000002 template naming mariadb, which is refused;
- the registry against the table;
- the index lint passes.

- [ ] Commit per repository.

## Task 3 (backend): create, archive, nightly, health, status

**Files:** `mariadb/prepare.go`, `configure.go`, `status.go`, `retention.go`. Tests
mirror the Postgres engine's.

- **`PrepareDoc`:**
  - the archive's env (`WALG_S3_PREFIX`, the endpoint, path style, compression,
    `WALG_PREVENT_…` where it applies) and secrets, as files;
  - for a volume archive, the subpath mount and `WALG_FILE_PREFIX`;
  - the scripts as config files;
  - the locked command and entrypoint, from the spec, §3;
  - the archive timeout passed to `walg-archive`.
- **`Configure`:** the app-managed setting and the nightly, health and restore-test jobs.
- **`Status`:**
  - from `walg-health`'s outputs (`LAST_ARCHIVED_AT`, `PENDING_BYTES`, `DB_BYTES`,
    `FREE_BYTES`);
  - from `walg-nightly`'s (`BACKUPS`);
  - the restore window, from the oldest backup to the last log pushed.
- **`NightlyBackupArgs`, `Retention`:** as Postgres's, with WAL-G backup names.

**Live (dind):**
- create `mariadb-managed` 11.8 with S3;
- a write is archived within the timeout;
- run the nightly; health is green;
- the Database tab shows status and backups in MariaDB's words.

- [ ] Commit per repository.

## Task 4 (backend): restore into a new app

**Files:** `mariadb/restore.go`, with `RestoreSource` (the source's archive and target
as env and secrets for `walg-entrypoint`'s restore mode), `RecoveryCheck` and
`FinishRestore`. The restore task's steps are engine-neutral.

**Tests:** for each target (latest, time, backup), the restore env and how a source is
read.

**Live (dind):**
- each target restores into a new app with the rows it should have;
- the restored app's first nightly backs up into its own prefix;
- the switch moves a consumer to it.

- [ ] Commit per repository.

## Task 5 (backend, app-templates): restore points and deploy points

**Files:** `mariadb/mark.go`, with `MarkRestorePoint` (`walg-mark` waits until the
point's log is pushed) and `MarkResult` (the `position`). A restore to a point means
`walg-replay` stops at the position.

**Tests:** parsing the position, and the mark command.

**Live (dind):**
- `hivepaas db mark before-x`, then rows written, then a restore to `before-x`, which
  has the rows up to the mark;
- a deploy with "restore point before deploy" on a MariaDB marks one.

- [ ] Commit per repository.

## Task 6 (backend, app-templates, dashboard): restore test, verifier, dump, e2e and the cluster

- **Restore test:** `walg-restore-test` in the container, and in a verifier app (the
  Postgres verifier's spec, with MariaDB's scratch directory).
- **Dump:** the engine's data-backup job with `mariadb-dump`, as Task 3 of Postgres
  phase 2 made it.
- **e2e level-0:** a MariaDB's create form, Database tab and settings. There is no
  Replicas section, and the copy is MariaDB's.
- **Live:**
  - on dind, the restore test passes and fails as it should, and a dump is taken;
  - on the dev cluster, a MariaDB on a volume pinned to the worker, restored to the
    other node; and a MariaDB on NFS, moved.
- **Release:** `release.json` pins the app-templates commit, for the user to sign.

- [ ] Commit per repository.

## After the plan

- Phase 2: replicas, backup from a replica, promotion, and the cluster's stable address
  (`mariadb://`… or a list of hosts for the connectors that take one).
- arm64: run the aarch64 artifacts on an arm64 node, for both engines.
