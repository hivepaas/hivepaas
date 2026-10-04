# Calls and routes from OBI

Phase 3 of `2026-10-02-app-performance-observability-design.md`. That spec's
design section was written before phases 1 and 2 were built; this one takes
what they settled - the agent writes rows to its own stdout, the log collector
takes them - and what OBI has become since the spike.

## Why

Traefik's access log counts the requests that reach an app by its domains.
It does not see:

- **requests between apps inside a project** - a frontend calling its API on
  the project network, a worker calling a service;
- **the calls an app makes** - to its database, its cache, other apps, outside
  hosts - their rate, errors and time;
- **routes**: `/users/1` and `/users/2` are two paths to Traefik; the query
  folds numbers and ids, but not `/users/alice`.

OBI (OpenTelemetry eBPF Instrumentation) sees them from the kernel, without
touching the app's code. It costs memory on every node it runs on, so it is
opt-in, per node and per app.

**In scope:** server requests by route (all of them, not only Traefik's), the
calls an app makes, by peer; on the Metrics tab. **Not in scope:** traces
(phase 4, unchanged), SQL text, headers and bodies, autoscaling on these
numbers (a follow-up: the Requests signal would then see internal traffic).

## What has changed since the spike (2026-10-02, OBI v0.8.0)

- **OBI v0.14.0** (2026-10-02). It attaches uprobes and kprobes without
  `CAP_SYS_ADMIN` on most kernels, falling back to tracefs. Go context
  propagation still needs it, and HivePaaS does not use that: it reads OBI's
  data, it does not propagate traces into apps.
- **Selection.** `target_pids` (`OTEL_EBPF_TARGET_PID`), `container_name`,
  `containers_only` and the others. A dynamic PID selector exists in OBI's Go
  packages (`MutablePIDSelector`: add and remove PIDs without a restart). The
  standalone binary is not documented to expose it.
- **Metric features**: `application` (`http.server.request.duration`,
  `http.client.request.duration`, `db.client.operation.duration`, RPC,
  messaging) and `application_service_graph` (calls between services).
- **Map sizing.** `maps_config.global_scale_factor` (-3 to 3) is documented
  still. In v0.8.0, -2 made thousands of "sizes of Go and BPF maps don't match"
  errors; this is to measure again.
- **As built in phases 1 and 2:** the agent writes `"hp":"resources"` rows to
  its stdout, labelled `hivepaas.component=agent`. vlagent collects them with
  its buffer and forwards, and the API reads them with LogsQL. OBI's numbers
  take the same path: no OTLP into vlagent, no new network path, no buffer in
  the agent.

## Step 0: measured (2026-10-03), go

On a small Linux node:
- 1 vCPU, 961 MB of memory, Ubuntu 24.04, kernel 6.8, KVM;
- a node of a HivePaaS swarm, its agent and vlagent running.

The spike's Node.js-and-Postgres app was the load: 2,000 requests at 16 at
once, then 2,000 at 64. OBI v0.14.0 read the app by its container name and
exported Prometheus metrics. Everything was removed afterwards: containers,
images, files.

| Run | OBI's memory | Counted at 16 / 64 | Notes |
|---|---|---|---|
| privileged, default maps | 215 MiB | 2000 / 2000 | node down to 185 MB available |
| privileged, `global_scale_factor` -1 | 136 MiB | 2000 / 2000 | no map mismatch |
| privileged, -2 | 122 MiB | 2000 / 2000 | no map mismatch: v0.8.0's bug is fixed |
| privileged, -3 | 51 MiB | 1305 / 1203 | loses a third: not usable |
| **least capabilities, -2** | **73-98 MiB** | **2000 / 2000** | 2 "join target netns: operation not permitted" errors, no count lost |

- **Memory**: about 100 MiB at -2, the least capabilities; -3 loses counts.
- **Privileges**: no `--privileged`, and no `CAP_SYS_ADMIN`. It needs:
  - `CAP_BPF`, `CAP_PERFMON`, `CAP_SYS_PTRACE`, `CAP_NET_RAW`,
    `CAP_DAC_READ_SEARCH` and `CAP_CHECKPOINT_RESTORE`;
  - the host's PID namespace;
  - `/sys/kernel/tracing` and `/sys/kernel/debug` mounted.

  Joining another process's network namespace needs `CAP_SYS_ADMIN`; without
  it two errors are logged and nothing is lost.
- **Selection by `container_name`** works, and needs the Docker socket (read
  only): without it OBI cannot name containers, and instruments nothing. It
  picks up:
  - a container started after OBI, including a server that is a shell's child;
  - a container restarted.

  OBI is not restarted when tasks change.
- **Metrics under bursts are whole**: 2,000 of 2,000 at 64 at once, where
  spans lost 30% in the spike. Counts can come from OBI - but see the load
  test below: under sustained load, -2 dropped most of them until
  `wakeup_len` was lowered.
- **Attributes**: every series carries `container_name`, and needs no pid or
  cgroup lookup:
  - servers: route templated (`/u/*`), method, status;
  - HTTP clients: `server_address` and `server_port`, as the app named its
    peer (`hp-obi-app:18555`, a service's name);
  - database clients: `db_system_name`, `db_namespace`, `db_operation_name`,
    but `server_address` is `outgoing`, with no peer address.
- **Prometheus has it all**: every family needed is on OBI's Prometheus
  endpoint (`http_server_request_duration_seconds`,
  `http_client_request_duration_seconds`,
  `db_client_operation_duration_seconds`). The service graph is not, and is not
  needed.
- **Not measured**: OBI's CPU on a loaded 1-vCPU node. Measured since, on
  Docker Desktop: below.

## Load test (2026-10-03, Docker Desktop)

A Node.js app pinned to one CPU, as on a 1-vCPU node, and OBI on the same
CPU: `/` answers `ok`; `/chain` calls another app first. `oha` sent the load
from other CPUs, at fixed rates and as fast as it could. The VM: 10 CPUs,
7.6 GiB, kernel 7.0 (LinuxKit). Each run was made without OBI, with it, and
without it again; the numbers without are the two runs' range.

**Counted.** OBI reads its ring buffer once `ebpf.wakeup_len` events have
gathered, 500 by default. At -2 that is more than the buffer holds in time:
OBI dropped the rest, logging nothing.

| Capacity, `wakeup_len` | 500 req/s | 2,000 req/s | as fast as possible |
|---|---|---|---|
| Small (-2), 500 | 42% | 12% | under 1% (of ~44,000/s) |
| Small (-2), 500, `high_request_volume` | 42% | 12% | under 1% |
| Small (-2), 64 | 100% | 100% | 100% (of ~42,000/s) |
| Medium (-1), 500 | 100% | 100% | 100% |
| Large (0), 500 | 100% | 100% | 100% |

The agent now sets `wakeup_len` 64, 128 and 256 for Small, Medium and Large.
With 64, OBI on the app's CPU counted 1,404,887 of 1,405,000 requests and
378,831 of 379,036 calls over the four runs. At low rates the last few events
wait for the next wake-up: a scrape later, not lost.

**Cost**, Small, `wakeup_len` 64, OBI on the app's CPU:

| Run | Without OBI | With OBI |
|---|---|---|
| `/` at 2,000 req/s: app's CPU | 0.19-0.21 cores | 0.21, OBI 0.05 |
| `/chain` at 1,000 req/s: app's CPU | 0.26-0.28 cores | 0.26, OBI 0.05 |
| `/` as fast as possible | 54,000-56,000 req/s | 31,000 (app 0.77, OBI 0.23) |
| `/chain` as fast as possible | 17,000 req/s | 11,100 (app 0.83, OBI 0.17) |

- **Latency** at the fixed rates moved within the runs' noise: p99 3.3-7.2 ms
  without, 2.9-3.3 ms with.
- **CPU per request**: about 14 µs - 6.6 µs in the app's own probes, 7.4 µs in
  OBI. A trivial handler's throughput fell by 35-44% with its CPU full; one
  using a millisecond of CPU a request would lose about 1.4%, by the same
  14 µs. Below saturation the app's CPU did not move, and OBI took about
  25 µs an event.
- **Memory** grows with a node's CPUs and with load: Small took 144-229 MiB,
  Medium 199, Large 284 on these 10 CPUs, against 73-98 MiB at -2 on 1 vCPU.
  The capacities' figures are a 1-vCPU node's. The container's limit is
  512 MiB.

## Design

### Shape

```text
node:  OBI - a plain container the agent starts: host PID namespace, the
       least capabilities, the Docker socket read-only, maps sized by the
       node's capacity, in the agent's network namespace; its Prometheus
       endpoint on localhost
agent: scrapes it every 15 s; container_name -> the container's labels
       (hivepaas.app.id); drops what is no opted-in app's
       cumulative -> one row per series per 15 s: "hp":"routes", "hp":"calls"
       -> its stdout -> vlagent -> VictoriaLogs, as its resource rows
API:   per app: routes; dependencies (peer, rate, errors, latency)
UI:    the Metrics tab: routes in the HTTP view, a Dependencies view
```

### OBI on a node, run by the agent

- **Started by the agent.** The agent is a global Swarm service with the
  Docker socket. It starts OBI as a plain container, because Swarm cannot give
  a service the host's PID namespace. It keeps OBI running, removes it when the
  node or the feature is turned off, and replaces it when the agent restarts.
- **In the agent's network namespace** (`--network container:<agent>`): its
  Prometheus endpoint is on the agent's `localhost`, and nothing outside the
  node reaches it.
- **Its privileges**, as step 0 found them: the six capabilities, host PID
  namespace, tracefs and debugfs mounted, the Docker socket read-only to name
  containers; not `--privileged`.
- **Image pinned** in the release, as VictoriaLogs and vlagent are; pulled
  only on nodes where it is on.
- **What it watches**: the containers of opted-in apps on the node, by
  `container_name` patterns - a task's container is `<service>.<slot>.<task>`.
  Never a port, never all containers. Tasks that start, stop or restart are
  picked up as they come (step 0). OBI is restarted, debounced, only when the
  set of apps changes.
- **Its config**: maps sized by the node's capacity (below); metrics feature
  `application`; the Prometheus endpoint on localhost; traces off until
  phase 4.
- **Capacity, per node, chosen with a recommendation.** OBI allocates its eBPF
  maps whole when it starts. Their size is how many requests and connections
  it tracks at once, and its memory, idle or not. Too small loses what does
  not fit, silently: -3 lost a third. An administrator chooses per node; the
  default is HivePaaS's recommendation, by the node's memory:

  | Capacity | `global_scale_factor` | Memory | Tracked at once | Recommended for |
  |---|---|---|---|---|
  | Small | -2 | about 100 MiB | about 7,500 | nodes under 8 GB |
  | Medium | -1 | about 140 MiB | about 15,000 | 8 to 32 GB |
  | Large | 0 | about 215 MiB | about 30,000 | 32 GB and more |

  By what a node reads, which is a little under what it was sold with: from
  7.5 GB medium, from 30 GB large.

  The agent reports, in its status row, the node's memory, the recommended
  capacity and the one it runs with. Preflight asks for twice the chosen
  capacity's memory free.
- **Preflight**, run by the agent and reported to the settings page:
  - kernel 5.8 or later, with BTF (`/sys/kernel/btf/vmlinux`);
  - not a container-based VPS (OpenVZ, LXC);
  - tracefs present;
  - lockdown not `confidentiality`;
  - free memory for OBI.

  A node that fails is listed with why and gets no OBI. Its Traefik and
  resource numbers still work.

### Turning it on, and telling the agents

- **System → Logging**: "Routes and calls (eBPF)" - off by default. Its nodes
  are listed with their preflight, a per-node switch, and a capacity. The
  capacity is "Recommended (Small)" by default - the level shown for the
  node's memory - or Small, Medium or Large, each with its memory and how many
  it tracks at once.
- **App → Feature Settings**: on or off per app, off by default. An app's
  numbers are collected only while both it and its node are on.
- **The agents are told** as the Docker API feature tells them: a sync RPC,
  `SyncPerformance`, sent when either setting changes, and the agent reads
  the same state on its own tick. The agent then starts, reconfigures or
  removes its OBI.

### The agent's part

- **Scrapes OBI's Prometheus endpoint** every 15 s on localhost, with the
  Prometheus text parser: no OTLP receiver, no protobuf. Traces, in phase 4,
  bring OTLP back.
- **Attribution** by each series' `container_name`: the container's labels,
  from the Docker API, cached by name. A series that is no opted-in app's is
  dropped, never written.
- **Cumulative to per-interval**: the agent keeps each series' last value, and
  writes the difference every 15 s. A counter that went down (OBI restarted)
  starts the series again.
- **Rows**, one per series per 15 s, only where the count moved:
  - `"hp":"routes"`: app, method, route (templated by OBI), status class,
    count, errors, the histogram's bucket counts and sum;
  - `"hp":"calls"`: app, kind (http, sql, redis, grpc, kafka, dns...), peer -
    `server_address:server_port` for HTTP and RPC; for a database, its system
    and namespace (OBI gives no address) - operation, count, errors, buckets,
    sum.
- **Volume**: 50 routes and 10 peers is 60 rows every 15 s for a busy app,
  about 350,000 a day, a few MB after compression; nothing at all for an idle
  one.

### Peers

An HTTP or RPC client series names its peer as the app did. Apps in a
project call each other by service name or alias (`hp-obi-app:18555` in
step 0). The API resolves a name to the app behind it in the env; an IP to
today's task or service IP; anything else stays a host.

A database call has no address. It is named by its system and database (`postgresql`,
`postgres`). The API matches it to the env's database app of that engine and
database when there is exactly one, from its kind settings. History is shown
with today's names.

### API

- `GET .../apps/{app}/route-metrics?range=`: per route - requests, errors,
  p50/p95/p99 - and series. Quantiles come from the summed buckets in Go: a
  quantile of rows' quantiles would be wrong.
- `GET .../apps/{app}/dependency-metrics?range=`: per peer - kind, operations,
  requests, errors, p50/p95 - and series.
- Each answers `available: false` with a reason:
  - the feature is off for the app or its node;
  - no node of the app's passed preflight;
  - stored logs are off.
- MCP: `get_app_route_metrics`, `get_app_dependency_metrics`, as the HTTP and
  resource ones.

### Dashboard

- **Metrics → HTTP**: a Routes table under the paths. Routes are OBI's
  templates and count every request, including those from inside the project.
  It is shown only while OBI is on for the app; otherwise a line says how to
  turn it on.
- **Metrics → Dependencies**, a new view: the app's peers - other apps by name,
  databases, outside hosts - with requests a second, errors and p95, and a
  chart per peer kind.
- **System → Logging**: the switch, the nodes with their preflight and
  switches, the memory cost.
- **App → Feature Settings**: the per-app switch.

### Privacy, tenants

- OBI watches only opted-in apps' containers. The agent drops any series it
  cannot attribute to one.
- No SQL text, headers or bodies: OBI's metrics carry the operation, not
  the statement.
- The API matches the app id on every query, as the other metrics do.

## Risks

- **Memory**: about 100 MiB a node at -2 (215 at default maps). On a 1 GB node
  it is a tenth of the memory. The node switch and the cost shown are the
  answers.
- **The Docker socket in OBI**: read-only, to name containers. OBI holds the
  capabilities above already, on a node whose agent has the socket too. Its
  image is upstream's (`otel/ebpf-instrument`), pinned by digest in the
  release.
- **OBI is beta (0.x)**: the image is pinned, and the agent's mapping is tested
  against recorded scrapes of that version's Prometheus endpoint.
- **Kernels**: older ones, and container-based VPS, get nothing. Preflight
  says so.
- **Restarts** lose a few seconds of numbers. With `container_name` selection
  they happen only when the set of apps changes.
- **Peers named by today's IPs** (above).

## Phases

1. **Step 0**: measure v0.14.0, and decide. **Done** - go; the findings are
   above, and the design takes them.
2. **The agent**: preflight, OBI's lifecycle, the scraper, attribution,
   per-interval rows. Tests against recorded scrapes and a fake Docker; a run
   on the local stack. **Done**. As built:
   - **No `SyncPerformance` RPC**: there is no protoc on the build machine to
     generate it, and the agent already reads the database. It reads the
     settings every 30 s, as it reads the Docker API's. A switch turned takes
     30 s at most.
   - **The node's status** - preflight, wanted, running, apps - is a row the
     agent writes every minute while the feature is on and the logs are
     stored, `"hp":"obi"`. The settings show each node's latest, as the other
     rows are read.
   - **While the feature is off** - the default, on every node of every
     installation - the agent reads the setting every 30 s and does nothing
     else: one look for an OBI a previous agent left, when it starts, then no
     Docker call, no preflight, no status row, its scrape and status timers
     stopped. The API answers an app's routes and calls from the switches
     alone, and the settings page reads no status.
   - **The settings**: `LoggingSettings.Performance` {enabled, nodes by swarm
     node id}, and `AppFeatureSettings.PerformanceSettings` {enabled}.
   - **OBI's configuration** is copied to its image's root,
     `/hivepaas-obi.yaml`: the image has no shell and no `/etc/obi`.
   - **A live test** (`HP_TEST_OBI_DOCKER=1`) runs the agent's own start on
     Docker Desktop, in a stand-in agent's network namespace. It counts a
     Node.js app's requests by its swarm-named container. A busybox `httpd`,
     which forks a process per request, was not counted.
3. **The backend**: the two settings, `SyncPerformance`, the queries and the
   APIs, MCP. **Done**. As built:
   - **The nodes' settings** are their own endpoints,
     `GET/PUT /system/settings/logging/performance`, saved with the logging
     settings - their `updateVer` - and applying nothing: the agents read them
     within 30 s. A save of the logging settings keeps them. They cannot be
     saved before the logging settings are. The GET lists the swarm's nodes:
     hostname, role, memory, the recommended and the chosen capacity, and the
     latest status its agent wrote in the last 3 minutes; and the capacities,
     with their memory and how many they track.
   - **The app's switch** is `performanceSettings` in its Feature Settings,
     kept as it is by an update that leaves it out.
   - **The queries** sum the rows in VictoriaLogs - the agent's, by its
     daemon-written identity, and the app's id - by step and by group. Rows
     carry their buckets on fixed bounds (5 ms to 10 s, and +Inf), so the
     quantiles are read in Go from the summed buckets, interpolated within
     one; past 10 s a quantile reads 10000.
   - **The reasons**, in order: agent-unlabelled; the logs' (disabled,
     apps-not-collected, no-query-endpoint); performance-disabled;
     app-disabled; node-disabled (no node running the app's tasks is on);
     node-unsupported (none of those can, with the preflight's reasons). An
     app running nowhere has its past shown. `nodes` and `nodesCovered` say
     how much of the app is counted.
   - **Peers**: by the name called - the app's key (its alias in the env's
     network), its service's name, `tasks.<service>` - or an address of a
     task or a service today. A database by the engines its system is spoken
     by (postgresql: postgres; mysql: mysql, mariadb; redis: redis, valkey,
     dragonfly, keydb) and its database, when exactly one app matches; a
     cache, which names no database, by its engine when it is the only one.
   - **MCP**: `get_app_route_metrics`, `get_app_dependency_metrics`.
4. **The dashboard and docs**: the settings, the Routes table, the
   Dependencies view, a docs page. **Done**. As built:
   - **Routes is a view of its own** beside HTTP and Dependencies, not a
     table under the HTTP paths: an app without a domain has no HTTP
     numbers, and its routes - the requests from inside its project - are
     what it has. The HTTP view's note points to it. Both views say why when
     they cannot be shown, with a link to where it is turned on, and how many
     of the app's nodes measure it.
   - **System → Logging → Routes and Calls** is a page of its own, saved
     apart; after a save of either page both refetch, their version shared.
   - **Docs**: the app's Metrics page and System settings' Logging section.
5. Later: traces (the observability spec's phase 4); the Requests autoscale
   signal from OBI's server metrics, for internal traffic.
