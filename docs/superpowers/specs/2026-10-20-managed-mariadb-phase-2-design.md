# Managed MariaDB phase 2: replicas, promotion and the cluster's address

Date: 2026-10-20. Based on [the spike](2026-10-20-mariadb-phase-2-spike.md). It brings
to MariaDB what Postgres has had since its phase 2:
- read replicas, added and removed;
- base backups taken on a replica;
- promotion of a replica over its fenced primary;
- a stable address for the cluster's writer.

The user chose the address form: a Swarm network alias that follows the writer
(2026-10-20).

## 1. What changes in the scripts first

These change how phase 1 databases restore. They come before any replica, and they
keep restoring the archives phase 1 wrote.

- **The replay goes by GTID, file by file.**
  - `walg-replay` replays the logs in upload order (`binlogs_order`). Each file is
    replayed from the GTID state reached so far, starting at the backup's GTID in
    `mariadb_backup_binlog_info`.
  - The `mariadb-binlog` answer per file decides:

    | Answer | What the replay does |
    |---|---|
    | succeeds | applies the file, and moves the state |
    | "never reached expected GTID state" | skips the file: all of it is applied |
    | "missing data for domain" | fails, naming the file: a gap |

  - The state is a GTID list, so there can be several domains.
  - A time stop stays `--stop-datetime`.
  - A point stops at its GTID, `--stop-position=<gtid>`, which includes it.
  - Points made in phase 1 carry file:offset. They keep working: the file and offset
    of the primary's own logs are still a stop.
- **Each server's binary logs have a name of their own.**
  - The primary keeps `mariadb-bin`, so phase 1 databases do not change.
  - A replica writes `mariadb-bin-<member key>`.
  - The engine puts the basename in the member's locked command. A promoted replica
    keeps its name.
- **The chain check goes by GTID.** `walg-nightly retain` reads each archived log's
  starting GTID list and its end. It fails where a file starts past what the previous
  ones reached. Name numbers no longer matter.
- **`walg-mark`** prints `GTID=`, as it does already. The engine records it as the
  point's position, and `POSITION=` for display only.

## 2. Replicas

A replica is a logical child app with its own app-managed setting (`MemberOf`), as for
Postgres. Its scripts take what the engine gives:

| Variable | Meaning |
|---|---|
| `HIVEPAAS_STANDBY_HOST` | the writer's key |
| `HIVEPAAS_STANDBY_NAME` | the member key, its `--report-host` |
| `HIVEPAAS_STANDBY_PASSWORD_FILE` | the replication password, a secret file |

Its locked command adds:
- `--server-id=<n>`: 1 for the anchor; each member takes the next free number,
  recorded on the member;
- `--log-bin=mariadb-bin-<key>`;
- `--read-only`.

### `walg-entrypoint`, standby mode

1. **Seed.** An empty data directory is seeded from the archive:
   - the latest base backup is fetched;
   - the server starts on its socket alone;
   - `walg-replay` brings it to the last archived log;
   - `gtid_slave_pos` is set to the state reached.

   A replica it cannot seed idles, and never starts a database of its own.
2. **Connect.** At every start:
   - `CHANGE MASTER TO MASTER_HOST=<writer>, MASTER_USER=hivepaas_replica,
     MASTER_PASSWORD=<file>, MASTER_USE_GTID=slave_pos`, then `START SLAVE`;
   - local administration runs with `sql_log_bin=0`, so no errant GTID reaches a
     promotion.

   The password goes in by a SQL file read from the secret, never on argv.
3. **Serve,** with the archive loop in replica mode:
   - **It never pushes.** It purges its own logs older than a margin (two archive
     timeouts, at least an hour). That keeps what a promotion may need to give the
     archive.
   - **It catches up.** When the IO thread stops with error 1236 (the writer purged
     what it needs), it stops replication, replays the newly archived logs from
     `gtid_slave_pos`, and starts again.

### The other scripts

- **`walg-replication`** (on the writer, password on stdin):
  - creates `hivepaas_replica@'%'` with `REPLICATION SLAVE`;
  - sets its password;
  - runs again as it is.

  It runs as `mysql@localhost`, which has `GRANT OPTION` (the phase 1 fix).
- **`walg-health` on a replica** prints:
  - `REPLICA_STATE` (streaming, connecting or catching up), `LAG_SECONDS`
    (`Seconds_Behind_Master`) and `GTID` (applied);
  - none of a primary's outputs.

  On the writer, `REPLICAS` lists `SHOW SLAVE HOSTS` by report host. The engine joins
  the two for each member's status.
- **Backups taken on a replica:**
  - the nightly's backup step runs on the replica, as for Postgres (`BackupFrom`);
  - retention runs on the writer;
  - the backup restores by its GTID (§1), whatever file it names.

## 3. Promotion

The engine fences the writer first: the `HIVEPAAS_FENCED` variable, and walg-entrypoint
idles, as for Postgres. Then it runs `walg-promote` on the member.

`walg-promote` does this:
1. `STOP SLAVE; RESET SLAVE ALL; SET GLOBAL read_only=0`.
2. It finds the archive's last GTID: the end of the last log the archive lists.
3. It closes its own log (`FLUSH BINARY LOGS`). Then it pushes its own logs whose end
   (`BINLOG_GTID_POS(file, size)`) is past that GTID: the writer's last transactions,
   which it may never have archived. Their overlap with the archive is what the replay
   skips.
4. It leaves the promoted marker, so that the next start is a primary's whatever the
   variables say. Then the archive loop pushes as a primary's does.

Exit 3 means it certainly promoted nothing, as for Postgres.

The other replicas are restarted with the promoted member as their host. One that
lagged catches up from the archive.

The writer's state (`Writer`, `FencedAt`) and moving the jobs are the engine's, as
for Postgres.

## 4. The cluster's address

- **The alias.** The writer's service carries the network alias `<anchor key>-rw` on
  the project's network.
  - The anchor has it from the start: a database alone is its own writer.
  - At promotion the engine moves the alias:
    - it takes it off the fenced anchor's service, which is idle;
    - it puts it on the promoted member's, whose service restarts once.
  - The alias is an app network setting of the writer, which the engine owns and
    locks.
- **System variables** of the cluster's app:
  - `HIVEPAAS_DATABASE_HOST_RW` is the alias;
  - `HIVEPAAS_DATABASE_HOSTS_RO` lists the replicas' keys, comma-separated.

  `HIVEPAAS_DATABASE_URL_RW` and `_RO` stay Postgres's, which has a multi-host form.
  The envlink cluster group offers the two hosts.
- **Clients** connect to the alias and are moved by DNS. A connection open at
  promotion breaks, and the client's reconnection finds the new writer.

## 5. Capabilities and the dashboard

- MariaDB's capabilities become `{Replicas: true, VerifierApp: true}`.
- The dashboard's Replicas section, "Base backups taken on", promotion and the
  cluster's Danger zone appear for it as they do for Postgres.
- The engine table gains each engine's cluster-address copy: the multi-host URL for
  Postgres, the `-rw` alias for MariaDB.

## 6. Testing

- **Harness** (`scripts/test/walg-mariadb/run.sh`):
  - restores of phase 1 archives still pass;
  - two replicas seeded under writes, one late (catch-up after 1236);
  - a backup on a replica, restored;
  - a promotion with the writer's last transactions unarchived, then a restore across
    it;
  - the chain check across two basenames;
  - points by GTID;
  - `GRANT` replays.
- **Unit:** the engine's `Replica`, `Promote`, member status, server ids and the
  alias settings, as the Postgres engine's are tested.
- **e2e level-0:** a MariaDB's Replicas section and promotion, mocked.
- **Live:**
  - on dind: replicas, backup-from, promote, the alias followed by a consumer;
  - on the dev cluster: a replica on the other node, promoted.
