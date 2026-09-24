# Docker API Access - Granting It Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A template can give its app the Docker API, and the rest of HivePaaS keeps that access working: deployments, the storage and network screens, cloning, deletion and export.

**Architecture:**
- **Service helpers.** `dockerapiservice` gains a small API over a service spec: `Attach` and `Detach` for the socket mount and the app's network, plus `ApplyToService`, which reads the setting and does whichever is right.
- **The block.** A new block, `settings.dockerApi`, is checked by the buildable subset and built into the setting row plus the attachments.
- **Templates.** Templates read the block the way they read capabilities: from the template, never overridden by a version, gated on Write on the Cluster module.
- **Everything else follows.** Every place that rewrites a service's mounts or networks, or copies them elsewhere, gains one call to these helpers.

**Tech Stack:** Go 1.27, fx, moby API types, the engine and agent of plans 1 and 2.

**Spec:** `docs/superpowers/specs/2026-09-24-docker-api-access-design.md` (§1, §2, §7, §8, §9, §10).

This is the first half of spec §14's plan 3. Import (checking who may grant access, and re-attaching it), the privileged-apps switch in import, reserved volume names in import, and the settings screen's API follow in the next plan.

## Global Constraints

- **The block shape** is the entity's JSON, as every settings block is: `settings.dockerApi: {images, sharedDirs, networks, allow, limits: {containers, memory, cpus}}`.
  - `memory` is a data size such as `2gb`, and `cpus` a number such as `2`.
  - Bounds: 1-20 images, at most 5 shared directories, 0-50 containers (0 is the default of 5), memory 0 or at least 6mb, cpus 0 or more.
- **Names:**
  - socket mount: volume `hp-dapi-sock-<app id>` at `/var/run/hivepaas`;
  - app network: `hp-dapi-<app id>`, an attachable overlay labeled `hivepaas.docker-api.network=<app id>`;
  - system environment variable: `HIVEPAAS_DOCKER_HOST=unix:///var/run/hivepaas/docker.sock`.
- **Gate:** giving access is refused without Write on the Cluster module, exactly like capabilities.
- **Screens and export do not show** the socket mount or the app's network. Saving a screen keeps them. Export leaves them out, and import attaches them again in the next plan.
- **Before this work is done:** `go build ./...`, `golangci-lint run ./...` (0 issues, 120 columns, US spelling), `go test ./...`, and `make gen-swag` after the template DTO changes. Tests use testify `assert`; `require` is not vendored.
- **Git:** work on branch `feat/docker-api-grant`. Every commit ends with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`. At the end, merge into `main` locally and delete the branch. Do not push.

---

### Task 1: The service spec helpers, access, and the system variable

**Files:**
- Modify: `hivepaas_app/service/dockerapiservice/types.go`
- Create: `hivepaas_app/service/dockerapiservice/attach.go`
- Test: `hivepaas_app/service/dockerapiservice/attach_test.go`
- Modify: `hivepaas_app/service/dockerapiservice/service.go`
- Create: `hivepaas_app/service/dockerapiservice/dockerapiserviceimpl/access.go`
- Test: `hivepaas_app/service/dockerapiservice/dockerapiserviceimpl/access_test.go`
- Modify: `hivepaas_app/base/env_var.go`
- Modify: `hivepaas_app/service/envvarservice/envvarserviceimpl/app_system_env.go`
- Test: `hivepaas_app/service/envvarservice/envvarserviceimpl/app_system_env_test.go`

**Interfaces:**
- Produces:
  - in `dockerapiservice`:
    - `SocketMountTarget`;
    - `SocketMount(appID) mount.Mount`, `IsSocketMount(*mount.Mount) bool`, `IsAppNetworkName(string) bool`;
    - `Attach(spec *swarm.ServiceSpec, appID, networkID string)`, `Detach(spec *swarm.ServiceSpec, networkID string)`;
    - `KeepSocketMounts(final, current []mount.Mount) []mount.Mount`;
    - `KeepAppNetwork(final, current []swarm.NetworkAttachmentConfig, networkID string) []swarm.NetworkAttachmentConfig`.
  - `Service` methods: `AccessOf(ctx, db, appID) (*entity.AppDockerAPISettings, error)`, `EnsureNetwork(ctx, appID) (string, error)`, `AppNetworkID(ctx, appID) (string, error)`, `ApplyToService(ctx, db, appID, *swarm.ServiceSpec) error`, `DetachFromService(ctx, appID, *swarm.ServiceSpec) error`, `RemoveApp(ctx, appID) error`.
  - `base.AppSystemEnvVarDockerHost`.

- [ ] **Step 1: Create the branch**

```bash
cd /Users/tnt/go/src/github.com/hivepaas/hivepaas
git checkout -b feat/docker-api-grant
```

- [ ] **Step 2: Write the failing tests**

`hivepaas_app/service/dockerapiservice/attach_test.go`:

```go
package dockerapiservice

import (
	"path"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
)

func appSpec(mounts []mount.Mount, networks ...string) *swarm.ServiceSpec {
	spec := &swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Mounts: mounts}}}
	for _, network := range networks {
		spec.TaskTemplate.Networks = append(spec.TaskTemplate.Networks, swarm.NetworkAttachmentConfig{Target: network})
	}
	return spec
}

var dataMount = mount.Mount{Type: mount.TypeVolume, Source: "hp-vol-data", Target: "/data"}

func TestTheSocketIsMountedWhereTheProxyPutsIt(t *testing.T) {
	assert.Equal(t, path.Dir(dockerproxy.SocketPath), SocketMountTarget)
}

func TestAttachGivesTheSocketAndTheNetworkOnce(t *testing.T) {
	spec := appSpec([]mount.Mount{dataMount}, "project-net")
	Attach(spec, "app1", "net-app1")
	Attach(spec, "app1", "net-app1")

	assert.Equal(t, []mount.Mount{dataMount, SocketMount("app1")}, spec.TaskTemplate.ContainerSpec.Mounts)
	assert.Equal(t, []swarm.NetworkAttachmentConfig{{Target: "project-net"}, {Target: "net-app1"}},
		spec.TaskTemplate.Networks)
}

func TestAttachReplacesAnotherAppsSocket(t *testing.T) {
	spec := appSpec([]mount.Mount{dataMount, SocketMount("source-app")})
	Attach(spec, "copy", "net-copy")
	assert.Equal(t, []mount.Mount{dataMount, SocketMount("copy")}, spec.TaskTemplate.ContainerSpec.Mounts)
}

func TestDetachTakesTheSocketAndTheNetworkAway(t *testing.T) {
	spec := appSpec([]mount.Mount{dataMount, SocketMount("app1")}, "project-net", "net-app1")
	Detach(spec, "net-app1")
	assert.Equal(t, []mount.Mount{dataMount}, spec.TaskTemplate.ContainerSpec.Mounts)
	assert.Equal(t, []swarm.NetworkAttachmentConfig{{Target: "project-net"}}, spec.TaskTemplate.Networks)

	// An app that never had a network of its own keeps every network it has.
	spec = appSpec(nil, "project-net")
	Detach(spec, "")
	assert.Equal(t, []swarm.NetworkAttachmentConfig{{Target: "project-net"}}, spec.TaskTemplate.Networks)
}

func TestKeepSocketMountsCarriesTheSocketThroughAScreensSave(t *testing.T) {
	current := []mount.Mount{dataMount, SocketMount("app1")}
	other := mount.Mount{Type: mount.TypeVolume, Source: "hp-vol-data", Target: "/other"}
	// The screen does not show the socket, so what it saves lacks it.
	assert.Equal(t, []mount.Mount{other, SocketMount("app1")}, KeepSocketMounts([]mount.Mount{other}, current))
	// A client that sent it back anyway does not get it twice.
	assert.Equal(t, []mount.Mount{other, SocketMount("app1")},
		KeepSocketMounts([]mount.Mount{other, SocketMount("app1")}, current))
}

func TestKeepAppNetworkCarriesTheNetworkThroughAScreensSave(t *testing.T) {
	current := []swarm.NetworkAttachmentConfig{{Target: "project-net"}, {Target: "net-app1"}}
	final := []swarm.NetworkAttachmentConfig{{Target: "project-net", Aliases: []string{"web"}}}
	assert.Equal(t, []swarm.NetworkAttachmentConfig{{Target: "project-net", Aliases: []string{"web"}},
		{Target: "net-app1"}}, KeepAppNetwork(final, current, "net-app1"))
	assert.Equal(t, final, KeepAppNetwork(final, current, ""), "no network of its own, nothing to keep")
	assert.Equal(t, final, KeepAppNetwork(final, current[:1], "net-app1"), "not attached, not added")
}

func TestAppNetworkNamesAreRecognized(t *testing.T) {
	assert.True(t, IsAppNetworkName(NetworkName("app1")))
	assert.False(t, IsAppNetworkName("shop_prod_net"))
	socket := SocketMount("app1")
	assert.True(t, IsSocketMount(&socket))
	assert.False(t, IsSocketMount(&dataMount))
}
```

`hivepaas_app/service/dockerapiservice/dockerapiserviceimpl/access_test.go`:

```go
package dockerapiserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// fakeNetworks holds networks by name, and records what is created and removed.
type fakeNetworks struct {
	docker.Manager
	byName  map[string]string
	created map[string]client.NetworkCreateOptions
	removed []string
}

func (f *fakeNetworks) NetworkInspect(_ context.Context, name string, _ ...docker.NetworkInspectOption) (
	*client.NetworkInspectResult, error) {
	id, found := f.byName[name]
	if !found {
		return nil, hperrors.Wrap(hperrors.ErrInfraNotFound)
	}
	return &client.NetworkInspectResult{Network: network.Inspect{Network: network.Network{ID: id, Name: name}}}, nil
}

func (f *fakeNetworks) NetworkCreate(_ context.Context, name string, options ...docker.NetworkCreateOption) (
	*client.NetworkCreateResult, error) {
	opts := client.NetworkCreateOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	f.created[name] = opts
	f.byName[name] = "id-" + name
	return &client.NetworkCreateResult{ID: "id-" + name}, nil
}

func (f *fakeNetworks) NetworkRemove(_ context.Context, name string, _ ...docker.NetworkRemoveOption) (
	*client.NetworkRemoveResult, error) {
	f.removed = append(f.removed, name)
	return &client.NetworkRemoveResult{}, nil
}

func accessService(settings ...*entity.Setting) (*service, *fakeNetworks) {
	networks := &fakeNetworks{byName: map[string]string{}, created: map[string]client.NetworkCreateOptions{}}
	return &service{settingRepo: &fakeSettingRepo{settings: settings}, dockerManager: networks}, networks
}

func TestEnsureNetworkCreatesTheAppsNetworkOnce(t *testing.T) {
	svc, networks := accessService()
	id, err := svc.EnsureNetwork(context.Background(), "app1")
	assert.NoError(t, err)
	assert.Equal(t, "id-hp-dapi-app1", id)
	created := networks.created["hp-dapi-app1"]
	assert.Equal(t, docker.NetworkDriverOverlay, created.Driver)
	assert.True(t, created.Attachable)
	assert.Equal(t, map[string]string{dockerapiservice.NetworkLabel: "app1"}, created.Labels)

	networks.created = map[string]client.NetworkCreateOptions{}
	id, err = svc.EnsureNetwork(context.Background(), "app1")
	assert.NoError(t, err)
	assert.Equal(t, "id-hp-dapi-app1", id)
	assert.Empty(t, networks.created)
}

func TestApplyToServiceAttachesAnAppWithAccess(t *testing.T) {
	svc, _ := accessService(dockerAPISetting("app1", `{"images":["alpine"]}`))
	spec := &swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{}}}
	assert.NoError(t, svc.ApplyToService(context.Background(), nil, "app1", spec))
	assert.Equal(t, []mount.Mount{dockerapiservice.SocketMount("app1")}, spec.TaskTemplate.ContainerSpec.Mounts)
	assert.Equal(t, []swarm.NetworkAttachmentConfig{{Target: "id-hp-dapi-app1"}}, spec.TaskTemplate.Networks)
}

func TestApplyToServiceDetachesAnAppWithout(t *testing.T) {
	svc, networks := accessService()
	networks.byName["hp-dapi-app1"] = "id-hp-dapi-app1"
	spec := &swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{Mounts: []mount.Mount{dockerapiservice.SocketMount("app1")}},
		Networks:      []swarm.NetworkAttachmentConfig{{Target: "id-hp-dapi-app1"}},
	}}
	assert.NoError(t, svc.ApplyToService(context.Background(), nil, "app1", spec))
	assert.Empty(t, spec.TaskTemplate.ContainerSpec.Mounts)
	assert.Empty(t, spec.TaskTemplate.Networks)
}

// fakeCluster is a cluster's nodes and networks at once, for a call that needs
// both.
type fakeCluster struct {
	fakeNetworks
	nodes []swarm.Node
}

func (f *fakeCluster) NodeList(_ context.Context, _ ...docker.NodeListOption) (*client.NodeListResult, error) {
	return &client.NodeListResult{Items: f.nodes}, nil
}

func TestRemoveAppRemovesItsNetworkAndAsksTheAgents(t *testing.T) {
	var calls []string
	svc := fanOutService(&calls, "")
	cluster := &fakeCluster{fakeNetworks: fakeNetworks{byName: map[string]string{}},
		nodes: svc.dockerManager.(*fakeNodes).nodes}
	svc.dockerManager = cluster
	networks := &cluster.fakeNetworks

	assert.NoError(t, svc.RemoveApp(context.Background(), "app1"))
	assert.Equal(t, []string{"remove app1 agent-on-n1", "remove app1 agent-on-n2"}, calls)
	assert.Equal(t, []string{"hp-dapi-app1"}, networks.removed)
}
```

In `hivepaas_app/service/envvarservice/envvarserviceimpl/app_system_env_test.go`, add:

```go
func TestDockerAPIEnvVarsNameTheSocketOnlyForAnAppWithAccess(t *testing.T) {
	assert.Empty(t, dockerAPIEnvVars(false))
	envs := dockerAPIEnvVars(true)
	if assert.Len(t, envs, 1) {
		assert.Equal(t, base.AppSystemEnvVarDockerHost, envs[0].Key)
		assert.Equal(t, "unix:///var/run/hivepaas/docker.sock", envs[0].Value)
		assert.False(t, envs[0].IsShared, "another app has no use for this app's socket")
	}
}
```

- [ ] **Step 3: Run them to see them fail**

Run: `go test ./hivepaas_app/service/dockerapiservice/... ./hivepaas_app/service/envvarservice/...`
Expected: FAIL, `undefined: SocketMountTarget`, `undefined: dockerAPIEnvVars`.

- [ ] **Step 4: Write the helpers and the methods**

Add to the `const` block of `hivepaas_app/service/dockerapiservice/types.go`:

```go
	// SocketMountTarget is where an app's socket volume is mounted: the directory
	// of dockerproxy.SocketPath.
	SocketMountTarget = "/var/run/hivepaas"
```

`hivepaas_app/service/dockerapiservice/attach.go`:

```go
package dockerapiservice

import (
	"slices"
	"strings"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
)

// SocketMount is the mount that puts an app's socket where the app looks for it.
func SocketMount(appID string) mount.Mount {
	return mount.Mount{Type: mount.TypeVolume, Source: SocketVolumeName(appID), Target: SocketMountTarget}
}

// IsSocketMount reports a mount of any app's socket volume.
func IsSocketMount(m *mount.Mount) bool {
	return m.Type == mount.TypeVolume && strings.HasPrefix(m.Source, SocketVolumePrefix)
}

// IsAppNetworkName reports the name of any app's own network.
func IsAppNetworkName(name string) bool {
	return strings.HasPrefix(name, NetworkPrefix)
}

func isSocketMount(m mount.Mount) bool {
	return IsSocketMount(&m)
}

// Attach gives a service spec an app's socket and its network, once each. A
// socket of another app - the one a clone was copied from - is replaced.
func Attach(spec *swarm.ServiceSpec, appID, networkID string) {
	container := spec.TaskTemplate.ContainerSpec
	if container == nil {
		container = &swarm.ContainerSpec{}
		spec.TaskTemplate.ContainerSpec = container
	}
	container.Mounts = append(slices.DeleteFunc(container.Mounts, isSocketMount), SocketMount(appID))
	attached := slices.ContainsFunc(spec.TaskTemplate.Networks, func(n swarm.NetworkAttachmentConfig) bool {
		return n.Target == networkID
	})
	if !attached {
		spec.TaskTemplate.Networks = append(spec.TaskTemplate.Networks, swarm.NetworkAttachmentConfig{Target: networkID})
	}
}

// Detach takes every socket mount off a spec, and the attachment to networkID
// when there is one.
func Detach(spec *swarm.ServiceSpec, networkID string) {
	if container := spec.TaskTemplate.ContainerSpec; container != nil {
		container.Mounts = slices.DeleteFunc(container.Mounts, isSocketMount)
	}
	if networkID == "" {
		return
	}
	spec.TaskTemplate.Networks = slices.DeleteFunc(spec.TaskTemplate.Networks,
		func(n swarm.NetworkAttachmentConfig) bool { return n.Target == networkID })
}

// KeepSocketMounts is what a screen that rewrites an app's mounts saves: the
// mounts it built, and the socket the service already has. The storage screen
// does not show the socket, since it is given with the app's access rather than
// chosen there, so what it sends back never has it.
func KeepSocketMounts(final, current []mount.Mount) []mount.Mount {
	kept := slices.DeleteFunc(slices.Clone(final), isSocketMount)
	for _, m := range current {
		if isSocketMount(m) {
			kept = append(kept, m)
		}
	}
	return kept
}

// KeepAppNetwork is the same for the network screen and the app's own network,
// networkID, when the service is attached to it.
func KeepAppNetwork(final, current []swarm.NetworkAttachmentConfig,
	networkID string) []swarm.NetworkAttachmentConfig {
	if networkID == "" {
		return final
	}
	for _, n := range current {
		if n.Target != networkID {
			continue
		}
		kept := slices.DeleteFunc(slices.Clone(final), func(f swarm.NetworkAttachmentConfig) bool {
			return f.Target == networkID
		})
		return append(kept, n)
	}
	return final
}
```

Add to the `Service` interface in `hivepaas_app/service/dockerapiservice/service.go`, importing `swarm` and `entity`:

```go
	// AccessOf is an app's Docker API setting, or nil when it has none.
	AccessOf(ctx context.Context, db database.IDB, appID string) (*entity.AppDockerAPISettings, error)
	// EnsureNetwork returns the id of the app's own network, creating it first
	// when it does not exist.
	EnsureNetwork(ctx context.Context, appID string) (string, error)
	// AppNetworkID is the id of the app's own network, "" when it has none.
	AppNetworkID(ctx context.Context, appID string) (string, error)
	// ApplyToService gives an app's service spec what its access needs - its
	// socket and its network - or takes them away when it has none.
	ApplyToService(ctx context.Context, db database.IDB, appID string, spec *swarm.ServiceSpec) error
	// DetachFromService takes an app's socket and network off a spec: a clone's,
	// which starts as a copy of that app's.
	DetachFromService(ctx context.Context, appID string, spec *swarm.ServiceSpec) error
	// RemoveApp removes what an app's access made: its children and socket
	// volumes on every node, and its network. Its setting goes with the app's.
	RemoveApp(ctx context.Context, appID string) error
```

`hivepaas_app/service/dockerapiservice/dockerapiserviceimpl/access.go`:

```go
package dockerapiserviceimpl

import (
	"context"
	"errors"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func (s *service) AccessOf(ctx context.Context, db database.IDB, appID string) (
	*entity.AppDockerAPISettings, error) {
	settings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDockerAPI),
		bunex.SelectWhere("setting.object_id = ?", appID),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(settings) == 0 {
		return nil, nil
	}
	access, err := settings[0].AsAppDockerAPISettings()
	return access, hperrors.Wrap(err)
}

func (s *service) AppNetworkID(ctx context.Context, appID string) (string, error) {
	inspect, err := s.dockerManager.NetworkInspect(ctx, dockerapiservice.NetworkName(appID))
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return "", nil
		}
		return "", hperrors.Wrap(err)
	}
	return inspect.Network.ID, nil
}

// EnsureNetwork makes the app's own network an attachable overlay: the app's
// task is a swarm service and joins it that way, and its children are plain
// containers on whichever node the task runs.
func (s *service) EnsureNetwork(ctx context.Context, appID string) (string, error) {
	id, err := s.AppNetworkID(ctx, appID)
	if err != nil || id != "" {
		return id, err
	}
	resp, err := s.dockerManager.NetworkCreate(ctx, dockerapiservice.NetworkName(appID),
		func(opts *client.NetworkCreateOptions) {
			opts.Driver = docker.NetworkDriverOverlay
			opts.Scope = docker.NetworkScopeSwarm
			opts.Attachable = true
			opts.Options = map[string]string{docker.NetworkOptionDriverMTU: docker.DefaultOverlayNetworkMTU}
			opts.Labels = map[string]string{dockerapiservice.NetworkLabel: appID}
		})
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return resp.ID, nil
}

func (s *service) ApplyToService(ctx context.Context, db database.IDB, appID string,
	spec *swarm.ServiceSpec) error {
	access, err := s.AccessOf(ctx, db, appID)
	if err != nil {
		return err
	}
	if access == nil {
		return s.DetachFromService(ctx, appID, spec)
	}
	networkID, err := s.EnsureNetwork(ctx, appID)
	if err != nil {
		return err
	}
	dockerapiservice.Attach(spec, appID, networkID)
	return nil
}

func (s *service) DetachFromService(ctx context.Context, appID string, spec *swarm.ServiceSpec) error {
	networkID, err := s.AppNetworkID(ctx, appID)
	if err != nil {
		return err
	}
	dockerapiservice.Detach(spec, networkID)
	return nil
}

// RemoveApp does what it can. It runs once the app's service is removed, and
// the agents' sweep and a later delete of the network take care of what a node
// that is down, or a task still shutting down, keeps for now.
func (s *service) RemoveApp(ctx context.Context, appID string) error {
	errs := []error{s.RemoveAppObjects(ctx, appID)}
	_, err := s.dockerManager.NetworkRemove(ctx, dockerapiservice.NetworkName(appID))
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		errs = append(errs, hperrors.Wrap(err))
	}
	return errors.Join(errs...)
}
```

In `hivepaas_app/base/env_var.go`, after `AppSystemEnvVarID`:

```go
	// AppSystemEnvVarDockerHost is where an app given the Docker API reaches it,
	// in the form DOCKER_HOST takes. A template sets DOCKER_HOST, or whatever
	// variable its app reads, to ${HIVEPAAS_DOCKER_HOST}.
	AppSystemEnvVarDockerHost = "HIVEPAAS_DOCKER_HOST"
```

In `hivepaas_app/service/envvarservice/envvarserviceimpl/app_system_env.go`:
- add `base.SettingTypeAppDockerAPI` to the `setting.type IN (?)` list;
- after `result = append(result, kindEnvs...)`, add:

```go
	result = append(result, dockerAPIEnvVars(
		settinghelper.FindSettingByType(settings, base.SettingTypeAppDockerAPI) != nil)...)
```

- add the function at the end of the file, importing `pkg/dockerproxy`:

```go
// dockerAPIEnvVars is where an app given the Docker API finds it. It is not
// shared: another app has no use for this app's socket, and no way to reach it.
func dockerAPIEnvVars(hasAccess bool) []*envvarservice.EnvVar {
	if !hasAccess {
		return nil
	}
	return []*envvarservice.EnvVar{{EnvVar: &entity.EnvVar{
		Key:   base.AppSystemEnvVarDockerHost,
		Value: "unix://" + dockerproxy.SocketPath,
	}}}
}
```

- [ ] **Step 5: Run the tests, then lint**

Run: `go test ./hivepaas_app/service/dockerapiservice/... ./hivepaas_app/service/envvarservice/... && golangci-lint run ./hivepaas_app/service/dockerapiservice/... ./hivepaas_app/service/envvarservice/... ./hivepaas_app/base/...`
Expected: `ok`, `0 issues`.

- [ ] **Step 6: Commit**

```bash
git add hivepaas_app
git commit -m "feat(dockerapi): give a service spec an app's socket and network, and name the socket

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 2: The `settings.dockerApi` block

**Files:**
- Modify: `hivepaas_app/entity/setting_app_docker_api.go` (limits in the units a person writes)
- Modify: `hivepaas_app/entity/setting_app_docker_api_test.go`
- Modify: `hivepaas_app/service/dockerapiservice/dockerapiserviceimpl/policies.go`, `policies_test.go`
- Create: `hivepaas_app/service/specservice/specmodel/docker_api.go`
- Test: `hivepaas_app/service/specservice/specmodel/docker_api_test.go`
- Modify: `hivepaas_app/service/specservice/specmodel/buildable.go`, `buildable_test.go`
- Modify: `hivepaas_app/service/specservice/specserviceimpl/build.go`, `build_settings.go`, `service.go`, `build_test.go`

**Interfaces:**
- Consumes: Task 1's `EnsureNetwork`, `Attach`.
- Produces:
  - `entity.AppDockerAPILimits{Containers int; Memory unit.DataSize; CPUs float64}`;
  - `specmodel.BlockSettingsDockerAPI`, `specmodel.DockerAPIIn(settings map[string]any) (*entity.AppDockerAPISettings, error)`, `specmodel.DockerAPIProblem(*entity.AppDockerAPISettings) string`;
  - the builder `buildDockerAPI`;
  - `specserviceimpl.New` taking a `dockerapiservice.Service`.

- [ ] **Step 1: Write the failing tests**

Change `hivepaas_app/entity/setting_app_docker_api_test.go` so that the stored limits read `"limits":{"containers":3,"memory":"2gb","cpus":2}` and the expected `Limits` is `AppDockerAPILimits{Containers: 3, Memory: 2 * unit.GB, CPUs: 2}`, importing `pkg/unit`.

In `hivepaas_app/service/dockerapiservice/dockerapiserviceimpl/policies_test.go`, change the autobase setting's limits to `"limits":{"containers":3,"memory":"2gb","cpus":0.5}` and its expected `Limits` to `dockerproxy.Limits{Containers: 3, Memory: 2 << 30, NanoCPUs: 500_000_000}`.

`hivepaas_app/service/specservice/specmodel/docker_api_test.go`:

```go
package specmodel

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

func TestDockerAPIInReadsTheBlock(t *testing.T) {
	doc := decodeDoc(t, `
settings:
  dockerApi:
    images: [autobase/automation]
    sharedDirs: [/var/lib/autobase/ansible]
    limits: {containers: 3, memory: 2gb, cpus: 1.5}
`)
	got, err := DockerAPIIn(doc.Settings)
	assert.NoError(t, err)
	assert.Equal(t, &entity.AppDockerAPISettings{
		Images: []string{"autobase/automation"}, SharedDirs: []string{"/var/lib/autobase/ansible"},
		Limits: entity.AppDockerAPILimits{Containers: 3, Memory: 2 * unit.GB, CPUs: 1.5},
	}, got)

	got, err = DockerAPIIn(decodeDoc(t, "settings:\n  routing: {port: 80}\n").Settings)
	assert.NoError(t, err)
	assert.Nil(t, got)

	_, err = DockerAPIIn(decodeDoc(t, "settings:\n  dockerApi: {images: [a], privileged: true}\n").Settings)
	assert.ErrorIs(t, err, hperrors.ErrSpecBlockInvalid)
}

func TestDockerAPIProblem(t *testing.T) {
	ok := func() *entity.AppDockerAPISettings {
		return &entity.AppDockerAPISettings{Images: []string{"alpine"}}
	}
	assert.Empty(t, DockerAPIProblem(ok()))

	tooMany := make([]string, MaxDockerAPIImages+1)
	for i := range tooMany {
		tooMany[i] = "alpine"
	}
	cases := map[string]func(s *entity.AppDockerAPISettings){
		"images: at least one":       func(s *entity.AppDockerAPISettings) { s.Images = nil },
		"images: at most":            func(s *entity.AppDockerAPISettings) { s.Images = tooMany },
		"images[0]":                  func(s *entity.AppDockerAPISettings) { s.Images = []string{"alp ine"} },
		"sharedDirs: at most":        func(s *entity.AppDockerAPISettings) { s.SharedDirs = []string{"/a", "/b", "/c", "/d", "/e", "/f"} },
		"sharedDirs[0]":              func(s *entity.AppDockerAPISettings) { s.SharedDirs = []string{"data"} },
		"sharedDirs[0]: / is":        func(s *entity.AppDockerAPISettings) { s.SharedDirs = []string{"/"} },
		"sharedDirs[0]: /a/../b is":  func(s *entity.AppDockerAPISettings) { s.SharedDirs = []string{"/a/../b"} },
		"networks[0]":                func(s *entity.AppDockerAPISettings) { s.Networks = []string{"hivepaas_net"} },
		"allow[0]":                   func(s *entity.AppDockerAPISettings) { s.Allow = []string{"build"} },
		"limits.containers":          func(s *entity.AppDockerAPISettings) { s.Limits.Containers = 51 },
		"limits.memory":              func(s *entity.AppDockerAPISettings) { s.Limits.Memory = 1 * unit.MB },
		"limits.cpus":                func(s *entity.AppDockerAPISettings) { s.Limits.CPUs = -1 },
	}
	for want, change := range cases {
		settings := ok()
		change(settings)
		problem := DockerAPIProblem(settings)
		assert.True(t, strings.Contains(problem, want), "%s: %q", want, problem)
	}
}

func TestCheckBuildableTakesTheDockerAPIBlock(t *testing.T) {
	doc := decodeDoc(t, `
deployment:
  storage:
    mounts:
      /var/lib/autobase: {type: volume, source: vol-1}
settings:
  dockerApi: {images: [autobase/automation], sharedDirs: [/var/lib/autobase/ansible]}
`)
	assert.NoError(t, CheckBuildable(doc))

	doc.Settings["dockerApi"] = map[string]any{"images": []any{"alpine"}, "sharedDirs": []any{"/srv/work"}}
	err := CheckBuildable(doc)
	assert.ErrorIs(t, err, hperrors.ErrSpecBlockInvalid)
	assert.ErrorContains(t, err, "/srv/work is on none of the app's storage mounts")

	doc.Settings["dockerApi"] = map[string]any{"images": []any{}}
	assert.ErrorIs(t, CheckBuildable(doc), hperrors.ErrSpecBlockInvalid)
}
```

In `hivepaas_app/service/specservice/specmodel/buildable_test.go`, add a `dockerApi` block to `buildableDocYAML` under `settings:`, after `routing`:

```yaml
  dockerApi:
    images: [autobase/automation]
    sharedDirs: [/var/lib/postgresql/data/logs]
```

In `hivepaas_app/service/specservice/specserviceimpl/build_test.go`:
- add the same `dockerApi` block to `buildDocYAML`, with `sharedDirs: [/var/lib/postgresql/data]`;
- in `TestBuildAppBuildsEverySupportedBlock`, give the service a fake Docker API service, `svc := &service{volumeService: volumes, dockerAPIService: &fakeDockerAPIService{}}`, and expect five settings (`assert.Len(t, byType, 5)`);
- add this fake and a test to the file:

```go
type fakeDockerAPIService struct {
	dockerapiservice.Service
}

func (fakeDockerAPIService) EnsureNetwork(_ context.Context, appID string) (string, error) {
	return "net-" + appID, nil
}

func TestBuildDockerAPIGivesTheAppItsSocketAndNetwork(t *testing.T) {
	useDataKey(t)
	svc := &service{volumeService: &fakeBuildVolumeService{}, dockerAPIService: &fakeDockerAPIService{}}
	req := buildReq(t, buildDocYAML)

	resp, err := svc.BuildApp(context.Background(), nil, req)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	setting := settinghelper.FindSettingByType(resp.Settings, base.SettingTypeAppDockerAPI)
	if assert.NotNil(t, setting) {
		assert.Equal(t, []string{"autobase/automation"}, setting.MustAsAppDockerAPISettings().Images)
	}
	assert.Contains(t, req.Spec.TaskTemplate.ContainerSpec.Mounts, dockerapiservice.SocketMount("app-1"))
	assert.Contains(t, req.Spec.TaskTemplate.Networks, swarm.NetworkAttachmentConfig{Target: "net-app-1"})
}
```

- [ ] **Step 2: Run them to see them fail**

Run: `go test ./hivepaas_app/entity/ ./hivepaas_app/service/specservice/... ./hivepaas_app/service/dockerapiservice/...`
Expected: FAIL: the entity has no `CPUs`, `DockerAPIIn` is undefined, and the builder is missing.

- [ ] **Step 3: Write it**

In `hivepaas_app/entity/setting_app_docker_api.go`, import `pkg/unit`, and make the limits read the way `deployment.resources.limits` does:

```go
// AppDockerAPILimits bound what the app's children use, in the units the
// deployment's own resource limits take. Zero is the default.
type AppDockerAPILimits struct {
	// Containers is how many children may exist at once.
	Containers int `json:"containers,omitempty"`
	// Memory is the most one child may have.
	Memory unit.DataSize `json:"memory,omitempty"`
	// CPUs is the most processor time one child may have.
	CPUs float64 `json:"cpus,omitempty"`
}
```

In `hivepaas_app/service/dockerapiservice/dockerapiserviceimpl/policies.go`, the limits become:

```go
		Limits: dockerproxy.Limits{
			Containers: gofn.Coalesce(data.Limits.Containers, dockerapiservice.DefaultContainers),
			Memory:     gofn.Coalesce(data.Limits.Memory.Bytes(), dockerapiservice.DefaultMemory),
			NanoCPUs:   gofn.Coalesce(int64(data.Limits.CPUs*nanoPerCPU), dockerapiservice.DefaultNanoCPUs),
		},
```

with `const nanoPerCPU = 1_000_000_000` in that file.

`hivepaas_app/service/specservice/specmodel/docker_api.go`:

```go
package specmodel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

const (
	// MaxDockerAPIImages is how many image patterns a document may give an app's
	// children. A job runner takes "*"; an app that runs a few of its own names
	// them.
	MaxDockerAPIImages = 20
	// MaxDockerAPISharedDirs is how many of its directories an app may share.
	MaxDockerAPISharedDirs = 5
	// MaxDockerAPIContainers is the most children one app may keep at once.
	MaxDockerAPIContainers = 50
	// minDockerAPIMemory is the least memory docker starts a container with.
	minDockerAPIMemory = 6 * unit.MB

	dockerAPIPrefix = "settings.dockerApi."
)

var (
	dockerAPINetworks = []string{entity.DockerAPINetworkEnv}
	dockerAPIGroups   = []string{string(dockerproxy.GroupExec), string(dockerproxy.GroupFiles),
		string(dockerproxy.GroupVolumes), string(dockerproxy.GroupNetworks), string(dockerproxy.GroupNestedSocket)}
)

// DockerAPIIn reads the dockerApi block of a document's settings, and is nil
// when there is none. A field the setting does not have is refused rather than
// dropped: it would be a grant that is not applied.
func DockerAPIIn(settings map[string]any) (*entity.AppDockerAPISettings, error) {
	body, found := settings[SingletonBlockName(base.SettingTypeAppDockerAPI)]
	if !found {
		return nil, nil
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, invalid("%s%s", dockerAPIPrefix, err.Error())
	}
	out := &entity.AppDockerAPISettings{}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(out); err != nil {
		return nil, invalid("%s%s", dockerAPIPrefix, err.Error())
	}
	return out, nil
}

// DockerAPIProblem says what is wrong with a Docker API block, and is empty when
// nothing is.
func DockerAPIProblem(s *entity.AppDockerAPISettings) string {
	if s == nil {
		return ""
	}
	if problem := dockerAPICountProblem(s); problem != "" {
		return problem
	}
	for i, image := range s.Images {
		if strings.TrimSpace(image) == "" || strings.ContainsAny(image, " \t\n") {
			return fmt.Sprintf("%simages[%d]: %q is not an image", dockerAPIPrefix, i, image)
		}
	}
	for i, dir := range s.SharedDirs {
		if !path.IsAbs(dir) || path.Clean(dir) != dir || dir == "/" {
			return fmt.Sprintf("%ssharedDirs[%d]: %s is not an absolute path to a directory below /",
				dockerAPIPrefix, i, dir)
		}
	}
	for i, network := range s.Networks {
		if !slices.Contains(dockerAPINetworks, network) {
			return fmt.Sprintf("%snetworks[%d]: %q is not one of %v", dockerAPIPrefix, i, network, dockerAPINetworks)
		}
	}
	for i, group := range s.Allow {
		if !slices.Contains(dockerAPIGroups, group) {
			return fmt.Sprintf("%sallow[%d]: %q is not one of %v", dockerAPIPrefix, i, group, dockerAPIGroups)
		}
	}
	return dockerAPILimitsProblem(s.Limits)
}

func dockerAPICountProblem(s *entity.AppDockerAPISettings) string {
	switch {
	case len(s.Images) == 0:
		return dockerAPIPrefix + "images: at least one image is needed"
	case len(s.Images) > MaxDockerAPIImages:
		return fmt.Sprintf("%simages: at most %d, and this has %d", dockerAPIPrefix, MaxDockerAPIImages, len(s.Images))
	case len(s.SharedDirs) > MaxDockerAPISharedDirs:
		return fmt.Sprintf("%ssharedDirs: at most %d, and this has %d",
			dockerAPIPrefix, MaxDockerAPISharedDirs, len(s.SharedDirs))
	}
	return ""
}

func dockerAPILimitsProblem(limits entity.AppDockerAPILimits) string {
	switch {
	case limits.Containers < 0 || limits.Containers > MaxDockerAPIContainers:
		return fmt.Sprintf("%slimits.containers: %d is not between 0 and %d",
			dockerAPIPrefix, limits.Containers, MaxDockerAPIContainers)
	case limits.Memory != 0 && limits.Memory < minDockerAPIMemory:
		return fmt.Sprintf("%slimits.memory: %s is less than the %s docker starts a container with",
			dockerAPIPrefix, limits.Memory, minDockerAPIMemory)
	case limits.CPUs < 0:
		return fmt.Sprintf("%slimits.cpus: %v is below zero", dockerAPIPrefix, limits.CPUs)
	}
	return ""
}

// checkDockerAPI refuses a block that is wrong in itself, and a shared
// directory the app does not keep on its own storage: what a child is given is
// the directory the app mounts there, so there has to be one.
func checkDockerAPI(doc *AppDoc) error {
	settings, err := DockerAPIIn(doc.Settings)
	if err != nil || settings == nil {
		return err
	}
	if problem := DockerAPIProblem(settings); problem != "" {
		return invalid("%s", problem)
	}
	var targets []string
	if doc.Deployment != nil && doc.Deployment.Storage != nil {
		targets = slices.Collect(maps.Keys(doc.Deployment.Storage.Mounts))
	}
	for _, dir := range settings.SharedDirs {
		covered := slices.ContainsFunc(targets, func(target string) bool {
			return dir == target || strings.HasPrefix(dir, strings.TrimSuffix(target, "/")+"/")
		})
		if !covered {
			return invalid("%ssharedDirs: %s is on none of the app's storage mounts", dockerAPIPrefix, dir)
		}
	}
	return nil
}
```

In `hivepaas_app/service/specservice/specmodel/buildable.go`:
- add `BlockSettingsDockerAPI    Block = "settings.dockerApi"` to the block constants, after `BlockSettingsRouting`;
- append `BlockSettingsDockerAPI` to `BuildableBlocks`;
- add `{base.SettingTypeAppDockerAPI, BlockSettingsDockerAPI},` to the singleton list in `presentSettingsBlocks`, after routing;
- in `checkSettings`, name the block `dockerAPI := SingletonBlockName(base.SettingTypeAppDockerAPI)` and let it through with the others that need no check of their own (`case kind, envVars, dockerAPI:`), since `checkDockerAPI` reads the whole document;
- make `CheckBuildable` end with:

```go
	if err := checkSettings(doc.Settings); err != nil {
		return err
	}
	return checkDockerAPI(doc)
```

In `hivepaas_app/service/specservice/specserviceimpl/service.go`:
- add a field `dockerAPIService dockerapiservice.Service`;
- add a `New` parameter `dockerAPIService dockerapiservice.Service` after `domainService`;
- set the field.

In `hivepaas_app/service/specservice/specserviceimpl/build.go`, register the builder after routing:

```go
		specmodel.BlockSettingsDockerAPI:    s.buildDockerAPI,
```

Add to `hivepaas_app/service/specservice/specserviceimpl/build_settings.go`, importing `dockerapiservice`:

```go
// buildDockerAPI gives the app the Docker API: the setting that says what it may
// do, and on its service, the socket and the network its children join. The
// network is created here because the service is: a spec cannot name one that
// does not exist.
func (s *service) buildDockerAPI(ctx context.Context, state *buildState) error {
	block := specmodel.BlockSettingsDockerAPI
	settings := &entity.AppDockerAPISettings{}
	body := state.req.Doc.Settings[specmodel.SingletonBlockName(base.SettingTypeAppDockerAPI)]
	if err := decodeBlock(block, body, settings); err != nil {
		return err
	}
	networkID, err := s.dockerAPIService.EnsureNetwork(ctx, state.req.App.ID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	dockerapiservice.Attach(state.req.Spec, state.req.App.ID, networkID)
	return state.addSetting(base.SettingTypeAppDockerAPI, entity.CurrentAppDockerAPIVersion, false, settings)
}
```

- [ ] **Step 4: Run the tests, then lint**

Run: `gofmt -w hivepaas_app && go build ./... && go test ./hivepaas_app/entity/ ./hivepaas_app/service/specservice/... ./hivepaas_app/service/dockerapiservice/... ./hivepaas_app/cmd/... && golangci-lint run ./hivepaas_app/...`
Expected: `ok` everywhere, including `TestBuilderRegistryCoversEveryBuildableBlock` and the fx wiring tests, and `0 issues`.

- [ ] **Step 5: Commit**

```bash
git add hivepaas_app
git commit -m "feat(spec): a settings.dockerApi block, built into the setting and the app's socket and network

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 3: Templates give it, behind Write on the Cluster module

**Files:**
- Create: `hivepaas_app/service/apptemplateservice/templatemodel/docker_api.go`
- Test: `hivepaas_app/service/apptemplateservice/templatemodel/docker_api_test.go`
- Modify: `hivepaas_app/service/apptemplateservice/templatemodel/validate.go`
- Create: `hivepaas_app/usecase/apptemplateuc/app_create_docker_api.go`
- Test: `hivepaas_app/usecase/apptemplateuc/app_create_docker_api_test.go`
- Modify: `hivepaas_app/usecase/apptemplateuc/app_create_from_template.go`, `app_preflight.go`
- Modify: `hivepaas_app/usecase/apptemplateuc/apptemplatedto/template_get.go`
- Regenerate: `docs/openapi/swagger.json`

**Interfaces:**
- Consumes: `specmodel.DockerAPIIn`, `specmodel.DockerAPIProblem`.
- Produces: `(*templatemodel.Template).DockerAPI() (*entity.AppDockerAPISettings, error)`, `(*Template).RequiresDockerAPI() bool`, `(*apptemplateuc.UC).checkDockerAPI`, `AppTemplateDockerAPIResp` on the template detail (`dockerApi`, for the app and each dependency).

- [ ] **Step 1: Write the failing tests**

`hivepaas_app/service/apptemplateservice/templatemodel/docker_api_test.go`:

```go
package templatemodel

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const dockerAPITemplateYAML = `
apiVersion: hivepaas.com/v1
kind: AppTemplate
metadata:
  name: autobase
  title: Autobase
  tagline: PostgreSQL clusters
  description: Test.
  categories: [databases/sql]
  icon: icons/autobase.svg
  requires: {versionCode: v000001}
versions:
  - {name: "2.11", release: "2.11.0", default: true, image: "autobase/console:2.11.0"}
app:
  deployment:
    source: {activeMethod: image, imageSource: {image: "${{ image }}"}}
    storage:
      mounts:
        /var/lib/autobase: {type: volume, source: "${{ params.dataVolume }}"}
  settings:
    dockerApi:
      images: [autobase/automation]
      sharedDirs: [/var/lib/autobase/ansible]
parameters:
  - {name: dataVolume, title: Data volume, type: volume}
`

func TestDockerAPIIsReadFromTheTemplate(t *testing.T) {
	tmpl, err := DecodeTemplate([]byte(dockerAPITemplateYAML))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	access, err := tmpl.DockerAPI()
	assert.NoError(t, err)
	if assert.NotNil(t, access) {
		assert.Equal(t, []string{"autobase/automation"}, access.Images)
	}
	assert.True(t, tmpl.RequiresDockerAPI())
}

func TestDockerAPIMayNotDependOnWhatAPersonFillsIn(t *testing.T) {
	tmpl, err := DecodeTemplate([]byte(strings.Replace(dockerAPITemplateYAML, "images: [autobase/automation]",
		`images: ["${{ params.image }}"]`, 1)))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.ErrorContains(t, tmpl.Validate("autobase.yaml"), "app.settings.dockerApi: a placeholder here")
}

func TestAVersionMayNotChangeTheDockerAPI(t *testing.T) {
	tmpl, err := DecodeTemplate([]byte(strings.Replace(dockerAPITemplateYAML,
		`default: true, image: "autobase/console:2.11.0"}`,
		`default: true, image: "autobase/console:2.11.0",`+
			` override: {app: {settings: {dockerApi: {images: ["*"]}}}}}`, 1)))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.ErrorContains(t, tmpl.Validate("autobase.yaml"), "the Docker API is declared once, in app")
}
```

`hivepaas_app/usecase/apptemplateuc/app_create_docker_api_test.go`:

```go
package apptemplateuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func dockerAPIApps(t *testing.T, withAccess bool) []*appToProvision {
	t.Helper()
	settings := map[string]any{"routing": map[string]any{"port": 80}}
	if withAccess {
		settings["dockerApi"] = map[string]any{"images": []any{"autobase/automation"}}
	}
	return []*appToProvision{{
		id: "app-1", name: "autobase",
		result: &templaterender.Result{Doc: &specmodel.AppDoc{Settings: settings}},
	}}
}

func TestCheckDockerAPINeedsWriteOnTheClusterModule(t *testing.T) {
	permissions := &fakePermissionManager{granted: false}
	uc := &UC{permissionManager: permissions}

	err := uc.checkDockerAPI(context.Background(), &basedto.Auth{}, dockerAPIApps(t, true))
	assert.ErrorIs(t, err, hperrors.ErrUnauthorized)
	assert.ErrorContains(t, err, "autobase")
	if assert.Len(t, permissions.checked, 1) {
		check, ok := permissions.checked[0].(*permission.ModuleAccessCheck)
		if assert.True(t, ok) {
			assert.Equal(t, base.ResourceModuleCluster, check.Module)
			assert.Equal(t, base.ActionTypeWrite, check.Action)
		}
	}

	permissions.granted = true
	assert.NoError(t, uc.checkDockerAPI(context.Background(), &basedto.Auth{}, dockerAPIApps(t, true)))
}

func TestCheckDockerAPIAsksNothingOfATemplateWithout(t *testing.T) {
	permissions := &fakePermissionManager{granted: false}
	uc := &UC{permissionManager: permissions}
	assert.NoError(t, uc.checkDockerAPI(context.Background(), &basedto.Auth{}, dockerAPIApps(t, false)))
	assert.Empty(t, permissions.checked)
}
```

- [ ] **Step 2: Run them to see them fail**

Run: `go test ./hivepaas_app/service/apptemplateservice/templatemodel/ ./hivepaas_app/usecase/apptemplateuc/`
Expected: FAIL, `tmpl.DockerAPI undefined`, `uc.checkDockerAPI undefined`.

- [ ] **Step 3: Write it**

`hivepaas_app/service/apptemplateservice/templatemodel/docker_api.go`:

```go
package templatemodel

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// dockerAPIPath is where a template gives its app the Docker API: the spec's own
// path, because a template's app block is a spec document.
var dockerAPIPath = []string{"settings", "dockerApi"}

// DockerAPI is the Docker API provisioning this template gives its app, and is
// nil for a template that gives none. Like capabilities, it is read from the
// template rather than from a render, and a version may not override it: what
// the store and the deploy dialog show is what is granted.
func (t *Template) DockerAPI() (*entity.AppDockerAPISettings, error) {
	access, problem := dockerAPIIn(t.App)
	if problem != "" {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail("%s", problem)
	}
	return access, nil
}

// RequiresDockerAPI reports whether any app of the template is given the Docker
// API. One whose block cannot be read counts, as for capabilities.
func (t *Template) RequiresDockerAPI() bool {
	for _, app := range t.appTrees() {
		access, problem := dockerAPIIn(app)
		if problem != "" || access != nil {
			return true
		}
	}
	return false
}

// dockerAPIIn reads the block out of an app tree, and says what is wrong with it
// instead when it cannot. Both are empty for a tree that gives none.
func dockerAPIIn(app map[string]any) (*entity.AppDockerAPISettings, string) {
	where := "app." + strings.Join(dockerAPIPath, ".")
	settings, _ := app[dockerAPIPath[0]].(map[string]any)
	body, found := settings[dockerAPIPath[1]]
	if !found {
		return nil, ""
	}
	encoded, err := yaml.Marshal(body)
	if err != nil {
		return nil, fmt.Sprintf("%s: %s", where, err.Error())
	}
	if strings.Contains(string(encoded), placeholderMark) {
		return nil, where + ": a placeholder here would make what is granted depend on what a person fills in"
	}
	access, err := specmodel.DockerAPIIn(settings)
	if err != nil {
		return nil, fmt.Sprintf("%s: %s", where, err.Error())
	}
	return access, ""
}

// validateDockerAPI checks the block a template declares, and refuses one a
// version declares: the Docker API an app is given is declared once.
func validateDockerAPI(t *Template, p *problems) {
	for i, app := range t.appTrees() {
		where := "app"
		if t.HasComponents() {
			where = fmt.Sprintf("components[%s].app", t.Components[i].Name)
		}
		access, problem := dockerAPIIn(app)
		if problem != "" {
			p.add("%s", problem)
		}
		if problem = specmodel.DockerAPIProblem(access); problem != "" {
			p.add("%s.%s", where, problem)
		}
	}
	for _, version := range t.Versions {
		if version == nil || version.Override == nil {
			continue
		}
		overridden, problem := dockerAPIIn(version.Override.App)
		if problem != "" || overridden != nil {
			p.add("versions[%s].override.app.%s: the Docker API is declared once, in app",
				version.Name, strings.Join(dockerAPIPath, "."))
		}
	}
}
```

In `hivepaas_app/service/apptemplateservice/templatemodel/validate.go`, call `validateDockerAPI(t, &p)` right after `validateCapabilities(t, &p)`.

`hivepaas_app/usecase/apptemplateuc/app_create_docker_api.go`:

```go
package apptemplateuc

import (
	"context"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// checkDockerAPI refuses a request whose apps would be given the Docker API,
// unless the caller may grant that. The gate is the one capabilities have, Write
// on the Cluster module: the app's children run on the cluster's nodes, as many
// as its limits allow, running what its images allow.
func (uc *UC) checkDockerAPI(ctx context.Context, auth *basedto.Auth, apps []*appToProvision) error {
	block := specmodel.SingletonBlockName(base.SettingTypeAppDockerAPI)
	var asking []string
	for _, target := range apps {
		if _, found := target.result.Doc.Settings[block]; found {
			asking = append(asking, target.name)
		}
	}
	if len(asking) == 0 {
		return nil
	}
	hasPerm, err := uc.permissionManager.CheckAccess(ctx, uc.db, auth, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleCluster,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	if hasPerm {
		return nil
	}
	return hperrors.Wrap(hperrors.ErrUnauthorized).
		WithExtraDetail("%s: giving an app the Docker API needs Write permission on the Cluster module",
			strings.Join(asking, ", ")).
		WithMsgLog("creating an app from a template with the Docker API requires Write on the Cluster module")
}
```

In `hivepaas_app/usecase/apptemplateuc/app_create_from_template.go`, after the `checkCapabilities` call:

```go
	if err = uc.checkDockerAPI(ctx, auth, apps); err != nil {
		return nil, hperrors.Wrap(err)
	}
```

In `hivepaas_app/usecase/apptemplateuc/app_preflight.go`, in `collectIssues`, after the capabilities entry:

```go
		func() error { return uc.checkDockerAPI(ctx, auth, apps) },
```

In `hivepaas_app/usecase/apptemplateuc/apptemplatedto/template_get.go`:
- add to the template response, after `Capabilities`:

```go
	// DockerAPI is the Docker API creating this template gives the app, and is
	// null for the templates that give none. Creating one needs Write on the
	// Cluster module, like capabilities.
	DockerAPI *AppTemplateDockerAPIResp `json:"dockerApi"`
```

- add the same field, with a comment naming the dependency's app, to `AppTemplateDependencyResp` after its `Capabilities`;
- add the type:

```go
// AppTemplateDockerAPIResp is the dockerApi block of the template, as the
// template wrote it: a version cannot override it.
type AppTemplateDockerAPIResp struct {
	Images     []string `json:"images"`
	SharedDirs []string `json:"sharedDirs,omitempty"`
	Networks   []string `json:"networks,omitempty"`
	Allow      []string `json:"allow,omitempty"`
	// Containers, Memory and CPUs are the limits of the app's children, zero
	// where the template leaves the default.
	Containers int     `json:"containers,omitempty"`
	Memory     string  `json:"memory,omitempty"`
	CPUs       float64 `json:"cpus,omitempty"`
}
```

- fill it with `resp.DockerAPI = transformDockerAPI(tmpl.Template)` next to `resp.Capabilities`, and `DockerAPI: transformDockerAPI(dep.Template),` in the dependency literal;
- add the transform after `transformCapabilities`:

```go
// transformDockerAPI reads the block out of the template, and leaves out one
// that cannot be read, for the reason transformCapabilities does.
func transformDockerAPI(tmpl *templatemodel.Template) *AppTemplateDockerAPIResp {
	access, err := tmpl.DockerAPI()
	if err != nil || access == nil {
		return nil
	}
	resp := &AppTemplateDockerAPIResp{
		Images: access.Images, SharedDirs: access.SharedDirs, Networks: access.Networks, Allow: access.Allow,
		Containers: access.Limits.Containers, CPUs: access.Limits.CPUs,
	}
	if access.Limits.Memory > 0 {
		resp.Memory = access.Limits.Memory.String()
	}
	return resp
}
```

Run `make gen-swag` and check that `docs/openapi/swagger.json` changes only by the new type and the two fields.

- [ ] **Step 4: Run the tests, then lint**

Run: `go test ./hivepaas_app/service/apptemplateservice/... ./hivepaas_app/usecase/apptemplateuc/... && golangci-lint run ./hivepaas_app/...`
Expected: `ok`, `0 issues`.

- [ ] **Step 5: Commit**

```bash
git add hivepaas_app docs/openapi/swagger.json
git commit -m "feat(templates): a template may give its app the Docker API, behind Write on the Cluster module

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 4: Deployments keep it, deletion removes it, a clone does not inherit it

**Files:**
- Create: `hivepaas_app/service/appdeploymentservice/appdeploymentserviceimpl/deployment_docker_api.go`
- Modify: `.../appdeploymentserviceimpl/image_deploy_apply_svc.go`, `repo_deploy_apply_svc.go`, `service.go`
- Modify: `hivepaas_app/service/appservice/appserviceimpl/deletion.go`, `service.go`
- Modify: `hivepaas_app/service/appcloneservice/appcloneserviceimpl/clone_3_swarm_service.go`, `service.go`

**Interfaces:**
- Consumes: `AccessOf`, `SyncAgents`, `ApplyToService`, `RemoveApp`, `DetachFromService`.

- [ ] **Step 1: Deployments**

Each of the three services gets a `dockerAPIService dockerapiservice.Service` field and `New` parameter, placed alphabetically among the other services it takes.

`hivepaas_app/service/appdeploymentservice/appdeploymentserviceimpl/deployment_docker_api.go`:

```go
package appdeploymentserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

// prepareDockerAPI has every node's agent serve the socket of an app given the
// Docker API before its new task looks for it. The agents would get there at
// their next tick; a task that starts first finds no socket, and many apps give
// up on that and restart. A failure here is the agents', not the deployment's:
// it is logged, and the deployment goes on.
func (s *service) prepareDockerAPI(ctx context.Context, db database.IDB, data *appDeploymentData) error {
	access, err := s.dockerAPIService.AccessOf(ctx, db, data.App.ID)
	if err != nil || access == nil {
		return hperrors.Wrap(err)
	}
	if err = s.dockerAPIService.SyncAgents(ctx); err != nil {
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
			"Not every node serves the app's Docker API yet: "+err.Error(), tasklog.TsNow))
	}
	return nil
}
```

In `imageDeployStepServiceApply` and in `repoDeployStepServiceApply`:
- before the `s.dockerManager.ServiceUpdateFunc(` call, add:

```go
	if err = s.prepareDockerAPI(ctx, db, data.appDeploymentData); err != nil {
		return hperrors.Wrap(err)
	}
```

- inside the update callback, before `placementReq.Service = svc`, add:

```go
			// The socket and the network follow the app's access on every
			// deployment, whatever a screen or an older release left on the service.
			if err := s.dockerAPIService.ApplyToService(ctx, db, data.App.ID, &svc.Spec); err != nil {
				return false, hperrors.Wrap(err)
			}
```

- [ ] **Step 2: Deletion and clones**

In `deleteAppInDocker` of `hivepaas_app/service/appservice/appserviceimpl/deletion.go`, add this with the other cleanups that are not fatal, after `deleteDockerSecretsAndConfigs`:

```go
	_ = s.dockerAPIService.RemoveApp(ctx, app.ID)
```

and extend the comment above them to say so: "Neither of these is fatal" becomes "None of these is fatal", and it adds a line that what the Docker API's removal leaves, the agents sweep.

In `hivepaas_app/service/appcloneservice/appcloneserviceimpl/clone_3_swarm_service.go`, after `destSvc.Spec.TaskTemplate.Networks = newNetAttachments`:

```go
	// A clone starts with a copy of the source app's service, which carries the
	// source app's socket and network. Access is given, not copied: the clone has
	// no app-docker-api setting, so it gets neither.
	if err = s.dockerAPIService.DetachFromService(ctx, srcApp.ID, &destSvc.Spec); err != nil {
		return hperrors.Wrap(err)
	}
```

- [ ] **Step 3: Build, test, lint**

Run: `go build ./... && go test ./hivepaas_app/service/appdeploymentservice/... ./hivepaas_app/service/appservice/... ./hivepaas_app/service/appcloneservice/... ./hivepaas_app/cmd/... && golangci-lint run ./hivepaas_app/...`
Expected: `ok`, `0 issues`. The fx wiring tests confirm that the three services resolve with the new dependency.

If a test in these packages builds one of the three services through `New`, pass `nil` for the new parameter where the test does not reach the Docker API. Where it does, give it a fake that embeds `dockerapiservice.Service`.

- [ ] **Step 4: Commit**

```bash
git add hivepaas_app
git commit -m "feat(dockerapi): deployments keep an app's Docker API, deletion removes it, clones start without it

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 5: The storage and network screens, and export

**Files:**
- Modify: `hivepaas_app/usecase/appsettingsuc/appsettingsdto/storage_settings_get.go`, `network_settings_get.go`
- Modify: `hivepaas_app/usecase/appsettingsuc/storage_settings_update.go`, `network_settings_update.go`, `uc.go`
- Modify: `hivepaas_app/service/specservice/specserviceimpl/storage_map.go`, `swarm_map.go`
- Test: `hivepaas_app/usecase/appsettingsuc/appsettingsdto/docker_api_hidden_test.go`
- Test: `hivepaas_app/service/specservice/specserviceimpl/export_docker_api_test.go`

**Interfaces:**
- Consumes: `IsSocketMount`, `IsAppNetworkName`, `KeepSocketMounts`, `KeepAppNetwork`, `AppNetworkID`.

- [ ] **Step 1: Write the failing tests**

`hivepaas_app/usecase/appsettingsuc/appsettingsdto/docker_api_hidden_test.go`:

```go
package appsettingsdto

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
)

func TestTheStorageScreenDoesNotShowTheSocket(t *testing.T) {
	data := mount.Mount{Type: mount.TypeVolume, Source: "hp-vol-data", Target: "/data"}
	mounts, err := TransformStorageMounts(&StorageSettingsTransformInput{
		Service: &swarm.Service{Spec: swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{Mounts: []mount.Mount{data, dockerapiservice.SocketMount("app1")}},
		}}},
		MountKeyCalculator: func(m *mount.Mount) string { return m.Target },
	})
	assert.NoError(t, err)
	if assert.Len(t, mounts, 1) {
		assert.Equal(t, "/data", mounts[0].Target)
	}
}

func TestTheNetworkScreenDoesNotShowTheAppsOwnNetwork(t *testing.T) {
	attachments, err := TransformNetworkAttachments(
		[]swarm.NetworkAttachmentConfig{{Target: "n-project"}, {Target: "n-dapi"}},
		&NetworkTransformationInput{DockerNetworks: map[string]*network.Summary{
			"n-project": {Network: network.Network{ID: "n-project", Name: "shop_prod_net"}},
			"n-dapi":    {Network: network.Network{ID: "n-dapi", Name: dockerapiservice.NetworkName("app1")}},
		}})
	assert.NoError(t, err)
	if assert.Len(t, attachments, 1) {
		assert.Equal(t, "shop_prod_net", attachments[0].Name)
	}
}
```

`hivepaas_app/service/specservice/specserviceimpl/export_docker_api_test.go`:

```go
package specserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// The socket and the network are this installation's, named after this app's
// id. Import gives an app with the setting its own.
func TestExportLeavesOutTheAppsSocketAndNetwork(t *testing.T) {
	task := &swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{Mounts: []mount.Mount{
			{Type: mount.TypeTmpfs, Target: "/scratch"},
			dockerapiservice.SocketMount("app1"),
		}},
		Networks: []swarm.NetworkAttachmentConfig{{Target: "n-project"}, {Target: "n-dapi"}},
	}

	storage, err := mapAppStorage(task, nil, func(string) (string, *specmodel.ExternalRef) { return "", nil })
	assert.NoError(t, err)
	if assert.NotNil(t, storage) {
		assert.Len(t, storage.DockerMounts, 1)
		assert.Contains(t, storage.DockerMounts, "/scratch")
	}

	networks := mapNetworks(task, nil, map[string]string{
		"n-project": "shop_prod_net", "n-dapi": dockerapiservice.NetworkName("app1"),
	})
	if assert.NotNil(t, networks) && assert.Len(t, networks.Attachments, 1) {
		assert.Equal(t, "shop_prod_net", networks.Attachments[0].Name)
	}
}
```

- [ ] **Step 2: Run them to see them fail**

Run: `go test ./hivepaas_app/usecase/appsettingsuc/appsettingsdto/ ./hivepaas_app/service/specservice/specserviceimpl/ -run 'Socket|OwnNetwork|LeavesOut'`
Expected: FAIL: the socket mount and the app's network are listed.

- [ ] **Step 3: Write it**

In `TransformStorageMounts` (`storage_settings_get.go`), skip the socket as the first thing in the loop:

```go
		// The Docker API socket is given with the app's access, not chosen here.
		if dockerapiservice.IsSocketMount(&mounts[i]) {
			continue
		}
```

In `TransformNetworkAttachments` (`network_settings_get.go`), after the name is resolved from `input.DockerNetworks`, skip the app's own network:

```go
		// The app's own network comes with its Docker API access, not from here.
		if dockerapiservice.IsAppNetworkName(itemResp.Name) {
			continue
		}
```

In `UpdateAppStorageSettings` (`storage_settings_update.go`), the final mounts keep the socket:

```go
		data.FinalMounts = dockerapiservice.KeepSocketMounts(built.Mounts,
			data.Service.Spec.TaskTemplate.ContainerSpec.Mounts)
```

In `network_settings_update.go`:
- `appsettingsuc.UC` gains a `dockerAPIService dockerapiservice.Service` field and a `New` parameter;
- at the end of `loadAppNetworkSettingsForUpdate`, before `return nil`, the final networks keep the app's own:

```go
	// The app's own network is not listed on the screen, so what it saves never
	// names it; the attachment the service has is kept.
	appNetworkID, err := uc.dockerAPIService.AppNetworkID(ctx, app.ID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.FinalNetworks = dockerapiservice.KeepAppNetwork(data.FinalNetworks, currNetworks, appNetworkID)
```

In `mapAppStorage` (`storage_map.go`), skip the socket right after the shared-memory mount:

```go
		// The socket is this installation's, named after this app's id; import
		// gives an app with the setting its own.
		if dockerapiservice.IsSocketMount(m) {
			continue
		}
```

In `mapNetworks` (`swarm_map.go`), skip the app's own network once `name` is resolved:

```go
		if dockerapiservice.IsAppNetworkName(name) {
			continue
		}
```

- [ ] **Step 4: Run the tests, then lint**

Run: `go build ./... && go test ./hivepaas_app/usecase/appsettingsuc/... ./hivepaas_app/service/specservice/... ./hivepaas_app/cmd/... && golangci-lint run ./hivepaas_app/...`
Expected: `ok`, `0 issues`.

- [ ] **Step 5: Commit**

```bash
git add hivepaas_app
git commit -m "feat(dockerapi): screens keep the socket and network they do not show, and export leaves them out

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 6: Gates and merge

- [ ] **Step 1: Whole-repo gates**

Run: `go build ./... && golangci-lint run ./... && go test ./... && make gen-swag && git status --short docs/openapi`
Expected: build succeeds, `0 issues`, every package `ok`, and `gen-swag` changes nothing beyond what Task 3 committed.

- [ ] **Step 2: Merge**

```bash
git checkout main
git merge --no-ff feat/docker-api-grant -m "Merge branch 'feat/docker-api-grant'

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
go build ./... && go test ./hivepaas_app/service/dockerapiservice/... ./hivepaas_app/service/specservice/...
git branch -d feat/docker-api-grant
```
