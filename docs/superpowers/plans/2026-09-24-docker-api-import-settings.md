# Docker API Access - Import and the Settings Screen's API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Import gives an app the Docker API only when the operator may grant it, and gives it back the socket and network export left out. Raw host mounts in an import also need the privileged-apps switch. An app's settings gain a Docker API screen's endpoints.

**Architecture:**
- **Import, planning.** One more check in `checkPermissions`, beside capabilities:
  - it validates the block;
  - it skips the app with `DOCKER_API_NOT_PERMITTED` when the operator lacks Write on the Cluster module.
- **Host mounts in import.** `checkHostMounts` now:
  - refuses outright a mount of an `hp-dapi-sock-*` volume;
  - refuses every other host mount while the switch is off. The switch reaches the planner as a request field set by the use case.
- **Import, writing.**
  - A created app's service gets its socket and network while it is provisioned.
  - After the commit, the agents are synced first. Then each updated app's service is brought to its access with `ApplyToService`, in the same single update as its deployment blocks.
- **The screen's API.**
  - `GET` and `PUT /apps/{id}/docker-api-settings`.
  - Turning access on, or letting it do more, needs Write on the Cluster module; the rest needs Write on the app.
  - The row is saved in the transaction. The agents, the service and the environment follow after the commit, as the env vars screen does.

**Tech Stack:** Go 1.27, gin handlers, swag, the services of plans 2 and 3.

**Spec:** `docs/superpowers/specs/2026-09-24-docker-api-access-design.md` (§1, §7, §8, §9).

## Global Constraints

- **Gate:** Write on the Cluster module whenever access is granted:
  - an import that creates the block or changes it;
  - the screen turning it on or widening it.

  Narrowing or turning it off needs only the app's Write.
- **Issue code:** `DOCKER_API_NOT_PERMITTED`, severity `skipped`.
- **Reserved names:** a docker mount of a volume named `hp-dapi-sock-*` is never imported.
- **The switch:** `Security.AllowPrivilegedApps` off refuses every import of a new or changed raw host mount, whoever asks.
- **Routes:** `GET|PUT /projects/{projectID}/{projectEnv}/apps/{appID}/docker-api-settings`.
- **Wire format:**
  - `limits.memory` is a data size string such as `"2gb"`;
  - `enabled: false` keeps the row's policy and disables it.
- **Before this work is done:**
  - `go build ./...`;
  - `golangci-lint run ./...` (0 issues, 120 columns, US spelling);
  - `go test ./...`;
  - `make gen-swag`.

  Tests use testify `assert`; `require` is not vendored.
- **Git:**
  - branch `feat/docker-api-import`;
  - every commit ends with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`;
  - merge into `main` locally at the end and delete the branch. Do not push.

---

### Task 1: Import checks who may grant the Docker API, and gives it back

**Files:**
- Modify: `hivepaas_app/service/specservice/specmodel/importplan.go` (code)
- Modify: `hivepaas_app/service/specservice/specmodel/docker_api.go` (`SharedDirsProblem`)
- Modify: `hivepaas_app/service/specservice/specserviceimpl/import_checks.go` (`checkDockerAPI`)
- Modify: `hivepaas_app/service/specservice/specserviceimpl/import_policy.go`
- Modify: `hivepaas_app/service/specservice/specserviceimpl/import_plan.go` (`restarts`)
- Modify: `hivepaas_app/service/specservice/specserviceimpl/import_apply_apps.go` (`configureApp`)
- Modify: `hivepaas_app/service/specservice/specserviceimpl/import_phase2.go`, `import_write.go` (`afterCommit`)
- Test: `hivepaas_app/service/specservice/specserviceimpl/import_docker_api_test.go`

**Interfaces:**
- Produces: `specmodel.CodeDockerAPINotPermitted`, `specmodel.SharedDirsProblem(dirs, targets []string) string`.

- [ ] **Step 1: Write the failing tests** in `import_docker_api_test.go`:
  - an app given the block is skipped with `DOCKER_API_NOT_PERMITTED` when `MayWriteCluster` says no, and updated with the change `settings.dockerApi` when it says yes;
  - a block wrong in itself (no images) fails the plan with `ErrSpecBundleInvalid`;
  - a created app's provisioned spec has `SocketMount(appID)` and the network `EnsureNetwork` returned;
  - after the commit, an updated app's service update carries its socket, and `SyncAgents` ran before it.

  The fake:

```go
type fakeImportDockerAPI struct {
	dockerapiservice.Service
	calls []string
}

func (f *fakeImportDockerAPI) EnsureNetwork(_ context.Context, appID string) (string, error) {
	f.calls = append(f.calls, "network "+appID)
	return "net-" + appID, nil
}

func (f *fakeImportDockerAPI) SyncAgents(context.Context) error {
	f.calls = append(f.calls, "sync")
	return nil
}

// ApplyToService attaches as the real one does for an app with access.
func (f *fakeImportDockerAPI) ApplyToService(_ context.Context, _ database.IDB, appID string,
	spec *swarm.ServiceSpec) error {
	f.calls = append(f.calls, "apply "+appID)
	dockerapiservice.Attach(spec, appID, "net-"+appID)
	return nil
}
```

- [ ] **Step 2: Run them to see them fail**

Run: `go test ./service/specservice/specserviceimpl/ -run DockerAPI`
Expected: FAIL. The code is undefined, and nothing attaches.

- [ ] **Step 3: Implement.**
  - **specmodel.**
    - Add `CodeDockerAPINotPermitted = "DOCKER_API_NOT_PERMITTED"` beside `CodeCapabilityNotPermitted`.
    - Move the coverage loop out of `checkDockerAPI` into `SharedDirsProblem(dirs, targets []string) string`, keeping its message. `checkDockerAPI` calls it.
  - **The planner's check.** `checkPermissions` calls `p.checkDockerAPI(ctx, node)` after `checkSharedMounts`, with the same skip guard:

```go
// checkDockerAPI refuses a Docker API block wrong in itself, and skips an app
// created with the block, or updated so that the block changes, when the
// operator may not grant it. Narrowing is asked about too: the planner does not
// weigh one policy against another, and a bundle is not the screen.
func (p *planner) checkDockerAPI(ctx context.Context, node *specmodel.PlanNode) error {
	name := specmodel.SingletonBlockName(base.SettingTypeAppDockerAPI)
	body, found := p.apps[node.Path].doc.Settings[name]
	if !found || !writesBlock(node, string(specmodel.BlockSettingsDockerAPI)) {
		return nil
	}
	_, data, err := decodeImportedSetting(specmodel.BlockSettingsDockerAPI, base.SettingTypeAppDockerAPI, "", body)
	if errors.Is(err, hperrors.ErrDataVerNewerThanSystemVer) {
		return nil // SETTING_VERSION_NEWER says so
	}
	if err != nil {
		return err
	}
	access, _ := data.(*entity.AppDockerAPISettings)
	if problem := specmodel.DockerAPIProblem(access); problem != "" {
		return hperrors.Wrap(hperrors.ErrSpecBundleInvalid).WithExtraDetail("%s: %s", node.Path, problem)
	}
	allowed, err := p.mayWriteCluster(ctx)
	if err != nil || allowed {
		return err
	}
	p.skipNode(node, specmodel.Issue{
		Severity: specmodel.SeveritySkipped, Code: specmodel.CodeDockerAPINotPermitted, Path: node.Path,
		Detail: map[string]any{refInSetting: string(specmodel.BlockSettingsDockerAPI)},
		Action: "not imported: giving an app the Docker API needs Write permission on the Cluster module",
	})
	return nil
}
```

  - **Import policy.** The policy covers scoped settings only; an app's are built with the app. Replace `reasonGrantsDockerAPI` with `"only an app is given the Docker API: a scope's setting would grant nothing"`.
  - **Restarts.** `restarts` also returns true for `string(specmodel.BlockSettingsDockerAPI)`: the service gains or loses the socket.
  - **A created app.** In `configureApp`, after `BuildApp`, call `w.giveDockerAPI(ctx, node, app, spec)`:

```go
// giveDockerAPI gives a new app's service the socket and network of the access
// its settings grant. The setting is provisioned with the app, so it cannot be
// read back the way ApplyToService reads it.
func (w *writer) giveDockerAPI(
	ctx context.Context, node *specmodel.PlanNode, app *entity.App, spec *swarm.ServiceSpec,
) error {
	for _, setting := range w.appSettings[node.Path] {
		if setting.Type != base.SettingTypeAppDockerAPI || setting.Status != base.SettingStatusActive {
			continue
		}
		networkID, err := w.p.s.dockerAPIService.EnsureNetwork(ctx, app.ID)
		if err != nil {
			return hperrors.Wrap(err)
		}
		dockerapiservice.Attach(spec, app.ID, networkID)
	}
	return nil
}
```

  - **An updated app** (phase 2).
    - In `applyToService`, call `updateService` when a deployment is prepared or `node.Changes` holds `settings.dockerApi`.
    - `updateService` builds the prepared deployment only when there is one. It then always calls `dockerAPIService.ApplyToService(ctx, db, app.ID, &svc.Spec)` before `ServiceUpdate`, because rebuilding storage or networks drops the socket and the app's network.
  - **After the commit.** `afterCommit` calls `w.syncDockerAPI(ctx)` before `applyToServices`. It calls `SyncAgents` when a written setting is `app-docker-api`, and ignores the error: the agents reconcile within 30 seconds.
  - **Test setup.** Tests that reach these paths set `svc.dockerAPIService = &fakeImportDockerAPI{}`.

- [ ] **Step 4: Run the package's tests**

Run: `go test ./service/specservice/...`
Expected: PASS

- [ ] **Step 5: Commit** `feat(spec): import gives an app the Docker API only when the operator may, and gives back its socket and network`

### Task 2: Host mounts in import need the switch; socket volumes are never imported

**Files:**
- Modify: `hivepaas_app/service/specservice/types.go` (`AllowPrivilegedApps`)
- Modify: `hivepaas_app/service/specservice/specserviceimpl/import_checks.go` (`checkHostMounts`)
- Modify: `hivepaas_app/usecase/specuc/import_validate.go` (`importReq`)
- Modify: `hivepaas_app/config/security.go` (the TODO)
- Test: `hivepaas_app/service/specservice/specserviceimpl/import_host_access_test.go`

- [ ] **Step 1: Write the failing tests.**
  - `planAskingCluster` sets `AllowPrivilegedApps: true`.
  - New: with the switch off, the host mount skips the app, and `MayWriteCluster` is not asked.
  - New: a docker mount `{type: volume, source: hp-dapi-sock-app_9}` skips the app even with the switch on and Cluster Write.
- [ ] **Step 2: Run to see them fail**: `go test ./service/specservice/specserviceimpl/ -run Host`
- [ ] **Step 3: Implement.**
  - **The request field.** `ValidateImportReq.AllowPrivilegedApps bool`, documented as the operator's switch. False refuses.
  - **`checkHostMounts`.**
    - After the `writesBlock` guard, collect the targets of docker mounts whose type is `volume` and whose source has `dockerapiservice.SocketVolumePrefix`. Any such target skips the app with `CodeHostMountNotPermitted` and the action `not imported: an app's Docker API socket comes with its access and is never mounted`.
    - After computing `changed`, when `!p.req.AllowPrivilegedApps`, skip the app with the action `not imported: this installation does not let apps reach the host; an administrator turns that on in the security settings`.
  - **The use case.** `importReq` sets `AllowPrivilegedApps` from `config.Current()`, and treats a nil config as off.
  - **The config comment.** Replace the TODO on the switch with what reads it now: import's host mounts.
- [ ] **Step 4: Run** `go test ./service/specservice/... ./usecase/specuc/...`. Expected: PASS
- [ ] **Step 5: Commit** `feat(spec): an imported host mount needs the privileged-apps switch, and an app's socket is never imported`

### Task 3: The Docker API screen's endpoints

**Files:**
- Create: `hivepaas_app/service/dockerapiservice/widen.go` (+ `widen_test.go`)
- Create: `hivepaas_app/usecase/appsettingsuc/appsettingsdto/docker_api_settings_get.go`, `docker_api_settings_update.go`
- Create: `hivepaas_app/usecase/appsettingsuc/docker_api_settings_get.go`, `docker_api_settings_update.go` (+ `docker_api_settings_update_test.go`)
- Create: `hivepaas_app/interface/api/handler/appsettingshandler/docker_api_settings.go`
- Modify: `hivepaas_app/interface/api/server/router_apps.go`
- Regenerate: `docs/openapi/swagger.json`

**Interfaces:**
- Produces:
  - `dockerapiservice.Widens(prev, next *entity.AppDockerAPISettings) bool`;
  - `appsettingsdto.AppDockerAPISettingsResp` and `AppDockerAPILimits` — the dashboard (plan 5) reads these;
  - `UC.GetAppDockerAPISettings` and `UC.UpdateAppDockerAPISettings`.

- [ ] **Step 1: Write the failing tests.**
  - **`Widens`** is true for:
    - nil to something;
    - an added image, unless the old list had `"*"`;
    - an added dir, network or group;
    - a higher effective limit, where zero is the default.

    It is false for narrowing, for turning off, and for no change.
  - **The use case**, with `answeringPermissionManager` and `serviceUpdatingDockerManager` from `storage_settings_update_test.go`:
    - widening asks for Cluster Write and is refused without it;
    - narrowing asks nothing;
    - a shared dir on no volume mount of the service is refused;
    - granting syncs the agents, then updates the service with the socket, then applies env vars;
    - turning off detaches the service and removes the app's children, and never syncs.
- [ ] **Step 2: Run to see them fail**: `go test ./service/dockerapiservice/ ./usecase/appsettingsuc/...`
- [ ] **Step 3: Implement.**

```go
// Widens reports whether next lets an app do anything prev did not: access where
// it had none, an image, directory, network or group it did not have, a higher
// limit. Granting that takes what giving access takes; taking any of it away
// does not. A pattern is compared as written: one covered by another prev
// already had still counts as added.
func Widens(prev, next *entity.AppDockerAPISettings) bool {
	switch {
	case next == nil:
		return false
	case prev == nil:
		return true
	}
	if !slices.Contains(prev.Images, "*") && adds(prev.Images, next.Images) {
		return true
	}
	if adds(prev.SharedDirs, next.SharedDirs) || adds(prev.Networks, next.Networks) || adds(prev.Allow, next.Allow) {
		return true
	}
	p, n := effectiveLimits(prev.Limits), effectiveLimits(next.Limits)
	return n.Containers > p.Containers || n.Memory > p.Memory || n.CPUs > p.CPUs
}
```

  - **The DTOs.**
    - `AppDockerAPISettingsResp` has `enabled`, `images`, `sharedDirs`, `networks`, `allow`, `limits`, `defaultLimits` and `updateVer`. Lists are never null.
    - `AppDockerAPILimits` has `containers`, `memory` (`unit.DataSize`, tagged `swaggertype:"string"`) and `cpus`.
    - `UpdateAppDockerAPISettingsReq` has the same fields minus `defaultLimits`, plus `ToEntity()`.
  - **GET.** Load the app, list its `app-docker-api` row whatever its status, and transform it.
  - **PUT, in the transaction.**
    - Load the app `FOR UPDATE`, its row, and its service.
    - A row whose `UpdateVer` differs from the request's is `ErrUpdateVerMismatched`.
    - `Prev` is the row's data while active.
    - **Enabled.** `Next = req.ToEntity()`. Check `DockerAPIProblem`, then `SharedDirsProblem` against the service's `volume` mounts, the socket excluded. A problem is `ErrValueInvalid` with the problem as extra detail.
    - **Disabled with no row.** A no-op.
    - **The gate.** When `Widens(Prev, Next)`, check Write on the Cluster module, and refuse without it with `ErrUnauthorized`.
    - **The row.** Upsert it with `UpdateVer+1`. Its status is active with `Next`'s data, or disabled with its data kept.
    - Audit section `docker-api`, with `enabled` and `widened`.
  - **PUT, after the commit** (`applyAppDockerAPI`). Errors are joined into `Meta.Warning`, as the env vars screen does.
    - With access: `SyncAgents`.
    - When access appears or goes:
      - `ServiceUpdateFunc` with `ApplyToService`;
      - then `BuildEnvVarsForAllAppsInScope(app scope)` and `ApplyEnvVarsForApps`, for `HIVEPAAS_DOCKER_HOST`.
    - When it goes: `RemoveAppObjects`. The app's network stays, unused, until the app is deleted; removing it now would race the task still leaving it.
  - **Handler and routes.** Two routes beside `resource-settings`, with swagger annotations in the file's style. The handler asks Read for GET and Write for PUT.
  - **Swagger.** `make gen-swag` from the repo root.
- [ ] **Step 4: Run** `go test ./service/dockerapiservice/... ./usecase/appsettingsuc/...`. Expected: PASS
- [ ] **Step 5: Commit** `feat(settings): a Docker API screen's endpoints, widening behind Write on the Cluster module`

### Task 4: Gates and merge

- [ ] `go build ./...`, `golangci-lint run ./...`, `go test ./...` in `hivepaas_app`; `make gen-swag` leaves nothing to commit.
- [ ] Spec §14: plan 4 is done; the dashboard (plan 5) renders `DOCKER_API_NOT_PERMITTED` and the screen.
- [ ] Merge `feat/docker-api-import` into `main` locally, and delete the branch.
