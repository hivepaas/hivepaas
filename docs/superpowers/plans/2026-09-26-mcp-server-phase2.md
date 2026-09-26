# The MCP Server, Phase 2 - Implementation Plan

**Spec:** `docs/superpowers/specs/2026-09-26-mcp-server-phase2-design.md`.
**Branches:** `feat/mcp-write` in hivepaas (from `feat/mcp-server`), `feat/mcp-write` in the dashboard (from
`feat/mcp-settings`). Merged `--no-ff` after phase 1, never pushed.

## Review focus

- **Nothing changes before apply.** Every `plan_*` test asserts the fake router saw no request but reads.
- **Apply sends exactly the plan.** The stored request is what is sent; the tool's input is not read again.
- **The three switches.** `TestWritesNeedAllThreeSwitches`: `allowWrite` off, or a key without write and
  execute, lists no write tool and refuses an earlier token.
- **A plan is bound and unreadable at rest.** `TestAPlanIsBoundToItsCaller`, `TestAPlanIsUnreadableAtRest`,
  `TestAPlanIsUsedOnce`.
- **Secrets.** An install's parameter values are in neither the audit entry nor the stored plan in clear.

## Files

| file | task | what |
|---|---|---|
| `usecase/systemsettings/mcpuc/enabled.go` | 1 | `Current(ctx)`: enabled and allowWrite, one cache |
| `interface/mcp/caller.go`, `endpoint.go` | 1 | the caller's key ID; two servers, one chosen per request |
| `repository/cacherepository/mcp_plan_repo.go` | 2 | `Set`, `GetDel` of sealed bytes |
| `interface/mcp/plans.go` | 2 | tokens, sealing, the stored plan |
| `interface/mcp/tool_write.go` | 3 | `planTool`, `apply_plan` |
| `interface/mcp/tools_app_actions.go` | 4 | `plan_restart_app`, `plan_redeploy_app` |
| `interface/mcp/tools_install.go` | 5 | `plan_install_app` |
| `interface/mcp/tools_app_config_write.go` | 6 | `plan_update_app_config` |
| `interface/mcp/tools_sched_write.go` | 7 | `plan_create_sched_job` |
| `interface/mcp/resources.go`, `endpoint.go` | 8 | prompts and instructions |
| dashboard `routes/ai/` | 9 | Allow changes; a pasted key fills the snippets |

---

### Task 1: The switch, and two servers

- `mcpuc`: replace the cached bool with the cached `entity.MCPSettings`; `IsEnabled` stays, `Current(ctx)`
  answers both fields. An update refreshes the cache.
- `caller` gains `keyID`: `authhandler.GetAPIKeyAuth` already parses it; the endpoint reads it the same way
  (a small exported `APIKeyID(ctx *gin.Context) string` beside it) and `withCaller` stores it.
- `NewEndpoint` builds a read server from the read tools and a full server from every tool.
  `StreamableHTTPHandler`'s `getServer` is called per request: full when `Current().AllowWrite` and the key's
  `AccessAction` has `Write` or `Exec` (a nil `AccessAction` - a key with no limit - counts as both).
- `Kind` gains `KindPlan` for `plan_*` and `KindApply` for `apply_plan`; `TestEveryToolIsReadAndDescribed`
  becomes `TestEveryToolIsDescribed`, and checks read tools are all in the read server.
- Test `TestWritesNeedAllThreeSwitches` (listing half; the apply half in Task 3).

### Task 2: The plan store

- `cacherepository.MCPPlanRepo`: `Set(ctx, id string, sealed []byte, ttl)`, `GetDel(ctx, id) ([]byte, error)`
  (nil, nil when absent), keys `mcp:plan:<id>`. Wired in `provides.go`, into `mcp.Services`.
- `plans.go`:

```go
type storedPlan struct {
	Tool    string          `json:"tool"`
	UserID  string          `json:"userId"`
	KeyID   string          `json:"keyId"`
	Method  string          `json:"method"`
	Path    string          `json:"path"`
	Body    json.RawMessage `json:"body,omitempty"`
	Summary string          `json:"summary"`
	// Check is what apply compares before sending: the tool's own, as JSON.
	Check json.RawMessage `json:"check,omitempty"`
}
```

  `savePlan(ctx, caller, plan) (token string, err)`: ULID, 32 random bytes, AES-256-GCM with the ULID as
  additional data, `Set` with `planTTL = 10 * time.Minute`, token `mcpp_<id>.<base64url key>`.
  `takePlan(ctx, caller, token) (*storedPlan, error)`: parse, `GetDel`, open, compare user and key; every
  failure is the same `InputError`: "no such plan: plans last ten minutes and are used once; plan again".
- Tests over an in-memory fake repo: round trip; other user, other key, used, expired (absent), tampered
  token, tampered ciphertext; the sealed bytes contain no field of the body.

### Task 3: `planTool` and `apply_plan`

- `planTool[In, Out](name, title, description, kind, run func(ctx, call, in) (Out, *storedPlan, error))`:
  audited like `readTool` (with `forAudit`), annotated `ReadOnlyHint: true`. When `run` answers a plan, it is
  saved and the answer gains `planToken`, `expiresAt` and `next: "show the person this plan; call apply_plan
  with planToken only once they agree"`. Out embeds `planAnswer{PlanToken, ExpiresAt, Next}` - a struct field
  named `plan`, since the schema generator does not flatten embedding.
- `apply_plan(planToken)`: `DestructiveHint: true`, `IdempotentHint: false`. Checks `Current().AllowWrite`,
  takes the plan, runs the tool's re-check (a registry `checks map[string]func(ctx, call, check) error`),
  dispatches `Method Path Body`, and answers `{tool, summary, result, follow}` where `result` is the
  endpoint's `data` and `follow` names the read tools to watch it with. Its audit detail carries the plan's
  tool and summary, not its body.
- Test the apply half of `TestWritesNeedAllThreeSwitches`, and `TestApplySendsExactlyThePlan`.

### Task 4: Restart and redeploy

- `plan_restart_app(project, env, app)`: `get_app_status`'s tasks now; plan `POST <app>/restart` `{}`.
- `plan_redeploy_app(project, env, app, imageTag?, noCache?)`: the latest deployment's source (as `get_app`
  reads it). Image app: before `image:tag`, after `image:<imageTag>` or the same; a repository app refuses
  `imageTag`. Plan `POST <app>/deploy {activeMethod, imageSource: {imageTag}, noCache}`; `Check` is the
  source seen, and apply re-reads it and refuses a changed one.
- Tests: the plan's answer; no write before apply; the sent body; drift refused.

### Task 5: Install

- `plan_install_app`: the fields of `preflight_install`. Reads `GET /app-templates/:name` for the version,
  variant, dependencies and components, runs the preflight, and answers what would be created: the app
  (`name`), each component (`<name>-<component>` is the create's naming - read it from
  `apptemplateuc` rather than guess), each dependency, with images; the preflight's issues and leftover
  storage. Issues mean no token. Plan `POST <env>/apps/from-template` with the body.
- `forAudit` as `preflightInput`; the stored plan is sealed. Test that neither holds a parameter's value.

### Task 6: Configuration

- `plan_update_app_config(project, env, app, yaml)`: export the app; parse the model's YAML into a
  `yaml.Node` (refuse what is not a mapping); replace `apps.<key>` in the env document; re-tar and gzip,
  file order and the manifest kept. `POST <env>/spec/import/validate` with `{bundle, selection: {include:
  [<the app node's path>]}, options: {existing: update}}`; find the app's node by kind and key.
- Answer the node's action, changes, `restart`, `deploy`, issues and notes, and the summary counts; a node
  whose action is not `update`, or a blocking issue, gets no token. A document with no change answers
  "nothing to change". Plan `POST <env>/spec/import/apply` with the same body plus `planHash`.
- Tests with a bundle built as the export builds it: the replaced document; the request; a `planHash`
  refusal from the endpoint surfaced as "the app changed since; plan again".

### Task 7: Scheduled jobs

- `plan_create_sched_job(project, env, app, name, cronExpr | interval, timeZone, command, timeout?, maxRetry?)`:
  the next five runs as `explain_schedule` computes them, with the same `initialTime`. Plan
  `POST <app>/sched-jobs` with `{name, jobType: container-command, schedule, app: {id}, command: {kind,
  command}, timeout, maxRetry}` - the kind and fields read from the dashboard's own create request.

### Task 8: Prompts and instructions

- The server instructions, on the full server only, add: "Tools named plan_* change nothing and answer a
  plan; show it to the person and call apply_plan only once they agree."
- `install_app` and `debug_app` prompts as in the spec §7, on the full server.

### Task 9: Dashboard

- The MCP tab: **Allow changes** beside Enabled, saved with it, with what it allows.
- **Use your own key**: two inputs, key ID and secret, kept in component state only, which fill the three
  snippets. A line under them: read tools need read; restart and redeploy need execute; install, config and
  jobs need write; create the key in Profile › API keys.
- `npx tsc --noEmit -p .`, eslint, prettier; commit.

### Task 10: End to end, and merge

- `TestAgainstARealServer` gains, when `HP_TEST_MCP_WRITE=1`, `plan_restart_app` and `apply_plan` on the
  given app.
- By hand, with the person's go-ahead to restart the backend: Claude Code plans and applies a restart, an
  env var change and an install on the dev cluster.
- `go build ./... && golangci-lint run ./... && go test ./...`; merge phase 1 then phase 2, `--no-ff`, both
  repositories.
