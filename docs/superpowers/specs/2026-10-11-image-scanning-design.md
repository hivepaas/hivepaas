# Scanning the images apps run

Status: phase 1 approved 2026-10-11, being built.

## Why

HivePaaS scans its own images in CI (Trivy in `release.yml`, `kopia.yml`,
`go.yml`, with `.trivyignore`), but nothing looks at the images the apps on an
installation run. A Debian base built a year ago, an `npm` tree locked to old
versions, a template's database image left on a tag no one bumps: each carries
known vulnerabilities, often with a fix one rebuild away, and nobody sees them.

Phase 1 makes them seen. It finds what each app's image carries, says what can
be fixed, and lets a team set aside what does not apply to them, with a reason
and an end date. It blocks nothing: a deployment goes ahead whatever a scan
finds. A gate is phase 3.

## What was measured

On the dev machine's Docker VM, 2026-10-10 and 2026-10-11.

**Trivy 0.75.0 on its own** (`trivy image`, standalone, DB local):

| Image | Size | Cold scan | Peak RAM |
|---|---|---|---|
| dozzle | 68 MB | 0.95 s | 182 MiB |
| rabbitmq (alpine) | 169 MB | 2.3 s | 209 MiB |
| an app, p1-f1 | 254 MB | 3.7 s | 133 MiB |
| hivepaas-devtools | 773 MB | 12.6 s | 289 MiB |

Warm (layers already analyzed): 0.11 - 0.15 s, 110 - 125 MiB. CPU while cold:
25 - 40 % of a core. The app p1-f1 had 212 findings, 108 distinct IDs, 54 HIGH,
49 with a fix, in 48 packages (Debian and `node-pkg`).

**Trivy client and server**, the shape this design uses:

- The image is 183 MB; its `trivy` binary alone is 160 MB.
- The server downloaded its DB (from `mirror.gcr.io/aquasec/trivy-db:2`) and
  was listening after 19 s; the DB is 1.4 GB on disk. Idle, and while serving
  scans, it held **72 - 82 MiB**: the DB is memory-mapped.
- A client scan, in a fresh container each time (no cache of its own), times
  including the container's start: the table below.
- **The layer cache is the server's.** The client asks the server which layers
  it lacks and analyzes only those; a second client on a fresh container was
  warm. For the two images above the cache was 480 KB.
- **The client copies the image to `/tmp`.** With `--image-src docker`, Trivy
  exports the whole image to a temporary tar before reading it: a read-only
  container with a 64 MB tmpfs failed with `no space left on device`. A scan
  needs free disk as large as the image, for as long as it runs. That tar is
  also why the cgroup's `memory.peak` read 958 MiB with no limit: page cache,
  given back under a limit, as the 256 MiB run shows.

| Image | Cold | Again (server knows the layers) |
|---|---|---|
| hivepaas-agent-dev, 412 MB | 5.0 s | 0.34 s |
| hivepaas-devtools, 773 MB | 14.4 s; 17.1 s at 1 CPU / 512 MiB; 19.3 s at 1 CPU / 256 MiB; **killed at 128 MiB** | 0.39 s, ~40 MiB |

**OSV-Scanner 2.6.0**: a 315 MB image, 3 s on p1-f1, 83 distinct IDs in 20
packages, against Trivy's 108 in 48. It sends each image's package list to
osv.dev unless run offline with a local DB per ecosystem, and it exits 1
whenever it finds anything.

**So: Trivy.** It is what CI already uses, it works offline from a DB it
downloads, and in client and server mode one small server holds the DB and the
layer cache for the whole cluster while each node only reads its own images.

## The shape

```text
                    hivepaas (app or worker)
                 │ scan app A: its task's container on node N → image sha256:…
                 ▼  (docker inspect, through node N's agent)
     runner on node N  ◄── exec `trivy image --server …` (1 CPU / 512 MiB, the runner's)
     │ /var/run/docker.sock: reads the image
     │ the env's overlay network
     ▼
     hivepaas_scanning_trivy (trivy server: DB + layer cache)
```

Built as two system apps in env `scanning` of the hidden `hivepaas` project,
both on the release's Trivy image (`trivyImage` in release.json, pinned by
digest, moved by the system updater with the other system apps): Trivy's
server, and a runner, one task on every node, that the scans run in.

### The server: a system app

An addon like logging and the registry: off until an administrator turns it
on (System > Image Scanning). `systemappservice.Provision` creates app `trivy`,
one replica, from a document built in Go (`imagescanserviceimpl/appdoc.go`):

- `sh -c 'exec trivy server --listen 0.0.0.0:4954 --cache-dir /data --token
  "$(cat /run/hivepaas/trivy-token)"'`, plus `--db-repository` for a mirror;
- the token a secret of the app, mounted as a file: one in an argument or an
  environment variable would be in the service's spec;
- a volume of its node, `hivepaas-trivy-data`, at `/data`: the DB and the
  layer cache. A server moved to another node fetches the DB again;
- limits 1 CPU, 1 GiB;
- no route: the runner reaches it as `trivy`, its key, on the env's network.

Trivy's server keeps its DB current by itself: it checks every hour and holds
requests while it swaps the new DB in. Upstream rebuilds the DB every 6 hours.

The DB's source is a setting, `dbRepository`, empty for Trivy's default. An
installation without internet pushes the DB to its own registry (the internal
zot will do) with `oras` and points `dbRepository` there; the website's page
says how.

Turning it off removes both apps. The volumes and the results stay; the
results are shown as last scanned at their time.

### The scan: a command run in the runner

The `trivy` binary is 160 MB, and the agent image 412 MB: bundling it, as
kopia is, would make every installation pull it whether it scans or not. A
container started per scan was the first plan, and would not keep: the daily
system cleanup prunes the nodes' unused images and anonymous volumes, so each
node would pull the 183 MB image again every day. So the runner is an app of
its own, `trivy-runner`, global: an idle shell on every node, which keeps the
image and the Java DB's volume in use, and which every scan of the node is a
command run in (`containerexecservice`, through the node's agent - no new
agent RPC).

- Its Docker socket is the app's Docker API access in host mode, a setting
  written on it when it is made: every deployment takes a socket away from an
  app whose access does not give one, and a template's document may not ask
  for host mode, so the access is the way a socket survives.
- Its volume `hivepaas-image-scan-cache` holds the Java DB, at Trivy's cache
  dir.
- Its limits, 1 CPU and 512 MiB by default (256 measured enough for 773 MB;
  128 was killed), are every scan's; the settings screen sets them.

A scan of app A, by `imagescanservice.ScanApp`:

1. The running tasks of A's service, by node; on each node, `docker inspect
   --format '{{.Image}} {{.Config.Image}}'` of their containers, through the
   node's agent, gives the image IDs and references. One node per image ID.
2. For each image, unless scanned within the hour (Scan now always scans): a
   Redis lock on the node, one scan at a time there; then, in the node's
   runner, `trivy image --server http://trivy:4954 --token "$(cat …)"
   --image-src docker --scanners vuln --list-all-pkgs=false --disable-telemetry
   --skip-version-check --no-progress --format json --timeout 10m [--java-db-repository …]
   [--skip-java-db-update] <image ID>`.
3. Its stdout is the report, parsed into the image's scan and findings and
   written in a transaction of its own (`ReplaceScan`); its stderr is the
   task's log. A scan that fails keeps the image's last findings and says why.

A scan that runs out of disk fails with Trivy's error, shown as the scan's.

Checked before building (2026-10-11):

- **Trivy takes an image ID** with `--image-src docker`: the report's
  `Metadata.ImageID` is the ID given, with the image's tags and repo digests.
- **The Java DB is the client's.** The jar analyzer runs where the layers are,
  and needs Trivy's Java DB on the first image with a jar: a 935 MiB download,
  1.4 GB on disk, about 2.5 minutes. `--offline-scan` does not avoid it, and
  `--skip-java-db-update` fails without one ("The first run cannot skip
  downloading Java DB"). So each node keeps it in a volume,
  `hivepaas-image-scan-cache`, mounted at the client's cache dir; it is fetched
  only on a node that runs an image with jars. Trivy renews it after 3 days;
  the scan passes `--skip-java-db-update` while the node's copy is under 7 days
  old, so a node fetches it once a week at most. `javaDbRepository` points it
  at a mirror, as `dbRepository` does the DB.
- **The server's DB swap** fits: downloading the DB was killed at 256 MiB, and
  took 22 - 24 s at 512 and 768 MiB. The 1 GiB limit stands.
- The scan also passes `--list-all-pkgs=false` (the report is findings only:
  5 KB for `alpine:3.10`, 41 KB for one log4j jar), `--disable-telemetry`,
  `--skip-version-check` and `--no-progress`. A wrong token fails with
  `twirp error unauthenticated: invalid token`.

### Why not the alternatives

- **Scan from the registry** (`--image-src remote`, at the server): no work on
  the nodes, but a local-only image, built on a single node with no registry,
  could not be scanned, and the server would need every registry's credentials,
  ECR tokens included.
- **A Swarm job per scan**: the result would have to be gathered from the job's
  logs, and placement on the node that has the image is what the agent already
  does.

## What is scanned

The image each app's running tasks run, by **image ID** as the node's Docker
gives it: the config digest with Docker's own image store, the manifest's with
the containerd store (as in dind). Services HivePaaS creates do not carry a
digest, since they deploy with `QueryRegistry` off; the container says what it
was made from. Every kind of app in a project: built from a repo, from an
image, a function, from compose, from a template. An app whose tasks run two
images (two platforms, a rollout) has both scanned.

Not in phase 1: HivePaaS's own apps. CI scans HivePaaS's images; Traefik, zot
and the logging images are phase 2. Scan now on one is refused.

An app with no running task is not scanned: what it ran last stays shown, with
why it was not scanned again.

## When

1. **After a deployment that ended done.** The deploy task's executor queues
   the app's scan once the deployment committed (the deployment service cannot
   depend on the scanner: the scanner provisions apps, which deploys them). It
   never delays or fails the deployment.
2. **Every day.** A system job, `Image scan`, at 02:00 in the installation's
   timezone (after backup repo cleanup at 01:30, in `dailyJobs`), scans again
   every image apps run: the DB has news an image did not have yesterday.
   Nearly all of it is warm, under half a second an image. It also removes the
   results of images no app has run for 30 days.
3. **When asked:** Scan now, on the app's Security tab.

Scans 1 and 3 are tasks, `task:image-scan`, the app's: they show in its tasks
with Trivy's log, retry once after 30 s, and time out after 15 minutes. A scan
asked while one of the app's is waiting to start is that one. The daily scan is
a scheduled job (`SchedJobTypeImageScan`, made with the settings and on while
scanning is), its run a task of its own. An image scanned in the last hour is
not scanned again by 1 or 2; 3 always scans.

`SchedJobTypeImageScan` is a variable, not a constant of its type, and is not
in `AllSchedJobTypes`: swag lists a type's constants as the values a request
may send, and the job type no client may create would raise the API level -
refusing every older CLI's writes - for nothing.

## What is kept

Results belong to an image, not to an app: two apps on the same image share
them.

```sql
create table image_scans (
  image_id      varchar(100) primary key, -- sha256:…
  image_ref     varchar not null,         -- the reference last seen, for display
  platform      varchar(50),              -- linux/amd64
  os_family     varchar(50), os_name varchar(100),
  os_eosl       boolean not null default false, -- the OS is past its support
  status        varchar(20) not null,     -- done | failed
  error         varchar,
  trivy_version varchar(50),
  db_updated_at timestamptz,              -- the DB the scan was matched against
  scanned_at    timestamptz not null,
  updated_at    timestamptz not null default current_timestamp
);

create table image_scan_findings (
  image_id          varchar(100) not null references image_scans (image_id) on delete cascade,
  vuln_id           varchar(100) not null, -- CVE-…, GHSA-…
  pkg_name          varchar(255) not null,
  pkg_type          varchar(50)  not null, -- debian, node-pkg, gobinary, jar…
  pkg_path          varchar      not null default '',
  installed_version varchar(255) not null,
  fixed_version     varchar,               -- null: no fix yet
  severity          varchar(20)  not null, -- CRITICAL | HIGH | MEDIUM | LOW | UNKNOWN
  title             varchar,
  url               varchar,               -- Trivy's PrimaryURL
  published_at      timestamptz,
  primary key (image_id, vuln_id, pkg_name, pkg_path, installed_version)
);

-- What each app's tasks ran when it was last scanned.
create table app_image_scans (
  app_id     varchar(100) primary key,
  image_ids  varchar(100)[] not null default '{}',
  error      varchar,                     -- why none could be scanned
  scanned_at timestamptz not null,
  reported   varchar[]                    -- what the daily scan last reported
);
```

A scan replaces its image's findings in one transaction; one that fails marks
the image's scan failed and keeps its findings. The report itself is
not kept: rows answer what phase 1 asks ("what does this app carry", "which
apps carry CVE-X"), and a description is one link away. About 250 bytes a
finding: 200 apps with p1-f1's 212 findings is about 10 MB.

## Setting findings aside

A team decides a finding does not apply to them - the package is in the image,
its code never runs. That is a setting, `vulnerability-ignore`, at any scope a
secret can have (global, project, env, app) and inherited down the same way:

```json
{ "vulnId": "CVE-2025-1234", "package": "libxml2",
  "reason": "the parser is never fed outside input", "expireAt": "2027-01-11T00:00:00Z",
  "inheritable": true }
```

- `reason` is required; `package` is optional (all packages when empty);
  `expireAt` is optional and at most a year away. It is the setting's own
  `expireAt`, which settings already turn expired by. The name is derived -
  `CVE-2025-1234 (libxml2)` - so one scope sets one vulnerability aside once.
  Who wrote it is kept with it, by username.
- It applies when results are read, never at scan time: adding or removing one
  needs no new scan, and an expired one stops applying by itself.
- Writing one is a write on its scope (global: the `vulnerability-ignore`
  resource of the settings module), audited like any setting. The nearest
  scope's ignore is the one shown when several apply.
- Spec export and import skip it, and the image scan settings, in phase 1.

## What is shown

Counts are after ignores. **Fixable** means a fixed version exists. The headline
is fixable CRITICAL and HIGH.

- **The app's Security tab**, after Metrics:
  - the image (reference, short ID, platform), the OS, and a warning when the OS
    is past its support; when it was scanned, its status, and Scan now;
  - counts by severity, each with its fixable part;
  - the findings by package: installed and fixed version, the IDs (each a link),
    the highest severity; Fixable only on by default; Ignore… on a row opens a
    dialog (scope, reason, until);
  - Ignored (n), folded: each with its reason, who, until, and Remove.
- **The app's header**: a shield beside the status, red with a count when there
  is a fixable CRITICAL, amber for fixable HIGH only, nothing otherwise; grey
  for not scanned or failed.
- **Home > Needs attention**: a new kind, `app-vulnerable` (warning), for an app
  with a fixable CRITICAL. Only CRITICAL, so the list does not fill with HIGHs.
- **System > Image Scanning**, beside Logging and Registry, three tabs:
  - Configuration: on/off, DB source, the scan's CPU and memory, the daily job's
    schedule and notification;
  - Images: every app's image, its project and env, CRITICAL and HIGH with their
    fixable part, last scanned, status;
  - Actions: the server's state, its DB's date, Run the daily scan now.

## Notification

The daily job's result notification, through the channels it is given (default
ones by default), as the system cleanup's is. It is sent only when a run found
something new: an app with a fixable CRITICAL or HIGH that its previous scan did
not have, after ignores. It lists them by app. New template
`image_scan_notification` for email, Slack, Discord, Telegram and Lark.

A scan after a deployment notifies nobody in phase 1: the app's header and the
attention list show it at once, and the next daily run reports it.

## API

Read through the usecase layer, DTOs per ARCHITECTURE.md §2. All new
operations: the API level is unchanged.

| Operation | What |
|---|---|
| `GET` / `PUT /system/settings/image-scan` | the addon and the daily scan: `enabled`, `dbRepository`, `javaDbRepository`, `schedule`, `notification`, `cpuLimit` and `memoryLimit` (the runner's); GET adds `scannerStatus` - server running, runner nodes, image, the DB's date, the last scan |
| `POST /system/settings/image-scan/exec` | run the daily scan now: its task |
| `GET /system/settings/image-scan/apps` | every scanned app, its images and counts, the worst first (paged) |
| `GET /projects/:projectID/:projectEnv/apps/:appID/image-scan` | the Security tab: images, counts, findings, those ignored with their ignore, the scan pending |
| `POST /projects/:projectID/:projectEnv/apps/:appID/image-scan` | Scan now: its task |
| `…/vulnerability-ignores` at `/settings`, `/projects/:projectID`, `/projects/:projectID/:projectEnv`, `…/apps/:appID` | ignores, as each scope's settings are |
| `GET /home/attention` | gains `app-vulnerable`, with `criticalFixable` |

Permissions: the addon and the apps list, the `image-scan` resource of the
system module; reading an app's results, reading the app; Scan now, writing
it.

## Building it, in order

1. Check what this design assumes (done, above).
2. Entities, migration, the `vulnerability-ignore` setting.
3. `imagescanservice`: the addon (Apply, Sync, Status), the scan, the daily
   run, its notification; the attention kind.
4. API, swagger.
5. Dashboard: System > Image Scanning, the Security tab, the badge, the
   attention kind.
6. Website: a page on image scanning, its air-gapped DB.
7. e2e, level 2, in the dind env (it needs the internet for the DB): turn it on;
   deploy an image with known fixable findings and an EOSL OS (`alpine:3.10`:
   CVE-2021-36159 in apk-tools, CRITICAL, fixed in 2.10.7-r0); the tab and badge
   show them; ignore one and the count drops; Scan now; the daily run; turn it
   off and the results stay.

Tried on the dind env (2026-10-11), through the API: switched on, both apps
ran as designed - the runner global, with the host socket and the cache
volume; the server fetched its DB. A deployment of `alpine:3.10` queued its
scan; Scan now found CVE-2021-36159 and the EOSL OS in 0.3 s; ignores at the
app and at the project set it aside; the daily run skipped the image scanned
within the hour; switched off, both apps went and the results stayed.

## Decided (2026-10-11)

1. Ignores at all four scopes: global, project, env, app.
2. The attention list for fixable CRITICAL only.
3. No notification from a scan after a deployment in phase 1.
4. HivePaaS's own apps left to phase 2.

## Later

- **Phase 2**: the CLI (`hivepaas app scan`, `hivepaas app vulns`) and MCP; a
  project's overview; HivePaaS's own apps; ignores in spec export and import; an
  SBOM to download (CycloneDX); secrets and misconfiguration scanners.
- **Phase 3**: a gate, per project: a deployment whose image has a fixable
  CRITICAL is refused, or held for approval; a scan of a built image before it
  rolls out.
