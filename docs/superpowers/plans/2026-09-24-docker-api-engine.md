# Docker API Access - Engine Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A Go package, `hivepaas_app/pkg/dockerproxy`, that serves one app's Docker socket: it lets through what the app's policy allows, rewrites what it must, and refuses the rest with a Docker error.

**Architecture:** An `http.Handler` in front of the Docker daemon. A table of endpoints (§3) decides what may be asked. Handlers judge each request, with lookups to the daemon for ownership, and either forward it through `httputil.ReverseProxy` or answer `403 {"message":"hivepaas: ..."}`. Container create is judged by allowlists of fields (§4) and rewritten: network, storage, limits, owner label, API version. The package knows nothing of HivePaaS's database or services. The agent (plan 2) builds a `Policy` and serves a `Proxy` per app.

**Tech Stack:** Go 1.27, standard library (`net/http`, `httputil`), `pkg/imageref` for image references, testify, `github.com/moby/moby/api` types in tests only.

**Spec:** `docs/superpowers/specs/2026-09-24-docker-api-access-design.md` (§3, §4, §5, §13).

## Global Constraints

- Package path: `github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy`. It imports nothing from `hivepaas_app` except `pkg/imageref`.
- Deny by default. Every endpoint, container field and mount is refused unless a rule in this plan lets it through.
- Refusals are `403` with body `{"message":"hivepaas: <rule>"}`. A failure to judge (daemon down, lookup error) is `502` with the same shape.
- Owner label: `hivepaas.docker-api.app=<app id>`. The app's socket inside its containers: `/var/run/hivepaas/docker.sock`.
- A rewritten create is sent at API version 1.45 or higher (`VolumeOptions.Subpath` needs 1.45).
- Code style: gofmt and gci, 120-column lines, US spelling in comments (`labeled`, `normalization`). Comments say why, in the voice of the surrounding code.
- Gates for every task: `go test ./hivepaas_app/pkg/dockerproxy/...` passes and `golangci-lint run ./hivepaas_app/pkg/dockerproxy/...` reports `0 issues`. Last task: `go build ./...`, `golangci-lint run ./...` (0 issues) and `go test ./...` over the whole repo.
- Work on branch `feat/docker-api-engine`. Every commit ends with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`. At the end, merge into `main` locally and delete the branch. Do not push.

## File Structure

| File | Responsibility |
|---|---|
| `policy.go` | Package doc, `Policy`, `Limits`, `Group`, exported constants |
| `errors.go` | `refusalError`, `refusef` |
| `fields.go` | `isZero`, neutral values, `checkFields`: the deny-by-default rule |
| `images.go` | `matchImage` |
| `daemon.go` | Lookups to the daemon: `get`, `post`, `labelFilter` |
| `proxy.go` | `Proxy`, `New`, `SetPolicy`, `ServeHTTP`, forward and refuse, `Decision` |
| `routes.go` | The endpoint table |
| `core.go` | Ping, version, info, image read and pull handlers; `fetch` |
| `body.go` | Reading, replacing and picking apart JSON bodies |
| `ownership.go` | Containers and execs of the app: per-id checks, list filtering, exec create |
| `objects.go` | Volumes and networks: create, read, delete, list, connect |
| `create.go` | Container create: fields, image, network, count, version |
| `limits.go` | Memory, CPU, pids and swap |
| `storage.go` | Binds and mounts: shared directories, the nested socket, named volumes |
| `*_test.go`, `testdata/` | Fake daemon, test world, recorded client bodies, fuzz test |

---

### Task 1: Policy, refusals, the deny-by-default rule, image matching

**Files:**
- Create: `hivepaas_app/pkg/dockerproxy/policy.go`
- Create: `hivepaas_app/pkg/dockerproxy/errors.go`
- Create: `hivepaas_app/pkg/dockerproxy/fields.go`
- Create: `hivepaas_app/pkg/dockerproxy/images.go`
- Test: `hivepaas_app/pkg/dockerproxy/fields_test.go`
- Test: `hivepaas_app/pkg/dockerproxy/images_test.go`

**Interfaces:**
- Produces: `type Policy struct{AppID, ServiceID string; Images, SharedDirs []string; Network string; Networks []string; SocketVolume string; Allow []Group; Limits Limits}`, `type Limits struct{Containers int; Memory, NanoCPUs int64}`, `type Group string` with `GroupExec`, `GroupFiles`, `GroupVolumes`, `GroupNetworks`, `GroupNestedSocket`; constants `OwnerLabel`, `SocketPath`, `SocketFile`; `refusef(format string, args ...any) error` returning `*refusalError`; `isZero(v any) bool`; `checkFields(where string, obj map[string]any, allowed, nullOnly []string) error`; `matchImage(patterns []string, ref string) bool`.

- [ ] **Step 1: Create the branch**

```bash
cd /Users/tnt/go/src/github.com/hivepaas/hivepaas
git checkout -b feat/docker-api-engine
```

- [ ] **Step 2: Write the failing tests**

`hivepaas_app/pkg/dockerproxy/fields_test.go`:

```go
package dockerproxy

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsZeroIsWhatAClientSendsForUnset(t *testing.T) {
	unset := []any{
		nil, false, "", int64(0), json.Number("0"), json.Number("0.0"), []any{}, map[string]any{},
		map[string]any{"Type": "", "Config": map[string]any{}},
	}
	for _, v := range unset {
		assert.True(t, isZero(v), "%#v", v)
	}
	set := []any{
		true, "host", int64(3), json.Number("1"), json.Number("-1"), []any{"SYS_ADMIN"},
		[]any{json.Number("0")}, map[string]any{"Type": "syslog"},
	}
	for _, v := range set {
		assert.False(t, isZero(v), "%#v", v)
	}
}

func TestCheckFieldsRefusesAFieldItDoesNotKnow(t *testing.T) {
	obj := map[string]any{"Image": "alpine", "Privileged": false, "MaskedPaths": nil}
	require.NoError(t, checkFields("HostConfig", obj, []string{"Image"}, []string{"MaskedPaths"}))

	obj["Privileged"] = true
	err := checkFields("HostConfig", obj, []string{"Image"}, []string{"MaskedPaths"})
	var refusal *refusalError
	require.ErrorAs(t, err, &refusal)
	assert.EqualError(t, err, "HostConfig.Privileged is not allowed")
}

func TestCheckFieldsTakesOnlyNullWhereAnEmptyListMeansSomething(t *testing.T) {
	// An empty MaskedPaths is how --security-opt systempaths=unconfined unmasks /proc.
	err := checkFields("HostConfig", map[string]any{"MaskedPaths": []any{}}, nil, []string{"MaskedPaths"})
	assert.EqualError(t, err, "HostConfig.MaskedPaths must not be set")
}

func TestCheckFieldsKnowsTheCLIsUnsetSwappiness(t *testing.T) {
	assert.NoError(t, checkFields("HostConfig", map[string]any{"MemorySwappiness": json.Number("-1")}, nil, nil))
	assert.EqualError(t,
		checkFields("HostConfig", map[string]any{"MemorySwappiness": json.Number("60")}, nil, nil),
		"HostConfig.MemorySwappiness is not allowed")
}
```

`hivepaas_app/pkg/dockerproxy/images_test.go`:

```go
package dockerproxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchImage(t *testing.T) {
	tests := []struct {
		patterns []string
		ref      string
		want     bool
	}{
		{[]string{"*"}, "anything/at:all", true},
		{[]string{"*"}, "", false},
		{nil, "alpine", false},
		// A pattern without a tag takes every tag and digest of its repository.
		{[]string{"alpine"}, "alpine:3", true},
		{[]string{"alpine"}, "docker.io/library/alpine:3.22", true},
		{[]string{"alpine"}, "alpine@sha256:abc", true},
		// No tag is latest.
		{[]string{"alpine:3"}, "alpine", false},
		{[]string{"alpine:3"}, "alpine:3", true},
		{[]string{"alpine:3.*"}, "alpine:3.22", true},
		{[]string{"autobase/automation:2.11.0"}, "autobase/automation:2.11.0", true},
		{[]string{"autobase/automation:2.11.0"}, "autobase/automation:2.12.0", false},
		{[]string{"autobase/automation:2.11.0"}, "autobase/console:2.11.0", false},
		{[]string{"openruntimes/*"}, "openruntimes/node:v5-22", true},
		{[]string{"openruntimes/*"}, "evil/node:v5-22", false},
		{[]string{"ghcr.io/nextcloud/*"}, "ghcr.io/nextcloud/app:1", true},
		{[]string{"ghcr.io/nextcloud/*"}, "docker.io/nextcloud/app:1", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, matchImage(tt.patterns, tt.ref), "%v %s", tt.patterns, tt.ref)
	}
}
```

- [ ] **Step 3: Run the tests to see them fail**

Run: `go test ./hivepaas_app/pkg/dockerproxy/...`
Expected: FAIL, `undefined: isZero`, `undefined: matchImage`.

- [ ] **Step 4: Write the package**

`hivepaas_app/pkg/dockerproxy/policy.go`:

```go
// Package dockerproxy gives an app the Docker API without the Docker socket.
//
// An app that starts containers of its own - a CI runner, Autobase, Appwrite's
// executor - is installed upstream by mounting /var/run/docker.sock into it,
// which makes it root on its node. The proxy stands between such an app and the
// daemon: it lets through what the app's policy allows, rewrites what it must,
// and refuses the rest with a Docker error the app logs like any other.
//
// Everything is refused unless a rule lets it through. See
// docs/superpowers/specs/2026-09-24-docker-api-access-design.md.
package dockerproxy

// Group names endpoints beyond the core that a policy may allow.
type Group string

const (
	// GroupExec is running commands in the app's children.
	GroupExec Group = "exec"
	// GroupFiles is copying files into and out of the app's children.
	GroupFiles Group = "files"
	// GroupVolumes is volumes of the app's own, and naming them in mounts.
	GroupVolumes Group = "volumes"
	// GroupNetworks is networks of the app's own, and connecting to them.
	GroupNetworks Group = "networks"
	// GroupNestedSocket lets a child mount the app's socket, so that what the
	// child starts goes through this proxy under the same policy.
	GroupNestedSocket Group = "nestedSocket"
)

const (
	// OwnerLabel marks everything the proxy creates for an app, with the app's id.
	OwnerLabel = "hivepaas.docker-api.app"
	// SocketPath is where an app with access finds its socket.
	SocketPath = "/var/run/hivepaas/docker.sock"
	// SocketFile is the socket's name inside the app's socket volume.
	SocketFile = "docker.sock"
)

// Limits bound what an app's children may use.
type Limits struct {
	// Containers is how many children may exist at once, running or not.
	Containers int
	// Memory is the most one child may ask for, in bytes, and what it gets when
	// it asks for nothing.
	Memory int64
	// NanoCPUs is the same for processor time, in billionths of a CPU.
	NanoCPUs int64
}

// Policy is what one app may do through its socket.
type Policy struct {
	// AppID is written into OwnerLabel on everything created for the app.
	AppID string
	// ServiceID is the app's swarm service. Its task containers are the app
	// itself: they hold the mounts shared directories are found in, and they may
	// connect themselves to the app's networks.
	ServiceID string
	// Images are patterns over the images children may run (see matchImage).
	Images []string
	// SharedDirs are directories of the app a child may bind. Each lies on one of
	// the app's volume mounts.
	SharedDirs []string
	// Network is the app's own network, which children join unless they name
	// another the policy allows.
	Network string
	// Networks are other networks children may join, by name.
	Networks []string
	// SocketVolume is the volume holding the app's socket on every node.
	SocketVolume string
	// Allow are the groups of endpoints beyond the core.
	Allow []Group
	Limits Limits
}
```

`hivepaas_app/pkg/dockerproxy/errors.go`:

```go
package dockerproxy

import "fmt"

// refusalError is a request the policy does not allow, as opposed to a failure
// to judge it. It is the app's to fix, and it is answered with 403.
type refusalError struct {
	msg string
}

func (e *refusalError) Error() string {
	return e.msg
}

// refusef returns a refusal saying which rule the request broke.
func refusef(format string, args ...any) error {
	return &refusalError{msg: fmt.Sprintf(format, args...)}
}
```

`hivepaas_app/pkg/dockerproxy/fields.go`:

```go
package dockerproxy

import (
	"encoding/json"
	"maps"
	"slices"
)

// neutralValues are what a client sends for "not set" where zero would mean
// something. The docker CLI sends MemorySwappiness -1.
var neutralValues = map[string][]string{
	"MemorySwappiness": {"-1"},
}

// isZero reports whether v is what a client sends for a field it did not set.
// Go clients marshal whole structs, so most of a HostConfig arrives as null, "",
// 0, false, an empty list, or an object holding only those.
func isZero(v any) bool {
	switch value := v.(type) {
	case nil:
		return true
	case bool:
		return !value
	case string:
		return value == ""
	case int64:
		return value == 0
	case json.Number:
		f, err := value.Float64()
		return err == nil && f == 0
	case []any:
		return len(value) == 0
	case map[string]any:
		for _, inner := range value {
			if !isZero(inner) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func isNeutral(field string, v any) bool {
	n, ok := v.(json.Number)
	return ok && slices.Contains(neutralValues[field], n.String())
}

// checkFields refuses any field of obj outside allowed that carries a value.
// A field in nullOnly must be absent or null, because an empty value there
// means something: an empty MaskedPaths unmasks /proc.
func checkFields(where string, obj map[string]any, allowed, nullOnly []string) error {
	for _, field := range slices.Sorted(maps.Keys(obj)) {
		value := obj[field]
		switch {
		case slices.Contains(nullOnly, field):
			if value != nil {
				return refusef("%s.%s must not be set", where, field)
			}
		case slices.Contains(allowed, field), isZero(value), isNeutral(field, value):
		default:
			return refusef("%s.%s is not allowed", where, field)
		}
	}
	return nil
}
```

`hivepaas_app/pkg/dockerproxy/images.go`:

```go
package dockerproxy

import (
	"path"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/imageref"
)

// matchImage reports whether ref is an image one of the patterns allows.
//
// A pattern is a repository, optionally with a tag, as a person writes it in a
// template: "autobase/automation:2.11.0", "openruntimes/*", "*". Repositories are
// compared after Docker's own normalization, so "alpine" and
// "docker.io/library/alpine" are the same image. A pattern without a tag allows
// every tag and digest of its repository, and "*" allows everything.
func matchImage(patterns []string, ref string) bool {
	if ref == "" {
		return false
	}
	image := imageref.Parse(ref)
	repository := imageref.NormalizeRepository(image.Repository)
	tag := image.Tag
	if tag == "" && image.Digest == "" {
		tag = "latest"
	}
	for _, pattern := range patterns {
		if pattern == "*" {
			return true
		}
		want := imageref.Parse(pattern)
		if ok, _ := path.Match(imageref.NormalizeRepository(want.Repository), repository); !ok {
			continue
		}
		if want.Tag == "" {
			return true
		}
		if ok, _ := path.Match(want.Tag, tag); ok && tag != "" {
			return true
		}
	}
	return false
}
```

- [ ] **Step 5: Run the tests to see them pass, then lint**

Run: `go test ./hivepaas_app/pkg/dockerproxy/... && golangci-lint run ./hivepaas_app/pkg/dockerproxy/...`
Expected: `ok`, `0 issues`.

- [ ] **Step 6: Commit**

```bash
git add hivepaas_app/pkg/dockerproxy
git commit -m "feat(dockerproxy): policy, refusals and the deny-by-default field rule

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 2: The proxy, the endpoint table, and a fake daemon to test it against

**Files:**
- Create: `hivepaas_app/pkg/dockerproxy/daemon.go`
- Create: `hivepaas_app/pkg/dockerproxy/proxy.go`
- Create: `hivepaas_app/pkg/dockerproxy/routes.go`
- Create: `hivepaas_app/pkg/dockerproxy/core.go`
- Modify: `hivepaas_app/pkg/dockerproxy/policy.go` (add `allows`)
- Test: `hivepaas_app/pkg/dockerproxy/fake_daemon_test.go`
- Test: `hivepaas_app/pkg/dockerproxy/world_test.go`
- Test: `hivepaas_app/pkg/dockerproxy/proxy_test.go`

**Interfaces:**
- Consumes: Task 1.
- Produces: `type Options struct{Upstream http.RoundTripper; OnDecision func(Decision)}`, `type Decision struct{AppID, Method, Path string; Allowed bool; Reason string}`, `func New(policy *Policy, opts Options) *Proxy`, `func (p *Proxy) SetPolicy(*Policy)`, `(*Proxy).ServeHTTP`; internal `call{w, r, policy, args}`, `(*Proxy).forward(c, reason)`, `(*Proxy).refuse(c, err)`, `(*Proxy).decide(c, allowed, reason)`, `(*Proxy).fetch(c, out) bool`, `writeJSON`, `writeError`, `on(methods, pattern string, group Group, handle func(*Proxy, *call)) route`, `idPart`, `var routes []route`; `daemon.get(ctx, path, out) (bool, error)`, `errDaemon`, `daemonHost`, `contentTypeJSON`, `versionPrefix`. Test harness: `newWorld(t testing.TB, policy *Policy) *world`, `testPolicy() *Policy`, `(*world).do(t, method, path string, body any) (int, []byte)`, `refusalMessage(t, raw) string`, `(*world).reached(method, suffix string) bool`, `(*world).forwarded(t, method, suffix string) (string, map[string]any)`.

- [ ] **Step 1: Write the fake daemon and the test world**

`hivepaas_app/pkg/dockerproxy/fake_daemon_test.go`:

```go
package dockerproxy

import (
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"slices"
	"strings"
	"sync"
)

type fakeContainer struct {
	labels  map[string]string
	mounts  []map[string]any
	running bool
}

type fakeNetwork struct {
	name   string
	labels map[string]string
}

type recordedRequest struct {
	method string
	path   string
	query  string
	body   []byte
}

// fakeDaemon answers the lookups the proxy makes from a world a test sets up,
// and records every request that reaches it.
type fakeDaemon struct {
	mu         sync.Mutex
	containers map[string]*fakeContainer
	execs      map[string]string
	volumes    map[string]map[string]string
	networks   map[string]*fakeNetwork
	requests   []recordedRequest
}

func (f *fakeDaemon) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	path := versionPrefix.ReplaceAllString(r.URL.Path, "")
	f.mu.Lock()
	f.requests = append(f.requests, recordedRequest{r.Method, r.URL.Path, r.URL.RawQuery, body})
	f.mu.Unlock()
	if r.Method == http.MethodPost && strings.HasSuffix(path, "/attach") {
		f.attach(w)
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	switch {
	case r.Method == http.MethodGet && path == "/containers/json":
		f.listContainers(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/containers/") && strings.HasSuffix(path, "/json"):
		f.inspectContainer(w, strings.TrimSuffix(strings.TrimPrefix(path, "/containers/"), "/json"))
	case r.Method == http.MethodPost && path == "/containers/create":
		writeJSON(w, http.StatusCreated, map[string]any{"Id": "created", "Warnings": []string{}})
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/exec/"):
		f.inspectExec(w, strings.TrimSuffix(strings.TrimPrefix(path, "/exec/"), "/json"))
	case r.Method == http.MethodGet && path == "/volumes":
		f.listVolumes(w)
	case r.Method == http.MethodPost && path == "/volumes/create":
		f.createVolume(w, body)
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/volumes/"):
		f.inspectVolume(w, strings.TrimPrefix(path, "/volumes/"))
	case r.Method == http.MethodGet && path == "/networks":
		f.listNetworks(w)
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/networks/"):
		f.inspectNetwork(w, strings.TrimPrefix(path, "/networks/"))
	case r.Method == http.MethodGet && path == "/info":
		writeJSON(w, http.StatusOK, map[string]any{
			"ServerVersion": "29.8.0", "OSType": "linux",
			"Swarm": map[string]any{"NodeID": "node1"}, "Labels": []string{"zone=a"},
			"RegistryConfig": map[string]any{"Mirrors": []string{}},
		})
	default:
		writeJSON(w, http.StatusOK, map[string]any{})
	}
}

func (f *fakeDaemon) listContainers(w http.ResponseWriter, r *http.Request) {
	all := r.URL.Query().Get("all") == "1"
	var filters map[string][]string
	_ = json.Unmarshal([]byte(r.URL.Query().Get("filters")), &filters)
	out := []map[string]any{}
	for _, id := range slices.Sorted(maps.Keys(f.containers)) {
		c := f.containers[id]
		if (all || c.running) && hasLabels(c.labels, filters["label"]) {
			out = append(out, map[string]any{"Id": id, "Labels": c.labels})
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func hasLabels(labels map[string]string, wanted []string) bool {
	for _, pair := range wanted {
		key, value, _ := strings.Cut(pair, "=")
		if labels[key] != value {
			return false
		}
	}
	return true
}

func (f *fakeDaemon) inspectContainer(w http.ResponseWriter, id string) {
	c, found := f.containers[id]
	if !found {
		writeError(w, http.StatusNotFound, "No such container: "+id)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"Id": id, "Config": map[string]any{"Labels": c.labels}, "HostConfig": map[string]any{"Mounts": c.mounts},
	})
}

func (f *fakeDaemon) inspectExec(w http.ResponseWriter, id string) {
	container, found := f.execs[id]
	if !found {
		writeError(w, http.StatusNotFound, "No such exec instance: "+id)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ID": id, "ContainerID": container})
}

func (f *fakeDaemon) listVolumes(w http.ResponseWriter) {
	volumes := []map[string]any{}
	for _, name := range slices.Sorted(maps.Keys(f.volumes)) {
		volumes = append(volumes, map[string]any{"Name": name, "Labels": f.volumes[name]})
	}
	writeJSON(w, http.StatusOK, map[string]any{"Volumes": volumes, "Warnings": nil})
}

func (f *fakeDaemon) createVolume(w http.ResponseWriter, body []byte) {
	var req struct {
		Name   string
		Labels map[string]string
	}
	_ = json.Unmarshal(body, &req)
	f.volumes[req.Name] = req.Labels
	writeJSON(w, http.StatusCreated, map[string]any{"Name": req.Name, "Labels": req.Labels})
}

func (f *fakeDaemon) inspectVolume(w http.ResponseWriter, name string) {
	labels, found := f.volumes[name]
	if !found {
		writeError(w, http.StatusNotFound, "no such volume")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"Name": name, "Labels": labels})
}

func (f *fakeDaemon) listNetworks(w http.ResponseWriter) {
	out := []map[string]any{}
	for _, id := range slices.Sorted(maps.Keys(f.networks)) {
		n := f.networks[id]
		out = append(out, map[string]any{"Id": id, "Name": n.name, "Labels": n.labels})
	}
	writeJSON(w, http.StatusOK, out)
}

func (f *fakeDaemon) inspectNetwork(w http.ResponseWriter, idOrName string) {
	for id, n := range f.networks {
		if id == idOrName || n.name == idOrName {
			writeJSON(w, http.StatusOK, map[string]any{"Id": id, "Name": n.name, "Labels": n.labels})
			return
		}
	}
	writeError(w, http.StatusNotFound, "network "+idOrName+" not found")
}

// attach answers the way the daemon does: it takes the connection over and
// streams.
func (f *fakeDaemon) attach(w http.ResponseWriter) {
	conn, buf, err := w.(http.Hijacker).Hijack()
	if err != nil {
		return
	}
	defer conn.Close()
	_, _ = buf.WriteString("HTTP/1.1 101 UPGRADED\r\nContent-Type: application/vnd.docker.raw-stream\r\n" +
		"Connection: Upgrade\r\nUpgrade: tcp\r\n\r\nstream-ok\n")
	_ = buf.Flush()
}
```

`hivepaas_app/pkg/dockerproxy/world_test.go`:

```go
package dockerproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// world is a proxy in front of a fake daemon holding:
//   - task1, the app's own task (service svc1), with its volume hp-vol-data
//     mounted at /var/lib/autobase under subpath app1-key;
//   - child1, a container the app started, with exec exec-child;
//   - other1, a container of another app, with exec exec-other;
//   - volumes cache-app1 (the app's), cache-app2 (another app's) and the app's
//     socket volume;
//   - networks hp-dapi-app1 (the app's, n-app), jobnet (created by the app,
//     n-job), othernet (another app's), proj_env_net (the app's env) and
//     hivepaas_net.
type world struct {
	daemon *fakeDaemon
	proxy  *Proxy
	url    string

	mu        sync.Mutex
	decisions []Decision
}

func testPolicy() *Policy {
	return &Policy{
		AppID:        "app1",
		ServiceID:    "svc1",
		Images:       []string{"alpine", "autobase/automation:2.11.0"},
		SharedDirs:   []string{"/var/lib/autobase/ansible"},
		Network:      "hp-dapi-app1",
		Networks:     []string{"proj_env_net"},
		SocketVolume: "hp-dapi-sock-app1",
		Limits:       Limits{Containers: 3, Memory: 1 << 30, NanoCPUs: 1_000_000_000},
	}
}

func newWorld(t testing.TB, policy *Policy) *world {
	t.Helper()
	fake := &fakeDaemon{
		containers: map[string]*fakeContainer{
			"task1": {labels: map[string]string{"com.docker.swarm.service.id": "svc1"}, running: true,
				mounts: []map[string]any{{
					"Type": "volume", "Source": "hp-vol-data", "Target": "/var/lib/autobase",
					"VolumeOptions": map[string]any{"Subpath": "app1-key"},
				}}},
			"child1": {labels: map[string]string{OwnerLabel: "app1"}, running: true},
			"other1": {labels: map[string]string{OwnerLabel: "app2"}, running: true},
		},
		execs: map[string]string{"exec-child": "child1", "exec-other": "other1"},
		volumes: map[string]map[string]string{
			"cache-app1": {OwnerLabel: "app1"}, "cache-app2": {OwnerLabel: "app2"}, "hp-dapi-sock-app1": {},
		},
		networks: map[string]*fakeNetwork{
			"n-app":   {name: "hp-dapi-app1", labels: map[string]string{}},
			"n-job":   {name: "jobnet", labels: map[string]string{OwnerLabel: "app1"}},
			"n-other": {name: "othernet", labels: map[string]string{OwnerLabel: "app2"}},
			"n-env":   {name: "proj_env_net", labels: map[string]string{}},
			"n-hp":    {name: "hivepaas_net", labels: map[string]string{}},
		},
	}
	daemonServer := httptest.NewServer(fake)
	t.Cleanup(daemonServer.Close)
	addr := daemonServer.Listener.Addr().String()
	upstream := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "tcp", addr)
	}}

	w := &world{daemon: fake}
	w.proxy = New(policy, Options{Upstream: upstream, OnDecision: func(d Decision) {
		w.mu.Lock()
		defer w.mu.Unlock()
		w.decisions = append(w.decisions, d)
	}})
	proxyServer := httptest.NewServer(w.proxy)
	t.Cleanup(proxyServer.Close)
	w.url = proxyServer.URL
	return w
}

// do sends a request to the proxy. body is sent as it is when it is bytes, and
// as JSON otherwise.
func (w *world) do(t testing.TB, method, path string, body any) (int, []byte) {
	t.Helper()
	var reader io.Reader
	switch b := body.(type) {
	case nil:
	case []byte:
		reader = bytes.NewReader(b)
	default:
		raw, err := json.Marshal(b)
		require.NoError(t, err)
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(context.Background(), method, w.url+path, reader)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, raw
}

// refusalMessage is the message of an answer the proxy wrote itself.
func refusalMessage(t testing.TB, raw []byte) string {
	t.Helper()
	var answer struct {
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal(raw, &answer), string(raw))
	return answer.Message
}

// reached reports whether a request with that method and path suffix reached
// the daemon.
func (w *world) reached(method, suffix string) bool {
	w.daemon.mu.Lock()
	defer w.daemon.mu.Unlock()
	for _, req := range w.daemon.requests {
		if req.method == method && strings.HasSuffix(req.path, suffix) {
			return true
		}
	}
	return false
}

// forwarded returns the path and decoded body of the last request with that
// method and path suffix that reached the daemon.
func (w *world) forwarded(t testing.TB, method, suffix string) (string, map[string]any) {
	t.Helper()
	w.daemon.mu.Lock()
	defer w.daemon.mu.Unlock()
	for i := len(w.daemon.requests) - 1; i >= 0; i-- {
		req := w.daemon.requests[i]
		if req.method != method || !strings.HasSuffix(req.path, suffix) {
			continue
		}
		body := map[string]any{}
		if len(req.body) > 0 {
			decoder := json.NewDecoder(bytes.NewReader(req.body))
			decoder.UseNumber()
			require.NoError(t, decoder.Decode(&body))
		}
		return req.path, body
	}
	t.Fatalf("no %s ...%s reached the daemon", method, suffix)
	return "", nil
}
```

- [ ] **Step 2: Write the failing tests**

`hivepaas_app/pkg/dockerproxy/proxy_test.go`:

```go
package dockerproxy

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefusesEveryEndpointOutsideTheTable(t *testing.T) {
	w := newWorld(t, testPolicy())
	for _, endpoint := range []struct{ method, path string }{
		{http.MethodPost, "/v1.51/build"},
		{http.MethodPost, "/v1.51/session"},
		{http.MethodGet, "/v1.51/events"},
		{http.MethodPost, "/v1.51/containers/child1/update"},
		{http.MethodPost, "/v1.51/containers/prune"},
		{http.MethodGet, "/v1.51/containers/child1/export"},
		{http.MethodPost, "/v1.51/commit"},
		{http.MethodPost, "/v1.51/swarm/init"},
		{http.MethodGet, "/v1.51/services"},
		{http.MethodGet, "/v1.51/secrets"},
		{http.MethodPost, "/v1.51/plugins/pull"},
		{http.MethodGet, "/v1.51/system/df"},
		{http.MethodPost, "/v1.51/images/load"},
	} {
		status, raw := w.do(t, endpoint.method, endpoint.path, nil)
		assert.Equal(t, http.StatusForbidden, status, endpoint.path)
		assert.Contains(t, refusalMessage(t, raw), "is not an endpoint this app may use", endpoint.path)
	}
	assert.Empty(t, w.daemon.requests)
}

func TestRefusesAPathThatIsNotInPlainForm(t *testing.T) {
	// A path the proxy reads one way and the daemon another would be judged by the
	// wrong rule.
	w := newWorld(t, testPolicy())
	for _, path := range []string{
		"/v1.51/containers/..%2Fbuild/json",
		"/v1.51/containers//json",
		"/v1.51/containers/child1/../../build",
	} {
		status, raw := w.do(t, http.MethodGet, path, nil)
		assert.Equal(t, http.StatusForbidden, status, path)
		assert.Contains(t, refusalMessage(t, raw), "is not in plain form", path)
	}
}

func TestPassesPingVersionAndImageReads(t *testing.T) {
	w := newWorld(t, testPolicy())
	for _, endpoint := range []struct{ method, path string }{
		{http.MethodGet, "/_ping"},
		{http.MethodHead, "/_ping"},
		{http.MethodGet, "/v1.51/version"},
		{http.MethodGet, "/v1.51/images/json"},
		{http.MethodGet, "/v1.51/images/autobase/automation:2.11.0/json"},
	} {
		status, _ := w.do(t, endpoint.method, endpoint.path, nil)
		assert.Equal(t, http.StatusOK, status, endpoint.path)
	}
}

func TestInfoLeavesOutTheCluster(t *testing.T) {
	w := newWorld(t, testPolicy())
	status, raw := w.do(t, http.MethodGet, "/v1.51/info", nil)
	require.Equal(t, http.StatusOK, status)
	var info map[string]any
	require.NoError(t, json.Unmarshal(raw, &info))
	assert.Equal(t, "29.8.0", info["ServerVersion"])
	assert.NotContains(t, info, "Swarm")
	assert.NotContains(t, info, "Labels")
	assert.NotContains(t, info, "RegistryConfig")
}

func TestPullTakesOnlyThePolicysImages(t *testing.T) {
	w := newWorld(t, testPolicy())
	status, _ := w.do(t, http.MethodPost, "/v1.51/images/create?fromImage=alpine&tag=3", nil)
	assert.Equal(t, http.StatusOK, status)
	path, _ := w.forwarded(t, http.MethodPost, "/images/create")
	assert.Equal(t, "/v1.51/images/create", path)

	status, raw := w.do(t, http.MethodPost, "/v1.51/images/create?fromImage=busybox&tag=1", nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: image busybox:1 is not allowed", refusalMessage(t, raw))

	status, raw = w.do(t, http.MethodPost, "/v1.51/images/create?fromSrc=http://example.com/rootfs.tar", nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: importing an image is not allowed", refusalMessage(t, raw))
}

func TestDecisionsAreReported(t *testing.T) {
	w := newWorld(t, testPolicy())
	w.do(t, http.MethodGet, "/_ping", nil)
	w.do(t, http.MethodPost, "/v1.51/build", nil)

	w.mu.Lock()
	defer w.mu.Unlock()
	require.Len(t, w.decisions, 2)
	assert.Equal(t, Decision{AppID: "app1", Method: http.MethodGet, Path: "/_ping", Allowed: true,
		Reason: "read-only endpoint"}, w.decisions[0])
	assert.False(t, w.decisions[1].Allowed)
	assert.Equal(t, "POST /build is not an endpoint this app may use", w.decisions[1].Reason)
}

func TestSetPolicyAppliesToTheNextRequest(t *testing.T) {
	w := newWorld(t, testPolicy())
	status, _ := w.do(t, http.MethodPost, "/v1.51/images/create?fromImage=busybox&tag=1", nil)
	assert.Equal(t, http.StatusForbidden, status)

	policy := testPolicy()
	policy.Images = []string{"busybox"}
	w.proxy.SetPolicy(policy)
	status, _ = w.do(t, http.MethodPost, "/v1.51/images/create?fromImage=busybox&tag=1", nil)
	assert.Equal(t, http.StatusOK, status)
}

func TestAnUnreachableDaemonIsNotARefusal(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	require.NoError(t, ln.Close())
	upstream := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "tcp", addr)
	}}
	server := httptest.NewServer(New(testPolicy(), Options{Upstream: upstream}))
	defer server.Close()

	w := &world{url: server.URL}
	status, raw := w.do(t, http.MethodGet, "/_ping", nil)
	assert.Equal(t, http.StatusBadGateway, status)
	assert.Contains(t, refusalMessage(t, raw), "hivepaas: the Docker daemon did not answer")
}
```

- [ ] **Step 3: Run the tests to see them fail**

Run: `go test ./hivepaas_app/pkg/dockerproxy/...`
Expected: FAIL, `undefined: New`, `undefined: versionPrefix`, `undefined: writeJSON`.

- [ ] **Step 4: Write the proxy**

Add to `hivepaas_app/pkg/dockerproxy/policy.go`, after the `Policy` type, with `"slices"` imported:

```go
func (p *Policy) allows(group Group) bool {
	return group == "" || slices.Contains(p.Allow, group)
}
```

`hivepaas_app/pkg/dockerproxy/daemon.go`:

```go
package dockerproxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const (
	// daemonHost is the host requests to the daemon carry. The transport dials the
	// daemon's socket whatever the address, so it only has to be a valid name.
	daemonHost = "docker"

	contentTypeJSON = "application/json"
)

// errDaemon is the daemon answering a lookup in a way the proxy cannot use.
var errDaemon = errors.New("unexpected answer from the Docker daemon")

// daemon asks the Docker daemon what the proxy needs to know to judge a request.
type daemon struct {
	client *http.Client
}

// get fetches path and decodes the answer into out. Not found is not an error:
// the bool says whether there was anything.
func (d *daemon) get(ctx context.Context, path string, out any) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+daemonHost+path, nil)
	if err != nil {
		return false, fmt.Errorf("GET %s: %w", path, err)
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("GET %s: %w", path, err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		if err = json.NewDecoder(resp.Body).Decode(out); err != nil {
			return false, fmt.Errorf("GET %s: %w", path, err)
		}
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("%w: GET %s answered %d", errDaemon, path, resp.StatusCode)
	}
}
```

`hivepaas_app/pkg/dockerproxy/proxy.go`:

```go
package dockerproxy

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httputil"
	"path"
	"regexp"
	"slices"
	"sync/atomic"
)

// flushImmediately makes the forwarder write each chunk as it arrives: logs,
// attach and pull progress are streams.
const flushImmediately = -1

// versionPrefix is the API version a client may put in front of any path.
var versionPrefix = regexp.MustCompile(`^/v[0-9]+\.[0-9]+`)

// Decision is what the proxy did with one request.
type Decision struct {
	AppID   string
	Method  string
	Path    string
	Allowed bool
	// Reason is the rule that let the request through, or the one it broke.
	Reason string
}

// Options are what a proxy needs besides its policy.
type Options struct {
	// Upstream reaches the Docker daemon. Requests are sent to http://docker/...,
	// so it is a transport that dials the daemon's socket whatever the address.
	Upstream http.RoundTripper
	// OnDecision, when set, is told about every request.
	OnDecision func(Decision)
}

// Proxy serves one app's socket.
type Proxy struct {
	policy     atomic.Pointer[Policy]
	daemon     *daemon
	forwarder  *httputil.ReverseProxy
	onDecision func(Decision)
}

// New returns a proxy enforcing policy.
func New(policy *Policy, opts Options) *Proxy {
	p := &Proxy{
		daemon:     &daemon{client: &http.Client{Transport: opts.Upstream}},
		onDecision: opts.OnDecision,
	}
	p.policy.Store(policy)
	p.forwarder = &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.Out.URL.Scheme, r.Out.URL.Host, r.Out.Host = "http", daemonHost, daemonHost
		},
		Transport:     opts.Upstream,
		FlushInterval: flushImmediately,
		ErrorHandler: func(w http.ResponseWriter, _ *http.Request, err error) {
			writeError(w, http.StatusBadGateway, "hivepaas: the Docker daemon did not answer: "+err.Error())
		},
	}
	return p
}

// SetPolicy replaces the policy. The next request is judged by the new one; a
// request in flight finishes under the old.
func (p *Proxy) SetPolicy(policy *Policy) {
	p.policy.Store(policy)
}

// call is one request being judged.
type call struct {
	w      http.ResponseWriter
	r      *http.Request
	policy *Policy
	// args are the parts of the path the route captured: an id, a name.
	args []string
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := &call{w: w, r: r, policy: p.policy.Load()}
	endpoint := versionPrefix.ReplaceAllString(r.URL.Path, "")
	// A path the proxy could read one way and the daemon another would be judged
	// by the wrong rule, so only the plain form is taken.
	if r.URL.RawPath != "" || endpoint != path.Clean(endpoint) {
		p.refuse(c, refusef("the path %s is not in plain form", r.URL.Path))
		return
	}
	for _, rt := range routes {
		if !slices.Contains(rt.methods, r.Method) {
			continue
		}
		match := rt.pattern.FindStringSubmatch(endpoint)
		if match == nil {
			continue
		}
		if !c.policy.allows(rt.group) {
			p.refuse(c, refusef("%s is not allowed for this app", rt.group))
			return
		}
		c.args = match[1:]
		rt.handle(p, c)
		return
	}
	p.refuse(c, refusef("%s %s is not an endpoint this app may use", r.Method, endpoint))
}

func (p *Proxy) forward(c *call, reason string) {
	p.decide(c, true, reason)
	p.forwarder.ServeHTTP(c.w, c.r)
}

// refuse answers the way the daemon answers a request it will not do, so that
// the app reports it the way it reports any daemon error.
func (p *Proxy) refuse(c *call, err error) {
	status := http.StatusBadGateway
	var refusal *refusalError
	if errors.As(err, &refusal) {
		status = http.StatusForbidden
	}
	p.decide(c, false, err.Error())
	writeError(c.w, status, "hivepaas: "+err.Error())
}

func (p *Proxy) decide(c *call, allowed bool, reason string) {
	if p.onDecision == nil {
		return
	}
	p.onDecision(Decision{
		AppID: c.policy.AppID, Method: c.r.Method, Path: c.r.URL.Path, Allowed: allowed, Reason: reason,
	})
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"message": msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
```

`hivepaas_app/pkg/dockerproxy/routes.go`:

```go
package dockerproxy

import (
	"regexp"
	"strings"
)

type route struct {
	methods []string
	pattern *regexp.Regexp
	group   Group
	handle  func(p *Proxy, c *call)
}

func on(methods, pattern string, group Group, handle func(*Proxy, *call)) route {
	return route{
		methods: strings.Split(methods, ","),
		pattern: regexp.MustCompile("^" + pattern + "$"),
		group:   group,
		handle:  handle,
	}
}

// idPart captures an id or a name in a path.
const idPart = `([^/]+)`

// routes are every endpoint an app may reach, after its API version is taken
// off. Anything else is refused.
var routes = []route{
	on("GET,HEAD", `/_ping`, "", (*Proxy).pass),
	on("GET", `/version`, "", (*Proxy).pass),
	on("GET", `/info`, "", (*Proxy).info),
	on("GET", `/images/json`, "", (*Proxy).pass),
	on("GET", `/images/(.+)/json`, "", (*Proxy).pass),
	on("POST", `/images/create`, "", (*Proxy).pull),
}
```

`hivepaas_app/pkg/dockerproxy/core.go`:

```go
package dockerproxy

import (
	"fmt"
	"net/http"
	"strings"
)

// infoClusterFields are what /info says about the cluster and the node's place
// in it, rather than about the engine an app is talking to.
var infoClusterFields = []string{"Swarm", "Labels", "RegistryConfig"}

func (p *Proxy) pass(c *call) {
	p.forward(c, "read-only endpoint")
}

func (p *Proxy) info(c *call) {
	var answer map[string]any
	if !p.fetch(c, &answer) {
		return
	}
	for _, field := range infoClusterFields {
		delete(answer, field)
	}
	p.decide(c, true, "info, without the cluster")
	writeJSON(c.w, http.StatusOK, answer)
}

func (p *Proxy) pull(c *call) {
	query := c.r.URL.Query()
	if query.Get("fromSrc") != "" {
		p.refuse(c, refusef("importing an image is not allowed"))
		return
	}
	ref := query.Get("fromImage")
	if tag := query.Get("tag"); tag != "" {
		separator := ":"
		if strings.HasPrefix(tag, "sha256:") {
			separator = "@"
		}
		ref += separator + tag
	}
	if !matchImage(c.policy.Images, ref) {
		p.refuse(c, refusef("image %s is not allowed", ref))
		return
	}
	p.forward(c, "image "+ref)
}

// fetch reads what the client asked for from the daemon, for the proxy to trim
// before answering. It answers the client itself when that fails.
func (p *Proxy) fetch(c *call, out any) bool {
	found, err := p.daemon.get(c.r.Context(), c.r.URL.RequestURI(), out)
	if err == nil && !found {
		err = fmt.Errorf("%w: GET %s answered 404", errDaemon, c.r.URL.Path)
	}
	if err != nil {
		p.refuse(c, err)
		return false
	}
	return true
}
```

- [ ] **Step 5: Run the tests to see them pass, then lint**

Run: `go test ./hivepaas_app/pkg/dockerproxy/... && golangci-lint run ./hivepaas_app/pkg/dockerproxy/...`
Expected: `ok`, `0 issues`.

- [ ] **Step 6: Commit**

```bash
git add hivepaas_app/pkg/dockerproxy
git commit -m "feat(dockerproxy): serve an endpoint table in front of the daemon

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 3: Containers and execs of the app

**Files:**
- Create: `hivepaas_app/pkg/dockerproxy/body.go`
- Create: `hivepaas_app/pkg/dockerproxy/ownership.go`
- Modify: `hivepaas_app/pkg/dockerproxy/routes.go` (add container and exec routes)
- Test: `hivepaas_app/pkg/dockerproxy/ownership_test.go`

**Interfaces:**
- Consumes: Task 2.
- Produces: `readBody(r *http.Request) (map[string]any, error)`, `setBody(r *http.Request, body map[string]any)`, `object(v any) map[string]any`, `text(v any) string`, `fieldLabels`, `(*Proxy).containerLabels(ctx, id) (map[string]string, error)`, `(*Proxy).child(ctx, policy, id) error`, `ownedBy(labels any, appID string) bool`, `keep(items []map[string]any, ok func(map[string]any) bool) []map[string]any`.

- [ ] **Step 1: Write the failing tests**

`hivepaas_app/pkg/dockerproxy/ownership_test.go`:

```go
package dockerproxy

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTheAppsChildrenAreReachable(t *testing.T) {
	w := newWorld(t, testPolicy())
	for _, endpoint := range []struct{ method, path string }{
		{http.MethodGet, "/v1.51/containers/child1/json"},
		{http.MethodGet, "/v1.51/containers/child1/logs?stdout=1"},
		{http.MethodPost, "/v1.51/containers/child1/start"},
		{http.MethodPost, "/v1.51/containers/child1/wait"},
		{http.MethodDelete, "/v1.51/containers/child1?force=1"},
	} {
		status, raw := w.do(t, endpoint.method, endpoint.path, nil)
		assert.Equal(t, http.StatusOK, status, "%s %s", endpoint.path, raw)
	}
}

func TestOtherContainersAreNot(t *testing.T) {
	w := newWorld(t, testPolicy())
	tests := map[string]string{
		// Another app's child, and the app's own task: the app may run containers,
		// not reach into the ones HivePaaS runs.
		"/v1.51/containers/other1/json": "container other1 is not one this app started",
		"/v1.51/containers/task1/json":  "container task1 is not one this app started",
		"/v1.51/containers/missing/json": "container missing does not exist",
	}
	for path, message := range tests {
		status, raw := w.do(t, http.MethodGet, path, nil)
		assert.Equal(t, http.StatusForbidden, status, path)
		assert.Equal(t, "hivepaas: "+message, refusalMessage(t, raw), path)
	}
	status, _ := w.do(t, http.MethodDelete, "/v1.51/containers/other1", nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.False(t, w.reached(http.MethodDelete, "/containers/other1"))
}

func TestTheContainerListShowsOnlyTheAppsChildren(t *testing.T) {
	w := newWorld(t, testPolicy())
	status, raw := w.do(t, http.MethodGet, "/v1.51/containers/json?all=1", nil)
	require.Equal(t, http.StatusOK, status)
	var containers []struct {
		ID string `json:"Id"`
	}
	require.NoError(t, json.Unmarshal(raw, &containers))
	require.Len(t, containers, 1)
	assert.Equal(t, "child1", containers[0].ID)
}

func TestAttachStreamsThroughTheProxy(t *testing.T) {
	w := newWorld(t, testPolicy())
	conn, err := net.Dial("tcp", strings.TrimPrefix(w.url, "http://"))
	require.NoError(t, err)
	defer conn.Close()
	_, err = conn.Write([]byte("POST /v1.51/containers/child1/attach?stream=1&stdout=1 HTTP/1.1\r\n" +
		"Host: docker\r\nConnection: Upgrade\r\nUpgrade: tcp\r\n\r\n"))
	require.NoError(t, err)
	require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)))
	got, _ := io.ReadAll(conn)
	assert.Contains(t, string(got), "101")
	assert.Contains(t, string(got), "stream-ok")
}

func TestGroupsThePolicyDoesNotAllowAreRefused(t *testing.T) {
	w := newWorld(t, testPolicy())
	status, raw := w.do(t, http.MethodPost, "/v1.51/containers/child1/exec", map[string]any{"Cmd": []string{"sh"}})
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: exec is not allowed for this app", refusalMessage(t, raw))

	status, raw = w.do(t, http.MethodGet, "/v1.51/containers/child1/archive?path=/", nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: files is not allowed for this app", refusalMessage(t, raw))
}

func TestExecRunsOnlyInTheAppsChildren(t *testing.T) {
	policy := testPolicy()
	policy.Allow = []Group{GroupExec, GroupFiles}
	w := newWorld(t, policy)

	status, raw := w.do(t, http.MethodPost, "/v1.51/containers/child1/exec",
		map[string]any{"Cmd": []string{"sh", "-c", "id"}, "AttachStdout": true, "Privileged": false})
	assert.Equal(t, http.StatusOK, status, string(raw))

	status, raw = w.do(t, http.MethodPost, "/v1.51/containers/child1/exec",
		map[string]any{"Cmd": []string{"sh"}, "Privileged": true})
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: exec.Privileged is not allowed", refusalMessage(t, raw))

	status, _ = w.do(t, http.MethodPost, "/v1.51/containers/other1/exec", map[string]any{"Cmd": []string{"sh"}})
	assert.Equal(t, http.StatusForbidden, status)

	status, _ = w.do(t, http.MethodPost, "/v1.51/exec/exec-child/start", map[string]any{"Detach": false})
	assert.Equal(t, http.StatusOK, status)
	status, raw = w.do(t, http.MethodPost, "/v1.51/exec/exec-other/start", map[string]any{"Detach": false})
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: container other1 is not one this app started", refusalMessage(t, raw))
	status, raw = w.do(t, http.MethodGet, "/v1.51/exec/missing/json", nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: exec missing does not exist", refusalMessage(t, raw))

	status, _ = w.do(t, http.MethodPut, "/v1.51/containers/child1/archive?path=/tmp", []byte("tar"))
	assert.Equal(t, http.StatusOK, status)
}
```

- [ ] **Step 2: Run the tests to see them fail**

Run: `go test ./hivepaas_app/pkg/dockerproxy/...`
Expected: FAIL: the container paths answer 403 `is not an endpoint this app may use`.

- [ ] **Step 3: Write the handlers**

`hivepaas_app/pkg/dockerproxy/body.go`:

```go
package dockerproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// maxBody bounds what the proxy reads to judge a request. A container create is
// a few kilobytes; a body near this size is not one.
const maxBody = 1 << 20

// readBody decodes a request's JSON object. Numbers stay json.Number, so that
// what is passed on is what was sent.
func readBody(r *http.Request) (map[string]any, error) {
	raw, err := io.ReadAll(io.LimitReader(r.Body, maxBody+1))
	if err != nil {
		return nil, fmt.Errorf("reading the request: %w", err)
	}
	if len(raw) > maxBody {
		return nil, refusef("the request body is larger than %d bytes", maxBody)
	}
	var body map[string]any
	if len(bytes.TrimSpace(raw)) > 0 {
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		if err = decoder.Decode(&body); err != nil {
			return nil, refusef("the request body is not a JSON object")
		}
	}
	if body == nil {
		body = map[string]any{}
	}
	return body, nil
}

// setBody replaces the request's body with body, as JSON.
func setBody(r *http.Request, body map[string]any) {
	// A map of what readBody decoded, and of the strings, numbers and maps the
	// proxy put in it, always marshals.
	raw, _ := json.Marshal(body)
	r.Body = io.NopCloser(bytes.NewReader(raw))
	r.ContentLength = int64(len(raw))
	r.TransferEncoding = nil
	r.Header.Set("Content-Length", strconv.Itoa(len(raw)))
	r.Header.Set("Content-Type", contentTypeJSON)
}

// object is v as a JSON object, or nil when it is not one.
func object(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

// text is v as a string, or "" when it is not one.
func text(v any) string {
	s, _ := v.(string)
	return s
}
```

`hivepaas_app/pkg/dockerproxy/ownership.go`:

```go
package dockerproxy

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// fieldLabels is where Docker keeps an object's labels.
const fieldLabels = "Labels"

// execFields are what an exec in a child may ask for. Privileged is not among
// them.
var execFields = []string{"User", "Cmd", "Env", "WorkingDir", "Tty", "AttachStdin", "AttachStdout",
	"AttachStderr", "DetachKeys", "ConsoleSize"}

type containerInspect struct {
	Config struct {
		Labels map[string]string
	}
}

// containerLabels returns a container's labels, refusing one that does not exist.
func (p *Proxy) containerLabels(ctx context.Context, id string) (map[string]string, error) {
	var info containerInspect
	found, err := p.daemon.get(ctx, "/containers/"+url.PathEscape(id)+"/json", &info)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, refusef("container %s does not exist", id)
	}
	return info.Config.Labels, nil
}

// child checks that a container is one the app started.
func (p *Proxy) child(ctx context.Context, policy *Policy, id string) error {
	labels, err := p.containerLabels(ctx, id)
	if err != nil {
		return err
	}
	if labels[OwnerLabel] != policy.AppID {
		return refusef("container %s is not one this app started", id)
	}
	return nil
}

// ownedBy reports whether labels, as a list answer carries them, mark the app's.
func ownedBy(labels any, appID string) bool {
	return text(object(labels)[OwnerLabel]) == appID
}

// keep returns the items ok allows, and an empty list rather than none.
func keep(items []map[string]any, ok func(map[string]any) bool) []map[string]any {
	kept := []map[string]any{}
	for _, item := range items {
		if ok(item) {
			kept = append(kept, item)
		}
	}
	return kept
}

func (p *Proxy) listContainers(c *call) {
	var items []map[string]any
	if !p.fetch(c, &items) {
		return
	}
	kept := keep(items, func(item map[string]any) bool { return ownedBy(item[fieldLabels], c.policy.AppID) })
	p.decide(c, true, fmt.Sprintf("%d of %d containers are the app's", len(kept), len(items)))
	writeJSON(c.w, http.StatusOK, kept)
}

func (p *Proxy) onChild(c *call) {
	if err := p.child(c.r.Context(), c.policy, c.args[0]); err != nil {
		p.refuse(c, err)
		return
	}
	p.forward(c, "a child of the app")
}

func (p *Proxy) onExec(c *call) {
	ctx := c.r.Context()
	var exec struct {
		ContainerID string
	}
	found, err := p.daemon.get(ctx, "/exec/"+url.PathEscape(c.args[0])+"/json", &exec)
	if err == nil && !found {
		err = refusef("exec %s does not exist", c.args[0])
	}
	if err == nil {
		err = p.child(ctx, c.policy, exec.ContainerID)
	}
	if err != nil {
		p.refuse(c, err)
		return
	}
	p.forward(c, "an exec in a child of the app")
}

func (p *Proxy) execCreate(c *call) {
	err := p.child(c.r.Context(), c.policy, c.args[0])
	var body map[string]any
	if err == nil {
		body, err = readBody(c.r)
	}
	if err == nil {
		err = checkFields("exec", body, execFields, nil)
	}
	if err != nil {
		p.refuse(c, err)
		return
	}
	setBody(c.r, body)
	p.forward(c, "an exec in a child of the app")
}
```

In `hivepaas_app/pkg/dockerproxy/routes.go`, append to `routes` after the `/images/create` entry:

```go
	on("GET", `/containers/json`, "", (*Proxy).listContainers),
	on("GET", `/containers/`+idPart+`/(?:json|logs|stats|top)`, "", (*Proxy).onChild),
	on("POST", `/containers/`+idPart+`/(?:start|stop|kill|wait|restart|resize|attach)`, "", (*Proxy).onChild),
	on("DELETE", `/containers/`+idPart, "", (*Proxy).onChild),
	on("GET,PUT,HEAD", `/containers/`+idPart+`/archive`, GroupFiles, (*Proxy).onChild),
	on("POST", `/containers/`+idPart+`/exec`, GroupExec, (*Proxy).execCreate),
	on("POST", `/exec/`+idPart+`/(?:start|resize)`, GroupExec, (*Proxy).onExec),
	on("GET", `/exec/`+idPart+`/json`, GroupExec, (*Proxy).onExec),
```

- [ ] **Step 4: Run the tests to see them pass, then lint**

Run: `go test ./hivepaas_app/pkg/dockerproxy/... && golangci-lint run ./hivepaas_app/pkg/dockerproxy/...`
Expected: `ok`, `0 issues`.

- [ ] **Step 5: Commit**

```bash
git add hivepaas_app/pkg/dockerproxy
git commit -m "feat(dockerproxy): reach only the containers and execs the app started

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 4: Volumes and networks of the app

**Files:**
- Create: `hivepaas_app/pkg/dockerproxy/objects.go`
- Create: `hivepaas_app/pkg/dockerproxy/testdata/cli-network-create.json`
- Create: `hivepaas_app/pkg/dockerproxy/testdata/cli-volume-create-bind-opts.json`
- Modify: `hivepaas_app/pkg/dockerproxy/body.go` (add `list`)
- Modify: `hivepaas_app/pkg/dockerproxy/policy.go` (add `joinable`)
- Modify: `hivepaas_app/pkg/dockerproxy/routes.go` (add volume and network routes)
- Modify: `hivepaas_app/pkg/dockerproxy/world_test.go` (add `fixture`)
- Test: `hivepaas_app/pkg/dockerproxy/objects_test.go`

**Interfaces:**
- Consumes: Task 3.
- Produces: `list(v any) []any`, `(*Policy).joinable(name string) bool`, `ownLabels(labels map[string]any, appID string) map[string]any`, `endpointFields`, `(*Proxy).usableNetwork(ctx, policy, idOrName string) error`, `(*Proxy).childOrSelf(ctx, policy, id string) error`, `serviceIDLabel`; test helper `fixture(t testing.TB, name string) map[string]any`.

- [ ] **Step 1: Add the recorded bodies**

Recorded from the docker CLI 29.8 (`docker network create jobnet` and `docker volume create --opt type=none --opt o=bind --opt device=/ evil`).

`hivepaas_app/pkg/dockerproxy/testdata/cli-network-create.json`:

```json
{"Name":"jobnet","Driver":"bridge","Scope":"","IPAM":{"Driver":"default","Options":{},"Config":[]},"Internal":false,"Attachable":false,"Ingress":false,"ConfigOnly":false,"ConfigFrom":null,"Options":{},"Labels":{}}
```

`hivepaas_app/pkg/dockerproxy/testdata/cli-volume-create-bind-opts.json`:

```json
{"Driver":"local","DriverOpts":{"device":"/","o":"bind","type":"none"},"Name":"evil"}
```

Add to `hivepaas_app/pkg/dockerproxy/world_test.go`, with `"os"` and `"path/filepath"` imported:

```go
// fixture decodes a request body recorded from a real client.
func fixture(t testing.TB, name string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)
	body := map[string]any{}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&body))
	return body
}
```

- [ ] **Step 2: Write the failing tests**

`hivepaas_app/pkg/dockerproxy/objects_test.go`:

```go
package dockerproxy

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func volumesAndNetworks() *Policy {
	policy := testPolicy()
	policy.Allow = []Group{GroupVolumes, GroupNetworks}
	return policy
}

func TestVolumesAreTheAppsOwn(t *testing.T) {
	w := newWorld(t, volumesAndNetworks())

	status, raw := w.do(t, http.MethodGet, "/v1.51/volumes", nil)
	require.Equal(t, http.StatusOK, status)
	var answer struct {
		Volumes []struct{ Name string }
	}
	require.NoError(t, json.Unmarshal(raw, &answer))
	require.Len(t, answer.Volumes, 1)
	assert.Equal(t, "cache-app1", answer.Volumes[0].Name)

	status, _ = w.do(t, http.MethodPost, "/v1.51/volumes/create", map[string]any{
		"Name": "cache2", "Labels": map[string]any{"keep": "yes", OwnerLabel: "app2"},
	})
	assert.Equal(t, http.StatusCreated, status)
	_, body := w.forwarded(t, http.MethodPost, "/volumes/create")
	assert.Equal(t, map[string]any{"keep": "yes", OwnerLabel: "app1"}, body["Labels"])

	status, raw = w.do(t, http.MethodPost, "/v1.51/volumes/create", fixture(t, "cli-volume-create-bind-opts.json"))
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: Volume.DriverOpts is not allowed", refusalMessage(t, raw))

	status, raw = w.do(t, http.MethodPost, "/v1.51/volumes/create", map[string]any{"Name": "x", "Driver": "nfs"})
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: volume driver nfs is not allowed", refusalMessage(t, raw))

	status, _ = w.do(t, http.MethodGet, "/v1.51/volumes/cache-app1", nil)
	assert.Equal(t, http.StatusOK, status)
	status, raw = w.do(t, http.MethodDelete, "/v1.51/volumes/cache-app2", nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: volume cache-app2 is not one this app created", refusalMessage(t, raw))
	status, _ = w.do(t, http.MethodDelete, "/v1.51/volumes/hp-dapi-sock-app1", nil)
	assert.Equal(t, http.StatusForbidden, status)
	status, _ = w.do(t, http.MethodDelete, "/v1.51/volumes/cache-app1", nil)
	assert.Equal(t, http.StatusOK, status)
}

func TestNetworksAreTheAppsOwn(t *testing.T) {
	w := newWorld(t, volumesAndNetworks())

	status, raw := w.do(t, http.MethodPost, "/v1.51/networks/create", fixture(t, "cli-network-create.json"))
	require.Equal(t, http.StatusOK, status, string(raw))
	_, body := w.forwarded(t, http.MethodPost, "/networks/create")
	assert.Equal(t, map[string]any{OwnerLabel: "app1"}, body["Labels"])

	refused := map[string]map[string]any{
		"network driver overlay is not allowed":  {"Name": "n", "Driver": "overlay"},
		"network scope swarm is not allowed":     {"Name": "n", "Scope": "swarm"},
		"Network.Options is not allowed":         {"Name": "n", "Options": map[string]any{"com.docker.network.bridge.name": "docker0"}},
		"Network.Ingress is not allowed":         {"Name": "n", "Ingress": true},
		"IPAM driver custom is not allowed":      {"Name": "n", "IPAM": map[string]any{"Driver": "custom"}},
		"Network.IPAM.Config is not allowed":     {"Name": "n", "IPAM": map[string]any{"Config": []any{map[string]any{"Subnet": "10.0.0.0/8"}}}},
	}
	for message, request := range refused {
		status, raw = w.do(t, http.MethodPost, "/v1.51/networks/create", request)
		assert.Equal(t, http.StatusForbidden, status, message)
		assert.Equal(t, "hivepaas: "+message, refusalMessage(t, raw))
	}

	status, raw = w.do(t, http.MethodGet, "/v1.51/networks", nil)
	require.Equal(t, http.StatusOK, status)
	var networks []struct{ Name string }
	require.NoError(t, json.Unmarshal(raw, &networks))
	names := make([]string, 0, len(networks))
	for _, n := range networks {
		names = append(names, n.Name)
	}
	assert.ElementsMatch(t, []string{"hp-dapi-app1", "jobnet", "proj_env_net"}, names)

	for path, want := range map[string]int{
		"/v1.51/networks/n-job":        http.StatusOK,
		"/v1.51/networks/hp-dapi-app1": http.StatusOK,
		"/v1.51/networks/othernet":     http.StatusForbidden,
		"/v1.51/networks/hivepaas_net": http.StatusForbidden,
	} {
		status, _ = w.do(t, http.MethodGet, path, nil)
		assert.Equal(t, want, status, path)
	}

	// The app may remove what it created, not the network HivePaaS gave it.
	status, _ = w.do(t, http.MethodDelete, "/v1.51/networks/jobnet", nil)
	assert.Equal(t, http.StatusOK, status)
	status, raw = w.do(t, http.MethodDelete, "/v1.51/networks/hp-dapi-app1", nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: network hp-dapi-app1 is not one this app created", refusalMessage(t, raw))
}

func TestConnectTakesTheAppsChildrenAndTheAppItself(t *testing.T) {
	w := newWorld(t, volumesAndNetworks())
	tests := []struct {
		network string
		body    map[string]any
		want    int
	}{
		{"jobnet", map[string]any{"Container": "child1"}, http.StatusOK},
		{"jobnet", map[string]any{"Container": "child1", "EndpointConfig": map[string]any{"Aliases": []string{"db"}}},
			http.StatusOK},
		// Appwrite's executor connects itself to the network of its runtimes.
		{"jobnet", map[string]any{"Container": "task1"}, http.StatusOK},
		{"proj_env_net", map[string]any{"Container": "child1"}, http.StatusOK},
		{"jobnet", map[string]any{"Container": "other1"}, http.StatusForbidden},
		{"othernet", map[string]any{"Container": "child1"}, http.StatusForbidden},
		{"jobnet", map[string]any{"Container": "child1",
			"EndpointConfig": map[string]any{"IPAMConfig": map[string]any{"IPv4Address": "10.0.0.9"}}},
			http.StatusForbidden},
	}
	for _, tt := range tests {
		status, raw := w.do(t, http.MethodPost, "/v1.51/networks/"+tt.network+"/connect", tt.body)
		assert.Equal(t, tt.want, status, "%s %v %s", tt.network, tt.body, raw)
	}
	status, _ := w.do(t, http.MethodPost, "/v1.51/networks/jobnet/disconnect",
		map[string]any{"Container": "child1", "Force": true})
	assert.Equal(t, http.StatusOK, status)
}

func TestReservedLabelsCannotBeSet(t *testing.T) {
	got := ownLabels(map[string]any{
		"keep": "yes", OwnerLabel: "app2", "com.docker.swarm.service.id": "svc1", "hivepaas.app.info": "{}",
	}, "app1")
	assert.Equal(t, map[string]any{"keep": "yes", OwnerLabel: "app1"}, got)
}
```

- [ ] **Step 3: Run the tests to see them fail**

Run: `go test ./hivepaas_app/pkg/dockerproxy/...`
Expected: FAIL, `undefined: ownLabels`.

- [ ] **Step 4: Write the handlers**

Add to `hivepaas_app/pkg/dockerproxy/body.go`:

```go
// list is v as a JSON array, or nil when it is not one.
func list(v any) []any {
	l, _ := v.([]any)
	return l
}
```

Add to `hivepaas_app/pkg/dockerproxy/policy.go`:

```go
// joinable reports whether a child may join the network of that name.
func (p *Policy) joinable(name string) bool {
	return name == p.Network || slices.Contains(p.Networks, name)
}
```

`hivepaas_app/pkg/dockerproxy/objects.go`:

```go
package dockerproxy

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

const (
	kindVolumes  = "volumes"
	kindNetworks = "networks"

	// serviceIDLabel is how Docker marks the containers of a swarm service's tasks.
	serviceIDLabel = "com.docker.swarm.service.id"
)

var (
	volumeCreateFields  = []string{"Name", "Driver", fieldLabels}
	volumeDrivers       = []string{"", "local"}
	networkCreateFields = []string{"Name", "CheckDuplicate", "Driver", "Scope", "Internal", "Attachable",
		"EnableIPv4", "EnableIPv6", fieldLabels, "IPAM"}
	networkDrivers       = []string{"", "bridge"}
	networkScopes        = []string{"", "local"}
	ipamDrivers          = []string{"", "default"}
	networkConnectFields = []string{"Container", "EndpointConfig", "Force"}
	// endpointFields are what a child may say about a network it joins. A static
	// address or a driver option is a decision about the network, not the child.
	endpointFields = []string{"Aliases", "DNSNames"}
	// reservedLabelPrefixes are labels a client may not set: Docker's, which mark
	// swarm tasks, and HivePaaS's, which mark what it owns.
	reservedLabelPrefixes = []string{"com.docker.", "hivepaas."}
)

// labeled is the part of an inspect answer that says what an object is called
// and whose it is.
type labeled struct {
	Name   string
	Labels map[string]string
}

func (p *Proxy) inspect(ctx context.Context, kind, id string) (labeled, bool, error) {
	var info labeled
	found, err := p.daemon.get(ctx, "/"+kind+"/"+url.PathEscape(id), &info)
	return info, found, err
}

// ownLabels returns the client's labels less the reserved ones, with the owner.
func ownLabels(labels map[string]any, appID string) map[string]any {
	out := map[string]any{}
	for key, value := range labels {
		reserved := slices.ContainsFunc(reservedLabelPrefixes, func(prefix string) bool {
			return strings.HasPrefix(key, prefix)
		})
		if !reserved {
			out[key] = value
		}
	}
	out[OwnerLabel] = appID
	return out
}

func (p *Proxy) listVolumes(c *call) {
	var answer map[string]any
	if !p.fetch(c, &answer) {
		return
	}
	items := make([]map[string]any, 0, len(list(answer["Volumes"])))
	for _, item := range list(answer["Volumes"]) {
		items = append(items, object(item))
	}
	kept := keep(items, func(item map[string]any) bool { return ownedBy(item[fieldLabels], c.policy.AppID) })
	answer["Volumes"] = kept
	p.decide(c, true, fmt.Sprintf("%d of %d volumes are the app's", len(kept), len(items)))
	writeJSON(c.w, http.StatusOK, answer)
}

func (p *Proxy) volumeCreate(c *call) {
	body, err := readBody(c.r)
	if err == nil {
		err = checkFields("Volume", body, volumeCreateFields, nil)
	}
	if driver := text(body["Driver"]); err == nil && !slices.Contains(volumeDrivers, driver) {
		err = refusef("volume driver %s is not allowed", driver)
	}
	if err != nil {
		p.refuse(c, err)
		return
	}
	body[fieldLabels] = ownLabels(object(body[fieldLabels]), c.policy.AppID)
	setBody(c.r, body)
	p.forward(c, "a volume of the app's own")
}

func (p *Proxy) onVolume(c *call) {
	info, found, err := p.inspect(c.r.Context(), kindVolumes, c.args[0])
	if err == nil && (!found || info.Labels[OwnerLabel] != c.policy.AppID) {
		err = refusef("volume %s is not one this app created", c.args[0])
	}
	if err != nil {
		p.refuse(c, err)
		return
	}
	p.forward(c, "a volume of the app's own")
}

func (p *Proxy) listNetworks(c *call) {
	var items []map[string]any
	if !p.fetch(c, &items) {
		return
	}
	kept := keep(items, func(item map[string]any) bool {
		return ownedBy(item[fieldLabels], c.policy.AppID) || c.policy.joinable(text(item["Name"]))
	})
	p.decide(c, true, fmt.Sprintf("%d of %d networks are the app's", len(kept), len(items)))
	writeJSON(c.w, http.StatusOK, kept)
}

func (p *Proxy) networkCreate(c *call) {
	body, err := readBody(c.r)
	if err == nil {
		err = checkNetworkCreate(body)
	}
	if err != nil {
		p.refuse(c, err)
		return
	}
	body[fieldLabels] = ownLabels(object(body[fieldLabels]), c.policy.AppID)
	setBody(c.r, body)
	p.forward(c, "a network of the app's own")
}

// checkNetworkCreate takes a plain bridge on this node. Anything else - an
// overlay across the cluster, a driver option naming a host interface, an
// address range of the operator's - is a decision about the host.
func checkNetworkCreate(body map[string]any) error {
	if err := checkFields("Network", body, networkCreateFields, nil); err != nil {
		return err
	}
	if driver := text(body["Driver"]); !slices.Contains(networkDrivers, driver) {
		return refusef("network driver %s is not allowed", driver)
	}
	if scope := text(body["Scope"]); !slices.Contains(networkScopes, scope) {
		return refusef("network scope %s is not allowed", scope)
	}
	ipam := object(body["IPAM"])
	if driver := text(ipam["Driver"]); !slices.Contains(ipamDrivers, driver) {
		return refusef("IPAM driver %s is not allowed", driver)
	}
	return checkFields("Network.IPAM", ipam, []string{"Driver"}, nil)
}

func (p *Proxy) networkRead(c *call) {
	if err := p.usableNetwork(c.r.Context(), c.policy, c.args[0]); err != nil {
		p.refuse(c, err)
		return
	}
	p.forward(c, "a network the app may use")
}

func (p *Proxy) networkDelete(c *call) {
	info, found, err := p.inspect(c.r.Context(), kindNetworks, c.args[0])
	if err == nil && (!found || info.Labels[OwnerLabel] != c.policy.AppID) {
		err = refusef("network %s is not one this app created", c.args[0])
	}
	if err != nil {
		p.refuse(c, err)
		return
	}
	p.forward(c, "a network of the app's own")
}

func (p *Proxy) networkConnect(c *call) {
	ctx := c.r.Context()
	body, err := readBody(c.r)
	if err == nil {
		err = checkFields("connect", body, networkConnectFields, nil)
	}
	if err == nil {
		err = checkFields("connect.EndpointConfig", object(body["EndpointConfig"]), endpointFields, nil)
	}
	if err == nil {
		err = p.usableNetwork(ctx, c.policy, c.args[0])
	}
	if err == nil {
		err = p.childOrSelf(ctx, c.policy, text(body["Container"]))
	}
	if err != nil {
		p.refuse(c, err)
		return
	}
	setBody(c.r, body)
	p.forward(c, "a network the app may use")
}

// usableNetwork checks that children may join the network: the app's own, one
// the policy names, or one the app created.
func (p *Proxy) usableNetwork(ctx context.Context, policy *Policy, idOrName string) error {
	if policy.joinable(idOrName) {
		return nil
	}
	info, found, err := p.inspect(ctx, kindNetworks, idOrName)
	if err != nil {
		return err
	}
	if !found {
		return refusef("network %s does not exist", idOrName)
	}
	if policy.joinable(info.Name) || info.Labels[OwnerLabel] == policy.AppID {
		return nil
	}
	return refusef("network %s is not one this app may use", idOrName)
}

// childOrSelf also lets through the app's own task containers, which may
// connect themselves to the networks of their children.
func (p *Proxy) childOrSelf(ctx context.Context, policy *Policy, id string) error {
	labels, err := p.containerLabels(ctx, id)
	if err != nil {
		return err
	}
	if labels[OwnerLabel] == policy.AppID {
		return nil
	}
	if labels[OwnerLabel] == "" && policy.ServiceID != "" && labels[serviceIDLabel] == policy.ServiceID {
		return nil
	}
	return refusef("container %s is not one this app started", id)
}
```

In `hivepaas_app/pkg/dockerproxy/routes.go`, append to `routes`:

```go
	on("GET", `/volumes`, GroupVolumes, (*Proxy).listVolumes),
	on("POST", `/volumes/create`, GroupVolumes, (*Proxy).volumeCreate),
	on("GET,DELETE", `/volumes/`+idPart, GroupVolumes, (*Proxy).onVolume),
	on("GET", `/networks`, GroupNetworks, (*Proxy).listNetworks),
	on("POST", `/networks/create`, GroupNetworks, (*Proxy).networkCreate),
	on("GET", `/networks/`+idPart, GroupNetworks, (*Proxy).networkRead),
	on("DELETE", `/networks/`+idPart, GroupNetworks, (*Proxy).networkDelete),
	on("POST", `/networks/`+idPart+`/(?:connect|disconnect)`, GroupNetworks, (*Proxy).networkConnect),
```

- [ ] **Step 5: Run the tests to see them pass, then lint**

Run: `go test ./hivepaas_app/pkg/dockerproxy/... && golangci-lint run ./hivepaas_app/pkg/dockerproxy/...`
Expected: `ok`, `0 issues`.

- [ ] **Step 6: Commit**

```bash
git add hivepaas_app/pkg/dockerproxy
git commit -m "feat(dockerproxy): volumes and networks of the app's own

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 5: Container create - fields, image, network, limits, owner, count, version

**Files:**
- Create: `hivepaas_app/pkg/dockerproxy/create.go`
- Create: `hivepaas_app/pkg/dockerproxy/limits.go`
- Create: `hivepaas_app/pkg/dockerproxy/testdata/cli-create-plain.json`
- Modify: `hivepaas_app/pkg/dockerproxy/body.go` (add `number`)
- Modify: `hivepaas_app/pkg/dockerproxy/daemon.go` (add `labelFilter`)
- Modify: `hivepaas_app/pkg/dockerproxy/routes.go` (add the create route)
- Modify: `hivepaas_app/pkg/dockerproxy/world_test.go` (add `set`)
- Test: `hivepaas_app/pkg/dockerproxy/create_test.go`

**Interfaces:**
- Consumes: Task 4.
- Produces: `(*Proxy).checkCreate(ctx, policy, body map[string]any) error` (Task 6 adds a storage step to it), `hostFields` (Task 6 adds `Binds` and `Mounts`), `number(v any) (int64, error)`, `labelFilter(key, value string) string`, `raiseVersion(r *http.Request)`; test helper `set(body map[string]any, dotted string, value any) map[string]any`.

Until Task 6, `Binds` and `Mounts` are not in `hostFields`, so a create that names any storage is refused.

- [ ] **Step 1: Add the recorded body and the test helper**

Recorded from `docker create alpine:3 true`, docker CLI 29.8.

`hivepaas_app/pkg/dockerproxy/testdata/cli-create-plain.json`:

```json
{"Hostname":"","Domainname":"","User":"","AttachStdin":false,"AttachStdout":true,"AttachStderr":true,"Tty":false,"OpenStdin":false,"StdinOnce":false,"Env":null,"Cmd":["true"],"Image":"alpine:3","Volumes":{},"WorkingDir":"","Entrypoint":null,"Labels":{},"HostConfig":{"Binds":null,"ContainerIDFile":"","LogConfig":{"Type":"","Config":{}},"NetworkMode":"default","PortBindings":{},"RestartPolicy":{"Name":"no","MaximumRetryCount":0},"AutoRemove":false,"VolumeDriver":"","VolumesFrom":null,"ConsoleSize":[0,0],"CapAdd":null,"CapDrop":null,"CgroupnsMode":"","Dns":null,"DnsOptions":[],"DnsSearch":[],"ExtraHosts":null,"GroupAdd":null,"IpcMode":"","Cgroup":"","Links":null,"OomScoreAdj":0,"PidMode":"","Privileged":false,"PublishAllPorts":false,"ReadonlyRootfs":false,"SecurityOpt":null,"UTSMode":"","UsernsMode":"","ShmSize":0,"Isolation":"","CpuShares":0,"Memory":0,"NanoCpus":0,"CgroupParent":"","BlkioWeight":0,"BlkioWeightDevice":[],"BlkioDeviceReadBps":[],"BlkioDeviceWriteBps":[],"BlkioDeviceReadIOps":[],"BlkioDeviceWriteIOps":[],"CpuPeriod":0,"CpuQuota":0,"CpuRealtimePeriod":0,"CpuRealtimeRuntime":0,"CpusetCpus":"","CpusetMems":"","Devices":[],"DeviceCgroupRules":null,"DeviceRequests":null,"MemoryReservation":0,"MemorySwap":0,"MemorySwappiness":-1,"OomKillDisable":false,"PidsLimit":0,"Ulimits":null,"CpuCount":0,"CpuPercent":0,"IOMaximumIOps":0,"IOMaximumBandwidth":0,"MaskedPaths":null,"ReadonlyPaths":null},"NetworkingConfig":{"EndpointsConfig":{"default":{"IPAMConfig":null,"Links":null,"Aliases":null,"DriverOpts":null,"GwPriority":0,"NetworkID":"","EndpointID":"","Gateway":"","IPAddress":"","MacAddress":"","IPPrefixLen":0,"IPv6Gateway":"","GlobalIPv6Address":"","GlobalIPv6PrefixLen":0,"DNSNames":null}}}}
```

Add to `hivepaas_app/pkg/dockerproxy/world_test.go`:

```go
// set changes a field of a decoded body by its dotted path, and returns the body.
func set(body map[string]any, dotted string, value any) map[string]any {
	keys := strings.Split(dotted, ".")
	obj := body
	for _, key := range keys[:len(keys)-1] {
		next, ok := obj[key].(map[string]any)
		if !ok {
			next = map[string]any{}
			obj[key] = next
		}
		obj = next
	}
	obj[keys[len(keys)-1]] = value
	return body
}
```

- [ ] **Step 2: Write the failing tests**

`hivepaas_app/pkg/dockerproxy/create_test.go`:

```go
package dockerproxy

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const createPath = "/v1.51/containers/create"

func TestCreateTakesWhatTheCLISendsForAPlainContainer(t *testing.T) {
	w := newWorld(t, testPolicy())
	status, raw := w.do(t, http.MethodPost, createPath+"?name=job1", fixture(t, "cli-create-plain.json"))
	require.Equal(t, http.StatusCreated, status, string(raw))

	path, body := w.forwarded(t, http.MethodPost, "/containers/create")
	assert.Equal(t, createPath, path)
	host := object(body["HostConfig"])
	assert.Equal(t, "hp-dapi-app1", host["NetworkMode"])
	endpoints := object(object(body["NetworkingConfig"])["EndpointsConfig"])
	assert.Contains(t, endpoints, "hp-dapi-app1")
	assert.NotContains(t, endpoints, "default")
	assert.Equal(t, json.Number("1073741824"), host["Memory"])
	assert.Equal(t, json.Number("1000000000"), host["NanoCpus"])
	assert.Equal(t, json.Number("1024"), host["PidsLimit"])
	assert.Equal(t, map[string]any{OwnerLabel: "app1"}, body["Labels"])
}

func TestCreateRefusesEveryFieldThatReachesPastTheContainer(t *testing.T) {
	w := newWorld(t, testPolicy())
	fields := map[string]any{
		"HostConfig.Privileged": true,
		"HostConfig.CapAdd":     []any{"SYS_ADMIN"},
		"HostConfig.Devices": []any{map[string]any{
			"PathOnHost": "/dev/kmsg", "PathInContainer": "/dev/kmsg", "CgroupPermissions": "rwm",
		}},
		"HostConfig.DeviceRequests":    []any{map[string]any{"Count": -1, "Capabilities": []any{[]any{"gpu"}}}},
		"HostConfig.DeviceCgroupRules": []any{"c 1:3 rwm"},
		"HostConfig.PidMode":           "host",
		"HostConfig.IpcMode":           "host",
		"HostConfig.UTSMode":           "host",
		"HostConfig.UsernsMode":        "host",
		"HostConfig.CgroupnsMode":      "host",
		"HostConfig.Cgroup":            "container:other1",
		"HostConfig.SecurityOpt":       []any{"seccomp=unconfined"},
		"HostConfig.Runtime":           "runc",
		"HostConfig.Sysctls":           map[string]any{"net.ipv4.ip_forward": "1"},
		"HostConfig.CgroupParent":      "/",
		"HostConfig.OomKillDisable":    true,
		"HostConfig.OomScoreAdj":       -1000,
		"HostConfig.VolumesFrom":       []any{"other1"},
		"HostConfig.Links":             []any{"other1:db"},
		"HostConfig.PortBindings":      map[string]any{"22/tcp": []any{map[string]any{"HostPort": "2222"}}},
		"HostConfig.PublishAllPorts":   true,
		"HostConfig.VolumeDriver":      "local",
		"HostConfig.MemorySwappiness":  60,
		"HostConfig.MaskedPaths":       []any{},
		"HostConfig.ReadonlyPaths":     []any{},
	}
	for field, value := range fields {
		status, raw := w.do(t, http.MethodPost, createPath, set(fixture(t, "cli-create-plain.json"), field, value))
		assert.Equal(t, http.StatusForbidden, status, field)
		assert.Contains(t, refusalMessage(t, raw), field, field)
	}
	assert.False(t, w.reached(http.MethodPost, "/containers/create"))
}

func TestCreateTakesOnlyThePolicysImages(t *testing.T) {
	w := newWorld(t, testPolicy())
	status, raw := w.do(t, http.MethodPost, createPath, set(fixture(t, "cli-create-plain.json"), "Image", "busybox"))
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: image busybox is not allowed", refusalMessage(t, raw))
}

func TestCreatePutsTheDefaultNetworksOnTheAppsNetwork(t *testing.T) {
	w := newWorld(t, testPolicy())
	for _, mode := range []string{"", "default", "bridge", "host"} {
		status, raw := w.do(t, http.MethodPost, createPath,
			set(fixture(t, "cli-create-plain.json"), "HostConfig.NetworkMode", mode))
		require.Equal(t, http.StatusCreated, status, "%q %s", mode, raw)
		_, body := w.forwarded(t, http.MethodPost, "/containers/create")
		assert.Equal(t, "hp-dapi-app1", object(body["HostConfig"])["NetworkMode"], mode)
	}
}

func TestCreateJoinsOnlyNetworksTheAppMayUse(t *testing.T) {
	w := newWorld(t, testPolicy())
	for mode, want := range map[string]int{
		"jobnet":           http.StatusCreated,
		"n-job":            http.StatusCreated,
		"proj_env_net":     http.StatusCreated,
		"none":             http.StatusCreated,
		"othernet":         http.StatusForbidden,
		"hivepaas_net":     http.StatusForbidden,
		"container:other1": http.StatusForbidden,
	} {
		body := set(fixture(t, "cli-create-plain.json"), "HostConfig.NetworkMode", mode)
		body["NetworkingConfig"] = map[string]any{}
		status, raw := w.do(t, http.MethodPost, createPath, body)
		assert.Equal(t, want, status, "%s %s", mode, raw)
	}

	body := set(fixture(t, "cli-create-plain.json"), "HostConfig.NetworkMode", "jobnet")
	body["NetworkingConfig"] = map[string]any{"EndpointsConfig": map[string]any{"othernet": map[string]any{}}}
	status, raw := w.do(t, http.MethodPost, createPath, body)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: network othernet is not one this app may use", refusalMessage(t, raw))

	body["NetworkingConfig"] = map[string]any{"EndpointsConfig": map[string]any{
		"jobnet": map[string]any{"IPAMConfig": map[string]any{"IPv4Address": "10.0.0.9"}},
	}}
	status, raw = w.do(t, http.MethodPost, createPath, body)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: EndpointsConfig.jobnet.IPAMConfig is not allowed", refusalMessage(t, raw))
}

func TestCreateGivesTheLimitsAndRefusesMore(t *testing.T) {
	w := newWorld(t, testPolicy())
	refused := map[string]map[string]any{
		"memory":            {"HostConfig.Memory": 2 << 30},
		"cpus":              {"HostConfig.NanoCpus": 2_000_000_000},
		"cpu quota":         {"HostConfig.CpuQuota": 200000, "HostConfig.CpuPeriod": 100000},
		"cpu period":        {"HostConfig.CpuQuota": 1000, "HostConfig.CpuPeriod": 10},
		"pids limit":        {"HostConfig.PidsLimit": 99999},
		"unlimited swap":    {"HostConfig.MemorySwap": -1},
		"log driver syslog": {"HostConfig.LogConfig": map[string]any{"Type": "syslog"}},
	}
	for what, fields := range refused {
		body := fixture(t, "cli-create-plain.json")
		for field, value := range fields {
			set(body, field, value)
		}
		status, raw := w.do(t, http.MethodPost, createPath, body)
		assert.Equal(t, http.StatusForbidden, status, what)
		assert.Contains(t, refusalMessage(t, raw), what)
	}

	body := set(fixture(t, "cli-create-plain.json"), "HostConfig.CpuQuota", 50000)
	status, raw := w.do(t, http.MethodPost, createPath, body)
	require.Equal(t, http.StatusCreated, status, string(raw))
	_, forwarded := w.forwarded(t, http.MethodPost, "/containers/create")
	assert.Equal(t, json.Number("0"), object(forwarded["HostConfig"])["NanoCpus"],
		"a quota is the child's own limit, and docker refuses both at once")
}

func TestCreateReplacesTheOwnerLabel(t *testing.T) {
	w := newWorld(t, testPolicy())
	body := set(fixture(t, "cli-create-plain.json"), "Labels", map[string]any{
		"keep": "yes", OwnerLabel: "app2", "com.docker.swarm.service.id": "svc1",
	})
	status, _ := w.do(t, http.MethodPost, createPath, body)
	require.Equal(t, http.StatusCreated, status)
	_, forwarded := w.forwarded(t, http.MethodPost, "/containers/create")
	assert.Equal(t, map[string]any{"keep": "yes", OwnerLabel: "app1"}, forwarded["Labels"])
}

func TestCreateStopsAtTheContainerLimit(t *testing.T) {
	policy := testPolicy()
	policy.Limits.Containers = 1
	w := newWorld(t, policy)
	status, raw := w.do(t, http.MethodPost, createPath, fixture(t, "cli-create-plain.json"))
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: this app already has 1 containers, its limit", refusalMessage(t, raw))
}

func TestCreateSpeaksAtLeastTheVersionSubpathsNeed(t *testing.T) {
	w := newWorld(t, testPolicy())
	for sent, want := range map[string]string{
		"/v1.44/containers/create": "/v1.45/containers/create",
		"/v1.51/containers/create": "/v1.51/containers/create",
		"/containers/create":       "/containers/create",
	} {
		status, _ := w.do(t, http.MethodPost, sent, fixture(t, "cli-create-plain.json"))
		require.Equal(t, http.StatusCreated, status, sent)
		path, _ := w.forwarded(t, http.MethodPost, "/containers/create")
		assert.Equal(t, want, path)
	}
}

func TestCreateRefusesStorageUntilItIsJudged(t *testing.T) {
	w := newWorld(t, testPolicy())
	body := set(fixture(t, "cli-create-plain.json"), "HostConfig.Binds", []any{"/:/host"})
	status, _ := w.do(t, http.MethodPost, createPath, body)
	assert.Equal(t, http.StatusForbidden, status)
}
```

- [ ] **Step 3: Run the tests to see them fail**

Run: `go test ./hivepaas_app/pkg/dockerproxy/...`
Expected: FAIL: `/containers/create` answers 403 `is not an endpoint this app may use`.

- [ ] **Step 4: Write create**

Add to `hivepaas_app/pkg/dockerproxy/body.go`, with `"encoding/json"` already imported:

```go
// number reads a whole number the client sent, or one the proxy set. Absent and
// null are zero.
func number(v any) (int64, error) {
	switch n := v.(type) {
	case nil:
		return 0, nil
	case int64:
		return n, nil
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			return 0, refusef("%s is not a whole number", n)
		}
		return i, nil
	default:
		return 0, refusef("%v is not a number", v)
	}
}
```

Add to `hivepaas_app/pkg/dockerproxy/daemon.go`, importing `"net/url"`:

```go
// labelFilter is a filters query parameter selecting by one label.
func labelFilter(key, value string) string {
	raw, _ := json.Marshal(map[string][]string{"label": {key + "=" + value}})
	return url.QueryEscape(string(raw))
}
```

`hivepaas_app/pkg/dockerproxy/limits.go`:

```go
package dockerproxy

const (
	defaultPidsLimit = 1024
	maxPidsLimit     = 4096
	// A CPU period is what a quota is a share of, and what docker accepts for it.
	defaultCPUPeriod = 100_000
	minCPUPeriod     = 1_000
	maxCPUPeriod     = 1_000_000
	nanoPerCPU       = 1_000_000_000
)

// applyLimits gives a child the policy's limits when it asks for none, and
// refuses one that asks for more.
func applyLimits(host map[string]any, limits Limits) error {
	memory, err := number(host["Memory"])
	if err != nil {
		return err
	}
	switch {
	case memory <= 0:
		host["Memory"] = limits.Memory
	case memory > limits.Memory:
		return refusef("memory %d is more than the %d this app's children may have", memory, limits.Memory)
	}
	swap, err := number(host["MemorySwap"])
	if err != nil {
		return err
	}
	if swap < 0 {
		return refusef("unlimited swap is not allowed")
	}
	if err = applyCPU(host, limits.NanoCPUs); err != nil {
		return err
	}
	return applyPids(host)
}

// applyCPU caps processor time, given either way docker takes it: NanoCpus, or
// a quota of a period.
func applyCPU(host map[string]any, limit int64) error {
	quota, err := number(host["CpuQuota"])
	if err != nil {
		return err
	}
	if quota <= 0 {
		nano, err := number(host["NanoCpus"])
		if err != nil {
			return err
		}
		switch {
		case nano <= 0:
			host["NanoCpus"] = limit
		case nano > limit:
			return refusef("cpus %.2f is more than the %.2f this app's children may have",
				float64(nano)/nanoPerCPU, float64(limit)/nanoPerCPU)
		}
		return nil
	}
	period, err := number(host["CpuPeriod"])
	if err != nil {
		return err
	}
	if period == 0 {
		period = defaultCPUPeriod
	}
	if period < minCPUPeriod || period > maxCPUPeriod {
		return refusef("cpu period %d is outside %d-%d", period, minCPUPeriod, maxCPUPeriod)
	}
	// Compared as quotas: the limit times a period stays far inside an int64,
	// where the quota times a billion need not.
	if allowed := limit * period / nanoPerCPU; quota > allowed {
		return refusef("cpu quota %d of %d is more than the %.2f cpus this app's children may have",
			quota, period, float64(limit)/nanoPerCPU)
	}
	return nil
}

func applyPids(host map[string]any) error {
	pids, err := number(host["PidsLimit"])
	if err != nil {
		return err
	}
	switch {
	case pids <= 0:
		host["PidsLimit"] = int64(defaultPidsLimit)
	case pids > maxPidsLimit:
		return refusef("pids limit %d is more than %d", pids, maxPidsLimit)
	}
	return nil
}
```

`hivepaas_app/pkg/dockerproxy/create.go`:

```go
package dockerproxy

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

// minCreateMajor and minCreateMinor are the API version a rewritten create
// needs: a mount's VolumeOptions.Subpath came in 1.45.
const (
	minCreateMajor = 1
	minCreateMinor = 45
)

const (
	networkModeNone      = "none"
	networkModeContainer = "container:"
)

var (
	// configFields are the fields of a container's Config a child may set.
	configFields = []string{"Hostname", "Domainname", "User", "AttachStdin", "AttachStdout", "AttachStderr",
		"ExposedPorts", "Tty", "OpenStdin", "StdinOnce", "Env", "Cmd", "Healthcheck", "ArgsEscaped", "Image",
		"Volumes", "WorkingDir", "Entrypoint", "NetworkDisabled", fieldLabels, "StopSignal", "StopTimeout",
		"Shell", "HostConfig", "NetworkingConfig"}
	// hostFields are the fields of a HostConfig a child may set. Everything that
	// reaches past the container - Privileged, CapAdd, Devices, the host's
	// namespaces, SecurityOpt, Runtime, Sysctls, CgroupParent, VolumesFrom,
	// Links, PortBindings - is absent, and so refused.
	hostFields = []string{"NetworkMode", "RestartPolicy", "AutoRemove", "Memory", "MemorySwap",
		"MemoryReservation", "NanoCpus", "CpuQuota", "CpuPeriod", "CpuShares", "PidsLimit", "ShmSize", "Dns",
		"DnsOptions", "DnsSearch", "ExtraHosts", "LogConfig", "Init", "ReadonlyRootfs", "Tmpfs", "CapDrop",
		"Ulimits", "GroupAdd", "ConsoleSize", "Isolation"}
	// hostNullOnly are fields whose empty list unmasks /proc.
	hostNullOnly     = []string{"MaskedPaths", "ReadonlyPaths"}
	networkingFields = []string{"EndpointsConfig"}
	// logDrivers write where docker logs reads, and nowhere else.
	logDrivers = []string{"", "json-file", "local"}
	// defaultNetworkModes are what a client asks for when it names no network of
	// its own, and they all become the app's network. host is among them because
	// the apps that ask for it - Autobase's automation - want the internet, which
	// the app's network reaches too.
	defaultNetworkModes = []string{"", "default", "bridge", "host"}
)

func (p *Proxy) create(c *call) {
	body, err := readBody(c.r)
	if err == nil {
		err = p.checkCreate(c.r.Context(), c.policy, body)
	}
	if err != nil {
		p.refuse(c, err)
		return
	}
	setBody(c.r, body)
	raiseVersion(c.r)
	p.forward(c, "a child of the app")
}

// checkCreate judges a container create and rewrites it into what the app may
// have. It changes body in place.
func (p *Proxy) checkCreate(ctx context.Context, policy *Policy, body map[string]any) error {
	if err := checkFields("Config", body, configFields, nil); err != nil {
		return err
	}
	if image := text(body["Image"]); !matchImage(policy.Images, image) {
		return refusef("image %s is not allowed", image)
	}
	host := object(body["HostConfig"])
	if host == nil {
		host = map[string]any{}
		body["HostConfig"] = host
	}
	if err := checkFields("HostConfig", host, hostFields, hostNullOnly); err != nil {
		return err
	}
	if driver := text(object(host["LogConfig"])["Type"]); !slices.Contains(logDrivers, driver) {
		return refusef("log driver %s is not allowed", driver)
	}
	if err := p.rewriteNetwork(ctx, policy, body, host); err != nil {
		return err
	}
	if err := applyLimits(host, policy.Limits); err != nil {
		return err
	}
	body[fieldLabels] = ownLabels(object(body[fieldLabels]), policy.AppID)
	return p.checkCount(ctx, policy)
}

// rewriteNetwork puts a child on the app's network unless it names another the
// app may use.
func (p *Proxy) rewriteNetwork(ctx context.Context, policy *Policy, body, host map[string]any) error {
	mode := text(host["NetworkMode"])
	switch {
	case slices.Contains(defaultNetworkModes, mode):
		host["NetworkMode"] = policy.Network
	case mode == networkModeNone:
	case strings.HasPrefix(mode, networkModeContainer):
		return refusef("sharing another container's network is not allowed")
	default:
		if err := p.usableNetwork(ctx, policy, mode); err != nil {
			return err
		}
	}
	networking := object(body["NetworkingConfig"])
	if err := checkFields("NetworkingConfig", networking, networkingFields, nil); err != nil {
		return err
	}
	endpoints := object(networking["EndpointsConfig"])
	for _, name := range slices.Sorted(maps.Keys(endpoints)) {
		if err := checkFields("EndpointsConfig."+name, object(endpoints[name]), endpointFields, nil); err != nil {
			return err
		}
		if slices.Contains(defaultNetworkModes, name) {
			endpoints[policy.Network] = endpoints[name]
			delete(endpoints, name)
			continue
		}
		if err := p.usableNetwork(ctx, policy, name); err != nil {
			return err
		}
	}
	return nil
}

func (p *Proxy) checkCount(ctx context.Context, policy *Policy) error {
	var children []struct {
		ID string `json:"Id"`
	}
	query := "/containers/json?all=1&filters=" + labelFilter(OwnerLabel, policy.AppID)
	if _, err := p.daemon.get(ctx, query, &children); err != nil {
		return err
	}
	if len(children) >= policy.Limits.Containers {
		return refusef("this app already has %d containers, its limit", len(children))
	}
	return nil
}

// raiseVersion sends a create at the version it needs when the client asked for
// an older one. What the client sent means the same at the higher version, and a
// request with no version already gets the daemon's own.
func raiseVersion(r *http.Request) {
	current := versionPrefix.FindString(r.URL.Path)
	if current == "" {
		return
	}
	major, minor := splitVersion(strings.TrimPrefix(current, "/v"))
	if major > minCreateMajor || (major == minCreateMajor && minor >= minCreateMinor) {
		return
	}
	r.URL.Path = fmt.Sprintf("/v%d.%d", minCreateMajor, minCreateMinor) + strings.TrimPrefix(r.URL.Path, current)
	r.URL.RawPath = ""
}

func splitVersion(v string) (int, int) {
	major, minor, _ := strings.Cut(v, ".")
	majorNumber, _ := strconv.Atoi(major)
	minorNumber, _ := strconv.Atoi(minor)
	return majorNumber, minorNumber
}
```

In `hivepaas_app/pkg/dockerproxy/routes.go`, add after the `/containers/json` entry:

```go
	on("POST", `/containers/create`, "", (*Proxy).create),
```

- [ ] **Step 5: Run the tests to see them pass, then lint**

Run: `go test ./hivepaas_app/pkg/dockerproxy/... && golangci-lint run ./hivepaas_app/pkg/dockerproxy/...`
Expected: `ok`, `0 issues`.

- [ ] **Step 6: Commit**

```bash
git add hivepaas_app/pkg/dockerproxy
git commit -m "feat(dockerproxy): judge and rewrite a container create

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 6: Storage - shared directories, the nested socket, named volumes

**Files:**
- Create: `hivepaas_app/pkg/dockerproxy/storage.go`
- Create: `hivepaas_app/pkg/dockerproxy/testdata/cli-create-bind-host.json`
- Modify: `hivepaas_app/pkg/dockerproxy/create.go` (`hostFields` gains `Binds` and `Mounts`; `checkCreate` calls `rewriteStorage`)
- Modify: `hivepaas_app/pkg/dockerproxy/daemon.go` (add `createVolume`)
- Modify: `hivepaas_app/pkg/dockerproxy/create_test.go` (drop `TestCreateRefusesStorageUntilItIsJudged`)
- Test: `hivepaas_app/pkg/dockerproxy/storage_test.go`

**Interfaces:**
- Consumes: Task 5.
- Produces: `(*Proxy).rewriteStorage(ctx, policy, host map[string]any) error`, `(*daemon).createVolume(ctx, name string, labels map[string]string) error`.

- [ ] **Step 1: Add the recorded body**

Recorded from `docker create -v /var/lib/autobase/ansible:/tmp/ansible --network host -m 512m alpine:3`, docker CLI 29.8 - what Autobase's console asks for, in the CLI's shape.

`hivepaas_app/pkg/dockerproxy/testdata/cli-create-bind-host.json`:

```json
{"Hostname":"","Domainname":"","User":"","AttachStdin":false,"AttachStdout":true,"AttachStderr":true,"Tty":false,"OpenStdin":false,"StdinOnce":false,"Env":null,"Cmd":null,"Image":"alpine:3","Volumes":{},"WorkingDir":"","Entrypoint":null,"Labels":{},"HostConfig":{"Binds":["/var/lib/autobase/ansible:/tmp/ansible"],"ContainerIDFile":"","LogConfig":{"Type":"","Config":{}},"NetworkMode":"host","PortBindings":{},"RestartPolicy":{"Name":"no","MaximumRetryCount":0},"AutoRemove":false,"VolumeDriver":"","VolumesFrom":null,"ConsoleSize":[0,0],"CapAdd":null,"CapDrop":null,"CgroupnsMode":"","Dns":null,"DnsOptions":[],"DnsSearch":[],"ExtraHosts":null,"GroupAdd":null,"IpcMode":"","Cgroup":"","Links":null,"OomScoreAdj":0,"PidMode":"","Privileged":false,"PublishAllPorts":false,"ReadonlyRootfs":false,"SecurityOpt":null,"UTSMode":"","UsernsMode":"","ShmSize":0,"Isolation":"","CpuShares":0,"Memory":536870912,"NanoCpus":0,"CgroupParent":"","BlkioWeight":0,"BlkioWeightDevice":[],"BlkioDeviceReadBps":[],"BlkioDeviceWriteBps":[],"BlkioDeviceReadIOps":[],"BlkioDeviceWriteIOps":[],"CpuPeriod":0,"CpuQuota":0,"CpuRealtimePeriod":0,"CpuRealtimeRuntime":0,"CpusetCpus":"","CpusetMems":"","Devices":[],"DeviceCgroupRules":null,"DeviceRequests":null,"MemoryReservation":0,"MemorySwap":0,"MemorySwappiness":-1,"OomKillDisable":false,"PidsLimit":0,"Ulimits":null,"CpuCount":0,"CpuPercent":0,"IOMaximumIOps":0,"IOMaximumBandwidth":0,"MaskedPaths":null,"ReadonlyPaths":null},"NetworkingConfig":{"EndpointsConfig":{}}}
```

- [ ] **Step 2: Write the failing tests**

Delete `TestCreateRefusesStorageUntilItIsJudged` from `create_test.go`: storage is judged from now on.

`hivepaas_app/pkg/dockerproxy/storage_test.go`:

```go
package dockerproxy

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createWith sends the plain CLI create with the given binds and mounts.
func createWith(t *testing.T, w *world, binds, mounts []any) (int, string) {
	t.Helper()
	body := fixture(t, "cli-create-plain.json")
	set(body, "HostConfig.Binds", binds)
	set(body, "HostConfig.Mounts", mounts)
	status, raw := w.do(t, http.MethodPost, createPath, body)
	if status == http.StatusCreated {
		return status, ""
	}
	return status, refusalMessage(t, raw)
}

func forwardedStorage(t *testing.T, w *world) ([]any, []any) {
	t.Helper()
	_, body := w.forwarded(t, http.MethodPost, "/containers/create")
	host := object(body["HostConfig"])
	return list(host["Binds"]), list(host["Mounts"])
}

func TestCreateMountsASharedDirectoryFromTheAppsVolume(t *testing.T) {
	w := newWorld(t, testPolicy())
	status, raw := w.do(t, http.MethodPost, createPath, fixture(t, "cli-create-bind-host.json"))
	require.Equal(t, http.StatusCreated, status, string(raw))

	binds, mounts := forwardedStorage(t, w)
	assert.Empty(t, binds)
	assert.Equal(t, []any{map[string]any{
		"Type": "volume", "Source": "hp-vol-data", "Target": "/tmp/ansible", "ReadOnly": false,
		"VolumeOptions": map[string]any{"NoCopy": true, "Subpath": "app1-key/ansible"},
	}}, mounts)
}

func TestCreateReachesNoOtherPathOfTheHost(t *testing.T) {
	w := newWorld(t, testPolicy())
	for _, bind := range []string{
		"/:/host",
		"/var/run/docker.sock:/var/run/docker.sock",
		// The app's own mount, but not a directory it shares.
		"/var/lib/autobase:/data",
		"/var/lib/autobase/ansiblex:/data",
		"/var/lib/autobase/ansible/../../..:/data",
	} {
		status, message := createWith(t, w, []any{bind}, nil)
		assert.Equal(t, http.StatusForbidden, status, bind)
		assert.Contains(t, message, "is not a shared directory of this app", bind)
	}
	status, message := createWith(t, w, nil, []any{map[string]any{"Type": "bind", "Source": "/etc", "Target": "/x"}})
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: /etc is not a shared directory of this app", message)
	assert.False(t, w.reached(http.MethodPost, "/containers/create"))
}

func TestCreateRefusesASharedDirectoryTheAppKeepsNoVolumeFor(t *testing.T) {
	policy := testPolicy()
	policy.SharedDirs = []string{"/srv/work"}
	w := newWorld(t, policy)
	status, message := createWith(t, w, []any{"/srv/work:/work"}, nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: shared directory /srv/work is not on a volume of the app", message)
}

func TestCreateFindsSharedDirectoriesOnlyInTheAppsOwnTask(t *testing.T) {
	w := newWorld(t, testPolicy())
	w.daemon.mu.Lock()
	delete(w.daemon.containers, "task1")
	// A child cannot set swarm labels through the proxy, but one made some other
	// way still must not stand in for the app.
	w.daemon.containers["spoof"] = &fakeContainer{running: true,
		labels: map[string]string{serviceIDLabel: "svc1", OwnerLabel: "app1"},
		mounts: []map[string]any{{"Type": "volume", "Source": "cache-app2", "Target": "/var/lib/autobase"}}}
	w.daemon.mu.Unlock()

	status, message := createWith(t, w, []any{"/var/lib/autobase/ansible:/tmp/ansible"}, nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: the app is not running on this node", message)
}

func TestCreateMountsTheAppsSocketOnlyWhenNestingIsAllowed(t *testing.T) {
	w := newWorld(t, testPolicy())
	status, message := createWith(t, w, []any{SocketPath + ":/var/run/docker.sock"}, nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: mounting the app's socket is not allowed for this app", message)
	status, _ = createWith(t, w, []any{"hp-dapi-sock-app1:/run/hp"}, nil)
	assert.Equal(t, http.StatusForbidden, status)

	policy := testPolicy()
	policy.Allow = []Group{GroupNestedSocket}
	w = newWorld(t, policy)
	status, message = createWith(t, w, []any{SocketPath + ":/var/run/docker.sock"}, nil)
	require.Equal(t, http.StatusCreated, status, message)
	_, mounts := forwardedStorage(t, w)
	assert.Equal(t, []any{map[string]any{
		"Type": "volume", "Source": "hp-dapi-sock-app1", "Target": "/var/run/docker.sock", "ReadOnly": false,
		"VolumeOptions": map[string]any{"NoCopy": true, "Subpath": "docker.sock"},
	}}, mounts)
}

func TestCreateTakesOnlyTheAppsNamedVolumes(t *testing.T) {
	w := newWorld(t, testPolicy())
	status, message := createWith(t, w, []any{"cache-app1:/cache"}, nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: volumes is not allowed for this app", message)

	policy := testPolicy()
	policy.Allow = []Group{GroupVolumes}
	w = newWorld(t, policy)

	status, message = createWith(t, w, []any{"cache-app1:/cache:ro"}, nil)
	require.Equal(t, http.StatusCreated, status, message)
	binds, _ := forwardedStorage(t, w)
	assert.Equal(t, []any{"cache-app1:/cache:ro"}, binds)

	status, message = createWith(t, w, []any{"cache-app2:/cache"}, nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: volume cache-app2 is not one this app created", message)

	// act names volumes in Binds and never creates them.
	status, message = createWith(t, w, []any{"act-toolcache:/opt/hostedtoolcache"}, nil)
	require.Equal(t, http.StatusCreated, status, message)
	w.daemon.mu.Lock()
	assert.Equal(t, map[string]string{OwnerLabel: "app1"}, w.daemon.volumes["act-toolcache"])
	w.daemon.mu.Unlock()

	status, _ = createWith(t, w, nil, []any{map[string]any{"Type": "volume", "Source": "cache-app2", "Target": "/c"}})
	assert.Equal(t, http.StatusForbidden, status)
	status, message = createWith(t, w, nil, []any{map[string]any{"Type": "volume", "Source": "cache-app1",
		"Target": "/c", "VolumeOptions": map[string]any{"DriverConfig": map[string]any{"Name": "local",
			"Options": map[string]any{"device": "/", "o": "bind", "type": "none"}}}}})
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: Mount.VolumeOptions.DriverConfig is not allowed", message)
	status, message = createWith(t, w, nil, []any{map[string]any{"Type": "volume", "Source": "cache-app1",
		"Target": "/c", "VolumeOptions": map[string]any{"Subpath": "../../etc"}}})
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: volume subpath ../../etc leaves the volume", message)
}

func TestCreateRefusesBindModesThatReachTheHost(t *testing.T) {
	policy := testPolicy()
	policy.Allow = []Group{GroupVolumes}
	w := newWorld(t, policy)
	for _, bind := range []string{"cache-app1:/c:z", "cache-app1:/c:Z", "cache-app1:/c:rshared", "a:b:c:d", "/x"} {
		status, _ := createWith(t, w, []any{bind}, nil)
		assert.Equal(t, http.StatusForbidden, status, bind)
	}
}

func TestCreateTakesTmpfsAndRefusesOtherMountTypes(t *testing.T) {
	w := newWorld(t, testPolicy())
	status, message := createWith(t, w, nil, []any{map[string]any{"Type": "tmpfs", "Target": "/scratch",
		"TmpfsOptions": map[string]any{"SizeBytes": 1 << 20}}})
	require.Equal(t, http.StatusCreated, status, message)

	for _, kind := range []string{"npipe", "cluster", "image"} {
		status, message = createWith(t, w, nil, []any{map[string]any{"Type": kind, "Source": "x", "Target": "/x"}})
		assert.Equal(t, http.StatusForbidden, status, kind)
		assert.Equal(t, "hivepaas: mount type "+kind+" is not allowed", message)
	}
	status, message = createWith(t, w, nil, []any{map[string]any{"Type": "bind",
		"Source": "/var/lib/autobase/ansible", "Target": "/x", "BindOptions": map[string]any{"Propagation": "rshared"}}})
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: Mount.BindOptions is not allowed", message)
}
```

- [ ] **Step 3: Run the tests to see them fail**

Run: `go test ./hivepaas_app/pkg/dockerproxy/...`
Expected: FAIL: the shared-directory create answers 403 `HostConfig.Binds is not allowed`.

- [ ] **Step 4: Write storage**

Add to `hivepaas_app/pkg/dockerproxy/daemon.go`, importing `"bytes"` and `"io"`:

```go
// maxErrorDetail is how much of a refusal from the daemon is kept to explain it.
const maxErrorDetail = 512

// createVolume creates a volume with labels, and fails unless the daemon
// accepts it.
func (d *daemon) createVolume(ctx context.Context, name string, labels map[string]string) error {
	const path = "/volumes/create"
	raw, err := json.Marshal(map[string]any{"Name": name, fieldLabels: labels})
	if err != nil {
		return fmt.Errorf("POST %s: %w", path, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+daemonHost+path, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("POST %s: %w", path, err)
	}
	req.Header.Set("Content-Type", contentTypeJSON)
	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("POST %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorDetail))
		return fmt.Errorf("%w: POST %s answered %d: %s", errDaemon, path, resp.StatusCode, bytes.TrimSpace(detail))
	}
	return nil
}
```

`hivepaas_app/pkg/dockerproxy/storage.go`:

```go
package dockerproxy

import (
	"context"
	"net/url"
	"path"
	"slices"
	"strings"
)

const (
	mountTypeVolume = "volume"
	mountTypeBind   = "bind"
	mountTypeTmpfs  = "tmpfs"

	// A bind is source:target, or source:target:modes.
	bindMinParts = 2
	bindMaxParts = 3
)

var (
	// bindModes are the modes a child may give a bind. Relabeling (z, Z) and
	// propagation act on the host's side of the mount.
	bindModes          = []string{"ro", "rw", "nocopy"}
	mountFields        = []string{"Type", "Source", "Target", "ReadOnly", "Consistency", "VolumeOptions", "TmpfsOptions"}
	volumeOptionFields = []string{"NoCopy", fieldLabels, "Subpath"}
)

// taskMount is one mount of the app's own task, as its container reports it.
type taskMount struct {
	Type          string
	Source        string
	Target        string
	VolumeOptions *struct {
		Subpath string
	}
}

// rewriteStorage turns every bind and mount of a child into one the app may give
// it, or refuses. A path becomes a mount of the directory the app itself has
// there; a named volume must be the app's.
func (p *Proxy) rewriteStorage(ctx context.Context, policy *Policy, host map[string]any) error {
	var mounts []any
	for _, raw := range list(host["Mounts"]) {
		mount, err := p.mount(ctx, policy, object(raw))
		if err != nil {
			return err
		}
		mounts = append(mounts, mount)
	}
	var binds []any
	for _, raw := range list(host["Binds"]) {
		spec := text(raw)
		mount, err := p.bind(ctx, policy, spec)
		if err != nil {
			return err
		}
		if mount == nil {
			binds = append(binds, spec)
		} else {
			mounts = append(mounts, mount)
		}
	}
	host["Binds"] = binds
	host["Mounts"] = mounts
	return nil
}

// bind judges one entry of Binds. A named volume stays a bind; a path comes back
// as the mount that replaces it.
func (p *Proxy) bind(ctx context.Context, policy *Policy, spec string) (map[string]any, error) {
	parts := strings.Split(spec, ":")
	if len(parts) < bindMinParts || len(parts) > bindMaxParts {
		return nil, refusef("bind %q is not source:target[:modes]", spec)
	}
	source, target := parts[0], parts[1]
	readOnly := false
	if len(parts) == bindMaxParts {
		for _, mode := range strings.Split(parts[2], ",") {
			if !slices.Contains(bindModes, mode) {
				return nil, refusef("bind mode %s is not allowed", mode)
			}
			readOnly = readOnly || mode == "ro"
		}
	}
	if strings.HasPrefix(source, "/") {
		return p.hostPath(ctx, policy, source, target, readOnly)
	}
	return nil, p.namedVolume(ctx, policy, source)
}

// mount judges one entry of Mounts, and returns what replaces it.
func (p *Proxy) mount(ctx context.Context, policy *Policy, m map[string]any) (map[string]any, error) {
	if err := checkFields("Mount", m, mountFields, nil); err != nil {
		return nil, err
	}
	switch kind := text(m["Type"]); kind {
	case mountTypeTmpfs:
		return m, nil
	case mountTypeVolume:
		options := object(m["VolumeOptions"])
		if err := checkFields("Mount.VolumeOptions", options, volumeOptionFields, nil); err != nil {
			return nil, err
		}
		if subpath := text(options["Subpath"]); subpath != "" && !localPath(subpath) {
			return nil, refusef("volume subpath %s leaves the volume", subpath)
		}
		if source := text(m["Source"]); source != "" {
			if err := p.namedVolume(ctx, policy, source); err != nil {
				return nil, err
			}
		}
		return m, nil
	case mountTypeBind:
		readOnly, _ := m["ReadOnly"].(bool)
		return p.hostPath(ctx, policy, text(m["Source"]), text(m["Target"]), readOnly)
	default:
		return nil, refusef("mount type %s is not allowed", kind)
	}
}

// hostPath turns a bind of a path into a mount of the directory the app itself
// has at that path, or of the app's socket. Nothing else of the host is
// reachable.
func (p *Proxy) hostPath(
	ctx context.Context, policy *Policy, source, target string, readOnly bool,
) (map[string]any, error) {
	clean := path.Clean(source)
	if clean == SocketPath {
		if !policy.allows(GroupNestedSocket) {
			return nil, refusef("mounting the app's socket is not allowed for this app")
		}
		return volumeMount(policy.SocketVolume, SocketFile, target, readOnly), nil
	}
	dir := longestCover(policy.SharedDirs, clean)
	if dir == "" {
		return nil, refusef("%s is not a shared directory of this app", source)
	}
	mounts, err := p.taskMounts(ctx, policy)
	if err != nil {
		return nil, err
	}
	var best *taskMount
	for i := range mounts {
		m := &mounts[i]
		if m.Type == mountTypeVolume && covers(m.Target, dir) && (best == nil || len(m.Target) > len(best.Target)) {
			best = m
		}
	}
	if best == nil {
		return nil, refusef("shared directory %s is not on a volume of the app", dir)
	}
	subpath := ""
	if best.VolumeOptions != nil {
		subpath = best.VolumeOptions.Subpath
	}
	rest := strings.TrimPrefix(clean, strings.TrimSuffix(best.Target, "/"))
	return volumeMount(best.Source, strings.TrimPrefix(path.Join(subpath, rest), "/"), target, readOnly), nil
}

// taskMounts are the mounts of the app's own task on this node. They are read
// from the task's container rather than from the service because a worker node
// cannot read services, and the child is being created on this node, beside
// the task.
func (p *Proxy) taskMounts(ctx context.Context, policy *Policy) ([]taskMount, error) {
	var tasks []struct {
		ID     string `json:"Id"`
		Labels map[string]string
	}
	query := "/containers/json?filters=" + labelFilter(serviceIDLabel, policy.ServiceID)
	if _, err := p.daemon.get(ctx, query, &tasks); err != nil {
		return nil, err
	}
	for _, task := range tasks {
		// A container carrying the owner label is a child, whatever else it says.
		if task.Labels[OwnerLabel] != "" {
			continue
		}
		var info struct {
			HostConfig struct {
				Mounts []taskMount
			}
		}
		found, err := p.daemon.get(ctx, "/containers/"+url.PathEscape(task.ID)+"/json", &info)
		if err != nil {
			return nil, err
		}
		if found {
			return info.HostConfig.Mounts, nil
		}
	}
	return nil, refusef("the app is not running on this node")
}

// namedVolume checks a volume a child names: the app's own, created for it when
// it does not exist yet, since act names volumes in Binds and never creates
// them.
func (p *Proxy) namedVolume(ctx context.Context, policy *Policy, name string) error {
	if name == "" {
		return refusef("a volume must be named")
	}
	if name == policy.SocketVolume {
		if !policy.allows(GroupNestedSocket) {
			return refusef("mounting the app's socket is not allowed for this app")
		}
		return nil
	}
	if !policy.allows(GroupVolumes) {
		return refusef("%s is not allowed for this app", GroupVolumes)
	}
	info, found, err := p.inspect(ctx, kindVolumes, name)
	if err != nil {
		return err
	}
	if !found {
		return p.daemon.createVolume(ctx, name, map[string]string{OwnerLabel: policy.AppID})
	}
	if info.Labels[OwnerLabel] != policy.AppID {
		return refusef("volume %s is not one this app created", name)
	}
	return nil
}

// volumeMount is a mount of a directory of a volume. NoCopy keeps docker from
// copying what an image holds at the target into the app's directory.
func volumeMount(source, subpath, target string, readOnly bool) map[string]any {
	options := map[string]any{"NoCopy": true}
	if subpath != "" {
		options["Subpath"] = subpath
	}
	return map[string]any{
		"Type": mountTypeVolume, "Source": source, "Target": target, "ReadOnly": readOnly,
		"VolumeOptions": options,
	}
}

// covers reports whether p is dir or lies below it.
func covers(dir, p string) bool {
	if dir == "" {
		return false
	}
	return p == dir || strings.HasPrefix(p, strings.TrimSuffix(dir, "/")+"/")
}

func longestCover(dirs []string, p string) string {
	best := ""
	for _, dir := range dirs {
		if covers(dir, p) && len(dir) > len(best) {
			best = dir
		}
	}
	return best
}

// localPath reports whether a subpath stays inside its volume.
func localPath(p string) bool {
	clean := path.Clean(p)
	return !path.IsAbs(clean) && clean != ".." && !strings.HasPrefix(clean, "../")
}
```

In `hivepaas_app/pkg/dockerproxy/create.go`, add `"Binds", "Mounts"` at the front of `hostFields`, and change its comment's first sentence to say storage is judged separately:

```go
	// hostFields are the fields of a HostConfig a child may set; Binds and Mounts
	// are then judged by rewriteStorage. Everything that reaches past the
	// container - Privileged, CapAdd, Devices, the host's namespaces, SecurityOpt,
	// Runtime, Sysctls, CgroupParent, VolumesFrom, Links, PortBindings - is
	// absent, and so refused.
	hostFields = []string{"Binds", "Mounts", "NetworkMode", "RestartPolicy", "AutoRemove", "Memory",
		"MemorySwap", "MemoryReservation", "NanoCpus", "CpuQuota", "CpuPeriod", "CpuShares", "PidsLimit",
		"ShmSize", "Dns", "DnsOptions", "DnsSearch", "ExtraHosts", "LogConfig", "Init", "ReadonlyRootfs", "Tmpfs",
		"CapDrop", "Ulimits", "GroupAdd", "ConsoleSize", "Isolation"}
```

and in `checkCreate`, before `rewriteNetwork`:

```go
	if err := p.rewriteStorage(ctx, policy, host); err != nil {
		return err
	}
```

- [ ] **Step 5: Run the tests to see them pass, then lint**

Run: `go test ./hivepaas_app/pkg/dockerproxy/... && golangci-lint run ./hivepaas_app/pkg/dockerproxy/...`
Expected: `ok`, `0 issues`.

- [ ] **Step 6: Commit**

```bash
git add hivepaas_app/pkg/dockerproxy
git commit -m "feat(dockerproxy): give a child the app's shared directories and nothing else of the host

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 7: What Go clients send, the fuzz test, the whole-repo gates

**Files:**
- Test: `hivepaas_app/pkg/dockerproxy/sdk_test.go`
- Test: `hivepaas_app/pkg/dockerproxy/fuzz_test.go`

**Interfaces:**
- Consumes: everything above.

- [ ] **Step 1: Write the tests**

`hivepaas_app/pkg/dockerproxy/sdk_test.go`:

```go
package dockerproxy

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sdkCreate is a create as a Go client - act, Woodpecker, Autobase - sends it:
// whole structs, with every field it did not set at its zero value.
func sdkCreate(t *testing.T, cfg *container.Config, host *container.HostConfig,
	networking *network.NetworkingConfig) []byte {
	t.Helper()
	raw, err := json.Marshal(container.CreateRequest{Config: cfg, HostConfig: host, NetworkingConfig: networking})
	require.NoError(t, err)
	return raw
}

func TestCreateTakesWhatGoClientsSend(t *testing.T) {
	policy := testPolicy()
	policy.Images = []string{"*"}
	policy.Allow = []Group{GroupVolumes, GroupNetworks, GroupNestedSocket}
	w := newWorld(t, policy)

	// act: a job container on the network it made, a cache volume, the socket.
	act := sdkCreate(t,
		&container.Config{Image: "node:20", Entrypoint: []string{"/bin/sleep", "10800"}},
		&container.HostConfig{
			NetworkMode: "jobnet",
			Binds:       []string{"act-toolcache:/opt/hostedtoolcache", SocketPath + ":/var/run/docker.sock"},
		},
		&network.NetworkingConfig{EndpointsConfig: map[string]*network.EndpointSettings{
			"jobnet": {Aliases: []string{"job"}},
		}})
	status, raw := w.do(t, http.MethodPost, "/v1.44/containers/create", act)
	require.Equal(t, http.StatusCreated, status, string(raw))
	path, _ := w.forwarded(t, http.MethodPost, "/containers/create")
	assert.Equal(t, "/v1.45/containers/create", path)

	// Autobase: host networking and the log directory it shares with the playbook.
	autobase := sdkCreate(t,
		&container.Config{Image: "autobase/automation:2.11.0", Tty: true},
		&container.HostConfig{NetworkMode: "host", Mounts: []mount.Mount{{
			Type: mount.TypeBind, Source: "/var/lib/autobase/ansible", Target: "/tmp/ansible",
		}}},
		nil)
	status, raw = w.do(t, http.MethodPost, "/v1.47/containers/create", autobase)
	require.Equal(t, http.StatusCreated, status, string(raw))
	_, body := w.forwarded(t, http.MethodPost, "/containers/create")
	host := object(body["HostConfig"])
	assert.Equal(t, "hp-dapi-app1", host["NetworkMode"])
	assert.Equal(t, "app1-key/ansible", object(object(list(host["Mounts"])[0])["VolumeOptions"])["Subpath"])
}
```

`hivepaas_app/pkg/dockerproxy/fuzz_test.go`:

```go
package dockerproxy

import (
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// FuzzCreate sends the proxy arbitrary creates. Whatever it lets through must
// be confined: nothing that reaches past the container, no path of the host, no
// network or volume of anyone else, and the owner label.
func FuzzCreate(f *testing.F) {
	for _, name := range []string{"cli-create-plain.json", "cli-create-bind-host.json"} {
		raw, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			f.Fatal(err)
		}
		f.Add(raw)
	}
	f.Add([]byte(`{"Image":"alpine","HostConfig":{"Privileged":true}}`))
	f.Add([]byte(`{"Image":"alpine","HostConfig":{"Binds":["cache-app1:/c","/var/lib/autobase/ansible:/a"]}}`))
	f.Add([]byte(`{"Image":"alpine","HostConfig":{"Mounts":[{"Type":"volume","Source":"cache-app2","Target":"/x"}]}}`))
	f.Add([]byte(`{"Image":"alpine","HostConfig":{"NetworkMode":"jobnet","CpuQuota":1,"CpuPeriod":1000}}`))

	policy := testPolicy()
	policy.Allow = []Group{GroupVolumes, GroupNetworks, GroupNestedSocket}
	w := newWorld(f, policy)

	f.Fuzz(func(t *testing.T, raw []byte) {
		status, _ := w.do(t, http.MethodPost, createPath, raw)
		if status != http.StatusCreated {
			return
		}
		_, body := w.forwarded(t, http.MethodPost, "/containers/create")
		assertConfined(t, policy, body)
	})
}

// reachingFields are the HostConfig fields that reach past the container.
var reachingFields = []string{"Privileged", "CapAdd", "Devices", "DeviceRequests", "DeviceCgroupRules",
	"PidMode", "IpcMode", "UTSMode", "UsernsMode", "CgroupnsMode", "Cgroup", "SecurityOpt", "Runtime",
	"Sysctls", "CgroupParent", "OomKillDisable", "VolumesFrom", "Links", "PortBindings", "PublishAllPorts",
	"VolumeDriver"}

func assertConfined(t *testing.T, policy *Policy, body map[string]any) {
	t.Helper()
	host := object(body["HostConfig"])
	for _, field := range reachingFields {
		assert.True(t, isZero(host[field]), "HostConfig.%s reached the daemon: %v", field, host[field])
	}
	assert.Nil(t, host["MaskedPaths"])
	assert.Nil(t, host["ReadonlyPaths"])
	for _, bind := range list(host["Binds"]) {
		source := strings.Split(text(bind), ":")[0]
		assert.False(t, strings.HasPrefix(source, "/"), "a path of the host reached the daemon: %v", bind)
		assert.NotEqual(t, "cache-app2", source)
	}
	for _, raw := range list(host["Mounts"]) {
		m := object(raw)
		assert.Contains(t, []string{"volume", "tmpfs"}, text(m["Type"]), "mount %v", m)
		assert.NotEqual(t, "cache-app2", text(m["Source"]))
	}
	mode := text(host["NetworkMode"])
	usable := []string{"none", policy.Network, "jobnet", "n-job", "n-app", "proj_env_net", "n-env"}
	assert.True(t, slices.Contains(usable, mode), "network mode %q", mode)
	assert.Equal(t, policy.AppID, text(object(body["Labels"])[OwnerLabel]))
}
```

- [ ] **Step 2: Run the tests**

Run: `go test ./hivepaas_app/pkg/dockerproxy/...`
Expected: `ok`. If `TestCreateTakesWhatGoClientsSend` fails on a field the SDK sends, check it against the SDK's struct: a field it always sends with a value that means "unset" goes into `neutralValues` with a comment naming the client. A field it sends with a real value is refused on purpose.

- [ ] **Step 3: Fuzz**

Run: `go test ./hivepaas_app/pkg/dockerproxy/ -run '^$' -fuzz FuzzCreate -fuzztime 60s`
Expected: `PASS`, no failing input written to `testdata/fuzz/`. A failing input is a hole: fix the rule, keep the input in `testdata/fuzz/FuzzCreate/` as a regression, and run again.

- [ ] **Step 4: Whole-repo gates**

Run: `go build ./... && golangci-lint run ./... && go test ./...`
Expected: build succeeds, `0 issues`, all packages `ok`.

- [ ] **Step 5: Commit**

```bash
git add hivepaas_app/pkg/dockerproxy
git commit -m "test(dockerproxy): what Go clients send, and a fuzz test that what passes is confined

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

- [ ] **Step 6: Merge**

```bash
git checkout main
git merge --no-ff feat/docker-api-engine -m "Merge branch 'feat/docker-api-engine'

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
go build ./... && go test ./hivepaas_app/pkg/dockerproxy/...
git branch -d feat/docker-api-engine
```
