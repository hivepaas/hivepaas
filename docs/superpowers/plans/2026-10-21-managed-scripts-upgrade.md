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

## Task 3 (dashboard, CLI)

- **Dashboard:** the Database tab's Scripts section, with the state, the changes,
  members out of step, and "Update scripts" with its confirmation.
- **e2e level-0:** each state, and the upgrade asked for.
- **CLI:** the client regenerated.
- **Live (dind):** the section on lc-maria-latest-0003 before and after.
