# Scanning the images apps run

Status: phase 1 approved 2026-10-11 and built with a Trivy server and a runner
on every node (ec854198, never released). Redesigned 2026-10-14, this version,
and rebuilt the same day: no service runs while nothing is scanned; Trivy's data
is on a volume the administrator chooses.

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

## Why it was redesigned

The first build ran Trivy's server (one replica, its DB on a volume of its
node) and a runner on every node (an idle shell with the node's Docker socket
and its own copy of the Java DB, that scans were commands run in). Against it:

- **A runner on every node, for ever.** A container with the node's Docker
  socket, idle on every node of the cluster whether anything is scanned or not.
- **The Java DB on every node.** Trivy reads jars where the layers are, so in
  client and server mode the Java DB (935 MiB to download, 1.4 GB on disk) is the
  client's: every node running a Java image kept its own.
- **The server's data was a volume of whichever node it ran on.** Moved to
  another node, it fetched its DB again and left the old volume behind.

The runner existed because the daily cleanup was thought to prune unused images
daily, so a container started per scan would have pulled Trivy's image again
every day. The cleanup in fact pruned only untagged images; since 317b9314 it
removes images left unused for its retention and always keeps the running
release's, Trivy's included. A container per scan is now cheap.

## What was measured

On the dev machine's Docker VM, 2026-10-10 and 2026-10-11.

**Trivy 0.75.0 standalone** (`trivy image`, DB local) - the shape this design
uses:

| Image | Size | Cold scan | Peak RAM |
|---|---|---|---|
| dozzle | 68 MB | 0.95 s | 182 MiB |
| rabbitmq (alpine) | 169 MB | 2.3 s | 209 MiB |
| an app, p1-f1 | 254 MB | 3.7 s | 133 MiB |
| hivepaas-devtools | 773 MB | 12.6 s | 289 MiB |

Warm (layers already analyzed): 0.11 - 0.15 s, 110 - 125 MiB. CPU while cold:
25 - 40 % of a core. The app p1-f1 had 212 findings, 108 distinct IDs, 54 HIGH,
49 with a fix, in 48 packages (Debian and `node-pkg`).

From the client and server build, still true standalone:

- The image is 183 MB; its `trivy` binary alone is 160 MB.
- **The DB** (from `mirror.gcr.io/aquasec/trivy-db:2`) is 1.4 GB on disk.
  Downloading it was killed at 256 MiB and took 22 - 24 s at 512 MiB.
- **The Java DB** is needed on the first image with a jar: a 935 MiB download,
  1.4 GB on disk, about 2.5 minutes. `--skip-java-db-update` fails without one
  ("The first run cannot skip downloading Java DB").
- **With `--image-src docker`, Trivy copies the image to `/tmp`** before reading
  it: a scan needs free disk as large as the image while it runs. A scan of the
  773 MB image ran at 256 MiB and was killed at 128 MiB.
- **Trivy takes an image ID** with `--image-src docker`: the report's
  `Metadata.ImageID` is the ID given, with the image's tags and repo digests.
- `--list-all-pkgs=false` keeps the report to findings: 5 KB for `alpine:3.10`,
  41 KB for one log4j jar.

**OSV-Scanner 2.6.0**: a 315 MB image, 3 s on p1-f1, 83 distinct IDs in 20
packages, against Trivy's 108 in 48. It sends each image's package list to
osv.dev unless run offline with a local DB per ecosystem, and it exits 1
whenever it finds anything. **So: Trivy**, which CI already uses and which works
offline from a DB it downloads.

## The shape

```text
hivepaas (app or worker)
  │ scan app A: its task on node N runs image sha256:…   (docker inspect, through N's agent)
  │ the volume decides the scan's node S: its own node, or N when it is shared
  ▼
agent on node S ── ImageScanRun ──► container hivepaas-cont-…, image trivyImage
                                     │ /data            the chosen volume: DB, Java DB, layer cache
                                     │ docker.sock      only when the image is on S
                                     │ registry auth    a file, only when it is not
                                     ▼
                                     stdout: the report, back over the RPC
```

Nothing runs between scans. A scan is a container the scan node's agent starts,
on the release's Trivy image (`trivyImage` in release.json, pinned by digest),
and removes when it ends.

### Where Trivy's data is

Switching scanning on requires a **volume**: a cluster volume, chosen as
Logging's and the Registry's are. Trivy keeps everything there, under
`hivepaas/scanning/trivy/` - where a system app of env `scanning` would keep its
data - so the volume can hold other things: the DB, the Java DB, and the layer
cache - what lets an image already scanned be scanned again in under half
a second.

The volume's pin decides where scans run:

| The volume | Scans run on | The image is read |
|---|---|---|
| **Pinned to a node** (a node, the current node) | that node | from that node's Docker when it is there; otherwise from its registry |
| **Pinned by a node label** | a node carrying the label: the one running the image if one does, else the first by ID | the same |
| **All nodes** (shared storage, such as NFS) | the node running the image | from that node's Docker |

- A **Tmpfs** volume is refused: it is memory, emptied when each scan's
  container stops, so every scan would download about 3 GB.
- A volume pinned by label on several nodes is a different directory on each:
  each of them that scans keeps its own copy. The settings screen says so.
- **All nodes** is the volume's own claim, as it is for apps: HivePaaS cannot
  tell shared storage from a node's disk, and the volumes page already warns
  about it.
- The volume can be changed at any time. Nothing is copied: the next scan
  downloads the DBs to the new one. The old volume is left as it is.
- The image scan settings reference the volume while scanning is on, so it
  cannot be deleted while scans use it. Switched off, the volume stays chosen
  but may go: switching on again asks for one that is there.

### Reading an image from its registry

With a volume pinned to node S, an app on node N runs an image S may not have.
The scan then reads it from its registry, with `--image-src remote`, by digest,
so what is scanned is exactly what N runs:

- the reference is the repository and digest the image on N carries: its ID
  with the containerd image store, its repo digest for that repository with
  Docker's own store;
- `--platform` is N's (`os/arch`), so a multi-platform image is read for the
  node that runs it. No variant: Docker 29's image inspect leaves it out when
  an image has none, and a template naming it fails; `linux/arm64` reads
  arm64's v8;
- the credentials are those the app's deployment pulls with (its image
  source's registry auth, or the registry its builds are pushed to), written to
  a Docker `config.json` copied into the container before it starts, never in
  an argument or the environment, where the container's inspect would show
  them; ECR's are the current token HivePaaS renews;
- Trivy asks its layer cache which layers it lacks and downloads only those, so
  an image is downloaded once, not at every daily scan.

An image whose repository has no digest on N - built there and never pushed,
or loaded by hand - can only be read on N. On a single node S is N. In a
cluster of more than one node, an app's builds are pushed to a registry
(HivePaaS refuses otherwise), so this is an image someone put on N by hand: it
is not scanned, and the app's tab says why ("on node N only, in no registry").

### The scan's container

`ImageScanRun`, a new agent RPC, as `FunctionTestRun` is: the scan node's agent
pulls Trivy's image when it is missing, creates the container, copies the
registry auth into it, starts it, streams back its stdout - the report - and
gives its stderr as the log, and removes it.

- `trivy image --cache-dir /data/trivy --scanners vuln --list-all-pkgs=false
  --disable-telemetry --skip-version-check --no-progress --format json
  --timeout 10m [--db-repository …] [--java-db-repository …]
  [--skip-java-db-update] (--image-src docker <image ID> | --image-src remote
  --platform … <repository@digest>)`.
- The volume at `/data`, mounted as an app's would be (a managed volume carries
  its driver and options, so the node builds it when it does not have it).
- The Docker socket only for `--image-src docker`; the copy of the image Trivy
  makes goes to the container's own `/tmp`, on the node's disk.
- Limits: 1 CPU and 512 MiB by default, set on the settings screen. 512 is what
  a DB download takes; a scan alone took 256 for a 773 MB image.
- Labeled a temporary resource: the cleanup removes one left behind by an agent
  that died mid-scan.
- Its logging driver is `none`, and what it prints is read by attaching to it
  before it starts: the collector of stored logs reads every container's
  json-file log, and the report would land in it.

**The DBs.** Trivy downloads its DB when the copy on the volume is past the
next update it announces (upstream rebuilds it every 6 hours), so the first
scan of a day may take 20 - 30 s more. The Java DB is downloaded by the first
scan that meets a jar; the scan passes `--skip-java-db-update` while the copy
is under 7 days old (Trivy would renew it after 3), so it is downloaded once a
week at most. `dbRepository` and `javaDbRepository` point them at a mirror: an
installation without internet pushes both to its own registry with `oras`; the
website's page says how.

**One scan at a time per volume.** Trivy's cache is a bbolt file one process
holds at a time, and on shared storage two nodes opening it at once is asking
for corruption. Each scan takes a Redis lock named after the volume: on a pinned
volume that serializes a node, on a shared one the cluster. Scans are short and
few - after a deployment, and once a day - so the wait is seconds.

### Why not the alternatives

- **The server and the runners** (the first build): see "Why it was
  redesigned".
- **Always from the registry**, on the volume's node: no socket at all, but a
  single node's local builds are in no registry, and every image would be
  downloaded once even when the volume is shared and the node has it.
- **A Swarm job per scan**: its result would have to be gathered from its logs,
  and placing it is what the agent already does.

## What is scanned

The image each app's running tasks run, by **image ID** as the node's Docker
gives it: the config digest with Docker's own image store, the manifest's with
the containerd store (as in dind). Services HivePaaS creates do not carry a
digest, since they deploy with `QueryRegistry` off; the container says what it
was made from. Every kind of app in a project: built from a repo, from an
image, a function, from compose, from a template. An app whose tasks run two
images (two platforms, a rollout) has both scanned; an image several nodes run
is scanned once.

Not in phase 1: HivePaaS's own apps. CI scans HivePaaS's images; Traefik, zot
and the logging images are phase 2. Scan now on one is refused.

An app with no running task is not scanned: what it ran last stays shown, with
why it was not scanned again.

## When

1. **After a deployment that ended done.** The deploy task's executor queues
   the app's scan once the deployment committed. It never delays or fails the
   deployment.
2. **Every day.** A system job, `Image scan`, at 02:00 in the installation's
   timezone (after backup repo cleanup at 01:30, in `dailyJobs`), scans again
   every image apps run: the DB has news an image did not have yesterday.
   Nearly all of it is warm: under half a second an image, plus the container's
   start. It also removes the results of images no app has run for 30 days.
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

Unchanged from the first build. A scan replaces its image's findings in one
transaction; one that fails marks the image's scan failed and keeps its
findings. The report itself is not kept: rows answer what phase 1 asks ("what
does this app carry", "which apps carry CVE-X"), and a description is one link
away. About 250 bytes a finding: 200 apps with p1-f1's 212 findings is about
10 MB.

## Setting findings aside

Unchanged. A team decides a finding does not apply to them - the package is in
the image, its code never runs. That is a setting, `vulnerability-ignore`, at
any scope a secret can have (global, project, env, app) and inherited down the
same way:

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
    is past its support; when it was scanned, where it was read (on its node,
    or from its registry), its status, and Scan now;
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
  - Configuration: on/off, **the volume** (required; the picker leaves Tmpfs
    out, and says where scans will run for the one chosen), DB sources, the
    scan's CPU and memory, the daily job's schedule and notification;
  - Apps: every app's image, its project and env, CRITICAL and HIGH with their
    fixable part, last scanned, status;
  - Actions: the volume and where scans run, Trivy's image, the DB's date (the
    last scan's), the last scan and its error, and Run the daily scan now.

## Notification

Unchanged. The daily job's result notification, through the channels it is
given (default ones by default), as the system cleanup's is. It is sent only
when a run found something new: an app with a fixable CRITICAL or HIGH that its
previous scan did not have, after ignores. It lists them by app, with the
`image_scan_notification` template for email, Slack, Discord, Telegram and
Lark.

A scan after a deployment notifies nobody in phase 1: the app's header and the
attention list show it at once, and the next daily run reports it.

## API

Read through the usecase layer, DTOs per ARCHITECTURE.md §2. All new
operations: the API level is unchanged.

| Operation | What |
|---|---|
| `GET` / `PUT /system/settings/image-scan` | `enabled`, **`volume`** (required when enabled), `dbRepository`, `javaDbRepository`, `schedule`, `notification`, `cpuLimit` and `memoryLimit` (the scan container's); GET adds `scannerStatus`: the volume's name and pin, the nodes scans run on, Trivy's image, the DB's date, the last scan and its error |
| `POST /system/settings/image-scan/exec` | run the daily scan now: its task |
| `GET /system/settings/image-scan/apps` | every scanned app, its images and counts, the worst first (paged) |
| `GET /projects/:projectID/:projectEnv/apps/:appID/image-scan` | the Security tab: images, counts, findings, those ignored with their ignore, the scan pending |
| `POST /projects/:projectID/:projectEnv/apps/:appID/image-scan` | Scan now: its task |
| `…/vulnerability-ignores` at `/settings`, `/projects/:projectID`, `/projects/:projectID/:projectEnv`, `…/apps/:appID` | ignores, as each scope's settings are |
| `GET /home/attention` | gains `app-vulnerable`, with `criticalFixable` |

Permissions: the addon and the apps list, the `image-scan` resource of the
system module; reading an app's results, reading the app; Scan now, writing
it.

## Rebuilding it, in order

Done 2026-10-14, in this order. None of the first build was released, so
nothing is migrated: what it made on the dind env - apps `trivy` and
`trivy-runner`, volumes `hivepaas-trivy-data` and `hivepaas-image-scan-cache` -
was removed by hand. Its settings request changed, but the operation was never
released either: its fingerprint was dropped from `api-level.json` so that the
change counts as the operation's addition, and the API level stays 2.

1. **Check what this design assumes.** Done on the dind env (Docker 29.9,
   containerd store, arm64), 2026-10-14, Trivy 0.75.0, a fresh volume:
   - **Standalone, on a volume**: the first scan of `alpine:3.10` downloaded the
     DB and scanned in 25 s; it passed at 512, 384 and **256 MiB** (the server's
     DB swap had been killed at 256). 512 stays the default.
   - **The DB's date is not in the standalone report** (`Trivy` holds only
     `Version`). `trivy version --format json --cache-dir …` gives
     `VulnerabilityDB.UpdatedAt` (and `JavaDB` once there): the container runs
     it after the scan, and prints both in one JSON object on stdout.
   - **`--image-src remote` by digest** took 3 s and 36 MiB (no copy of the
     image is made); the second scan logged no missing layer and downloaded
     nothing. **`--platform` is needed every time**: without it Trivy reads
     `linux/amd64` even on an arm64 node, and an arm64-only index fails ("no
     child with platform linux/amd64").
   - **The report's `Metadata.ImageID`, read remotely, is the config's digest**,
     not the ID the node gives (the index's, with the containerd store): results
     are keyed by the node's ID, as given to the scan.
   - **Credentials**: a password-protected registry refused the scan
     (`UNAUTHORIZED`), then accepted it with a `config.json` copied into the
     container, as a directory under `/tmp`, before it started, and
     `DOCKER_CONFIG` pointing there. A plain-HTTP registry needs `--insecure`:
     passed when the node's Docker lists the registry as insecure. The internal
     registry is HTTPS, on its domain.
   - **A second Trivy on the same cache fails** after 5 s: "cache may be in use
     by another process: timeout". The lock per volume is what prevents it.
   - A warm scan, container started and removed: 2 - 3 s.
2. **Take out** the server and the runner: their documents (`appdoc.go`), the
   system apps' provisioning and sync, the token, the host access, env
   `scanning`, and the system updater's steps for them. `trivyImage` stays in
   release.json; the updater no longer pre-pulls it - the agent pulls it when a
   scan needs it, and the cleanup keeps it.
3. **Settings**: `volume`, validated (exists, not Tmpfs), referenced; the status
   as above.
4. **`ImageScanRun`**: the agent RPC and the container.
5. **The scan**: the scan node from the volume, the source (`docker` or
   `remote`), the reference, platform and credentials, the lock per volume.
6. API, swagger.
7. Dashboard: the volume on Configuration, Actions' status; the Security tab
   says where an image was read. The rest stays.
8. Website: the image scanning page: the volume, where scans run, the
   registry read, the air-gapped DBs.
9. e2e, level 2 (rows 2.121 - 2.123), in the dind env: the test makes a volume
   pinned to the node; the rest as before - `alpine:3.10` (CVE-2021-36159 in
   apk-tools, CRITICAL, fixed in 2.10.7-r0, and an EOSL OS), ignore, Scan now,
   the daily run, switched off with the results kept. They pass in 2.2 minutes
   (7 with the server and the runners).

Tried on the dind env, through the API: switched on without a volume, and with
a Tmpfs one, refused with the reason; with one pinned to the node, Scan now
started the container through the node's agent, which fetched the DB onto the
volume (`hivepaas/scanning/trivy/db`, 1.5 GB) and scanned in 20 s; the next
scan took 0.6 s; no container was left; the volume could not be deleted while
scanning was on, and could once it was off. Read from a registry - a node
without the image - is covered by the measurements above and by unit tests: a
single-node env always has the image it runs.

## Decided

2026-10-11:

1. Ignores at all four scopes: global, project, env, app.
2. The attention list for fixable CRITICAL only.
3. No notification from a scan after a deployment in phase 1.
4. HivePaaS's own apps left to phase 2.

2026-10-14:

5. No service runs between scans: a container per scan, started by the scan
   node's agent; the Docker socket only in it, only to read a local image.
6. Trivy's data on a volume the administrator must choose; its pin decides
   where scans run; Tmpfs refused; the volume can be changed freely.

## Later

- **Phase 2**: the CLI (`hivepaas app scan`, `hivepaas app vulns`) and MCP; a
  project's overview; HivePaaS's own apps; ignores in spec export and import; an
  SBOM to download (CycloneDX); secrets and misconfiguration scanners.
- **Phase 3**: a gate, per project: a deployment whose image has a fixable
  CRITICAL is refused, or held for approval; a scan of a built image before it
  rolls out.
- Several images in one container, if a daily run over many images spends
  too long starting containers.
