# Docker API Access - Host Mode Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** An administrator, with the privileged-apps switch on, can give an app the node's own Docker socket instead of the proxy. Everything that gives, keeps, exports or imports Docker API access knows the two modes.

**Architecture:**
- **The mode.** The `app-docker-api` setting gains `mode`. `dockerapiservice` turns it into mounts:
  - `proxy` is the socket volume and the app's network, as before;
  - `host` is a bind of `/var/run/docker.sock` at the same path.
- **Every path follows the mode.** `ApplyToService`, `Detach` and the import writer all read it. The agents' policies leave host-mode apps out.
- **One gate for raw access to the host,** checked in two places: the screen's `PUT` and the import planner. It requires the switch on and an administrator. Import's raw host mounts move to the same gate.
- **The security settings** list the apps in host mode.

**Tech Stack:** Go 1.27, the services of plans 2-4.

**Spec:** `docs/superpowers/specs/2026-09-24-docker-api-access-design.md` §15 (with §9's switch paragraph).

## Global Constraints

- `mode` is `proxy`, `host`, or empty for `proxy`. Any other value is refused.
- **Host mode:**
  - the mount is `{type: bind, source: /var/run/docker.sock, target: /var/run/docker.sock}`;
  - `HIVEPAAS_DOCKER_HOST=unix:///var/run/docker.sock`;
  - no socket volume, no app network, no policy on any agent;
  - the rest of the block is kept but not checked.
- **Gate:** the switch on **and** an administrator, for:
  - entering host mode on the screen, which includes turning a disabled host mode back on;
  - an import that creates or changes a block in host mode;
  - an import's raw host mounts.
- **Codes:**
  - `DOCKER_SOCKET_NOT_PERMITTED` for host mode in an import;
  - `HOST_MOUNT_NOT_PERMITTED` stays for raw host mounts.
- **Templates never ask for host mode.** Both the template validator and the buildable subset refuse it.
- **Before this work is done:**
  - `go build ./...`;
  - `golangci-lint run ./...` (0 issues, 120 columns, US spelling);
  - `go test ./...`;
  - `make gen-swag`.

  Tests use testify `assert`; `require` is not vendored.
- **Git:**
  - branch `feat/docker-api-host-mode`;
  - every commit ends with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`;
  - merge into `main` locally at the end and delete the branch. Do not push.

---

### Task 1: The mode, and what it gives a service

**Files:**
- Modify: `hivepaas_app/entity/setting_app_docker_api.go`: `Mode`, `DockerAPIModeProxy`, `DockerAPIModeHost`, `IsHostMode()`
- Modify: `hivepaas_app/service/specservice/specmodel/docker_api.go`: check the mode; host mode skips the rest; the buildable subset refuses host mode
- Modify: `hivepaas_app/service/apptemplateservice/templatemodel/docker_api.go`: refuse host mode
- Modify: `hivepaas_app/service/dockerapiservice/types.go`, `attach.go`, `widen.go`
- Modify: `hivepaas_app/service/dockerapiservice/service.go`, `dockerapiserviceimpl/access.go`, `policies.go`: `HostModeApps`
- Modify: `hivepaas_app/service/envvarservice/envvarserviceimpl/app_system_env.go`
- Tests: beside each

**Interfaces (produced):**
- `entity.AppDockerAPISettings.Mode string` (`json:"mode,omitempty"`), and `(*AppDockerAPISettings).IsHostMode() bool`, which is nil-safe.
- In `dockerapiservice`:
  - `HostSocketPath = "/var/run/docker.sock"`;
  - `HostSocketMount() mount.Mount`;
  - `IsSocketMount(m)`, now true for the host socket mount too;
  - `AttachHost(spec)`;
  - `EntersHostMode(prev, next) bool`.
- `Service.HostModeApps(ctx, db) ([]*entity.App, error)`, which loads each app's Project and ProjectEnv.
- `Widens`:
  - an access in host mode covers every other, so leaving host mode never widens;
  - entering host mode always widens.

- [ ] **Step 1: Write the failing tests.**
  - **Mode validation:** `DockerAPIProblem` refuses `mode: root` and accepts `mode: host` with no images.
  - **Templates:**
    - the buildable subset refuses `mode: host`;
    - a template declaring `mode: host` fails validation.
  - **Attach:**
    - `AttachHost` puts `HostSocketMount()` on a spec, once, and removes a socket volume;
    - `Attach` removes `HostSocketMount()`;
    - `Detach` removes both;
    - `KeepSocketMounts` keeps the host socket;
    - `EntersHostMode` holds for nil→host, proxy→host and disabled→host, and not for host→host or host→proxy.
  - **`ApplyToService`**, with a fake `AccessOf` through the setting repo:
    - host mode gives the bind and no app network;
    - proxy mode gives the volume and the network.
  - **`Policies`** leaves out a host-mode setting; **`HostModeApps`** returns only host-mode apps.
  - **`dockerAPIEnvVars`** gives `unix:///var/run/docker.sock` in host mode and the proxy's path otherwise.
- [ ] **Step 2: Run to see them fail:** `go test ./entity/... ./service/specservice/... ./service/apptemplateservice/... ./service/dockerapiservice/... ./service/envvarservice/...`
- [ ] **Step 3: Implement.**
  - **`ApplyToService`:**
    - with no access, detach;
    - in host mode, detach the proxy's parts and `AttachHost`;
    - in proxy mode, `EnsureNetwork` and `Attach`, which also drops the host socket.
  - **`dockerAPIEnvVars`** takes the setting, or nil, and returns an error when it cannot be read.
- [ ] **Step 4: Run the tests again.** Expected: PASS
- [ ] **Step 5: Commit** `feat(dockerapi): host mode, the node's own socket bound where apps look for it`

### Task 2: Import

**Files:**
- Modify: `hivepaas_app/service/specservice/types.go`: `Admin bool`; the doc of `MayWriteCluster` no longer names host mounts
- Modify: `hivepaas_app/service/specservice/specserviceimpl/import_checks.go`, `import_apply_apps.go`
- Modify: `hivepaas_app/service/specservice/specmodel/importplan.go`: `CodeDockerSocketNotPermitted = "DOCKER_SOCKET_NOT_PERMITTED"`
- Modify: `hivepaas_app/usecase/specuc/import_validate.go`: `Admin: auth.User.IsAdmin()`
- Tests: `import_docker_api_test.go`, `import_host_access_test.go`, `usecase/specuc/import_validate_test.go`

- [ ] **Step 1: Write the failing tests.**
  - **A host-mode block** is:
    - skipped with `DOCKER_SOCKET_NOT_PERMITTED` when the switch is off, or when the caller is not an administrator, and the action says which;
    - planned as an update when both hold, without asking `MayWriteCluster`.
  - **A raw host mount:**
    - skipped when the caller is not an administrator, whatever `MayWriteCluster` says;
    - kept when the caller is, with the switch on.
  - **A docker mount of exactly `HostSocketMount()`** is refused like a socket volume.
  - **A created app in host mode** is provisioned with `HostSocketMount()` and without calling `EnsureNetwork`.
  - **The use case** passes `Admin`.
- [ ] **Step 2: Run to see them fail:** `go test ./service/specservice/... ./usecase/specuc/...`
- [ ] **Step 3: Implement.**
  - **`checkDockerAPI`** checks host mode before the Cluster Write check, and returns after it.
  - **`checkHostMounts`** checks the switch, then `p.req.Admin`. It no longer asks `mayWriteCluster`, whose only callers are now capabilities and the proxy's Docker API.
  - **`socketMounts`** also names a docker mount equal to the host socket bind.
  - **`giveDockerAPI`** reads the setting's mode.
- [ ] **Step 4: Run the tests again.** Expected: PASS
- [ ] **Step 5: Commit** `feat(spec): import gives host mode and raw host mounts only with the switch on and an administrator`

### Task 3: The screen's API and the security settings

**Files:**
- Modify: `hivepaas_app/usecase/appsettingsuc/appsettingsdto/docker_api_settings_get.go`, `docker_api_settings_update.go`: `mode`, and `hostMode: {available, blockedBy}`
- Modify: `hivepaas_app/usecase/appsettingsuc/docker_api_settings_get.go`, `docker_api_settings_update.go` (+ test)
- Modify: `hivepaas_app/usecase/system/hpappsettingsuc/uc.go`, `security_settings_get.go`, `hpappsettingsdto/security_settings_get.go`: `privilegedApps`
- Regenerate: `docs/openapi/swagger.json`

- [ ] **Step 1: Write the failing tests.**
  - **`hostModeBlockedBy(cfg, auth)`** is `"switch"`, `"admin"` or `""`, with the switch checked first.
  - **`checkDockerAPIGrant`:**
    - entering host mode without the switch or without admin is `ErrUnauthorized`, and never asks for Cluster Write;
    - host→proxy asks nothing.
  - **`dockerAPIProblem`** in host mode does not check `sharedDirs` against the service.
  - **`applyAppDockerAPI`:**
    - on proxy→host it syncs, updates the service to the host socket and applies the environment;
    - on host→host it only syncs.
  - **`TransformSecuritySettings`** lists the apps given to it, and gives an empty list for none.
- [ ] **Step 2: Run to see them fail.**
- [ ] **Step 3: Implement.**
  - **`GET`** takes the caller and adds `hostMode`.
  - **`PUT`:**
    - applies the host gate when `EntersHostMode`, and the Cluster Write gate when only `Widens`;
    - updates the service and the environment when the presence or the mode changes.
  - **Security settings:** `UC` gets `dockerAPIService`. `GetSecuritySettings` adds `privilegedApps: [{appId, appName, projectId, projectName, projectEnvId, projectEnvName}]`.
- [ ] **Step 4: Run the tests, then `make gen-swag`.** Expected: PASS
- [ ] **Step 5: Commit** `feat(settings): host mode on the Docker API screen, and the apps in it on the security settings`

### Task 4: Gates and merge

- [ ] `go build ./...`, `golangci-lint run ./...` and `go test ./...` from the repo root.
- [ ] Merge `feat/docker-api-host-mode` into `main` locally, and delete the branch.
