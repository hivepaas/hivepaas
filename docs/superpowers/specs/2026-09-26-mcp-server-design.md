# An MCP Server for HivePaaS

An AI assistant is useful to a HivePaaS user when it can see what HivePaaS sees:
why an app keeps restarting, what its last hundred log lines say, which template
would install the database it needs, whether a cron expression means what they
think. This design has HivePaaS serve the Model Context Protocol, so that the
assistant a person already uses - Claude Code, Claude Desktop, an editor - can ask
HivePaaS directly, as that person, and nothing more.

It covers the server and its read-only tools. Tools that change things, and an
assistant built into the dashboard, are the phases after it (§10), and are shaped
here so that they need no second tool layer.

---

## Decisions

1. **HivePaaS is the server; the model is the client's.** The assistant, its
   model and its provider key live in the client. HivePaaS holds no AI provider
   key for this, pays for no tokens, and sends nothing to a model provider: the
   client asks, HivePaaS answers.
2. **One tool layer, two doors.** Tools are defined once, in a service. The MCP
   endpoint is the first door; the dashboard's own assistant (§10.2) is the
   second, and calls the same tools in-process.
3. **A tool acts as the person who called it, through the dashboard's own
   endpoints.** It sends the request the dashboard would send to the backend's
   router, in-process, carrying the caller's already-verified `*basedto.Auth`.
   Permissions - many of them checked in the handler, not the use case - the
   key's access actions, validation, audit and secret masking are the ones that
   already apply; a tool can see and do exactly what its caller can in the
   dashboard, and cannot drift from it.
4. **Read-only first.** Phase 1 ships no tool that changes anything. A tool that
   writes comes with a plan and an explicit apply (§10.1), never in one call.
5. **Answers are bounded.** A model reads every byte it is given. Logs come as a
   tail with a limit, lists are paged, and nothing a tool returns is larger than
   `MaxToolOutput` (64 KB), cut with a note saying so.
6. **Secrets never leave as values.** A tool returns a variable's name and whether
   it is a secret or a reference, never a secret's value; `Reveal` is not a tool.
7. **Off until an administrator turns it on.** A global switch enables the
   endpoint; with it off, `/mcp` answers 404.
8. **Authentication is an API key.** The existing keys, with their access actions,
   so a person can give an assistant a key that reads and nothing else. OAuth, for
   clients that only speak it, comes later (§11).

## 1. Protocol and transport

- **Streamable HTTP**, protocol revision `2026-07-28`, at `<API base path>/mcp`
  on the backend's own listener - `https://hivepaas.example.com/api/mcp`. No new
  port, no new service, and the reverse proxy already routes it.
- **Stateless.** Every request carries its key and is served on its own; no MCP
  session is kept between requests. Nothing a phase 1 tool does needs one, and a
  stateless endpoint survives the backend restarting or running more than one
  replica.
- **The official Go SDK**, `github.com/modelcontextprotocol/go-sdk` v1.8.0,
  vendored. It carries the JSON-RPC framing, the protocol negotiation and the
  JSON Schema of each tool's input from its Go type.
- The server declares **tools**, **resources** and **prompts**, and nothing else.
  No sampling, no elicitation: those ask the client for things, and phase 1 has
  nothing to ask.

## 2. Authentication

The endpoint takes an API key the way the rest of the API does, in either form:

- `HIVEPAAS-API-KEY-ID` and `HIVEPAAS-API-SECRET-KEY` headers, as today;
- `Authorization: Bearer <keyId>:<secret>`, because some MCP clients let a
  person set that header and no other.

Both go through `sessionuc.GetCurrentAuthByAPIKey`. The key's `AccessAction`
bounds what its tools may do: a key without `read` can call nothing, and in
phase 2 a key without `write` can plan a change but not apply it.

A dashboard session token is refused here: it expires within minutes and a person
would be pasting it into a config file. The MCP tab (§8) sends them to create a
key instead.

## 3. The tool layer

`interface/mcp` holds a registry of tools:

```go
type Tool struct {
	Name        string      // snake_case, unique: "get_app_logs"
	Title       string      // what a client shows a person
	Description string      // what a model reads to decide to call it
	Kind        ToolKind    // read, write, destructive
	Input       any         // a Go struct; its JSON Schema is the tool's input schema
	Handle      func(ctx context.Context, auth *basedto.Auth, input any) (*ToolResult, error)
}
```

- `Handle` dispatches to the backend's own router, in-process: the method, path,
  query and body the dashboard would send, answered by the same handler. The
  caller's `*basedto.Auth` travels in the request's context, under a key only
  this package can set, and `authhandler` takes it from there instead of from a
  header; a request from outside can carry headers but never a Go context, so it
  cannot claim one. No credential is verified twice, and no handler learns it is
  being called by a tool.
- A tool that needs something no endpoint offers gets an endpoint first, so the
  dashboard can have it too. The tool then trims the endpoint's answer to what a
  model needs.

  *Found while planning:* the first version of this design had tools call use
  cases directly. Many permission checks live in the handlers - task lists,
  settings lists, the per-module checks passed to `GetCurrentAuth` - so a tool
  calling the use case would have had to repeat each of them, and one it missed
  would be a way around it.
- `ToolResult` is text for the model and, where the SDK allows it, structured
  content of the same data. Errors a person can act on - not found, not
  permitted, invalid input - are tool results with `isError`, worded for the
  model; anything else is a protocol error and is logged.
- `Kind` is what the phases, the key's access actions and the global switch
  (§8) check before `Handle` runs.

The same package adapts the registry to the SDK's server and mounts it on the
router, behind the switch and the API key check. The dashboard's assistant (§10.2)
runs in the same process and dispatches the same way.

### Naming things

A person says "the api app in shop's staging env", not an id. Every tool that
takes an app takes `project`, `env` and `app` as keys **or** names, resolved the
way the dashboard's search resolves them and filtered by `Visibility`. An
ambiguous name is an error listing the candidates, never a guess.

## 4. Phase 1 tools

All `read`. Each dispatches to the endpoint whose use case is named.

| Tool | What it answers | The endpoint's use case |
|---|---|---|
| `list_projects` | projects and their envs the caller can see | `projectuc.ListProject` |
| `list_apps` | apps of an env, with status, image, replicas running/desired | `appuc.ListApp` |
| `get_app` | one app: kind, image, status, routing, mounts, last deployments | `appuc.GetApp`, `appdeploymentuc` |
| `get_app_status` | why an app is not running: its tasks' states, exit codes and errors | `appuc.GetApp`, swarm tasks |
| `get_app_logs` | a tail of an app's logs: `tail` (≤ 500), `since`, `grep` | `appuc.GetAppLogs` (`Follow` off) |
| `get_app_config` | an app as a spec document: variables by name, secrets as names | `specuc.ExportSpec`, app scope |
| `list_attention` | what needs attention, as the Home card shows it | `homeuc.GetHomeAttention` |
| `list_tasks` | recent tasks - deployments, backups, jobs - with state | `taskuc.ListTask` |
| `get_task_logs` | a tail of one task's logs | `taskuc.GetTaskLogs` |
| `list_nodes` | nodes, their state and memory, as the cluster screen shows | `cluster` use cases |
| `search_templates` | store templates matching words, with tagline and versions | `apptemplateuc.ListAppTemplates` |
| `get_template` | one template: parameters, dependencies, what it grants | `apptemplateuc.GetAppTemplate` |
| `preflight_install` | what installing a template would create and what stops it | `apptemplateuc.PreflightAppFromTemplate` |
| `list_sched_jobs` | scheduled jobs of a scope, with their next runs | `schedjobuc.ListSchedJob` |
| `explain_schedule` | the next runs of a cron expression, in the job's time zone | `schedjobuc.CalcNextRuns` |

`preflight_install` is read-only even though it is about installing: it creates
nothing, and it is how phase 2's `install_app` is planned.

`get_app_config` exports with `SecretsMode` fixed at `omit`, the export's own
default, and takes no parameter that could change it: the encrypted and plaintext
modes are reveals, gated and audited as such, and no tool offers one.

`get_app_logs` greps on the server, before the tail: a model asking for "errors
in the last hour" should get the errors, not the last 500 lines that happen to
hold none.

## 5. Resources and prompts

**Resources** are documents a client can attach to a conversation:

- `hivepaas://templates/{name}` - a template's description, as the store shows it;
- `hivepaas://docs/docker-api` - what the Docker API lets an app do, from the
  guide the dashboard shows.

**Prompts** are starting points a client offers a person:

- `debug_app(project, env, app)` - gathers status, the last errors in the logs
  and the last deployment, and asks for a diagnosis;
- `install_app(what)` - searches the store, compares candidates, and preflights
  the chosen one.

A prompt only arranges tool calls and text; it grants nothing a tool does not.

## 6. What a tool returns

- **Bounded:** `MaxToolOutput` per result, lists paged by `limit` and `cursor`,
  logs by `tail`. A cut result says it was cut and how to ask for the rest.
- **No secret values:** variables by name, with `secret: true` or the reference
  they hold (`${postgres.HIVEPAAS_PASSWORD}`); the `MaskSecrets` path the env
  screen uses.
- **Ids alongside names**, so the next call can be exact.
- **Times in RFC 3339 UTC**, sizes in bytes and in words.

## 7. Audit

Every tool call is an audit entry with source `mcp`, the tool's name and its
input, less anything marked secret. A read that the dashboard would not audit is
still audited here: an assistant reading logs is worth being able to see, and
the entries are what §8 shows. A call refused by permission is audited as
refused.

## 8. Settings and the dashboard

A new global singleton setting, `mcp`:

```json
{ "enabled": false, "allowWrite": false }
```

- `enabled` switches the endpoint; only an administrator changes it.
- `allowWrite` is read in phase 2: with it off, write tools are not listed at all.

**System settings › AI**, a new page, with one tab in phase 1:

- **MCP server** - the switch, the endpoint URL, a button to create a read-only
  API key, and ready-to-copy client configuration: `claude mcp add` for Claude
  Code, the JSON for Claude Desktop and editors, each with the key filled in the
  moment it is created and never shown again.
- The last 20 MCP calls, from the audit log, so a person can see what their
  assistant has been doing.

The **AI provider** tab joins it in phase 3 (§10.2).

## 9. Testing

- **Registry:** every tool's input has a schema, every tool's `Kind` is set, names
  are unique, and a tool the key's access actions do not allow is refused before
  `Handle` runs.
- **Each tool, table-driven,** against use case fakes: its answer for a visible
  app, for one the caller cannot see (not found, as the API answers), for an
  ambiguous name, and the output bound.
- **Secrets:** a golden test that no tool's output contains a value stored as
  secret, run over a seeded project.
- **The endpoint,** with the SDK's own client over HTTP: initialize, list tools,
  call one; refused with the switch off, without a key, with a key lacking `read`.
- **By hand:** Claude Code connected to the local stack, debugging an app that
  crash-loops and installing nothing.

## 10. The phases after this one

### 10.1 Phase 2: tools that change things

Each write is two tools: one that plans and one that applies what was planned.

- `plan_*` returns what would change - a spec diff, a preflight - and a
  `planToken`, kept 10 minutes in redis, bound to the caller and to the exact
  change.
- `apply_plan(planToken)` applies it, if nothing it depends on has changed since
  (the settings' `updateVer`), and returns the task to follow.

A model cannot skip the plan, and a client shows the person the plan before it
calls `apply_plan`. The first write tools: `install_app` (from a template),
`restart_app`, `redeploy_app`, `update_app_config` (a spec change, validated and
diffed by `specuc`'s import), and `create_sched_job`. `destructive` tools -
deleting an app, a volume, a project - are not in phase 2.

### 10.2 Phase 3: the dashboard's own assistant

- A setting `ai-provider`, global, its key an `EncryptedField`: provider
  (`anthropic`, or `openai-compatible` with a base URL, which covers OpenAI,
  gateways and Ollama), model, a monthly token budget, and whether logs may be
  sent to the provider at all.
- A chat panel in the dashboard, streaming over SSE from an agent loop in the
  backend that calls the same registry in-process, as the signed-in person.
- The Home card's **Needs attention** items gain an "Ask the assistant" action
  that opens the panel with the item as context.

## 11. Not in this design

- **OAuth.** Claude Desktop's and claude.ai's remote connectors authenticate with
  OAuth, per the MCP authorization spec; HivePaaS would be the authorization
  server. It is its own design, and API keys serve Claude Code and editors today.
- **A stdio server** (`hivepaas mcp` run locally). The HTTP endpoint serves every
  client that matters, and a stdio binary would be one more thing to ship per
  platform.
- **Tools across installations.** One endpoint is one HivePaaS.
- **Following logs.** A model cannot read a stream; `get_app_logs` answers once.
