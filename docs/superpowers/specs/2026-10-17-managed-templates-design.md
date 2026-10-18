# Managed templates: Postgres with point-in-time recovery

Status: draft 2026-10-17, revised the same day with the answers to its open questions
(see Decided). Nothing is built.

## Why

A template creates an app and then lets go of it. That is enough for most of the
store, and not for a database, whose hard part comes after the first deploy:
continuous WAL archiving, base backups and their retention, checking the chain is
whole, proving a restore works, restoring to a point in time, and saying when
any of it has stopped working.

Trying to do WAL-G with an ordinary template ran into the template system's edges,
each one by design:

- **No parameter can name a backup repository or a cloud storage.** `ParamType` is
  a closed set (`templatemodel/template.go`), because the dashboard draws one form
  field per type.
- **A template cannot create scheduled jobs.** `specmodel.BuildableBlocks` lists
  them under the phase 3 TODO of the app templates design (§12).
- **A template mounts volumes only.** `checkStorage` refuses any mount whose type is
  not `volume` (`specmodel/buildable.go`), so a tool binary cannot be put into the
  container.
- **A template has no logic.** An S3 repository, a volume repository and a custom
  bucket need different handling. YAML can only substitute.
- **Nothing runs after creation.** There is no place for PITR, a restore test or
  archiver health.

The other way out, building `postgres + wal-g` images, makes HivePaaS rebuild every
database image on every upstream minor, for every major and architecture. It
would also push those images into `release.json`, which pins the system apps of
one HivePaaS release. A user's database is not one of those, and its versions move
monthly, on their own cadence.

The tools do not need an image. WAL-G publishes binaries that run on the official
Debian images as they are. Measured on `wal-g-pg-20.04-amd64` v3.0.8:

- 66,978,616 bytes, and its sha256 matches the published `.sha256`;
- `NEEDED` only `libc`, `libm`, `libpthread` and `libresolv`; libsodium, lz4 and
  brotli are linked in;
- the newest symbol version it needs is `GLIBC_2.29`. `postgres:17` and
  `postgres:18` on Debian have 2.36 (bookworm) or 2.41 (trixie);
- `aarch64` builds of the same are published. They were not inspected.

Alpine images are out: musl is not glibc.

## Goals

- A **Postgres 17 and 18** database app that archives WAL to S3, so that what a
  failure loses is about `archive_timeout`, 60 seconds by default, not the time since
  the last dump.
- **Base backups and retention, a chain check, a scheduled restore test, and archiver
  health**, all created with the app and notifying on failure.
- **Restore to a point in time into a new app.** The source app is never touched.
- **No image is built by HivePaaS.** Images, tool versions and their hashes reach
  installations through the template channel, which is already signed.
- **One mechanism that MySQL, MariaDB and the rest can use later.**

Not in this design: automatic failover, synchronous replicas, major version
upgrades, in-place restore, Alpine variants, MySQL and MariaDB.

## The shape

```
app-templates (data, signed, frequent)          hivepaas (code, per release)
──────────────────────────────────────          ─────────────────────────────
postgres-managed.yaml                           managedservice
  versions 17 / 18 → official images              Engine "postgres", engineApi 1..1
  managed: {engine: postgres, engineApi: 1}       Configure → jobs, mounts, args, env
  artifacts: wal-g → url + sha256 per arch        Locked fields, Status, RestoreToNewApp
  configFiles: walg-* scripts                   agent: ArtifactEnsure → node cache
```

**Data goes in the template and logic goes in Go.** A managed template is an
ordinary template plus two blocks. It renders through the same pipeline into the
same settings. Then the engine named by `managed.engine` adds what it owns, and
from then on it looks after the app.

The rule that keeps this honest: **when a template seems to need a condition or a
loop, that part belongs to the engine.** The template only sets the engine's
defaults.

### Why not the alternatives

| | Against it |
|---|---|
| A plain template | The five edges above. |
| HivePaaS-built images | A build per engine × major × minor × arch, scanned and signed. Upstream security fixes wait for our pipeline. |
| `release.json` | It pins the system apps of one release. A database's image moves faster than HivePaaS does. |
| Download at container start | Puts the network and GitHub on the start path of a database, on every restart, and needs a downloader in the image. |
| A sidecar container | `archive_command` runs inside the postgres process, so the binary must be in that container. Swarm has no pods. |

## 1. Template schema

```yaml
metadata:
  name: postgres-managed
  requires: {versionCode: v000002}
managed:
  engine: postgres
  engineApi: 1
  postgres:                      # defaults for the create form, all optional
    archiveTimeout: 60s
    baseBackupSchedule: "0 2 * * *"
    retainFull: 7
    restoreTestSchedule: "0 4 * * 0"
artifacts:
  - name: wal-g
    version: 3.0.8
    mountPath: /opt/hivepaas/bin/wal-g
    files:
      amd64:
        url: https://github.com/wal-g/wal-g/releases/download/v3.0.8/wal-g-pg-20.04-amd64
        sha256: 9a09c2b1afad6a4e7d87444b34726dd098b60ec816032af921d2f887e6e285c5
        size: 66978616
      arm64:
        url: https://github.com/wal-g/wal-g/releases/download/v3.0.8/wal-g-pg-20.04-aarch64
        sha256: …
        size: …
versions:
  - name: "17"     # image: postgres:17.x-trixie, pinned like any template's
  - name: "18"     # image: postgres:18.x-trixie; PGDATA moves as in postgres.yaml
```

- **`managed.engine`** is one of the engines the binary knows: `postgres` for now.
- **`managed.engineApi`** is the version of the contract between template and
  engine: which scripts exist, their arguments, exit codes and environment names.
  An engine supports a range. A template outside it is incompatible, exactly like a
  template that needs a newer `versionCode`: it is listed and cannot be read.
- **`managed.<engine>`** is a small block of defaults, decoded strictly by that
  engine.
- **`artifacts[]`** each have a name, a version, an absolute `mountPath`, and
  `files` by architecture, each with an `url` (https only), a `sha256` and a `size`.
  The template file is hash-chained from the signed release info (release info →
  `index.json` → template), so a hash in the template is trusted as much as the
  template is.
- **Version code.** The two blocks are new to the schema, so this adds
  `VersionCodeV2` and moves `CurrentVersion` to it in the same release, following the
  app-templates README. It is the first move past `v000001`.
- **`requires.features`** names the license features a template needs (§8). It is new
  in `v000002` as well, and it is not specific to managed templates.
- **Lint** checks that:
  - the engine is known and `engineApi` is in its range;
  - every URL is https and every hash is 64 hex characters;
  - `amd64` is present;
  - a `mountPath` does not sit under a volume mount;
  - no variant name or image tag of a managed template says `alpine`. This is a
    heuristic, and it is documented as one.
- **The index** gains `managed: <engine>` per entry, so the store can mark it. The
  index is decoded leniently, so older installations ignore the field.

## 2. Artifacts

### On the node

Each node keeps a content-addressed cache at
`<HP_STORAGE_HOST_DIR>/cache/artifacts/sha256/<hex>`, mode `0755`. The layout is the
one the template cache uses on the control plane. The agent writes it at
`/host<dir>/…`, since the agent mounts `/:/host`.

### Getting it there

A new agent RPC, `ArtifactService.Ensure{url, sha256, size}`, has the agent of a node:

1. download from the URL itself, not through the backend;
2. hash while streaming;
3. write to a temp file next to the target and rename it into place;
4. return the host path.

It is idempotent: a file already there with the right size and hash is a no-op. A
hash mismatch removes the temp file and fails, naming the artifact.

**Artifacts are downloaded from where their project publishes them.** HivePaaS
re-hosts nothing, so it takes on no redistribution, no storage and no pipeline. What
the source can do wrong is caught:

- A re-uploaded or tampered file fails the hash and is not used.
- A release that was deleted fails the download. Nodes that already hold the file keep
  it, because the cache is keyed by hash.

Each node downloads a version once.

An optional mirror base URL, as a system setting, is tried before the source. It is
for installations that cannot reach GitHub (air-gapped, egress-filtered) or that want
no dependency on it. The mirror is trusted no more than the source: the hash decides.

`FileService` was the closest existing path and is not reused: its bytes travel
internet → backend → agent, it creates files `0600`, and it checks no hash.

### Which nodes

Swarm reschedules tasks without HivePaaS: node loss, scale, and about forty other
`ServiceUpdate` callers. A check at deploy time alone is not enough. Phase 1 makes
the question small:

- **A managed database's data volume must be pinned to a node.** The app is then
  pinned through the existing `VolumePin` constraint, and the artifact is ensured on
  that node only. An unpinned (shared) volume is refused, with the reason given.
- The node's architecture is read from Swarm (`Description.Platform.Architecture`,
  mapping `x86_64` to amd64 and `aarch64` to arm64). Nothing stores it today.

Unpinned volumes are Later. They need artifacts on every node that is eligible for
the app, and HivePaaS has no helper that evaluates a spec's constraints against
nodes.

### When

- **At creation, before `ServiceCreate`.** `ProvisionApp` creates the service, and so
  starts a task, before the first deployment. The managed path ensures the artifact
  before `ProvisionApps`.
- **At every deploy**, next to `prepareDockerAPI`. Unlike the Docker API, a failure
  here **fails the deploy**: the task would not start anyway.

### The mount

The mount is a bind mount from the cache file to `mountPath`, read-only, **without
`CreateMountpoint`**: Docker must never create a directory where the file belongs.

It is HivePaaS's own mount, treated like the Docker API socket mount:

- hidden from the storage GET;
- left out of the storage version and of spec export;
- kept by a storage save (the `KeepSocketMounts` pattern);
- applied again on every deploy.

Templates still cannot write bind mounts. Only the engine adds this one.

### Cleanup and trust

- **Cleanup.** Node cleanup removes an artifact that no app has referenced for its
  retention. It never removes one that a running app references.
- **Trust.** The hash comes from the signed chain. Downloads are https only and
  capped at 256 MiB.
- **Scanning.** Scanning the binary with Trivy (Go module metadata) is Later.

## 3. The engine

### Interface

A new `service/managedservice`, with one engine per database:

```go
type Engine interface {
    Name() EngineType
    APIRange() (min, max int)
    // Configure returns what the engine owns of an app being created from its template:
    // settings (jobs, scripts, secrets), the artifact mounts and the container args.
    Configure(ctx context.Context, req *ConfigureReq) (*ConfigureResp, error)
    // Apply re-applies what the engine owns to a service spec, at every deploy.
    Apply(ctx context.Context, req *ApplyReq) error
    Locked() []Field
    Status(ctx context.Context, req *StatusReq) (*Status, error)
    PlanRestore(ctx context.Context, req *RestoreReq) (*RestorePlan, error)
}
```

It plugs into `CreateAppFromTemplate`, through the `Configure` closure that
`appProvisionService.ProvisionApps` already takes. `CreateFunction` is the
precedent: one transaction, the app plus its settings. `ApplyAppConfiguration`
already schedules any scheduled-job settings that `Configure` returns.

### State: the `app-managed` setting

A new setting type, scope app, versioned with a `Migrate` like every other. The
artifacts an app was created with are not in it: they are in `app-artifacts`, which
§2's mounts read whatever the engine.

```go
type AppManagedSettings struct {
    Engine    EngineType
    EngineAPI int
    Postgres  *ManagedPostgres
}

type ManagedPostgres struct {
    Storage        ManagedStorage   // where WAL-G writes
    EncryptionKey  EncryptedField   // WALG_LIBSODIUM_KEY
    ArchiveTimeout time.Duration
    RetainFull     int
    Jobs           ManagedPostgresJobs // the scheduled jobs it created, by role
    RestorePoints  []RestorePoint      // §4, named by users
}

type RestorePoint struct {
    Name      string     // unique per app, at most 63 bytes, not starting with "hp-"
    LSN       string     // returned by pg_create_restore_point; what a restore targets
    Time      time.Time
    CreatedBy ObjectID
    Note      string
    KeepUntil *time.Time // set: retention keeps it restorable until then
}

type ManagedStorage struct {
    BackupRepo   ObjectID // either: a backup repository, reused (S3 or volume)
    CloudStorage ObjectID // or: any cloud storage setting
    KeyAuth      ObjectID // optional: a narrower key than the storage's own
    Prefix       string   // walg/<app key>, set at creation, never changed
}
```

**There are three places WAL-G can write to.** Two of them resolve to a `CloudStorage`
setting, which already exists with its key in an encrypted `KeyAuth`.

- **Reuse an S3 backup repository.** Its cloud storage is used with the prefix
  `<repo prefix>/walg/<app key>`, which no Kopia repository uses.
- **A cloud storage of the user's choice.** Any `CloudStorage` setting, used by a
  repository or not.
- **Reuse a volume backup repository.** WAL-G writes with its file storage
  (`WALG_FILE_PREFIX`). The engine mounts the subpath `<repo prefix>/walg/<app key>` of
  the repository's volume into the container, as an ordinary volume mount. That needs
  the volume to be reachable where the database runs: pinned to the database's node,
  or unpinned (shared storage). A volume pinned to another node is refused, with the
  reason given.

  **The form warns before it saves, and Status keeps saying it afterwards:**
  - **On the database's node**, the archive is not independent. Losing the node or its
    disk loses the database and every way back. The warning asks for a repository on
    other storage, or a second copy off the node.
  - **On shared storage**, the warning is milder: the archive is only as independent as
    that storage is from the database's disk.

  Status shows the archive's placement as `off-site`, `shared` or `same-node`.
- **Credentials resolve through `cloudstorageservice.S3Config`**, the reuse point the
  registry, files and job output already share. No other code reads a repository's
  keys today, and this keeps it that way.
- **They reach the container as Swarm secrets**, not as env. A wrapper script exports
  them for the one process that runs `wal-g`, so they never show in
  `docker inspect`.
- **The key's reach.** A repository's key can also delete its Kopia snapshots. The
  form offers a separate `KeyAuth`, and says why. HivePaaS cannot mint a narrower key
  itself, because IAM differs per provider.
- **Key rotation.** A managed app references its storage and key through resource
  links, so changing either marks the app for a redeploy.

### What the engine sets on Postgres

- **Command arguments:**
  - `archive_mode=on`, `archive_timeout=<n>s`, `wal_level=replica`;
  - `archive_command='/opt/hivepaas/bin/walg-run wal-push %p'`;
  - `shared_buffers` derived from the memory limit.

  These are re-applied on every deploy. Arguments the user adds come after them.
- **Environment:** `PGDATA` per major (18 moves it, as `postgres.yaml` already handles),
  and `WALG_*` for compression, prefix, `WALG_PREVENT_WAL_OVERWRITE=true` and delta
  steps. These names are reserved on a managed app.
- **Scripts,** shipped in the template as `settings.configFiles` (Swarm configs, under
  128 KiB each). The engine checks that the scripts its `engineApi` names are present.

| Script | Does |
|---|---|
| `walg-lib` | What the others share, sourced: outputs, the local server, the backup list |
| `walg-run` | Exports the secrets, then runs `wal-g` with its arguments, as the postgres user. With `--source`, the archive a restored database reads |
| `walg-fetch` | `wal-fetch` for `restore_command`. Exit 74 (not found) passes through; any other error becomes 255, so recovery stops instead of ending early at an older point |
| `walg-nightly` | `backup`: `backup-push $PGDATA`. `retain [<backup>]`: `delete before FIND_FULL <backup> --confirm` on the boundary the engine computes (§4, "Retention and kept points"), then `wal-verify integrity timeline`. Both print the backups kept |
| `walg-health` | Reads `pg_stat_archiver`, `pg_wal` and the archive's backlog; prints JSON and the outputs Status reads |
| `walg-mark` | `walg-mark <name> <seconds>`: a restore point, once its WAL is archived (§4, "Restore points") |
| `walg-restore-test` | §5 |
| `walg-entrypoint` | The entrypoint of a restored database (§4); `status` and `health` beside it |

As built, the health check fails when the last WAL archived failed and none has been
since, or when a segment has waited three archive timeouts: both are read from the
server, with no state between runs. The container's health check is `walg-entrypoint
health`, which is `pg_isready` except while a restore fetches or recovers. Before 18,
the data is in `/var/lib/postgresql/data/pgdata`, so that a restore test has room on
the data volume beside `PGDATA`, as it has under 18's `/var/lib/postgresql`.

### Jobs the engine creates

They run in a running task of the app, keep their history and outputs, and notify on
failure through the scope's default notification.

| Job | Default schedule | Runs |
|---|---|---|
| Nightly base backup | `0 2 * * *` | a `managed-nightly` job: `walg-nightly backup`, then the engine's retention decision, then `walg-nightly retain` |
| Archiver health | every 5 min | `walg-health` |
| Restore test | `0 4 * * 0` | `walg-restore-test`, a `container-command` job, or `managed-restore-test` in a verifier |

**Jobs are owned.** `SchedJob` gains `ManagedBy ObjectID`, pointing at the
`app-managed` setting. `UpdateSchedJob` and `DeleteSchedJob` refuse such a job; it is
changed through the managed settings use case. Today nothing refuses edits to the jobs
HivePaaS makes for itself either (image-scan, license-refresh, SSL renewal), and the
same field could later protect them. That is outside this design.

### Locked fields

There is no generic field lock today. Function apps are the precedent: a predicate on
the app-kind setting, and a check in each settings use case
(`appsettingsuc/function_rules.go`), sometimes a refusal, sometimes a silent pin. Phase
1 follows it, and always refuses, with `ErrAppManagedFieldLocked` naming the field:

| Where | Locked |
|---|---|
| Kind | category and engine |
| Deployment | the image repository. The tag may move within its major, through the existing Postgres tag-family logic of image overrides |
| Container | the engine's arguments. The user's own come after them |
| Storage | the data volume's mount and the artifact mounts |
| Env vars | `WALG_*`, `AWS_*`, `PGDATA` |
| Scheduled jobs | the managed jobs |

**Unmanage** is the way out. It removes the `app-managed` setting and its jobs, and
leaves the app as an ordinary app with its data and its archive. It is audited.

### Status

Status is computed on read from what the jobs last wrote. There is no background
work, and no exec into the container on read.

- **Archiving:** `ok` / `failing` / `stale`, with `last_archived_time` and `failed_count`.
- **The size of `pg_wal`.**
- **The restore window:** from the earliest retained base backup to the last archived
  WAL.
- **The last base backup:** its time and size.
- **The last restore test:** passed or failed, the time it took, and the point it
  reached.
- **Restore points:** each with the date until which it can still be restored. That
  date is the point's `KeepUntil`, or for a point that is not kept, when retention
  will drop the base backup before it.

It is shown on the app's page, and it bounds the time a PITR may pick.

## 4. Restore to a point in time, into a new app

### What to restore to

The user picks one of four targets. Each restores into a new app.

| Target | For | Recovery settings |
|---|---|---|
| **Latest** | A lost node or disk: as much as there is | No target. Every archived WAL is replayed, then the server promotes |
| **A time** | "Someone deleted it at 14:32" | `recovery_target_time`, with its zone |
| **A restore point** | "Just before migration v42" | `recovery_target_lsn` of the point, `recovery_target_inclusive = on` |
| **A base backup** | The quickest restore that is certainly consistent | That backup, `recovery_target = 'immediate'` |

The form offers only what lies inside the restore window. For a time or a point, the
engine fetches the latest base backup that starts before the target.

### Restore points

A user names a moment, to come back to it exactly. This is what
`pg_create_restore_point` is for. A point is created from the dashboard, the API or
the CLI (`hivepaas db mark <app> <name>`). The engine:

1. checks the name: unique in the app, at most 63 bytes, and not starting with `hp-`,
   which is reserved for HivePaaS's own points, such as the restore test's;
2. runs `pg_create_restore_point(name)` and keeps the LSN it returns;
3. runs `pg_switch_wal()` and waits until `pg_stat_archiver` shows that segment
   archived, so the point is restorable as soon as it is listed;
4. records the point (name, LSN, time, who, note) in `app-managed`, audited.

**The name is a label. The LSN is what a restore targets.** Postgres does not require
names to be unique, and recovery by name stops at the first one it meets. Targeting
the LSN is exact whatever is in the WAL. Deleting a point removes the record only; the
mark stays in the WAL, harmlessly.

### Retention and kept points

A point can be restored only while the base backup before it and every WAL segment
after it are kept. With the default retention of 7 full backups, a point made today
stops being restorable after about a week. Status gives each point that date.

**A point can be kept until a date**, up to 90 days by default, and the date is
required. Retention then keeps it restorable:

- `walg-retain` no longer runs `delete retain FULL <n>`. The engine computes the
  boundary: the older of
  - the n-th newest full backup, and
  - the full backup that starts before the oldest kept point;

  and runs `delete before FIND_FULL <that backup>`. WAL-G deletes WAL only before the
  oldest backup it keeps, so all WAL from the boundary on stays, and every kept point
  is reachable.
- **This costs storage**: every backup and all WAL from the boundary on. The form
  estimates it from the archive's growth per day before the point is kept. Status shows
  what the kept points hold back.
- When `KeepUntil` passes, the point stops counting, and the next run of retention
  moves the boundary forward.

For months or years, a `pg_dump` of the moment into a Kopia repository (phase 2) is far
smaller than holding back the archive.

**Before every deploy** (phase 2): an option on an app that links to a managed
database makes a point named `deploy-<app>-<deployment number>` before its pre-deploy
command. A migration that breaks data then has its way back already marked.

### The steps

1. **Pick a target.** The user picks one of the four targets above, inside the restore
   window, and a data volume for the new app.
2. **Create the new app.** The use case creates `<app>-restored-<yyyymmdd-hhmm>` from
   the same managed template, version and parameters, with a new prefix
   `walg/<new app key>`. It is not a logical child of the source: it must outlive it.
3. **Recover.** The new app's container starts through `walg-entrypoint`, with the
   source prefix and the target time in its environment. Since `PGDATA` is empty, it:
   - runs `backup-fetch` from the source prefix: the latest base backup before the
     target;
   - writes `recovery.signal`, `restore_command='walg-fetch %f %p'` (reading the
     **source** prefix), the target's settings from the table above, and
     `recovery_target_action='promote'`;
   - execs the official entrypoint.
4. **Archive to the new prefix.** WAL is not archived during recovery
   (`archive_mode=on`, not `always`). After promotion, the new timeline archives to the
   **new** prefix. The source prefix is only ever read. Together with
   `WALG_PREVENT_WAL_OVERWRITE`, that rules out two servers writing one archive.
5. **Take a first base backup.** A task waits for `pg_is_in_recovery()` to be false,
   then runs `walg-backup` once and marks the app ready. Recovery that ends before the
   target is reached fails the task with Postgres's own message.
6. **Switch.** The user checks the new app and repoints its consumers. Swapping names
   is Later.

The task's steps keep their state in the task (`Continue()`, as the job sequence
does). Each step can be run again safely.

As built:
- The source's archive reaches the new app as its own variables would, prefixed
  with `HIVEPAAS_RESTORE_`.
- Its keys, including its encryption key, are files of `/run/secrets/walg/restore`.
- The target is given as the backup to fetch and the recovery settings to write.
- The new app's engine jobs stay disabled until it has recovered.
- The first base backup is a run of its nightly job.
- Once it is taken, all of the restore's variables, keys and mounts are removed from
  the app.

## 5. The restore test

A backup nobody has restored is a hope. The weekly job proves the whole chain, up to
the moment it runs, without writing to the user's tables:

1. On the live database, run `pg_create_restore_point('hp-test-<ts>')`, then
   `pg_switch_wal()`, then wait until `pg_stat_archiver` shows that segment archived.
2. Run `backup-fetch` into a scratch directory on the data volume, after checking that
   free space is at least 1.2 × the database's size. Without the room, the job fails
   and says so.
3. Start a second postgres from it with:
   - `recovery_target_name='hp-test-<ts>'`;
   - `recovery_target_action='promote'`;
   - `archive_mode=off`, which is essential: otherwise the copy pushes into the real
     archive;
   - `port=5433` and `listen_addresses=''`.
4. Wait for promotion. If the chain is broken, Postgres stops with "recovery ended
   before configured recovery target was reached", and the job fails.
5. Run the checks: `pg_database` is readable, plus an optional query of the user's
   (`restoreCheckQuery`).
6. Stop the copy and remove the scratch directory. The job prints the restore's
   duration (the real recovery time), its size and the point it reached. Status reads
   these.

### Where it runs

The user chooses one of two places, at creation and later:

- **In the database's container.** Nothing more to run. It suits small databases. It
  takes the database's disk, CPU and memory while it runs, and it proves recovery on
  the same node.
- **In a verifier app.** The engine creates it on a node the user picks, preferably
  another one. It runs the same image and artifact, on its own volume. Its key is
  read-only where the storage allows one, and its scale is zero between runs, so it
  uses nothing in between. The job scales it up, runs `walg-restore-test` in it, and
  scales it down again. This **also proves that the database can be rebuilt after its
  node is lost**, and it keeps the test's load off production.

  The verifier runs no server of its own, so step 1 reaches the live database over
  the project's network, at the host in `HIVEPAAS_RESTORE_SOURCE_HOST`. Steps 2 to 6
  run in `HIVEPAAS_RESTORE_SCRATCH_DIR`, on the verifier's volume. A verifier needs
  an archive in object storage: it cannot reach one on a volume.

  The verifier is a logical child of the database app. It is deleted with it and not
  counted as an app.

**The container is the default.** A size threshold cannot pick for the user at
creation, because a new database has no size yet. Status suggests the verifier when
either of these turns true:

- the volume's free space falls under 1.5 × the database's size, so the next test
  would not fit;
- the last test took longer than a set time (default 30 minutes), so it weighs on
  production for that long.

The test itself refuses to start without the room (step 2).

## 6. Versions and updates

- **Postgres minors** move with the template's pinned image, through the existing
  image override. **Majors** do not move: going from 17 to 18 is "restore into a new
  app", by dump, until a `pg_upgrade` flow exists.
- **Artifacts** are recorded in `app-managed` when the app is created. Template
  updates (phase 2 of app templates) do not exist yet. Until they do, "Update tools"
  on a managed app reads only the `artifacts` of the template's current revision, which
  is an engine-owned field and safe to replace, then ensures and redeploys. The new
  artifact has a new hash, and so a new path, so the old file stays in place until the
  new task runs.

## 7. Testing

- **Schema:** `Validate` and lint on valid and broken fixtures (engine, `engineApi`, URL,
  hash, `alpine`). Compat CI with a release before `v000002`.
- **Agent:** `Ensure` against a local HTTP server. A good file; a wrong hash, which
  leaves nothing behind; a cut connection; two `Ensure` calls at once; a file already
  present.
- **Engine:** golden `Configure` output per major, the locked-field refusals per use
  case, and the job ownership refusals.
- **End to end, on the Linux server:**
  - Create PG17 and PG18 managed apps on a pinned volume, write rows, and wait for the
    archive.
  - Restore to a time between two writes into a new app, and find exactly the earlier
    rows.
  - Run the restore test.
  - Break the key and see archiver health fail and notify.
  - Fill a small volume and see the status turn `failing` before Postgres stops.

## 8. Paid templates

**A template says which feature it needs. A license says which features it grants.
The binary checks one against the other.**

```yaml
metadata:
  requires:
    versionCode: v000002
    features: [managed-databases]
```

- **Templates name features, not plans.** Which plan grants `managed-databases` is
  decided by the key. That way prices and plans can change without touching a
  template, and one feature can be sold in more than one plan.
- **The key** gains `features: [...]`, the way it carries `limits` today. A feature left
  out is not granted. Bee's features are compiled in, as its limits are.
- **Templates are signed**, so a template cannot be edited to drop its
  `requires.features`. In line with the licensing design, this is a fair-use gate, not
  DRM. `HP_TEMPLATES_DIR`, which is dev only, does not check it.
- **The store** shows such a template with its feature and the plans that grant it.
  The create form explains what is missing instead of failing at submit.
  `ERR_PLAN_FEATURE_REQUIRED` mirrors `ERR_PLAN_LIMIT_REACHED`.

**The licensing design's rule holds: a limit only ever stops creating.** Without the
feature, either on Bee or after a key's grace:

| Keeps working | Stops |
|---|---|
| Existing managed apps, archiving, base backups, retention, verify, health, restore tests, status, notifications | Creating a managed app from a paid template |
| **PITR into a new app**, and unmanage | Adding replicas or a verifier to an existing one |

Restoring is never gated. A license that lapses must not stand between a user and
their data. The app a restore creates is counted, and allowed, as part of the app it
restores.

Feature gating is generic: any template can name a feature. Managed databases are its
first use.

## 9. Clusters

The same engine extends to clusters. This makes the design's choices load-bearing:
the engine owns the topology, and the template only supplies images, artifacts and
defaults.

**The shape.** A managed cluster is one managed app, the primary, with members that
are its logical children. Each member is pinned to a different node by its own
volume.

- `app-managed` gains `Members []{AppID, Role, Node}` and a `Topology`.
- Template components (`MaxComponents = 30`) already create several apps from one
  template, but the members are added and removed after creation, so the engine
  creates them, not the template.

**What is easy, and comes first:**

1. **Async read replicas.** `pg_basebackup` from the primary, or seeded from the
   archive with `backup-fetch`, which keeps the load off the primary. Then streaming,
   with `restore_command` as the fallback when the primary has recycled WAL. The
   template `postgres-replica` already holds the replica side. The engine adds:
   - the primary's `pg_hba` entry;
   - an optional replication slot with `max_slot_wal_keep_size`;
   - lag in Status.
2. **Base backups from a replica.** `backup-push` runs on a standby. The nightly job
   then targets a replica and spares the primary.
3. **Manual promotion.** An audited action:
   - stop the old primary's service, which is the fence;
   - `pg_promote()` on the chosen replica;
   - move archiving to it, since only the primary archives;
   - repoint the other replicas;
   - move the stable address (below).

**The stable address.** Clients must reach "the primary" whichever member it is. Each
member's service has its own name.

1. **By default, a multi-host connection string.** The cluster shares
   `postgres://…@<m1>,<m2>,<m3>/<db>?target_session_attrs=read-write` with the apps
   that link to it, and `target_session_attrs=prefer-standby` for reads. The client
   library tries each host and keeps the one that accepts writes. There is nothing
   to run, nothing to move on promotion, and no hop in the path. libpq (Python's
   psycopg, Ruby's pg, PHP's pdo_pgsql) supports it since PG10, as do Go's `pgx` and
   JDBC (`targetServerType=primary`).

   It is safe only because promotion fences first: the old primary's service is
   stopped before the replica is promoted, so no client finds two servers accepting
   writes.
2. **For clients that cannot, a proxy.** node-postgres, for one, takes a single host.
   The cluster offers a `<cluster>-rw` service running PgBouncer, two replicas and
   stateless. The engine rewrites its config to point at the new primary on promotion,
   as a Swarm config, and it restarts in seconds. It is opt-in, because it adds a hop
   and a service to run.

**A moved network alias was considered and dropped.** In Swarm an alias belongs to a
service's network attachment. Moving it is a service update on both members, and
that restarts the newly promoted primary right after its promotion. While both
services hold the alias, its DNS answers with both of them.

**What is hard, and comes last:**

- **Synchronous replication** (`synchronous_standby_names = 'ANY 1 (...)'`). It is easy
  to turn on, and with fewer than two replicas it is a trap: one replica down blocks
  every commit. The engine allows it only with two or more.
- **Automatic failover.** The difficult part is not promoting, it is being sure the old
  primary is dead. HivePaaS can fence a primary whose node it reaches by stopping its
  service. It cannot fence one on a node it has lost, and promoting then risks two
  primaries. Doing this safely takes a consensus store and a fencing story:
  - **Patroni with etcd** is the proven one. It is Python plus a DCS, not a single
    binary, so it would be this design's first exception to "no image".
  - **pg_auto_failover** needs its extension installed into the image too.

  Either way it is its own design, after manual promotion has run in production.

**Postgres and WAL-G in a cluster:**

- **One archive per cluster**, not per member. Only the primary pushes, and the prefix
  follows the cluster. `wal-verify timeline` catches two members pushing.
- **PITR into a new app** restores a single primary. Its replicas are added afterwards.

## 10. Phases

| Phase | Work |
|---|---|
| 1 | `v000002` with `managed`, `artifacts` and `requires.features`; the key's `features`; agent `Ensure`, cache and cleanup; the hidden artifact mount; `app-managed`; the Postgres engine with its jobs, locks and status; S3, cloud-storage and volume archives, with their placement warning; the restore test, in the container or a verifier app; restore into a new app to the latest, a time, a restore point or a base backup; restore points, kept or not, and the retention that keeps them; `postgres-managed` for 17 and 18 with its scripts |
| 2 | A restore point before every deploy; unpinned volumes; scanning artifacts; swapping after a restore; async replicas, backups from a replica and manual promotion with the stable address; a nightly `pg_dump` into a Kopia repository alongside |
| 3 | MySQL and MariaDB: `xtrabackup` and `mariadb-backup` as artifacts, coupled to the server's version; Redis and Valkey; major upgrades; synchronous replication |
| Later | Automatic failover (its own design) |

## Decided

- HivePaaS builds no database images. Managed templates live in `app-templates`, and
  `release.json` stays for system apps.
- Postgres 17 and 18, Debian images only.
- **Scripts ship in the template.** A fix to a script is a template change, with
  `engineApi` as the contract the engine checks.
- **Artifacts are downloaded from their project's own release**, verified by hash, with
  an optional mirror tried first.
- **Locked fields are refused**, never re-applied silently.
- **The restore test runs in the database's container or in a verifier app**, as the
  user chooses.
- **WAL-G may write to a volume backup repository**, with a warning, before saving and
  in Status, when the archive is not independent of the database.
- **Paid features are declared by templates** (`requires.features`) and granted by the
  key. Restoring is never gated.
- PITR always restores into a new app, to one of four targets: the latest, a time, a
  restore point, or a base backup.
- **Restore points are named by users and restored by LSN.** A point is kept only
  until a date, and a kept point holds retention back for as long as it is kept.

- **`managed-databases` is granted by every paid plan**: Comb, Hive, Colony, Apiary
  and the trial. It is not granted by Bee.
- **A cluster's stable address is a multi-host connection string**, with PgBouncer as
  an opt-in proxy for clients that take a single host.
- **The restore test defaults to the container.** Status suggests the verifier from
  free space and test duration, not from a size fixed at creation.

## Open questions

None at this revision.

## Found along the way

- The app-templates README lists seven parameter types and at most three
  dependencies. The code has nine types (`domain`, `app`) and `MaxDependencies = 5`.
- The phase 3 TODO in `specmodel.BuildableBlocks` still lists config files and
  secrets, which are built.
- `base/system_job.go` declares `SystemJobName` and nothing uses it.
- Restores today go only into an existing app (`backupsnapshotuc/restore.go`). §4
  is the first path that creates one.
