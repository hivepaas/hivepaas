# Apps that scale themselves

## Why

Functions scale themselves (spec `2026-10-03-function-autoscale-design.md`):
their replicas follow the calls their runtime logs. Every other app still has
its **Replicas** set by hand, in Availability & Scaling: kept for the peak all
day, or short of it.

HivePaaS now stores two numbers that say how busy any app is, with no new
component:

- **Requests**, from Traefik's JSON access log: every request reaching the app
  by its domains, with its duration (`Duration`, ns), its service
  `svc-<lowercased app id>-<n>@swarm`.
- **CPU**, from the agent's `"hp":"resources"` rows: every container's cores,
  and its limit (`cpuLimit`, from `cpu.max`), every 15 s.

This spec scales an app's replicas from them, with the function autoscale's job,
setting, decision and history.

**In scope:** Replicated apps; between Min and Max, from 1 up; requests and CPU.
**Not in scope:** memory (it does not fall when replicas are added: scaling on
it only grows); queue length and custom metrics; scaling to zero; schedules
(Min by time of day); vertical scaling. All are follow-ups.

## The signals

An app scales on one or both, its choice; with both, the larger answer wins, as
Kubernetes' HPA does.

**Requests in flight per instance**, by Little's law, as for functions:

```text
inFlight = sum(min(OriginDuration, window) of the requests that ended in the window) / window
desired  = ceil(inFlight / requestsTarget)      requestsTarget: requests per instance
```

- From Traefik's access log, so only requests through the app's domains: a
  worker, or an app called inside its project, has none - it scales on CPU.
- One LogsQL query a run for every app on it: the access log lines of the last
  60 s whose `ServiceName` matches `^svc-(<id>|<id>...)-[0-9]+@swarm$`, the app
  id cut out of it with `replace_regexp`, `sum(Duration)` and `count()` by it;
  and both again over the window's last 15 s, `if (_time:>=...)`, for a
  burst.
- A request's time is its `OriginDuration`, how long Traefik waited on the
  app: 0 for one no replica answered, which is counted but does not raise
  `inFlight`. Not `Duration`, which adds Traefik's time and a slow client's;
  and not filtered by `OriginStatus`, which Traefik 3.7 writes 0 for nearly
  every request the app answered (22 of 61,460 carried one, on a local
  install) - a filter on it read a loaded app as idle.
- **A request counts for the window at most** (`math min(Duration, window)`):
  a WebSocket, a stream of events or a long poll is logged when it ends, with
  its whole duration - an hour's, ending in the minute, would read as 60
  requests at once. While it is open it is not seen at all: an app serving
  mostly those scales on CPU. A duration that is not a number is not summed -
  `min()` of one is the window.
- **The window ends 10 s before now**, for every signal: the lines of the last
  seconds may not have arrived, later from a node further away. With several
  Traefik replicas each request is still logged once, by the one that served
  it, and the agent on every node ships its lines: the sum is the app's.

**CPU per instance**, the HPA formula:

```text
utilization = sum(cpu) / sum(cpuLimit)   over the app's containers, the last 60 s
desired     = ceil(current * utilization / cpuTarget)      cpuTarget: 70 % by default
```

- One query a run for every app on it: the agent's rows of the last 60 s, each
  container's average, summed by app.
- **The limit**: `cpuLimit` from the container's cgroup. An app without a CPU
  limit falls back to its CPU **reservation**, from its service spec; with
  neither, CPU cannot be used for it, and the section says so.
- **Tolerance**: within 10 % of the target, no change - CPU moves a little
  every run, and a replica each way every minute is noise.

## The decision

The function autoscale's, its hysteresis shared, the signal apart:

1. **Skip** an app that is stopped, being deployed, or not Replicated.
2. **Hold** when a signal it uses cannot be read: logging off, the query
   failed, the access log not ready (`AccessLogReadiness`), the agent not
   labelled. No data is never read as no load. With two signals, one that can
   be read is enough.
3. `desired` from the signals, clamped to `[min, max]`.
4. **Up**: after 2 runs in a row above (30 s), then every run that still asks
   for more: the count is kept after a scale-out, and a run not above starts
   it again. **At once for a burst**: the requests of the window's last 15 s
   needing twice the replicas running or more - Knative's panic mode, at its
   default - the larger of the two asks taken; an app does not say it turned
   a request away, but a burst is how that starts. CPU has no burst: it stops
   at its limit, and a starting container's would read as one. After a
   scale-out, a **cooldown** of 1 min before the next one from CPU: a starting
   container burns CPU, and would ask for more. A run grows the replicas to
   twice over or by 4, the larger, at most - the HPA's default.
5. **Down**: once below for the scale-in delay, half the gap a run - as
   functions.
6. **The cluster is full**: a scale-out Swarm cannot place - tasks pending for
   want of CPU or memory reservations, or of a node the constraints allow -
   is not followed by another. When the service's running tasks
   (`ServiceStatus`, one `ServiceList` a run) stay below its desired for 2 min,
   the app holds its count, and its section says why, until they run.

## What it refuses, what it warns of

- **Not Replicated**: paused, as a function is.
- **A port published in host mode**: one replica a node at most - refused,
  with why.
- **A writable volume or bind mount**: allowed, with a warning in the section -
  the replicas share the volume on one node, and each has its own on another.
  A database must not autoscale.
- **Sessions in memory**: a warning in the docs, as for any app with more than
  one replica.

## Settings

The same **Autoscale** section, in Availability & Scaling, for every app; the
same `app-autoscale` setting, with its fields for apps:

| Field | Default | |
|---|---|---|
| **Autoscale** | off | |
| **Min Replicas**, **Max Replicas** | 1, 5 | as for functions |
| **Requests Per Instance** | off | requests in flight one instance takes: 1 to 1000 |
| **CPU Target** | 70 % | of the limit, or reservation, per instance: 10 to 100 |
| **Scale-in Delay** | 5 min | as for functions |

- At least one of Requests and CPU.
- A function keeps its **Target** (of its Concurrency) and does not show these.
- Each signal says whether it can be used now, and why not: no domain, the
  access log not ready, no CPU limit nor reservation, the agent not labelled.

## The job, the service

- **One job for every app**: `function-autoscale` becomes `app-autoscale`, and
  the job's setting is renamed by `EnsureJob`, which finds it by either kind -
  no migration.
- `functionautoscaleservice` becomes `appautoscaleservice`: one run reads the
  invocation lines for the functions, the access log and the agent's rows for
  the other apps - three queries at most, whatever the number of apps.
- `decide` splits into the signal - a function's, an app's - and the shared
  hysteresis, `settle`.
- The saved task's output stays as it is: an app's scaling lists its in-flight
  requests, its CPU, the signal that decided.

## API, dashboard, MCP

- `GET/PUT .../apps/{app}/autoscale` take any app; the answer adds each
  signal's readiness, and the warnings.
- The **Metrics** tab draws the replicas on the **Requests** and **CPU** charts,
  as on a function's **Calls**.
- MCP: the same `autoscale` kind of the app settings tools.
- Docs: Resources and placement, a section on autoscale.

## Risks

- **Reaction time, 15 to 45 s for a burst of requests, 45 s to 2 min
  otherwise**: a request is logged when it ends, the agent's rows come every
  15 s, 2 runs above. Min is the answer for a known peak.
- **A slow start**: an app that takes a minute to answer is scaled before it
  helps, and more if CPU climbs meanwhile - the cooldown bounds it; a health
  check keeps Traefik from sending it requests before it answers.
- **Scale-in drops requests**: Swarm stops a task with SIGTERM while Traefik may
  still route to it for a moment - an app should finish its requests on
  SIGTERM, with a stop grace period; the docs say so.
- **Internal traffic** is not seen by the Requests signal: CPU covers it.

## Phases

1. The backend: the setting's fields, the service renamed and generalized,
   the two queries, the decision split, the refusals and the full-cluster
   hold, the API for every app; tests against a fake swarm and a live
   VictoriaLogs with seeded Traefik and agent rows. **Done**. As built: a run
   lists the apps' services once (`ServiceList`, status on) rather than
   inspecting each; a signal whose query fails, or whose source is not
   marked, holds only the apps that scale on it alone; CPU with no row for a
   running app holds it (the agent not reporting is not an idle app);
   `paused` is answered whether autoscale is on or off, so the section can
   say it cannot be turned on; the replicas on the Requests and CPU charts
   move to phase 2, with the dashboard.
2. The dashboard's section for apps, the replicas on the Requests and CPU
   charts, MCP, docs. **Done**. As built: the section is
   the same for every app, its fields by whether it is a function; each
   signal has its own switch and target, and says under it why it cannot be
   read; an app's unreadable signals do not stop it being turned on in the
   form - the API refuses, with why - but a port in host mode does; the
   texts of the access log's and the agent's reasons are shared with the
   Metrics tab. Later: only turning autoscale on is refused while what it
   reads cannot be read - one already on is saved, and says it is paused -
   with errors of their own (`ERR_AUTOSCALE_*`) that say why in words; the
   form refuses it first.
3. A burst and a rising load, for apps and functions alike (2026-10-04).
   **Done**. The requests' and the calls' queries sum the window's last 15 s
   apart in the same query - 4 to 10 ms more a run, measured on 120,000
   access log lines a minute; a burst there scales out at once; the count of
   runs above is kept after a scale-out, so a load still rising is given more
   every run, not every other; and every scale-out is bounded, to twice over
   or by 4 a run.
4. Later: memory, queue length and custom metrics, Min by schedule, scaling to
   zero.
