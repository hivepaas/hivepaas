# Managed templates: MariaDB with point-in-time recovery

Date: 2026-10-19. Status: design, for the user's approval.

## Why

Managed PostgreSQL is built: phases 1 and 2 of
`2026-10-17-managed-templates-design.md`. MariaDB is the next engine the user asked
for. The spike (`2026-10-19-mariadb-spike.md`) shows it fits the same shape:
- the official `mariadb` image;
- WAL-G downloaded by its sha256 and bind-mounted;
- the scripts shipped in the template;
- the engine as code in HivePaaS.

MySQL is out of scope. Its image has no XtraBackup, and XtraBackup is neither one static
binary nor free of version pinning, so it breaks "no image of our own".

## Goals of phase 1

A MariaDB the user creates from `mariadb-managed` gets what a managed PostgreSQL got in
phase 1:
- continuous archiving;
- a nightly base backup with retention;
- a health check and a restore test;
- restore into a new app, to the latest, a time, a named restore point or a base backup;
- a dump job and a restore point before each deploy.

The rest of the managed machinery is used as it is:
- artifacts, the license feature, locked fields;
- restore into a new app, the verifier;
- unpinned volumes, artifact scanning, the switch, deleting a cluster.

**Not in phase 1:** replicas, promotion and the cluster's stable address. They are
phase 2, on GTID replication.

## The shape

The engine-neutral parts stay where they are. Two changes come first, so that a
second engine is added rather than grafted on:

1. **The engine's state stops being called Postgres** (§1). `AppManagedSettings.Postgres`
   becomes `.Database`, of type `ManagedDatabase`. The few engine-specific fields
   become neutral:
   - `LSN` becomes `Position`, a string the engine defines;
   - `PgWalBytes` becomes `PendingBytes`.

   Nothing has released the managed templates yet, but a migration moves stored
   settings anyway: dev servers and test stacks have some.
2. **An engine says what it can do** (§2). The dashboard and the use cases ask instead
   of assuming. MariaDB in phase 1 has no replicas.

Then the MariaDB engine, its template and its scripts (§3-§5).

## 1. Engine-neutral state

| Today | Becomes | Why |
|---|---|---|
| `AppManagedSettings.Postgres *ManagedPostgres` (`json:"postgres"`) | `.Database *ManagedDatabase` (`json:"database"`) | Every field under it is the archive's, the jobs', the restore's or the members': none is Postgres's. |
| `ManagedRestorePoint.LSN`, `ManagedRestoreTarget.LSN` (`lsn`) | `Position` (`position`) | An LSN for Postgres; for MariaDB, binlog file:offset and the GTID position. The engine reads what it wrote. |
| `ManagedArchivingResp.PgWalBytes` (`pgWalBytes`) | `PendingBytes` (`pendingBytes`) | What waits on the node to be archived: `pg_wal`, or binary logs not yet pushed. |
| `ManagedPostgresJobs` | `ManagedJobs` | The same four roles: nightly, health, restore test, dump. |
| `templatemodel.ManagedPostgresDefaults` | `ManagedDefaults`, read from `managed.postgres` or `managed.mariadb` | The same four defaults. The YAML key stays the engine's, so `postgres-managed` is unchanged. |

**Migration of stored settings.** `CurrentAppManagedVersion` moves from 1 to 2.
- A legacy field `json:"postgres,omitempty"` stays on the struct, to be read during
  migration only.
- `Migrate` moves `postgres` to `database`, and `lsn` to `position` in its restore
  points and restore target.

**The wire.** The changes are in responses only: `lsn` becomes `position`, `pgWalBytes`
becomes `pendingBytes`. The API level does not move, and the CLI client is regenerated.

**Kept Postgres-only, behind the engine:**
- the libpq cluster URLs (`HIVEPAAS_DATABASE_URL_RW/RO`) and the envlink cluster group:
  they are built only for an engine that has replicas;
- the 63-byte point-name limit: it becomes the engine's (`Engine.PointNameMax()`).
  MariaDB has no limit of its own, and keeps 63 so that names read the same.

## 2. What an engine can do

```go
type Capabilities struct {
	Replicas   bool // add/remove replicas, back up from one, promote; the cluster URLs
	VerifierApp bool // the restore test in an app of its own
}
func (Engine) Capabilities() Capabilities
```

- **Use cases** refuse what the engine lacks, with
  `ERR_APP_MANAGED_UNSUPPORTED_BY_ENGINE` ("{{.Engine}} databases do not have
  {{.What}} yet"). This covers the replica, promote and backup-from endpoints.
- **The settings response** carries `capabilities`. The dashboard hides the Replicas
  section, and the "Base backups taken on" select, for an engine without them.
- **The dashboard names engines from one table**: `postgres` is PostgreSQL, `mariadb`
  is MariaDB. Engine-specific copy (WAL or binary logs, `pg_dump` or `mariadb-dump`, the
  restore test's example query) comes from the same table. The "restore point before
  deploy" select offers every managed database, not only `postgres` ones.

## 3. The engine

`service/managedservice/mariadb` implements `managedmodel.Engine`, alongside
`postgres`. The same methods apply, and these are the differences:

| Method | MariaDB |
|---|---|
| `PrepareDoc` | It sets the following, as Postgres's does: the archive's settings and its secrets (S3 keys and the libsodium key, as files); the scripts' config files; the locked command `mariadbd --log-bin=mariadb-bin --binlog-format=ROW --server-id=1 --log-slave-updates --binlog-expire-logs-seconds=<2 days>`, the user's arguments after it; and the entrypoint `walg-entrypoint`. |
| `Configure` | The same jobs as Postgres: `managed-nightly` (walg-nightly), health (walg-health, every 5 minutes) and the restore test (walg-restore-test, in the container). |
| `Status` | Read from the outputs of the health and nightly jobs, which use the same keys as Postgres's, `PENDING_BYTES` in place of `PG_WAL_BYTES`. The restore window runs from the oldest base backup to the last binary log pushed. |
| `MarkRestorePoint` / `MarkResult` | `walg-mark` reads `@@gtid_binlog_pos`, the binary log's file and offset, and the time, then flushes and pushes the log. It prints `POSITION=<file>:<offset>/<gtid>` once that log is archived. |
| `RecoveryCheck`, `FinishRestore` | The same contract: up and serving once the replay is done. |
| `Retention` | `wal-g delete retain FULL n`. The binary logs before the oldest kept backup go with it, as the spike showed. |
| `Replica`, `Promote` | `ErrUnsupported`. The capabilities say so first. |

**Dump.** The engine's data-backup job runs
`mariadb-dump --single-transaction --routines --events --triggers <db>` into a Kopia
repository, as Postgres's `pg_dump` job does.

**Credentials.**
- The scripts run as root inside the container, and reach the server on its socket.
- The harness will show whether `unix_socket` lets WAL-G's Go MySQL driver in with no
  password. If it does not, a backup user is made at the first start, with the
  privileges `mariadb-backup` needs, and its password reaches the scripts as a Swarm
  secret.
- In both cases, `mariadb-backup` and the replay read a tmpfs option file. No password
  is ever on argv or in WAL-G's command-line arguments.

## 4. The template

`templates/mariadb-managed.yaml` is generated from `scripts/walg-mariadb/` by its
`generate.sh`, as `postgres-managed` is.
- **Versions:** 11.8 and 11.4, the store's `mariadb` template's
  (`mariadb:11.8.9-noble`, `mariadb:11.4.13-noble`).
- **Artifact:** WAL-G v3.0.9 `wal-g-mysql-20.04-{amd64,aarch64}`, by sha256.
- **Managed block:** `managed: {engine: mariadb, engineApi: 1, mariadb: {archiveTimeout: 60s,
  baseBackupSchedule, retainFull: 7, restoreTestSchedule}}`.
- **Version code:** `requires.versionCode: v000003`. A v000002 HivePaaS decodes
  templates strictly, and would refuse `managed.mariadb`. `CurrentVersion` becomes
  `v000003`: "a managed template may name engine mariadb".
- **Parameters, routing and kind settings** are the `mariadb` template's: database,
  user, password, root password, port 3306, `category: database, engine: mariadb`.

## 5. The scripts

| Script | What it does |
|---|---|
| `walg-lib` | Shared helpers: `sql` (`mariadb` on the socket), `data_root`, the refusal file, `out`. |
| `walg-run` | WAL-G with the archive's keys exported for its process alone, the DSN built from the socket, and the stream commands set. |
| `walg-entrypoint` | Normal start: guard against a fresh start over an archive with history (as Postgres's), start the archive loop beside the server, hand over to `docker-entrypoint.sh`. Restore: fetch, prepare, start the server on the socket only, replay to the target, then serve; the state is kept beside the data, as Postgres's restore keeps it. Also `status` and `health`. |
| `walg-archive` | The loop. Every archive timeout, if the active binary log has events past its header: `FLUSH BINARY LOGS`, `binlog-push`, `PURGE BINARY LOGS TO` the last one pushed. It records the time of its last success, which health reads. |
| `walg-replay` | `binlog-fetch --since LATEST --until T`. Then `mariadb-binlog`: from the base backup's start position (`mariadb_backup_binlog_info`, or `xtrabackup_binlog_info` on 10.11), `--stop-datetime` for a time, `--stop-position` in the point's file for a point. Piped into `mariadb`. |
| `walg-health` | Whether archiving keeps up: the loop's last success against three timeouts, and the pending binary logs. Prints the same JSON and outputs as Postgres's. |
| `walg-nightly` | `backup` (`backup-push`) and `retain` (`delete retain FULL n`), with Postgres's outputs. |
| `walg-mark` | The restore point (§3). |
| `walg-restore-test` | Fetches the latest base backup into a scratch directory and replays to the latest. A second `mariadbd` on that directory, on a socket of its own, answers the check query; then it is stopped and removed. |

**The harness:** `scripts/test/walg-mariadb/run.sh`, for 11.8 and 11.4, as
`scripts/test/walg/run.sh` is for Postgres. It covers:
- archiving, under writes and while idle (the RPO);
- the replay's start position when a log rotates during the base backup;
- restores to the latest, a time, a point and a backup;
- retention;
- the guard;
- the restore test;
- secrets on no argv.

## 6. Testing

- **Unit:** the neutral state's migration, capabilities, and the MariaDB engine's
  `PrepareDoc`, `Configure`, `Status`, mark and retention, as the Postgres engine's
  are tested.
- **Harness:** §5.
- **e2e level-0:** the create form, the Database tab and the settings for a MariaDB
  (engine copy, no Replicas section), with the API mocked as for Postgres.
- **Live:**
  - on dind: create, archive, nightly, mark, restore to a point and to a time,
    restore test, dump;
  - on the dev cluster: the same across two nodes;
  - the Postgres live checks rerun after the refactor of §1.

## Decided

- MariaDB first; MySQL is not planned.
- Versions 11.8 and 11.4.
- State renamed to `Database`, with a migration.
- Capabilities, not engine names, decide what is offered.
- The archive loop and the replay are our scripts; WAL-G's `binlog-replay` is not used.
- Replicas and promotion are phase 2.

## Open questions

1. **Root on the socket, or a backup user?** The harness decides. Root is simpler if
   WAL-G's driver gets in by `unix_socket`.
2. **An idle database's archive loop.** It flushes only when there are events, so an
   idle database adds no files. Should an idle day still push one log, so that health
   can tell "idle" from "stuck"? The proposal is no: health reads the loop's last
   successful pass, not the last log.
