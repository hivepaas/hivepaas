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

## Task 1 (backend): points by GTID

- **`MarkResult`** stores the GTID (`GTID=`). A phase 1 point's file:offset still
  restores, by `STOP_POSITION`.
- **The restore target** of a point is `HIVEPAAS_RESTORE_STOP_GTID` when its position
  is a GTID list.
- **Tests:** parsing, and a target of each form.

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
