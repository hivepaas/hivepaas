# Managed MariaDB phase 2: plan

Design: [2026-10-20-managed-mariadb-phase-2-design.md](../specs/2026-10-20-managed-mariadb-phase-2-design.md).

Each task is approved before it starts. Each task is test-first: lint the whole repo,
run every test and the harness, check it live, commit per repository, merge locally,
then write as-built notes here.

## Task 0 (app-templates): the GTID replay, logs by member, points by GTID

**Before replicas.** Phase 1 archives keep restoring.
- **`walg-replay`:**
  - replays in upload order, file by file, carrying the GTID state;
  - skips, applies or fails per the design §1;
  - takes `--stop-gtid`;
  - keeps `--position` for phase 1's points;
  - keeps `--until`.
- **`walg-entrypoint`:** in restore mode, `HIVEPAAS_RESTORE_STOP_GTID`.
- **`walg-restore-test`:** stops at its mark's GTID.
- **`walg-nightly retain`:** the chain check by GTID.
- **The log basename:** the scripts no longer assume `mariadb-bin`. The archive loop
  and the purge read the active log's name from the server.
- **Harness:**
  - every existing check passes;
  - the replay across logs of two basenames (written by restarting a database with
    another `--log-bin`);
  - a point by GTID;
  - a gap named.

**As built** (2026-10-20, app-templates `db74d7e`, merged `9c46124`):
- **`walg-replay`:**
  - the stop flag is `--gtid GTIDS`, not `--stop-gtid`;
  - before the loop it runs `ensure_grant_option`;
  - a stop GTID the archive never reaches fails the restore ("The archive does not
    reach …");
  - its last line names the logs replayed and skipped, and the state reached.
- **`walg-restore-test`** stops at its mark's GTID when it has one, and at its
  file:offset otherwise.
- **The chain check goes per basename, not by GTID.** Each basename's logs must
  follow one another by number. The oldest backup's basename is checked from after
  its `binlog_start`; every other basename from its first log. The replay itself
  still catches a GTID gap, and names the file.
  - A backup taken before any log was archived has no `binlog_start`. The check used
    to read that as "No base backup" and checked nothing. It now checks every
    basename from its first log.
- **No change for the basename.** The archive loop and the purge already read the
  active log from `show master status`. At a rename, `FLUSHED_AT` compares offsets
  across files, which costs at most one extra rotation.
- **A new basename starts an empty GTID state.** The server keeps the state in
  `<basename>.state`, so a database restarted under a new `--log-bin` numbers from
  `0-1-1` again. The harness copies the state file to simulate a rename.
  - Never rename the anchor's logs.
  - A replica's state comes through replication (`--log-slave-updates`), so it
    needs no copy (Task 2).
- **WAL-G** pushed `mariadb-bin-b.*` after `mariadb-bin.*`, although `-b` sorts
  before `.`.
- **Harness:** 268/268 on 11.8 and 11.4. It adds:
  - a point by GTID (30 rows);
  - a restore of latest across two basenames (50 rows);
  - the chain naming both basenames;
  - a deleted middle log: retain fails naming it, and a restore fails with "binary
    logs break at".
- **Live:** none for this task, which changes the scripts only. The harness runs
  them in the template's own images. The dind check comes with Task 2.

## Task 1 (backend): points by GTID

- **`MarkResult`** stores the GTID (`GTID=`). A phase 1 point's file:offset still
  restores, by `STOP_POSITION`.
- **The restore target** of a point is `HIVEPAAS_RESTORE_STOP_GTID` when its position
  is a GTID list.
- **Tests:** parsing, and a target of each form.

**As built** (2026-10-20, backend `8529f16f`, merged `63e03446`; CLI `75e0f75`, spec
descriptions only):
- **`MarkResult`** returns the `GTID=` list. It falls back to `POSITION=` when the
  GTID is empty: a database with no transaction yet.
- **The point target:** a GTID list gives `HIVEPAAS_RESTORE_STOP_GTID`, and
  FILE:OFFSET gives `HIVEPAAS_RESTORE_STOP_POSITION`. Anything else is refused.
- **A guard against old scripts.** A template whose `walg-entrypoint` does not
  mention `HIVEPAAS_RESTORE_STOP_GTID` would ignore the stop and replay to the last
  log archived. So a restore to a GTID point with such a template is refused
  (`ErrAppManagedRestoreTargetInvalid`, "update the template, or restore to the
  point's time"). The check is `walgkit.ScriptMentions`.
  - Points already made before this deploy are FILE:OFFSET, and keep restoring
    with any template.
  - **Release:** pin app-templates to `9c46124` or later in the same release as this
    backend. Otherwise new points are made by GTID and refused until the template
    ships.
- **The API level stays 4.** The DTO descriptions of `position` name the GTID list.
- **Dashboard:** nothing to change; it never shows a point's position.
- **Live (dind, `lc/check24.sh`):** run with app-templates `9c46124` loaded.
  - A point marked on lc-maria-latest-0003 got `0-1-23`. Its restore replayed 14
    logs, skipped 1, stopped at `0-1-23`, and held 10/10 rows.
  - The earlier point `before-x-0223` (`mariadb-bin.000007:971`) restored with the
    new scripts by STOP_POSITION, 10/10 rows.

## Task 2 (app-templates, backend): replicas

- **Scripts:**
  - `walg-entrypoint`'s standby mode (seed, connect, idle when it cannot seed);
  - the replica archive loop (purge with a margin, catch up on 1236);
  - `walg-replication`;
  - replica outputs in `walg-health`.
- **Engine:**
  - `Replica`: the standby variables and secret, the server id and log basename in
    the member's command, `--read-only`, `--report-host`;
  - `AllowReplicas` via `walg-replication`;
  - members' status;
  - `Capabilities{Replicas: true}`.
- **Harness:**
  - two replicas under writes;
  - a late one catching up;
  - a backup on a replica, restored.
- **Live (dind):** add and remove a replica; backup-from.

## Task 3 (app-templates, backend): promotion

- **Scripts:** `walg-promote` (design §3); the fenced idle as Postgres's.
- **Engine:**
  - `Promote`: the fence variables, the promote command, standby keys;
  - the other replicas pointed at the promoted member.
- **Harness:**
  - a promotion with the writer's last transactions unarchived;
  - a restore across it;
  - the other replica following.
- **Live (dind):** promote; the cluster deleted whole.

## Task 4 (backend, dashboard): the cluster's address

- **The `<anchor>-rw` alias:**
  - on the anchor's network settings at creation, locked;
  - moved at promotion, with the promoted member's service updated.
- **System variables:** `HIVEPAAS_DATABASE_HOST_RW` and `HIVEPAAS_DATABASE_HOSTS_RO`;
  the envlink cluster group.
- **Dashboard:**
  - the engine table's address copy;
  - the Replicas section and promotion for MariaDB;
  - e2e level-0.
- **Live (dind):** a consumer on `${db.HIVEPAAS_DATABASE_HOST_RW}` writes before and
  after a promotion.

## Task 5: the dev cluster and the release

- **On the dev cluster:**
  - a replica on the other node;
  - promoted;
  - the consumer follows the alias.
- **The release:** `release.json` pins app-templates, for the user to sign.
