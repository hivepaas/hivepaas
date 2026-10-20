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

**As built** (2026-10-20, app-templates `4d0446d`, merged `39c9720`; backend
`f7794cb9`, merged `30a1047e`; API level 4, no CLI or dashboard change):
- **The member's server arguments come by variable, not in its command.** This
  differs from the design.
  - The engine gives `HIVEPAAS_SERVER_ID` and `HIVEPAAS_LOG_BIN`
    (`mariadb-bin-<member key>`) as locked variables.
  - `walg-entrypoint` appends them to `mariadbd`'s arguments; the last option wins.
  - A standby also gets `--read-only` and `--report-host=<member key>`.
  - So the generic replica code keeps copying the primary's command unchanged.
  - **Task 3 must not shed `HIVEPAAS_SERVER_ID` or `HIVEPAAS_LOG_BIN` at
    promotion.** A promoted member keeps its name and id, and a new basename would
    start an empty GTID state (Task 0).
- **The server id** is the first one no member has, from 2, recorded on the member
  (`ManagedMember.ServerID`, `ReplicaSpec.ServerID`).
- **`ReplicaSpec.FileOwner`** replaces the use case's hard-coded postgres uid. Both
  are 999.
- **Shared code moved to walgkit:** `walgkit.Replica` and `walgkit.MembersStatus`
  came from the Postgres engine, along with the constants (`EnvStandby*`,
  `ReplicaPasswordFile`, `OutReplicas`).
- **Seeding:**
  - The latest backup is fetched, and the server started on its socket with the
    member's arguments.
  - `walg-replay --since <backup> --state-out` replays the logs since, into the
    member's own logs with their GTIDs.
  - When nothing was replayed, `gtid_binlog_state` is set to the backup's state,
    so that the member's logs go on from the cluster's state.
  - Then `gtid_slave_pos` is set to the state reached.
  - The member reseeds at each start until a seed completes (state `standby`).
- **The replica's archive loop:**
  - It never pushes.
  - It runs `CHANGE MASTER … MASTER_USE_GTID=slave_pos` when `Master_Host`
    differs from `HIVEPAAS_STANDBY_HOST`. The password is read from its file into
    the client's stdin.
  - On `Last_IO_Errno=1236` it catches up: `walg-replay --since <oldest backup>
    --from <gtid_slave_pos>`, then a new `gtid_slave_pos` and `START SLAVE`.
    mysql@localhost replays into a read-only server.
  - It purges its own logs older than `max(2 × archive timeout, 1h)`.
- **Member status is read on the writer.** This differs from the design, where
  each replica reported its own.
  - `walg-replication` grants `REPLICATION SLAVE, SLAVE MONITOR` and keeps the
    password in `/var/lib/mysql/hivepaas-replication.cnf`. The file is mode 600,
    owned by mysql, and sits beside the data directory, so it is in no base backup.
  - The writer's `walg-health` asks each host from `SHOW SLAVE HOSTS` for
    `SHOW SLAVE STATUS`. It prints `name@state@bytes@seconds`, as Postgres does.
  - A replica that is catching up is not connected, so it shows as disconnected.
- **The multi-host `HIVEPAAS_DATABASE_URL_RW` and `_RO`** are now generated for
  Postgres only.
- **Promotion** is refused ("promotion") until Task 3. MariaDB's `Capabilities`
  are `{Replicas: true}`, so the dashboard offers a promote button that the API
  refuses until then.
- **Harness:** 166/166 on each of 11.8 and 11.4. It covers:
  - `walg-replication`;
  - two replicas seeded under writes;
  - read-only, logs of the member's own name with the primary's GTIDs, and none
    pushed;
  - `REPLICAS`;
  - no password on any process's arguments;
  - a stopped replica catching up from the archive after 1236;
  - a backup taken on a replica, restored, then retain;
  - a replica the archive cannot seed, which idles.

  `ONLY=<test>` runs one standalone test.
- **Live (dind, `lc/check25.sh`).** A new lc-maria2 (APP_MARIA2) got replica
  lc-maria2-r1:
  - read-only, server id 2, log `mariadb-bin-lc-maria2-r1`, password file 0400
    owned by mysql;
  - streaming per the health check, with 20/20 rows;
  - `backupFrom` → nightly on the replica (WAL-G hostname `lc-maria2-r1`) →
    restore latest held 25/25 rows;
  - removed, with its service and volume gone.

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

**As built** (2026-10-20, app-templates `4bb31fa`, merged `65d2d5e`; backend
`724ade59`, merged `7835d0d3`; API level 4, no CLI or dashboard change):
- **There is no `walg-promote`; promotion is `walg-entrypoint promote`.** The
  backend allows at most 10 config files per app, and the MariaDB template already
  ships 10 (walg-replication was the tenth).
  - Raising the limit would fail the lint that older supported releases run on the
    template, so the limit stays.
  - **Engine:** `walgkit.Profile.Promotion{Command, Requires}`. MariaDB's is
    `walg-entrypoint promote`, gated on `walg-replication`.
    - The Task 2 template has walg-replication but no promote. It was never
      released; only dind's lc-maria2 had it, and that is deleted.
  - Postgres keeps `walg-promote` as the default.
- **What promote does:**
  - It stops and forgets replication (`STOP SLAVE; RESET SLAVE ALL`), runs
    `read_only=0`, and leaves the marker.
  - Then it runs `FLUSH BINARY LOGS` and `walg-archive once`, which pushes every
    closed log of the member. WAL-G's cache on a replica is empty, so all are
    pushed.
  - **This differs from the design:** it pushes all of them, not only those past
    the archive's last GTID by `BINLOG_GTID_POS`. The replay skips the overlap,
    and the chain check covers each basename from its first log.
  - Exit codes are those of Postgres (0, 3, 1). It runs again as it is.
  - `walg-lib`'s `standby()` is false once the marker exists. So the running
    archive loop turns into a primary's at once, and a restart serves as primary
    without `--read-only`.
- **`HIVEPAAS_FENCED`:** `walg-entrypoint` idles, healthy, with no server.
- **Engine refactor:**
  - `walgkit.Promote` and the writer's recording (`promoteWriter`,
    `pointJobsAtWriter`) moved from the Postgres engine.
  - `walgkit.Reconfigure` now promotes for every engine.
  - `StandbyEnvKeys` are the three standby variables only. The member keeps
    `HIVEPAAS_SERVER_ID` and `HIVEPAAS_LOG_BIN`.
- **Not handled:** a replica ahead of the promoted member (one that got
  transactions the promoted member did not) meets 1236 against it. It then catches
  up from an archive that lacks its extra transactions, which the cluster has lost
  anyway.
- **Harness:** 192/192 on each of 11.8 and 11.4.
  - **Promotion:** the primary's archive loop is stopped, rows are written and
    reach both replicas but are never archived, and the primary is stopped.
  - **The promoted member:** r1 is promoted, writes under `0-2-…`, and archives
    its logs. Promoting again is idempotent.
  - **After promotion:**
    - r2 is recreated pointed at r1 and follows it without reseeding, and r1's
      health lists it as streaming;
    - a restore across the promotion holds all 40 rows;
    - retain passes over both basenames;
    - the fenced old primary idles, healthy, and promote there exits 3;
    - r1, restarted, is still the primary.
- **Live (dind, `lc/check26.sh` and `check26b.sh`):**
  - **Setup:** lc-maria3 with replicas r1 and r2; r1 was promoted.
  - **The cluster after:**
    - the anchor's service is at 0/0;
    - r1 writes and has its marker;
    - r2 is restarted onto r1 (`Master_Host`) and holds r1's rows;
    - members show primary and streaming;
    - health and nightly ran on r1.
  - **Restore:** latest held 36/36 rows.
  - **Cluster delete:** `DELETE ?deleteCluster=true` removed every service, and
    the volumes were removed.

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

**As built** (2026-10-21, backend `3bbfb582`, merged `73b82424`; dashboard
`3439de9b`, merged `b534baca`; no template change; the API level stays 4, so there
is no CLI change):
- **No app network setting holds the alias.** Network attachments live on the Swarm
  service alone, so the alias is recorded and enforced this way:
  - `ManagedDatabase.Alias` is the cluster's `<anchor key>-rw`.
  - `ManagedLocks.Aliases` is on whichever app carries the alias.
  - `walgkit.Profile.WriterAlias` turns it on; it is MariaDB's only.
  - `walgkit.Configure` records both.
  - The creation step (`apptemplateuc` `managedPlan.configure`) gives the new
    service each locked alias, on the network it is reached by its key on
    (`appservice.AddNetworkAlias`).
  - Restores into a new app go through the same step, so a restored database has
    its own `-rw`.
  - Replicas and verifiers copy no alias.
- **The lock:** the network settings update refuses a locked alias missing from
  the final attachments (`checkManagedAliases`, ERR_APP_MANAGED_FIELD_LOCKED
  "networks").
- **At promotion:** `PromoteSpec.Alias` drives a new step `moveAlias`, which runs
  after `record` and before `restart`:
  - the locks move first (`lockAliases` on the fenced app and the promoted app);
  - then `managedService.MoveAlias` updates the fenced service, removing the
    alias, and the promoted one, adding it. The promoted member restarts once.
  - A failure only warns, naming the alias to add by hand.
- **A deployment keeps the alias.** `applyNetworks` rewrites attachments only when
  the document names some; a managed template names none.
- **System variables:**
  - `HIVEPAAS_DATABASE_HOST_RW` is `Database.WriterHost`: the alias, or the
    writer's key for a database made before it.
  - `HIVEPAAS_DATABASE_HOSTS_RO` is the replicas' keys, or the writer's host when
    there is none.
  - They are shared, are not secrets, and are MariaDB's only. Postgres keeps its
    multi-host URLs.
- **envlink:** `Facts.HasClusterHost` and `clusterHostGroup`, recommended for a
  MySQL-family target. It suggests DATABASE_URL over HOST_RW, DB_HOST,
  DB_READ_HOSTS, and the individual parts.
- **Dashboard:**
  - `ManagedEngineCopy.clusterAddress` holds `{write, read, holds}`: the URL
    variables for Postgres and unknown engines, the HOST variables for MariaDB.
  - The Replicas section and the promotion's confirmation read it.
  - e2e level-0 18/18, with "a MariaDB's replicas name its writer's host, and one
    is promoted".
- **Live (dind, `lc/check27.sh`).** A new lc-maria4 with replica r1, and a consumer
  lc-mclient (the mariadb image, `sleep infinity`) whose env names
  `${lc-maria4.HIVEPAAS_DATABASE_HOST_RW}`, which resolved to `lc-maria4-rw`:
  - **Lock:** the service carried `lc-maria4-rw`, and removing it from the
    network settings was refused (locked).
  - **Promotion:** the consumer wrote through the alias. After r1 was promoted,
    the alias was on r1 and gone from the anchor. The consumer wrote again 9s
    later, its configuration unchanged.
  - **After:** the lock held on r1, and a redeploy of r1 kept the alias, with the
    consumer still writing.
  - **Cleanup:** the cluster, the consumer and the volumes were deleted.

## Task 5: the dev cluster and the release

- **On the dev cluster:**
  - a replica on the other node;
  - promoted;
  - the consumer follows the alias.
- **The release:** `release.json` pins app-templates, for the user to sign.
