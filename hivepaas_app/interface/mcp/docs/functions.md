# Functions on HivePaaS

A function is an app whose deployment is code and a runtime: HivePaaS builds it
on the runtime's image and runs it as a server that calls the handler once per
request. It is created with `plan_create_function`, its code is changed with
`plan_update_app_settings` (kind `deployment`, `functionSource`), and it is
deployed again with `plan_redeploy_app`.

## Runtimes

| Runtime | Language | Entrypoint by default | Handler by default |
|---|---|---|---|
| `node24` | JavaScript, or TypeScript with its types removed (`.ts`, `.mts`, `.cts`) | `index.js` | `default` (the default export) |
| `bun1` | JavaScript or TypeScript, transpiled whole | `index.ts` | `default` |
| `python313` | Python | `main.py` | `handler` |
| `go127` | Go, compiled | `.` (the package's directory) | `Handle` |

TypeScript on `node24` allows only syntax that can be erased: no `enum`, no
`namespace` holding values, no parameter properties, no decorators; an import
names the file with its extension (`./lib/util.ts`) and a type is imported with
`import type`. On `bun1` all of TypeScript runs. Types are checked by neither.

## The handler

```js
// node24, bun1
export default async function (req, ctx) {
  ctx.log(`${req.method} ${req.path}`)
  return { status: 200, body: { hello: req.query.name ?? "world" } }
}
```

```python
# python313 - a function or a coroutine
def handler(req, ctx):
    return {"status": 200, "body": {"hello": req.query.get("name", "world")}}
```

```go
// go127 - module with a go.mod; the runtime's package is added when it builds
package function

import (
	"context"

	"github.com/hivepaas/function-runtimes/hivepaas"
)

func Handle(ctx context.Context, req *hivepaas.Request) (*hivepaas.Response, error) {
	return hivepaas.JSON(200, map[string]string{"hello": req.Query.Get("name")})
}
```

- **The request**: method, path, query (first value; every value as
  `queryAll` / `query_all` / `Query[name]`), headers in lower case, body as
  bytes, `text()` or `json()`.
- **The response**: an object with `status`, `headers`, `body`. A body that is
  text is sent as `text/plain`, bytes as `application/octet-stream`, anything
  else as JSON. Returning nothing answers 204.
- **The runtime answers itself** 500 when the handler throws (the error goes to
  the log, never the client), 504 past the timeout, 429 past the concurrency,
  413 for a body over the limit.

## Libraries

A manifest beside the code is installed when the function is built:
`package.json` (npm for `node24`, bun for `bun1`), `requirements.txt` (uv for
`python313`), `go.mod` (`go127`). A lock file is made when there is none.
System packages (`apt`) go in `functionSource.systemPackages`.

## Limits

Per call: a timeout (30 s by default, up to 15 min), a body size (6 MiB), and a
number of calls running at once per instance (16). They are fields of
`functionSource`.

## Reaching it

Created with a domain, a function is routed there over HTTPS from its first
deployment. Without one it is reached inside its project, by its service's
name, on port 8080. A scheduled job calls it with a request on a schedule:
`plan_create_sched_job` with `functionInvoke` (method, path, headers, body)
instead of a command; `plan_run_sched_job` runs one now.

## Trying and watching it

- `plan_test_run_function` calls the code once - the function's saved code, or
  files given - with a request, in a throwaway container, with the function's
  variables and secrets. Nothing is saved or deployed. Its answer is the
  response, the log and, when the libraries were installed, their lock files.
- `get_function_metrics` counts its calls, failures, 5xx answers and durations
  over 1h, 6h, 24h or 7d, from the line its runtime logs for every call.
- `search_app_logs` reads its log: every call is a line
  `{"hp":"invocation", "requestId", "status", "durationMs", "outcome"}`, and
  `ctx.log` lines carry the call's `requestId`.
