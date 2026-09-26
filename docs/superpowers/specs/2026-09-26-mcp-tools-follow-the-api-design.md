# MCP Tools That Follow the API

Phases 1 and 2 (`2026-09-26-mcp-server-design.md`, `2026-09-26-mcp-server-phase2-design.md`)
gave the MCP server tools shaped by what an assistant asks. Several of them drifted
from the API they call: they declare their own copies of its types, read an app's
configuration through the configuration spec's export, and change it through its
import. This revision lines the tools up with the API: a tool is an endpoint, or a
GET and PUT pair, of one use case, and speaks that endpoint's data.

---

## Decisions

1. **A tool follows a use case's endpoint.** The app's use cases - `appactionuc`,
   `appdeploymentuc`, `appsettingsuc`, `apppreviewuc` - and their endpoints are the
   map. A tool is named after what the endpoint does, takes what it takes, and
   answers what it answers.
2. **The API's types, not copies.** A tool decodes an answer into the use case's own
   DTO (`appsettingsdto.GetAppEnvVarsResp`, `appdto.ListAppResp`) rather than a type
   of its own that mirrors it. What a model is given is the endpoint's `data`, as
   JSON, unchanged - except that a list or a log is bounded, as every answer is.
3. **Settings are read and changed one kind at a time.** An app's settings are kinds,
   each a GET and PUT pair: env vars, resources, routing, and so on. An assistant asked
   about env vars reads env vars, and changes them, and nothing else is fetched.
4. **The configuration spec is not a tool.** Export and import are for installing
   again elsewhere; their documents are rewritten for that, and are not the settings
   the API reads and writes. `get_app_config` and `plan_update_app_config` go.
5. **Plans and their safety stay as phase 2 made them.** A change is still a plan, then
   an apply, bound to its caller, sealed, used once. A settings change is refused at
   apply when the settings moved since the plan: every settings update checks
   `updateVer` (`ErrUpdateVerMismatched`), which takes the place of the import's
   `planHash`.

## 1. Settings

A registry maps each kind to its endpoint pair and what the pair needs:

| kind | endpoint (under the app) | change needs |
|---|---|---|
| `env-vars` | `/env-vars` | Write |
| `deployment` | `/deployment-settings` | Write |
| `routing` | `/routing-settings` | Write |
| `service` | `/service-settings` | Write |
| `network` | `/network-settings` | Write |
| `resource` | `/resource-settings` | Write |
| `container` | `/container-settings` | Write |
| `storage` | `/storage-settings` | Write |
| `feature` | `/feature-settings` | Write |
| `kind` | `/kind-settings` | Write |
| `docker-api` | `/docker-api-settings` | Write |

Clone settings are not here: cloning is an action, not a setting.

**`get_app_settings(project, env, app, kind)`** sends the GET and answers its `data`.
Secrets are masked as the endpoint masks them: the tool never asks for
`revealSecrets`.

**`plan_update_app_settings(project, env, app, kind, changes)`**:

- `changes` is a JSON merge patch (RFC 7396) of what `get_app_settings` answered: the
  fields to change, and nothing else. A list is replaced whole - to add an env var, a
  model sends the whole list - and `null` removes a field.
- The plan reads the settings now, drops what the kind answers but does not take (the
  env vars' inherited lists, `updateVer`), applies the patch, and answers each changed
  field as `{path, before, after}`: the person sees values, not setting names.
- The plan keeps the PUT body with the `updateVer` it read. Apply sends it; a
  mismatch is refused as "the settings changed since the plan; plan again".
- A masked secret sent back unchanged keeps the stored secret, as the dashboard's
  forms rely on; the plan never shows a secret's value.
- The endpoints have no dry run: a value they refuse is refused at apply, in their
  words. A `dryRun` query on the settings updates would move that refusal into the
  plan; it is a change to the settings handlers, proposed separately.

## 2. Actions and deployments

| tool | endpoint | needs |
|---|---|---|
| `plan_restart_app` | `POST /restart` | Execute |
| `plan_redeploy_app(noCache)` | `POST /deploy` - answers `deploymentId`, `taskId` | Execute |
| `plan_set_app_running(running)` | `POST /running-status` | Execute |
| `list_app_deployments` | `GET /deployments` | Read |
| `get_app_deployment` | `GET /deployments/:id`, and the last lines of `/deployments/:id/logs` | Read |
| `plan_cancel_deployment` | `POST /deployments/:id/cancel` | Execute |

A redeploy deploys the app's source as its settings say, and no longer changes it:
another image tag or branch is a change to the `deployment` settings, planned with
`plan_update_app_settings`, then a redeploy. `plan_redeploy_app` loses `imageTag` and
`repoRef`, which the endpoint no longer takes.

## 3. What stays, and what is refactored

- The read tools keep their names and answers, and decode into the use cases' DTOs
  instead of their own mirror types (`apiApp`, `apiTask`, `apiNode`, ...).
- `plan_install_app` and `preflight_install` already speak the template endpoints'
  request; they keep it.
- `plan_create_sched_job` stays. The app's other collections - scheduled jobs beyond
  creating one, config files, setting mounts, secrets - get tools of the same shape
  (list, get, plan create, plan update) in the revision after this one; secrets'
  values are not read or written by any tool.

## 4. Testing

- Every kind in the registry has a route in the backend's router, found by method
  and path, so a renamed endpoint fails a test rather than a user.
- `plan_update_app_settings`: the patch applied to the settings read, the diff shown,
  the body sent at apply, and an `ErrUpdateVerMismatched` refusal worded as "plan
  again"; the env vars' inherited lists are never sent.
- The read tools answer as before, decoded through the DTOs.
- Against a running server: read and change an env var, change a resource limit,
  redeploy, and follow the deployment to its end.
