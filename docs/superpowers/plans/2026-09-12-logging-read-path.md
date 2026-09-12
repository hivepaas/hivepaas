# Logging: read path and dashboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let an operator turn logging on from the dashboard and let anyone who can read an app see that app's stored logs - including logs of containers that no longer exist - without ever being able to see another app's.

**Architecture:** First a fix: plan 2's collector and backend share no network, so nothing is ever ingested. Then the read path: the API takes structured parameters, `services/logging/victorialogs` is the only place LogsQL is written, and the app's identity is put into every query by the service layer, never by the caller. The dashboard gets a logging settings page and a History tab beside the existing live logs.

**Tech Stack:** Go, VictoriaLogs v1.52.0 HTTP query API (`/select/logsql/query`, NDJSON), docker swarm overlay networks, React 19 + TypeScript + react-hook-form + zod + TanStack Query (dashboard).

**Spec:** `docs/superpowers/specs/2026-09-12-logging-design.md` - sections 6 (Read path) and 7 (Provenance) are what this plan implements. Its "What implementation changed" section overrides earlier sections.

## Global Constraints

- Go repo: `/Users/tnt/go/src/github.com/hivepaas/hivepaas`, branch `logging-read-path`. Dashboard repo: `/Users/tnt/go/src/github.com/hivepaas/hivepaas-dashboard`, branch `logging-read-path`. Run every command from the repo root it belongs to, using absolute paths.
- Only `github.com/stretchr/testify/assert` is vendored. **`testify/require` is NOT.** Use `assert`, plus `t.Fatal`/`t.Fatalf` where execution must stop.
- Never add a dependency, in either repo.
- `services/logging` and everything under it may import **only** `hivepaas_app/hperrors` and `hivepaas_app/pkg/*` (enforced by `services/logging/loggingmodel/types_test.go`).
- Errors are declared locally with `hperrors.NewErr(base, "ERR_X")` and translated in `hivepaas_app/pkg/translation/messages/en/errors.logging.en.toml`. `go run ./tools/errcodelint` must exit 0: every declared code translated, and every declared identifier referenced somewhere.
- **No raw LogsQL from any client** (spec section 6.3). No request field, anywhere, carries query text.
- **Every `unpack_json` HivePaaS emits carries `result_prefix`** (spec section 7).
- **The scope of a query is set by the server**: the app's identity comes from the app HivePaaS loaded, never from a request field. A query with no scope is refused.
- VictoriaLogs never gets a published port, and never joins `hivepaas_net` (the routing network every publicly exposed app is on). It has no authentication.
- DI is fx-style: constructors in `hivepaas_app/registry/provides.go` are wired by parameter type. `go test ./hivepaas_app/cmd/internal/ -run TestFxGraphResolves` proves the graph resolves.
- Lint: `golangci-lint run ./<changed packages>/...` must print `0 issues.` - commit only after it does. American spelling (`misspell`).
- Dashboard has no unit test runner. Its gate is `npx tsc --noEmit -p .`, `npx eslint <changed files> --max-warnings 0`, `npx prettier --check <changed files>`.
- Commit messages end with `Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>`.

### Verified facts this plan relies on

Measured on 2026-09-12 against Docker 29.7.2 (swarm, one node) and `victoriametrics/victoria-logs:v1.52.0`. Do not redesign around a contradicting guess.

- **Two swarm services with no shared overlay network cannot reach each other by name**: `curl http://<service>:9428/health` exits 6 (could not resolve host). On a shared overlay it returns 200. Plan 2's `deploy.go` passes no `Networks`, so vlagent cannot reach VictoriaLogs.
- The HivePaaS API service `hivepaas_app` is attached to `hivepaas_local_net` (the stack's private network: db, redis, worker) and `hivepaas_net` (routing). `hpappservice.Service.GetHpAppSwarmService(ctx)` returns it; its `Spec.TaskTemplate.Networks[].Target` are network IDs. `networkservice.Service.GetGlobalRoutingNetworkID(ctx)` returns `hivepaas_net`'s ID.
- Docker's network `name` filter matches substrings; compare `Name` exactly after listing. `client.NetworkListResult{Items []network.Summary}`, and `network.Summary` embeds `network.Network` (fields `ID`, `Name`).
- LogsQL string literals use Go quoting: `strconv.Quote` output is parsed correctly, including `\"`, `\\`, tabs and non-ASCII. An injected `x" OR "attrs.hivepaas.app.id":="APP2` inside a quoted literal stays one literal and matches nothing. An unescaped `"""` is a parse error.
- Exact field match: `"attrs.hivepaas.app.id":="APP1"`. Stream: `stream:in("stderr")`. Case-insensitive substring over the message: `~"(?i)` + `regexp.QuoteMeta(text)` + `"` (a phrase filter such as `"info li"` does NOT match `info line`: phrases match whole words).
- `| unpack_json from _msg fields (level) result_prefix "app."` - this order; `result_prefix` before `fields` is a parse error. Lines that are not JSON pass through without `app.level`.
- Level filter after unpacking: `| filter "app.level":~"(?i)^(error|warn)$"` works (matches `ERROR` and `warn`). `:=~` is a parse error.
- `| sort by (_time desc) | limit N | fields _time, _msg, stream, app.level` returns newest first, only those fields, one JSON object per line, every value a string. Absent fields are omitted from the object.
- Query parameters `start` / `end` (RFC3339) bound the time range.
- `/insert/jsonline` needs `Content-Type: application/stream+json`. Without it the server answers 200 and **stores nothing**.
- `/select/logsql/tail` exists (live tail). This plan does not use it: live logs already stream from docker.

---

## File Structure

| File | Responsibility |
|---|---|
| `hivepaas_app/service/loggingservice/loggingserviceimpl/network.go` | The logging overlay network; the API's private networks |
| `hivepaas_app/service/loggingservice/loggingserviceimpl/deploy.go` | Attach networks when deploying (modify) |
| `hivepaas_app/service/loggingservice/loggingserviceimpl/service.go` | New dependencies (modify) |
| `services/logging/loggingmodel/types.go` | Structured `QueryReq`, `LogEntry`, `QueryResp` (modify) |
| `services/logging/victorialogs/query.go` | `BuildQuery` - the only LogsQL writer |
| `services/logging/victorialogs/client.go` | `Query` over HTTP, TLS options (modify) |
| `services/logging/vlagent/configure.go` | `AttrField` - where a container label is stored (modify) |
| `hivepaas_app/service/loggingservice/service.go` | `QueryAppLogs`, `AppHistory` on the interface (modify) |
| `hivepaas_app/service/loggingservice/loggingserviceimpl/query.go` | Scope, endpoint, availability |
| `hivepaas_app/usecase/appuc/appdto/logs_history.go` | Request/response of the history API |
| `hivepaas_app/usecase/appuc/logs_history.go` | The use case and frame mapping |
| `hivepaas_app/interface/api/handler/apphandler/logs.go` | Handler (modify) |
| dashboard `src/application/modules/system-settings/**/hivepaas-logging-settings*` | Settings API layer and page |
| dashboard `src/application/modules/projects/**/logs/**` | History API and tab |

---

### Task 1: The collector and the backend share a network

**Files:**
- Create: `hivepaas_app/service/loggingservice/loggingserviceimpl/network.go`
- Create: `hivepaas_app/service/loggingservice/loggingserviceimpl/network_test.go`
- Modify: `hivepaas_app/service/loggingservice/loggingserviceimpl/service.go`
- Modify: `hivepaas_app/service/loggingservice/loggingserviceimpl/config.go`
- Modify: `hivepaas_app/service/loggingservice/loggingserviceimpl/deploy.go`
- Modify: `hivepaas_app/service/loggingservice/loggingserviceimpl/deploy_test.go`
- Modify: `hivepaas_app/service/loggingservice/errors.go`
- Modify: `hivepaas_app/pkg/translation/messages/en/errors.logging.en.toml`

**Interfaces:**
- Consumes: `hpappservice.Service.GetHpAppSwarmService(ctx) (*swarm.Service, error)`, `networkservice.Service.GetGlobalRoutingNetworkID(ctx) (string, error)`, `docker.Manager.NetworkList(ctx, ...docker.NetworkListOption) (*client.NetworkListResult, error)`, `docker.Manager.NetworkCreate(ctx, name string, ...docker.NetworkCreateOption) (*client.NetworkCreateResult, error)`.
- Produces: `const NetworkLogging = "hivepaas_logging_net"`; `func backendBaseURL() string` returning `http://hivepaas-victoria-logs:9428` (Task 5 uses it); `loggingservice.ErrAPINetworkMissing`; `New(...)` gains `hpAppService hpappservice.Service, networkService networkservice.Service`.

Why two networks for the backend: the collector reaches VictoriaLogs over `hivepaas_logging_net`, which only the logging stack joins. The API reaches it over its own private network(s). Putting VictoriaLogs on `hivepaas_net` instead would hand every publicly exposed app an unauthenticated view of every project's logs. Putting vlagent on the API's private network would let a process running on every node reach the database. Neither is done. The API service itself is not modified - it would restart the process serving the request.

- [ ] **Step 1: Declare the error**

Append to the `var (...)` block in `hivepaas_app/service/loggingservice/errors.go`:

```go
	ErrAPINetworkMissing  = hperrors.NewErr(hperrors.ErrActionFailed, "ERR_LOGGING_SVC_API_NETWORK_MISSING")
```

Append to `errors.logging.en.toml`:

```toml
ERR_LOGGING_SVC_API_NETWORK_MISSING = "HivePaaS's own service has no private network the logging backend can join"
```

- [ ] **Step 2: Write the failing tests**

Create `hivepaas_app/service/loggingservice/loggingserviceimpl/network_test.go`:

```go
package loggingserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	testRoutingNetID = "net-routing"
	testLocalNetID   = "net-local"
)

// fakeHpApp is HivePaaS's own service, attached to the routing network and one
// private network - the shape deployment/local/hivepaas.yaml gives it.
type fakeHpApp struct {
	hpappservice.Service
	networks []string
}

func (f *fakeHpApp) GetHpAppSwarmService(_ context.Context) (*swarm.Service, error) {
	svc := &swarm.Service{}
	for _, n := range f.networks {
		svc.Spec.TaskTemplate.Networks = append(svc.Spec.TaskTemplate.Networks,
			swarm.NetworkAttachmentConfig{Target: n})
	}
	return svc, nil
}

type fakeNetworkService struct {
	networkservice.Service
}

func (f *fakeNetworkService) GetGlobalRoutingNetworkID(_ context.Context) (string, error) {
	return testRoutingNetID, nil
}

func (f *fakeDocker) NetworkList(
	_ context.Context, _ ...docker.NetworkListOption,
) (*client.NetworkListResult, error) {
	return &client.NetworkListResult{Items: f.networks}, nil
}

func (f *fakeDocker) NetworkCreate(
	_ context.Context, name string, opts ...docker.NetworkCreateOption,
) (*client.NetworkCreateResult, error) {
	o := client.NetworkCreateOptions{}
	for _, opt := range opts {
		opt(&o)
	}
	f.networksCreated = append(f.networksCreated, o)
	id := "id-" + name
	f.networks = append(f.networks, network.Summary{Network: network.Network{ID: id, Name: name}})
	return &client.NetworkCreateResult{ID: id}, nil
}

func networkTargets(spec *swarm.ServiceSpec) []string {
	out := []string{}
	for _, n := range spec.TaskTemplate.Networks {
		out = append(out, n.Target)
	}
	return out
}

func TestEnsureLoggingNetworkCreatesAnOverlayOnce(t *testing.T) {
	fd := &fakeDocker{}
	s := newTestService(fd, nil)

	id1, err := s.ensureLoggingNetwork(context.Background())
	assert.NoError(t, err)
	id2, err := s.ensureLoggingNetwork(context.Background())
	assert.NoError(t, err)

	assert.Equal(t, id1, id2)
	if assert.Len(t, fd.networksCreated, 1) {
		assert.Equal(t, docker.NetworkDriverOverlay, fd.networksCreated[0].Driver)
		assert.Equal(t, docker.NetworkScopeSwarm, fd.networksCreated[0].Scope)
		assert.False(t, fd.networksCreated[0].Attachable, "nothing outside the stack may join it")
	}
}

func TestEnsureLoggingNetworkIgnoresANameThatOnlyContainsIt(t *testing.T) {
	// docker's name filter matches substrings
	fd := &fakeDocker{networks: []network.Summary{{Network: network.Network{
		ID: "other", Name: "p1_" + NetworkLogging + "_copy",
	}}}}
	s := newTestService(fd, nil)

	id, err := s.ensureLoggingNetwork(context.Background())
	assert.NoError(t, err)
	assert.NotEqual(t, "other", id)
	assert.Len(t, fd.networksCreated, 1)
}

func TestAPIPrivateNetworksLeavesTheRoutingNetworkOut(t *testing.T) {
	s := newTestService(&fakeDocker{}, nil)

	nets, err := s.apiPrivateNetworks(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, []string{testLocalNetID}, nets)
}

func TestAPIPrivateNetworksRefusesWhenOnlyTheRoutingNetworkIsLeft(t *testing.T) {
	s := newTestService(&fakeDocker{}, nil)
	s.hpAppService = &fakeHpApp{networks: []string{testRoutingNetID}}

	_, err := s.apiPrivateNetworks(context.Background())
	assert.ErrorIs(t, err, loggingservice.ErrAPINetworkMissing)
}

func TestDeployJoinsTheStackOnlyToNetworksItNeeds(t *testing.T) {
	fd := &fakeDocker{}
	s := newTestService(fd, storedSetting(t, enabledConfig()))

	assert.NoError(t, s.Apply(context.Background(), nil))

	var backend, collector *swarm.ServiceSpec
	for _, spec := range fd.created {
		switch spec.Name {
		case ServiceNameBackend:
			backend = spec
		case ServiceNameCollector:
			collector = spec
		}
	}
	if backend == nil || collector == nil {
		t.Fatalf("expected both services, created %d", len(fd.created))
	}
	logNet := "id-" + NetworkLogging
	assert.ElementsMatch(t, []string{logNet, testLocalNetID}, networkTargets(backend))
	assert.Equal(t, []string{logNet}, networkTargets(collector))
	assert.NotContains(t, networkTargets(backend), testRoutingNetID)
}
```

In `deploy_test.go`, add two fields to `fakeDocker`:

```go
	networks        []network.Summary
	networksCreated []client.NetworkCreateOptions
```

and add this helper (import `hivepaas_app/entity` is already there):

```go
// newTestService builds the service with every dependency faked. setting may be
// nil for "never configured".
func newTestService(fd *fakeDocker, setting *entity.Setting) *service {
	return &service{
		dockerManager:  fd,
		settingRepo:    &fakeSettingRepo{setting: setting},
		appRepo:        &fakeAppRepo{},
		hpAppService:   &fakeHpApp{networks: []string{testRoutingNetID, testLocalNetID}},
		networkService: &fakeNetworkService{},
	}
}
```

Then replace **every** `&service{...}` literal in `deploy_test.go` with `newTestService(fd, <the setting that test used>)`, keeping any extra field the literal set by assigning it afterwards. `storedSetting(t, cfg)` is the row-shaped setting helper already in `deploy_test.go` - use it, do not add a second one.

- [ ] **Step 3: Run the tests to verify they fail**

Run: `go test ./hivepaas_app/service/loggingservice/loggingserviceimpl/`
Expected: build failure - `s.ensureLoggingNetwork undefined`, `unknown field hpAppService`.

- [ ] **Step 4: Implement**

In `service.go`, add the two fields and constructor parameters (keep the existing ones and their order; append the new ones at the end):

```go
	hpAppService   hpappservice.Service
	networkService networkservice.Service
```

In `config.go`, add to the `const` block:

```go
	// NetworkLogging is the overlay the collector and the backend talk over.
	// Only the logging stack joins it.
	NetworkLogging = "hivepaas_logging_net"
```

and below `toEndpoint`:

```go
// backendBaseURL is where a HivePaaS-run backend answers, from any service on a
// network it is attached to.
func backendBaseURL() string {
	return fmt.Sprintf("http://%s:%d", ServiceNameBackend, victorialogs.DefaultHTTPPort)
}
```

Create `network.go`:

```go
package loggingserviceimpl

import (
	"context"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// ensureLoggingNetwork returns the overlay the collector and the backend share,
// creating it the first time.
//
// Swarm services resolve each other by name only across a network they both
// join. Without one, vlagent's writes to hivepaas-victoria-logs fail to
// resolve and nothing is ever stored.
//
// Not attachable: a container started by hand has no business on it.
func (s *service) ensureLoggingNetwork(ctx context.Context) (string, error) {
	list, err := s.dockerManager.NetworkList(ctx, func(opts *client.NetworkListOptions) {
		docker.FilterAdd(&opts.Filters, "name", NetworkLogging)
	})
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	// The name filter matches substrings.
	for i := range list.Items {
		if list.Items[i].Name == NetworkLogging {
			return list.Items[i].ID, nil
		}
	}

	resp, err := s.dockerManager.NetworkCreate(ctx, NetworkLogging, func(opts *client.NetworkCreateOptions) {
		opts.Driver = docker.NetworkDriverOverlay
		opts.Scope = docker.NetworkScopeSwarm
		opts.Attachable = false
		opts.Options = map[string]string{docker.NetworkOptionDriverMTU: docker.DefaultOverlayNetworkMTU}
		opts.Labels = map[string]string{LabelManagedBy: "true"}
	})
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return resp.ID, nil
}

// apiPrivateNetworks is every network HivePaaS's own API is on except the
// routing network.
//
// The backend joins these so that the API can query it. The routing network is
// left out because every publicly exposed app is on it, and the backend has no
// authentication.
func (s *service) apiPrivateNetworks(ctx context.Context) ([]string, error) {
	svc, err := s.hpAppService.GetHpAppSwarmService(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	routingID, err := s.networkService.GetGlobalRoutingNetworkID(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	out := []string{}
	for _, n := range svc.Spec.TaskTemplate.Networks {
		if n.Target == routingID || n.Target == base.NetworkGlobalRouting {
			continue
		}
		out = append(out, n.Target)
	}
	if len(out) == 0 {
		return nil, hperrors.Wrap(loggingservice.ErrAPINetworkMissing)
	}
	return out, nil
}
```

In `deploy.go`:
1. At the top of `deploy`, before `deployBackend`:
```go
	logNet, err := s.ensureLoggingNetwork(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
```
2. Pass `logNet` into `deployBackend(ctx, cfg, logNet)` and `deployCollector(ctx, cfg, ingestURL, logNet)`, changing both signatures.
3. In `deployCollector`, add `Networks: []string{logNet},` to its `swarmSpecOpts`.
4. In `deployBackend`, before `toSwarmServiceSpec`:
```go
	apiNets, err := s.apiPrivateNetworks(ctx)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
```
and add `Networks: append([]string{logNet}, apiNets...),` to its `swarmSpecOpts`.
5. Replace the returned ingest URL with `backendBaseURL() + victorialogs.IngestPath`.

The network is not removed by `TearDown`: an empty overlay costs nothing, and removing it while a task is still draining fails.

In `hivepaas_app/registry/provides.go` nothing changes - fx fills the new parameters.

- [ ] **Step 5: Run tests, the fx graph and lint**

```bash
go test ./hivepaas_app/service/loggingservice/... && \
go test ./hivepaas_app/cmd/internal/ -run TestFxGraphResolves && \
go run ./tools/errcodelint && \
golangci-lint run ./hivepaas_app/service/loggingservice/...
```
Expected: all pass, `0 issues.`

- [ ] **Step 6: Commit**

```bash
git add hivepaas_app/service/loggingservice hivepaas_app/pkg/translation/messages/en/errors.logging.en.toml
git commit -m "fix(logging): give the collector and the backend a network to meet on

Swarm services resolve each other only across a shared overlay, and
plan 2 attached none, so vlagent could never reach VictoriaLogs.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

### Task 2: A query is structured, and always scoped

**Files:**
- Modify: `services/logging/loggingmodel/types.go` (the `QueryReq`, `LogEntry`, `QueryResp` block at the end)
- Modify: `services/logging/loggingmodel/errors.go`
- Modify: `services/logging/logging.go`
- Modify: `hivepaas_app/pkg/translation/messages/en/errors.logging.en.toml`

**Interfaces:**
- Produces (used by Tasks 3-5 exactly):

```go
type FieldMatch struct{ Field, Value string }
type QueryReq struct {
	Match    []FieldMatch
	Contains string
	Levels   []string
	Streams  []string
	Start    time.Time
	End      time.Time
	Limit    int
}
type LogEntry struct {
	Time    time.Time
	Message string
	Stream  string
	Level   string
}
type QueryResp struct {
	Entries   []LogEntry
	Truncated bool
}
const MaxQueryLimit = 5000
var ErrQueryScopeRequired // ERR_LOGGING_QUERY_SCOPE_REQUIRED
```
and in package `logging`: `FieldMatch = loggingmodel.FieldMatch`, `MaxQueryLimit`, `ErrQueryScopeRequired` re-exported.

The old `QueryReq.Query string` is removed. A string field would be an invitation to pass LogsQL through.

- [ ] **Step 1: Replace the query types**

In `services/logging/loggingmodel/types.go`, replace everything from `// QueryReq is a search over stored logs.` to the end of the file with:

```go
// MaxQueryLimit caps how many lines one query returns.
const MaxQueryLimit = 5000

// FieldMatch is one field that must equal one value exactly.
type FieldMatch struct {
	Field string
	Value string
}

// QueryReq is a search over stored logs, in parameters rather than query text.
//
// There is deliberately no field for query text. A backend's query language has
// operators that change how a prepended filter binds, so accepting text and
// narrowing it is the SQL-concatenation flaw over again.
type QueryReq struct {
	// Match is the scope. Every entry must hold, and a request without one is
	// refused: the caller - not whoever asked it - decides what may be seen,
	// and puts that here.
	Match []FieldMatch
	// Contains is a case-insensitive substring of the message.
	Contains string
	// Levels keeps lines whose JSON message carries one of these levels,
	// compared case-insensitively. Lines that are not JSON have no level.
	Levels []string
	// Streams keeps lines written to these streams: stdout, stderr.
	Streams []string
	Start   time.Time
	End     time.Time
	// Limit is how many of the newest matching lines to return.
	Limit int
}

// LogEntry is one stored line.
type LogEntry struct {
	Time    time.Time
	Message string
	// Stream is stdout or stderr, empty when the source has no such notion.
	Stream string
	// Level is the message's own level when it is JSON carrying one.
	Level string
}

// QueryResp is what a search found, oldest first.
type QueryResp struct {
	Entries []LogEntry
	// Truncated says the limit was reached: older matching lines exist.
	Truncated bool
}
```

- [ ] **Step 2: Declare the error**

In `services/logging/loggingmodel/errors.go`, under `// Talking to a backend`:

```go
	ErrQueryScopeRequired = hperrors.NewErr(hperrors.ErrBadRequest, "ERR_LOGGING_QUERY_SCOPE_REQUIRED")
```

In `errors.logging.en.toml`, after `ERR_LOGGING_QUERY_INVALID`:

```toml
ERR_LOGGING_QUERY_SCOPE_REQUIRED = "A log query must be limited to what it may read"
```

In `services/logging/logging.go`, add to the error block `ErrQueryScopeRequired = loggingmodel.ErrQueryScopeRequired`, to the type block `FieldMatch = loggingmodel.FieldMatch`, and a const block:

```go
const MaxQueryLimit = loggingmodel.MaxQueryLimit
```

- [ ] **Step 3: Build**

Run: `go build ./... && go vet ./services/logging/...`
Expected: `go build` fails in `services/logging/victorialogs/client.go` only if it referenced `req.Query` - it does not (it ignored the request). If anything else references `QueryReq.Query` or `LogEntry.Fields`, fix it by removing the reference. `go run ./tools/errcodelint` must exit 0 - the re-export in `logging.go` references `ErrQueryScopeRequired`.

- [ ] **Step 4: Commit**

```bash
go test ./services/logging/... && go run ./tools/errcodelint && golangci-lint run ./services/logging/...
git add services/logging hivepaas_app/pkg/translation/messages/en/errors.logging.en.toml
git commit -m "feat(logging): describe a query in parameters, never in query text

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

### Task 3: BuildQuery - the only place LogsQL is written

**Files:**
- Create: `services/logging/victorialogs/query.go`
- Create: `services/logging/victorialogs/query_test.go`

**Interfaces:**
- Consumes: Task 2's `loggingmodel.QueryReq`, `ErrQueryScopeRequired`, `ErrQueryInvalid`, `MaxQueryLimit`.
- Produces: `func BuildQuery(req *loggingmodel.QueryReq) (string, error)`; `const UnpackPrefix = "app."`; `const LevelField = "app.level"`.

- [ ] **Step 1: Write the failing tests**

Create `services/logging/victorialogs/query_test.go`:

```go
package victorialogs

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

const appField = "attrs.hivepaas.app.id"

func scoped(mod func(r *loggingmodel.QueryReq)) *loggingmodel.QueryReq {
	r := &loggingmodel.QueryReq{
		Match: []loggingmodel.FieldMatch{{Field: appField, Value: "APP1"}},
		Limit: 100,
	}
	if mod != nil {
		mod(r)
	}
	return r
}

const tail = ` | unpack_json from _msg fields (level) result_prefix "app."` +
	` | sort by (_time desc) | limit 100 | fields _time, _msg, stream, app.level`

func TestBuildQueryScopeOnly(t *testing.T) {
	q, err := BuildQuery(scoped(nil))
	assert.NoError(t, err)
	assert.Equal(t, `"attrs.hivepaas.app.id":="APP1"`+tail, q)
}

func TestBuildQueryRefusesAnUnscopedRequest(t *testing.T) {
	_, err := BuildQuery(&loggingmodel.QueryReq{Limit: 10})
	assert.ErrorIs(t, err, loggingmodel.ErrQueryScopeRequired)

	_, err = BuildQuery(&loggingmodel.QueryReq{Limit: 10, Match: []loggingmodel.FieldMatch{{Value: "x"}}})
	assert.ErrorIs(t, err, loggingmodel.ErrQueryScopeRequired, "a match with no field scopes nothing")
}

func TestBuildQueryRefusesALimitOutOfRange(t *testing.T) {
	for _, limit := range []int{0, -1, loggingmodel.MaxQueryLimit + 1} {
		_, err := BuildQuery(scoped(func(r *loggingmodel.QueryReq) { r.Limit = limit }))
		assert.ErrorIs(t, err, loggingmodel.ErrQueryInvalid, "limit %d", limit)
	}
}

func TestBuildQueryEveryParameter(t *testing.T) {
	q, err := BuildQuery(scoped(func(r *loggingmodel.QueryReq) {
		r.Contains = "Timeout (x+y)"
		r.Streams = []string{"stderr"}
		r.Levels = []string{"error", "warn"}
	}))
	assert.NoError(t, err)
	assert.Equal(t,
		`"attrs.hivepaas.app.id":="APP1" AND stream:in("stderr") AND ~"(?i)Timeout \\(x\\+y\\)"`+
			` | unpack_json from _msg fields (level) result_prefix "app."`+
			` | filter "app.level":~"(?i)^(error|warn)$"`+
			` | sort by (_time desc) | limit 100 | fields _time, _msg, stream, app.level`,
		q)
}

// Each of these tries to break out of the literal it is placed in. Built with
// Go quoting, each stays a single literal - verified against VictoriaLogs
// v1.52.0, where the first returns no line of APP2.
func TestBuildQueryKeepsHostileInputInsideItsLiteral(t *testing.T) {
	hostile := []string{
		`x" OR "attrs.hivepaas.app.id":="APP2`,
		`") OR ("attrs.hivepaas.app.id":="APP2`,
		`x | delete attrs.hivepaas.app.id`,
		"back\\slash \" and \n newline",
		`"""`,
	}
	for _, h := range hostile {
		q, err := BuildQuery(scoped(func(r *loggingmodel.QueryReq) { r.Contains = h }))
		assert.NoError(t, err)
		assert.Contains(t, q, `"attrs.hivepaas.app.id":="APP1" AND ~`, h)
		// The scope and the pipes after it are untouched: nothing the input
		// held became syntax.
		assert.Equal(t, 4, countOutsideLiterals(q, "|"), h) // unpack_json, sort, limit, fields
		assert.NotContains(t, stripLiterals(q), "APP2", h)
	}

	for _, h := range hostile {
		q, err := BuildQuery(scoped(func(r *loggingmodel.QueryReq) { r.Match[0].Value = h }))
		assert.NoError(t, err)
		assert.NotContains(t, stripLiterals(q), "APP2", h)
	}
	for _, h := range hostile {
		q, err := BuildQuery(scoped(func(r *loggingmodel.QueryReq) { r.Levels = []string{h} }))
		assert.NoError(t, err)
		assert.NotContains(t, stripLiterals(q), "APP2", h)
	}
}

func TestBuildQueryAlwaysPrefixesUnpackedFields(t *testing.T) {
	// Spec section 7: an unprefixed unpack_json lets a line's own JSON
	// overwrite the daemon's attrs.hivepaas.app.id at query time.
	for _, r := range []*loggingmodel.QueryReq{
		scoped(nil),
		scoped(func(r *loggingmodel.QueryReq) { r.Levels = []string{"error"} }),
	} {
		q, err := BuildQuery(r)
		assert.NoError(t, err)
		assert.Contains(t, q, `unpack_json from _msg fields (level) result_prefix "app."`)
		assert.Equal(t, 1, countOutsideLiterals(q, "unpack_json"))
	}
}

// stripLiterals removes every double-quoted literal, honoring backslash
// escapes, leaving only what VictoriaLogs parses as syntax.
func stripLiterals(q string) string {
	out := make([]rune, 0, len(q))
	in, esc := false, false
	for _, c := range q {
		switch {
		case esc:
			esc = false
		case in && c == '\\':
			esc = true
		case c == '"':
			in = !in
		case !in:
			out = append(out, c)
		}
	}
	return string(out)
}

func countOutsideLiterals(q, sub string) int {
	s := stripLiterals(q)
	n := 0
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			n++
		}
	}
	return n
}
```

Exactly four pipes stay outside literals when no level filter is set: `unpack_json`, `sort`, `limit`, `fields`. A fifth would mean input became syntax.

- [ ] **Step 2: Run to verify failure**

Run: `go test ./services/logging/victorialogs/ -run BuildQuery`
Expected: build failure, `undefined: BuildQuery`.

- [ ] **Step 3: Implement**

Create `services/logging/victorialogs/query.go`:

```go
package victorialogs

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

const (
	// UnpackPrefix is put in front of every field read out of a message.
	//
	// Without it, a line printing {"attrs.hivepaas.app.id":"OTHER"} would
	// overwrite the daemon's own field of that name for the rest of the query
	// - measured, not assumed. Nothing may unpack without it.
	UnpackPrefix = "app."

	// LevelField is the message's own level once unpacked.
	LevelField = UnpackPrefix + "level"
)

// BuildQuery turns a structured request into LogsQL.
//
// It is the only place HivePaaS writes LogsQL. Every value from the request is
// placed inside a Go-quoted literal, which LogsQL parses the same way, so no
// value can become syntax; the structure around them is fixed here.
func BuildQuery(req *loggingmodel.QueryReq) (string, error) {
	if len(req.Match) == 0 {
		return "", hperrors.Wrap(loggingmodel.ErrQueryScopeRequired)
	}
	if req.Limit <= 0 || req.Limit > loggingmodel.MaxQueryLimit {
		return "", hperrors.Wrap(loggingmodel.ErrQueryInvalid).
			WithExtraDetail("limit must be between 1 and %d", loggingmodel.MaxQueryLimit)
	}

	filters := make([]string, 0, len(req.Match)+2) //nolint:mnd // streams and contains
	for _, m := range req.Match {
		if m.Field == "" {
			return "", hperrors.Wrap(loggingmodel.ErrQueryScopeRequired)
		}
		filters = append(filters, strconv.Quote(m.Field)+":="+strconv.Quote(m.Value))
	}
	if len(req.Streams) > 0 {
		filters = append(filters, "stream:in("+quoteAll(req.Streams)+")")
	}
	if req.Contains != "" {
		filters = append(filters, "~"+strconv.Quote("(?i)"+regexp.QuoteMeta(req.Contains)))
	}

	var b strings.Builder
	b.WriteString(strings.Join(filters, " AND "))
	b.WriteString(" | unpack_json from _msg fields (level) result_prefix ")
	b.WriteString(strconv.Quote(UnpackPrefix))
	if len(req.Levels) > 0 {
		quoted := make([]string, 0, len(req.Levels))
		for _, l := range req.Levels {
			quoted = append(quoted, regexp.QuoteMeta(l))
		}
		b.WriteString(" | filter ")
		b.WriteString(strconv.Quote(LevelField))
		b.WriteString(":~")
		b.WriteString(strconv.Quote("(?i)^(" + strings.Join(quoted, "|") + ")$"))
	}
	fmt.Fprintf(&b, " | sort by (_time desc) | limit %d | fields _time, _msg, stream, %s", req.Limit, LevelField)
	return b.String(), nil
}

func quoteAll(values []string) string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		out = append(out, strconv.Quote(v))
	}
	return strings.Join(out, ",")
}
```

- [ ] **Step 4: Run tests and lint**

```bash
go test ./services/logging/... && go run ./tools/errcodelint && golangci-lint run ./services/logging/...
```
Expected: PASS, errcodelint exit 0, `0 issues.`

- [ ] **Step 5: Commit**

```bash
git add services/logging/victorialogs/query.go services/logging/victorialogs/query_test.go
git commit -m "feat(logging): build LogsQL from parameters, in one place

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

### Task 4: Query over HTTP, with a regression test against a real VictoriaLogs

**Files:**
- Modify: `services/logging/victorialogs/client.go`
- Modify: `services/logging/victorialogs/client_test.go`
- Create: `services/logging/victorialogs/query_live_test.go`

**Interfaces:**
- Consumes: `BuildQuery`, `QueryPath`, `loggingmodel.Endpoint` (with `TLSSkipVerify`).
- Produces: `(*Client).Query(ctx, *loggingmodel.QueryReq) (*loggingmodel.QueryResp, error)` - entries oldest first, `Truncated` when `len == Limit`.

- [ ] **Step 1: Write the failing httptest tests**

Append to `services/logging/victorialogs/client_test.go` (add imports `io`, `net/http`, `net/http/httptest`, `net/url`, `strings`, `time`, `loggingmodel` as needed):

```go
func TestQuerySendsTheBuiltQueryAndTimeRange(t *testing.T) {
	var got url.Values
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, QueryPath, r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		body, _ := io.ReadAll(r.Body)
		got, _ = url.ParseQuery(string(body))
		auth = r.Header.Get("Authorization")
		_, _ = io.WriteString(w,
			`{"_time":"2026-09-12T08:00:02.5Z","_msg":"second","stream":"stderr","app.level":"error"}`+"\n"+
				`{"_time":"2026-09-12T08:00:01Z","_msg":"first","stream":"stdout"}`+"\n")
	}))
	defer srv.Close()

	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: srv.URL, BearerToken: "tok"}})
	start := time.Date(2026, 9, 12, 7, 0, 0, 0, time.UTC)
	end := time.Date(2026, 9, 12, 9, 0, 0, 0, time.UTC)
	resp, err := c.Query(context.Background(), &loggingmodel.QueryReq{
		Match: []loggingmodel.FieldMatch{{Field: "attrs.hivepaas.app.id", Value: "APP1"}},
		Start: start, End: end, Limit: 2,
	})
	assert.NoError(t, err)

	want, _ := BuildQuery(&loggingmodel.QueryReq{
		Match: []loggingmodel.FieldMatch{{Field: "attrs.hivepaas.app.id", Value: "APP1"}}, Limit: 2,
	})
	assert.Equal(t, want, got.Get("query"))
	assert.Equal(t, "2026-09-12T07:00:00Z", got.Get("start"))
	assert.Equal(t, "2026-09-12T09:00:00Z", got.Get("end"))
	assert.Equal(t, "Bearer tok", auth)

	if assert.Len(t, resp.Entries, 2) {
		assert.Equal(t, "first", resp.Entries[0].Message, "oldest first")
		assert.Equal(t, "stdout", resp.Entries[0].Stream)
		assert.Equal(t, "", resp.Entries[0].Level)
		assert.Equal(t, "second", resp.Entries[1].Message)
		assert.Equal(t, "error", resp.Entries[1].Level)
		assert.Equal(t, 500*time.Millisecond, resp.Entries[1].Time.Sub(resp.Entries[0].Time)-time.Second)
	}
	assert.True(t, resp.Truncated, "as many lines as the limit means older ones may exist")
}

func TestQueryMapsABadRequestToQueryInvalid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "cannot parse query", http.StatusBadRequest)
	}))
	defer srv.Close()

	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: srv.URL}})
	_, err := c.Query(context.Background(), &loggingmodel.QueryReq{
		Match: []loggingmodel.FieldMatch{{Field: "f", Value: "v"}}, Limit: 1,
	})
	assert.ErrorIs(t, err, loggingmodel.ErrQueryInvalid)
}

func TestQueryMapsAnUnreachableBackend(t *testing.T) {
	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: "http://127.0.0.1:1"}})
	_, err := c.Query(context.Background(), &loggingmodel.QueryReq{
		Match: []loggingmodel.FieldMatch{{Field: "f", Value: "v"}}, Limit: 1,
	})
	assert.ErrorIs(t, err, loggingmodel.ErrBackendUnreachable)
}

func TestQueryHonorsTLSSkipVerify(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "")
	}))
	defer srv.Close()
	req := &loggingmodel.QueryReq{Match: []loggingmodel.FieldMatch{{Field: "f", Value: "v"}}, Limit: 1}

	_, err := New(&Config{Endpoint: loggingmodel.Endpoint{URL: srv.URL}}).Query(context.Background(), req)
	assert.ErrorIs(t, err, loggingmodel.ErrBackendUnreachable, "a self-signed cert is refused by default")

	resp, err := New(&Config{Endpoint: loggingmodel.Endpoint{URL: srv.URL, TLSSkipVerify: true}}).
		Query(context.Background(), req)
	assert.NoError(t, err)
	assert.Empty(t, resp.Entries)
	assert.False(t, resp.Truncated)
}

func TestQueryRefusesAnUnscopedRequestWithoutCalling(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { called = true }))
	defer srv.Close()

	_, err := New(&Config{Endpoint: loggingmodel.Endpoint{URL: srv.URL}}).
		Query(context.Background(), &loggingmodel.QueryReq{Limit: 1})
	assert.ErrorIs(t, err, loggingmodel.ErrQueryScopeRequired)
	assert.False(t, called)
}
```

The `Time.Sub` assertion checks both entries parsed: `08:00:02.5 - 08:00:01 = 1.5s`, minus `1s` is `500ms`.

- [ ] **Step 2: Run to verify failure**

Run: `go test ./services/logging/victorialogs/ -run Query`
Expected: FAIL - the current `Query` returns `ErrQueryInvalid` for everything.

- [ ] **Step 3: Implement**

In `client.go`, replace the `Query` stub and change `Ping` to use `c.httpClient(pingTimeout)` instead of `http.DefaultClient` (and drop its own `context.WithTimeout` in favor of the client timeout, or keep both - keep both). Add:

```go
const (
	queryTimeout = 30 * time.Second
	// maxLineBytes bounds one returned line. A log line longer than this is
	// cut by the scanner and reported as an error rather than read unbounded.
	maxLineBytes = 1 << 20
	// errorBodyBytes is how much of an error response is kept for the detail.
	errorBodyBytes = 512
)

// Query runs a search. The request is turned into LogsQL by BuildQuery, never
// passed through.
func (c *Client) Query(ctx context.Context, req *loggingmodel.QueryReq) (*loggingmodel.QueryResp, error) {
	q, err := BuildQuery(req)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if c.cfg.Endpoint.URL == "" {
		return nil, hperrors.Wrap(loggingmodel.ErrBackendUnreachable).WithExtraDetail("no query endpoint")
	}

	form := url.Values{"query": {q}}
	if !req.Start.IsZero() {
		form.Set("start", req.Start.UTC().Format(time.RFC3339Nano))
	}
	if !req.End.IsZero() {
		form.Set("end", req.End.UTC().Format(time.RFC3339Nano))
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimSuffix(c.cfg.Endpoint.URL, "/")+QueryPath, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	applyAuth(httpReq, &c.cfg.Endpoint)

	resp, err := c.httpClient(queryTimeout).Do(httpReq)
	if err != nil {
		return nil, hperrors.Wrap(loggingmodel.ErrBackendUnreachable).WithExtraDetail("%s", err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, errorBodyBytes))
		base := loggingmodel.ErrBackendUnreachable
		if resp.StatusCode == http.StatusBadRequest {
			base = loggingmodel.ErrQueryInvalid
		}
		return nil, hperrors.Wrap(base).WithExtraDetail("%d: %s", resp.StatusCode, string(body))
	}

	entries, err := readEntries(resp.Body)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &loggingmodel.QueryResp{Entries: entries, Truncated: len(entries) == req.Limit}, nil
}

// readEntries parses one JSON object per line, newest first as the query sorts
// them, and returns them oldest first - the order a person reads.
func readEntries(r io.Reader) ([]loggingmodel.LogEntry, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), maxLineBytes) //nolint:mnd // initial buffer
	var out []loggingmodel.LogEntry
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var row map[string]string
		if err := json.Unmarshal(line, &row); err != nil {
			return nil, hperrors.Wrap(loggingmodel.ErrQueryInvalid).WithExtraDetail("unreadable line: %s", err.Error())
		}
		ts, err := time.Parse(time.RFC3339Nano, row["_time"])
		if err != nil {
			return nil, hperrors.Wrap(loggingmodel.ErrQueryInvalid).WithExtraDetail("unreadable _time: %s", err.Error())
		}
		out = append(out, loggingmodel.LogEntry{
			Time:    ts,
			Message: row["_msg"],
			Stream:  row["stream"],
			Level:   row[LevelField],
		})
	}
	if err := sc.Err(); err != nil {
		return nil, hperrors.Wrap(loggingmodel.ErrBackendUnreachable).WithExtraDetail("%s", err.Error())
	}
	slices.Reverse(out)
	return out, nil
}

// httpClient honors the endpoint's TLS setting. http.DefaultClient would
// silently ignore TLSSkipVerify.
func (c *Client) httpClient(timeout time.Duration) *http.Client {
	tr := http.DefaultTransport.(*http.Transport).Clone() //nolint:forcetypeassert // stdlib guarantees it
	if c.cfg.Endpoint.TLSSkipVerify {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // the operator asked for it
	}
	return &http.Client{Transport: tr, Timeout: timeout}
}
```

Imports to add: `bufio`, `crypto/tls`, `encoding/json`, `io`, `net/url`, `slices`.

- [ ] **Step 4: Write the live regression test**

Create `services/logging/victorialogs/query_live_test.go`. It is the spec's section 7 regression suite, run against a real VictoriaLogs when `HP_TEST_VICTORIALOGS_URL` is set and skipped otherwise:

```go
package victorialogs

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

// Run with:
//
//	docker run -d --rm --name vl-test -p 19428:9428 victoriametrics/victoria-logs:v1.52.0
//	HP_TEST_VICTORIALOGS_URL=http://localhost:19428 go test ./services/logging/victorialogs/ -run Live
//	docker rm -f vl-test
func TestLiveQueryCannotLeaveItsScope(t *testing.T) {
	base := os.Getenv("HP_TEST_VICTORIALOGS_URL")
	if base == "" {
		t.Skip("HP_TEST_VICTORIALOGS_URL not set")
	}
	run := time.Now().UTC().Format("150405.000000000")
	self, other := "APP-"+run, "OTHER-"+run
	field := "attrs.hivepaas.app.id"

	lines := strings.Join([]string{
		// A line from self whose JSON forges another app's identity, in both
		// spellings the attack could use.
		`{"_msg":"{\"level\":\"ERROR\",\"attrs.hivepaas.app.id\":\"` + other + `\",\"hivepaas.app.id\":\"` + other +
			`\"}","` + field + `":"` + self + `","stream":"stderr"}`,
		`{"_msg":"plain line from self","` + field + `":"` + self + `","stream":"stdout"}`,
		`{"_msg":"secret of other","` + field + `":"` + other + `","stream":"stdout"}`,
	}, "\n") + "\n"
	ingest, err := http.NewRequest(http.MethodPost, base+"/insert/jsonline", strings.NewReader(lines))
	if err != nil {
		t.Fatal(err)
	}
	// Without this content type the server answers 200 and stores nothing.
	ingest.Header.Set("Content-Type", "application/stream+json")
	resp, err := http.DefaultClient.Do(ingest)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()

	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: base}})
	scope := []loggingmodel.FieldMatch{{Field: field, Value: self}}
	query := func(mod func(r *loggingmodel.QueryReq)) []loggingmodel.LogEntry {
		r := &loggingmodel.QueryReq{Match: scope, Limit: 100}
		if mod != nil {
			mod(r)
		}
		var got *loggingmodel.QueryResp
		for range 20 { // ingestion becomes visible within a second or two
			got, err = c.Query(context.Background(), r)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Entries) > 0 || mod != nil {
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		return got.Entries
	}

	all := query(nil)
	assert.Len(t, all, 2, "both of self's lines, the forged one included")
	for _, e := range all {
		assert.NotContains(t, e.Message, "secret of other")
	}

	for _, hostile := range []string{
		`x" OR "` + field + `":="` + other,
		`") OR ("` + field + `":="` + other,
		"secret of other",
	} {
		got := query(func(r *loggingmodel.QueryReq) { r.Contains = hostile })
		for _, e := range got {
			assert.NotContains(t, e.Message, "secret of other", hostile)
		}
	}

	errs := query(func(r *loggingmodel.QueryReq) { r.Levels = []string{"error"} })
	if assert.Len(t, errs, 1) {
		assert.Equal(t, "ERROR", errs[0].Level, "level comes from the message, matched case-insensitively")
	}

	// The forged field inside self's message must not re-scope the query to
	// other: scoping to other returns only other's own line.
	c2 := &loggingmodel.QueryReq{Match: []loggingmodel.FieldMatch{{Field: field, Value: other}}, Limit: 100}
	otherLines, err := c.Query(context.Background(), c2)
	assert.NoError(t, err)
	if assert.Len(t, otherLines.Entries, 1) {
		assert.Equal(t, "secret of other", otherLines.Entries[0].Message)
	}
}
```

- [ ] **Step 5: Run unit tests, then the live test**

```bash
go test ./services/logging/... && \
docker run -d --rm --name vl-test -p 19428:9428 victoriametrics/victoria-logs:v1.52.0 && \
for i in $(seq 1 20); do curl -sf localhost:19428/health >/dev/null && break; sleep 1; done && \
HP_TEST_VICTORIALOGS_URL=http://localhost:19428 go test ./services/logging/victorialogs/ -run Live -v; \
docker rm -f vl-test
```
Expected: unit tests PASS; `TestLiveQueryCannotLeaveItsScope` PASS (not SKIP). If it fails, the fix is in `BuildQuery` or `readEntries` - never loosen the assertions.

- [ ] **Step 6: Lint and commit**

```bash
golangci-lint run ./services/logging/...
git add services/logging/victorialogs
git commit -m "feat(logging): query VictoriaLogs, with provenance regression tests

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

### Task 5: The service scopes every query to the app

**Files:**
- Modify: `services/logging/vlagent/configure.go`
- Modify: `services/logging/vlagent/configure_test.go`
- Modify: `hivepaas_app/service/loggingservice/service.go`
- Modify: `hivepaas_app/service/loggingservice/errors.go`
- Modify: `hivepaas_app/pkg/translation/messages/en/errors.logging.en.toml`
- Modify: `hivepaas_app/service/loggingservice/loggingserviceimpl/service.go`
- Create: `hivepaas_app/service/loggingservice/loggingserviceimpl/query.go`
- Create: `hivepaas_app/service/loggingservice/loggingserviceimpl/query_test.go`

**Interfaces:**
- Consumes: `backendBaseURL()` (Task 1), `logging.QueryReq`/`FieldMatch`/`QueryResp` (Task 2), `excludedApps` (existing), `appservice.LabelLogAppID`.
- Produces:

```go
// package vlagent
func AttrField(label string) string // "attrs." + label

// package loggingservice
type AppLogQuery struct {
	Contains string
	Levels   []string
	Streams  []string
	Start    time.Time
	End      time.Time
	Limit    int
}
type HistoryUnavailableReason string
const (
	HistoryReasonDisabled         HistoryUnavailableReason = "disabled"
	HistoryReasonAppsNotCollected HistoryUnavailableReason = "apps-not-collected"
	HistoryReasonNoQueryEndpoint  HistoryUnavailableReason = "no-query-endpoint"
	HistoryReasonDriverUnreadable HistoryUnavailableReason = "driver-unreadable"
	HistoryReasonIdentityMissing  HistoryUnavailableReason = "identity-missing"
)
type AppHistory struct {
	Available bool
	Reason    HistoryUnavailableReason
}
// on Service:
QueryAppLogs(ctx context.Context, db database.IDB, app *entity.App, q *AppLogQuery) (*logging.QueryResp, error)
AppHistory(ctx context.Context, db database.IDB, app *entity.App) (*AppHistory, error)
var ErrNotEnabled, ErrQueryEndpointMissing
```

- [ ] **Step 1: `AttrField`, test first**

Append to `services/logging/vlagent/configure_test.go`:

```go
func TestAttrFieldIsWhereADaemonLabelLands(t *testing.T) {
	// Measured: json-file writes the labels option's values into the line's
	// attrs object, and vlagent flattens it with this prefix.
	assert.Equal(t, "attrs.hivepaas.app.id", AttrField("hivepaas.app.id"))
}
```

Append to `configure.go`:

```go
// AttrField is the field a container label named in json-file's `labels`
// option is stored under. The daemon writes the label into the line's attrs
// object, and vlagent flattens that object with this prefix.
func AttrField(label string) string {
	return "attrs." + label
}
```

Run: `go test ./services/logging/vlagent/` - PASS.

- [ ] **Step 2: Interface and errors**

In `hivepaas_app/service/loggingservice/service.go`, add the types from **Interfaces** above (imports `time`, `entity`, `services/logging`) and the two methods on `Service`, each with a doc comment:

```go
	// QueryAppLogs searches one app's stored logs. The app's identity is put
	// into the query here; nothing in q can widen it.
	QueryAppLogs(ctx context.Context, db database.IDB, app *entity.App, q *AppLogQuery) (*logging.QueryResp, error)

	// AppHistory says whether stored logs can be shown for the app, and if not,
	// why - so the dashboard can say so instead of showing an empty list.
	AppHistory(ctx context.Context, db database.IDB, app *entity.App) (*AppHistory, error)
```

In `errors.go`:

```go
	ErrNotEnabled           = hperrors.NewErr(hperrors.ErrUnavailable, "ERR_LOGGING_SVC_NOT_ENABLED")
	ErrQueryEndpointMissing = hperrors.NewErr(hperrors.ErrUnavailable, "ERR_LOGGING_SVC_QUERY_ENDPOINT_MISSING")
```

In the toml:

```toml
ERR_LOGGING_SVC_NOT_ENABLED = "Log collection is not enabled"
ERR_LOGGING_SVC_QUERY_ENDPOINT_MISSING = "The logging backend has no query endpoint configured"
```

- [ ] **Step 3: Write the failing tests**

Create `loggingserviceimpl/query_test.go`:

```go
package loggingserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/services/logging"
)

// fakeBackend records the request and the endpoint it was built for.
type fakeBackend struct {
	logging.Backend
	endpoint logging.Endpoint
	got      *logging.QueryReq
}

func (f *fakeBackend) Query(_ context.Context, req *logging.QueryReq) (*logging.QueryResp, error) {
	f.got = req
	return &logging.QueryResp{}, nil
}

func withFakeBackend(s *service) *fakeBackend {
	fb := &fakeBackend{}
	s.newBackend = func(_ logging.BackendType, cfg *logging.BackendConfig) (logging.Backend, error) {
		fb.endpoint = cfg.VictoriaLogs.Endpoint
		return fb, nil
	}
	return fb
}

func TestQueryAppLogsScopesToTheLoadedApp(t *testing.T) {
	s := newTestService(&fakeDocker{}, storedSetting(t, enabledConfig()))
	fb := withFakeBackend(s)

	_, err := s.QueryAppLogs(context.Background(), nil, &entity.App{ID: "APP1"},
		&loggingservice.AppLogQuery{Contains: "boom", Levels: []string{"error"}, Limit: 50})
	assert.NoError(t, err)

	if assert.NotNil(t, fb.got) {
		assert.Equal(t, []logging.FieldMatch{{Field: "attrs." + appservice.LabelLogAppID, Value: "APP1"}}, fb.got.Match)
		assert.Equal(t, "boom", fb.got.Contains)
		assert.Equal(t, []string{"error"}, fb.got.Levels)
		assert.Equal(t, 50, fb.got.Limit)
	}
	assert.Equal(t, backendBaseURL(), fb.endpoint.URL, "a managed backend is reached by service name")
}

func TestQueryAppLogsRefusesAnAppWithoutAnID(t *testing.T) {
	s := newTestService(&fakeDocker{}, storedSetting(t, enabledConfig()))
	fb := withFakeBackend(s)

	_, err := s.QueryAppLogs(context.Background(), nil, &entity.App{}, &loggingservice.AppLogQuery{Limit: 1})
	assert.Error(t, err)
	assert.Nil(t, fb.got, "an empty id would match every line with no app at all")
}

func TestQueryAppLogsWhenDisabled(t *testing.T) {
	cfg := enabledConfig()
	cfg.Enabled = false
	for _, setting := range []*entity.Setting{nil, storedSetting(t, cfg)} {
		s := newTestService(&fakeDocker{}, setting)
		withFakeBackend(s)
		_, err := s.QueryAppLogs(context.Background(), nil, &entity.App{ID: "APP1"},
			&loggingservice.AppLogQuery{Limit: 1})
		assert.ErrorIs(t, err, loggingservice.ErrNotEnabled)
	}
}

func TestQueryAppLogsUsesTheQueryEndpointOfAnUnmanagedBackend(t *testing.T) {
	cfg := enabledConfig()
	cfg.Backend.Managed = false
	cfg.Backend.Ingest = &entity.LoggingEndpoint{URL: "http://theirs:9428/insert"}
	s := newTestService(&fakeDocker{}, storedSetting(t, cfg))
	withFakeBackend(s)

	_, err := s.QueryAppLogs(context.Background(), nil, &entity.App{ID: "APP1"}, &loggingservice.AppLogQuery{Limit: 1})
	assert.ErrorIs(t, err, loggingservice.ErrQueryEndpointMissing)

	cfg.Backend.Query = &entity.LoggingEndpoint{URL: "http://theirs:9428"}
	s = newTestService(&fakeDocker{}, storedSetting(t, cfg))
	fb := withFakeBackend(s)
	_, err = s.QueryAppLogs(context.Background(), nil, &entity.App{ID: "APP1"}, &loggingservice.AppLogQuery{Limit: 1})
	assert.NoError(t, err)
	assert.Equal(t, "http://theirs:9428", fb.endpoint.URL)
}

func (f *fakeDocker) withService(id string, spec swarm.ServiceSpec) *fakeDocker {
	if f.inspected == nil {
		f.inspected = map[string]swarm.Service{}
	}
	f.inspected[id] = swarm.Service{ID: id, Spec: spec}
	return f
}

func collectibleSpec(appID string) swarm.ServiceSpec {
	return swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{Labels: map[string]string{appservice.LabelLogAppID: appID}},
		LogDriver:     appservice.DefaultLogDriver(),
	}}
}

func TestAppHistory(t *testing.T) {
	disabled := enabledConfig()
	disabled.Enabled = false
	noApps := enabledConfig()
	noApps.Sources = entity.LoggingSources{HivePaaS: true}
	unmanagedNoQuery := enabledConfig()
	unmanagedNoQuery.Backend.Managed = false
	unmanagedNoQuery.Backend.Ingest = &entity.LoggingEndpoint{URL: "http://theirs/insert"}

	local := collectibleSpec("APP1")
	local.TaskTemplate.LogDriver = &swarm.Driver{Name: "local"}
	noIdentity := collectibleSpec("SOURCE-APP") // cloned before the deep-copy fix

	cases := []struct {
		name    string
		setting *entity.Setting
		spec    *swarm.ServiceSpec
		want    loggingservice.AppHistory
	}{
		{"never configured", nil, nil, loggingservice.AppHistory{Reason: loggingservice.HistoryReasonDisabled}},
		{"disabled", storedSetting(t, disabled), nil, loggingservice.AppHistory{Reason: loggingservice.HistoryReasonDisabled}},
		{"apps not collected", storedSetting(t, noApps), nil,
			loggingservice.AppHistory{Reason: loggingservice.HistoryReasonAppsNotCollected}},
		{"no query endpoint", storedSetting(t, unmanagedNoQuery), nil,
			loggingservice.AppHistory{Reason: loggingservice.HistoryReasonNoQueryEndpoint}},
		{"local driver", storedSetting(t, enabledConfig()), &local,
			loggingservice.AppHistory{Reason: loggingservice.HistoryReasonDriverUnreadable}},
		{"wrong identity", storedSetting(t, enabledConfig()), &noIdentity,
			loggingservice.AppHistory{Reason: loggingservice.HistoryReasonIdentityMissing}},
		{"collected", storedSetting(t, enabledConfig()), new(collectibleSpec("APP1")),
			loggingservice.AppHistory{Available: true}},
		// The service is gone: what it logged is exactly what history is for.
		{"service removed", storedSetting(t, enabledConfig()), nil, loggingservice.AppHistory{Available: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fd := &fakeDocker{}
			if tc.spec != nil {
				fd.withService("svc-app", *tc.spec)
			}
			s := newTestService(fd, tc.setting)
			got, err := s.AppHistory(context.Background(), nil, &entity.App{ID: "APP1", ServiceID: "svc-app"})
			assert.NoError(t, err)
			assert.Equal(t, tc.want, *got)
		})
	}
}

```

`new(collectibleSpec("APP1"))` needs Go 1.26 (`new(expr)`); the repo already uses it (`new(true)` in `appuc/logs_get.go`).

The existing `fakeDocker.ServiceInspect` answers by name from `existing`. Extend it so an id present in `f.inspected` returns that service first:

```go
	if svc, ok := f.inspected[serviceID]; ok {
		return &client.ServiceInspectResult{Service: svc}, nil
	}
```

and add `inspected map[string]swarm.Service` to `fakeDocker`.

- [ ] **Step 4: Run to verify failure**

Run: `go test ./hivepaas_app/service/loggingservice/loggingserviceimpl/ -run 'QueryAppLogs|AppHistory'`
Expected: build failure - `s.newBackend undefined`, `s.QueryAppLogs undefined`.

- [ ] **Step 5: Implement**

In `loggingserviceimpl/service.go` add the field and set it in `New`:

```go
	// newBackend builds a backend client; a field so tests can substitute one.
	newBackend func(logging.BackendType, *logging.BackendConfig) (logging.Backend, error)
```
```go
		newBackend:     logging.NewBackend,
```

Create `loggingserviceimpl/query.go`:

```go
package loggingserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/services/logging"
	"github.com/hivepaas/hivepaas/services/logging/victorialogs"
	"github.com/hivepaas/hivepaas/services/logging/vlagent"
)

// QueryAppLogs searches one app's stored logs.
//
// The scope comes from app, which the caller loaded and authorized, and is the
// only Match the backend receives. Nothing in q is a filter on identity.
func (s *service) QueryAppLogs(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	q *loggingservice.AppLogQuery,
) (*logging.QueryResp, error) {
	if app == nil || app.ID == "" {
		// An empty value would match every line that carries no app at all.
		return nil, hperrors.Wrap(logging.ErrQueryScopeRequired)
	}
	cfg, err := s.loadEnabled(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	ep, err := queryEndpoint(cfg)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	backend, err := s.newBackend(logging.BackendType(cfg.Backend.Type),
		&logging.BackendConfig{VictoriaLogs: &victorialogs.Config{Endpoint: ep}})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := backend.Query(ctx, &logging.QueryReq{
		Match:    []logging.FieldMatch{{Field: vlagent.AttrField(appservice.LabelLogAppID), Value: app.ID}},
		Contains: q.Contains,
		Levels:   q.Levels,
		Streams:  q.Streams,
		Start:    q.Start,
		End:      q.End,
		Limit:    q.Limit,
	})
	return resp, hperrors.Wrap(err)
}

// AppHistory says whether stored logs can be shown for the app.
//
// A service that no longer exists is not a reason: its past output is what
// history is for.
func (s *service) AppHistory(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
) (*loggingservice.AppHistory, error) {
	cfg, err := s.loadEnabled(ctx, db)
	if errors.Is(err, loggingservice.ErrNotEnabled) {
		return &loggingservice.AppHistory{Reason: loggingservice.HistoryReasonDisabled}, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !cfg.Sources.Apps {
		return &loggingservice.AppHistory{Reason: loggingservice.HistoryReasonAppsNotCollected}, nil
	}
	if _, err := queryEndpoint(cfg); err != nil {
		return &loggingservice.AppHistory{Reason: loggingservice.HistoryReasonNoQueryEndpoint}, nil
	}

	if app.ServiceID != "" {
		inspect, err := s.dockerManager.ServiceInspect(ctx, app.ServiceID)
		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return nil, hperrors.Wrap(err)
		}
		if err == nil && inspect != nil {
			excluded := excludedApps([]*entity.App{app}, []swarm.Service{inspect.Service})
			if len(excluded) > 0 {
				return &loggingservice.AppHistory{Reason: historyReason(excluded[0].Reason)}, nil
			}
		}
	}
	return &loggingservice.AppHistory{Available: true}, nil
}

func historyReason(r loggingservice.ExcludedReason) loggingservice.HistoryUnavailableReason {
	if r == loggingservice.ExcludedReasonDriverUnreadable {
		return loggingservice.HistoryReasonDriverUnreadable
	}
	return loggingservice.HistoryReasonIdentityMissing
}

// loadEnabled returns the decrypted configuration, or ErrNotEnabled.
func (s *service) loadEnabled(ctx context.Context, db database.IDB) (*entity.Logging, error) {
	setting, err := s.settingRepo.GetSingle(ctx, db, nil, base.SettingTypeLogging, true)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	if setting == nil {
		return nil, hperrors.Wrap(loggingservice.ErrNotEnabled)
	}
	cfg, err := setting.AsLogging()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if cfg == nil || !cfg.Enabled {
		return nil, hperrors.Wrap(loggingservice.ErrNotEnabled)
	}
	if err := cfg.Decrypt(); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return cfg, nil
}

// queryEndpoint is where to read from: the backend HivePaaS runs, reached by
// service name over the API's private network, or the one the operator named.
func queryEndpoint(cfg *entity.Logging) (logging.Endpoint, error) {
	if cfg.Backend.Managed {
		return logging.Endpoint{URL: backendBaseURL()}, nil
	}
	if cfg.Backend.Query == nil || cfg.Backend.Query.URL == "" {
		return logging.Endpoint{}, hperrors.Wrap(loggingservice.ErrQueryEndpointMissing)
	}
	return toEndpoint(cfg.Backend.Query)
}
```

(Import `github.com/moby/moby/api/types/swarm` for `[]swarm.Service`.)

- [ ] **Step 6: Run tests, fx graph, errcodelint, lint**

```bash
go test ./hivepaas_app/service/loggingservice/... ./services/logging/... && \
go test ./hivepaas_app/cmd/internal/ -run TestFxGraphResolves && \
go run ./tools/errcodelint && \
golangci-lint run ./hivepaas_app/service/loggingservice/... ./services/logging/...
```
Expected: PASS, exit 0, `0 issues.` errcodelint will flag `ErrNotEnabled`/`ErrQueryEndpointMissing` only if unreferenced - both are referenced in `query.go`.

- [ ] **Step 7: Commit**

```bash
git add services/logging/vlagent hivepaas_app/service/loggingservice hivepaas_app/pkg/translation/messages/en/errors.logging.en.toml
git commit -m "feat(logging): query an app's stored logs, scoped by the service

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

### Task 6: The app log history API

**Files:**
- Create: `hivepaas_app/usecase/appuc/appdto/logs_history.go`
- Create: `hivepaas_app/usecase/appuc/appdto/logs_history_test.go`
- Create: `hivepaas_app/usecase/appuc/logs_history.go`
- Create: `hivepaas_app/usecase/appuc/logs_history_test.go`
- Modify: `hivepaas_app/usecase/appuc/uc.go` (the file holding `type UC struct` and `New`)
- Modify: `hivepaas_app/usecase/appuc/appdto/logs_info.go`
- Modify: `hivepaas_app/usecase/appuc/logs_info.go`
- Modify: `hivepaas_app/interface/api/handler/apphandler/logs.go`
- Modify: `hivepaas_app/interface/api/server/router_apps.go:157-160`
- Modify: `docs/openapi/swagger.json` (regenerated)

**Interfaces:**
- Consumes: `loggingservice.Service.QueryAppLogs`, `AppHistory`, `AppLogQuery` (Task 5); `tasklog.LogFrame{Type, Data, Ts}`, `tasklog.LogType*`.
- Produces (the dashboard in Task 9 relies on this JSON exactly):
  - `GET /projects/{projectID}/{projectEnv}/apps/{appID}/logs/history?start=&end=&limit=&search=&levels=error,warn&streams=stderr`
  - response `{"meta":..., "data": {"logs": [{"type":"out|err|warn|debug","data":"...","ts":"RFC3339Nano"}], "truncated": bool, "nextEnd": "RFC3339Nano"|absent}}` - `logs` oldest first; `nextEnd` present only when `truncated`, and is what to pass as `end` to get the page before.
  - `GET .../logs/info` gains `"history": {"available": bool, "reason": "disabled|apps-not-collected|no-query-endpoint|driver-unreadable|identity-missing"|absent}`.

- [ ] **Step 1: DTO tests first**

Create `appdto/logs_history_test.go`:

```go
package appdto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func validReq() *GetAppLogHistoryReq {
	return &GetAppLogHistoryReq{
		ProjectID: "01J0000000000000000000000P", ProjectEnvID: "01J000000000000000000000PE",
		AppID: "01J0000000000000000000000A",
	}
}

func TestLogHistoryReqDefaults(t *testing.T) {
	now := time.Date(2026, 9, 12, 10, 0, 0, 0, time.UTC)
	req := validReq()
	req.ApplyDefaults(now)
	assert.Equal(t, now, req.End)
	assert.Equal(t, now.Add(-time.Hour), req.Start)
	assert.Equal(t, DefaultLogHistoryLimit, req.Limit)
}

func TestLogHistoryReqValidate(t *testing.T) {
	now := time.Date(2026, 9, 12, 10, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		mod  func(r *GetAppLogHistoryReq)
		ok   bool
	}{
		{"defaults", func(*GetAppLogHistoryReq) {}, true},
		{"levels", func(r *GetAppLogHistoryReq) { r.Levels = "error, WARN" }, true},
		{"unknown level", func(r *GetAppLogHistoryReq) { r.Levels = "error,verbose" }, false},
		{"streams", func(r *GetAppLogHistoryReq) { r.Streams = "stderr" }, true},
		{"unknown stream", func(r *GetAppLogHistoryReq) { r.Streams = "stdin" }, false},
		{"limit too big", func(r *GetAppLogHistoryReq) { r.Limit = 5001 }, false},
		{"limit negative", func(r *GetAppLogHistoryReq) { r.Limit = -1 }, false},
		{"search too long", func(r *GetAppLogHistoryReq) { r.Search = string(make([]byte, 257)) }, false},
		{"end before start", func(r *GetAppLogHistoryReq) { r.Start = now; r.End = now.Add(-time.Minute) }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := validReq()
			tc.mod(r)
			errs := r.Validate()
			if tc.ok {
				assert.Empty(t, errs)
			} else {
				assert.NotEmpty(t, errs)
			}
		})
	}
}

func TestLogHistoryReqSplitsLists(t *testing.T) {
	r := validReq()
	r.Levels = " error, WARN ,,"
	r.Streams = "stderr"
	assert.Equal(t, []string{"error", "warn"}, r.LevelList())
	assert.Equal(t, []string{"stderr"}, r.StreamList())
}
```

If `validReq`'s IDs fail `basedto.ValidateID`, replace them with IDs copied from an existing appdto test (`grep -rn 'ProjectID:' hivepaas_app/usecase/appuc/appdto/*_test.go`).

- [ ] **Step 2: DTO**

Create `appdto/logs_history.go`:

```go
package appdto

import (
	"strings"
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

const (
	DefaultLogHistoryLimit  = 500
	MaxLogHistoryLimit      = 5000
	maxLogHistorySearchLen  = 256
	defaultLogHistoryWindow = time.Hour
)

var (
	logHistoryLevels  = []string{"trace", "debug", "info", "warn", "warning", "error", "fatal", "panic"}
	logHistoryStreams = []string{"stdout", "stderr"}
)

// GetAppLogHistoryReq is a search over an app's stored logs.
//
// Levels and Streams are comma-separated. There is no field for query text:
// see services/logging/victorialogs/query.go.
type GetAppLogHistoryReq struct {
	ProjectID    string    `json:"-"`
	ProjectEnvID string    `json:"-"`
	AppID        string    `json:"-"`
	Start        time.Time `json:"-" mapstructure:"start"`
	End          time.Time `json:"-" mapstructure:"end"`
	Limit        int       `json:"-" mapstructure:"limit"`
	Search       string    `json:"-" mapstructure:"search"`
	Levels       string    `json:"-" mapstructure:"levels"`
	Streams      string    `json:"-" mapstructure:"streams"`
}

func NewGetAppLogHistoryReq() *GetAppLogHistoryReq {
	return &GetAppLogHistoryReq{}
}

// ApplyDefaults fills what was not asked: the last hour, up to now.
func (req *GetAppLogHistoryReq) ApplyDefaults(now time.Time) {
	if req.End.IsZero() {
		req.End = now
	}
	if req.Start.IsZero() {
		req.Start = req.End.Add(-defaultLogHistoryWindow)
	}
	if req.Limit == 0 {
		req.Limit = DefaultLogHistoryLimit
	}
}

func (req *GetAppLogHistoryReq) LevelList() []string  { return splitList(req.Levels) }
func (req *GetAppLogHistoryReq) StreamList() []string { return splitList(req.Streams) }

func splitList(s string) []string {
	out := []string{}
	for _, p := range strings.Split(s, ",") {
		if p = strings.ToLower(strings.TrimSpace(p)); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func (req *GetAppLogHistoryReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateNumber(&req.Limit, false, 0, MaxLogHistoryLimit, "limit")...)
	validators = append(validators, basedto.ValidateStr(&req.Search, false, 0, maxLogHistorySearchLen, "search")...)
	validators = append(validators, basedto.ValidateSlice(req.LevelList(), false, 0, logHistoryLevels, "levels")...)
	validators = append(validators, basedto.ValidateSlice(req.StreamList(), false, 0, logHistoryStreams, "streams")...)
	validators = append(validators, basedto.ValidateCond(
		req.Start.IsZero() || req.End.IsZero() || req.Start.Before(req.End), "end")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppLogHistoryResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *AppLogHistoryDataResp `json:"data"`
}

type AppLogHistoryDataResp struct {
	// Logs are oldest first.
	Logs      []*tasklog.LogFrame `json:"logs"`
	Truncated bool                `json:"truncated"`
	// NextEnd is the `end` that returns the page before this one. Set only
	// when Truncated.
	NextEnd *time.Time `json:"nextEnd,omitempty"`
}
```

`ValidateNumber(v, required, min, max)` with `min = 0` accepts 0 (defaulted later) and rejects negatives. Check the signatures of `ValidateNumber`, `ValidateSlice` and `ValidateCond` in `hivepaas_app/basedto/validation*.go` - the call shapes above match what is there on 2026-09-12.

In `appdto/logs_info.go`, add to `AppLogsInfoDataResp`:

```go
	// History says whether stored logs can be shown, and if not why.
	History *AppLogHistoryInfoResp `json:"history"`
```
```go
type AppLogHistoryInfoResp struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}
```

Run: `go test ./hivepaas_app/usecase/appuc/appdto/ -run LogHistory` - PASS.

- [ ] **Step 3: Frame mapping, test first**

Create `appuc/logs_history_test.go`:

```go
package appuc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/services/logging"
)

func TestHistoryFrameType(t *testing.T) {
	cases := []struct {
		entry logging.LogEntry
		want  tasklog.LogType
	}{
		{logging.LogEntry{Stream: "stdout"}, tasklog.LogTypeOut},
		{logging.LogEntry{Stream: "stderr"}, tasklog.LogTypeErr},
		{logging.LogEntry{Stream: "stdout", Level: "ERROR"}, tasklog.LogTypeErr},
		{logging.LogEntry{Stream: "stdout", Level: "fatal"}, tasklog.LogTypeErr},
		{logging.LogEntry{Stream: "stderr", Level: "warning"}, tasklog.LogTypeWarn},
		{logging.LogEntry{Stream: "stderr", Level: "debug"}, tasklog.LogTypeDebug},
		{logging.LogEntry{Stream: "stderr", Level: "info"}, tasklog.LogTypeOut},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, historyFrameType(&tc.entry), "%+v", tc.entry)
	}
}

func TestHistoryPage(t *testing.T) {
	t1 := time.Date(2026, 9, 12, 8, 0, 1, 0, time.UTC)
	t2 := t1.Add(time.Second)
	data := toHistoryData(&logging.QueryResp{
		Entries:   []logging.LogEntry{{Time: t1, Message: "a"}, {Time: t2, Message: "b"}},
		Truncated: true,
	})
	if assert.Len(t, data.Logs, 2) {
		assert.Equal(t, "a", data.Logs[0].Data)
		assert.Equal(t, t1, data.Logs[0].Ts)
	}
	if assert.NotNil(t, data.NextEnd) {
		assert.Equal(t, t1.Add(-time.Nanosecond), *data.NextEnd, "strictly before the oldest line shown")
	}

	data = toHistoryData(&logging.QueryResp{Entries: []logging.LogEntry{{Time: t1}}})
	assert.Nil(t, data.NextEnd)
	data = toHistoryData(&logging.QueryResp{})
	assert.NotNil(t, data.Logs, "an empty page is [], not null")
}
```

- [ ] **Step 4: Use case**

Add `loggingService loggingservice.Service` as the last field of `UC` and the last parameter of `New` in the appuc file that declares them, assigning it in the literal.

Create `appuc/logs_history.go`:

```go
package appuc

import (
	"context"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
	"github.com/hivepaas/hivepaas/services/logging"
)

// GetAppLogHistory reads an app's stored logs.
//
// The app is loaded by project and id from the path, which the handler has
// already authorized; that app - not anything in the request - is what the
// query is scoped to.
func (uc *UC) GetAppLogHistory(
	ctx context.Context,
	_ *basedto.Auth,
	req *appdto.GetAppLogHistoryReq,
) (*appdto.GetAppLogHistoryResp, error) {
	app, featureSettings, err := uc.appService.LoadAppWithFeatureSettings(ctx, uc.db, req.ProjectID, req.AppID,
		true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if featureSettings.LoggingSettings != nil && !featureSettings.LoggingSettings.Enabled {
		return nil, hperrors.NewUnavailable("App logs")
	}

	req.ApplyDefaults(timeutil.NowUTC())
	resp, err := uc.loggingService.QueryAppLogs(ctx, uc.db, app, &loggingservice.AppLogQuery{
		Contains: req.Search,
		Levels:   req.LevelList(),
		Streams:  req.StreamList(),
		Start:    req.Start,
		End:      req.End,
		Limit:    req.Limit,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &appdto.GetAppLogHistoryResp{Data: toHistoryData(resp)}, nil
}

func toHistoryData(resp *logging.QueryResp) *appdto.AppLogHistoryDataResp {
	data := &appdto.AppLogHistoryDataResp{
		Logs:      make([]*tasklog.LogFrame, 0, len(resp.Entries)),
		Truncated: resp.Truncated,
	}
	for i := range resp.Entries {
		e := &resp.Entries[i]
		data.Logs = append(data.Logs, &tasklog.LogFrame{Type: historyFrameType(e), Data: e.Message, Ts: e.Time})
	}
	if resp.Truncated && len(resp.Entries) > 0 {
		// One nanosecond before the oldest line shown. VictoriaLogs' end is
		// inclusive, so the oldest line itself would come back again.
		next := resp.Entries[0].Time.Add(-time.Nanosecond)
		data.NextEnd = &next
	}
	return data
}

// historyFrameType colors a stored line the way the live viewer colors one: by
// the message's own level when it has one, by stream otherwise.
func historyFrameType(e *logging.LogEntry) tasklog.LogType {
	switch strings.ToLower(e.Level) {
	case "error", "fatal", "panic":
		return tasklog.LogTypeErr
	case "warn", "warning":
		return tasklog.LogTypeWarn
	case "debug", "trace":
		return tasklog.LogTypeDebug
	case "":
		if e.Stream == "stderr" {
			return tasklog.LogTypeErr
		}
	}
	return tasklog.LogTypeOut
}
```

In `appuc/logs_info.go`, after `resp` is built and before returning it, add:

```go
	history, err := uc.loggingService.AppHistory(ctx, uc.db, app)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp.Data.History = &appdto.AppLogHistoryInfoResp{Available: history.Available, Reason: string(history.Reason)}
```

Put it before the existing `if app.ServiceID == ""` early return if that return would otherwise skip it - no: that early return errors out, which is existing behavior for the live logs; leave it.

- [ ] **Step 5: Handler and route**

Append to `apphandler/logs.go`:

```go
// GetAppLogHistory Gets stored app logs
// @Summary Gets stored app logs
// @Description Searches the logs collected for the app, including those of containers that no longer exist.
// @Description Parameters are structured; no query text is accepted.
// @Tags    apps
// @Produce json
// @Id      getAppLogHistory
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   start query string false "`start=YYYY-MM-DDTHH:mm:SSZ`, default one hour before end"
// @Param   end query string false "`end=YYYY-MM-DDTHH:mm:SSZ`, default now; pass `nextEnd` to page back"
// @Param   limit query int false "lines to return, 1-5000, default 500"
// @Param   search query string false "case-insensitive substring of the message"
// @Param   levels query string false "comma-separated: trace,debug,info,warn,warning,error,fatal,panic"
// @Param   streams query string false "comma-separated: stdout,stderr"
// @Success 200 {object} appdto.GetAppLogHistoryResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/logs/history [get]
func (h *Handler) GetAppLogHistory(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdto.NewGetAppLogHistoryReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appUC.GetAppLogHistory(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}
```

In `router_apps.go`, in the `{ // Logs` block, add:

```go
		appGroup.GET("/:appID/logs/history", appHandler.GetAppLogHistory)
```

- [ ] **Step 6: Run everything**

```bash
go build ./... && \
go test ./hivepaas_app/usecase/appuc/... ./hivepaas_app/service/loggingservice/... && \
go test ./hivepaas_app/cmd/internal/ -run TestFxGraphResolves && \
go run ./tools/errcodelint && \
golangci-lint run ./hivepaas_app/usecase/appuc/... ./hivepaas_app/interface/api/...
```
Expected: PASS, exit 0, `0 issues.` If a test elsewhere constructs `appuc.New(...)` positionally, add `nil` for the new last parameter there.

Check that `start` with nanoseconds parses (the dashboard sends `nextEnd` back verbatim): find the mapstructure decode hook used by `ParseAndValidateRequest` (`grep -rn 'StringToTimeHookFunc\|time.RFC3339' hivepaas_app/interface/api/handler/basehandler`). If it parses with `time.RFC3339`, Go accepts fractional seconds there - no change. If it uses a layout without fractions, parse `end`/`start` with `time.RFC3339Nano` explicitly and note it in the report.

- [ ] **Step 7: Regenerate swagger and commit**

```bash
./tools/swag/swag.sh
git add hivepaas_app/usecase/appuc hivepaas_app/interface/api docs/openapi/swagger.json
git commit -m "feat(logging): app log history API

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

### Task 7: Dashboard - logging settings API layer

All paths below are under `/Users/tnt/go/src/github.com/hivepaas/hivepaas-dashboard/src/application/modules/system-settings/`. Model every file on its `system-ssl-renewal` sibling - same structure, same names with `hivepaas-logging-settings` in place of `system-ssl-renewal`.

**Files:**
- Create: `domain/hivepaas-logging-settings.entity.ts`, export from `domain/index.ts`
- Create: `api/services/hivepaas-logging-settings-services/{hivepaas-logging-settings.api.contracts.ts,hivepaas-logging-settings.api.validator.ts,hivepaas-logging-settings.api.ts,index.ts}`, export from `api/services/index.ts`
- Modify: `api/api-context/system-settings.api.context.ts` - add `hivepaasLoggingSettings: new HivePaaSLoggingSettingsApi(new HivePaaSLoggingSettingsApiValidator())` under `systemSettings`
- Create: `api/hooks/use-hivepaas-logging-settings.api.ts`, export from `api/hooks/index.ts`
- Modify: `data/constants/system-settings.query-keys.ts` - add `"system-settings.hivepaas.logging.find-one": "system-settings.hivepaas.logging.find-one",`
- Create: `data/queries/hivepaas-logging-settings.queries.ts`, `data/commands/hivepaas-logging-settings.commands.ts`, export both

**Interfaces:**
- Consumes: `GET/PUT /system/hivepaas/logging-settings` (plan 2). Request body `{ "data": SettingsData }`; response `{ "data": SettingsData, "secretMasked"?: bool, "status"?: { "backendReady": bool, "excludedApps": [{appId,name,reason,driver?}] } }`. A secret comes back as `"********"`; sending `"********"` back keeps the stored one.
- Produces: `HivePaaSLoggingSettings`, `HivePaaSLoggingStatus`, `LOGGING_MASKED_SECRET`, `HivePaaSLoggingSettingsQueries.useFindOne()`, `HivePaaSLoggingSettingsCommands.useUpdateOne()` taking `{ payload: HivePaaSLoggingSettings }`.

- [ ] **Step 1: Domain**

`domain/hivepaas-logging-settings.entity.ts`:

```ts
/** What the server sends in place of a stored secret. Sending it back keeps the secret. */
export const LOGGING_MASKED_SECRET = "********";

export type HivePaaSLoggingEndpoint = {
    url: string;
    username?: string;
    password?: string;
    bearerToken?: string;
    headers?: Record<string, string>;
    tlsSkipVerify?: boolean;
};

export type HivePaaSLoggingSettings = {
    enabled: boolean;
    sources: { apps: boolean; hivepaas: boolean; traefikAccess: boolean; nodes: boolean };
    collector: { type: string; managed: boolean; image?: string };
    backend: {
        type: string;
        managed: boolean;
        ingest?: HivePaaSLoggingEndpoint | null;
        query?: HivePaaSLoggingEndpoint | null;
        victoriaLogs?: {
            image?: string;
            nodeId: string;
            volumeId: string;
            /** timeutil.Duration text: the server writes days as "30d", and accepts w/d/h/m/s. */
            retention: string;
            maxDiskUsagePercent?: number;
        } | null;
    };
    forwards: { name: string; format?: string; endpoint: HivePaaSLoggingEndpoint }[];
};

export type HivePaaSLoggingExcludedApp = {
    appId: string;
    name: string;
    reason: "driver-unreadable" | "identity-missing" | string;
    driver?: string;
};

export type HivePaaSLoggingStatus = {
    backendReady: boolean;
    excludedApps: HivePaaSLoggingExcludedApp[];
};
```

Retention is a string on the wire: `timeutil.Duration` marshals `720h` as `"30d"` (whole days first, then the remainder, e.g. `"1d12h"`) and parses anything `str2duration` accepts - `w`, `d`, `h`, `m`, `s` units, combined.

- [ ] **Step 2: Contracts, validator, API**

Contracts:

```ts
import type { HivePaaSLoggingSettings, HivePaaSLoggingStatus } from "~/system-settings/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type HivePaaSLoggingSettings_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type HivePaaSLoggingSettings_FindOne_Res = ApiResponseBase<{
    settings: HivePaaSLoggingSettings;
    status: HivePaaSLoggingStatus;
}>;

export type HivePaaSLoggingSettings_UpdateOne_Req = ApiRequestBase<{ payload: HivePaaSLoggingSettings }>;
export type HivePaaSLoggingSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
```

Validator (zod; `.nullish()` everywhere the server may omit):

```ts
import { type AxiosResponse } from "axios";
import { z } from "zod";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    HivePaaSLoggingSettings_FindOne_Res,
    HivePaaSLoggingSettings_UpdateOne_Res,
} from "./hivepaas-logging-settings.api.contracts";

const EndpointSchema = z.object({
    url: z.string().catch(""),
    username: z.string().optional(),
    password: z.string().optional(),
    bearerToken: z.string().optional(),
    headers: z.record(z.string(), z.string()).optional(),
    tlsSkipVerify: z.boolean().optional(),
});

const SettingsSchema = z.object({
    enabled: z.boolean().catch(false),
    sources: z.object({
        apps: z.boolean().catch(false),
        hivepaas: z.boolean().catch(false),
        traefikAccess: z.boolean().catch(false),
        nodes: z.boolean().catch(false),
    }),
    collector: z.object({ type: z.string().catch("vlagent"), managed: z.boolean().catch(true), image: z.string().optional() }),
    backend: z.object({
        type: z.string().catch("victoria-logs"),
        managed: z.boolean().catch(true),
        ingest: EndpointSchema.nullish(),
        query: EndpointSchema.nullish(),
        victoriaLogs: z
            .object({
                image: z.string().optional(),
                nodeId: z.string().catch(""),
                volumeId: z.string().catch(""),
                retention: z.string().catch("30d"),
                maxDiskUsagePercent: z.number().optional(),
            })
            .nullish(),
    }),
    forwards: z
        .array(z.object({ name: z.string(), format: z.string().optional(), endpoint: EndpointSchema }))
        .nullish()
        .transform(v => v ?? []),
});

const StatusSchema = z
    .object({
        backendReady: z.boolean().catch(false),
        excludedApps: z
            .array(
                z.object({
                    appId: z.string(),
                    name: z.string(),
                    reason: z.string(),
                    driver: z.string().optional(),
                }),
            )
            .nullish()
            .transform(v => v ?? []),
    })
    .nullish()
    .transform(v => v ?? { backendReady: false, excludedApps: [] });

const FindOneSchema = z.object({
    data: SettingsSchema,
    status: StatusSchema,
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({ meta: BaseMetaApiSchema.nullish() });

export class HivePaaSLoggingSettingsApiValidator {
    findOne = (response: AxiosResponse): HivePaaSLoggingSettings_FindOne_Res => {
        const { data, status, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return { data: { settings: data, status }, meta };
    };

    updateOne = (response: AxiosResponse): HivePaaSLoggingSettings_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
```

API class: copy `system-ssl-renewal.api.ts`, keep `findOne` and `updateOne`, drop `execute`, and use:

```ts
this.client.v1.get("/system/hivepaas/logging-settings", { signal })
this.client.v1.put("/system/hivepaas/logging-settings", { data: payload }, { signal })
```

Hook, queries and commands: copy the ssl-renewal ones, rename, drop `execute`, use query key `QK["system-settings.hivepaas.logging.find-one"]`, error messages `"Failed to get logging settings"` / `"Failed to update logging settings"`, and access `api.systemSettings.hivepaasLoggingSettings`.

- [ ] **Step 3: Verify**

```bash
cd /Users/tnt/go/src/github.com/hivepaas/hivepaas-dashboard && npx tsc --noEmit -p . && \
npx eslint src/application/modules/system-settings --max-warnings 0 && \
npx prettier --check src/application/modules/system-settings
```
Expected: no output from tsc, eslint clean, prettier "All matched files use Prettier code style!" (run `npx prettier --write` on the new files first if needed).

- [ ] **Step 4: Commit (dashboard repo)**

```bash
git add src/application/modules/system-settings
git commit -m "feat(logging): logging settings API layer

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

### Task 8: Dashboard - logging settings page

Paths under `.../src/application/modules/system-settings/` unless absolute.

**Files:**
- Modify: `src/application/shared/constants/route.constants.ts` - under `systemSettings.hivepaas`, after `security`:
  ```ts
  logging: {
      $pattern: "system/hivepaas/logging",
      $route: "/system/hivepaas/logging/",
  },
  ```
- Modify: `layouts/hivepaas/layout/hivepaas-layout.com.tsx` - add `{ label: "Logging", route: ROUTE.systemSettings.hivepaas.logging.$route, icon: ScrollText }` after Security (`ScrollText` from `lucide-react`).
- Modify: `system-settings.router.tsx` - a child `{ path: "logging", lazy: ... SystemSettingsHivePaaSLoggingRoute }` after `security`.
- Modify: `system-settings.module.ts` - export `SystemSettingsHivePaaSLoggingRoute`.
- Create: `routes/hivepaas/logging/{index.ts, route/index.ts, route/system-settings-hivepaas-logging.route.com.tsx, schemas/hivepaas-logging-settings.schema.ts, form/hivepaas-logging-settings.form-mappers.ts, building-blocks/logging-status-section.com.tsx}`; export from `routes/hivepaas/index.ts`.

**Interfaces:**
- Consumes: Task 7's queries/commands and types; `NodesQueries.useFindManyPaginated` (`~/cluster/data/queries`, rows `NodeDetails` with `id`, `name`, `hostname`); `ClusterVolumesQueries.useFindManyPaginated` (rows `ClusterVolume` with `id`, `name`, `nodeId?`).
- Produces: a page at `/system/hivepaas/logging/`.

What the page must do, section by section (each is one `<section>` with a heading, following the ssl-renewal form's layout classes):

1. **Enabled** - a checkbox. Off hides nothing; the server ignores the rest while off.
2. **Sources** - four checkboxes: Apps, HivePaaS, Traefik access log, Node logs. Under Node logs, a muted note: "Not collected yet: reserved for a later release." and render it disabled (the collector has no reader for it yet - `nodeLogGlob` is `/var/log/syslog`, which vlagent does not mount).
3. **Backend** - a two-option select: "Run VictoriaLogs in the cluster (managed)" / "Use my own endpoint".
   - Managed: Node (select from nodes, label `node.name || node.hostname`), Data volume (select from volumes; show only volumes whose `nodeId` is empty or equals the chosen node, and say so under the field: "Only volumes reachable from the chosen node."), Retention (text such as `30d`, default `30d`; VictoriaLogs keeps at least one day), Max disk usage % (number 1-100, optional).
   - Own endpoint: Ingest URL (required), Query URL (optional, with the note "Without it, stored logs cannot be shown in HivePaaS."), and for each endpoint Username, Password, Bearer token, Skip TLS verification.
4. **Forwards** - a list with add/remove; each row: Name (required, unique), Format (select `native` / `jsonline`, default `jsonline`), URL, Bearer token, Skip TLS verification.
5. **Status** (read-only, from the find-one response) - "Backend running: yes/no", and when `excludedApps` is non-empty a table: App, Reason, Log driver, with reason text:
   - `driver-unreadable` → "Its log driver ({driver || "daemon default"}) cannot be read. Switch it to json-file in the app's container settings."
   - `identity-missing` → "Created before logging existed. Save its container settings once to add its identity."

Secrets: a password/token field initialized from the server shows `********`; leaving it untouched sends `********` back, which keeps the stored value. Do not clear masked values on load.

- [ ] **Step 1: Schema and mappers**

`schemas/hivepaas-logging-settings.schema.ts`:

```ts
import { z } from "zod";

const EndpointForm = z.object({
    url: z.string().trim(),
    username: z.string().trim(),
    password: z.string(),
    bearerToken: z.string(),
    tlsSkipVerify: z.boolean(),
});

export const HivePaaSLoggingSettingsFormSchema = z
    .object({
        enabled: z.boolean(),
        sources: z.object({ apps: z.boolean(), hivepaas: z.boolean(), traefikAccess: z.boolean(), nodes: z.boolean() }),
        backendManaged: z.boolean(),
        nodeId: z.string(),
        volumeId: z.string(),
        retention: z.string().trim().regex(/^(\d+(w|d|h|m|s))+$/, "Use a duration such as 30d or 12h"),
        maxDiskUsagePercent: z.number().int().min(1).max(100).nullable(),
        ingest: EndpointForm,
        query: EndpointForm,
        forwards: z.array(EndpointForm.extend({ name: z.string().trim().min(1, "Required"), format: z.string() })),
    })
    .superRefine((v, ctx) => {
        if (!v.enabled) {
            return;
        }
        if (!Object.values(v.sources).some(Boolean)) {
            ctx.addIssue({ code: "custom", path: ["sources", "apps"], message: "Select at least one source" });
        }
        if (v.backendManaged && !v.nodeId) {
            ctx.addIssue({ code: "custom", path: ["nodeId"], message: "Required" });
        }
        if (v.backendManaged && !v.volumeId) {
            ctx.addIssue({ code: "custom", path: ["volumeId"], message: "Required" });
        }
        if (!v.backendManaged && !v.ingest.url) {
            ctx.addIssue({ code: "custom", path: ["ingest", "url"], message: "Required" });
        }
        const seen = new Set<string>();
        v.forwards.forEach((f, i) => {
            if (seen.has(f.name)) {
                ctx.addIssue({ code: "custom", path: ["forwards", i, "name"], message: "Names must be unique" });
            }
            seen.add(f.name);
            if (!f.url) {
                ctx.addIssue({ code: "custom", path: ["forwards", i, "url"], message: "Required" });
            }
        });
    });

export type HivePaaSLoggingSettingsFormInput = z.input<typeof HivePaaSLoggingSettingsFormSchema>;
export type HivePaaSLoggingSettingsFormOutput = z.output<typeof HivePaaSLoggingSettingsFormSchema>;
```

`form/hivepaas-logging-settings.form-mappers.ts`:

```ts
import type { HivePaaSLoggingEndpoint, HivePaaSLoggingSettings } from "~/system-settings/domain";

import type { HivePaaSLoggingSettingsFormInput, HivePaaSLoggingSettingsFormOutput } from "../schemas";

type EndpointForm = HivePaaSLoggingSettingsFormInput["ingest"];

const emptyEndpoint: EndpointForm = { url: "", username: "", password: "", bearerToken: "", tlsSkipVerify: false };

function toEndpointForm(ep?: HivePaaSLoggingEndpoint | null): EndpointForm {
    return {
        url: ep?.url ?? "",
        username: ep?.username ?? "",
        password: ep?.password ?? "",
        bearerToken: ep?.bearerToken ?? "",
        tlsSkipVerify: ep?.tlsSkipVerify ?? false,
    };
}

function toEndpoint(f: EndpointForm): HivePaaSLoggingEndpoint {
    return {
        url: f.url,
        ...(f.username ? { username: f.username } : {}),
        ...(f.password ? { password: f.password } : {}),
        ...(f.bearerToken ? { bearerToken: f.bearerToken } : {}),
        tlsSkipVerify: f.tlsSkipVerify,
    };
}

export function toFormInput(s?: HivePaaSLoggingSettings): HivePaaSLoggingSettingsFormInput {
    return {
        enabled: s?.enabled ?? false,
        sources: s?.sources ?? { apps: true, hivepaas: false, traefikAccess: false, nodes: false },
        backendManaged: s?.backend.managed ?? true,
        nodeId: s?.backend.victoriaLogs?.nodeId ?? "",
        volumeId: s?.backend.victoriaLogs?.volumeId ?? "",
        retention: s?.backend.victoriaLogs?.retention ?? "30d",
        maxDiskUsagePercent: s?.backend.victoriaLogs?.maxDiskUsagePercent ?? null,
        ingest: s?.backend.ingest ? toEndpointForm(s.backend.ingest) : emptyEndpoint,
        query: s?.backend.query ? toEndpointForm(s.backend.query) : emptyEndpoint,
        forwards: (s?.forwards ?? []).map(f => ({
            ...toEndpointForm(f.endpoint),
            name: f.name,
            format: f.format ?? "jsonline",
        })),
    };
}

export function toPayload(v: HivePaaSLoggingSettingsFormOutput): HivePaaSLoggingSettings {
    return {
        enabled: v.enabled,
        sources: { ...v.sources, nodes: false },
        collector: { type: "vlagent", managed: true },
        backend: v.backendManaged
            ? {
                  type: "victoria-logs",
                  managed: true,
                  victoriaLogs: {
                      nodeId: v.nodeId,
                      volumeId: v.volumeId,
                      retention: v.retention,
                      ...(v.maxDiskUsagePercent ? { maxDiskUsagePercent: v.maxDiskUsagePercent } : {}),
                  },
              }
            : {
                  type: "victoria-logs",
                  managed: false,
                  ingest: toEndpoint(v.ingest),
                  ...(v.query.url ? { query: toEndpoint(v.query) } : {}),
              },
        forwards: v.forwards.map(f => ({ name: f.name, format: f.format, endpoint: toEndpoint(f) })),
    };
}
```

- [ ] **Step 2: Route component**

`route/system-settings-hivepaas-logging.route.com.tsx` follows `system-settings-ssl-renewal-configuration.route.com.tsx`: `useConditionalModule({ id: MODULE_IDS.System })` for `canWrite`; `HivePaaSLoggingSettingsQueries.useFindOne()`; `HivePaaSLoggingSettingsCommands.useUpdateOne({ onSuccess: () => toast.success("Logging settings saved") })`; `AppLoader` while loading; a `react-hook-form` `useForm` with `zodResolver(HivePaaSLoggingSettingsFormSchema)` and `defaultValues: toFormInput(data?.data.settings)`, reset when the query data changes; `useFieldArray` for `forwards`; a `FormActionBar` with the `PermissionTooltipAction`-wrapped Save button copied from ssl-renewal; `onSubmit` → `update({ payload: toPayload(values) })`. On a server validation error, route `ValidationException` errors to `setError` exactly as the ssl-renewal form's `onError` does.

Build the fields from `@/components/ui` (`Checkbox`, `Input`, `InputNumber`, `Select`/`SelectTrigger`/`SelectValue`/`SelectContent`/`SelectItem`, `Label`, `Button`, `Table`). Use the node `Select` from `create-or-edit-volume.form.com.tsx` (lines ~480-495) as the pattern for the node and volume selects: `NodesQueries.useFindManyPaginated({ pagination: { pageIndex: 0, pageSize: 100 } })` and `ClusterVolumesQueries.useFindManyPaginated({ pagination: { pageIndex: 0, pageSize: 100 } })`. Password and token inputs use `InputPassword` from `@/components/ui/input-password`.

A save that enables logging deploys services and can take several seconds; keep the Save button in its loading state until the mutation settles (`isPending`).

`building-blocks/logging-status-section.com.tsx` renders section 5 from `data.data.status`, with the two reason texts above.

- [ ] **Step 3: Verify**

```bash
cd /Users/tnt/go/src/github.com/hivepaas/hivepaas-dashboard && npx tsc --noEmit -p . && \
npx eslint src/application/modules/system-settings src/application/shared/constants --max-warnings 0 && \
npx prettier --check src/application/modules/system-settings src/application/shared/constants/route.constants.ts
```

Then run the app against the local backend (`npm run dev`, backend on its usual dev port) and open `/system/hivepaas/logging/`: the page loads, Save with Enabled off succeeds, and a masked secret round-trips (save twice; the second save must not fail with `ERR_...SECRET...`). Record in the report what was checked by hand and what could not be (for example, if no backend was running).

- [ ] **Step 4: Commit (dashboard repo)**

```bash
git add src/application
git commit -m "feat(logging): logging settings page

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

### Task 9: Dashboard - History tab in app logs

Paths under `/Users/tnt/go/src/github.com/hivepaas/hivepaas-dashboard/src/application/modules/projects/`.

**Files:**
- Modify: `api/services/project-apps-services/logs/app-logs.api.contracts.ts`
- Modify: `api/services/project-apps-services/logs/app-logs.api.validator.ts`
- Modify: `api/services/project-apps-services/logs/app-logs.api.ts`
- Modify: `api/hooks/project-apps/use-app-logs.api.ts`
- Modify: `data/constants/*query-keys*.ts` - add `"projects.apps.logs.$.get-history": "projects.apps.logs.$.get-history",`
- Modify: `data/queries/project-apps/app-logs.queries.ts` - add `useGetHistory` (infinite)
- Create: `routes/single-project/single-app/tabs/logs/building-blocks/app-logs-history/app-logs-history.com.tsx` + `index.ts`, export from `building-blocks/index.ts`
- Modify: `routes/single-project/single-app/tabs/logs/route/app-logs.route.com.tsx`

**Interfaces:**
- Consumes: Task 6's `GET .../logs/history` and `logs/info`'s `history`.
- Produces: a "History" tab after the task tabs.

- [ ] **Step 1: Contracts and validator**

Contracts - add:

```ts
export type AppLogHistoryReason =
    | "disabled"
    | "apps-not-collected"
    | "no-query-endpoint"
    | "driver-unreadable"
    | "identity-missing";

export type AppLogs_GetHistory_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    start?: Date;
    /** Either a Date or a `nextEnd` string passed back verbatim - it has nanoseconds a Date would lose. */
    end?: Date | string;
    limit?: number;
    search?: string;
    levels?: string[];
    streams?: string[];
}>;

export type AppLogs_GetHistory_Res = ApiResponseBase<{
    logs: AppLogFrame[];
    truncated: boolean;
    nextEnd: string | null;
}>;
```

and extend `AppLogs_GetInfo_Res`'s data with `history: { available: boolean; reason: AppLogHistoryReason | null }`.

Validator - in `GetInfoSchema.data` add:

```ts
        history: z
            .object({ available: z.boolean().catch(false), reason: z.string().nullish().transform(v => v ?? null) })
            .nullish()
            .transform(v => v ?? { available: false, reason: "disabled" }),
```

(cast `reason` to `AppLogHistoryReason | null` in the returned object), and a new schema + method:

```ts
const GetHistorySchema = z.object({
    data: z.object({
        logs: z.array(AppLogFrameSchema).nullish().transform(v => v ?? []),
        truncated: z.boolean().catch(false),
        nextEnd: z.string().nullish().transform(v => v ?? null),
    }),
    meta: BaseMetaApiSchema.nullable(),
});

    getHistory = (response: AxiosResponse): AppLogs_GetHistory_Res => {
        return parseApiResponse({ response, schema: GetHistorySchema });
    };
```

API - add:

```ts
    async getHistory(
        request: AppLogs_GetHistory_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppLogs_GetHistory_Res, Error>> {
        const { projectID, env, appID, start, end, limit, search, levels, streams } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/${env}/apps/${appID}/logs/history`, {
                    params: {
                        ...(start ? { start: start.toISOString() } : {}),
                        ...(end ? { end: typeof end === "string" ? end : end.toISOString() } : {}),
                        ...(limit ? { limit } : {}),
                        ...(search ? { search } : {}),
                        ...(levels?.length ? { levels: levels.join(",") } : {}),
                        ...(streams?.length ? { streams: streams.join(",") } : {}),
                    },
                    signal,
                }),
            ).pipe(
                map(this.validator.getHistory),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
```

Hook - add `getHistory` to `queries` in `use-app-logs.api.ts`, same shape as `getLogs`, message `"Failed to get stored logs"`.

Query - add to `app-logs.queries.ts` (import `useInfiniteQuery`):

```ts
type GetHistoryReq = Omit<AppLogs_GetHistory_Req["data"], "end">;

function useGetHistory(request: GetHistoryReq, options: { enabled?: boolean } = {}) {
    const { queries } = useAppLogsApi();

    return useInfiniteQuery({
        queryKey: [QK["projects.apps.logs.$.get-history"], request],
        queryFn: ({ signal, pageParam }) => queries.getHistory({ ...request, end: pageParam }, signal),
        initialPageParam: undefined as string | undefined,
        getNextPageParam: last => (last.data.truncated && last.data.nextEnd ? last.data.nextEnd : undefined),
        ...options,
    });
}
```

and export it in `AppLogsQueries`.

- [ ] **Step 2: History component**

`app-logs-history.com.tsx` renders:

- When `history.available` is false, a paragraph instead of the viewer:
  - `disabled`: "Stored logs are off. An administrator can turn them on in System → HivePaaS → Logging."
  - `apps-not-collected`: "App logs are not collected. An administrator can turn them on in System → HivePaaS → Logging."
  - `no-query-endpoint`: "Logs go to an external backend HivePaaS has no query endpoint for."
  - `driver-unreadable`: "This app's log driver cannot be collected. Switch it to json-file in container settings."
  - `identity-missing`: "This app was created before logging existed. Save its container settings once so its logs can be identified."
- Otherwise the shared `LogsViewer` (`@application/shared/components`) with:
  - `frames`: `pages` reversed (older pages first) and flattened, each `{ type, data, ts }` - the frame shape already matches `LogsViewerFrame`.
  - `toolbarFilters`: a range select (`15m`, `1h` default, `6h`, `24h`, `7d` → sets `start = now - range`, and the request key changes so the query refetches), a search `Input` applied on Enter, a level select (`All` / `Errors` = `["error","fatal","panic"]` / `Warnings and errors` = those plus `["warn","warning"]`), a stream select (`All` / `stdout` / `stderr`).
  - `toolbarStart`: a "Load older" `Button`, disabled unless `hasNextPage`, calling `fetchNextPage()`, showing `isFetchingNextPage` as loading.
  - `onRefresh`: `refetch()`; `isRefreshPending`: `isFetching`.
  - `limit`: 500.
  - `downloadFileName`: `${appID}-history.log`.
  - forward `fontSize`, `themeId`, `height`, `isFullView`, `isFullHeight` from the route like `AppLogsViewer` receives them.
- Only fetch when the tab is active: pass `enabled: isActive && history.available`.

`start` is computed once per range choice (store the Date in state when the range changes), not on every render - otherwise the query key changes every render and refetches forever.

- [ ] **Step 3: The tab**

In `app-logs.route.com.tsx`:
- add `const HISTORY_TAB_ID = "history";`
- append `{ id: HISTORY_TAB_ID, label: "History" }` to `tabs` (after the task tabs)
- in the `tabs.map` that renders `TabsContent`, render `<AppLogsHistory ... />` for `HISTORY_TAB_ID` and `AppLogsViewer` for every other tab; the `tabStates` machinery is not used for History.
- pass `history={infoResponse?.data.history}`.

- [ ] **Step 4: Verify**

```bash
cd /Users/tnt/go/src/github.com/hivepaas/hivepaas-dashboard && npx tsc --noEmit -p . && \
npx eslint src/application/modules/projects --max-warnings 0 && \
npx prettier --check src/application/modules/projects
```

Then by hand, with the backend running: open an app's Logs page; the History tab appears; with logging off it shows the "Stored logs are off" text; the live tabs behave exactly as before. Record what was checked.

- [ ] **Step 5: Commit (dashboard repo)**

```bash
git add src/application/modules/projects
git commit -m "feat(logging): History tab for stored app logs

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

### Task 10: End-to-end on the local swarm, and the spec

This verifies what no unit test can: that the services Apply creates actually talk, and that a real container's line comes back through the API's query path. The local swarm (`docker info` → `Swarm.LocalNodeState: active`) already runs the HivePaaS stack (`hivepaas_*` services, `hivepaas_local_net`, `hivepaas_net`).

**Files:**
- Modify: `docs/superpowers/specs/2026-09-12-logging-design.md` ("What implementation changed")
- Create: `docs/superpowers/notes/2026-09-12-logging-e2e.md` (the record of this run)

- [ ] **Step 1: Deploy the stack the way Apply does**

Write a throwaway Go test file `hivepaas_app/service/loggingservice/loggingserviceimpl/e2e_local_test.go` guarded by `//go:build e2elocal`, which builds a `service` with the real `docker.Manager` (see how `registry/provides.go` constructs it), real `networkservice`/`hpappservice` fakes replaced by: `fakeHpApp{networks: []string{<ID of hivepaas_local_net>}}` and `fakeNetworkService` returning the ID of `hivepaas_net` (get both from `docker network ls --no-trunc`), a `fakeSettingRepo` holding `enabledConfig()` with `NodeID` = `docker node ls -q` and `VolumeID` = a volume created with `docker volume create hp-logging-e2e`, and calls `Apply`. Run it:

```bash
go test -tags e2elocal ./hivepaas_app/service/loggingservice/loggingserviceimpl/ -run E2E -v
```
Expected: `hivepaas-victoria-logs` and `hivepaas-vlagent` running (`docker service ls`), both with `hivepaas_logging_net`, and the backend also on `hivepaas_local_net` (`docker service inspect --format '{{json .Spec.TaskTemplate.Networks}}'`).

- [ ] **Step 2: A real line, collected**

```bash
docker service create -d --name hp-e2e-app \
  --container-label hivepaas.app.id=E2E-APP \
  --log-driver json-file --log-opt labels=hivepaas.app.id \
  alpine sh -c 'while true; do echo "{\"level\":\"error\",\"hivepaas.app.id\":\"FORGED\",\"msg\":\"e2e\"}"; sleep 2; done'
sleep 20
docker run --rm --network hivepaas_local_net curlimages/curl -s \
  http://hivepaas-victoria-logs:9428/select/logsql/query \
  --data-urlencode 'query="attrs.hivepaas.app.id":="E2E-APP" | limit 3'
```
Expected: lines with `"attrs.hivepaas.app.id":"E2E-APP"` and the forged value only inside `_msg`. `hivepaas_local_net` must be attachable for `docker run --network` to work; if it is not, run the curl as a one-shot service on that network instead.

If nothing comes back: `docker service logs hivepaas-vlagent` - a resolution error means Task 1 is wrong; a permission error on `/var/lib/docker/containers` on Docker Desktop is a known limitation of Desktop's VM, and must be recorded, not worked around in code.

- [ ] **Step 3: The API's query path**

Extend the e2e test to call `s.QueryAppLogs(ctx, nil, &entity.App{ID: "E2E-APP"}, &loggingservice.AppLogQuery{Limit: 5})` - but the test process runs on the host, not on `hivepaas_local_net`, so `backendBaseURL()` does not resolve from it. Run this one check from inside the network instead: build the test binary for linux (`GOOS=linux go test -c -tags e2elocal -o /tmp/lg.test ./hivepaas_app/service/loggingservice/loggingserviceimpl/`) and run it in a one-shot container on `hivepaas_local_net` with the docker socket mounted. Expected: entries returned, every one from E2E-APP, level `error`.

- [ ] **Step 4: Clean up**

```bash
docker service rm hp-e2e-app hivepaas-vlagent hivepaas-victoria-logs
sleep 5; docker network rm hivepaas_logging_net; docker volume rm hp-logging-e2e
```
Delete `e2e_local_test.go` - it is not committed. Record the commands and outputs in `docs/superpowers/notes/2026-09-12-logging-e2e.md`.

- [ ] **Step 5: Update the spec**

Add to "What implementation changed":

- **The stack needs its own network.** Swarm services resolve each other only across a shared overlay; plan 2 attached none. Collector and backend share `hivepaas_logging_net`; the backend also joins the API's private networks, never the routing network.
- **The read path takes structured parameters only**, and the scope is put in by the service from the loaded app. `QueryReq` has no query-text field at all; admin raw LogsQL is not offered.
- **Search is a case-insensitive substring**, built as an escaped regular expression. LogsQL's phrase filter matches whole words, which is not what a search box promises.
- **No `Tail` on `Backend`.** Live logs already stream from docker; history is paged with `nextEnd`.
- **Node logs are not collected yet** - the collector does not mount `/var/log`. The toggle is shown disabled.

- [ ] **Step 6: Commit**

```bash
git add docs/superpowers
git commit -m "docs(logging): record the end-to-end run and what the read path changed

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

Final gate before offering to merge: `./scripts/test.sh` ends with `DONE.` and no `FAIL`; `go run ./tools/errcodelint` exits 0; `golangci-lint run ./...` on changed packages prints `0 issues.`; dashboard `npx tsc --noEmit -p .` and `npm run lint` are clean.
