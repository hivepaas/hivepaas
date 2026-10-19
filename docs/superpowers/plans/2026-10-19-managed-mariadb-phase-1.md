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

- [x] Commit (app-templates).

**As built** (app-templates, branch `feat/mariadb-managed-scripts`, worktree
`../app-templates-wt`, commit 2bf4f03; **not merged**. CI's lint of the templates and
its `index -check` would fail on `main` until Task 2 teaches the backend engine
`mariadb`, so it merges with Task 2):
- **Credentials, open question 1 answered:**
  - root@localhost in the official image has a password only, no unix_socket.
  - The scripts reach the server as `mysql@localhost`, `IDENTIFIED VIA unix_socket`,
    and run as the OS user `mysql` through `gosu`. This works for the `mariadb`
    client, `mariadb-backup`, and WAL-G's Go driver (DSN
    `mysql@unix(/run/mysqld/mysqld.sock)/mysql`). No password is used anywhere.
  - The account comes from the image's own `MARIADB_MYSQL_LOCALHOST_USER=1` and
    `MARIADB_MYSQL_LOCALHOST_GRANTS=ALL`, in the template's env, not an initdb file.
    A restored database brings the source's. Task 3's `PrepareDoc` locks both
    variables.
  - Replaying arbitrary events needs ALL: DDL, `SET gtid_*`, and the privilege to
    replay binary logs.
  - A verifier's restore test reaches the live database over the network as root. Its
    password goes into an option file in `/dev/shm` and is never on argv.
- **The data directory** is `/var/lib/mysql/data` (`--datadir`), on a volume mounted
  at `/var/lib/mysql`. The state files sit beside it:
  - `hivepaas-archive.state` and `.lock`;
  - `hivepaas-restore.state`, `.log` and `.backup`;
  - `hivepaas-refused`;
  - the restore test's scratch copy;
  - the replay's fetched logs (`hivepaas-replay`).

  `walg-entrypoint` chowns the mount point and the loop's state files to `mysql`, for
  volumes Docker did not copy the image's ownership into.
- **The locked command** the harness uses, and Task 2 locks:
  `mariadbd --datadir=/var/lib/mysql/data --log-bin=mariadb-bin --binlog-format=ROW
  --server-id=1 --log-slave-updates --binlog-expire-logs-seconds=172800`. The
  entrypoint is `walg-entrypoint`, and `HIVEPAAS_ARCHIVE_TIMEOUT` (seconds, default
  60) sets the loop's period.
- **`walg-archive`:**
  - **Flush rule.** A pass flushes when the active log's end is past the offset where
    the last flush left a new log. A log the server only rotated to, or opened at
    start, has just its header and stays open: an idle database adds no log.
  - **Purge.** It then pushes, and purges to the active log.
  - **Retries.** Its push runs with `WALG_S3_MAX_RETRIES=3`. With WAL-G's default,
    an unreachable store held a pass, and the lock, for many minutes. The next pass
    is the retry.
  - **Locking.** It holds the lock on an fd (`flock`), so a pass and `walg-mark`
    never push at the same time.
- **WAL-G's MySQL output, verified:**
  - `backup-list --detail --json` gives `backup_name`, `start_local_time`,
    `compressed_size` and `binlog_start`.
  - Backup names are `stream_YYYYMMDDTHHMMSSZ`.
  - `binlog-list --json` gives `name`, `size` and `last_modified`.
  - `binlog_start` is the last binary log *already uploaded* when the backup started,
    not the one it starts in. The backup starts in a later log, and
    `mariadb_backup_binlog_info` names it.
  - `delete before FIND_FULL <name>` deletes every binary log uploaded before the
    backup's name time, `binlog_start` included. The log the backup starts in was
    still open then, so it is kept. The harness restores the oldest backup kept after
    retention, to the latest.
- **`walg-mark NAME SECONDS`:**
  - It prints `POSITION=<file>:<offset>` and `GTID=<gtid_binlog_pos>` on two lines,
    not the design's one `POSITION=<file>:<offset>/<gtid>`. Task 3's `MarkResult`
    reads `POSITION`.
  - It flushes and runs `walg-archive once` (under `timeout`) until the point's log
    is listed. Otherwise it exits 75.
- **`walg-replay`:**
  - It fetches into a work directory on the volume, and deletes each log once it is
    replayed.
  - A stop position in a log the archive lacks fails before anything is replayed.
- **`walg-nightly retain`** checks the chain in place of `wal-verify`. The archived
  logs from the one after the oldest backup's `binlog_start` must be numbered one
  after another. A log removed from the middle fails it, naming the log.
- **`walg-health`:**
  - It fails when the last pass failed, when no pass has succeeded for three periods,
    or when none has in three periods of uptime.
  - `PENDING_BYTES` and `READY_COUNT` come from `SHOW BINARY LOGS`.
  - `REPLICAS` is empty.
- **`walg-restore-test`:**
  - It marks a point, fetches into `<scratch>/hivepaas-restore-test/data`, and starts
    a second `mariadbd` on `/run/mysqld/hivepaas-restore-test.sock`, with
    `--skip-networking --skip-log-bin --innodb-buffer-pool-size=64M`.
  - It replays to the point and runs the check query on `MARIADB_DATABASE`.
  - `RESTORED_TO` is the position.
- **`walg-entrypoint` restore mode:**
  - It fetches, starts the server on its socket alone, replays, stops it, and serves.
  - The targets are:
    - `HIVEPAAS_RESTORE_TARGET_BACKUP` (or `LATEST`), or
      `HIVEPAAS_RESTORE_TARGET_BEFORE`;
    - then `HIVEPAAS_RESTORE_STOP_TIME`, `HIVEPAAS_RESTORE_STOP_POSITION`, or
      `HIVEPAAS_RESTORE_REPLAY=none`.
  - A restore stopped before it ends fetches again.
  - `status` prints the last lines the scripts and WAL-G wrote, not the server's.
- **Known noise:** each push makes the server log "Aborted connection … Got an error
  reading communication packets". WAL-G's driver closes without `COM_QUIT`, so this
  shows once a minute in the container's log.
- **Harness:** `scripts/test/walg-mariadb/run.sh` passes 194 checks, 97 on each of 11.8 and
  11.4. This includes a base backup taken while 60 flushes rotate the logs: its
  restore holds all 61 rows. No key and no root password was seen in any process's
  arguments during a backup or a restore.

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

- [x] Commit per repository.

**As built** (backend e603e1d0, merged as 5758a987; dashboard 3a4734de, merged as
11c41d01; CLI 441a2b4, merged as 90799de; the API changed in responses only, at level 4):
- **Startup data migration never runs.** This was found here and is not fixed:
  - `data_migrations` is created with the row `v1-20251019`. That sorts after every
    `v00000N` code, so `migration.ID >= base.CurrentVersion` always skips.
  - Turning it on is a task of its own. Setting updates do not bump a row's
    `version`, so old steps would run again over newer data. For example, the
    deployment settings' version-3 step would turn auto-deploy back on where it was
    turned off.
- **So the app-managed setting does not rely on migration:**
  - `AppManagedSettings.UnmarshalJSON` reads the old `postgres` and `lsn` as `database`
    and `position`, and clears the legacy fields. `Migrate` still moves them, tested
    both ways.
  - Every writer stores through `Setting.SetAppManagedData`, which sets version 2, and
    updates the `version` column. A row converts at its next save.
  - On dind the other app-managed rows stayed version 1 and read correctly. A restored
    one that was never saved again read its `target.position`.
- **`VersionCodeV3` is added here** and is `CurrentVersion`, following `base/version.go`
  (a code per data version). Task 2 extends its comment to engine `mariadb`, and adds
  no code. Nothing is released between Task 1 and Task 2: a v000003 HivePaaS would be
  offered `mariadb-managed`.
- **Engine lookup** is `service/managedservice/engines` (`Lookup`), not the managed
  service. The environment-variable service needs it, and asking the service would be
  a dependency cycle: envvar → managed → schedjobexec → command → envvar.
- **`templatemodel.EngineContract.Title`** ("PostgreSQL") names the engine in
  `ERR_APP_MANAGED_UNSUPPORTED_BY_ENGINE`, e.g. "PostgreSQL databases do not have
  replicas yet". Task 2 adds "MariaDB".
- **Refusals:**
  - Replica add and remove, promote, and backup-from naming a replica need
    `Replicas`. An empty backup-from, meaning the database itself, is every engine's.
  - Moving the restore test to a verifier app needs `VerifierApp`.
  - The deploy restore-point check asks for any managed database. Its refusal reads
    "restore points are made on a managed database, not on a replica". Deploy point
    names are still cut to 63 bytes, which every engine takes.
- **The point-name limit** is checked against the engine's `PointNameMax` once the app
  is loaded. Postgres gets the same field errors at 64 characters as before.
- **Dashboard:**
  - `managed-engines.ts` holds the one engine table: name, archive noun and its
    number, dump tool, example check query. MariaDB's row is already there.
  - The verifier-app option is hidden without `capabilities.verifierApp`.
  - The unmanage dialog uses the engine's terms. For Postgres one bullet now reads "the
    archiving of its WAL".
- **CLI:** `db mark` prints the position, and its help says "the log".
- **Checked:**
  - backend lint, and tests in 277 packages;
  - dashboard tsc, lint, prettier and build;
  - e2e level-0 15/15;
  - CLI build, test and lint.
  - On dind (APP_PG18), the settings answered `capabilities`, and the status answered
    `pendingBytes` and an old point's `position` read from its `lsn`. A new point was
    stored at version 2 with no `postgres` or `lsn` key.
- **Not rerun:** the Postgres harness. Its scripts did not change.

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

- [x] Commit per repository.

**As built** (backend 953fffc3; app-templates aa288f4, on Task 0's branch, merged with
it; dashboard 049a992c; the API is unchanged, at level 4, so no CLI change):
- **Contract version codes.** `EngineContract.VersionCode` is the first code that has
  the engine: postgres v000002, mariadb v000003. A template naming an engine must
  require it ("managed: engine mariadb needs requires.versionCode v000003 or later"),
  besides v000002 for `managed` itself.
- **Defaults.** `Managed.MariaDB` (`yaml:"mariadb"`) sits beside `Postgres`.
  `Defaults()` picks the template's engine's. Each key is refused for another engine,
  and checked as `managed.<engine>.…`, in a stable order.
- **`VersionCodeV3`** came with Task 1. Its comment now also says that a managed
  template may name engine mariadb.
- **The engine skeleton** (`service/managedservice/mariadb`):
  - `Capabilities{Replicas: false, VerifierApp: true}` and `PointNameMax` 63, as
    planned.
  - Every method with an error refuses with `ERR_APP_MANAGED_UNSUPPORTED_BY_ENGINE`,
    worded as what a MariaDB database does not have yet: archiving, a status,
    settings, restores, nightly backups, restore points, replicas.
  - The others return their zero value. No app reaches them, since none can be made.
  - It is registered in `engines.Lookup`, and the test holds the registry to the
    table.
- **The index.** `apptemplate lint` passes on 393 templates and `index -check` on the
  new index. The v000002 subset keeps 392: `mariadb-managed` is left out of the
  releases that cannot read it.
- **Dashboard.** No code change, since Task 1 made the store and the create form read
  the engine table. A level-0 e2e test now covers a mocked `mariadb-managed`:
  - the badge's tooltip names MariaDB;
  - the form speaks of binary logs and `mariadb-dump`;
  - the check-query placeholder is MariaDB's;
  - the create request names the template.

  The spec passes 16/16.
- **Live on dind:**
  - The store lists `mariadb-managed` as managed by mariadb and compatible, with its
    defaults. `postgres-managed` keeps its own.
  - Creating one answers 412 "MariaDB databases do not have archiving yet", and leaves
    no app.
- **Shipping.** The template reaches installs only with a new release.json pin to an
  app-templates commit that has it, re-signed. That waits until the engine creates
  databases (Task 3) at least.

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

- [x] Commit per repository.

**As built** (backend e127afca and 49a3d1bc; the CLI's spec copy
re-pinned for descriptions only, 7be2a7a; the API stays at level 4; no dashboard
change):
- **`service/managedservice/walgkit`, a refactor first** (e127afca). The two engines
  share what engine API 1's scripts promise and what an engine makes of it:
  - the paths, secret names, defaults and job timeouts;
  - the outputs;
  - the archive's variables and secret files;
  - `Configure` (the setting and its four jobs);
  - `Status`, retention and kept points;
  - `Reconfigure` (schedules, retention, check query, verifier move, dump, base
    backups' member);
  - the dump job.

  An engine passes a `Profile` with what differs:
  - its scripts and their uid;
  - its backup-name pattern and delta marker;
  - the walg-health output for pending bytes;
  - its dump commands;
  - a members reader (Postgres's, nil for MariaDB).

  The Postgres engine keeps its server command, restore, mark, replicas and promotion.
  Its tests pass unchanged apart from names. On dind after the refactor, pg18s (with a
  replica) ran health and nightly, and its status read as before, the replica
  streaming. `walgkit` has no tests of its own: both engines' suites test it.
- **`PrepareDoc`:**
  - It sets `WALG_COMPRESSION_METHOD=zstd`, `WALG_LIBSODIUM_KEY_TRANSFORM=hex` and
    `HIVEPAAS_WALG_SECRETS_DIR`.
  - `HIVEPAAS_ARCHIVE_TIMEOUT` is the archive timeout in seconds, read by
    `walg-archive` and `walg-health`.
  - It sets the S3 or volume archive's variables. There is no
    `WALG_PREVENT_WAL_OVERWRITE`, which `binlog-push` lacks.
  - The secret files belong to uid 999, the images' mysql user, at mode 400. The
    archive loop runs as mysql.
  - `MARIADB_MYSQL_LOCALHOST_USER=1` and `…_GRANTS=ALL` are added when the template
    lacks them. Either way they are owned (locked), with every variable it adds.
  - A restore is refused until Task 4.
- **`Configure`:**
  - The command is `mariadbd --datadir=/var/lib/mysql/data --log-bin=mariadb-bin
    --binlog-format=ROW --server-id=1 --log-slave-updates
    --binlog-expire-logs-seconds=172800`. It adds `--innodb-buffer-pool-size` at a
    quarter of the memory limit (at least 128M), as Postgres's shared buffers; none
    without a limit.
  - The entrypoint is `walg-entrypoint`.
- **The dump** came here rather than in Task 6. The create form dumps into the
  archive's repository by default, so a MariaDB `Configure` without it would drop that
  choice.
  - It runs `gosu mysql mariadb-dump --user=mysql --single-transaction --routines
    --events --triggers --hex-blob "$MARIADB_DATABASE"` into `<key>.sql`.
  - Restore runs `gosu mysql mariadb --user=mysql "$MARIADB_DATABASE"`. It is not one
    transaction, since MariaDB commits DDL, and it needs `mysql@localhost`: this
    database, or another managed one.
- **`Reconfigure`** is walgkit's. Moving the restore test into a verifier is refused
  ("MariaDB databases do not have verifier apps yet") until Task 6. The dashboard
  still offers the option, since the capability is true.
- **Still refused:**
  - restore (`PrepareDoc` with a restore, and `FinishRestore`);
  - restore points (`MarkRestorePoint`, `MarkResult`), so a deploy asking for a point
    on a MariaDB stops before it changes anything;
  - replicas and promotion.
- **DTO comments** now name both engines (the dump; `pendingBytes`). gen-swag changed
  descriptions only, at level 4.
- **Live on dind** (`lc/check19.sh`), `lc-maria` 11.8 archiving to lc-s3 and dumping
  to lc-dumps:
  - It started through `walg-entrypoint` with the locked command and the archive
    loop, its data in `/var/lib/mysql/data`.
  - A write's log was archived 58s later, unasked.
  - Nightly, health, dump and restore test all finished `done`.
  - Status:
    - archiving ok, 389 B pending;
    - the `stream_…` backup;
    - a restore window;
    - the dump's snapshot;
    - the test restored to `mariadb-bin.000003:737`.
  - A second nightly kept 2 backups. A settings PUT (retention, schedule) saved.
  - The Database tab reads "MariaDB, archived continuously under …" and "Binary logs on
    the node", with no Replicas section.

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

- [x] Commit per repository.

**As built** (backend b9dcaeb9; app-templates 286f507; API unchanged, at level 4; no CLI
or dashboard change):
- **walgkit has the shared restore** (`walgkit/restore.go`):
  - `PrepareRestore` sets the source archive's `HIVEPAAS_RESTORE_`-prefixed variables
    and its keys as `RESTORE_` secret files of the engine's user. A volume source is
    mounted read-only.
  - `RecoveryCheck` runs `walg-entrypoint status`.
  - `FirstBackupJob` is the nightly job.
  - `FinishRestore` drops the restore's variables, secrets, their mounts and the
    source mount, and marks the restore ready.

  `Profile.RestoreTarget` is the engine's part. Postgres keeps its
  `HIVEPAAS_RESTORE_RECOVERY` settings. Its tests pass, renamed.
- **MariaDB's targets:**

  | Target | Variables |
  |---|---|
  | latest | `TARGET_BACKUP=LATEST` |
  | time | `TARGET_BEFORE` and `STOP_TIME`, both RFC 3339 |
  | point | `TARGET_BEFORE=<point at>` and `STOP_POSITION=<file:offset>`, matched against `^[A-Za-z0-9_-]+\.[0-9]{6,}:[0-9]+$` |
  | backup | `TARGET_BACKUP=<stream_…>` and `REPLAY=none` |

  Anything else is `ERR_APP_MANAGED_RESTORE_TARGET_INVALID`, a Postgres LSN or base
  name included.
- **The restore task** is unchanged apart from its wording: "Recovered: the server is
  up, and serving".
- **Found here, fixed for MariaDB, open for Postgres:**
  - **Two servers on one volume.** The restore's finish updates the service twice
    within a second: the variables taken off, then the secret mounts refreshed. Swarm
    started the third task before the first had stopped, so two `mariadbd` ran on one
    data directory for about 0.7 s (`ddl_recovery.log not found` on the new one). A
    plain `service update --force` orders them correctly.
    - `walg-entrypoint` now takes an `flock` on `<volume>/hivepaas-server.lock`
      before anything else. It holds it in a process of its own
      (`hivepaas-data-lock`), because `mariadbd` closes the files it inherits. A
      second container waits up to `HIVEPAAS_DATA_LOCK_TIMEOUT` (300 s), then exits
      1.
    - Postgres was hit too: `pg18b-restored` logged "performing immediate shutdown
      because data directory lock file is invalid". Its own check caught it a minute
      later. Its `walg-entrypoint` has no such lock yet.
  - **A restore's state on a reused volume.** `hivepaas-restore.state` now records an
    id: a hash of the `HIVEPAAS_RESTORE_` variables and the app's own archive prefix.
    A volume restored before, into another app or to another target, is fetched again
    and not served as it was. Postgres's `walg-entrypoint` serves a "done" state or an
    existing `PG_VERSION` as it is.
- **Harness:**
  - 108 checks on 11.8: a second server waits or gives up, and the lock is held by its
    own process; latest restored into a volume restored to a time before holds the 40
    rows.
  - 11.4 passed the same, bar the lock-holder check, whose `pgrep` was fixed after.
- **Live on dind** (`lc/check20.sh`):
  - `lc-maria` held 10 rows, a base backup, then 20 rows, a moment, then 30 rows. It
    was restored into new apps:
    - latest: 30 rows;
    - the time: 20 rows;
    - the backup: 10 rows.
  - Each finished its task, served and took writes. Each had its first base backup
    in its own `walg/<key>/<id>`, and no `HIVEPAAS_RESTORE_` variable left on the
    service.
  - `lc-web`, with `DB_HOST=${lc-maria.HIVEPAAS_HOST}`, was listed as a consumer and
    switched to `${lc-maria-latest-0003.HIVEPAAS_HOST}`. `lc-maria` is switched, its
    jobs paused. `lc-maria-latest-0003` is the live MariaDB from here on.

## Task 5 (backend, app-templates): restore points and deploy points

**Files:** `mariadb/mark.go`, with `MarkRestorePoint` (`walg-mark` waits until the
point's log is pushed) and `MarkResult` (the `position`). A restore to a point means
`walg-replay` stops at the position.

**Tests:** parsing the position, and the mark command.

**Live (dind):**
- `hivepaas db mark before-x`, then rows written, then a restore to `before-x`, which
  has the rows up to the mark;
- a deploy with "restore point before deploy" on a MariaDB marks one.

- [x] Commit per repository.

**As built** (backend 5bb1208b; no script change, Task 0's `walg-mark` as it was; API
unchanged, at level 4):
- **`walgkit/mark.go`** is shared:
  - the point name's rules (`ValidPointName`, the reserved `hp-` prefix,
    `PointNameMax` 63);
  - `walg-mark NAME SECONDS` with a wait of two archive timeouts;
  - `MarkResult(stdout, key, pattern)`.

  Postgres reads `LSN`, MariaDB reads `POSITION`. The latter is checked against the
  restore target's pattern, so what is stored is what a restore takes.
- **Live on dind** (`lc/check21.sh`, on `lc-maria-latest-0003`):
  - A point was marked at `mariadb-bin.000007:971`, between 10 rows and 10 more. It is
    listed, restorable until the run that drops its backup.
  - Restored to the point, the new app holds the 10 rows.
  - With `restorePointBeforeDeploy` naming the MariaDB, each deploy of `lc-web`
    marked `deploy-lc-web-<time>` at the end of a fresh binary log.
- **Not run:** the CLI (`hivepaas db mark`). It posts the same endpoint, and its output
  names the position.

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
