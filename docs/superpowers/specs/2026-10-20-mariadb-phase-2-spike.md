# Managed MariaDB phase 2: spike

Date: 2026-10-20. Phase 2 adds what Postgres has had since its phase 2:
- read replicas, and base backups taken on one;
- promotion of a replica over its fenced primary;
- an address of the cluster.

The question is whether MariaDB can do these with the official image, WAL-G and our
scripts, as phase 1 does. It can, but some of phase 1 must change first.

Everything was run with plain containers on one host's Docker, against SeaweedFS as
S3. It used MariaDB 11.8.9 and the phase 1 template's own scripts. GTID support was
checked in `mariadb-binlog` 11.4 and 11.8.

## What was checked

| Check | Result |
|---|---|
| Seed a replica from the archive | It worked under steady writes, two writes a second: the latest base backup, `mariadb_backup_binlog_info` read for its GTID, the archived binary logs replayed, then `CHANGE MASTER TO … MASTER_USE_GTID=slave_pos`. Both replicas held the primary's rows exactly. |
| A replica that connects late | Waiting longer than one archive pass gave error 1236, "required binlog files have been purged". The archive loop purges the primary's logs as soon as they are pushed. Replaying the newly archived logs from the replica's GTID and starting replication again caught it up. |
| A base backup taken on a replica | The backup lands in the cluster's archive. Its `mariadb_backup_binlog_info` names the **replica's** own binary log (`mariadb-bin-s2.000001 103322`), which is not in the archive. A replay from its GTID (`0-1-391`) over the primary's archived logs restored it exactly. |
| `mariadb-binlog` with GTID positions | 11.4 and 11.8 take `--start-position` and `--stop-position` as GTID lists. A stop at a GTID includes that transaction: a point at `0-1-400` restored the 395 rows written up to it. |
| Promotion | Fence the old primary, then on the replica `STOP SLAVE; RESET SLAVE ALL; SET GLOBAL read_only=0`. Writes went on under the replica's server id (`0-2-…`). The other replica followed the new primary with `CHANGE MASTER TO MASTER_HOST=… MASTER_USE_GTID=slave_pos`. |
| Restore across a promotion | It restored the 40 rows exactly from the old primary's base backup. The logs replayed were the old primary's archived logs, then a replica's log that repeats them and goes further, then the new primary's. |

## What phase 1 must change first

1. **A replay of `GRANT` fails today. This is a phase 1 bug.**
   - `mysql@localhost` has `ALL` but not `GRANT OPTION`, and
     `MARIADB_MYSQL_LOCALHOST_GRANTS` cannot add it. The image puts the value before
     `ON *.*`.
   - So any `GRANT` or `REVOKE` in the source's binary logs stops a restore, a restore
     test or a replica's seeding. The restore fails at that line.
   - **The fix:** `mysql@localhost` gets `WITH GRANT OPTION` when the data directory is
     initialized. An init SQL file run by the image's entrypoint is enough;
     `walg-entrypoint` can put it in `/docker-entrypoint-initdb.d`.
   - A database restored from a backup keeps the source's grants. Databases made
     before the fix need the grant once, as root. Only the dev cluster has any.

2. **The replay goes by GTID, file by file.** Phase 1 replays from the backup's
   file:offset, and points are file:offset. Both break with replicas:
   - A backup taken on a replica records the replica's own file, which is not in the
     archive.
   - After a promotion, the logs of two servers are in one archive.

   The prototype's rule is per file, carrying the GTID state reached so far:

   | `mariadb-binlog --start-position=<state> <file>` | Meaning |
   |---|---|
   | succeeds | apply its output, and move the state to its last GTID |
   | "never reached expected GTID state" | the whole file is at or before the state: skip it |
   | "Binary logs are missing data for domain N…" | a gap: the restore fails, naming the file |

   One `mariadb-binlog` over all the files does not work. It stops at the first
   repeated GTID: "Found out of order GTID. Got 0-1-16 after 0-1-27".

   Points record their GTID (`walg-mark` already prints `GTID=`), and a restore stops
   at it. The point's file:offset stays for display only.

3. **Each server's binary logs have a name of their own.** A promoted replica numbers
   its logs from its own start. Under the shared name `mariadb-bin` they would
   overwrite the old primary's in the archive.
   - Renumbering does not work: `RESET MASTER TO n` empties the GTID state, and
     `SET GLOBAL gtid_binlog_state` resets the logs to `000001`.
   - So the basename is per member, for example `--log-bin=mariadb-bin-<member>`. The
     primary keeps `mariadb-bin`, so databases made in phase 1 do not change.
   - The replay orders logs by upload (`binlogs_order`), not by name.
   - The nightly chain check checks GTID continuity: each file's starting state is
     the previous file's end. It no longer checks name numbers.

## What phase 2 has to do, from the spike

- **Replica start (`walg-entrypoint`, standby mode):**
  - An empty data directory is seeded: fetch, replay the archive, set
    `gtid_slave_pos`, `CHANGE MASTER`, `START SLAVE`.
  - Every start sets `read_only`, `--server-id=<n>` unique per member,
    `--report-host=<member key>` (so the primary lists it), and the member's log
    basename.
  - Replicas do not archive: their archive loop only purges.
  - Local administration runs with `sql_log_bin=0`. The spike's own `GRANT` on a
    replica left an errant GTID (`0-2-1`), which would follow it into a promotion.
- **Catch-up:** error 1236 means the primary purged what the replica needs. The
  replica's loop answers it by replaying the newly archived logs from its GTID and
  starting replication again. This is the analogue of Postgres's `restore_command`
  fallback.
- **The replication user:** made on the primary by a `walg-replication`, with the
  password on stdin. It runs as `mysql@localhost` once that account can grant; until
  then, as root through an option file in `/dev/shm`.
- **Promotion (`walg-promote`):**
  - The old primary is fenced first, as for Postgres.
  - Then `STOP SLAVE; RESET SLAVE ALL; read_only=0`, a marker, and archiving on.
  - Before archiving, the promoted replica pushes its own logs that hold transactions
    the archive lacks. These are the old primary's last ones, which were never
    archived if it died. `BINLOG_GTID_POS(file, size)` gives each log's end state, so
    only the logs past the archive's last GTID are pushed. Their overlap with the
    archive is what the per-file replay skips.
  - The other replicas `CHANGE MASTER` to it. One that lagged catches up from the
    archive.
- **Replica status:**
  - The primary's `SHOW SLAVE HOSTS` lists the connected replicas by report host.
  - Lag is a replica's own reading (`Seconds_Behind_Master`, the GTID it has
    applied). So a replica's health output carries it, and the engine joins the two.
    Postgres reads both on the primary.
- **The cluster's address:**
  - MariaDB clients have no common multi-host form. Postgres's
    `target_session_attrs` has no equivalent in PHP, Python or Go drivers; only
    Connector/J (`jdbc:mariadb:sequential://` and `replication:`) and Node's pools
    know several hosts.
  - The likely form is a stable name for the writer:
    - a Swarm network alias `<cluster>-rw`, which the engine moves to the promoted
      member's service (one restart of it);
    - or `HIVEPAAS_DATABASE_HOST_RW`, rewritten on promotion, with consumers
      redeployed as the switch does;
    - plus `HIVEPAAS_DATABASE_HOSTS_RO`, the replicas.

  This is for the design to settle.

## Not looked at

- arm64.
- Several replication domains: a person's own `gtid_domain_id`. The rule above
  carries one state per domain, and the prototype handled one.
- Semi-synchronous replication. Replication here is asynchronous, as Postgres's
  replicas are.
