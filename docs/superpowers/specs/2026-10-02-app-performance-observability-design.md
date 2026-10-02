# App performance: requests, latency and traces without instrumenting

## Why

HivePaaS shows an app's logs - VictoriaLogs and vlagent, the logging stack in
the `logging` env of the `hivepaas` project - and checks its health. It shows
nothing of how the app answers: how many requests, how many fail, how long
they take, what the app calls and where the time goes. Tools that do this
without touching the app's code (Coroot, DeepFlow) watch the kernel with eBPF;
they bring their own storage and UI, see every project on a node, and are an
admin's add-on at best.

This spec puts the same kind of data into the stack HivePaaS already runs, and
on the app's own page, scoped by the app's permissions.

**In scope:** HTTP and gRPC servers, the calls an app makes (HTTP, SQL, Redis,
DNS...), traces of slow and failed requests, CPU and memory per container; the
agent, VictoriaLogs, an API, a dashboard tab. **Not in scope:** continuous
profiling, a database's internals, browser monitoring, alerting rules on these
numbers (a follow-up), autoscaling on them (a follow-up that reads the same
data).

## What was measured

A spike on Docker Desktop (kernel 7.0.12, BTF, 10 CPUs), with a Node.js app
calling Postgres, labelled `hivepaas.app.id` as HivePaaS labels its containers.

- **The collector: OBI v0.8.0** (OpenTelemetry eBPF Instrumentation, Grafana
  Beyla donated to OpenTelemetry), `--privileged --pid=host`, exporting OTLP.
  - It produced a server span per HTTP request with method, **route already
    templated** (`/users/1` becomes `/users/*`), status and duration; a client
    span per Postgres query (operation, system - **not the SQL text**, which
    only its stdout printer shows) and per DNS lookup; and the parent links
    between them: HTTP, then `processing`, then the queries.
  - **No container id and no `process.pid`** in the OTLP resource. The pid is
    the end of `service.instance.id` (`<host>:<pid>`). From it, the process's
    `/proc/<pid>/cgroup` gives the container id, and the container's labels the
    HivePaaS app: that worked for every span.
  - **Discovery must be told exactly what to watch.** Selecting by port 8080
    instrumented the other containers listening there too (php, a qemu-run
    binary), read-only, and injected a script into the Node.js process it
    watched. The probes went when OBI stopped.
- **Storage in VictoriaLogs**, one JSON line per span: about 30 bytes a span
  compressed (15x) on repetitive test traffic; 50-100 is the planning figure.
  A request makes 2-5 spans.
- **Queries**: `stats by (route) count(), count() if (status>=500),
  quantile(0.95, duration_us)` gave rate, errors and latency per route;
  `/select/logsql/stats_query_range` the same as Prometheus-shaped series; a
  trace by `trace_id` its whole tree.
- **Overhead on the app**: none measurable at 500 requests/s (noise ±10%).
- **Memory of OBI: about 220-240 MB, mostly not OBI's process.** Its Go heap is
  ~20 MB; 168 MB is kernel memory charged to its cgroup - 57 preallocated LRU
  hash maps of 30000 entries (`ongoing_http`, `ongoing_tcp_req`...), which do
  not shrink with fewer CPUs. `maps_config.global_scale_factor` shrinks them
  (-2: ~106 MB) but in v0.8.0 also shrinks a pid map below what its Go side
  writes - thousands of "sizes of Go and BPF maps don't match" errors at
  start; -3 lost spans outright. Not usable until fixed upstream.
- **coroot-node-agent**, for comparison: ~412 MB (144 MB heap, 255 MB maps); it
  names Swarm containers itself (`/swarm/_/<service>/<slot>`) and exports
  aggregated metrics.
- **Spans are lost under bursts.** At 16 concurrent requests every server span
  arrived; at 64 (several hundred a second) about 30% did not. Counting
  requests from spans is therefore wrong.

## Revised (2026-10-03): HTTP numbers from Traefik's access log

Every request that reaches an app by its domain goes through Traefik, which
already writes an access log (`--accesslog=true`, CLF), and vlagent already
collects every container's stdout into VictoriaLogs. The HTTP numbers come from
there first; OBI becomes an opt-in for what Traefik cannot see.

**Checked on the local stack (Traefik v3.7):**

- The access log lines of real requests are in VictoriaLogs, as `_msg` text,
  in a stream named only by the Traefik container's log file (its id).
- LogsQL computes the numbers from them at query time, as the functions'
  metrics already do from their invocation lines: `extract` on the CLF line,
  then `stats by (router, status) count(), quantile(0.5|0.95, dur)` gave
  `router-a1-0@swarm 200: 18 hits, p50 3 ms, p95 40 ms`.
- An app's routers are `router-<appKey>-<n>@swarm` and its services
  `svc-<appKey>-<n>`: a line maps back to its app. `ServiceURL` names the
  replica that answered.
- `--accesslog.format=json` with `--accesslog.fields.queryparameters.defaultmode=drop`
  is accepted by v3.7 and strips the query: `/login/cb?token=...&email=...` is
  logged as `/login/cb`. JSON adds `OriginStatus` (what the app answered; 0
  when the request never reached it), `DownstreamStatus` (what the client got,
  after middlewares), `OriginDuration` and `Duration` (nanoseconds),
  `ServiceName`, `RouterName`, `RetryAttempts`.

**Decided:**

- **Traefik's access log becomes JSON**, the query string dropped, headers
  dropped (Traefik's default), **the client's IP kept** (to trace an abuser in
  the logs). HivePaaS sets these arguments on its Traefik service; an operator
  who turns the access log off in Config Options gets a Performance tab that
  says why it is empty.
- **Traefik's log lines are labelled** so that a query matches them exactly,
  not by content and not by a container id that changes at each restart: the
  Traefik service's log options carry a HivePaaS identity, as an app's
  `LabelLogAppID` does.
- **Numbers per app**: requests, errors (`OriginStatus >= 500`, and 502/504
  from Traefik when the app is down), requests blocked before the app
  (`OriginStatus` 0 with a 4xx), p50/p95/p99 of `Duration` and of
  `OriginDuration`, per replica by `ServiceURL`; per path for the top paths, as
  the functions' metrics count their top 20.
- **Paths are not templated** by Traefik: `/users/1` and `/users/2` are two
  paths. The query reduces numbers and ids to `*` (`replace_regexp`) and keeps
  the top paths; route templates stay OBI's.
- **What Traefik cannot see** - calls between apps inside a project, calls to
  databases and outside hosts, the time inside the app - stays OBI's, opt-in
  per node, until its memory comes down.
- **Cost**: a line per request, as today with CLF - JSON is somewhat longer.
  The figure at 100 requests/s is to measure (planning: hundreds of MB a day
  before retention); there is no sampling in Traefik.

## Design

### Shape

```
node:  OBI (run by the agent) --OTLP traces+metrics, localhost--> agent
agent: pid -> container -> labels (app, project, env); drop what is not an app
       metrics -> one row per series per interval
       spans   -> kept when failed, slow, or sampled
       cgroup  -> one row per container per interval
       --> VictoriaLogs (/insert/jsonline), the logging stack's backend
API:   LogsQL built in one place, always matching the app id
UI:    the app's Performance tab
```

### Collection: OBI, run and scoped by the agent

- The agent, which already runs on every node with the Docker socket and the
  host mounted, starts OBI as a plain container (Swarm services cannot be
  privileged or share the host's pid namespace), keeps it running, removes it
  when the feature is turned off, and on nodes that join later.
- **The agent decides what OBI watches**: the pids of the containers of apps
  that have the feature on, by `OTEL_EBPF_TARGET_PID` (or discovery by
  container name, `hivepaas`-named services only). Never a port, never all.
  The list changes as tasks start and stop: OBI is restarted with the new list,
  debounced; the gap is a few seconds of missing data.
- **Metrics for counting, traces for looking.** OBI exports both over OTLP to
  the agent on localhost: request counts, errors and duration histograms come
  from its metrics, aggregated in OBI and not lost when spans are; spans only
  feed the trace view.
- The image is pinned in the release, like VictoriaLogs and vlagent.
- **Preflight per node**, shown in the settings: kernel >= 5.8 with BTF
  (`/sys/kernel/btf/vmlinux`), not OpenVZ/LXC (`systemd-detect-virt`), tracefs
  present, kernel lockdown not `confidentiality`. A node that fails is listed
  with the reason and gets no OBI; the cgroup numbers below still work there.

### The agent's part

- **Receives OTLP/HTTP** (`/v1/traces`, `/v1/metrics`) on an address only the
  node's OBI reaches.
- **Attributes** each resource to an app: pid from `service.instance.id`, then
  `/proc/<pid>/cgroup`, then the container's labels (cached by container id).
  Resources that are no HivePaaS app's container are dropped, never stored.
- **Resolves the peer** of client spans and metrics - an IP and port - to the
  app behind it when it is one (the agent knows the node's containers' IPs; the
  service's VIP for another node's), so "app A calls database B" is a row, not
  an address.
- **Metrics** are written as one row per series per interval (15 s): app,
  route, method, status class, peer app or address, count, error count, and the
  histogram's buckets or a few quantiles. A hundred routes is a hundred rows
  every 15 s, whatever the traffic.
- **Spans** are kept when the request failed (5xx, error status), when it was
  slower than a threshold (per app, 1 s by default), and otherwise at a sample
  rate (1% by default); internal spans (`in queue`, `processing`) are kept only
  inside a kept trace. A trace is kept or dropped whole, decided on its server
  span.
- **Container resources** from cgroup v2 every 15 s: CPU usage and throttling,
  memory and its limit, OOM kills, network bytes; one row per container.
- **Writes** batches to **the node's vlagent** (`/insert/jsonline`, port 9429),
  with `app_id` and `kind` (metric, span, resource) as stream fields: vlagent
  already buffers on disk and retries when VictoriaLogs is away, so the agent
  keeps no buffer of its own. The agent reaches the vlagent of its own node
  (attached to its network, or a port bound to localhost).

### Storage

The logging stack's VictoriaLogs, with its retention. Rows carry `app_id`,
`project_id`, `env`, `node`, `kind`, and the fields above, under a `perf.`
prefix so they never mix with the apps' own log fields. At 15 s rows and a 1%
sample, a busy app costs megabytes a day, not gigabytes.

When the numbers outgrow LogsQL at query time, the metric rows move to
VictoriaMetrics and the spans stay; the API hides the difference.

### API

`GET /projects/{p}/{env}/apps/{app}/performance?from=&to=&step=` and friends
(routes, calls, traces, a trace by id), under the app's read permission.
LogsQL is built in one function, as `BuildQuery` builds the log queries:
values quoted, the `app_id` match always there, nothing from the request
becoming syntax. Series come from `stats_query_range`.

### Dashboard

A **Performance** tab on the app:

- requests per second, error rate, p50/p95/p99, CPU and memory, over the
  chosen range;
- the routes, slowest and most failing first;
- the calls the app makes - other apps, databases, outside hosts - with their
  rate and latency;
- slow and failed traces, each opening as a waterfall, with a link to the
  app's logs at that time.

The dashboard has no chart library yet; a small one (uPlot) is added.

### Settings

- **System → Logging** (where the stack is turned on): "Collect performance
  data" with the node preflight; off by default, as it costs ~220 MB per node
  with OBI's maps at their default size.
- **App → Feature Settings**: on or off per app, the slow threshold and the
  sample rate.

### Multi-tenancy and privacy

- OBI is told only the pids of apps with the feature on; the agent drops any
  resource it cannot attribute.
- The API matches the app id on every query.
- SQL text and request headers and bodies are not collected (OBI's OTLP export
  leaves the SQL text out; header capture stays off). URLs are reduced to the
  route template.

## Risks and open questions

- **OBI's memory.** 220-240 MB per node is too much for a 1-2 GB VPS. Until
  `global_scale_factor` works, the feature is for nodes with room; to raise
  upstream with the measurements above, and to measure on a 2-vCPU node.
- **OBI is beta** (0.x); its attributes and options move. The image is pinned
  and the agent's mapping is tested against it.
- **Bursts drop spans.** Counts come from metrics; a trace view with gaps is
  acceptable. To check that OBI's metrics are complete under the same bursts.
- **Restarting OBI on every task change** loses a few seconds; whether OBI can
  reload its target list in place is to check.
- **No eBPF** (old kernels, container-based VPS): those nodes get the cgroup
  numbers and nothing else.

## Phases

1. **Container resources** (done: backend bf7ee931, dashboard 499e0cdb) - the
   agent's cgroup rows, the API, the tab with CPU and memory. No eBPF; every
   node. As built, the transport changed: the agent writes its rows to its own
   stdout, `"hp":"resources"`, and the log collector takes them with its other
   lines - no network path, vlagent's buffer, the external backend's
   credentials and forwards. Its lines carry `hivepaas.component=agent`
   (pkg/logidentity, shared with Traefik's); an existing install gets it with
   the agent's next image update. Rows are written only while stored logs are
   on. Checked against a real container's cgroup on a local node.
2. **HTTP numbers from Traefik** (done: backend b0822c28, dashboard 58349d98)
   - the access log as JSON (query dropped, IP kept), its lines labelled, the
   API's queries per app, path and replica, the tab's requests, errors and
   latency. No eBPF; every node. As built:
   - Traefik's routers, services and middlewares are named by the app's id
     first (f45814d3): by key, two apps of one key - or one named "app", like
     HivePaaS's own - made Traefik drop both routes. An app's lines are its
     services', `^svc-<id>-[0-9]+@swarm$`, matched exactly.
   - The lines are the proxy's by `attrs.hivepaas.component=traefik`, a
     container label json-file copies into each; a live test checks that a
     line an app printed is not counted.
   - New installs have it in the stack. An existing one gets it when an
     administrator saves Config Options with Access Log on - that option now
     writes the JSON form, and the apply adds the label - or with Traefik's
     next image update; until then the tab says why it is empty.
3. **Calls and inside the app** - OBI run by the agent, opt-in per node:
   calls between apps, to databases and outside hosts, route templates.
4. **Traces** - sampling, the trace list and waterfall, the link to logs.
