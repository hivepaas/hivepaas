# MCP Server, Phase 1 - Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** HivePaaS serves the Model Context Protocol at `<API base path>/mcp`: fifteen read-only tools, two resources and two prompts, answered as the person whose API key calls them, through the dashboard's own endpoints; off until an administrator turns it on from a new System settings › AI page.

**Architecture:** `interface/mcp` holds the tool registry, the endpoint and an in-process dispatcher. The endpoint is a gin route that checks the `mcp` setting, authenticates an API key, and hands the request to the SDK's stateless Streamable HTTP handler with the caller's `*basedto.Auth` in its context. A tool builds the request the dashboard would send and dispatches it to the backend's own gin engine; `authhandler` finds the auth in the context, so the handler runs its usual checks on the caller. The tool trims the answer, bounds it, and records one audit entry.

**Tech Stack:** Go 1.27, gin, `github.com/modelcontextprotocol/go-sdk` v1.8.0 (vendored; protocol `2026-07-28`), uber/fx wiring in `registry/provides.go`; the dashboard in React, TanStack Query, zod.

**Spec:** `docs/superpowers/specs/2026-09-26-mcp-server-design.md` - amended in the same commit as this plan: tools dispatch through the router instead of calling use cases (Decision 3, §3), found while planning (see "Found while planning").

## Global Constraints

- Phase 1 has no tool that writes. Every tool is `read`; a `write` or `destructive` tool in this phase is a review failure.
- A tool never passes a value from its input into a path unescaped: path segments go through `url.PathEscape`, queries through `url.Values`.
- Nothing a tool returns is larger than `MaxToolOutput = 64 << 10` bytes; a cut answer ends with the note `truncated: <n> bytes left out; <how to ask for the rest>`.
- No secret value in any tool's output: `get_app_config` exports with `secretsMode=omit` fixed; env variables come from the list endpoint, which masks them.
- The endpoint answers `404` while the setting is off, `401` without a valid API key, and refuses a dashboard session token (`Authorization: Bearer <jwt>`) with `401` and a message that says to use an API key.
- `/mcp` never dispatches to itself: the dispatcher refuses any path under `<base>/mcp`.
- `golangci-lint run ./...` clean (120 columns, US spelling), `go test ./...` green, `make gen-swag` after DTO changes; the dashboard's `tsc`, eslint and prettier clean.
- The local Docker Desktop runs the user's own HivePaaS: no task deploys to it. Manual checks use `make local-app-run` against the local stack only after the user says the backend may be restarted.
- Work on branch `feat/mcp-server` in `hivepaas` and in `hivepaas-dashboard`; merge into `main` locally with `--no-ff`; never push. Every commit ends with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

## Review Focus

- **The dispatched auth.** `authhandler` must take the auth from the context only when the dispatcher put it there (unexported key type), must still run `VerifyAuth` with the handler's access check, and a request from outside must have no way to set it. Pinned by `TestDispatchedAuthIsOnlyTakenFromTheContext` and `TestDispatchedAuthStillVerifiesAccess` (Task 2).
- **A fresh auth per dispatched request.** `permission.CheckAccess` writes the resources it allowed into `auth.AllowedResources`, and use cases filter lists by it; a tool that dispatches twice with one auth would have the second answer filtered by the first's check. `authhandler` hands each handler a copy holding only the user. Pinned by `TestEachDispatchedRequestGetsAFreshAuth` (Task 2), found while implementing Task 2.
- **An API key's access actions.** A key limited to `read` reaches read endpoints and nothing else, through a tool as through the API. Pinned by `TestToolsRespectTheKeysAccessActions` (Task 4).
- **Secrets.** `get_app_config` and `get_app` must not carry a secret value, whatever the app holds. Pinned by `TestNoToolOutputCarriesASecretValue` (Task 5), over a fake app whose secret and env values are distinctive strings.
- **Name resolution.** An ambiguous project, env or app name is an error listing the candidates; a name the caller cannot see is "not found", the same as a name that does not exist. Pinned by `TestResolveRefusesAmbiguity` and `TestResolveHidesWhatTheCallerCannotSee` (Task 5).
- **Bounds.** Logs by `tail` ≤ 500 and by grep before tail; every answer cut at `MaxToolOutput`. Pinned by `TestLogsGrepBeforeTail` and `TestAnswersAreBounded` (Tasks 3, 5).

## Before you start

- **Repositories:** `/Users/tiendc/go/src/github.com/hivepaas/hivepaas` and `/Users/tiendc/go/src/github.com/hivepaas/hivepaas-dashboard`, both on `main`. Go commands run from the first.
- **Read** `docs/ARCHITECTURE.md` §2 (request and response shapes) before Task 1, and the spec.
- **What was verified while planning:** the SDK's stateless `StreamableHTTPHandler` passes the HTTP request's context to tool handlers, so a value an outer handler puts there reaches the tool; `mcp.AddTool[In, Out]` produces the input schema from the struct, `jsonschema` tags becoming descriptions and non-`omitempty` fields becoming `required`; the SDK's own client, `StreamableClientTransport`, calls the tools over `httptest`. The spike is not committed.
- **Found while planning, and handled here:**
  - Permission checks live in handlers as often as in use cases (`GetAuthGlobalTasks`, the settings handlers' `ListSetting`, the access checks passed to `GetCurrentAuth`), so tools dispatch through the router (spec amended).
  - `POST .../apps/:appID/spec/export` answers `application/gzip`; `get_app_config` unpacks it.
  - `GET .../apps/:appID/logs` answers JSON only without the websocket upgrade, with `follow` forced off - which is what a tool sends.
  - The API key create DTO already takes `accessAction`, so the dashboard's "read-only key" needs no backend change.

## Files

| file | task | what |
|---|---|---|
| `vendor/`, `go.mod`, `go.sum` | 1 | the MCP Go SDK |
| `hivepaas_app/base/setting.go` | 1 | `SettingTypeMCP` |
| `hivepaas_app/entity/setting_mcp.go` | 1 | `MCPSettings` |
| `hivepaas_app/usecase/systemsettings/mcpuc/` | 1 | get and update, as `logginguc` |
| `hivepaas_app/interface/api/handler/systemsettingshandler/mcp.go` | 1 | `GET`/`PUT /system/settings/mcp` |
| `hivepaas_app/interface/api/server/router_system.go` | 1 | the two routes |
| `hivepaas_app/interface/api/handler/authhandler/dispatch.go` | 2 | the dispatched auth and the API-key-only auth |
| `hivepaas_app/base/audit.go` | 3 | `AuditLogTypeMCPToolCall`, `AuditLogSourceMCP` |
| `hivepaas_app/interface/mcp/dispatch.go` | 3 | the in-process dispatcher |
| `hivepaas_app/interface/mcp/tool.go` | 3 | the registry, `Call`, bounds, audit |
| `hivepaas_app/interface/mcp/endpoint.go` | 4 | the route, the switch, the key, the SDK handler |
| `hivepaas_app/interface/api/server/router_mcp.go` | 4 | mounting it |
| `hivepaas_app/interface/mcp/resolve.go` | 5 | project, env and app by key or name |
| `hivepaas_app/interface/mcp/tools_apps.go` | 5 | the app tools |
| `hivepaas_app/interface/mcp/tools_cluster.go` | 6 | attention, tasks, nodes |
| `hivepaas_app/interface/mcp/tools_store.go` | 6 | templates, preflight, schedules |
| `hivepaas_app/interface/mcp/resources.go` | 6 | resources and prompts |
| `hivepaas-dashboard/src/application/modules/system-settings/routes/ai/` | 7 | the AI page, MCP tab |

---

### Task 1: The SDK, the `mcp` setting and its endpoints

**Files:** `go.mod`, `go.sum`, `vendor/`; `base/setting.go`; `entity/setting_mcp.go` (+ `_test.go`); `usecase/systemsettings/mcpuc/{uc.go,settings_get.go,settings_update.go,mcpdto/*.go}`; `systemsettingshandler/mcp.go`; `router_system.go`; `registry/provides.go`.

- [ ] **Step 1: Vendor the SDK.**

```bash
go get github.com/modelcontextprotocol/go-sdk@v1.8.0
go mod tidy && go mod vendor
go build ./...
```

Expected: builds. The SDK brings `google/jsonschema-go`, `yosida95/uritemplate/v3`, `segmentio/encoding`, `golang.org/x/{oauth2,time}`; check each license is MIT, BSD or Apache-2.0 before committing (`go-licenses` is not required; read each `LICENSE`).

- [ ] **Step 2: The setting type and its data.** Add `SettingTypeMCP SettingType = "mcp"` to `base/setting.go`, in order. `entity/setting_mcp.go`, modelled on `setting_logging.go`:

```go
// MCPSettings is whether HivePaaS serves the Model Context Protocol, and what
// its tools may do. See docs/superpowers/specs/2026-09-26-mcp-server-design.md.
type MCPSettings struct {
	// Enabled serves <API base path>/mcp; off, the path answers 404.
	Enabled bool `json:"enabled"`
	// AllowWrite lists the tools that change things. Read from phase 2 on.
	AllowWrite bool `json:"allowWrite"`
}
```

with its parser registration, `GetType`, `GetRefObjectIDs` (empty), `AsMCPSettings`, and `CurrentMCPSettingsVersion = 1`. Test: a round trip through `SetData`/`AsMCPSettings`, and that an absent setting reads as `Enabled: false`.

- [ ] **Step 3: `mcpuc`.** Copy the shape of `usecase/systemsettings/logginguc`: `GetMCPSettings` returns the setting or the zero value when none exists; `UpdateMCPSettings` upserts the global singleton with `UpdateVer` checked, and records an audit entry the way `logginguc` does. Add `IsEnabled(ctx) (bool, error)` for the endpoint, reading through a 10-second cache so that every MCP request is not a query; an update clears the cache.

- [ ] **Step 4: Endpoints.** In `systemsettingshandler/mcp.go`, `GetMCPSettings` and `UpdateMCPSettings`, with the same access checks and swag annotations as the logging pair (`GetLoggingSettings`, `UpdateLoggingSettings`) - read for the settings module to get, write to update. Routes in `router_system.go`, beside logging:

```go
	// MCP settings
	{
		mcpGroup := systemSettingGroup.Group("/mcp")
		mcpGroup.GET("", systemSettingsHandler.GetMCPSettings)
		mcpGroup.PUT("", systemSettingsHandler.UpdateMCPSettings)
	}
```

Wire `mcpuc.New` in `registry/provides.go` and inject it into `systemsettingshandler.New`.

- [ ] **Step 5: Test and commit.**

```bash
go build ./... && go test ./hivepaas_app/entity/ ./hivepaas_app/usecase/systemsettings/mcpuc/...
make gen-swag && golangci-lint run ./...
git add -A && git commit -m "feat(mcp): the MCP Go SDK, and the setting that turns the server on"
```

---

### Task 2: The dispatched auth, and the API-key-only auth

**Files:** `interface/api/handler/authhandler/dispatch.go` (new), `handler.go`, `dispatch_test.go`.

- [ ] **Step 1: Write the failing tests** in `dispatch_test.go`:

  - `TestDispatchedAuthIsOnlyTakenFromTheContext`: a gin context whose request context carries `WithDispatchedAuth(ctx, auth)` gets that auth from `GetCurrentAuth`; the same request with no such context and no header gets `ErrNoSession`; a header named anything at all does not stand for it.
  - `TestDispatchedAuthStillVerifiesAccess`: with a dispatched auth whose `AuthClaims.AccessAction` is read-only, `GetCurrentAuth` with a write `ModuleAccessCheck` fails `ErrForbidden`, as it does for the same key over HTTP.
  - `TestGetAPIKeyAuthTakesBothForms`: the two `HIVEPAAS-API-*` headers, and `Authorization: Bearer <keyId>:<secret>`, reach `GetCurrentAuthByAPIKey` with the same id and secret (a fake `sessionUC`); `Bearer <jwt>` without a colon is refused with `ErrSessionAPIKeyInvalid` and never reaches `GetCurrentAuthByJWT`.

- [ ] **Step 2: Implement.**

```go
// Package-private: only a value this package puts in a context carries an
// auth, and a request from outside the process can set headers but never its
// context.
type dispatchedAuthKey struct{}

// WithDispatchedAuth is the context of a request the backend sends to its own
// router on behalf of a caller it has already authenticated - an MCP tool. The
// handler takes the caller from it, then checks access as it would for a
// request from outside.
func WithDispatchedAuth(ctx context.Context, auth *basedto.Auth) context.Context {
	return context.WithValue(ctx, dispatchedAuthKey{}, auth)
}

func dispatchedAuth(ctx context.Context) *basedto.Auth {
	auth, _ := ctx.Value(dispatchedAuthKey{}).(*basedto.Auth)
	return auth
}
```

At the top of `getCurrentAuth`:

```go
	if auth := dispatchedAuth(ctx.Request.Context()); auth != nil {
		return auth, nil
	}
```

and a new method for the MCP endpoint:

```go
// GetAPIKeyAuth authenticates a request by API key alone: the two
// HIVEPAAS-API-* headers, or "Authorization: Bearer <keyId>:<secret>" for
// clients that let a person set that header and no other. A session token is
// refused: it expires within minutes and would end up pasted in a config file.
func (h *Handler) GetAPIKeyAuth(ctx *gin.Context) (*basedto.Auth, error)
```

It splits the bearer value at the first `:`; no colon, an empty half, or a value that parses as a JWT is `ErrSessionAPIKeyInvalid`.

- [ ] **Step 3: Test and commit.**

```bash
go test ./hivepaas_app/interface/api/handler/authhandler/ && golangci-lint run ./...
git commit -am "feat(auth): a request the backend dispatches to itself carries its caller in its context"
```

---

### Task 3: The dispatcher, the registry, bounds and audit

**Files:** `base/audit.go`; `interface/mcp/{dispatch.go,tool.go,bound.go}` and their tests.

- [ ] **Step 1: Audit constants.** In `base/audit.go`: `AuditLogTypeMCPToolCall AuditLogType = "mcp-tool-call"` and `AuditLogSourceMCP AuditLogSource = "mcp"`, each with a comment, added to the lists the audit type endpoint returns.

- [ ] **Step 2: The dispatcher.**

```go
// Dispatcher sends a request to the backend's own router, as the caller: the
// request the dashboard would send, answered by the same handler, with the same
// checks. It never goes through the network.
type Dispatcher struct {
	handler  http.Handler // the gin engine
	basePath string       // the API's, e.g. "/api"
}

// Response is what the handler answered.
type Response struct {
	Status int
	Header http.Header
	Body   []byte
}

// Do dispatches one request. path is below the API's base path, its segments
// already escaped; query and body may be nil. A path under /mcp is refused.
func (d *Dispatcher) Do(ctx context.Context, auth *basedto.Auth, method, path string,
	query url.Values, body any) (*Response, error)
```

It builds the request with `http.NewRequestWithContext(authhandler.WithDispatchedAuth(ctx, auth), ...)`, JSON-encodes a non-nil body with `Content-Type: application/json`, copies `RemoteAddr` and `User-Agent` from the MCP request (kept in the context by the endpoint, Task 4) so audit entries show where the call came from, and records the answer in a small buffered `http.ResponseWriter` of its own - not `httptest.ResponseRecorder` in production code. A non-2xx answer becomes an `*APIError{Status, Code, Message}` decoded from the API's error body, so a tool can word it.

Tests (`dispatch_test.go`, over a gin engine with two test routes): the handler sees the caller through `authhandler.GetCurrentAuth`; the query and JSON body arrive; `/mcp/x` is refused without being served; a 404 and a 403 come back as `*APIError` with the API's message.

- [ ] **Step 3: The registry and `Call`.**

```go
// Kind is what a tool may do. Phase 1 registers read tools only.
type Kind string

const (
	KindRead        Kind = "read"
	KindWrite       Kind = "write"
	KindDestructive Kind = "destructive"
)

// Tool is one tool, as registered with the SDK's server.
type Tool struct {
	Name, Title, Description string
	Kind                     Kind
	add                      func(s *mcpsdk.Server, deps *Deps)
}

// Call is what a tool's run function works with: the caller and the
// dispatcher, and helpers that decode an answer.
type Call struct {
	Auth *basedto.Auth
	deps *Deps
}

func (c *Call) Get(ctx context.Context, path string, query url.Values, out any) error
func (c *Call) Post(ctx context.Context, path string, body, out any) error

// readTool registers a read tool whose input is In and whose answer is Out.
// It audits the call, bounds the answer, and turns an *APIError into a tool
// error the model can read.
func readTool[In, Out any](name, title, description string,
	run func(ctx context.Context, call *Call, in In) (Out, error)) Tool
```

`readTool`'s handler: take the auth from the context (put there by the endpoint; missing is a protocol error, never a tool result), record the audit entry through `auditservice.RecordAllowed` with type `mcp-tool-call`, source `mcp`, detail `{"tool": name, "input": in}` - and abandon the call if the record fails, as every recorded action does - then run, then bound. An `*APIError` becomes `CallToolResult{IsError: true}` with text such as `not found: no app "api" in env "staging" of project "shop" that you can see`. A refused access check is audited again with result `refused`.

- [ ] **Step 4: Bounds.** `bound.go`: `MaxToolOutput = 64 << 10`; `boundText(s string, hint string) string` cuts at a UTF-8 boundary and appends the note; `boundJSON(v any, hint string)` marshals, and if too large, lists shorten from the end until the answer fits, with a `truncated` field saying how many items were left out. Tests: `TestAnswersAreBounded` over a 1 MB log and a 5,000-item list; a multi-byte rune on the boundary is never split.

- [ ] **Step 5: Test and commit.**

```bash
go test ./hivepaas_app/interface/mcp/... && golangci-lint run ./...
git add -A && git commit -m "feat(mcp): a dispatcher to the backend's own router, and the tool registry"
```

---

### Task 4: The endpoint

**Files:** `interface/mcp/endpoint.go` (+ `_test.go`); `interface/api/server/router_mcp.go`; `router.go`; `registry/provides.go`.

- [ ] **Step 1: Write the failing tests** (`endpoint_test.go`), each with the SDK's client over `httptest`:
  - `TestEndpointIsOffUntilEnabled`: setting off, `POST /api/mcp` answers 404 and the SDK client's `Connect` fails.
  - `TestEndpointTakesAnAPIKeyOnly`: no key 401; a session token 401 with the API-key message; a valid key lists the fifteen tools.
  - `TestToolsRespectTheKeysAccessActions`: a key with `accessAction: {read: false}` gets a tool error on every tool; a read-only key reads.
  - `TestEveryToolIsReadAndDescribed`: every registered tool has `Kind == KindRead`, a title, a description of at least one sentence, and an input schema that is an object.

- [ ] **Step 2: Implement.**

```go
// Endpoint serves MCP at <base>/mcp: stateless Streamable HTTP, one server
// shared by every request, the caller in each request's context.
type Endpoint struct {
	server  *mcpsdk.Server
	handler *mcpsdk.StreamableHTTPHandler
	auth    *authhandler.Handler
	mcpUC   *mcpuc.UC
}

func NewEndpoint(deps *Deps, tools []Tool) *Endpoint {
	server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "hivepaas", Version: version.Current()}, nil)
	for _, tool := range tools {
		tool.add(server, deps)
	}
	addResources(server, deps)
	addPrompts(server)
	h := mcpsdk.NewStreamableHTTPHandler(func(*http.Request) *mcpsdk.Server { return server },
		&mcpsdk.StreamableHTTPOptions{Stateless: true})
	return &Endpoint{server: server, handler: h, auth: deps.AuthHandler, mcpUC: deps.MCPUC}
}

// Serve is the gin handler.
func (e *Endpoint) Serve(ctx *gin.Context) {
	enabled, err := e.mcpUC.IsEnabled(ctx)
	if err != nil || !enabled {
		ctx.String(http.StatusNotFound, "not found")
		return
	}
	auth, err := e.auth.GetAPIKeyAuth(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "an API key is required: " + err.Error()})
		return
	}
	req := ctx.Request.WithContext(withCaller(ctx.Request.Context(), auth, ctx.Request))
	e.handler.ServeHTTP(ctx.Writer, req)
}
```

`router_mcp.go` mounts `apiGroup.Any("/mcp", endpoint.Serve)`; `registerRoutes` calls it last. The endpoint is built in the server, which alone holds the gin engine the dispatcher needs: `Deps{Dispatcher: &Dispatcher{handler: s.engine, basePath: ...}, AuthHandler, MCPUC, AuditService}` - the last three come through the handler registry, wired in `provides.go`.

- [ ] **Step 3: Test and commit.**

```bash
go test ./hivepaas_app/interface/mcp/... ./hivepaas_app/interface/api/... && golangci-lint run ./...
git add -A && git commit -m "feat(mcp): the endpoint, off by default, by API key"
```

---

### Task 5: Names, and the app tools

**Files:** `interface/mcp/resolve.go`, `tools_apps.go`, and their tests.

- [ ] **Step 1: Resolution.** `resolveApp(ctx, call, project, env, app string) (*appRef, error)`:
  1. `GET /projects?search=<project>`; a project matches when its key equals the input, or its name does case-insensitively. None: `not found`. Several: an error listing `key (name)` for each.
  2. The env from the project's envs, by name or key the same way.
  3. `GET /projects/<id>/<env>/apps?search=<app>&getChildApps=true`, matched the same way.

  Every lookup goes through the dispatcher, so what the caller cannot see is not found. Tests: `TestResolveRefusesAmbiguity`, `TestResolveHidesWhatTheCallerCannotSee` (the fake handler answers the search as the permission layer would), `TestResolveTakesKeysAndNames`.

- [ ] **Step 2: The app tools.** Each input names `project`, `env`, `app` (`jsonschema:"key or name"`), resolved first.

| tool | dispatches | trims to |
|---|---|---|
| `list_projects` | `GET /projects` then each project's envs | key, name, envs |
| `list_apps` | `GET /projects/:id/:env/apps?getStats=true` | key, name, kind, image, status, replicas running/desired |
| `get_app` | `GET .../apps/:appID`, `GET .../apps/:appID/deployments?pageSize=5` | kind, image, status, routing domains, mounts, the last five deployments with state |
| `get_app_status` | `GET .../apps/:appID/service-tasks` | each task's state, desired state, exit code, error and time, newest first, at most 20 |
| `get_app_logs` | `GET .../apps/:appID/logs?tail=<n>&since=<t>&timestamps=true` | lines, grep applied before the tail |
| `get_app_config` | `POST .../apps/:appID/spec/export` with `secretsMode=omit` | the app's YAML document from the archive |

  `get_app_logs` input: `tail` (default 100, at most 500), `since` (RFC 3339 or a duration such as `1h`), `grep` (a case-insensitive substring; a regular expression when wrapped in `/.../`). With `grep`, it asks the endpoint for `tail = 5000` over the same `since`, filters, then keeps the last `tail`; the answer says how many lines matched. `get_app_config` unpacks the `application/gzip` archive with `archive/tar` and `compress/gzip`, and returns the app's document only.

- [ ] **Step 3: Tests.** Over a fake router with canned answers: each tool's trimmed answer; `TestLogsGrepBeforeTail`; `TestNoToolOutputCarriesASecretValue` - an app whose env values and secrets are `s3cr3t-<n>` and whose export archive (built by the real `specuc` in the test) is inspected: no tool's output contains `s3cr3t`.

- [ ] **Step 4: Commit.** `git commit -m "feat(mcp): tools that read apps - status, logs, config"`

---

### Task 6: Attention, tasks, nodes, the store, schedules, resources, prompts

**Files:** `tools_cluster.go`, `tools_store.go`, `resources.go`, and tests.

- [ ] **Step 1: The tools.**

| tool | dispatches | notes |
|---|---|---|
| `list_attention` | `GET /home/attention` | as the Home card: kind, subject, scope, since, last error |
| `list_tasks` | `GET /system/tasks?status=&type=&pageSize=` | newest first, at most 50 |
| `get_task_logs` | `GET /system/tasks/:id/logs?tail=` | bounded as app logs |
| `list_nodes` | `GET /cluster/nodes` | hostname, role, state, availability, CPUs, memory |
| `search_templates` | `GET /app-templates?search=` | name, title, tagline, categories, versions, whether it needs the Docker API |
| `get_template` | `GET /app-templates/:name` | parameters (secrets as "generated or asked"), dependencies, components, what each grants |
| `preflight_install` | `POST /projects/:id/:env/apps/from-template/preflight` | the same body the deploy dialog sends; the answer's issues and planned apps |
| `list_sched_jobs` | the app's `GET .../sched-jobs`, or the global `GET /settings/sched-jobs` | name, schedule, next runs, last result |
| `explain_schedule` | `POST /settings/sched-jobs/calc-next-runs` | the next runs, in the given time zone |

  Every path above was read from `router_*.go` while planning.

- [ ] **Step 2: Resources and prompts.** `hivepaas://templates/{name}` (a resource template) answers the template's description; `hivepaas://docs/docker-api` answers the permissions guide's text, kept in Go beside the dashboard's copy. Prompts `debug_app` and `install_app` return the instructions in the spec §5 as a user message.

- [ ] **Step 3: Test and commit.** `git commit -m "feat(mcp): tools for attention, tasks, nodes, the store and schedules"`

---

### Task 7: The dashboard - System settings › AI

**Files (dashboard):** `modules/system-settings/routes/ai/` (route, `mcp` tab, schema, form), the module's `data` (queries and commands for `/system/settings/mcp`), the settings layout's tabs, `ROUTE` constants.

- [ ] **Step 1: Data.** `MCPSettingsQueries.useFindOne` and `MCPSettingsCommands.useUpdate` over `GET`/`PUT /system/settings/mcp`, with the zod schema `{enabled: boolean, allowWrite: boolean, updateVer: number}`; the layout of `routes/logging`.

- [ ] **Step 2: The MCP tab.**
  - The switch, and the endpoint URL: the dashboard's origin plus the API base path plus `/mcp`, with a copy button.
  - **Create a read-only key**: `POST /users/current/settings/api-keys` with `name: "MCP - <date>"` and `accessAction: {read: true}`; the secret is shown once, and the three snippets below are filled with it.
  - Snippets, each with a copy button:
    - Claude Code: `claude mcp add --transport http hivepaas <url> --header "HIVEPAAS-API-KEY-ID: <id>" --header "HIVEPAAS-API-SECRET-KEY: <secret>"`
    - Claude Desktop and editors: the `mcpServers` JSON with `type: "http"`, `url` and the two headers.
    - A client that sets only `Authorization`: `Bearer <id>:<secret>`.
  - **Recent calls**: `GET /system/audit-logs?type=mcp-tool-call&pageSize=20`, showing the time, the user, the tool and whether it was refused.

- [ ] **Step 3: Check and commit.** `npx tsc --noEmit -p .`, eslint and prettier on the changed files; `git commit -m "feat(ai): System settings › AI, with the MCP server's tab"`

---

### Task 8: End to end, and merge

- [ ] **Step 1: An SDK client against a real server.** `interface/mcp/real_server_test.go` calls every tool once through `StreamableClientTransport` against a running HivePaaS, and is skipped unless `HP_TEST_MCP_URL`, `HP_TEST_MCP_KEY` and `HP_TEST_MCP_APP` are set - the pattern of the existing `real_*_test.go` files, which run against a real Docker daemon only when asked. (Planned as a build-tagged test that starts the whole server over a test database; the repository has no such harness, and building one is more than this phase needs.)

- [ ] **Step 2: By hand, with the user's go-ahead to restart the backend.** Turn the switch on in the dashboard, create the key, add the server to Claude Code with the snippet, and ask it: "why is <an app that crash-loops> not running?", "show me errors in <app>'s logs in the last hour", "what would installing postgres into <env> create?". Record what it called in the plan's last section.

- [ ] **Step 3: Merge.**

```bash
go build ./... && golangci-lint run ./... && go test ./...
git checkout main && git merge --no-ff feat/mcp-server
```

and the same for the dashboard.
