# Managed databases: upgrading their scripts, plan

Design: [2026-10-21-managed-scripts-upgrade-design.md](../specs/2026-10-21-managed-scripts-upgrade-design.md).

Each task is approved before it starts. Each task is test-first: lint the whole repo,
run every test, check it live, commit per repository, merge locally, then write
as-built notes here.

## Task 1 (backend): what the newer revision offers

- **Render:** the app's template at the installation's revision, for its version,
  variant and parameters, secrets left as placeholders. A deprecated version is
  not refused.
- **Fingerprints and changes:**
  - the scripts under `walgkit.ScriptDir` and the artifacts, of the render and of
    the app's settings;
  - the changes per script;
  - the members out of step.
- **States:** up-to-date, update-available, incompatible (with the reason),
  unavailable.
- **`GET .../managed/scripts`.**
- **Tests:** fingerprint order, each state, the changes.
- **Live (dind):**
  - lc-maria-latest-0003 and an old Postgres database show `update-available`,
    with their changes;
  - a database made from the current revision shows `up-to-date`.

**As built** (2026-10-21, backend `3a4663c1`, merged `7b110abb`; API level 4; the
CLI is regenerated in Task 3):
- **`templaterender.VersionTree(tmpl, version)`** is the app tree with the version's
  override applied and nothing substituted. A deprecated version is read; a
  template with components is refused.
- **`templaterender.Static(node)`** undoes the escapes and refuses any `${{ … }}`.
  Scripts use bash's `${VAR}`, which is no placeholder.
- **Render: no full render.** The design said "render with the app's parameters,
  secrets as placeholders". As built there is no render: only the
  `settings.configFiles` subtree of the version's tree is read, through `Static`.
  The scripts need no parameters, so neither the stored parameters nor generated
  secrets are involved. A placeholder in a script makes the revision
  `incompatible`.
- **Package `service/managedservice/managedscripts`:**
  - `FromTemplate` and `FromSettings` give the `Set`: the scripts by path (name,
    path, mode, decoded content, sha256) and the artifacts.
  - `Fingerprint` hashes one line per script (path, octal mode, sha256), then one
    per artifact by name (version, mount path, each build's sha256 by arch).
  - `Diff` lists the scripts by path (added, changed, removed), then the artifacts
    (added, changed with from and to, removed).
  - `Evaluate` checks the managed engine, then the engine API, then the version,
    then the scripts.
- **No `walgkit.CheckScripts` on the revision.** The design said the engine's own
  check runs on the render. As built it is not run: the release pins a revision
  this backend's lint passed. A revision that ships no script at all is
  `incompatible`.
- **`GET .../managed/scripts`:** `ManagedScriptsResp{state, reason,
  current{revision, fingerprint}, available{revision, fingerprint} or null,
  changes, outOfStep}`.
  - The current revision is the binding's `Base.Revision` until Task 2 records
    its own.
  - A member is refused (`ERR_APP_MANAGED_IS_MEMBER`, 412).
  - An app with no template binding is `unavailable`, and so is one whose template
    is gone from the store.
  - Members and the verifier are compared by fingerprint (`clusterMembers`).
- **Live (dind, `lc/check28.sh`; the dind source's revision reads `local`):**
  - lc-maria-latest-0003, with phase 1's scripts, shows `update-available`: 9
    changed and `walg-replication` added.
  - pg18s shows `update-available`: walg-entrypoint, walg-health and walg-lib
    changed.
  - A new lc-maria5 shows `up-to-date`, and its replica is not out of step.
  - The replica's own GET is refused with 412.
  - lc-maria5 and its replica are kept for Task 3.

## Task 2 (backend): applying

- **`POST .../managed/scripts/upgrade`** with `{fingerprint}`, under the cluster's
  row lock.
- **Settings:** the cluster's app, its members and its verifier get scripts added,
  changed and removed, `app-artifacts`, locks, `Scripts{Revision, Fingerprint,
  UpdatedAt}`, and an audit entry.
- **Deployments:** the running members first and the writer last; fenced and
  stopped apps are not deployed.
- **Tests:** applying to a cluster's settings; the guards (restoring, a member,
  the fingerprint changed, not `update-available`).
- **Live (dind):**
  - **MariaDB:** lc-maria-latest-0003 upgraded. Its container has the new scripts,
    it archives, a restore to a point works, and a replica is then added.
  - **Postgres:** an old database upgraded, still archiving.

**As built** (2026-10-21, backend `0affc1e7`, merged `ef0335b8`; app-templates fix
`5c57bd0`, merged `aeff621`; API level 4, with the POST's fingerprint added to
`api-level.json`):
- **`managedscripts.Plan(app, settings, available, now)`** returns a `Patch`:
  - created, updated and deleted settings, stamped with their update versions
    moved;
  - the lock ids gained and lost.

  A script is known by its config file's name:
  - **Changed:** its content changes in place, and so do its mount's path and mode.
  - **New:** it gets a config file and a mount named `EntryKeyFor(name)`, as a
    template makes them.
  - **No longer shipped:** both are soft-deleted (`DeletedAt`).
  - **Name taken:** a new script whose name one of the app's own config files has
    is refused.

  The artifacts setting is changed, created or deleted. It is not locked, as at
  creation.
- **`AppManagedSettings.Scripts{Revision, Fingerprint, UpdatedAt}`** is recorded
  on each engine setting: the cluster's app and its members. The verifier has no
  engine setting, so its scripts change alone. GET reads the current revision
  from this record.
- **`POST .../managed/scripts/upgrade` with `{fingerprint}`**, under `FOR UPDATE OF
  app`. It refuses:
  - a member;
  - a database still restoring (`ERR_APP_MANAGED_SWITCH_NOT_READY`);
  - nothing to apply (`ERR_ARGUMENT_INVALID`, 400). Up-to-date with a member out of
    step is allowed, which re-applies.
  - another fingerprint (`ERR_UPDATE_VER_MISMATCHED`, 412).

  Then it audits on the cluster's app, deploys every app with the writer last
  (`deployOrder`), and queues the tasks after the commit.
- **Stopped apps are deployed too.** A fenced member or a verifier between runs is
  deployed: a deployment keeps the service at its replicas, 0, and only updates its
  spec, so it takes the new files and starts with them. This differs from the
  design's "not deployed".
- **A bug found live in the handler:** it parsed the request with
  `ParseAndValidateRequest` and not `ParseAndValidateJSONBody`, so the
  fingerprint never arrived. Fixed before the merge.
- **A script bug found by the upgrade:** a phase 1 MariaDB, upgraded, refused a
  replica, because `walg-replication`'s GRANT failed: mysql@localhost had no GRANT
  OPTION. Its script now runs `ensure_grant_option` first (app-templates
  `5c57bd0`). The harness revokes the grant before `walg-replication`: replicas and
  promotion passed on 11.8 and 11.4. One run's timing-dependent
  `caught_up` check failed once on 11.4, and passed in two more runs.
  - **Release:** the release pin must move to app-templates `aeff621` or later.
- **Live (dind, `lc/check29.sh`, `check29b.sh` and `check29c.sh`):**
  - **lc-maria-latest-0003, phase 1's scripts:**
    - 10 changes offered; another fingerprint was refused with 412.
    - Upgraded: one deployment done, then up-to-date; applying again was refused
      with 400.
    - Its container has `walg-replication` and the new `walg-entrypoint`, and the
      archive loop runs.
    - Nightly ran, and a point marked at GTID `0-1-26` restored with 10/10 rows.
  - **Second upgrade, with the fixed `walg-replication`:** 1 change. The replica
    was then let in and streamed 25/25 rows, and was removed afterwards.
  - **lc-maria5 with its replica:** the replica was deployed first and the writer
    last. Nothing was out of step afterwards, the replica runs the new script and
    still streams.
  - **pg18s:** upgraded, up-to-date, its `walg-entrypoint` has the data lock, and
    its health check is done.

## Task 3 (dashboard, CLI)

- **Dashboard:** the Database tab's Scripts section, with the state, the changes,
  members out of step, and "Update scripts" with its confirmation.
- **e2e level-0:** each state, and the upgrade asked for.
- **CLI:** the client regenerated.
- **Live (dind):** the section on lc-maria-latest-0003 before and after.
