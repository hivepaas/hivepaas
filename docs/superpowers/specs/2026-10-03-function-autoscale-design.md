# Functions that scale themselves

## Why

A function runs **Concurrency** calls at once per instance (16 by default) and
answers `429` past that. Its **Replicas** are set by hand, in Availability &
Scaling. A function with a busy hour either keeps replicas for the peak all day
or turns callers away during it.

HivePaaS now stores, for every call, the line the function's runtime writes
(`"hp":"invocation"`, with `durationMs`, `status`, `outcome`) in the logging
stack's VictoriaLogs, and reads it for the function's Metrics. The same lines
say how busy the function is: this spec scales a function's replicas from them,
with no new component.

**In scope:** functions; replicas between a minimum and a maximum, from 1 up.
**Not in scope:** scaling to zero (nothing would be left to answer the first
call, nor to log the load that decides it - it needs an activator in front, as
Sablier is); other apps (a follow-up, from Traefik's requests and the agent's
CPU rows); vertical scaling.

## The signal

**Calls in flight**, by Little's law: over a window of `T` seconds, the time
the function's calls spent being handled, summed, divided by `T`, is the number
it was handling at once on average.

```
inFlight = sum(durationMs of the calls that ended in the window) / (T * 1000)
desired  = ceil(inFlight / (Concurrency * target))      target: 0.7 by default
```

- Read from the invocation lines, not from Traefik's access log: they count
  every call - by its domain, inside the project, scheduled - and a function's
  Concurrency is per call, not per request through the proxy.
- **Throttled calls** (`outcome: "throttled"`, the runtime's `429` past
  Concurrency) are the urgent signal: any in the window means too few
  instances now, whatever the average says. They take no time, so they do not
  raise `inFlight`; they raise `desired` to at least `current + 1`, more for
  many (`current * (1 + throttled / calls)`, capped at doubling).
- A line is written when a call ends: a call in progress is counted once it
  is done. With the default 30 s timeout, the window (60 s) holds most of them.

One LogsQL query a run, for **every** function with autoscale on, grouped by
app: `"hp":"invocation"` lines of the last 60 s, `sum(durationMs)`, `count()`,
`count() if (outcome = "throttled")`. Its cost does not grow with the number of
functions.

## The decision

Every run, for each function:

1. **Skip** a function that is stopped (`replicas: 0` is how HivePaaS stops an
   app), being deployed, or whose service is gone.
2. **Hold** - change nothing - when the numbers cannot be trusted: the query
   failed, the logging stack is off or not answering. No data is never read as
   no load.
3. `desired` from the signal, clamped to `[min, max]`.
4. **Up fast**: above the current replicas, scale at once (throttled calls), or
   after 2 runs in a row above (the average).
5. **Down slow**: below the current replicas, only once it has been below for
   the **scale-in delay** (5 min by default), and by at most half the gap a
   run, so a quiet minute does not undo a busy hour.
6. Scaling is a service update of `Mode.Replicated.Replicas` alone: the
   running instances are not restarted.

The state between runs - since when below, how many runs above - is kept in
Redis, per function; lost, it starts again, which only delays a scale-in.

## Where it runs

A **periodic job** of the worker (`PeriodicKind` `function-autoscale`, a global
`periodic-job` setting), at the periodic base interval, 15 s:

- The setting **exists only while a function has autoscale on**: created when
  the first one turns it on, removed when the last one turns it off. No
  function with autoscale, no runs at all.
- Periodic jobs already run under a Redis lock per job, so two workers never
  decide at once.
- A run that changes nothing **saves no task** (`SaveTask` false); a run that
  scales one saves one, whose output says what moved and why: function, from,
  to, in flight, throttled calls, Concurrency, target. That is the function's
  scaling history, without a task row every 15 s.
- The interval stays at 15 s while it runs: a function at rest is when the next
  burst comes, and a slower run then is a slower scale-out.

## Settings

The function's **Settings → Availability & Scaling**, an **Autoscale** section:

| Field | Default | |
|---|---|---|
| **Autoscale** | off | |
| **Min Replicas** | 1 | 1 or more |
| **Max Replicas** | 5 | at least Min, at most 50 |
| **Target** | 70 % | of Concurrency, per instance: 10 to 100 |
| **Scale-in Delay** | 5 min | 1 to 60 min |

- Stored as an app-scoped setting of its own (`app-autoscale`), not in the
  function's deployment: it is the app's scaling, as Replicas is, and other
  apps take it later.
- With Autoscale on, **Replicas** is shown read-only as the current count.
  Turning it on scales to at least Min at once; turning it off leaves the
  replicas as they are.
- A deployment does not touch the replicas (it already does not), so an
  autoscaled function keeps its count across deployments.
- **Logging off**: Autoscale cannot be turned on, and a function that has it
  says it is paused, with a link to System → Logging.

## API, dashboard, MCP

- `GET/PUT /projects/{p}/{env}/apps/{app}/autoscale`: the settings, and the
  current replicas, whether paused and why, the last decision.
- The function's **Metrics** tab draws the replicas over the range on its
  Calls chart; **Settings → Availability & Scaling** lists the last scaling
  events, from the saved tasks.
- MCP: the settings in `get_app` and their update in the app settings tools,
  as Replicas is.
- Docs: the functions page, a section on scaling.

## Risks

- **Reaction time, 20 to 60 s**: the line is shipped by vlagent within a few
  seconds, a run every 15 s, a new instance starts in seconds (more for a
  runtime that installs at start). A burst within seconds is answered `429`
  until then; Min is the answer for a known one.
- **Logs dropped on one node** look like less load there; scale-in is slow
  and bounded, so the cost is a step down, not an outage.
- **A long call** counts when it ends: a function whose calls run for minutes
  is better scaled by hand or by a longer window (a follow-up).
- **Stateful functions**: a function that keeps state in its instance should
  not autoscale; the docs say so.

## Phases

1. The setting, the periodic job, the query and the decision, the service
   update and the saved tasks; tests against a fake swarm and a live
   VictoriaLogs. **Done**, with `GET/PUT .../apps/{app}/autoscale`. As built:
   the job's setting is made once and turned on and off by its status, not
   removed; a function whose service lacks its log identity is passed over,
   as its calls would read as none; a saved task is the job's, one a run,
   listing every function it scaled - the history per function filters its
   output by app.
2. The settings section, the paused state, the replicas on the Metrics tab and
   the history; MCP; docs.
3. Later: other apps, from Traefik's requests and the agent's CPU rows.
