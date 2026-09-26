# MCP Tools That Follow the API - Implementation Plan

**Spec:** `docs/superpowers/specs/2026-09-26-mcp-tools-follow-the-api-design.md`. No `dryRun` in this scope.
**Branch:** `feat/mcp-tools-follow-the-api` in hivepaas; merged `--no-ff`, never pushed.

## Review focus

- **Only the kind asked for is fetched.** `get_app_settings` sends one GET.
- **The plan shows values.** Each changed field is `{path, before, after}`; a masked secret is never unmasked.
- **Apply sends what the plan showed, with the `updateVer` it read**; a mismatch is "plan again".
- **Every registry entry is a real route** (`TestMCPSettingsKindsAreRoutes` in the server package).

### Task 1: Settings kinds

- `interface/mcp/settings.go`: the registry (kind, path, the fields a GET answers that a PUT does not
  take), `SettingsEndpoints()` for the route test, `get_app_settings`, `plan_update_app_settings`.
- JSON merge patch (RFC 7396) and the field diff in `interface/mcp/mergepatch.go`, with tests.
- `interface/api/server/router_mcp_test.go`: every kind's GET and PUT are registered.
- Tests over a fake router: one GET; the patch, the diff, the body and its `updateVer`; the mismatch;
  env vars' inherited lists not sent; a masked value kept masked.

### Task 2: Actions and deployments

- `plan_redeploy_app` takes `noCache` only and follows the task the endpoint answers.
- `plan_set_app_running`, `list_app_deployments`, `get_app_deployment`, `plan_cancel_deployment`.

### Task 3: Remove the spec tools

- Delete `get_app_config`, `plan_update_app_config` and their bundle helpers and tests; the secrets test
  reads `kind` settings instead.

### Task 4: The read tools decode into the use cases' DTOs

- Replace `apiApp`, `apiAppDetail`, `apiDeployment`, `apiServiceTask`, `apiTask`, `apiNode`,
  `apiTemplateSummary`, `apiSchedJob`, `apiDeploySettings`, `apiProject`, `attentionItem`'s source with the
  DTOs; answers unchanged, tests unchanged.

### Task 5: Check and merge

- `go build ./... && golangci-lint run ./... && go test ./...`; `make gen-swag` if a DTO moved; merge.
