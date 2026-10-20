# Managed databases: upgrading their scripts

Date: 2026-10-21. It follows [the managed templates design](2026-10-17-managed-templates-design.md)
and the two engines' phases.

## The problem

A managed database keeps the scripts its template shipped when it was made. They
are config files that the engine locks. Nothing ever replaces them, so a database
never gets a fix made after its creation. Databases already running lack all of
these:
- the data lock, which keeps two containers off one volume;
- MariaDB's `GRANT` replay fix, without which a restore stops at a `GRANT`;
- the GTID replay;
- replicas and promotion. A database whose scripts predate replicas is refused one.

Each new engine adds databases that stay stuck on their scripts.

## What this does

A person sees that a managed database's scripts have a newer revision, reviews the
change, and applies it. HivePaaS then:
- replaces the scripts and the artifacts on the database, on each of its members
  and on its verifier;
- redeploys them, the member that writes last.

Only what the engine owns, and the template provides, changes:
- **The scripts:** the template's config files mounted under
  `/opt/hivepaas/bin`, each with its setting mount.
- **The artifacts:** the WAL-G binary, by version and hash for each architecture.

Nothing else changes:
- not the image, which the person moves within its line on the app's settings;
- not the engine's command, which the backend writes and not the template;
- not the variables, the parameters, or anything the person set.

The general template update (the app templates design, §12) is a three-way merge
of everything the person may edit. This is not that. The scripts are locked, so
there is nothing to merge: they are replaced. That update, when it comes, must
leave what the engine locks to this one. `AppTemplateSettings.Base` is not
touched.

Applying is always a person's act, as for templates: never automatic.

## 1. What the newer revision offers

The installation's templates are at one revision: the release's pin, or the local
directory in development.

**What HivePaaS reads from that revision:**
- It reads the app's template there, renders it for the app's version and variant
  with the app's parameters, and keeps secret placeholders unexpanded, as for the
  base (app templates design §4, "Two renders"). Scripts carry no placeholders.
- Rendering for an update does not refuse a deprecated version. Only creating an
  app does.
- From the render it keeps the config files whose `swarmRef.file.name` is under
  `walgkit.ScriptDir`, with their mode.
- From the template it keeps the artifacts.

**The fingerprint** is a sha256 over both, canonically ordered:
- the scripts' paths, modes and content hashes;
- the artifacts' names, versions and file hashes.

The current fingerprint is computed the same way from the app's own settings: its
config files mounted under the script directory, and its `app-artifacts`.

**States:**

| State | When |
|---|---|
| `up-to-date` | The fingerprints are equal. |
| `update-available` | They differ, and the revision can be applied (below). |
| `incompatible` | The revision's template names another engine, or another engine API than the app's. Or it lacks a script the engine's API needs (`walgkit.CheckScripts`). The reason is given. |
| `unavailable` | The app was not made from a template; the revision no longer has its template or its version; or the template is not a managed one. |

- **Changes are listed per script:** added, changed or removed. Artifacts are
  listed by version.
- **Members:** each member and the verifier is compared too. One whose scripts
  differ from the cluster's app is listed, for example after an apply that
  stopped halfway.
- **Engine API:** there are no migrations between engine APIs. A template whose
  `engineApi` differs from the app's is `incompatible` until an engine knows how
  to move a database from one API to the next.

## 2. Applying

`POST .../managed/scripts/upgrade` names the fingerprint the person reviewed. If
the revision changed since, the request is refused.

**Guards.** It holds the cluster's app row (`SELECT … FOR UPDATE OF app`) and
refuses:
- a database still restoring;
- a member, because the cluster is upgraded from its app;
- any state other than `update-available`.

**What it changes, in one transaction,** for the cluster's app, each member and the
verifier app:
- **Scripts by name:**
  - a changed script gets its new content, and its mount its new mode;
  - an added script gets a config file and a setting mount, both locked;
  - a removed script loses both. Only scripts under the script directory, and
    locked, are ever removed.
- **Artifacts:** `app-artifacts` is replaced by the revision's.
- **Locks:** `Locks.Settings` follows: new ids are added, and removed ones dropped.
- **Record:** the app-managed setting records `Scripts{Revision, Fingerprint,
  UpdatedAt}`.
- **Audit:** the cluster's app gets one entry, naming the revision and the scripts
  changed.

**Then it deploys:**
- every member that runs first (replicas, the verifier), and the member that
  writes last;
- each deployment ensures the artifacts on the node and applies the new mounts, as
  every deployment does;
- a fenced app or a stopped one is not deployed: it takes its scripts when it is
  started.

The response lists the deployments' tasks.

**Restarts.** Each member restarts once, and the writer is down while it restarts,
as in any deployment of the database. The data lock keeps the old and the new
container apart. The deployments are queued tasks: the writer's is queued last,
but HivePaaS does not wait for the replicas to come back first.

## 3. The engine's part

- The scripts are `walgkit.ScriptDir`'s, for every WAL-G engine.
- The engine API check is `Engine.APIRange` and the app's `EngineAPI`.
- The script check is the engine's own (`walgkit.CheckScripts` with its profile),
  run on the render.

There is no new engine method. If an engine owns something else a template
provides, the engine is asked for it then.

## 4. API, dashboard and CLI

**API:**
- `GET .../managed/scripts`: the state, the reason, the current and available
  revisions and fingerprints, the changes, and the members out of step.
- `POST .../managed/scripts/upgrade` with `{fingerprint}`: the deployments' tasks.

The API level stays 4, which is unreleased.

**Dashboard:** the Database tab gets a "Scripts" section with the state and the
changes. "Update scripts" confirms that each member restarts once and the writer
last. It covers e2e level-0 cases for each state.

**CLI:** regenerated. A `hivepaas db scripts` command can come later.

## 5. Testing

- **Unit:**
  - the fingerprint and its order;
  - each state;
  - the changes;
  - applying to a cluster's settings, with an added, a changed and a removed
    script and a new artifact;
  - a fenced member not deployed;
  - the fingerprint guard.
- **Live (dind):**
  - **MariaDB:** lc-maria-latest-0003, made with phase 1's scripts, is upgraded to
    the pinned revision. It keeps archiving, restores to a point, then takes a
    replica, which its old scripts refused.
  - **Postgres:** a database made before the data lock is upgraded and keeps
    archiving.
  - **Restart:** each member restarts once.

## 6. Not covered

- Engine API migrations (API 2).
- The engine's own drift: a server command the backend now writes differently.
- An update policy, or upgrades applied without a person.
- The general template update of §12, which must leave locked fields alone.
