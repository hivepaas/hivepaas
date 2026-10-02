# Function metrics

## Why

A function's author cannot see how it is doing: how often it is called, how
often a call fails, how long calls take. Every call already says so: each
runtime writes, on standard output, one line per call -

```json
{"hp":"invocation","requestId":"…","method":"GET","path":"/","status":200,"durationMs":12.5,"outcome":"ok"}
```

- and HivePaaS already collects an app's output into VictoriaLogs and queries it
with LogsQL (the app's Logs, History tab). Metrics are those lines, counted.

**In scope:** a function's calls, failures and durations over a time range, on a
tab of its own. **Not in scope:** alerts, autoscaling, metrics of apps that are
not functions, a metrics store (Prometheus, VictoriaMetrics).

## What is counted

- **A call** is an invocation line in the app's logs: HTTP requests, and
  scheduled calls going through `serve` (`hivepaas-runtime call`, runtimes
  1.2.0 and later). Test runs and calls through `invoke` write their lines
  elsewhere, and are not counted.
- **Failed**: an outcome other than `ok` - `error`, `timeout`, `throttled`,
  `request_too_large`, `response_too_large`, `not_loaded`. A status the handler
  chose, 4xx or 5xx, is not a failure of the call; 5xx answers are counted
  apart, as `errors5xx`.
- **Durations**: `durationMs`, how long the handler ran - not the network nor
  the proxy; p50, p95 and p99, as VictoriaLogs' `quantile` gives them, which is
  close rather than exact (two runs over the same lines gave a p50 of 20.0 and
  20.1 ms).
- Every instance of the function, together.

## Measured

VictoriaLogs v1.52.0, the version HivePaaS deploys, on this machine, 300,000
invocation lines over 24 hours among 1.1 million lines of 21 apps:

| Query | Time |
|---|---|
| totals over 24 h | 180 ms |
| 1 h by 1 min | 26 ms |
| 6 h by 5 min | 110 ms |
| 24 h by 15 min | 260 ms |
| 24 h by outcome | 100 ms |

The buckets' calls added up to the totals. A bucket without a call is not
returned, and buckets come in no order.

## The query

One LogsQL query per answer, built in `services/logging/victorialogs` beside
`BuildQuery`, by the same rules: the app's id and every value in a Go-quoted
literal, the structure fixed:

```
_time:<range> AND "attrs.hivepaas.app.id":="<app>" AND "\"hp\":\"invocation\""
| unpack_json from _msg fields (hp, outcome, status, durationMs) result_prefix "app."
| filter "app.hp":="invocation"
| stats by (_time:<step>) count() calls, count() if ("app.outcome":!="ok") failed,
    count() if ("app.status":>=500) errors5xx,
    quantile(0.5, "app.durationMs") p50, quantile(0.95, "app.durationMs") p95,
    quantile(0.99, "app.durationMs") p99
| sort by (_time)
```

- The phrase `"hp":"invocation"` narrows the lines before they are unpacked; the
  `filter` after it is what decides.
- Unpacked fields carry the `app.` prefix, as every unpacking in HivePaaS does,
  so that a line cannot overwrite the collector's own fields.
- A second query, without `by (_time…)`, gives the totals, and a third the count
  by outcome.

A handler can print a line that looks like an invocation line, and it is
counted: a function can distort only its own metrics.

## The API

`GET /projects/{id}/{env}/apps/{appId}/function-metrics?range=24h`

- `range`: `1h`, `6h`, `24h` or `7d`; anything else is refused. The step is the
  range's: 1 min, 5 min, 15 min, 1 h.
- The answer: the range, its step, the totals (`calls`, `failed`, `errors5xx`,
  `p50`, `p95`, `p99`, and `byOutcome`), and the series - one point per step,
  oldest first, with the steps without a call filled with zeros and no
  durations.
- Who may read the app's logs may read its metrics; the app must be a function.
- When there is nothing to query, the answer says why, as the History tab does:
  logging disabled, apps' logs not collected, a backend without a query
  endpoint. A range longer than the logs' retention is answered for what is
  kept, and says so.

## The dashboard

A **Metrics** tab on a function's page:

- the range - 1 h, 6 h, 24 h, 7 d - **24 h by default**, the last chosen
  remembered in the browser;
- the totals: calls, failure rate, p95;
- a chart of calls and failed calls over time, and one of p50, p95, p99;
- the count by outcome;
- the reason in place of the charts when there is nothing to query, linking to
  the logging settings.

Charts use shadcn/ui's `chart` component, on `recharts` - the dashboard's first
chart, and a new dependency.

## Testing

- The query builder: the query for each range, values quoted, the app's id
  required.
- The service: the step by range; buckets sorted and gaps filled; totals and
  outcomes read; the reasons; a range past retention.
- The API: a function's metrics; refused for an app that is not a function, for
  a range not allowed, for a caller who may not read the app's logs.
- Live, against a VictoriaLogs container, as `query_live_test.go` does: lines
  ingested, the answer's counts and series.
- The dashboard: `lint:ci`, `build`; the tab in the browser, by the user.

## Later

- Alerts on failure rate or p95.
- A metrics store, when autoscaling needs numbers every few seconds.

## Changes after the plan

Writing the plan (`docs/superpowers/plans/2026-10-02-function-metrics.md`)
settled what follows:

- **The charts use `recharts` directly**, in a component of the tab's own, its
  colors the theme's (`--chart-*`, `--destructive`, `--border`), rather than
  shadcn/ui's `chart` wrapper, which the dashboard does not have and would come
  through the shadcn CLI.
- **`recharts` 3.10.1 is added with yarn**: `yarn.lock` is the dashboard's lock
  file - it changes with `package.json` - and `package-lock.json` is behind it
  (it lacks `@fontsource-variable/geist`), so `npm` would rewrite both.
- **The reasons' texts** - why the logs cannot be read - move to a file of their
  own, shared by the History tab and the Metrics tab.
- **A range longer than the logs' retention** starts at the first whole step the
  logs still hold.
- **The live test** (`stats_live_test.go`) ran against VictoriaLogs v1.52.0: an
  app's handler lines that mention an invocation, and another app's line naming
  this one, are not counted.

## Added: calls by path

On a server the first day's calls were mostly scanners - `/.env`, `/.git/config`,
`/wp-login.php` - which the totals cannot tell from a function's own. A fourth
query counts the calls by method and path, the 20 most called first, ties in
path order:

```
… | stats by ("app.method", "app.path") <the counts> | sort by (calls desc, "app.path", "app.method") limit 20
```

The counts gain `errors4xx` (`"app.status":>=400 "app.status":<500`), in the
totals and by path: a handler that answers 404 to what it does not serve shows
the scanners there. The Metrics tab lists the paths under the charts. A path is
counted as the runtime wrote it, ids and all: a function serving `/users/123`
has as many paths as users, and only the 20 most called are listed.
