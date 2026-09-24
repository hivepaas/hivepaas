# Docker API Access for Apps

Some apps do their work by starting containers: a CI runner starts one per job,
Autobase starts an Ansible container per cluster operation, Appwrite starts one
per function runtime. Upstream, every one of them is installed by mounting
`/var/run/docker.sock` into it. On a node that is root; on a manager it is the
whole cluster.

This document designs giving such an app the Docker API without the socket: a
proxy that HivePaaS runs, holding the socket itself, lets the app do what its
template declares and nothing else.

---

## What it is for

A survey of the popular self-hosted apps that ask for the socket (September 2026)
sorted them by what they do with it:

| group | apps | what they call |
|---|---|---|
| job runners | Autobase, Kestra, Appwrite executor, Gitea/Forgejo runner (act), Woodpecker agent, GitLab runner, Jenkins docker agents, Windmill docker jobs, JupyterHub DockerSpawner, Coder | image pull and inspect; container create, start, wait, attach, logs, remove; exec; copying files in and out; volumes and networks of their own |
| observers | Uptime Kuma, Homepage, Glance, Homarr, Beszel agent, Dozzle | list, inspect, logs, stats and events of containers they did not create |
| stack managers | Nextcloud AIO master container, mailcow, watchtower, docker-volume-backup | long-lived containers beside themselves, and restarting other apps' containers |
| Docker managers | Portainer, Dockge, Komodo, Coolify | everything |

Job runners are the group this design serves. The catalogue already has Gitea,
Forgejo, Jenkins, Windmill and Jupyter without the part of each that runs jobs,
and has refused Autobase for want of it. Observers come later (see *Not in this
design*). Stack managers fight HivePaaS for the containers it runs. Docker
managers cannot be given anything less than everything.

Read from source for this design: the Autobase console 2.11, utopia-php
orchestration (Appwrite 2.3's executor), nektos/act (the Gitea runner),
Woodpecker's docker backend, Dozzle, Kestra's and Nextcloud AIO's compose files.

## Decisions

1. **No app is given the Docker socket.** An app with Docker API access talks to
   a proxy on its node, and the proxy holds the socket.
2. **One engine, configured by data.** A template declares a policy (§1); nothing
   about a particular app is code. An app that needs something the policy cannot
   say waits until the policy can say it, for every app at once.
3. **Deny by default.** Every endpoint, every field of a container and every
   mount is refused unless a rule lets it through (§3, §4). Docker adds fields;
   a field the proxy does not know is refused, not passed.
4. **What an app starts stays with the app.** Its children run on its node, on
   its own network, with its own storage, labelled as its own. It sees and
   touches nothing else (§4, §5).
5. **The gate is Write on the Cluster module, not the privileged-apps switch.**
   Nothing the proxy allows reaches the host, so granting it is the same kind of
   decision as granting a capability (§9). The switch stays what it was: raw
   access to the host, which only import can ask for.
6. **Measured before designed.** A spike ran the Autobase console and the Gitea
   runner through a prototype of this engine, unchanged, on Docker 29.8. What it
   found is written into the rules below where it applies.

---

## 1. The policy

A new singleton setting type, `app-docker-api`, written in an app document as
`settings.dockerApi`:

```yaml
settings:
  dockerApi:
    images: ["autobase/automation:2.11.0"]
    sharedDirs: ["/var/lib/autobase/ansible"]
    networks: []
    allow: []
    limits: {containers: 3, memory: 2Gi, cpus: 2}
```

| field | meaning | default |
|---|---|---|
| `images` | The images children may run, as patterns over the reference the client names (`path.Match` on `repository:tag`, after Docker's own normalisation of `docker.io/` and `library/`). `"*"` is any image. | required, not empty |
| `sharedDirs` | Directories of the app a child may bind, at most 5. Each is an absolute path in the app's container that equals or lies below the target of one of the app's own storage mounts. | none |
| `networks` | Networks children may join besides their own. The only value in this phase is `env`: the app's project-env network, for a runner whose jobs clone from a forge in the same env. | none |
| `allow` | Groups of endpoints beyond the core (§3): `exec`, `files`, `volumes`, `networks`, `nestedSocket`. | none |
| `limits.containers` | Children that may exist at once, running or not. At most 50. | 5 |
| `limits.memory`, `limits.cpus` | The most one child may ask for, and what it gets when it asks for nothing. | 1Gi, 1 |

The buildable subset (`specmodel/buildable.go`) accepts the block with exactly
these fields and bounds. Build refuses a `sharedDirs` entry that no mount of the
app covers, and one covered by a cluster volume: a container outside Swarm
cannot mount a CSI volume.

Export writes the block like any other setting. Import reads it back (§9).

## 2. What the app sees

- A socket at `/var/run/hivepaas/docker.sock`, and `DOCKER_HOST` set to
  `unix:///var/run/hivepaas/docker.sock` unless the app sets `DOCKER_HOST`
  itself. A template for an app that reads another variable - Autobase reads
  `PG_CONSOLE_DOCKER_HOST` - sets that one to the same value.
- A network of its own, `hp-dapi-<app id>`. Its children join it by default.
- Refusals as Docker errors: status 403 and a message starting `hivepaas:` that
  names the rule, such as `hivepaas: HostConfig.Privileged is not allowed`. The
  app logs it as it logs any daemon error, which is where the person debugging it
  will look.

## 3. Endpoints

The API version prefix (`/v1.xx`) is stripped before matching, and kept on the
forwarded request except where §4 says otherwise. A path that is not in plain
form - an escaped character, `..`, a doubled slash - is refused, since the proxy
and the daemon could read it differently.

| endpoint | group | rule |
|---|---|---|
| `GET,HEAD /_ping`, `GET /version` | core | passed |
| `GET /info` | core | passed with `Swarm`, `Labels` and `RegistryConfig` removed |
| `GET /images/json`, `GET /images/{name}/json` | core | passed |
| `POST /images/create` | core | the image must match `images`; `fromSrc` (import) refused |
| `GET /containers/json` | core | filtered to the app's children |
| `POST /containers/create` | core | §4 |
| `GET /containers/{id}/json`, `logs`, `stats`, `top` | core | the app's children only |
| `POST /containers/{id}/start`, `stop`, `kill`, `wait`, `restart`, `resize`, `attach`; `DELETE /containers/{id}` | core | the app's children only |
| `GET,PUT,HEAD /containers/{id}/archive` | files | the app's children only |
| `POST /containers/{id}/exec` | exec | the app's children only; the body may not ask for `Privileged` |
| `POST /exec/{id}/start`, `resize`; `GET /exec/{id}/json` | exec | the exec's container must be one of the app's children |
| `GET /volumes`, `POST /volumes/create`, `GET,DELETE /volumes/{name}` | volumes | filtered, labelled, the app's only; driver `local` with no options |
| `GET /networks`, `POST /networks/create`, `GET,DELETE /networks/{id}`, `POST /networks/{id}/connect`, `disconnect` | networks | filtered, labelled, the app's only; driver `bridge`, scope `local`, no options, default IPAM. Connect takes a child, or the app's own task (Appwrite's executor connects itself to its runtimes' network) |
| anything else | - | refused: build, BuildKit sessions, swarm, plugins, secrets, configs, system, prune, commit, export, events |

Streams and hijacked connections (attach, exec start, pull progress, logs
follow) pass through `httputil.ReverseProxy` with flushing on. The spike ran
all of them through it without special handling.

## 4. Creating a container

In order, for `POST /containers/create`:

1. **Fields.** `Config`, `HostConfig` and each `EndpointsConfig` entry have an
   allowlist. Any other field must be absent or hold the value a client sends
   for "not set":
   - zero - `null`, `""`, `0`, `false`, `[]`, `{}`, or an object of zeros - since
     Go clients marshal whole structs;
   - a **neutral value** named per field, where zero means something else. The
     docker CLI sends `MemorySwappiness: -1`.
   - `MaskedPaths` and `ReadonlyPaths` only `null`: an empty list there unmasks
     `/proc`.

   | object | allowed fields |
   |---|---|
   | `Config` | Hostname, Domainname, User, AttachStdin/out/err, ExposedPorts, Tty, OpenStdin, StdinOnce, Env, Cmd, Healthcheck, ArgsEscaped, Image, Volumes, WorkingDir, Entrypoint, NetworkDisabled, Labels, StopSignal, StopTimeout, Shell, HostConfig, NetworkingConfig |
   | `HostConfig` | Binds, Mounts, NetworkMode, RestartPolicy, AutoRemove, Memory, MemorySwap, MemoryReservation, NanoCpus, CpuQuota, CpuPeriod, CpuShares, PidsLimit, ShmSize, Dns, DnsOptions, DnsSearch, ExtraHosts, LogConfig (driver `json-file` or `local`), Init, ReadonlyRootfs, Tmpfs, CapDrop, Ulimits, GroupAdd, ConsoleSize, Isolation |
   | endpoint | Aliases, DNSNames |

   So the following are refused without a rule of their own:
   - `Privileged`, `CapAdd`, `Devices`, `DeviceRequests`, `DeviceCgroupRules`;
   - `PidMode`, `IpcMode`, `UTSMode`, `UsernsMode`, `CgroupnsMode`, `Cgroup`;
   - `SecurityOpt`, `Runtime`, `Sysctls`, `CgroupParent`, `OomKillDisable`;
   - `VolumesFrom`, `Links`, `PortBindings`, `PublishAllPorts`, `VolumeDriver`.

   Each of those has a test.
2. **Image.** `Image` must match `images`.
3. **Storage.** Every bind and mount is one of these, and nothing else:
   - **Under a shared directory.** It becomes a volume mount of the directory the
     app itself mounts there. The proxy reads the mounts of the app's own task
     container on this node (a worker node cannot read services), takes the
     mount whose target covers the path, and joins its volume subpath with the
     rest of the path. A container carrying the owner label is never taken for
     the app's task, whatever its other labels say. Autobase binds `/var/lib/autobase/ansible` into its
     automation container, and the container gets the same directory of the
     app's volume, which is how the console reads the log the playbook writes.
   - **The app's socket**, `/var/run/hivepaas/docker.sock`, with `nestedSocket`.
     It becomes a mount of the socket volume with subpath `docker.sock`, so a
     job's own `docker run` goes through this proxy under this policy. The Gitea
     runner binds the socket named by its `DOCKER_HOST` into every job by default.
   - **A named volume**, with `volumes`. The volume must be the app's. One that
     does not exist yet is created first, labelled: act names volumes in `Binds`
     and never calls create.
   - **A tmpfs.**
4. **Network.**
   - `""`, `default`, `bridge` and `host` become the app's network. The
     automation container asks for `host` only to reach the internet, which the
     app's network also reaches.
   - `none` stays.
   - `container:<id>` is refused.
   - Any other name must be the app's network, a network the app created, or an
     entry of `networks`.
   - `EndpointsConfig` keys follow the same rules.
5. **Resources.** `Memory` and `NanoCpus` (or `CpuQuota`/`CpuPeriod`) default to
   the limits and may not exceed them. `CpuPeriod` must be one docker accepts
   (1000-1000000). `PidsLimit` defaults to 1024 and may not exceed 4096.
   Unlimited swap (`MemorySwap: -1`) is refused.
6. **Ownership.** The label `hivepaas.docker-api.app=<app id>` is set, replacing
   whatever the client put there, and every label starting `com.docker.` or
   `hivepaas.` is dropped. Those mark swarm tasks and what HivePaaS owns. The same
   holds for the labels of a volume or network the app creates.
7. **Count.** The app's children, counted by that label, must be fewer than
   `limits.containers`.
8. **Version.** The rewritten request goes out at the proxy's own API version.
   `VolumeOptions.Subpath` needs 1.45, the Gitea runner speaks 1.44, and the
   create response has the same shape in every version.

## 5. Ownership and visibility

Everything the proxy creates for an app carries the label, and everything the app
names by id is checked against it:
- a container, by inspecting it;
- an exec, by looking up its container;
- a volume or network, by inspecting it.

Lists are fetched and filtered. The app's own tasks count as its own only for
network connect and disconnect.

What stays visible is the node's image list and a trimmed `/info`. act needs
both, and neither holds a secret.

## 6. Where the proxy runs

In the HivePaaS agent, the global service that already holds each node's socket
and the host's filesystem.

- **The socket lives in a volume.** For every app with access, each node's agent
  keeps a local volume `hp-dapi-sock-<app id>`, labelled with the app, and listens
  on `docker.sock` inside its mountpoint.
- **A volume, not a host directory,** because Swarm creates a missing local
  volume when a task starts, but refuses a task whose bind source does not exist.
  That was measured: `bindOptions.createMountpoint` does not survive into a
  service spec. So neither side waits for the other: whichever of the task and
  the agent comes first creates the volume.
- **The socket's path is the app's identity.** No token is needed: only the app's
  tasks mount that volume, and §9 keeps every other mount of it out.
- **Every agent serves every app with access.** It is one listener per app per
  node. It means the socket is there on whichever node Swarm puts the task, and
  the children start on that same node, beside the shared directory.
- **The agent reconciles** the listeners with the database:
  - on start;
  - every 30 seconds;
  - when the backend calls it (a new `DockerAPIService.Sync` over the existing
    agent gRPC) after an app gains, changes or loses access.
- **A change of policy applies to the next request.** Nothing is restarted.

## 7. Deployment

When an app has access, building its service adds three things, and removing
access takes them away on the next deployment:
- the socket volume mounted at `/var/run/hivepaas`;
- `DOCKER_HOST`, unless the app sets it;
- the app's network `hp-dapi-<app id>`.

That network is an attachable overlay, created on first need and labelled
`hivepaas.docker-api.network=<app id>`. It does not carry the owner label, so
the app can use it but not remove it. It is an overlay rather than a bridge so that the app's task, a swarm
service, can join it.

## 8. Lifecycle

- **Deleting the app, or removing its access.**
  1. The backend asks every node's agent to remove the app's children, networks
     and volumes, by label.
  2. It then removes the app's network and socket volumes.
- **Every agent collects on a timer.** It removes:
  - anything labelled for an app that no longer has access;
  - an exited child older than 24 hours;
  - a network of the app's with no endpoints, older than an hour. The spike's
    first failed jobs left their networks behind.

  Volumes stay while the app has access, since they are caches (`act-toolcache`).
- **Stopping the app** leaves its children alone. A job in flight finishes.

## 9. Permissions and surfaces

**Granting.** Access is granted whenever:
- an app is created from a template with the block;
- an import creates the block, or changes it;
- the settings screen adds it or widens it.

Each of these needs Write on the Cluster module, like capabilities do. Narrowing
access or removing it needs only the app's own Write.

**Where the gate appears:**

| surface | behaviour |
|---|---|
| Template create and preflight | refused with `DOCKER_API_NOT_PERMITTED`, reported by preflight like the capability check |
| Import | the app is skipped with `DOCKER_API_NOT_PERMITTED` (severity skipped) |
| App settings | a Docker API screen shows the policy and edits it under the same gate |
| Export | writes the block |
| Audit | the setting change is audited like any other |

**Reserved names.** `hp-dapi-sock-*` volumes cannot be mounted by any app. The
storage screen never offers them, and import refuses a docker mount that names
one, as a host mount.

**The privileged-apps switch.** This work closes the TODO on the switch in
`config/security.go`: import's check of raw host mounts (`HOST_MOUNT_NOT_PERMITTED`)
also requires the switch to be on. Templates never ask for raw host mounts, so
nothing here depends on the switch.

## 10. Dashboard

- **Template detail** shows a "Docker API access" block, the way it shows
  capabilities. It lists:
  - which images children may run;
  - which directories they share;
  - which groups are allowed;
  - the limits;
  - one sentence on what that means.
- **The create dialog** shows the same block, and the preflight issue when the
  person may not grant it.
- **App settings** gains the Docker API screen of §9.

## 11. Templates

Shipped with this work, in `app-templates`:
- **`autobase`**, the console, which the catalogue refused before. It has:
  - `images: [autobase/automation:<version>]`;
  - a shared `/var/lib/autobase/ansible` on the app's volume, with
    `PG_CONSOLE_DOCKER_LOGDIR` pointing at it;
  - the console token and encryption key as generated secrets;
  - dbdesk turned off.
- **`gitea-runner`**, act_runner, registered by a token parameter. It has:
  - `images: ["*"]`;
  - `allow: [exec, files, volumes, networks, nestedSocket]`;
  - `networks: [env]`, so that jobs can clone from a Gitea in the same env.

The spike ran both of them unchanged against this design's rules. Autobase:
- deployed through the proxy;
- the console read the playbook's log from the shared directory;
- the operation ended `success`.

Gitea runner: a plain job, a job with a `redis` service, and a job running
`docker run` inside itself all passed. In that job, `--privileged` and
`-v /:/host` were refused.

## 12. What remains possible, and why it is accepted

- **Children run images.** With `"*"`, a child runs any image, with egress. That
  is what a person with Cluster Write could create as an app anyway, and the
  limits bound it.
- **Registry credentials pass through.** A child pulls with the credentials the
  app sends in `X-Registry-Auth`. They are the app's own.
- **A child shares the kernel.** A kernel exploit from a child is the same risk
  as from any container. Sandbox runtimes on dedicated nodes (gVisor, Kata,
  sysbox) are the answer to that, and a separate one.
- **The agent parses what apps send.** The create checker is fuzzed, and every
  endpoint it does not know is refused before any body is read.
- **Docker keeps adding fields.** A new field is refused until the allowlist
  names it. The tests pin the request bodies of the clients that matter.

## 13. Testing

- **Engine, table driven.** The fixtures are request bodies the spike recorded
  from the docker CLI 29, the Go SDK (act at API 1.44) and Autobase. Every
  refused field in §4 has a test, and so does every rewrite. The create checker
  has a fuzz test.
- **Agent against a real daemon**, skipped without Docker. It covers:
  - reconcile creating and removing listeners;
  - a child created and cleaned up;
  - collection on the timer.
- **Manual, on the Linux test server.** Autobase creating a cluster on a real VM
  over SSH, and a Gitea workflow that clones the repository.

## 14. Plans

1. **Engine.** A package with no HivePaaS dependencies: endpoints, the create
   checker, rewrites, ownership, and tests. The rest of this design builds on it.
2. **Agent.**
   - hosting the engine;
   - socket volumes and reconcile;
   - `DockerAPIService.Sync`, removal by label, and collection on a timer.
3. **Backend.**
   - the setting type;
   - the buildable subset and builder;
   - export and import;
   - template checks and preflight;
   - deployment (mount, variable, network);
   - lifecycle on delete;
   - reserved volume names;
   - the switch in import.
4. **Dashboard.** Template detail, create dialog, app settings screen.
5. **Templates.** `autobase` and `gitea-runner`.

## Not in this design

- **Observers**: reading containers an app did not create. They need visibility
  scoped to a project or env and `Env` removed from inspect.
- **Build and BuildKit sessions** for CI jobs that build images.
- **Children that publish ports**, or that run on other nodes.
- **Devices and GPUs** for children.
- **Sandbox runtimes** for nodes.
- **A list of an app's children** in the dashboard.
