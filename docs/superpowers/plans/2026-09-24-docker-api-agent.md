# Docker API Access - Agent Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Each node's agent serves the Docker API of every app with access through the engine of plan 1. It keeps the sockets in step with the database, and removes what apps' children leave behind.

**Architecture:**
- **Setting type.** A new singleton setting type `app-docker-api` holds an app's policy.
- **Backend service.** `service/dockerapiservice` turns those settings into `dockerproxy.Policy` values, and fans a sync or a removal out to every node's agent.
- **Agent use case.** In the agent, `usecaseagent/dockerapiagentuc`:
  - keeps one unix socket per app, inside the app's socket volume, reached through the host's filesystem;
  - reconciles the sockets on start, every 30 seconds, and on a gRPC call;
  - sweeps children, networks and volumes by label every 10 minutes.

**Tech Stack:** Go 1.27, gRPC and protobuf (`protoc`, `protoc-gen-go`, `protoc-gen-go-grpc` are installed), fx, bun, `pkg/dockerproxy`, moby client v0.6.

**Spec:** `docs/superpowers/specs/2026-09-24-docker-api-access-design.md` (§1, §2, §6, §8).

## Global Constraints

- **Names:**
  - socket volume `hp-dapi-sock-<app id>`, labeled `hivepaas.docker-api.socket=<app id>`;
  - app network `hp-dapi-<app id>`, labeled `hivepaas.docker-api.network=<app id>`;
  - children's owner label `hivepaas.docker-api.app` (`dockerproxy.OwnerLabel`).
- **The socket inside an app:** `/var/run/hivepaas/docker.sock` (`dockerproxy.SocketPath`). In its volume the file is `docker.sock` (`dockerproxy.SocketFile`).
- **Defaults when the setting says nothing:** 5 containers, 1 GiB and 1 CPU per child.
- **Timers:** sockets are reconciled every 30 seconds; the sweep runs every 10 minutes.
- **The sweep removes:**
  - everything of an app without access;
  - an exited child older than 24 hours;
  - an empty network of the app's older than 1 hour.

  Volumes of an app with access stay.
- **A failed database read changes nothing.** Sockets stay as they are, and the sweep does not run, since without knowing who has access nothing is anybody's to remove.
- **Before this work is done:** `go build ./...`, `golangci-lint run ./...` over the whole repo (0 issues, 120 columns, US spelling), and `go test ./...` must all pass. Tests use testify `assert`, plus `if !assert... { t.FailNow() }` where going on is pointless. `require` is not vendored.
- **Git:** work on branch `feat/docker-api-agent`. Every commit ends with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`. At the end, merge into `main` locally and delete the branch. Do not push.

## File Structure

| File | Responsibility |
|---|---|
| `base/setting.go` | `SettingTypeAppDockerAPI` |
| `entity/setting_app_docker_api.go` | `AppDockerAPISettings`, its parser |
| `entity/setting_spec.go`, `specmodel/singleton.go`, `specserviceimpl/import_policy.go` | the new type in every registry that must list it |
| `service/dockerapiservice/types.go` | names, labels, defaults |
| `service/dockerapiservice/service.go` | `Service`: `Policies`, `SyncAgents`, `RemoveAppObjects` |
| `service/dockerapiservice/dockerapiserviceimpl/` | building policies, reaching the agents |
| `interface/agent/proto/docker_api.proto` (+ generated) | `DockerAPIService` gRPC |
| `interface/agent/client/dockerapiservice/service.go` | the backend's client for it |
| `interface/agent/server/server_docker_api.go` | the agent's side of it |
| `usecaseagent/dockerapiagentuc/host.go` | the sockets |
| `usecaseagent/dockerapiagentuc/sweep.go` | removing by label |
| `usecaseagent/dockerapiagentuc/uc.go` | `Sync`, `RemoveApp`, `Collect`, `Run`, `Close` |
| `cmd/internal/docker_api_host.go`, `cmd/agent/main.go`, `registry/provides.go` | wiring |

---

### Task 1: The `app-docker-api` setting type

**Files:**
- Modify: `hivepaas_app/base/setting.go`
- Create: `hivepaas_app/entity/setting_app_docker_api.go`
- Create: `hivepaas_app/entity/setting_app_docker_api_migration.go`
- Modify: `hivepaas_app/entity/setting_spec.go`
- Modify: `hivepaas_app/service/specservice/specmodel/singleton.go`
- Modify: `hivepaas_app/service/specservice/specserviceimpl/import_policy.go`
- Test: `hivepaas_app/entity/setting_app_docker_api_test.go`

**Interfaces:**
- Produces: `base.SettingTypeAppDockerAPI`, `entity.AppDockerAPISettings{Images, SharedDirs, Networks, Allow []string; Limits AppDockerAPILimits}`, `entity.AppDockerAPILimits{Containers int; Memory, NanoCPUs int64}`, `entity.DockerAPINetworkEnv`, `(*entity.Setting).AsAppDockerAPISettings()`.

- [ ] **Step 1: Create the branch**

```bash
cd /Users/tnt/go/src/github.com/hivepaas/hivepaas
git checkout -b feat/docker-api-agent
```

- [ ] **Step 2: Write the failing test**

`hivepaas_app/entity/setting_app_docker_api_test.go`:

```go
package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func TestAppDockerAPISettingsReadWhatWasStored(t *testing.T) {
	setting := &Setting{Type: base.SettingTypeAppDockerAPI, Data: `{"images":["autobase/automation:2.11.0"],` +
		`"sharedDirs":["/var/lib/autobase/ansible"],"networks":["env"],"allow":["exec"],` +
		`"limits":{"containers":3,"memory":2147483648,"nanoCpus":2000000000}}`}

	got, err := setting.AsAppDockerAPISettings()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, &AppDockerAPISettings{
		Images:     []string{"autobase/automation:2.11.0"},
		SharedDirs: []string{"/var/lib/autobase/ansible"},
		Networks:   []string{DockerAPINetworkEnv},
		Allow:      []string{"exec"},
		Limits:     AppDockerAPILimits{Containers: 3, Memory: 2 << 30, NanoCPUs: 2_000_000_000},
	}, got)
}
```

- [ ] **Step 3: Run it to see it fail**

Run: `go test ./hivepaas_app/entity/ -run TestAppDockerAPISettings`
Expected: FAIL, `undefined: base.SettingTypeAppDockerAPI`.

- [ ] **Step 4: Add the type**

In `hivepaas_app/base/setting.go`, after `SettingTypeAppDeployment`:

```go
	SettingTypeAppDockerAPI      SettingType = "app-docker-api"
```

`hivepaas_app/entity/setting_app_docker_api.go`:

```go
package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentAppDockerAPIVersion = 1

	// DockerAPINetworkEnv, in AppDockerAPISettings.Networks, is the app's own
	// project-env network.
	DockerAPINetworkEnv = "env"
)

var _ = registerSettingParser(base.SettingTypeAppDockerAPI, &appDockerAPISettingsParser{})

type appDockerAPISettingsParser struct {
}

func (s *appDockerAPISettingsParser) New() SettingData {
	return &AppDockerAPISettings{}
}

// AppDockerAPISettings is what an app may do through the Docker API HivePaaS
// serves it: the images its children run, the directories they share with it,
// the networks they join, and the endpoints beyond the core. An app without this
// setting has no socket at all. See
// docs/superpowers/specs/2026-09-24-docker-api-access-design.md.
type AppDockerAPISettings struct {
	// Images are patterns over what children may run; "*" is any image.
	Images []string `json:"images"`
	// SharedDirs are directories of the app's own storage a child may bind.
	SharedDirs []string `json:"sharedDirs,omitempty"`
	// Networks are networks children may join besides their own.
	Networks []string `json:"networks,omitempty"`
	// Allow are groups of endpoints beyond the core, as the proxy names them.
	Allow  []string           `json:"allow,omitempty"`
	Limits AppDockerAPILimits `json:"limits,omitzero"`
}

// AppDockerAPILimits bound what the app's children use. Zero is the default.
type AppDockerAPILimits struct {
	// Containers is how many children may exist at once.
	Containers int `json:"containers,omitempty"`
	// Memory is the most one child may have, in bytes.
	Memory int64 `json:"memory,omitempty"`
	// NanoCPUs is the most processor time one child may have, in billionths of
	// a CPU.
	NanoCPUs int64 `json:"nanoCpus,omitempty"`
}

func (s *AppDockerAPISettings) GetType() base.SettingType {
	return base.SettingTypeAppDockerAPI
}

func (s *AppDockerAPISettings) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *AppDockerAPISettings) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsAppDockerAPISettings() (*AppDockerAPISettings, error) {
	return parseSettingAs[*AppDockerAPISettings](s)
}

func (s *Setting) MustAsAppDockerAPISettings() *AppDockerAPISettings {
	return gofn.Must(s.AsAppDockerAPISettings())
}
```

Every setting type migrates its stored data between versions.
`hivepaas_app/entity/setting_app_docker_api_migration.go`:

```go
package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *AppDockerAPISettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAppDockerAPIVersion {
		return false, nil
	}
	if setting.Version > CurrentAppDockerAPIVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// Version 1 is the first, so an older row is one written before versions
	// were set: the data is already in its shape.
	setting.Version = CurrentAppDockerAPIVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
```

In `hivepaas_app/entity/setting_spec.go`, in the `registerDefaultSpecPolicies` list after `base.SettingTypeAppDeployment,`:

```go
		base.SettingTypeAppDockerAPI,
```

In `hivepaas_app/service/specservice/specmodel/singleton.go`, in `singletonBlockNames` after the `SettingTypeAppDeployment` entry:

```go
	base.SettingTypeAppDockerAPI:  "dockerApi",
```

In `hivepaas_app/service/specservice/specserviceimpl/import_policy.go`, add the reason to the `const` block with the others:

```go
	reasonGrantsDockerAPI = "it grants the Docker API, and import does not check who may grant it yet"
```

and the policy, with the skipped types:

```go
	base.SettingTypeAppDockerAPI:      {skip: reasonGrantsDockerAPI},
```

- [ ] **Step 5: Run the tests**

Run: `gofmt -w hivepaas_app && go test ./hivepaas_app/entity/... ./hivepaas_app/service/specservice/...`
Expected: `ok`. The registries' own tests pass: `TestEverySettingTypeIsClassifiedExactlyOnce`, `TestEverySettingTypeHasASpecPolicy`, `TestEveryBlockTypeHasAnImportPolicy`. If another coverage test names the new type, give it the entry the test asks for, the way `app-placement` has it.

- [ ] **Step 6: Commit**

```bash
git add hivepaas_app/base hivepaas_app/entity hivepaas_app/service/specservice
git commit -m "feat(settings): the app-docker-api setting type

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 2: The agent's sockets

**Files:**
- Create: `hivepaas_app/service/dockerapiservice/types.go`
- Create: `hivepaas_app/usecaseagent/dockerapiagentuc/host.go`
- Test: `hivepaas_app/usecaseagent/dockerapiagentuc/host_test.go`

**Interfaces:**
- Consumes: `dockerproxy.New`, `dockerproxy.Policy`, `(*dockerproxy.Proxy).SetPolicy`, `dockerproxy.SocketFile`.
- Produces: in `dockerapiservice`: `NetworkName(appID string) string`, `SocketVolumeName(appID string) string`, `SocketVolumeLabel`, `NetworkLabel`, `SocketVolumePrefix`, `DefaultContainers`, `DefaultMemory`, `DefaultNanoCPUs`. In `dockerapiagentuc`: `newSocketHost(logger, dockerManager, upstream) *socketHost`, `(*socketHost).reconcile(ctx, []*dockerproxy.Policy) error`, `closeApp(appID)`, `closeAll()`, `served() []string`, `newUpstream() (http.RoundTripper, error)`.

- [ ] **Step 1: Write the failing tests**

`hivepaas_app/usecaseagent/dockerapiagentuc/host_test.go`:

```go
package dockerapiagentuc

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/moby/moby/api/types/volume"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// fakeVolumes creates socket volumes in a temporary directory, as a node's
// daemon would under its data root.
type fakeVolumes struct {
	docker.Manager
	mu      sync.Mutex
	root    string
	created map[string]map[string]string
}

func (f *fakeVolumes) VolumeCreate(_ context.Context, options ...docker.VolumeCreateOption) (
	*client.VolumeCreateResult, error) {
	opts := client.VolumeCreateOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.created[opts.Name] = opts.Labels
	dir := filepath.Join(f.root, opts.Name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &client.VolumeCreateResult{Volume: volume.Volume{Name: opts.Name, Mountpoint: dir}}, nil
}

// shortTempDir is a directory with a path short enough for a unix socket, which
// macOS limits to 104 bytes and t.TempDir can exceed.
func shortTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "dapi")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return dir
}

// newTestHost is a host over fake volumes, in front of a daemon that answers
// every request with 200 and counts them.
func newTestHost(t *testing.T) (*socketHost, *fakeVolumes, *int) {
	t.Helper()
	requests := 0
	var mu sync.Mutex
	daemon := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		requests++
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
	}))
	t.Cleanup(daemon.Close)
	addr := daemon.Listener.Addr().String()
	upstream := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "tcp", addr)
	}}

	volumes := &fakeVolumes{root: shortTempDir(t), created: map[string]map[string]string{}}
	host := newSocketHost(logging.GlobalLogger(), volumes, upstream)
	t.Cleanup(host.closeAll)
	return host, volumes, &requests
}

func testPolicy(appID string) *dockerproxy.Policy {
	return &dockerproxy.Policy{
		AppID: appID, Images: []string{"alpine"}, Network: dockerapiservice.NetworkName(appID),
		SocketVolume: dockerapiservice.SocketVolumeName(appID),
		Limits:       dockerproxy.Limits{Containers: 1, Memory: 1 << 30, NanoCPUs: 1_000_000_000},
	}
}

func socketOf(volumes *fakeVolumes, appID string) string {
	return filepath.Join(volumes.root, dockerapiservice.SocketVolumeName(appID), dockerproxy.SocketFile)
}

// send makes a request over an app's socket.
func send(t *testing.T, socket, method, path string) (int, error) {
	t.Helper()
	httpClient := &http.Client{Transport: &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", socket)
		},
	}}
	req, err := http.NewRequestWithContext(context.Background(), method, "http://docker"+path, nil)
	if err != nil {
		return 0, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	_ = resp.Body.Close()
	return resp.StatusCode, nil
}

func TestReconcileServesEachAppOnASocketInItsVolume(t *testing.T) {
	host, volumes, requests := newTestHost(t)
	err := host.reconcile(context.Background(), []*dockerproxy.Policy{testPolicy("app1"), testPolicy("app2")})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, []string{"app1", "app2"}, host.served())
	assert.Equal(t, map[string]string{dockerapiservice.SocketVolumeLabel: "app1"},
		volumes.created[dockerapiservice.SocketVolumeName("app1")])

	for _, appID := range []string{"app1", "app2"} {
		status, err := send(t, socketOf(volumes, appID), http.MethodPost, "/v1.51/images/create?fromImage=alpine&tag=3")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, status)
	}
	assert.Equal(t, 2, *requests)
}

func TestReconcileGivesAChangedPolicyToTheOpenSocket(t *testing.T) {
	host, volumes, _ := newTestHost(t)
	ctx := context.Background()
	assert.NoError(t, host.reconcile(ctx, []*dockerproxy.Policy{testPolicy("app1")}))
	status, _ := send(t, socketOf(volumes, "app1"), http.MethodPost, "/v1.51/images/create?fromImage=busybox&tag=1")
	assert.Equal(t, http.StatusForbidden, status)

	changed := testPolicy("app1")
	changed.Images = []string{"busybox"}
	assert.NoError(t, host.reconcile(ctx, []*dockerproxy.Policy{changed}))
	status, _ = send(t, socketOf(volumes, "app1"), http.MethodPost, "/v1.51/images/create?fromImage=busybox&tag=1")
	assert.Equal(t, http.StatusOK, status)
}

func TestReconcileClosesTheSocketOfAnAppNoLongerListed(t *testing.T) {
	host, volumes, _ := newTestHost(t)
	ctx := context.Background()
	assert.NoError(t, host.reconcile(ctx, []*dockerproxy.Policy{testPolicy("app1")}))
	assert.NoError(t, host.reconcile(ctx, nil))

	assert.Empty(t, host.served())
	_, err := os.Stat(socketOf(volumes, "app1"))
	assert.ErrorIs(t, err, os.ErrNotExist)
	_, err = send(t, socketOf(volumes, "app1"), http.MethodGet, "/_ping")
	assert.Error(t, err)
}

func TestReconcileReplacesASocketAnEarlierAgentLeft(t *testing.T) {
	host, volumes, _ := newTestHost(t)
	socket := socketOf(volumes, "app1")
	if !assert.NoError(t, os.MkdirAll(filepath.Dir(socket), 0o755)) {
		t.FailNow()
	}
	assert.NoError(t, os.WriteFile(socket, []byte("stale"), 0o600))

	assert.NoError(t, host.reconcile(context.Background(), []*dockerproxy.Policy{testPolicy("app1")}))
	status, err := send(t, socket, http.MethodGet, "/_ping")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
}

func TestOneAppThatCannotBeServedDoesNotStopTheOthers(t *testing.T) {
	host, volumes, _ := newTestHost(t)
	unreachable := errors.New("socket volume unreachable")
	serveInVolume := host.socketDir
	host.socketDir = func(ctx context.Context, policy *dockerproxy.Policy) (string, error) {
		if policy.AppID == "broken" {
			return "", unreachable
		}
		return serveInVolume(ctx, policy)
	}

	err := host.reconcile(context.Background(), []*dockerproxy.Policy{testPolicy("broken"), testPolicy("app1")})
	assert.ErrorIs(t, err, unreachable)
	assert.Equal(t, []string{"app1"}, host.served())
	status, _ := send(t, socketOf(volumes, "app1"), http.MethodGet, "/_ping")
	assert.Equal(t, http.StatusOK, status)
}

func TestCloseAppStopsServingOneApp(t *testing.T) {
	host, _, _ := newTestHost(t)
	assert.NoError(t, host.reconcile(context.Background(),
		[]*dockerproxy.Policy{testPolicy("app1"), testPolicy("app2")}))
	host.closeApp("app1")
	assert.Equal(t, []string{"app2"}, host.served())
}
```

- [ ] **Step 2: Run them to see them fail**

Run: `go test ./hivepaas_app/usecaseagent/dockerapiagentuc/...`
Expected: FAIL, `no non-test Go files` or `undefined: newSocketHost`.

- [ ] **Step 3: Write the names and the host**

`hivepaas_app/service/dockerapiservice/types.go`:

```go
// Package dockerapiservice is the backend's side of giving apps the Docker API
// through the proxy each node's agent runs (pkg/dockerproxy): it turns
// app-docker-api settings into the proxy's policies, and asks the agents to act
// on a change now rather than at their next tick.
package dockerapiservice

const (
	// NetworkPrefix names an app's own network, which its children join.
	NetworkPrefix = "hp-dapi-"
	// NetworkLabel marks an app's own network, with the app's id. It is not the
	// proxy's owner label, so that the app can use the network but not remove it.
	NetworkLabel = "hivepaas.docker-api.network"

	// SocketVolumePrefix names the volume an app's socket lives in, on every node.
	SocketVolumePrefix = "hp-dapi-sock-"
	// SocketVolumeLabel marks a socket volume, with the app's id. It is not the
	// owner label either: the app must not see or remove the volume its own
	// socket lives in.
	SocketVolumeLabel = "hivepaas.docker-api.socket"

	// DefaultContainers, DefaultMemory and DefaultNanoCPUs are an app's limits
	// when its setting names none.
	DefaultContainers = 5
	DefaultMemory     = 1 << 30
	DefaultNanoCPUs   = 1_000_000_000
)

// NetworkName is the name of an app's own network.
func NetworkName(appID string) string {
	return NetworkPrefix + appID
}

// SocketVolumeName is the name of the volume an app's socket lives in.
func SocketVolumeName(appID string) string {
	return SocketVolumePrefix + appID
}
```

`hivepaas_app/usecaseagent/dockerapiagentuc/host.go`:

```go
package dockerapiagentuc

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	// socketMode lets an app running as any user reach its socket. Which apps
	// reach it is decided by who mounts the volume it lives in.
	socketMode = 0o666
	// readHeaderTimeout bounds a client that connects and says nothing. Nothing
	// else is bounded: logs and attach are streams.
	readHeaderTimeout = 10 * time.Second
)

// socketHost keeps one proxy socket per app with access, on this node.
type socketHost struct {
	mu      sync.Mutex
	sockets map[string]*appSocket

	logger   logging.Logger
	upstream http.RoundTripper
	// socketDir is where an app's socket goes on this node: its socket volume,
	// reached through the host's filesystem.
	socketDir func(ctx context.Context, policy *dockerproxy.Policy) (string, error)
}

type appSocket struct {
	proxy  *dockerproxy.Proxy
	server *http.Server
	path   string
}

func newSocketHost(logger logging.Logger, dockerManager docker.Manager, upstream http.RoundTripper) *socketHost {
	return &socketHost{
		sockets:  map[string]*appSocket{},
		logger:   logger,
		upstream: upstream,
		socketDir: func(ctx context.Context, policy *dockerproxy.Policy) (string, error) {
			return volumeSocketDir(ctx, dockerManager, policy)
		},
	}
}

// newUpstream reaches the daemon the way the docker CLI would, from the
// environment: on a node, its socket.
func newUpstream() (http.RoundTripper, error) {
	daemon, err := client.New(client.FromEnv)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	dial := daemon.Dialer()
	return &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return dial(ctx)
	}}, nil
}

// reconcile serves exactly the apps of policies. A new app gets a socket, an
// app whose policy changed has the new one from its next request, and an app no
// longer listed loses its socket. One app that cannot be served does not stop
// the others.
func (h *socketHost) reconcile(ctx context.Context, policies []*dockerproxy.Policy) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	wanted := make(map[string]bool, len(policies))
	var errs []error
	for _, policy := range policies {
		wanted[policy.AppID] = true
		if socket, found := h.sockets[policy.AppID]; found {
			socket.proxy.SetPolicy(policy)
			continue
		}
		if err := h.open(ctx, policy); err != nil {
			errs = append(errs, err)
		}
	}
	for appID, socket := range h.sockets {
		if !wanted[appID] {
			socket.close()
			delete(h.sockets, appID)
		}
	}
	return errors.Join(errs...)
}

func (h *socketHost) open(ctx context.Context, policy *dockerproxy.Policy) error {
	dir, err := h.socketDir(ctx, policy)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, dockerproxy.SocketFile)
	// A socket file outlives the agent that made it, and nothing can listen
	// where it is.
	if err = os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return hperrors.Wrap(err)
	}
	listener, err := (&net.ListenConfig{}).Listen(ctx, "unix", path)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if err = os.Chmod(path, socketMode); err != nil {
		_ = listener.Close()
		return hperrors.Wrap(err)
	}
	proxy := dockerproxy.New(policy, dockerproxy.Options{Upstream: h.upstream, OnDecision: h.logDecision})
	server := &http.Server{Handler: proxy, ReadHeaderTimeout: readHeaderTimeout}
	safego.GoWithLogger(h.logger, "dockerAPI.serve", func() {
		_ = server.Serve(listener)
	})
	h.sockets[policy.AppID] = &appSocket{proxy: proxy, server: server, path: path}
	return nil
}

func (s *appSocket) close() {
	_ = s.server.Close()
	_ = os.Remove(s.path)
}

// closeApp stops serving one app at once, rather than at the next reconcile.
func (h *socketHost) closeApp(appID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if socket, found := h.sockets[appID]; found {
		socket.close()
		delete(h.sockets, appID)
	}
}

// closeAll stops serving every app.
func (h *socketHost) closeAll() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for appID, socket := range h.sockets {
		socket.close()
		delete(h.sockets, appID)
	}
}

// served are the apps this node serves, sorted.
func (h *socketHost) served() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	apps := make([]string, 0, len(h.sockets))
	for appID := range h.sockets {
		apps = append(apps, appID)
	}
	sort.Strings(apps)
	return apps
}

func (h *socketHost) logDecision(d dockerproxy.Decision) {
	if d.Allowed {
		h.logger.Debugf("docker api: app %s: %s %s: %s", d.AppID, d.Method, d.Path, d.Reason)
		return
	}
	h.logger.Warnf("docker api: app %s refused %s %s: %s", d.AppID, d.Method, d.Path, d.Reason)
}

// volumeSocketDir is where an app's socket lives on this node: its socket
// volume. Creating a volume that exists returns it, so this also finds one that
// Swarm made first, when it started the app's task.
func volumeSocketDir(ctx context.Context, dockerManager docker.Manager, policy *dockerproxy.Policy) (string, error) {
	resp, err := dockerManager.VolumeCreate(ctx, func(opts *client.VolumeCreateOptions) {
		opts.Name = policy.SocketVolume
		opts.Labels = map[string]string{dockerapiservice.SocketVolumeLabel: policy.AppID}
	})
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	mountpoint := resp.Volume.Mountpoint
	// The agent's container reaches the host's filesystem under a prefix; an
	// agent running on the host itself reaches it as it is.
	candidates := []string{filepath.Join(volumeservice.HostPathPrefix, mountpoint), mountpoint}
	if i := slices.IndexFunc(candidates, isDir); i >= 0 {
		return candidates[i], nil
	}
	return "", hperrors.Wrap(hperrors.ErrInfraNotFound).
		WithMsgLog("socket volume %s is not reachable at %s", policy.SocketVolume, mountpoint)
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
```

- [ ] **Step 4: Run the tests, then lint the packages**

Run: `go test ./hivepaas_app/usecaseagent/dockerapiagentuc/... && golangci-lint run ./hivepaas_app/usecaseagent/dockerapiagentuc/... ./hivepaas_app/service/dockerapiservice/...`
Expected: `ok`. Lint may report that nothing outside the tests uses the host yet; Task 5 uses it.

- [ ] **Step 5: Commit**

```bash
git add hivepaas_app/service/dockerapiservice hivepaas_app/usecaseagent/dockerapiagentuc
git commit -m "feat(agent): serve each app with Docker API access on a socket in its volume

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 3: Removing what children leave

**Files:**
- Create: `hivepaas_app/usecaseagent/dockerapiagentuc/sweep.go`
- Test: `hivepaas_app/usecaseagent/dockerapiagentuc/sweep_test.go`

**Interfaces:**
- Consumes: `dockerproxy.OwnerLabel`, `dockerapiservice.SocketVolumePrefix`, `dockerapiservice.SocketVolumeName`.
- Produces: `type Removed struct{Containers, Networks, Volumes int}`, `type sweep struct{docker docker.Manager; now time.Time; gone func(appID string) bool; ageOut bool; only string}`, `(*sweep).run(ctx) (Removed, error)`.

- [ ] **Step 1: Write the failing tests**

`hivepaas_app/usecaseagent/dockerapiagentuc/sweep_test.go`:

```go
package dockerapiagentuc

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/volume"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

var sweepNow = time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)

// fakeNode is a node's daemon as far as a sweep asks: the listings honor the
// label and name filters a sweep uses, and removals are recorded.
type fakeNode struct {
	docker.Manager
	containers []container.Summary
	networks   []network.Summary
	attached   map[string]int
	volumes    []volume.Volume
	removed    []string
}

func owned(appID string) map[string]string {
	return map[string]string{dockerproxy.OwnerLabel: appID}
}

// matches applies a filters value of the kinds a sweep sends: "label" as key or
// key=value, and "name" as a substring.
func matches(filters client.Filters, name string, labels map[string]string) bool {
	for want := range filters["label"] {
		key, value, withValue := strings.Cut(want, "=")
		got, found := labels[key]
		if !found || (withValue && got != value) {
			return false
		}
	}
	for want := range filters["name"] {
		if !strings.Contains(name, want) {
			return false
		}
	}
	return true
}

func (f *fakeNode) ContainerList(_ context.Context, options ...docker.ContainerListOption) (
	*client.ContainerListResult, error) {
	opts := client.ContainerListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	out := &client.ContainerListResult{}
	for _, c := range f.containers {
		if matches(opts.Filters, c.ID, c.Labels) {
			out.Items = append(out.Items, c)
		}
	}
	return out, nil
}

func (f *fakeNode) ContainerRemove(_ context.Context, id string, _ ...docker.ContainerRemoveOption) (
	*client.ContainerRemoveResult, error) {
	f.removed = append(f.removed, "container "+id)
	return &client.ContainerRemoveResult{}, nil
}

func (f *fakeNode) NetworkList(_ context.Context, options ...docker.NetworkListOption) (
	*client.NetworkListResult, error) {
	opts := client.NetworkListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	out := &client.NetworkListResult{}
	for _, n := range f.networks {
		if matches(opts.Filters, n.Name, n.Labels) {
			out.Items = append(out.Items, n)
		}
	}
	return out, nil
}

func (f *fakeNode) NetworkInspect(_ context.Context, id string, _ ...docker.NetworkInspectOption) (
	*client.NetworkInspectResult, error) {
	attached := map[string]network.EndpointResource{}
	for i := range f.attached[id] {
		attached[string(rune('a'+i))] = network.EndpointResource{}
	}
	return &client.NetworkInspectResult{Network: network.Inspect{Containers: attached}}, nil
}

func (f *fakeNode) NetworkRemove(_ context.Context, id string, _ ...docker.NetworkRemoveOption) (
	*client.NetworkRemoveResult, error) {
	f.removed = append(f.removed, "network "+id)
	return &client.NetworkRemoveResult{}, nil
}

func (f *fakeNode) VolumeList(_ context.Context, options ...docker.VolumeListOption) (
	*client.VolumeListResult, error) {
	opts := client.VolumeListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	out := &client.VolumeListResult{}
	for _, v := range f.volumes {
		if matches(opts.Filters, v.Name, v.Labels) {
			out.Items = append(out.Items, v)
		}
	}
	return out, nil
}

func (f *fakeNode) VolumeRemove(_ context.Context, id string, _ bool, _ ...docker.VolumeRemoveOption) (
	*client.VolumeRemoveResult, error) {
	f.removed = append(f.removed, "volume "+id)
	return &client.VolumeRemoveResult{}, nil
}

func networkSummary(id string, labels map[string]string, created time.Time) network.Summary {
	return network.Summary{Network: network.Network{ID: id, Name: id, Labels: labels, Created: created}}
}

// aNode holds, for app1 (which keeps its access) and app2 (which lost it):
// running, recently exited and long-exited children, fresh, idle and busy
// networks, cache volumes and socket volumes - and objects of nobody's.
func aNode() *fakeNode {
	hour := time.Hour
	return &fakeNode{
		containers: []container.Summary{
			{ID: "app1-running", Labels: owned("app1"), State: container.StateRunning,
				Created: sweepNow.Add(-48 * hour).Unix()},
			{ID: "app1-exited-now", Labels: owned("app1"), State: container.StateExited,
				Created: sweepNow.Add(-hour).Unix()},
			{ID: "app1-exited-old", Labels: owned("app1"), State: container.StateExited,
				Created: sweepNow.Add(-25 * hour).Unix()},
			{ID: "app2-running", Labels: owned("app2"), State: container.StateRunning,
				Created: sweepNow.Unix()},
			{ID: "hivepaas-db", Labels: map[string]string{}, State: container.StateExited,
				Created: sweepNow.Add(-100 * hour).Unix()},
		},
		networks: []network.Summary{
			networkSummary("app1-fresh", owned("app1"), sweepNow.Add(-10*time.Minute)),
			networkSummary("app1-idle", owned("app1"), sweepNow.Add(-2*hour)),
			networkSummary("app1-busy", owned("app1"), sweepNow.Add(-2*hour)),
			networkSummary("app2-net", owned("app2"), sweepNow),
			networkSummary("hivepaas_net", map[string]string{}, sweepNow.Add(-100*hour)),
		},
		attached: map[string]int{"app1-busy": 1},
		volumes: []volume.Volume{
			{Name: "app1-cache", Labels: owned("app1")},
			{Name: "app2-cache", Labels: owned("app2")},
			{Name: dockerapiservice.SocketVolumeName("app1"), Labels: map[string]string{}},
			{Name: dockerapiservice.SocketVolumeName("app2"), Labels: map[string]string{}},
			{Name: "hp-vol-data", Labels: map[string]string{}},
		},
	}
}

func TestCollectRemovesWhatIsLeftBehind(t *testing.T) {
	node := aNode()
	access := map[string]bool{"app1": true}
	s := &sweep{docker: node, now: sweepNow, ageOut: true, gone: func(appID string) bool { return !access[appID] }}

	removed, err := s.run(context.Background())
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{
		"container app1-exited-old", "container app2-running",
		"network app1-idle", "network app2-net",
		"volume app2-cache", "volume " + dockerapiservice.SocketVolumeName("app2"),
	}, node.removed)
	assert.Equal(t, Removed{Containers: 2, Networks: 2, Volumes: 2}, removed)
}

func TestRemovingAnAppTouchesOnlyThatApp(t *testing.T) {
	node := aNode()
	s := &sweep{docker: node, now: sweepNow, only: "app1", gone: func(appID string) bool { return appID == "app1" }}

	removed, err := s.run(context.Background())
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{
		"container app1-running", "container app1-exited-now", "container app1-exited-old",
		"network app1-fresh", "network app1-idle", "network app1-busy",
		"volume app1-cache", "volume " + dockerapiservice.SocketVolumeName("app1"),
	}, node.removed)
	assert.Equal(t, Removed{Containers: 3, Networks: 3, Volumes: 2}, removed)
}
```

- [ ] **Step 2: Run them to see them fail**

Run: `go test ./hivepaas_app/usecaseagent/dockerapiagentuc/...`
Expected: FAIL, `undefined: sweep`.

- [ ] **Step 3: Write the sweep**

`hivepaas_app/usecaseagent/dockerapiagentuc/sweep.go`:

```go
package dockerapiagentuc

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	// exitedChildAge is how long an exited child is kept for the app to read.
	exitedChildAge = 24 * time.Hour
	// idleNetworkAge is how long a network of the app's may stand empty: a job
	// makes its network before its first container.
	idleNetworkAge = time.Hour
)

// finishedStates are the states of a child that is not coming back by itself.
var finishedStates = []container.ContainerState{container.StateCreated, container.StateExited, container.StateDead}

// Removed counts what a sweep took away.
type Removed struct {
	Containers int
	Networks   int
	Volumes    int
}

// sweep removes, on this node, what apps' children left: everything of an app
// that is gone and, when ageOut is set, an exited child older than a day and an
// empty network older than an hour of an app that stays. Volumes of an app that
// stays are kept, since they are its caches.
type sweep struct {
	docker docker.Manager
	now    time.Time
	// gone reports an app whose every object goes.
	gone   func(appID string) bool
	ageOut bool
	// only, when set, limits the sweep to one app's objects.
	only string
}

func (s *sweep) run(ctx context.Context) (Removed, error) {
	var removed Removed
	var errs []error
	var err error
	removed.Containers, err = s.containers(ctx)
	errs = append(errs, err)
	removed.Networks, err = s.networks(ctx)
	errs = append(errs, err)
	volumes, err := s.volumes(ctx)
	errs = append(errs, err)
	sockets, err := s.socketVolumes(ctx)
	errs = append(errs, err)
	removed.Volumes = volumes + sockets
	return removed, errors.Join(errs...)
}

// ownerFilter selects the objects the proxy created, of one app when the sweep
// is limited to it.
func (s *sweep) ownerFilter() string {
	if s.only != "" {
		return dockerproxy.OwnerLabel + "=" + s.only
	}
	return dockerproxy.OwnerLabel
}

func (s *sweep) containers(ctx context.Context) (int, error) {
	resp, err := s.docker.ContainerList(ctx, func(opts *client.ContainerListOptions) {
		opts.All = true
		docker.FilterAdd(&opts.Filters, "label", s.ownerFilter())
	})
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	count := 0
	var errs []error
	for _, c := range resp.Items {
		stale := s.gone(c.Labels[dockerproxy.OwnerLabel]) ||
			(s.ageOut && slices.Contains(finishedStates, c.State) &&
				s.now.Sub(time.Unix(c.Created, 0)) > exitedChildAge)
		if !stale {
			continue
		}
		_, err = s.docker.ContainerRemove(ctx, c.ID, func(opts *client.ContainerRemoveOptions) {
			opts.Force = true
			opts.RemoveVolumes = true
		})
		if err != nil {
			errs = append(errs, hperrors.Wrap(err))
			continue
		}
		count++
	}
	return count, errors.Join(errs...)
}

func (s *sweep) networks(ctx context.Context) (int, error) {
	resp, err := s.docker.NetworkList(ctx, func(opts *client.NetworkListOptions) {
		docker.FilterAdd(&opts.Filters, "label", s.ownerFilter())
	})
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	count := 0
	var errs []error
	for _, n := range resp.Items {
		if !s.gone(n.Labels[dockerproxy.OwnerLabel]) {
			idle, idleErr := s.idleNetwork(ctx, n.ID, n.Created)
			if idleErr != nil {
				errs = append(errs, idleErr)
			}
			if !idle {
				continue
			}
		}
		if _, err = s.docker.NetworkRemove(ctx, n.ID); err != nil {
			errs = append(errs, hperrors.Wrap(err))
			continue
		}
		count++
	}
	return count, errors.Join(errs...)
}

// idleNetwork reports a network of an app that stays, old enough and with
// nothing attached, when the sweep ages things out.
func (s *sweep) idleNetwork(ctx context.Context, id string, created time.Time) (bool, error) {
	if !s.ageOut || s.now.Sub(created) <= idleNetworkAge {
		return false, nil
	}
	inspect, err := s.docker.NetworkInspect(ctx, id)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	return len(inspect.Network.Containers) == 0, nil
}

func (s *sweep) volumes(ctx context.Context) (int, error) {
	resp, err := s.docker.VolumeList(ctx, func(opts *client.VolumeListOptions) {
		docker.FilterAdd(&opts.Filters, "label", s.ownerFilter())
	})
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	var names []string
	for _, v := range resp.Items {
		if s.gone(v.Labels[dockerproxy.OwnerLabel]) {
			names = append(names, v.Name)
		}
	}
	return s.removeVolumes(ctx, names)
}

// socketVolumes removes the socket volumes of apps that are gone. They are
// found by name: one that Swarm created when it started the app's task carries
// no label.
func (s *sweep) socketVolumes(ctx context.Context) (int, error) {
	prefix := dockerapiservice.SocketVolumePrefix
	if s.only != "" {
		prefix = dockerapiservice.SocketVolumeName(s.only)
	}
	resp, err := s.docker.VolumeList(ctx, func(opts *client.VolumeListOptions) {
		docker.FilterAdd(&opts.Filters, "name", prefix)
	})
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	var names []string
	for _, v := range resp.Items {
		appID, isSocket := strings.CutPrefix(v.Name, dockerapiservice.SocketVolumePrefix)
		if isSocket && s.gone(appID) {
			names = append(names, v.Name)
		}
	}
	return s.removeVolumes(ctx, names)
}

// removeVolumes removes what it can. A volume still mounted - by the task of an
// app being deleted - stays until a later sweep.
func (s *sweep) removeVolumes(ctx context.Context, names []string) (int, error) {
	count := 0
	var errs []error
	for _, name := range names {
		if _, err := s.docker.VolumeRemove(ctx, name, false); err != nil {
			errs = append(errs, hperrors.Wrap(err))
			continue
		}
		count++
	}
	return count, errors.Join(errs...)
}
```

- [ ] **Step 4: Run the tests, then lint the package**

Run: `go test ./hivepaas_app/usecaseagent/dockerapiagentuc/... && golangci-lint run ./hivepaas_app/usecaseagent/dockerapiagentuc/...`
Expected: `ok`. Lint may report that nothing outside the tests uses the sweep yet; Task 5 uses it.

- [ ] **Step 5: Commit**

```bash
git add hivepaas_app/usecaseagent/dockerapiagentuc
git commit -m "feat(agent): remove what apps' children leave on a node

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 4: The gRPC service, the agent client, and the backend service

**Files:**
- Create: `hivepaas_app/interface/agent/proto/docker_api.proto` (+ generated `docker_api.pb.go`, `docker_api_grpc.pb.go`)
- Modify: `hivepaas_app/interface/agent/proto/generate.go`
- Create: `hivepaas_app/interface/agent/client/dockerapiservice/service.go`
- Create: `hivepaas_app/service/dockerapiservice/service.go`
- Create: `hivepaas_app/service/dockerapiservice/dockerapiserviceimpl/service.go`
- Create: `hivepaas_app/service/dockerapiservice/dockerapiserviceimpl/policies.go`
- Create: `hivepaas_app/service/dockerapiservice/dockerapiserviceimpl/agents.go`
- Test: `hivepaas_app/service/dockerapiservice/dockerapiserviceimpl/policies_test.go`
- Test: `hivepaas_app/service/dockerapiservice/dockerapiserviceimpl/agents_test.go`

**Interfaces:**
- Consumes: `entity.AppDockerAPISettings`, `dockerapiservice` names (Task 2), `agentservice.Service.GetAgentAddrForNode`, `networkservice.Service.GetProjectNetworkName`.
- Produces:
  - gRPC `DockerAPIService` with `SyncDockerAPI(DockerAPISyncReq) returns (DockerAPISyncResp{apps})` and `RemoveDockerAPIApp(DockerAPIRemoveAppReq{app_id}) returns (DockerAPIRemoveAppResp{containers, networks, volumes})`.
  - Client: `dockerapiservice.NewDockerAPIServiceClient(addr) (DockerAPIServiceClient, error)` with `Sync(ctx) (int, error)`, `RemoveApp(ctx, appID) (*RemoveAppResult, error)`, `Close() error`.
  - Backend: `dockerapiservice.Service{Policies(ctx, db) ([]*dockerproxy.Policy, error); SyncAgents(ctx) error; RemoveAppObjects(ctx, appID) error}` and `dockerapiserviceimpl.New(settingRepo, appRepo, networkService, agentService, dockerManager) dockerapiservice.Service`.

- [ ] **Step 1: Write the proto and generate it**

`hivepaas_app/interface/agent/proto/docker_api.proto`:

```proto
syntax = "proto3";

package agent;

option go_package = "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto;agentproto";

// DockerAPIService is the agent's side of serving apps the Docker API through
// a proxy: every node's agent serves every app that has access.
service DockerAPIService {
  // SyncDockerAPI reads which apps have access and serves exactly those, now
  // rather than at the next tick.
  rpc SyncDockerAPI(DockerAPISyncReq) returns (DockerAPISyncResp);
  // RemoveDockerAPIApp stops serving an app and removes what its children left
  // on the node.
  rpc RemoveDockerAPIApp(DockerAPIRemoveAppReq) returns (DockerAPIRemoveAppResp);
}

message DockerAPISyncReq {}

message DockerAPISyncResp {
  int32 apps = 1;
}

message DockerAPIRemoveAppReq {
  string app_id = 1;
}

message DockerAPIRemoveAppResp {
  int32 containers = 1;
  int32 networks = 2;
  int32 volumes = 3;
}
```

Append to `hivepaas_app/interface/agent/proto/generate.go`:

```go
//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative docker_api.proto
```

Run: `cd hivepaas_app/interface/agent/proto && protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative docker_api.proto && cd - && go build ./hivepaas_app/interface/agent/proto/`
Expected: `docker_api.pb.go` and `docker_api_grpc.pb.go` appear, and the package builds.

- [ ] **Step 2: Write the client**

`hivepaas_app/interface/agent/client/dockerapiservice/service.go`:

```go
package dockerapiservice

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
)

// RemoveAppResult is what an agent removed of an app on its node.
type RemoveAppResult struct {
	Containers int
	Networks   int
	Volumes    int
}

// DockerAPIServiceClient reaches one node's agent about the Docker API it
// serves apps.
type DockerAPIServiceClient interface {
	// Sync has the agent serve exactly the apps that have access now, and
	// returns how many it serves.
	Sync(ctx context.Context) (int, error)
	// RemoveApp has the agent stop serving an app and remove what its children
	// left on the node.
	RemoveApp(ctx context.Context, appID string) (*RemoveAppResult, error)
	Close() error
}

type grpcDockerAPIServiceClient struct {
	protoClient agentproto.DockerAPIServiceClient
	conn        *grpc.ClientConn
}

func NewDockerAPIServiceClient(agentAddr string) (DockerAPIServiceClient, error) {
	conn, err := grpc.NewClient(agentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &grpcDockerAPIServiceClient{
		conn:        conn,
		protoClient: agentproto.NewDockerAPIServiceClient(conn),
	}, nil
}

func (c *grpcDockerAPIServiceClient) Close() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

func (c *grpcDockerAPIServiceClient) Sync(ctx context.Context) (int, error) {
	resp, err := c.protoClient.SyncDockerAPI(client.CreateAuthCtx(ctx), &agentproto.DockerAPISyncReq{})
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	return int(resp.GetApps()), nil
}

func (c *grpcDockerAPIServiceClient) RemoveApp(ctx context.Context, appID string) (*RemoveAppResult, error) {
	resp, err := c.protoClient.RemoveDockerAPIApp(client.CreateAuthCtx(ctx),
		&agentproto.DockerAPIRemoveAppReq{AppId: appID})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &RemoveAppResult{
		Containers: int(resp.GetContainers()),
		Networks:   int(resp.GetNetworks()),
		Volumes:    int(resp.GetVolumes()),
	}, nil
}
```

- [ ] **Step 3: Write the failing service tests**

`hivepaas_app/service/dockerapiservice/dockerapiserviceimpl/policies_test.go`:

```go
package dockerapiserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
)

type fakeSettingRepo struct {
	repository.SettingRepo
	settings []*entity.Setting
}

func (f *fakeSettingRepo) List(_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ *basedto.Paging,
	_ ...bunex.SelectQueryOption) ([]*entity.Setting, *basedto.PagingMeta, error) {
	return f.settings, nil, nil
}

type fakeAppRepo struct {
	repository.AppRepo
	apps []*entity.App
}

func (f *fakeAppRepo) ListByIDs(_ context.Context, _ database.IDB, _ string, ids []string,
	_ ...bunex.SelectQueryOption) ([]*entity.App, error) {
	var out []*entity.App
	for _, app := range f.apps {
		for _, id := range ids {
			if app.ID == id {
				out = append(out, app)
			}
		}
	}
	return out, nil
}

type fakeNetworkService struct {
	networkservice.Service
}

func (fakeNetworkService) GetProjectNetworkName(project *entity.Project, env string) string {
	return project.Key + "_" + env + "_net"
}

func dockerAPISetting(appID, data string) *entity.Setting {
	return &entity.Setting{ID: "s-" + appID, Scope: base.ObjectScopeApp, ObjectID: appID,
		Type: base.SettingTypeAppDockerAPI, Status: base.SettingStatusActive, Data: data}
}

func appIn(id string) *entity.App {
	return &entity.App{ID: id, ServiceID: "svc-" + id, Project: &entity.Project{Key: "shop"},
		ProjectEnv: &entity.ProjectEnv{Name: "prod"}}
}

func TestPoliciesComeFromTheAppsSettings(t *testing.T) {
	svc := &service{
		settingRepo: &fakeSettingRepo{settings: []*entity.Setting{
			dockerAPISetting("runner", `{"images":["*"],"networks":["env"],`+
				`"allow":["exec","files","volumes","networks","nestedSocket"]}`),
			dockerAPISetting("autobase", `{"images":["autobase/automation:2.11.0"],`+
				`"sharedDirs":["/var/lib/autobase/ansible"],"limits":{"containers":3,"memory":2147483648}}`),
			// An app deleted since: its setting row is still there for a moment.
			dockerAPISetting("gone", `{"images":["*"]}`),
		}},
		appRepo:        &fakeAppRepo{apps: []*entity.App{appIn("runner"), appIn("autobase")}},
		networkService: fakeNetworkService{},
	}

	policies, err := svc.Policies(context.Background(), nil)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, []*dockerproxy.Policy{
		{
			AppID: "runner", ServiceID: "svc-runner", Images: []string{"*"},
			Network: "hp-dapi-runner", Networks: []string{"shop_prod_net"}, SocketVolume: "hp-dapi-sock-runner",
			Allow: []dockerproxy.Group{dockerproxy.GroupExec, dockerproxy.GroupFiles, dockerproxy.GroupVolumes,
				dockerproxy.GroupNetworks, dockerproxy.GroupNestedSocket},
			Limits: dockerproxy.Limits{Containers: 5, Memory: 1 << 30, NanoCPUs: 1_000_000_000},
		},
		{
			AppID: "autobase", ServiceID: "svc-autobase", Images: []string{"autobase/automation:2.11.0"},
			SharedDirs: []string{"/var/lib/autobase/ansible"},
			Network:    "hp-dapi-autobase", SocketVolume: "hp-dapi-sock-autobase",
			Limits: dockerproxy.Limits{Containers: 3, Memory: 2 << 30, NanoCPUs: 1_000_000_000},
		},
	}, policies)
}
```

`hivepaas_app/service/dockerapiservice/dockerapiserviceimpl/agents_test.go`:

```go
package dockerapiserviceimpl

import (
	"context"
	"errors"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	dockerapiclient "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/dockerapiservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type fakeNodes struct {
	docker.Manager
	nodes []swarm.Node
}

func (f *fakeNodes) NodeList(_ context.Context, _ ...docker.NodeListOption) (*client.NodeListResult, error) {
	return &client.NodeListResult{Items: f.nodes}, nil
}

type fakeAgents struct {
	agentservice.Service
}

func (fakeAgents) GetAgentAddrForNode(_ context.Context, nodeID string) (string, error) {
	return "agent-on-" + nodeID, nil
}

// fakeAgentClient records what it is asked, and fails where told to.
type fakeAgentClient struct {
	addr  string
	calls *[]string
	fails bool
}

var errAgentDown = errors.New("agent down")

func (c *fakeAgentClient) Sync(context.Context) (int, error) {
	*c.calls = append(*c.calls, "sync "+c.addr)
	if c.fails {
		return 0, errAgentDown
	}
	return 1, nil
}

func (c *fakeAgentClient) RemoveApp(_ context.Context, appID string) (*dockerapiclient.RemoveAppResult, error) {
	*c.calls = append(*c.calls, "remove "+appID+" "+c.addr)
	return &dockerapiclient.RemoveAppResult{}, nil
}

func (c *fakeAgentClient) Close() error {
	return nil
}

func node(id string, state swarm.NodeState) swarm.Node {
	return swarm.Node{ID: id, Description: swarm.NodeDescription{Hostname: id}, Status: swarm.NodeStatus{State: state}}
}

func fanOutService(calls *[]string, failing string) *service {
	return &service{
		dockerManager: &fakeNodes{nodes: []swarm.Node{
			node("n1", swarm.NodeStateReady), node("n2", swarm.NodeStateReady), node("n3", swarm.NodeStateDown),
		}},
		agentService: fakeAgents{},
		agentClient: func(addr string) (dockerapiclient.DockerAPIServiceClient, error) {
			return &fakeAgentClient{addr: addr, calls: calls, fails: addr == "agent-on-"+failing}, nil
		},
	}
}

func TestSyncAgentsReachesEveryNodeThatIsUp(t *testing.T) {
	var calls []string
	assert.NoError(t, fanOutService(&calls, "").SyncAgents(context.Background()))
	assert.Equal(t, []string{"sync agent-on-n1", "sync agent-on-n2"}, calls)
}

func TestOneAgentFailingDoesNotKeepTheOthersFromHearing(t *testing.T) {
	var calls []string
	err := fanOutService(&calls, "n1").SyncAgents(context.Background())
	assert.ErrorIs(t, err, errAgentDown)
	assert.Contains(t, err.Error(), "node n1")
	assert.Equal(t, []string{"sync agent-on-n1", "sync agent-on-n2"}, calls)
}

func TestRemoveAppObjectsAsksEveryNode(t *testing.T) {
	var calls []string
	assert.NoError(t, fanOutService(&calls, "").RemoveAppObjects(context.Background(), "app1"))
	assert.Equal(t, []string{"remove app1 agent-on-n1", "remove app1 agent-on-n2"}, calls)
}
```

- [ ] **Step 4: Run them to see them fail**

Run: `go test ./hivepaas_app/service/dockerapiservice/...`
Expected: FAIL, `undefined: service`.

- [ ] **Step 5: Write the service**

`hivepaas_app/service/dockerapiservice/service.go`:

```go
package dockerapiservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
)

type Service interface {
	// Policies are what every app with Docker API access may do, read from the
	// database. Every node's agent serves exactly these.
	Policies(ctx context.Context, db database.IDB) ([]*dockerproxy.Policy, error)
	// SyncAgents asks every node's agent to serve exactly the apps that have
	// access now, rather than at its next tick.
	SyncAgents(ctx context.Context) error
	// RemoveAppObjects asks every node's agent to stop serving an app and to
	// remove what its children left: containers, networks and volumes.
	RemoveAppObjects(ctx context.Context, appID string) error
}
```

`hivepaas_app/service/dockerapiservice/dockerapiserviceimpl/service.go`:

```go
package dockerapiserviceimpl

import (
	dockerapiclient "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/dockerapiservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	settingRepo    repository.SettingRepo
	appRepo        repository.AppRepo
	networkService networkservice.Service
	agentService   agentservice.Service
	dockerManager  docker.Manager
	// agentClient reaches one node's agent.
	agentClient func(addr string) (dockerapiclient.DockerAPIServiceClient, error)
}

func New(
	settingRepo repository.SettingRepo,
	appRepo repository.AppRepo,
	networkService networkservice.Service,
	agentService agentservice.Service,
	dockerManager docker.Manager,
) dockerapiservice.Service {
	return &service{
		settingRepo:    settingRepo,
		appRepo:        appRepo,
		networkService: networkService,
		agentService:   agentService,
		dockerManager:  dockerManager,
		agentClient:    dockerapiclient.NewDockerAPIServiceClient,
	}
}
```

`hivepaas_app/service/dockerapiservice/dockerapiserviceimpl/policies.go`:

```go
package dockerapiserviceimpl

import (
	"context"
	"slices"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
)

func (s *service) Policies(ctx context.Context, db database.IDB) ([]*dockerproxy.Policy, error) {
	settings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDockerAPI),
		bunex.SelectWhere("setting.scope = ?", base.ObjectScopeApp),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	appIDs := make([]string, 0, len(settings))
	for _, setting := range settings {
		appIDs = append(appIDs, setting.ObjectID)
	}
	apps, err := s.appRepo.ListByIDs(ctx, db, "", appIDs,
		bunex.SelectRelation("Project"), bunex.SelectRelation("ProjectEnv"))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	appByID := make(map[string]*entity.App, len(apps))
	for _, app := range apps {
		appByID[app.ID] = app
	}

	policies := make([]*dockerproxy.Policy, 0, len(settings))
	for _, setting := range settings {
		// An app deleted a moment ago can still have its setting row.
		app := appByID[setting.ObjectID]
		if app == nil || app.Project == nil || app.ProjectEnv == nil {
			continue
		}
		data, err := setting.AsAppDockerAPISettings()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		envNetwork := s.networkService.GetProjectNetworkName(app.Project, app.ProjectEnv.Name)
		policies = append(policies, policyOf(app, data, envNetwork))
	}
	return policies, nil
}

// policyOf is what the proxy enforces for one app.
func policyOf(app *entity.App, data *entity.AppDockerAPISettings, envNetwork string) *dockerproxy.Policy {
	policy := &dockerproxy.Policy{
		AppID:        app.ID,
		ServiceID:    app.ServiceID,
		Images:       data.Images,
		SharedDirs:   data.SharedDirs,
		Network:      dockerapiservice.NetworkName(app.ID),
		SocketVolume: dockerapiservice.SocketVolumeName(app.ID),
		Limits: dockerproxy.Limits{
			Containers: gofn.Coalesce(data.Limits.Containers, dockerapiservice.DefaultContainers),
			Memory:     gofn.Coalesce(data.Limits.Memory, dockerapiservice.DefaultMemory),
			NanoCPUs:   gofn.Coalesce(data.Limits.NanoCPUs, dockerapiservice.DefaultNanoCPUs),
		},
	}
	if slices.Contains(data.Networks, entity.DockerAPINetworkEnv) {
		policy.Networks = []string{envNetwork}
	}
	for _, group := range data.Allow {
		policy.Allow = append(policy.Allow, dockerproxy.Group(group))
	}
	return policy
}
```

`hivepaas_app/service/dockerapiservice/dockerapiserviceimpl/agents.go`:

```go
package dockerapiserviceimpl

import (
	"context"
	"errors"
	"fmt"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	dockerapiclient "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/dockerapiservice"
)

func (s *service) SyncAgents(ctx context.Context) error {
	return s.eachAgent(ctx, func(agent dockerapiclient.DockerAPIServiceClient) error {
		_, err := agent.Sync(ctx)
		return hperrors.Wrap(err)
	})
}

func (s *service) RemoveAppObjects(ctx context.Context, appID string) error {
	return s.eachAgent(ctx, func(agent dockerapiclient.DockerAPIServiceClient) error {
		_, err := agent.RemoveApp(ctx, appID)
		return hperrors.Wrap(err)
	})
}

// eachAgent calls fn with the agent of every node that is up. A node that
// cannot be reached, or whose agent fails, is reported, and does not keep the
// others from hearing.
func (s *service) eachAgent(ctx context.Context, fn func(dockerapiclient.DockerAPIServiceClient) error) error {
	nodes, err := s.dockerManager.NodeList(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	var errs []error
	for i := range nodes.Items {
		node := &nodes.Items[i]
		if node.Status.State != swarm.NodeStateReady {
			continue
		}
		if err = s.callAgent(ctx, node.ID, fn); err != nil {
			errs = append(errs, fmt.Errorf("node %s: %w", node.Description.Hostname, err))
		}
	}
	return errors.Join(errs...)
}

func (s *service) callAgent(ctx context.Context, nodeID string,
	fn func(dockerapiclient.DockerAPIServiceClient) error) error {
	addr, err := s.agentService.GetAgentAddrForNode(ctx, nodeID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	agent, err := s.agentClient(addr)
	if err != nil {
		return hperrors.Wrap(err)
	}
	defer func() { _ = agent.Close() }()
	return fn(agent)
}
```

- [ ] **Step 6: Run the tests, then lint**

Run: `go test ./hivepaas_app/service/dockerapiservice/... && golangci-lint run ./hivepaas_app/service/dockerapiservice/... ./hivepaas_app/interface/agent/...`
Expected: `ok`, `0 issues`.

- [ ] **Step 7: Commit**

```bash
git add hivepaas_app/interface/agent hivepaas_app/service/dockerapiservice
git commit -m "feat(dockerapi): turn settings into policies, and reach every node's agent

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 5: The agent use case, its gRPC side, and the wiring

**Files:**
- Create: `hivepaas_app/usecaseagent/dockerapiagentuc/uc.go`
- Test: `hivepaas_app/usecaseagent/dockerapiagentuc/uc_test.go`
- Create: `hivepaas_app/interface/agent/server/server_docker_api.go`
- Modify: `hivepaas_app/interface/agent/server/server.go`
- Create: `hivepaas_app/cmd/internal/docker_api_host.go`
- Modify: `hivepaas_app/cmd/agent/main.go`
- Modify: `hivepaas_app/registry/provides.go`
- Test: `hivepaas_app/cmd/internal/wiring_agent_test.go`

**Interfaces:**
- Consumes: Tasks 2-4.
- Produces: `dockerapiagentuc.New(logger, db, dockerManager, dockerAPIService) (*UC, error)`, `(*UC).Sync(ctx) (int, error)`, `(*UC).RemoveApp(ctx, appID) (Removed, error)`, `(*UC).Collect(ctx) (Removed, error)`, `(*UC).Run(ctx)`, `(*UC).Close()`, `internal.InitDockerAPIHost`.

- [ ] **Step 1: Write the failing tests**

`hivepaas_app/usecaseagent/dockerapiagentuc/uc_test.go`:

```go
package dockerapiagentuc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
)

// fakePolicies answers Policies with what a test sets, or fails.
type fakePolicies struct {
	dockerapiservice.Service
	policies []*dockerproxy.Policy
	err      error
}

func (f *fakePolicies) Policies(context.Context, database.IDB) ([]*dockerproxy.Policy, error) {
	return f.policies, f.err
}

func newTestUC(t *testing.T, policies *fakePolicies, node *fakeNode) *UC {
	t.Helper()
	host, _, _ := newTestHost(t)
	return &UC{logger: logging.GlobalLogger(), dockerManager: node, dockerAPIService: policies, host: host}
}

func TestSyncServesTheAppsThatHaveAccess(t *testing.T) {
	uc := newTestUC(t, &fakePolicies{policies: []*dockerproxy.Policy{testPolicy("app1")}}, aNode())
	served, err := uc.Sync(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, served)
}

var errDatabaseDown = errors.New("database down")

func TestAFailedReadChangesNothing(t *testing.T) {
	policies := &fakePolicies{policies: []*dockerproxy.Policy{testPolicy("app1")}}
	node := aNode()
	uc := newTestUC(t, policies, node)
	_, err := uc.Sync(context.Background())
	assert.NoError(t, err)

	policies.err = errDatabaseDown
	_, err = uc.Sync(context.Background())
	assert.ErrorIs(t, err, errDatabaseDown)
	assert.Equal(t, []string{"app1"}, uc.host.served(), "the socket stays")

	_, err = uc.Collect(context.Background())
	assert.ErrorIs(t, err, errDatabaseDown)
	assert.Empty(t, node.removed, "without knowing who has access, nothing is anybody's to remove")
}

func TestRemoveAppStopsServingItAndRemovesItsObjects(t *testing.T) {
	node := aNode()
	uc := newTestUC(t, &fakePolicies{policies: []*dockerproxy.Policy{testPolicy("app1")}}, node)
	_, err := uc.Sync(context.Background())
	assert.NoError(t, err)

	removed, err := uc.RemoveApp(context.Background(), "app1")
	assert.NoError(t, err)
	assert.Empty(t, uc.host.served())
	assert.Equal(t, 3, removed.Containers)
}
```

- [ ] **Step 2: Run them to see them fail**

Run: `go test ./hivepaas_app/usecaseagent/dockerapiagentuc/...`
Expected: FAIL, `undefined: UC`.

- [ ] **Step 3: Write the use case**

`hivepaas_app/usecaseagent/dockerapiagentuc/uc.go`:

```go
package dockerapiagentuc

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	syncInterval    = 30 * time.Second
	collectInterval = 10 * time.Minute
)

// UC serves, on this node, the Docker API of every app that has access, and
// removes what the apps' children leave behind.
type UC struct {
	logger           logging.Logger
	db               *database.DB
	dockerManager    docker.Manager
	dockerAPIService dockerapiservice.Service
	host             *socketHost
}

func New(
	logger logging.Logger,
	db *database.DB,
	dockerManager docker.Manager,
	dockerAPIService dockerapiservice.Service,
) (*UC, error) {
	upstream, err := newUpstream()
	if err != nil {
		return nil, err
	}
	return &UC{
		logger:           logger,
		db:               db,
		dockerManager:    dockerManager,
		dockerAPIService: dockerAPIService,
		host:             newSocketHost(logger, dockerManager, upstream),
	}, nil
}

// Sync serves exactly the apps that have access now, and says how many this
// node serves. When the database cannot be read, nothing changes.
func (uc *UC) Sync(ctx context.Context) (int, error) {
	policies, err := uc.dockerAPIService.Policies(ctx, uc.db)
	if err != nil {
		return len(uc.host.served()), hperrors.Wrap(err)
	}
	err = uc.host.reconcile(ctx, policies)
	return len(uc.host.served()), err
}

// RemoveApp stops serving an app and removes what its children left on this
// node.
func (uc *UC) RemoveApp(ctx context.Context, appID string) (Removed, error) {
	uc.host.closeApp(appID)
	s := &sweep{docker: uc.dockerManager, now: time.Now(), only: appID,
		gone: func(id string) bool { return id == appID }}
	return s.run(ctx)
}

// Collect removes what apps' children left behind on this node.
func (uc *UC) Collect(ctx context.Context) (Removed, error) {
	policies, err := uc.dockerAPIService.Policies(ctx, uc.db)
	if err != nil {
		// Without knowing who has access, nothing is anybody's to remove.
		return Removed{}, hperrors.Wrap(err)
	}
	access := make(map[string]bool, len(policies))
	for _, policy := range policies {
		access[policy.AppID] = true
	}
	s := &sweep{docker: uc.dockerManager, now: time.Now(), ageOut: true,
		gone: func(appID string) bool { return !access[appID] }}
	return s.run(ctx)
}

// Run syncs at once and then on a timer, and collects on a longer one, until
// ctx ends.
func (uc *UC) Run(ctx context.Context) {
	uc.syncAndLog(ctx)
	syncTicker := time.NewTicker(syncInterval)
	defer syncTicker.Stop()
	collectTicker := time.NewTicker(collectInterval)
	defer collectTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-syncTicker.C:
			uc.syncAndLog(ctx)
		case <-collectTicker.C:
			uc.collectAndLog(ctx)
		}
	}
}

// Close stops serving every app.
func (uc *UC) Close() {
	uc.host.closeAll()
}

func (uc *UC) syncAndLog(ctx context.Context) {
	if _, err := uc.Sync(ctx); err != nil {
		uc.logger.Errorf("docker api: sync: %v", err)
	}
}

func (uc *UC) collectAndLog(ctx context.Context) {
	removed, err := uc.Collect(ctx)
	if err != nil {
		uc.logger.Errorf("docker api: collect: %v", err)
	}
	if removed != (Removed{}) {
		uc.logger.Infof("docker api: removed %d containers, %d networks and %d volumes left behind",
			removed.Containers, removed.Networks, removed.Volumes)
	}
}
```

- [ ] **Step 4: Write the gRPC side and the wiring**

`hivepaas_app/interface/agent/server/server_docker_api.go`:

```go
package server

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
)

func (s *AgentServer) SyncDockerAPI(
	ctx context.Context,
	_ *agentproto.DockerAPISyncReq,
) (*agentproto.DockerAPISyncResp, error) {
	apps, err := s.dockerAPIAgentUC.Sync(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &agentproto.DockerAPISyncResp{Apps: int32(apps)}, nil //nolint:gosec // a count of apps
}

func (s *AgentServer) RemoveDockerAPIApp(
	ctx context.Context,
	req *agentproto.DockerAPIRemoveAppReq,
) (*agentproto.DockerAPIRemoveAppResp, error) {
	removed, err := s.dockerAPIAgentUC.RemoveApp(ctx, req.GetAppId())
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	//nolint:gosec // counts of what one node held
	return &agentproto.DockerAPIRemoveAppResp{
		Containers: int32(removed.Containers),
		Networks:   int32(removed.Networks),
		Volumes:    int32(removed.Volumes),
	}, nil
}
```

In `hivepaas_app/interface/agent/server/server.go`:
- import `"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/dockerapiagentuc"`;
- embed `agentproto.UnimplementedDockerAPIServiceServer` in `AgentServer` after `UnimplementedContainerServiceServer`;
- add a field `dockerAPIAgentUC *dockerapiagentuc.UC`;
- add a constructor parameter `dockerAPIAgentUC *dockerapiagentuc.UC` after `containerAgentUC`, and set the field.

`hivepaas_app/cmd/internal/docker_api_host.go`:

```go
package internal

import (
	"context"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/dockerapiagentuc"
)

// InitDockerAPIHost serves, on this node, the Docker API of every app given it,
// and removes what their children leave behind. It runs in the agent, the one
// process on every node that holds the node's Docker socket.
func InitDockerAPIHost(lc fx.Lifecycle, uc *dockerapiagentuc.UC, logger logging.Logger) {
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			safego.GoWithLogger(logger, "dockerAPIHost", func() {
				uc.Run(ctx)
			})
			return nil
		},
		OnStop: func(_ context.Context) error {
			cancel()
			uc.Close()
			return nil
		},
	})
}
```

In `hivepaas_app/cmd/agent/main.go`, register the service with the others:

```go
				agentproto.RegisterDockerAPIServiceServer(s, agentSrv)
```

and invoke the host after the gRPC server:

```go
		fx.Invoke(internal.InitDockerAPIHost),
```

In `hivepaas_app/registry/provides.go`:
- import `dockerapiserviceimpl` and `dockerapiagentuc`;
- add `dockerapiserviceimpl.New,` to the services in alphabetical place;
- add `dockerapiagentuc.New,` to `// Use case: Agent`.

`hivepaas_app/cmd/internal/wiring_agent_test.go`:

```go
package internal

import (
	"testing"
	"time"

	"go.uber.org/fx"
	"google.golang.org/grpc"

	agentserver "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/server"
	"github.com/hivepaas/hivepaas/hivepaas_app/registry"
)

// TestAgentFxGraphResolves is TestFxGraphResolves for the agent: it mirrors
// cmd/agent/main.go, and an fx.Invoke added there belongs here too.
func TestAgentFxGraphResolves(t *testing.T) {
	const startTimeout = 60 * time.Second

	provides := make([]any, 0, len(registry.Provides)+1)
	provides = append(provides, func(_ *agentserver.AgentServer) GrpcRegistrar {
		return func(*grpc.Server) {}
	})
	provides = append(provides, registry.Provides...)

	err := fx.ValidateApp(
		fx.StartTimeout(startTimeout),
		fx.Provide(provides...),
		fx.Invoke(InitLogger),
		fx.Invoke(InitConfig),
		fx.Invoke(InitDBConnection),
		fx.Invoke(InitCache),
		fx.Invoke(InitDockerManager),
		fx.Invoke(InitSystemSettings),
		fx.Invoke(InitSystemEventBus),
		fx.Invoke(InitGrpcServer),
		fx.Invoke(InitDockerAPIHost),
	)
	if err != nil {
		t.Fatalf("the agent's fx graph does not resolve: %v", err)
	}
}
```

- [ ] **Step 5: Run the tests, then lint**

Run: `go build ./... && go test ./hivepaas_app/usecaseagent/... ./hivepaas_app/cmd/... ./hivepaas_app/interface/agent/... && golangci-lint run ./hivepaas_app/...`
Expected: `ok`, `0 issues`.

- [ ] **Step 6: Commit**

```bash
git add hivepaas_app
git commit -m "feat(agent): serve the Docker API of apps given it, on every node

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 6: Against a real daemon, and the gates

**Files:**
- Test: `hivepaas_app/usecaseagent/dockerapiagentuc/real_daemon_test.go`

- [ ] **Step 1: Write the test**

`hivepaas_app/usecaseagent/dockerapiagentuc/real_daemon_test.go`:

```go
package dockerapiagentuc

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/services/docker"
)

// TestAgainstARealDaemon runs a child through an app's socket against the
// Docker daemon of this machine, and sweeps it away. It needs a daemon, so it
// runs only when HP_TEST_DOCKER is set, and needs alpine:3 pullable.
func TestAgainstARealDaemon(t *testing.T) {
	if os.Getenv("HP_TEST_DOCKER") == "" {
		t.Skip("set HP_TEST_DOCKER=1 to run against this machine's Docker daemon")
	}
	ctx := context.Background()
	dockerManager, err := docker.New()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	upstream, err := newUpstream()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	id, err := ulid.NewStringULID()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	appID := "itest-" + id
	networkName := "hp-dapi-" + appID
	_, err = dockerManager.NetworkCreate(ctx, networkName, func(opts *client.NetworkCreateOptions) {
		opts.Driver = "bridge"
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	t.Cleanup(func() { _, _ = dockerManager.NetworkRemove(context.Background(), networkName) })

	dir := shortTempDir(t)
	host := newSocketHost(logging.GlobalLogger(), dockerManager, upstream)
	host.socketDir = func(context.Context, *dockerproxy.Policy) (string, error) { return dir, nil }
	t.Cleanup(host.closeAll)
	policy := &dockerproxy.Policy{AppID: appID, Images: []string{"alpine"}, Network: networkName,
		Limits: dockerproxy.Limits{Containers: 2, Memory: 64 << 20, NanoCPUs: 500_000_000}}
	if !assert.NoError(t, host.reconcile(ctx, []*dockerproxy.Policy{policy})) {
		t.FailNow()
	}

	// The app's side: a Docker client that knows nothing of the proxy.
	app, err := client.New(client.WithHost("unix://" + filepath.Join(dir, dockerproxy.SocketFile)))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	pull, err := app.ImagePull(ctx, "alpine:3", client.ImagePullOptions{})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	_, _ = io.Copy(io.Discard, pull)
	_ = pull.Close()

	created, err := app.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{Image: "alpine:3", Cmd: []string{"true"}},
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	_, err = app.ContainerStart(ctx, created.ID, client.ContainerStartOptions{})
	assert.NoError(t, err)
	wait := app.ContainerWait(ctx, created.ID, client.ContainerWaitOptions{})
	select {
	case <-wait.Result:
	case err = <-wait.Error:
		assert.NoError(t, err)
	case <-time.After(time.Minute):
		t.Fatal("the child did not finish")
	}

	inspect, err := dockerManager.ContainerInspect(ctx, created.ID)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, appID, inspect.Container.Config.Labels[dockerproxy.OwnerLabel])
	assert.Equal(t, networkName, string(inspect.Container.HostConfig.NetworkMode))
	assert.Equal(t, int64(64<<20), inspect.Container.HostConfig.Memory)

	_, err = app.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config:     &container.Config{Image: "alpine:3"},
		HostConfig: &container.HostConfig{Privileged: true},
	})
	assert.ErrorContains(t, err, "HostConfig.Privileged is not allowed")

	s := &sweep{docker: dockerManager, now: time.Now(), only: appID,
		gone: func(id string) bool { return id == appID }}
	removed, err := s.run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, removed.Containers)
	_, err = dockerManager.ContainerInspect(ctx, created.ID)
	assert.Error(t, err)
}
```

- [ ] **Step 2: Run it against this machine's daemon**

Run: `HP_TEST_DOCKER=1 go test ./hivepaas_app/usecaseagent/dockerapiagentuc/ -run TestAgainstARealDaemon -v`
Expected: `PASS`. Without `HP_TEST_DOCKER` it reports `SKIP`. On macOS this runs the host natively against Docker Desktop's daemon, which is what the test means to do: the socket is in a temporary directory, not a volume, so nothing depends on the VM's filesystem.

If `ContainerInspect` in `services/docker` returns a different shape than `inspect.Container`, read its signature and adjust the three field reads. What they check stays the same.

- [ ] **Step 3: Whole-repo gates**

Run: `go build ./... && golangci-lint run ./... && go test ./...`
Expected: build succeeds, `0 issues`, all packages `ok`.

- [ ] **Step 4: Commit and merge**

```bash
git add hivepaas_app/usecaseagent/dockerapiagentuc
git commit -m "test(agent): a child through an app's socket against a real daemon

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
git checkout main
git merge --no-ff feat/docker-api-agent -m "Merge branch 'feat/docker-api-agent'

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
go build ./... && go test ./hivepaas_app/usecaseagent/... ./hivepaas_app/service/dockerapiservice/...
git branch -d feat/docker-api-agent
```
