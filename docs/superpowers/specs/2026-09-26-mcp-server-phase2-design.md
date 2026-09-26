# The MCP Server, Phase 2: Tools That Change Things

Phase 1 (`2026-09-26-mcp-server-design.md`) let an assistant read HivePaaS: why an
app is down, what its logs say, what a template would install. The person then
went to the dashboard to act on it. This phase lets the assistant act - install
an app, restart or redeploy one, change its configuration, schedule a job -
without ever acting in one step the person did not see.

---

## Decisions

1. **Every change is a plan, then an apply.** A `plan_*` tool changes nothing: it
   checks the change the way the endpoint that makes it would, and answers what
   would happen and a `planToken`. `apply_plan(planToken)` makes exactly that
   change, and nothing else can be applied. A model cannot skip the plan, and the
   person sees the plan - in the model's answer - before the apply is called.
2. **Applying is dispatching, as reading is.** The plan stores the request the
   dashboard would send; the apply sends it to the backend's own router as the
   caller, as every phase 1 tool does. The endpoint's permission check, its
   validation, its audit entry and its refusals are the ones that apply.
3. **Three switches, all needed.** The `mcp` setting's `allowWrite` (off by
   default, an administrator's), a key whose access actions include write or
   execute, and the permission the endpoint itself checks. With either of the
   first two missing, the write tools are not listed at all.
4. **A plan is bound, short-lived and single-use.** To the user and the API key
   that made it, for ten minutes, and gone once applied or refused. A token from
   another key is not found, the same as one that expired.
5. **A plan holds nothing readable at rest.** It may carry an install's password.
   It is kept in redis encrypted with a key that exists only inside the token.
6. **Nothing destructive.** No tool deletes an app, a volume, a project, a
   setting or data. An app can be restarted, redeployed or reconfigured; removing
   it stays in the dashboard.

## 1. The tools

| plan tool | what the plan answers | apply sends | needs |
|---|---|---|---|
| `plan_install_app` | the template, version and variant; the apps that would be created - the app, its components, its dependencies - each with its name and image; the preflight's issues and leftover data | `POST .../apps/from-template` | write on the env |
| `plan_restart_app` | the app's containers now, and that the swarm replaces them (a forced update, as the dashboard's Restart) | `POST .../apps/:id/restart` | execute on the app |
| `plan_redeploy_app` | the source now and after: the image and tag, or the repository and ref; `imageTag` changes the tag for this deployment | `POST .../apps/:id/deploy` | execute on the app |
| `plan_update_app_config` | the configuration import's plan for the app: each setting that changes, whether it restarts or redeploys the app, the import's issues | `POST .../spec/import/apply` of the env | write on the env |
| `plan_create_sched_job` | the job, its command, and its next five runs in the zone given | `POST .../apps/:id/sched-jobs` | write on the app |

`apply_plan(planToken)` answers what was done and what to follow: the deployment
or task id, with `get_app_status`, `get_app_logs` or `get_task_logs` named to
follow it.

**Inputs** name the project, env and app by key or name, as in phase 1.

- `plan_install_app` takes the same fields as `preflight_install`. A secret
  parameter left out is generated, as in the dashboard.
- `plan_redeploy_app` takes `imageTag` (an image app) and `noCache` (a repository
  app); nothing else about the source changes here - that is configuration.
- `plan_update_app_config` takes `yaml`: the app's whole document, as
  `get_app_config` answers it, edited. A model reads the configuration, changes
  what it must, and hands the document back; the import says what differs.
- `plan_create_sched_job` takes `name`, `cronExpr` or `interval`, `timeZone`,
  `command` (run in the app's container), and optional `timeout` and `maxRetry`.

## 2. How `plan_update_app_config` works

The configuration spec already plans before it applies: `POST .../spec/import/validate`
answers an `ImportPlan` - per node, the settings that change, `restart`, `deploy`,
issues - and a `planHash`; `.../apply` takes the same bundle and that hash, and
refuses if the target changed since. The tool uses it unchanged:

1. Export the app (`secretsMode: omit`), as `get_app_config` does.
2. Replace the app's document in the env document with the model's YAML, and
   pack the bundle again.
3. Validate it at the env's scope, selecting the app's node only, with
   `existing: update`.
4. Answer the app's node: its action, its changes, `restart`, `deploy`, its
   issues and notes. A plan with an issue that blocks gets no token.

The token keeps the bundle, the selection and the `planHash`. What exists keeps
its secrets in every mode (`checkSecrets`), so an omit bundle changes no secret;
a secret the YAML adds is created empty, and the plan says so.

## 3. The plan token

```
mcpp_<planID>.<key>
```

- `planID`: a ULID; the redis key is `mcp:plan:<planID>`, with a ten-minute TTL.
- `key`: 32 random bytes, base64url, never stored. The value in redis is the plan
  sealed with it, AES-256-GCM, directly: the key is random already, so none of
  `cryptoutil`'s argon2 derivation, which is for passwords. Redis alone reads
  nothing.
- The plan: the tool, the caller's user ID and API key ID, the request (method,
  path, body), the summary shown, and what apply re-checks (§4).

`apply_plan` takes the plan with `GETDEL`, so it is used once whatever happens
next; a refused apply needs a new plan. A token that does not decrypt, has
expired, or belongs to another user or key is answered the same way: no such
plan.

## 4. What apply checks again

A plan is a promise about the state it saw. Apply refuses, and asks for a new
plan, when that state moved:

- `update_app_config`: the import's own `planHash` - the endpoint refuses.
- `redeploy_app`: the app's deployment source (image, or repository and ref) is
  compared with the plan's before sending.
- `install_app`: the create endpoint's own checks - a name taken since, a volume
  used since.
- `restart_app`, `create_sched_job`: nothing; they do not depend on what changed.

And, every time: `allowWrite` is still on, and the key still exists and may do
it - the dispatch runs as the caller, now.

## 5. Listing, annotations, audit

- **Listing.** The endpoint serves one of two servers per request: the read-only
  one of phase 1, or the full one when `allowWrite` is on and the key's access
  actions include write or execute. A client sees only what it can use.
- **Annotations.** `plan_*`: `readOnlyHint: true` - a plan changes nothing.
  `apply_plan`: `readOnlyHint: false`, `destructiveHint: true`,
  `idempotentHint: false`, so a client that asks before acting asks here.
- **The model is told**, in the server's instructions and in each plan's answer:
  show the person the plan and apply only once they agree.
- **Audit.** Every call is an `mcp-tool-call` entry, as in phase 1. A plan's entry
  names the plan; the apply's names the plan and the tool it applies, and its
  summary. The endpoint the apply reaches writes its own entry as well - the
  app update, the app creation - as for any call through the API.

## 6. Dashboard

The MCP tab (System settings › AI) gains:

- **Allow changes** (`allowWrite`), off by default, with what it lets an
  assistant do and that each change is planned first.
- **No button makes a key that can make changes.** The person creates it
  themselves, in Profile › API keys, choosing its access actions, and pastes its
  ID and secret into the tab, which fills the client snippets with it. The pasted
  key stays in the page: it is not saved or sent anywhere. The tab says which
  access actions each tool needs (write, execute).
- In **Recent calls**, an apply shows the tool it applied.

## 7. Prompts

- `install_app` goes on to `plan_install_app`, shows the plan, and applies once
  the person agrees, when changes are allowed.
- `debug_app` may end with a plan - a restart, a configuration change - and asks
  before applying it.

## 8. Testing

- **Plan store:** a token round-trips; another user's or key's token, an expired
  one, a used one and a tampered one are all "no such plan"; redis holds no
  plaintext of the body (`TestAPlanIsUnreadableAtRest`).
- **Gating:** with `allowWrite` off, or a read-only key, no write tool is listed
  and `apply_plan` of an earlier token is refused (`TestWritesNeedAllThreeSwitches`).
- **Each tool:** over a fake router, the plan's answer, that nothing was sent
  before apply, and that apply sends exactly the planned request.
- **Drift:** a redeploy whose source changed after the plan is refused.
- **Config:** a bundle edited by YAML validates to the expected changes, and a
  `planHash` mismatch from the endpoint is surfaced as "plan again".
- **Secrets:** an install's password is in neither the audit entry nor redis in
  clear.
- **By hand:** Claude Code installs a template, restarts an app and changes an
  env var, against the dev cluster, with the person approving each apply.

## 9. Not in this phase

- **Deleting anything** (§ Decisions 6).
- **Stopping or starting an app** (`running-status`): easy to add as a plan
  tool; left out until the ones above have been used.
- **Setting a secret's value.** A secret is set in the dashboard; the plan of a
  configuration change says which secrets it created empty.
- **Rollback.** A redeploy of an earlier image tag is the way back for an image
  app; a tool for it waits for the deployments to carry what they deployed.
