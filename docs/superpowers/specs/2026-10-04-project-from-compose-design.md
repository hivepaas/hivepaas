# Projects from a Docker Compose file

## Why

Most software people self-host ships a `docker-compose.yml`. To run one on
HivePaaS today, a user rebuilds it by hand - an app per service, its variables,
its volumes, its domains - and gets the names, the shared volumes and the
secrets wrong on the way. The templates were made the same way: a dozen of them
name the compose file they follow, and supabase, appwrite and invoice-ninja
rename its services into app keys by hand.

Spec import (`2026-09-23-config-spec-import-design.md`) already creates a
project, its envs and its apps with the full container, network and storage
model, from a plan the user reviews and accepts, and deploys them. What is
missing is a reader for the format everyone already has.

## What it does

Paste a compose file - with its `.env` and the files it reads - review what
each service becomes, and create the project: one env, an app per service,
deployed.

It is a converter in front of spec import: compose, read by compose-go, becomes
`specmodel` documents in memory, which the import's planner plans and its writer
writes. Provisioning, the checks and the permissions they ask, the issues, the
plan hash and the first deployments are the import's. The converter adds issues
of its own: what the compose file says that HivePaaS cannot do, said before
anything is created.

**In scope:** a new project with one env; services with an image; variables and
the `.env`; `environment` and `env_file`; `configs` and `secrets`; named volumes
and relative binds; published ports, as domains or as ports on the nodes;
command and entrypoint; healthcheck, restart, resources, replicas and mode,
placement, labels, logging, the container's options; profiles; `extends` and
anchors; `depends_on`, as an order.

**Not in scope, and why:**

- `build:` - HivePaaS builds from a Git repository, not a local directory. A
  service with an `image` and a `build` takes the image; one with only a
  `build` is skipped unless the review gives it an image. Phase 4 builds from
  the repository holding the compose file.
- An existing project or env - phase 4: there the import would update an app of
  the same key, which a compose import must not do.
- Updating a project from a changed compose file - phase 4.
- `privileged`, `devices`, `network_mode`, `pid`, `ipc` and the like: Swarm
  services have none of them (the mapping lists each).
- Several `-f` files merged; a directory archive instead of single files -
  phase 4.

## Reading the file

- **compose-go v2** (`github.com/compose-spec/compose-go/v2`, the Compose
  Specification's reference loader, Apache-2.0; v2.16.1 of 2026-09-29, Go
  1.24+): it parses, merges `extends` and YAML anchors, applies profiles,
  interpolates, and normalizes the short syntaxes - ports, volumes, durations,
  sizes, healthcheck tests. Not structs of our own: the short forms alone are a
  parser, and `docker compose config` is what users will compare with.
- **Nothing on the server is read:**
  - the loader runs in an empty scratch directory holding only the request's
    files, removed after;
  - a path that is absolute or leaves that directory (`..`) is refused before
    loading - in `env_file`, `extends.file`, `include`, `configs`, `secrets`;
    a bind's source is read as text, never opened;
  - `env_file` is read from the request by the converter, not by the loader
    (`SkipResolveEnvironment`);
  - no remote resource loader is registered: an `include` from Git or OCI
    fails;
  - the interpolation environment is what the request gives, never the
    process's.
- **Limits:** the compose file 1 MB; its files 5 MB in all, 500 KB each
  (Docker's limit for a config); 50 services.
- **The project's name** is the request's, else the file's `name:`; it is
  checked as `POST /projects` checks one - its form, `hivepaas` refused, not
  taken. The env is the request's, `production` by default, checked the same
  way.

## Variables

- The interpolation environment is the values typed in the review, over the
  `.env` text. Validate lists every variable the file uses
  (`template.ExtractVariables`): its default, whether it is required (`:?`),
  whether a value was given, whether it is secret.
- Required with no value: blocked until given. Used with no value and no
  default: empty, as compose does, with a warning.
- **A secret variable** - marked in the review; by default one whose name holds
  the word PASSWORD, PASSWD, PASS, SECRET, TOKEN, KEY, PRIVATE, CREDENTIALS or
  SALT - becomes an env secret, inheritable, and in an `environment` value it is
  written `${secrets.NAME}`: encrypted at rest and masked in the dashboard,
  while the container gets the same value at deploy. Elsewhere - a command, a
  healthcheck, a label - HivePaaS has no references, so the value is written
  there, and a note says where. (Interpolation answers a secret variable with a
  marker; the converter then writes the reference in environment values and the
  value everywhere else.)
- The review offers to generate an empty secret variable: 32 random hex
  characters, made in the browser and shown.
- **`$`:** an environment value that still holds `${` after interpolation -
  written `$${` in compose - is marked `literal`, so HivePaaS does not read it as
  a reference. HivePaaS never expands a command - a deployment splits it by
  shell rules, nothing more - so its `$` reaches the container as compose passes
  it.

## Files the compose file reads

`env_file`, the `file` of `configs` and `secrets`, single-file binds, the files
of `extends` and `include`: the request carries them by their path in the
compose file (`files: {"./nginx.conf": <base64>}`). Validate lists those used
and missing (`needs`, with what reads each), and the review asks for them,
pasted or picked. One still missing is created empty, its setting pending, with
a fixable issue: the app will not start right until it is filled.

## Services to apps

### Names and the network

- An app per service, named after it; its key is `CalcAppKey` of the name
  (lowercase, hyphens, 63 at most). Every app joins its env's network with its
  key as alias, so `db:5432` resolves as it does in compose.
- A name its key changes (`my_db` → `my-db`, `DB` → `db`) is added as a further
  alias on the env network, so `my_db:5432` still resolves; so are
  `container_name`, `links` and each network's `aliases`.
- Two services with one key: blocked - rename one.
- Every app is on the env's one network. Compose's own networks (`frontend`,
  `backend`) are not kept: every app reaches every other, as within one compose
  network. An external network is not joined (warning).

### The mapping

| Compose | Becomes | Notes |
| --- | --- | --- |
| `image` | `deployment.source{activeMethod: image, imageSource.image}` | No tag: it deploys `latest` (note). A private registry needs its credentials in Registries (note) |
| `build` | - | With an `image`: the image (warning). Without: skipped, unless the review gives an image |
| `command`, `entrypoint` | `source.command`, `source.entrypoint` (phase 0) | Each argv joined with `shellquote.Join`, the inverse of the `CmdSplit` a deployment applies - not `executil.ArgQuote`, which leaves words holding `${` or quotes alone. An empty `entrypoint`: warning |
| `working_dir` | `source.workingDir` | |
| `user`, `group_add`, `hostname`, `stop_signal`, `stop_grace_period`, `tty`, `stdin_open`, `read_only`, `init` | `container.*` | The hostname is the app key when not given: the import clears a field it is not given |
| `environment`, `env_file` | `settings.envVars` | `env_file` first and `environment` over it, as compose. Secret variables as references (above) |
| `secrets` | an env secret per top-level secret, mounted by `settings.settingMounts` at its target (`/run/secrets/<name>` by default), with uid, gid, mode | From `file` (the request's) or `environment` (a variable) |
| `configs` | an env config file per top-level config, mounted the same way (`/<name>` by default) | From `file`, `content` or `environment` |
| named volumes | a managed mount on the volume the review picks - the project's `default` unless another - at `subpath: <volume>` | Below |
| binds, `tmpfs` | below | |
| `ports` | the review's choice for each: a domain, a port on the nodes, or none | Below |
| `expose` | - | The env network reaches every port |
| `healthcheck` | `container.healthcheck` | `["CMD", ...]`: mode CMD, the argv joined; `CMD-SHELL` or a string: CMD-SHELL; `NONE` or `disable`: NONE. Interval, timeout, start period, start interval, retries |
| `restart`, `deploy.restart_policy` | `container.restartPolicy` | `always`, `unless-stopped`: any; `on-failure[:n]`: on-failure; `no`: none - or a job (below) |
| `deploy.replicas`, `scale`, `deploy.mode` | `service.modeSpec` | replicated, global, replicated-job, global-job |
| `deploy.resources`, `cpus`, `mem_limit`, `mem_reservation`, `pids_limit`, `shm_size`, `memswap_limit`, `mem_swappiness`, `oom_score_adj` | `resources` | GPUs (`gpus`, or a device reservation for `gpu`): `enableGPU` |
| `cap_add`, `cap_drop`, `ulimits`, `sysctls` | `resources.capabilities` | Write on the Cluster module, as the import asks; without it, dropped and the app kept (fixable) |
| `security_opt` | `container.privileges` | seccomp, apparmor, label, no-new-privileges; any other: warning |
| `deploy.placement` | `service.placement` | Warning: they name nodes and labels of the cluster it was written for |
| `labels`, `deploy.labels` | `container.containerLabels`, `container.serviceLabels` | `traefik.*`, `hivepaas.*`, `com.docker.stack.*` are dropped, as HivePaaS drops them (note); phase 4 reads Traefik's `Host()` rules as domains |
| `logging` | `container.logDriver` | |
| `extra_hosts`, `dns`, `dns_search`, `dns_opt` | `networks.hostsFileEntries`, `networks.dnsConfig` | |
| `networks` | the env network, with the aliases above | `ipv4_address`, `priority`, external networks: warning |
| `depends_on` | the order of creation and of first deployments | Below |
| `profiles` | the services of the profiles the request names, and those with none | As compose |
| `privileged`, `devices`, `network_mode`, `pid`, `ipc`, `uts`, `userns_mode`, `cgroup`, `cgroup_parent`, `runtime`, `isolation`, `platform`, `cpu_shares`, `cpuset`, `cpu_*`, `blkio_config`, `oom_kill_disable`, `storage_opt`, `volumes_from`, `external_links`, `mac_address`, `domainname`, `post_start`, `pre_stop`, `develop`, `pull_policy` | - | Warning, each named: Swarm services have none of them, or HivePaaS does not set it. For `privileged` and `devices` it says the app may not work |
| `x-*` | - | Ignored; phase 4 reads `x-hivepaas` |

### Volumes

- **A named volume** is a managed mount on the volume the review picks for the
  project: the project's own `default` - a directory on the node HivePaaS runs
  on, made with the project - unless the review picks a global cluster volume.
  It sits at `subpath: <volume name>`: the data lands in
  `<env>/<app>/<volume>`, and two volumes of one service stay apart.
- **Shared by several services**, it is owned by the first of them, in
  `depends_on` order, that writes to it; every other mounts the owner's
  directory with `sourceApp{app, write: not read_only}` and is created after it
  (the import already creates first an app another's mount reaches). Without
  this each app would get an empty directory of its own - as twenty's server
  and worker do today, a template to check.
- `external: true`: the review picks an existing cluster volume; until it does,
  a new directory, with a warning. A `driver` or `driver_opts` (NFS, ...): a
  warning - the project's volume is used; a cluster volume with that driver is
  the way (phase 4 makes one).
- **A relative bind**:
  - of a file the request carries, or whose source has an extension
    (`./nginx.conf:/etc/nginx/nginx.conf`): an env config file holding it,
    mounted at the target; still missing, it is empty, pending, fixable;
  - else, of a directory (`./data:/var/lib/mysql`): a managed mount like a named
    volume called after the path (`data`), which starts empty - a warning says
    its files are not copied.
- **An absolute bind**: `/var/run/docker.sock` is not mounted (warning): Docker
  API access is HivePaaS's way, and phase 4 sets it. Any other host path is a raw
  bind (`storage.dockerMounts`), which the import allows only with privileged
  apps on and an administrator; otherwise it is dropped and the app kept
  (fixable).
- `tmpfs`, and a volume of type tmpfs: `storage.dockerMounts`, tmpfs.

### Ports

Each published port is one of these, chosen in the review:

- **A domain**: HTTP through Traefik, with TLS -
  `routing{exposePublicly, domains: [{domain, containerPort: <target>}]}`.
  Suggested `<app key>-<project>.<root domain>` when HivePaaS has a root domain,
  the routing screen's suggestion; the project part with hyphens, as a DNS name
  needs. Taken: the import's `DOMAIN_IN_USE`.
- **A port on the nodes**: `networks.endpointSpec.ports[{target, published,
  protocol, publishMode}]`, compose's `mode` as the publish mode. Taken: the
  import's `PORT_IN_USE`.
- **None**: the env network reaches it anyway.

By default: a domain for a TCP port whose target is a usual HTTP one (80, 3000,
5000, 8000, 8080, 8888), when there is a root domain; none for a database's or a
broker's (5432, 3306, 1433, 27017, 6379, 11211, 9200, 5672, 9092) - a compose
file publishes them for a laptop, and a server should not; a port on the nodes
for every other, and for every UDP one (Traefik routes no UDP here). A port with
no published side (`"80"`): none.

### Order, start, jobs

- Apps are created, and their first deployments queued, in `depends_on` order; a
  cycle is blocked.
- HivePaaS does not wait for one to be healthy before it starts the next
  (`condition: service_healthy`). As with `docker stack deploy`, an app that
  starts before what it needs fails and is restarted until it answers; a note
  says so on each service with a condition.
- A service with `restart: "no"` that another waits on to complete
  (`condition: service_completed_successfully`) - a migration, a setup step -
  is a `replicated-job`: it runs to completion, and again on each deployment.

## API

- `POST /projects/from-compose/validate` and `POST /projects/from-compose/apply`.
  Not under the spec routes: they ask the permission `POST /projects` asks, not
  Write on the System module, which the global import asks because it can
  overwrite any project. The gates inside are the import's: Write on the Cluster
  module for capabilities, privileged apps for host binds, and the rest.
- The request, as YAML for brevity:

```yaml
compose: <text>
dotenv: <text>
files: {"./nginx.conf": <base64>}
variables: {DB_PASSWORD: {value: "...", secret: true}}
project: {name: blog, env: production}
profiles: [worker]
volume: "" # a global cluster volume's id; empty: the project's default
services:
  web:
    image: "" # for a service with only a build
    ports:
      - {published: 8080, target: 80, protocol: tcp, as: domain, domain: web-blog.example.com}
selection: {exclude: []} # the import's: to leave services out
deploy: true # the import's deployCreated
planHash: "..." # apply only
acceptIssues: true # apply only
```

- Validate answers what the review shows: the project's name and key; each
  service as an app - its key, image, ports and their choices, volumes, what
  was dropped; the variables; the files needed; and the import's plan - its
  tree, issues and summary - with its hash.
- Apply answers the plan's outcomes, the project's id and the deployments
  queued. Its audit entry: `project-from-compose`.

## Backend

- `service/composeservice`: `Read` - the loader, sandboxed, and the variables
  and files the file needs - and `Convert` - the documents, the issues, and the
  services as the review shows them.
- `usecase/composeuc`: validate and apply, the name checks of
  `projectuc.CreateProject`, then the import below.
- **Spec import, made able to take it:**
  1. Plan and apply a bundle in memory, with issues from the caller - counted in
     the plan hash - and the caller's permission checks: `planBundle` and
     `applyBundle`, exported as such. No reveal-secrets permission: the values
     come from the request, and nothing stored is shown.
  2. A new project's `default` volume as a mount's source: planned as there, as
     `PrepareNewProject` makes it before the apps. Today a bundle must carry it,
     and its host path trips the Cluster gate.
  3. The builder (`CheckImportable`, `BuildApp`) run at validate too: a block
     that would fail is an issue there, not a transaction rolled back at apply.
  4. A new project's name and envs checked as `POST /projects` checks them.
  5. The `COMPOSE_*` issue codes, with the dashboard's texts.
- **The converter's issues**, on the app's node, and the variables' on the
  env's:
  - blocked: `COMPOSE_VARIABLE_REQUIRED`, `COMPOSE_KEY_CONFLICT`,
    `COMPOSE_DEPENDS_CYCLE`;
  - skipped: `COMPOSE_NO_IMAGE`;
  - fixable: `COMPOSE_FILE_MISSING`, `COMPOSE_MOUNT_DROPPED`,
    `COMPOSE_CAPABILITY_DROPPED`;
  - warning: `COMPOSE_NOT_SUPPORTED` (the fields, named),
    `COMPOSE_BUILD_IGNORED`, `COMPOSE_DIRECTORY_EMPTY`,
    `COMPOSE_VOLUME_EXTERNAL`, `COMPOSE_NETWORK`, `COMPOSE_PLACEMENT`;
  - notes: `COMPOSE_SECRET_VARIABLE`, `COMPOSE_ALIAS_ADDED`,
    `COMPOSE_START_ORDER`, `COMPOSE_LABELS_DROPPED`, `COMPOSE_IMAGE_UNPINNED`.

### To check first

- A setting mount whose source is an env secret or config file - else a copy
  per app that mounts it.
- The env network of a new project as an attachment at plan time: it is made
  with the first app.
- compose-go's options that keep it off the disk: `SkipResolveEnvironment`, and
  the local loader of `extends` and `include` held to the scratch directory.
- `/projects/from-compose` before `/projects/:projectID` in the router; and
  `/projects/new/compose` before `/projects/:id` in the dashboard's.

## Phase 0: entrypoint

Compose files set `entrypoint` often, and HivePaaS cannot: a deployment sets the
container's arguments and always clears its command
(`dockerhelper.ContainerCommandApply`), so the image's entrypoint always runs.
Before the converter: `AppDeploymentSettings.Entrypoint`, applied at each
deployment as `ContainerSpec.Command` when set; its field on the deployment
settings screen; `source.entrypoint` in the spec and among the templates'
buildable fields; MCP's app settings; the docs.

Nothing changes for an app without one: empty is the image's entrypoint, as
now, so no data is migrated, and a deployment writes the service spec it wrote
before - the upgrade restarts nothing. What must change with it, or an app that
has one breaks:

- **The container settings screen** shows `Command` and `Args` joined as one
  command, and a save splits it all into the arguments and clears the command:
  one save would turn the entrypoint into arguments of the image's own. It
  shows the entrypoint apart, and keeps it. Its join quotes nothing either -
  today `sh -c "a && b"` comes back split into words on a save - so both are
  joined with `shellquote.Join`.
- **The deployment settings request** replaces the whole setting: a client that
  does not know the field would clear it. No compatibility is kept with such a
  client - nobody runs HivePaaS yet (2026-10-04) - so it is a plain string, as
  `command` is. MCP merges a patch into what it read, so it keeps it.
- **Docker's init** is decided from the image's entrypoint
  (`applyContainerInit`); with an entrypoint of the app's own, from that one -
  else `tini --` as the entrypoint would run under docker's init, two inits,
  which s6 refuses.
- **No new setting version**: a version here rewrites old data (2 removed
  fields, 3 turned `autoDeploy` on for old rows), and there is none to rewrite.
  A release before this one, importing a bundle that sets an entrypoint,
  ignores it as it ignores any key it does not know; a template that sets one
  says so in `requires.versionCode`.
- Already right: a clone without the deployment settings clears the command,
  and one with them keeps both; a function passes none; a build deploys through
  the same apply.

**As built** (2026-10-04): `dockerhelper.ContainerCommandApply` takes both
lines and returns an error where it panicked on a quote left open; both
settings requests refuse such a line; the screens show argv quoted
(`CommandLine`). The container screen's healthcheck is written as the spec's
(`HealthcheckTest`, shared): CMD-SHELL whole, not split; NONE kept; and an
empty mode with no command inherits the image's test - before, it wrote `[""]`,
which docker does not run at all. The export quotes CMD's argv, so an import
splits it back.

## Dashboard

- The projects page: **New Project** gets a menu - an empty project, or **From
  Docker Compose**: a page, `/projects/new/compose`, not a dialog, as the
  review is large.
- **The file**: the compose text, pasted or picked; the `.env`; the project's
  name and env; the profiles the file has.
- **The review**, validated again whenever something changes (400 ms, as spec
  import does):
  - the services as apps: image, replicas, each port's choice and domain, the
    volumes and where they land, what was dropped;
  - the variables: value, secret, generate;
  - the files needed: paste or pick each;
  - the volume for the project;
  - the import's plan tree and its issues - its component, reused - and deploy
    after, on by default.
- **Create** ("Create and accept N issues"): then the project, the deployments
  queued, the warnings.

## MCP, docs

- MCP (phase 3): `plan_create_project_from_compose`, then the apply with its
  plan hash once the user agrees, as `plan_install_app` does.
- Docs (phase 3): Deploying apps, "From a Docker Compose file": what each field
  becomes, what is left out, variables and secrets, volumes, ports.

## Tests

- **The converter**: a table per field; and a corpus of real compose files -
  those the templates were made from (twenty, dify, supabase, appwrite,
  invoice-ninja) and common ones (WordPress and MySQL, Gitea and Postgres, n8n,
  Outline, Plausible, Umami, Immich) - each with its expected documents and
  issues, checked against `docker compose config` for interpolation and the
  short forms.
- **The loader's safety**: absolute and `..` paths, remote includes, the
  process environment never read.
- **The usecase**: validate and apply on the test database with a fake swarm,
  as spec import's tests do.
- **Live**, on the local stack: a compose of an app and its database - created,
  deployed, reached by its domain, the app reaching the database by its compose
  name - then removed.

## Into an existing project

Agreed with the user on 2026-10-04: a compose file goes into a project that
exists, into one of its envs or a new one.

- **Entry.** The project's apps list: "New From ▾" holds Template and Docker
  Compose. The page is the same as a new project's, at
  `projects/:id/apps/from-compose`; the target replaces Project Name and
  Environment: the project shown, and either an existing env - the header's,
  when one is chosen - or a new one, named and coloured. A project with ten
  envs takes no new one.
- **API.** `POST /projects/{projectID}/from-compose/{validate,apply}`, with
  `POST /projects/{projectID}/spec/import`'s gate: Write on the project. The
  body is the new project's, with `project.env` the env's name, and
  `project.newEnv` saying it is to be created - refused when the name or its key
  is taken, and an existing env not found is refused too. `project.name` is not
  read.
- **Only created.** The import runs with `existing: keep`: nothing the project
  has is changed. The env's settings it lacks are created; those it has are
  kept.
- **What the env has.** The converter is given the env as export sees it,
  secrets omitted:
  - a service whose key, or name, is an app's key or alias here is blocked,
    `COMPOSE_APP_EXISTS`, until the review chooses: use the app there - the
    service is not created, and the others reach that app by the name - or
    another key. A key chosen is checked as a service's own;
  - an alias the env already answers to is not added, a warning,
    `COMPOSE_ALIAS_TAKEN`: the name reaches the app there;
  - a secret variable's env secret that exists is used as it is, a warning,
    `COMPOSE_SECRET_EXISTS`: the value given is not written. Overwriting it
    is the env's Secrets screen's;
  - a file's secret or config file whose name is taken is named
    `<name>-2`: a mount never reads a setting the request did not carry, so
    no reveal gate is needed.
- **Volumes** need nothing: an app's directory on the project's volume is
  `<env key>/<app key>/<subpath>`, so the same file in two envs, or beside
  other apps, shares no data.
- **Audit**: `compose-import`, at the project's scope, with the env and whether
  it was created.
- **As built** (2026-10-04): `specservice.CurrentEnv` is the env export sees,
  secrets omitted; the converter takes it as `ConvertReq.Existing`, and a
  service's `app` and `useExisting` as the review's choices. A service used as
  the app there is that app's own document, which `keep` leaves alone. MCP's
  tool takes `projectId`, `newEnv`, `apps` and `useExisting`.

## Opening a folder, includes, directories

Built on 2026-10-04, while the user was away, from what was agreed: "do
everything for compose, review later".

- **A folder, read in the browser.** Rather than an archive uploaded and kept
  on the server under an id - storage, expiry, unpacking with its zip-slip and
  symlink risks - the dashboard reads the folder itself (`webkitdirectory`):
  the shallowest compose file by compose's order of names, its `.env` (or it
  offers a `.env.example`, whose passwords are known to anybody), and from then
  on gives, from the folder, each file the review lists as needed. Nothing
  else leaves the browser, the request stays stateless, and its limits - 100
  files, 5 MB, 500 KB each - hold.
- **Includes and extends.** The loader resolves a relative path from the file
  naming it, which compose-go puts in the context (`consts.ComposeFileKey`), and
  answers `Dir` with the loaded file's directory, absolute and inside the
  scratch directory. compose-go makes an included file's paths absolute there;
  the reader makes them relative to the compose file again, and one climbing
  out into HivePaaS's data directory reads as leaving the directory. An
  include's `project_directory` is refused - an included file's directory is
  its project's - and a URL is not fetched. A compose file an include or
  extends reads that the request lacks, and an include's env file, are needs
  (`compose`, `env_file`) that stop the read, as a required variable does.
  The variables of included files are found once compose-go has read them, and
  the file read again with them; the `.env` beside an included file and an
  include's env files give values after the request's.
- **Only what compose-go reads is written.** The scratch directory gets a file
  of the request's only when compose-go asks for it: an included or extended
  compose file, the `.env` beside one, an include's env file. The converter
  reads everything else from the request itself. It is in the data directory's
  `tmp/<day>` (`fileutil.CreateTempDirInAppPath`), which the system cleanup
  sweeps.
- **Directories.** A directory a service mounts is a need (`directory`), given
  once a file under it is. It stays the app's own directory on the project's
  volume - what the app writes there is kept - and each file given under it is
  mounted at its place in it, read only, from an env config file: Docker mounts
  the config over the volume, nested mounts going deepest last. The dashboard
  gives a directory's files all or none, checked unless unchecked.
  Checked on Docker 29.8 with plain containers (runc makes a Swarm config's
  mount as any other): a file mounted in a writable volume's subpath works -
  the volume's files stay, and the app writes beside it - while one in a
  read-only volume fails to start ("make mountpoint ...: read-only file
  system"), Docker having no place to make for it. So a directory mounted
  read only with files given is those files alone, with no volume; one read
  from another service that writes it stays that service's, without them.
- **Traefik labels.** The hosts of a service's routers' `Host` rules - its
  labels' and its deploy's - are a port of the review's, source `labels`, on
  the container port they reach (the router's service's load balancer port,
  the one service's, or the service's one port): the first host the domain
  unless the review says otherwise, the others domains too. A published port
  of the same container port is then none by default. Path matchers are not
  kept; only a host a domain can be is taken.
- **Read-only volumes holding a file.** Found by the same check, and older: a
  file bind inside a volume mounted `:ro` - `site:/html:ro` with
  `./index.html:/html/index.html` - would not start either. Such a volume is
  mounted writable, with a warning (`COMPOSE_MOUNT_WRITABLE`).
- **Nested directories stay apart.** A directory under another mounted one
  is its own on the volume, as before, now with a warning
  (`COMPOSE_DIRECTORY_APART`). Making it the other's subdirectory was built
  and dropped: `volumeservice.MakeDirWritableCmd` opens up only the leaf it
  makes, and only while empty, so preparing the inner one first would leave
  the outer one root's, unwritable for a service not running as root, with
  deployments in no set order.
- **Not done, on purpose.** `docker.sock` stays dropped with its note: host
  mode is root on the node, which HivePaaS never grants from a template
  (`HostModeFromTemplate`), and the proxy is not the socket a compose file
  expects. Database images as the database kind wait for a design: the kind
  publishes credentials other apps link to, beside its own self env vars.
- **Setting mount keys.** Found on the way: the converter keyed setting mount
  entries after the mount (`secret-api_key`, `file-etc-nginx-nginx.conf`), and
  an entry whose key is not an entry key is never mounted. Each mount is now an
  entry of its own, keyed by `EntryKeyFor`, numbered when two would share one.

## Phases

0. Entrypoint for apps.
1. The backend: reader, converter, endpoints, the import's changes; the tests
   and the corpus. **Done** (2026-10-04). As built:
   - `service/composeservice` converts. compose-go loads the file in a scratch
     directory through a resource loader of its own, the only one it has, that
     serves the request's files and nothing else; env files and label files
     are read by the converter; a variable with neither a value nor a default
     is given an empty one, so that compose-go logs nothing of it. What a
     service says that HivePaaS does not carry is found by compose-go's
     unsupported-attribute check, from a list of patterns.
   - `POST /projects/from-compose/{validate,apply}` are the spec handler's,
     with `POST /projects`' gate; the usecase is spec import's neighbour in
     `specuc`, with its gates but no reveal gate; its audit entry is
     `compose-import`. The name checks of creating a project moved to
     `projectservice.CheckNewProjectName`, which both use.
   - Spec import: `PlanBundle` and `ApplyBundle`; the reader's issues on their
     nodes, in the hash, a skipped one skipping its node; the defaults a new
     project is given (`projectservice.NewProjectDefaults`) resolve in the
     planner and the writer; and a bundle narrower than global whose project
     is not here is compared with nothing, rather than with an export of the
     whole installation.
   - Not as planned: validate does not run the builder - the converter writes
     only blocks it builds; compose-go refuses a dependency cycle itself, as an
     error; and `depends_on` does not order creation - the import creates a
     volume's owner first, and the rest by key.
   - Added after the first try: a variable written out in a service's
     environment whose name reads as a secret's, with a value, is kept as a
     secret of the app (`COMPOSE_SECRET_ENV`), its variable `${secrets.NAME}`.
     A value a `${VARIABLE}` fills follows the variable instead; one the review
     says is not secret is read through a marker of its own, so that it is
     written as plain text.
2. The dashboard's page. **Done** (2026-10-04). As built: New Project is a
   menu - an empty project, or From Docker Compose, the page
   `/projects/new/compose`. Its API sits beside spec import's, in the
   operations module, and shares the plan's schema and its tree. The file is
   read again 500 ms after a change; a service the file gains is checked in the
   plan, one unchecked stays so.
3. Docs and the MCP tool. **Done** (2026-10-04): Deploying apps, "From a Docker
   Compose file", the quick start pointing at it; MCP's
   `plan_create_project_from_compose`, applied by `apply_plan`.
4. Later:
   - a directory archive instead of single files: done as a folder read in
     the browser, see "Opening a folder, includes, directories";
   - `build:` from the Git repository holding the compose file - it needs a
     build context directory for apps, which functions already have;
   - into an existing project or env: see "Into an existing project";
   - updating a project from a changed compose file: the import's diff, never
     deleting, the file kept with the project to compare with;
   - Traefik labels read as domains; `x-hivepaas` (domains, volume, autoscale,
     kind); `docker.sock` as Docker API access; database images given the
     database kind, for backups; volume drivers as cluster volumes; several
     `-f` files; a CLI command.

## Risks

- **Files written for a laptop**: database ports published, host paths bound,
  `privileged`. The defaults leave those out, and the review shows each before
  anything is created.
- **One node**: every app with data on the project's `default` volume runs on
  the node HivePaaS runs on, as a template's does. On several nodes, pick a
  cluster volume.
- **Start order**: a database's health is not waited for, so the first minutes
  may be restarts, as the notes say.
- **compose-go** brings a tree of its own (YAML, a JSON schema): pinned, and it
  only reads.
- **Spec import's own gaps**, found on the way and fixed with it: validate
  without the builder, and the default volume.

## Decisions, to confirm

1. A new project first; an existing env later.
2. Ports: a domain for HTTP, none for databases, a port on the nodes otherwise.
3. Secret-looking variables become env secrets by default.
4. `build:` needs an image in phase 1.
5. Entrypoint as phase 0.
6. The env is `production` by default.
