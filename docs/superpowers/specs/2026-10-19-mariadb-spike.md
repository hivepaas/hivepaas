# Managed MariaDB: spike

Date: 2026-10-19. The question: can a managed MariaDB be built the way the managed
PostgreSQL is, with no image of our own? That means the official `mariadb` image, plus
WAL-G downloaded by its sha256 and bind-mounted, plus scripts shipped in the template.

The answer is yes. Two things need scripts of our own, because WAL-G does not do them
for MariaDB:
- archiving on a timer;
- replaying the binary logs from the base backup's position.

Everything here was run with plain containers on one host's Docker, against SeaweedFS
as S3. It used MariaDB 10.11.19, 11.4 and 11.8.9.

## What was checked

| Check | Result |
|---|---|
| WAL-G for MySQL, v3.0.9 | `wal-g-mysql-{20.04,22.04,24.04}-{amd64,aarch64}`, one static binary each. The 20.04 amd64 build (sha256 `0fef8a61…3721`) runs in the 10.11 (Ubuntu 22.04) and 11.x (24.04) images with no missing library. |
| Tools in the official images | `mariadb-backup`, `mbstream` and `mariadb-binlog` are all present in 10.11, 11.4 and 11.8. No image of our own is needed. |
| Base backup | `wal-g backup-push` with `WALG_STREAM_CREATE_COMMAND="mariadb-backup --backup --stream=xbstream …"`. 55 MB came to 1.6 MB with zstd, encrypted with libsodium like the Postgres archive. |
| Restore of a base backup | `wal-g backup-fetch LATEST`, with `WALG_STREAM_RESTORE_COMMAND="mbstream -x -C /var/lib/mysql"` and `WALG_MYSQL_BACKUP_PREPARE_COMMAND="mariadb-backup --prepare --target-dir=/var/lib/mysql"`, then `chown mysql`. The server starts on it as it is. |
| Archiving the binary logs | `wal-g binlog-push` uploads the closed binary logs and leaves the one being written. |
| Restore to a time | 20 of 30 rows: the rows written after the target were left out. Run on 11.8. |
| Restore to the latest | 50 of 50 on 11.8, 25 of 25 on 11.4 and 10.11. |
| Retention | `wal-g delete retain FULL 1 --confirm` deletes the older base backups and the binary logs before the oldest kept one. |
| No secret on argv | `mariadb-backup --defaults-extra-file=/dev/shm/…cnf` reads the user and password from a file. |
| arm64 | The aarch64 builds exist, but were not run: this host has no arm64 emulation. The same is still true of the Postgres engine's arm64 artifact. |

## What WAL-G does not do for MariaDB

**Archiving on a timer.**
- MariaDB has no `archive_command`, and rotates a binary log only at `max_binlog_size`.
- `binlog-push` uploads closed logs only. Writes therefore reach the archive when the
  log rotates, which on a quiet database may be never.
- The entrypoint has to run a loop beside the server. Every archive timeout, it runs
  `FLUSH BINARY LOGS` if the active log has events past its header, then
  `binlog-push`. This is the equivalent of `archive_timeout` and `archive_command`, and
  the RPO is its period.

**Replaying from the base backup's position.**
- `wal-g binlog-replay` picks binary logs by upload time and replays every file whole.
- With MariaDB, the first file holds events the base backup already has, so a replay
  fails on duplicate keys or applies events twice. Here it fetched `mariadb-bin.000001`
  and `.000002`, while the backup had stopped at `.000002:991`.
- The replay is therefore ours:
  1. `wal-g binlog-fetch --since LATEST --until T` brings the logs, including the one
     T falls in.
  2. The start is read from the prepared data directory: `mariadb_backup_binlog_info`
     on 11.x, `xtrabackup_binlog_info` on 10.11.
  3. Each log from the start file on is piped to `mariadb`: the first with
     `mariadb-binlog --start-position`, all of them with `--stop-datetime` for a
     time target.
- `backup-push` warns that it cannot read `xtrabackup_checkpoints`, which MariaDB names
  `mariadb_backup_checkpoints`. Only incremental backups need it, and we take full ones.

## What this means for the engine

| Postgres engine | MariaDB engine |
|---|---|
| `archive_command`, `archive_timeout` | A loop in `walg-entrypoint`: flush when there are events, then `binlog-push`. Its last success is what health reads. |
| `walg-health` reads `pg_stat_archiver` | Read the last log pushed and the active log (`SHOW BINARY LOGS`), and the loop's last success. |
| Recovery settings and `recovery.signal` | Fetch and prepare, start the server, replay to the target, then serve. |
| Restore points by LSN (`pg_create_restore_point`) | Record `@@gtid_binlog_pos`, the binary log file and position, and the time at the mark. A restore to it uses `--stop-position` in that file. |
| Password from Swarm secrets | The same. `walg-run` builds `WALG_MYSQL_DATASOURCE_NAME` from the secret files for WAL-G's own process. Backup and replay read a tmpfs option file. The user is a backup user with the privileges `mariadb-backup` needs, not root. |
| Server-side WAL recycling | `PURGE BINARY LOGS TO` only for logs already pushed. `binlog_expire_logs_seconds` stays generous enough that the server never drops one first. |
| Dump: `pg_dump` | `mariadb-dump --single-transaction --routines --events`. |
| Replicas and promotion (phase 2) | GTID replication (`MASTER_USE_GTID=slave_pos`), seeded from the archive. Not looked at here. |

The server is started with `--log-bin`, `--binlog-format=ROW`, a `--server-id` and
`--log-slave-updates`. These go into the engine's locked command, as Postgres's
archive settings do.

## Proposed scope of MariaDB phase 1

**Versions:** the template offers what the store's `mariadb` template does, 11.8 and
11.4. 10.11 works the same, with the older name of the binlog info file.

**Built:**
- `mariadb-managed`, with the scripts as `postgres-managed` has them: entrypoint with
  the archive loop and the fresh-start guard, run, health, nightly (full backup and
  retention), restore test, mark and replay;
- the MariaDB engine (`managedmodel.Engine`), with prepare, configure, status, restore
  and restore points by GTID position;
- a dump job with `mariadb-dump`;
- the dashboard's Database tab made engine-aware.

**Reused:**
- artifacts and the cache, the license feature, locks;
- the managed jobs framework;
- restore into a new app, the verifier;
- unpinned volumes, artifact scanning and the switch.

**Phase 2:** replicas and promotion.

**Before building:** a harness like `scripts/test/walg/run.sh` for MariaDB, starting from
this spike's restore script. The parts that need most care are the archive loop's
RPO, under writes and while idle, and the replay's start position across a log
rotation during the base backup.
