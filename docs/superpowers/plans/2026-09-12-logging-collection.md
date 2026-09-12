# Logging: collection and forwarding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn logging on and have every node's container logs collected, stored in a HivePaaS-deployed VictoriaLogs, and optionally copied to a system HivePaaS does not own.

**Architecture:** `services/logging` describes what to run and how to talk to a backend, knowing nothing about docker or the database; `hivepaas_app/service/loggingservice` turns those descriptions into swarm services and owns the lifecycle. The collector runs one task per node in swarm global mode; the backend is pinned to one node with a volume.

**Tech Stack:** Go, VictoriaLogs `victoriametrics/victoria-logs`, vlagent `victoriametrics/vlagent:v1.52.0`, docker swarm, `github.com/moby/moby/api/types/swarm`.

**Spec:** `docs/superpowers/specs/2026-09-12-logging-design.md`

## Global Constraints

- Run every Go command from the repo root, `/Users/tnt/go/src/github.com/hivepaas/hivepaas`.
- Only `github.com/stretchr/testify/assert` is vendored. **`testify/require` is NOT.** Use `assert`, plus a bare `t.Fatal`/`t.Fatalf` where execution must stop.
- Never add a dependency.
- `services/logging` and every package under it may import **only** `hivepaas_app/hperrors` and `hivepaas_app/pkg/*`. Importing `entity`, `base`, `basedto`, `repository`, `infra/database`, or anything under `services/docker` from those packages is a defect, not a shortcut. `services/backup` is the model.
- Credentials cross into `services/logging` already decrypted, as plain `string`. `loggingmodel` never references `entity.EncryptedField`.
- DI is fx-style: a constructor added to `hivepaas_app/registry/provides.go` is wired automatically. There is no wire, no generated file.
- `entity.Setting.Data` is what persists. `MustSetData` caches the parsed struct in `s.parsedData`, and `parseSettingAs` returns that cache before it ever reads `Data` — so a test that stores and reads back through the same `*Setting` asserts nothing. Build row-shaped settings the way `entity/setting_reencrypt_test.go:18` does: `SetData` into one value, then assert against a **fresh** `&entity.Setting{ID, Type, Data}` carrying no cache.
- `logging.Logger` methods take no `ctx`: `Warnf(format, args...)`.
- American spelling; `misspell` is part of `make lint`.

### Verified facts this plan relies on

These were measured, not assumed. Do not redesign around a contradicting guess.

- Docker's `json-file` driver with `--log-opt labels=a,b` writes an `attrs` object into **every line**. vlagent flattens it to fields prefixed `attrs.`, e.g. `attrs.hivepaas.app.id`.
- vlagent does **not** parse a nested JSON string inside the line's `log` field. An app printing `{"attrs.hivepaas.app.id":"SPOOFED"}` stores that text in `_msg` and the daemon's value survives.
- vlagent accepts repeated `-remoteWrite.url`, with `-remoteWrite.format`, `-remoteWrite.bearerToken`, `-remoteWrite.headers`, `-remoteWrite.basicAuth.*` matched **positionally** to each URL. `-remoteWrite.format=jsonline` ships newline-delimited JSON to a non-VictoriaLogs endpoint.
- vlagent flags confirmed present in v1.52.0: `-fileCollector.glob`, `-fileCollector.excludeGlob`, `-fileCollector.extraFields`, `-fileCollector.streamFields`, `-fileCollector.checkpointsPath`, `-remoteWrite.url`, `-remoteWrite.format`, `-remoteWrite.bearerToken`, `-remoteWrite.headers`, `-remoteWrite.tmpDataPath`.
- VictoriaLogs flags confirmed present: `-storageDataPath`, `-retentionPeriod`, `-retention.maxDiskSpaceUsageBytes`, `-retention.maxDiskUsagePercent`, `-httpListenAddr`. Default HTTP port 9428; health at `/health`; ingest at `/internal/insert`.
- `swarm.ServiceMode.Global` exists in the vendored moby API. **HivePaaS has never used it** — every service it creates today sets `Replicated`.
- `clusterservice.Service` has `ServiceInspect`, `ServiceUpdate`, `ServiceRemove` but **no** `ServiceCreate`. Creating a service goes through `dockerManager.ServiceCreate`, as `usecase/appuc/create.go:69` does.
- A singleton setting is read with `settingRepo.GetSingle(ctx, db, nil, <type>, true)`, as `service/startupservice/startupserviceimpl/data.go:38` does.

---

## File Structure

| File | Responsibility |
|---|---|
| `services/logging/loggingmodel/types.go` | Endpoint, RuntimeSpec, Mount, Port, Resources, CollectSpec, Source, ForwardTarget |
| `services/logging/loggingmodel/service.go` | The `Backend`, `Collector`, `Deployer` interfaces |
| `services/logging/loggingmodel/errors.go` | `ERR_LOGGING_*` |
| `services/logging/victorialogs/deploy.go` | `Deployer`: VictoriaLogs' RuntimeSpec |
| `services/logging/victorialogs/client.go` | `Backend`: HTTP client, `Ping` |
| `services/logging/vlagent/configure.go` | `Collector`: CollectSpec → vlagent flags |
| `services/logging/logging.go` | Re-exported aliases |
| `services/logging/factory.go` | `NewBackend`, `NewCollector` |
| `hivepaas_app/entity/setting_logging.go` | The setting and its parser |
| `hivepaas_app/entity/setting_logging_migration.go` | `Migrate` |
| `hivepaas_app/service/loggingservice/service.go` | The service interface |
| `hivepaas_app/service/loggingservice/errors.go` | `ERR_LOGGING_SVC_*` |
| `hivepaas_app/service/loggingservice/loggingserviceimpl/config.go` | entity → loggingmodel, decrypting |
| `hivepaas_app/service/loggingservice/loggingserviceimpl/swarmspec.go` | RuntimeSpec → swarm.ServiceSpec |
| `hivepaas_app/service/loggingservice/loggingserviceimpl/deploy.go` | Apply / tear down |
| `hivepaas_app/usecase/appsettingsuc/log_driver.go` | Managed log driver for app services |
| `hivepaas_app/usecase/system/logginguc/` | Read and update the setting |

---

### Task 1: loggingmodel — types, interfaces, errors

**Files:**
- Create: `services/logging/loggingmodel/types.go`
- Create: `services/logging/loggingmodel/service.go`
- Create: `services/logging/loggingmodel/errors.go`
- Create: `services/logging/loggingmodel/types_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces: everything below. Later tasks refer to these names exactly.

- [ ] **Step 1: Write the types**

Create `services/logging/loggingmodel/types.go`:

```go
// Package loggingmodel holds the types and interfaces the logging
// implementations share.
//
// It deliberately depends on nothing from the application: no entity, no
// database, no docker. Configuration arrives as plain values - credentials
// included, already decrypted by the caller - so that an implementation can be
// tested without any of that machinery, and so that the application can change
// how it stores things without touching this package.
package loggingmodel

import "time"

// BackendType names a log store HivePaaS knows how to talk to.
type BackendType string

const (
	BackendTypeVictoriaLogs BackendType = "victoria-logs"
)

// CollectorType names a log collector HivePaaS knows how to configure.
type CollectorType string

const (
	CollectorTypeVlagent CollectorType = "vlagent"
)

// Endpoint is how to reach an HTTP log endpoint.
//
// Password and BearerToken are plaintext. The application decrypts before
// calling in, the same way backupmodel.StorageS3 receives a plain secret key.
type Endpoint struct {
	URL           string
	Username      string
	Password      string
	BearerToken   string
	Headers       map[string]string
	TLSSkipVerify bool
}

// SourceKind is what a set of log files holds.
type SourceKind string

const (
	SourceKindApp           SourceKind = "app"
	SourceKindHivePaaS      SourceKind = "hivepaas"
	SourceKindTraefikAccess SourceKind = "traefik-access"
	SourceKindNode          SourceKind = "node"
)

// Source is one set of files to collect.
type Source struct {
	Kind SourceKind
	// Glob matches the files to read, on the node the collector runs on.
	Glob string
	// Exclude drops files Glob would otherwise match. The collector's own
	// container and the backend's belong here: both write logs, and collecting
	// them feeds the collector its own output.
	Exclude string
	// Labels are added to every line from this source.
	Labels map[string]string
}

// ForwardTarget is a write-only copy of the collected stream.
//
// Format is what the target accepts rather than what the collector prefers, so
// it travels with the endpoint.
type ForwardTarget struct {
	Name     string
	Format   string
	Endpoint Endpoint
}

// CollectSpec is the whole job: read these, ship there, copy to those.
type CollectSpec struct {
	Ingest   Endpoint
	Forwards []ForwardTarget
	Sources  []Source
}

// Mount is a host path a container needs.
type Mount struct {
	Source   string
	Target   string
	ReadOnly bool
	// VolumeName is set instead of Source when the mount is a named volume.
	VolumeName string
}

// Port is a container port to expose.
type Port struct {
	Container uint32
	Published uint32
}

// Resources are the limits to run under. Zero means unset.
type Resources struct {
	CPULimit    float64
	MemoryLimit int64
}

// RuntimeSpec describes a container to run, in terms no orchestrator owns.
//
// This is the boundary: implementations describe, and the application decides
// what a swarm service made of that description looks like.
type RuntimeSpec struct {
	Image     string
	Args      []string
	Env       map[string]string
	Files     map[string][]byte
	Mounts    []Mount
	Ports     []Port
	Resources Resources
}

// QueryReq is a search over stored logs.
type QueryReq struct {
	Query string
	Start time.Time
	End   time.Time
	Limit int
}

// LogEntry is one stored line.
type LogEntry struct {
	Time    time.Time
	Message string
	Fields  map[string]string
}

// QueryResp is what a search found.
type QueryResp struct {
	Entries []LogEntry
}
```

- [ ] **Step 2: Write the interfaces**

Create `services/logging/loggingmodel/service.go`:

```go
package loggingmodel

import "context"

// Backend is the read path. It is required whoever owns the backend, because
// reading is the one thing HivePaaS always does itself.
type Backend interface {
	Query(ctx context.Context, req *QueryReq) (*QueryResp, error)
	Ping(ctx context.Context) error
}

// Collector turns a collection job into the configuration that performs it.
type Collector interface {
	Configure(spec *CollectSpec) (*RuntimeSpec, error)
}

// Deployer describes how to run something.
//
// It is separate from Backend because a backend the user runs needs only
// Backend: HivePaaS queries it and never deploys it. Only a managed component
// implements this.
type Deployer interface {
	RuntimeSpec() (*RuntimeSpec, error)
}
```

- [ ] **Step 3: Write the errors**

Create `services/logging/loggingmodel/errors.go`:

```go
package loggingmodel

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// NOTE: these live here rather than in package `logging` because `logging`
// imports the implementations (victorialogs, vlagent), so those packages cannot
// import it back. `logging` re-exports them.
//
// They are also deliberately not in hperrors/constants.go. Their translations
// live in errors.logging.en.toml, which names this file.
var (
	// Configuration
	ErrBackendUnsupported     = hperrors.NewErr(hperrors.ErrUnsupported, "ERR_LOGGING_BACKEND_UNSUPPORTED")
	ErrCollectorUnsupported   = hperrors.NewErr(hperrors.ErrUnsupported, "ERR_LOGGING_COLLECTOR_UNSUPPORTED")
	ErrIngestEndpointRequired = hperrors.NewErr(hperrors.ErrBadRequest, "ERR_LOGGING_INGEST_ENDPOINT_REQUIRED")
	ErrNoSources              = hperrors.NewErr(hperrors.ErrBadRequest, "ERR_LOGGING_NO_SOURCES")
	ErrForwardFormatInvalid   = hperrors.NewErr(hperrors.ErrBadRequest, "ERR_LOGGING_FORWARD_FORMAT_INVALID")

	// Talking to a backend
	ErrQueryInvalid       = hperrors.NewErr(hperrors.ErrBadRequest, "ERR_LOGGING_QUERY_INVALID")
	ErrBackendUnreachable = hperrors.NewErr(hperrors.ErrActionFailed, "ERR_LOGGING_BACKEND_UNREACHABLE")
)
```

- [ ] **Step 4: Add the translations**

Create `hivepaas_app/pkg/translation/messages/en/errors.logging.en.toml`:

```toml
# Logging errors
#
# Not declared in hperrors/constants.go. The sources are:
#   services/logging/loggingmodel/errors.go               (ERR_LOGGING_*)
#   hivepaas_app/service/loggingservice/errors.go         (ERR_LOGGING_SVC_*)
# Adding or removing an error in either file means editing this one.
ERR_LOGGING_BACKEND_UNSUPPORTED = "Logging backend '{{.Name}}' is not supported"
ERR_LOGGING_COLLECTOR_UNSUPPORTED = "Logging collector '{{.Name}}' is not supported"
ERR_LOGGING_INGEST_ENDPOINT_REQUIRED = "An ingest endpoint is required to collect logs"
ERR_LOGGING_NO_SOURCES = "At least one log source must be selected"
ERR_LOGGING_FORWARD_FORMAT_INVALID = "Forward format '{{.Name}}' is not supported"
ERR_LOGGING_QUERY_INVALID = "The log query is not valid"
ERR_LOGGING_BACKEND_UNREACHABLE = "The logging backend cannot be reached"
```

- [ ] **Step 5: Write the test that the boundary holds**

Create `services/logging/loggingmodel/types_test.go`:

```go
package loggingmodel

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// allowedImportPrefixes are the only hivepaas packages anything under
// services/logging may import. The rule is the point of this package: an
// implementation that reaches for entity or docker has moved application
// concerns into a layer that exists to be free of them.
var allowedImportPrefixes = []string{
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors",
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/",
	"github.com/hivepaas/hivepaas/services/logging",
}

func TestLoggingPackagesDoNotImportTheApplication(t *testing.T) {
	root := ".."

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		fset := token.NewFileSet()
		file, perr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if perr != nil {
			return perr
		}
		for _, imp := range file.Imports {
			p, uerr := strconv.Unquote(imp.Path.Value)
			if uerr != nil {
				return uerr
			}
			if !strings.HasPrefix(p, "github.com/hivepaas/hivepaas/") {
				continue // stdlib and third-party are fine
			}
			assert.True(t, allowed(p), "%s imports %s", path, p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}

func allowed(path string) bool {
	for _, prefix := range allowedImportPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
```

- [ ] **Step 6: Run the tests**

Run: `go test ./services/logging/... -count=1 -v`
Expected: PASS.

- [ ] **Step 7: Prove the boundary test works**

Add `_ "github.com/hivepaas/hivepaas/hivepaas_app/entity"` to `types.go`, re-run, and confirm `TestLoggingPackagesDoNotImportTheApplication` fails naming that import. Remove it and confirm the test passes again.

- [ ] **Step 8: Commit**

```bash
git add services/logging hivepaas_app/pkg/translation/messages/en/errors.logging.en.toml
git commit -m "feat(logging): loggingmodel types, interfaces and errors"
```

---

### Task 2: victorialogs — RuntimeSpec and Ping

**Files:**
- Create: `services/logging/victorialogs/deploy.go`
- Create: `services/logging/victorialogs/client.go`
- Create: `services/logging/victorialogs/deploy_test.go`
- Create: `services/logging/victorialogs/client_test.go`

**Interfaces:**
- Consumes: `loggingmodel.RuntimeSpec`, `Mount`, `Port`, `Endpoint`, `Backend`, `Deployer`, and the errors from Task 1.
- Produces:
  - `type Config struct { Image string; DataVolumeName string; Retention time.Duration; MaxDiskUsagePercent int; Endpoint loggingmodel.Endpoint }`
  - `func New(cfg *Config) *Client`
  - `func (c *Client) RuntimeSpec() (*loggingmodel.RuntimeSpec, error)`
  - `func (c *Client) Ping(ctx context.Context) error`
  - `func (c *Client) Query(ctx context.Context, req *loggingmodel.QueryReq) (*loggingmodel.QueryResp, error)`
  - `const DefaultImage, DefaultHTTPPort, DataPath, IngestPath, HealthPath`

Context: `Query` is stubbed here to satisfy `Backend`; the query builder is the next plan's subject. `Ping` is needed now because deployment waits on it.

- [ ] **Step 1: Write the failing test**

Create `services/logging/victorialogs/deploy_test.go`:

```go
package victorialogs

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRuntimeSpecCarriesRetentionAndStorage(t *testing.T) {
	c := New(&Config{DataVolumeName: "vol-123", Retention: 14 * 24 * time.Hour})

	spec, err := c.RuntimeSpec()
	if err != nil {
		t.Fatalf("RuntimeSpec: %v", err)
	}

	args := strings.Join(spec.Args, " ")
	assert.Contains(t, args, "-retentionPeriod=14d")
	assert.Contains(t, args, "-storageDataPath="+DataPath)
	assert.Equal(t, DefaultImage, spec.Image)

	// The data has to outlive the container, so the volume is mounted where
	// -storageDataPath points.
	if len(spec.Mounts) != 1 {
		t.Fatalf("want one mount, got %d", len(spec.Mounts))
	}
	assert.Equal(t, "vol-123", spec.Mounts[0].VolumeName)
	assert.Equal(t, DataPath, spec.Mounts[0].Target)
	assert.False(t, spec.Mounts[0].ReadOnly)
}

// Retention below a day is rejected by VictoriaLogs itself, so rounding down to
// zero days would produce a service that will not start.
func TestRuntimeSpecRoundsRetentionUpToWholeDays(t *testing.T) {
	c := New(&Config{DataVolumeName: "v", Retention: 30 * time.Minute})

	spec, err := c.RuntimeSpec()
	if err != nil {
		t.Fatalf("RuntimeSpec: %v", err)
	}

	assert.Contains(t, strings.Join(spec.Args, " "), "-retentionPeriod=1d")
}

func TestRuntimeSpecOmitsDiskLimitWhenUnset(t *testing.T) {
	c := New(&Config{DataVolumeName: "v", Retention: time.Hour})

	spec, err := c.RuntimeSpec()
	if err != nil {
		t.Fatalf("RuntimeSpec: %v", err)
	}

	assert.NotContains(t, strings.Join(spec.Args, " "), "maxDiskUsagePercent")
}

func TestRuntimeSpecPassesDiskLimit(t *testing.T) {
	c := New(&Config{DataVolumeName: "v", Retention: time.Hour, MaxDiskUsagePercent: 80})

	spec, err := c.RuntimeSpec()
	if err != nil {
		t.Fatalf("RuntimeSpec: %v", err)
	}

	assert.Contains(t, strings.Join(spec.Args, " "), "-retention.maxDiskUsagePercent=80")
}

func TestRuntimeSpecRequiresAVolume(t *testing.T) {
	c := New(&Config{Retention: time.Hour})

	_, err := c.RuntimeSpec()

	assert.Error(t, err)
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./services/logging/victorialogs/ -run TestRuntimeSpec -v`
Expected: FAIL to build, `undefined: New`.

- [ ] **Step 3: Write the implementation**

Create `services/logging/victorialogs/deploy.go`:

```go
// Package victorialogs talks to VictoriaLogs and describes how to run it.
package victorialogs

import (
	"fmt"
	"math"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

const (
	// DefaultImage is pinned rather than tracking latest: an unannounced storage
	// format change arriving on a restart is not a surprise worth having.
	DefaultImage = "victoriametrics/victoria-logs:v1.52.0"

	// DefaultHTTPPort is where VictoriaLogs serves ingest, query and health.
	DefaultHTTPPort = 9428

	// DataPath is where the data volume is mounted inside the container.
	DataPath = "/victoria-logs-data"

	// IngestPath is the native protocol endpoint vlagent writes to.
	IngestPath = "/internal/insert"

	// HealthPath answers OK once the service is serving.
	HealthPath = "/health"

	// QueryPath runs a LogsQL query.
	QueryPath = "/select/logsql/query"

	// minRetentionDays is VictoriaLogs' own floor; it rejects anything shorter.
	minRetentionDays = 1
)

type Config struct {
	Image          string
	DataVolumeName string
	Retention      time.Duration

	// MaxDiskUsagePercent caps the share of the filesystem the logs may take
	// before the oldest days are dropped. Zero leaves it unset.
	MaxDiskUsagePercent int

	// Endpoint is how to reach an already-running instance. It is what the read
	// path uses, and is unset while only RuntimeSpec is needed.
	Endpoint loggingmodel.Endpoint
}

type Client struct {
	cfg *Config
}

func New(cfg *Config) *Client {
	return &Client{cfg: cfg}
}

// RuntimeSpec describes the VictoriaLogs container.
func (c *Client) RuntimeSpec() (*loggingmodel.RuntimeSpec, error) {
	if c.cfg.DataVolumeName == "" {
		return nil, hperrors.Wrap(loggingmodel.ErrIngestEndpointRequired).
			WithExtraDetail("a data volume is required to run VictoriaLogs")
	}

	args := []string{
		fmt.Sprintf("-storageDataPath=%s", DataPath),
		fmt.Sprintf("-retentionPeriod=%dd", retentionDays(c.cfg.Retention)),
		fmt.Sprintf("-httpListenAddr=:%d", DefaultHTTPPort),
	}
	if c.cfg.MaxDiskUsagePercent > 0 {
		args = append(args, fmt.Sprintf("-retention.maxDiskUsagePercent=%d", c.cfg.MaxDiskUsagePercent))
	}

	image := c.cfg.Image
	if image == "" {
		image = DefaultImage
	}

	return &loggingmodel.RuntimeSpec{
		Image: image,
		Args:  args,
		Mounts: []loggingmodel.Mount{{
			VolumeName: c.cfg.DataVolumeName,
			Target:     DataPath,
		}},
		// No published port. VictoriaLogs has no authentication and no tenancy,
		// so it is reachable only on the overlay network the API shares with it.
		Ports: nil,
	}, nil
}

// retentionDays converts a duration to whole days, never below VictoriaLogs'
// own one-day minimum: a shorter setting would otherwise produce a service that
// refuses to start.
func retentionDays(d time.Duration) int {
	days := int(math.Ceil(d.Hours() / 24))
	if days < minRetentionDays {
		return minRetentionDays
	}
	return days
}
```

Create `services/logging/victorialogs/client.go`:

```go
package victorialogs

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

const pingTimeout = 5 * time.Second

// Ping reports whether the backend is serving.
func (c *Client) Ping(ctx context.Context) error {
	if c.cfg.Endpoint.URL == "" {
		return hperrors.Wrap(loggingmodel.ErrIngestEndpointRequired)
	}

	ctx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		strings.TrimSuffix(c.cfg.Endpoint.URL, "/")+HealthPath, nil)
	if err != nil {
		return hperrors.Wrap(err)
	}
	applyAuth(req, &c.cfg.Endpoint)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return hperrors.Wrap(loggingmodel.ErrBackendUnreachable).WithExtraDetail("%s", err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return hperrors.Wrap(loggingmodel.ErrBackendUnreachable).
			WithExtraDetail("health returned %d", resp.StatusCode)
	}
	return nil
}

// Query is the read path. Building LogsQL safely is its own subject and is not
// part of collection; this exists so that Client satisfies loggingmodel.Backend.
func (c *Client) Query(_ context.Context, _ *loggingmodel.QueryReq) (*loggingmodel.QueryResp, error) {
	return nil, hperrors.Wrap(loggingmodel.ErrQueryInvalid).
		WithExtraDetail("the read path is not implemented yet")
}

func applyAuth(req *http.Request, ep *loggingmodel.Endpoint) {
	if ep.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+ep.BearerToken)
	} else if ep.Username != "" {
		req.SetBasicAuth(ep.Username, ep.Password)
	}
	for k, v := range ep.Headers {
		req.Header.Set(k, v)
	}
}

// IngestURL is where a collector should write, derived from the endpoint.
func (c *Client) IngestURL() string {
	return fmt.Sprintf("%s%s", strings.TrimSuffix(c.cfg.Endpoint.URL, "/"), IngestPath)
}
```

- [ ] **Step 4: Write the Ping test**

Create `services/logging/victorialogs/client_test.go`:

```go
package victorialogs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

func TestPingSucceedsOnHealthyBackend(t *testing.T) {
	var gotPath, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotAuth = r.URL.Path, r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: srv.URL, BearerToken: "tok"}})

	assert.NoError(t, c.Ping(context.Background()))
	assert.Equal(t, HealthPath, gotPath)
	assert.Equal(t, "Bearer tok", gotAuth)
}

func TestPingFailsOnUnhealthyBackend(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: srv.URL}})

	assert.Error(t, c.Ping(context.Background()))
}

func TestIngestURLAppendsTheNativePath(t *testing.T) {
	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: "http://vlogs:9428/"}})

	assert.Equal(t, "http://vlogs:9428"+IngestPath, c.IngestURL())
}
```

- [ ] **Step 5: Run every test in the package**

Run: `go test ./services/logging/... -count=1`
Expected: PASS, including the import-boundary test from Task 1.

- [ ] **Step 6: Commit**

```bash
git add services/logging/victorialogs
git commit -m "feat(logging): victorialogs runtime spec and health check"
```

---

### Task 3: vlagent — CollectSpec to flags

**Files:**
- Create: `services/logging/vlagent/configure.go`
- Create: `services/logging/vlagent/configure_test.go`

**Interfaces:**
- Consumes: `loggingmodel.CollectSpec`, `Source`, `ForwardTarget`, `Endpoint`, `RuntimeSpec`, `Mount`, and the errors from Task 1.
- Produces:
  - `type Config struct { Image string }`
  - `func New(cfg *Config) *Collector`
  - `func (c *Collector) Configure(spec *loggingmodel.CollectSpec) (*loggingmodel.RuntimeSpec, error)`
  - `const DefaultImage, ContainersPath, CheckpointsPath, FormatNative, FormatJSONLine`

Context, and the reason this task is the fiddly one: every `-remoteWrite.*` flag is **positional**. The Nth `-remoteWrite.format` belongs to the Nth `-remoteWrite.url`. A destination with no bearer token still needs its slot in the bearer-token array, or the token of a later destination silently attaches to an earlier one — and that means shipping a customer's credential to the wrong system. The primary ingest is always destination zero; forwards follow in order.

- [ ] **Step 1: Write the failing test**

Create `services/logging/vlagent/configure_test.go`:

```go
package vlagent

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

// argsFor returns every value given for one repeated flag, in order.
func argsFor(args []string, name string) []string {
	var out []string
	prefix := name + "="
	for _, a := range args {
		if strings.HasPrefix(a, prefix) {
			out = append(out, strings.TrimPrefix(a, prefix))
		}
	}
	return out
}

func baseSpec() *loggingmodel.CollectSpec {
	return &loggingmodel.CollectSpec{
		Ingest:  loggingmodel.Endpoint{URL: "http://vlogs:9428/internal/insert"},
		Sources: []loggingmodel.Source{{Kind: loggingmodel.SourceKindApp, Glob: "/var/lib/docker/containers/*/*-json.log"}},
	}
}

func TestConfigureShipsToTheIngestEndpoint(t *testing.T) {
	spec, err := New(&Config{}).Configure(baseSpec())
	if err != nil {
		t.Fatalf("Configure: %v", err)
	}

	assert.Equal(t, []string{"http://vlogs:9428/internal/insert"}, argsFor(spec.Args, "-remoteWrite.url"))
	assert.Equal(t, []string{FormatNative}, argsFor(spec.Args, "-remoteWrite.format"))
	assert.Equal(t, DefaultImage, spec.Image)
}

func TestConfigureMountsTheContainerLogsReadOnly(t *testing.T) {
	spec, err := New(&Config{}).Configure(baseSpec())
	if err != nil {
		t.Fatalf("Configure: %v", err)
	}

	var found bool
	for _, m := range spec.Mounts {
		if m.Target == ContainersPath {
			found = true
			assert.True(t, m.ReadOnly, "the collector must never be able to write to the log files")
		}
	}
	assert.True(t, found, "container logs are not mounted")
}

// Every -remoteWrite.* array is matched to its url by position, so a
// destination without a credential still needs its slot. Getting this wrong
// sends one target's token to another.
func TestConfigureKeepsPerDestinationArraysAligned(t *testing.T) {
	in := baseSpec()
	in.Forwards = []loggingmodel.ForwardTarget{
		{Name: "no-auth", Format: FormatJSONLine, Endpoint: loggingmodel.Endpoint{URL: "http://a/ingest"}},
		{Name: "with-token", Format: FormatJSONLine, Endpoint: loggingmodel.Endpoint{URL: "http://b/ingest", BearerToken: "SECRET"}},
	}

	spec, err := New(&Config{}).Configure(in)
	if err != nil {
		t.Fatalf("Configure: %v", err)
	}

	urls := argsFor(spec.Args, "-remoteWrite.url")
	formats := argsFor(spec.Args, "-remoteWrite.format")
	tokens := argsFor(spec.Args, "-remoteWrite.bearerToken")

	assert.Equal(t, []string{"http://vlogs:9428/internal/insert", "http://a/ingest", "http://b/ingest"}, urls)
	assert.Equal(t, []string{FormatNative, FormatJSONLine, FormatJSONLine}, formats)
	// Three slots for three destinations; only the third carries a token.
	assert.Equal(t, []string{"", "", "SECRET"}, tokens)
}

func TestConfigureExcludesTheLoggingStackItself(t *testing.T) {
	in := baseSpec()
	in.Sources[0].Exclude = "/var/lib/docker/containers/deadbeef/*"

	spec, err := New(&Config{}).Configure(in)
	if err != nil {
		t.Fatalf("Configure: %v", err)
	}

	assert.Equal(t, []string{"/var/lib/docker/containers/deadbeef/*"}, argsFor(spec.Args, "-fileCollector.excludeGlob"))
}

// excludeGlob is matched to its glob by position too, so a source with nothing
// to exclude still occupies a slot.
func TestConfigureKeepsGlobAndExcludeAligned(t *testing.T) {
	in := baseSpec()
	in.Sources = []loggingmodel.Source{
		{Kind: loggingmodel.SourceKindApp, Glob: "/a/*.log"},
		{Kind: loggingmodel.SourceKindHivePaaS, Glob: "/b/*.log", Exclude: "/b/skip-*.log"},
	}

	spec, err := New(&Config{}).Configure(in)
	if err != nil {
		t.Fatalf("Configure: %v", err)
	}

	assert.Equal(t, []string{"/a/*.log", "/b/*.log"}, argsFor(spec.Args, "-fileCollector.glob"))
	assert.Equal(t, []string{"", "/b/skip-*.log"}, argsFor(spec.Args, "-fileCollector.excludeGlob"))
}

func TestConfigureRejectsAnEmptySpec(t *testing.T) {
	_, err := New(&Config{}).Configure(&loggingmodel.CollectSpec{
		Ingest: loggingmodel.Endpoint{URL: "http://x"},
	})

	assert.Error(t, err, "collecting nothing is a misconfiguration, not a quiet no-op")
}

func TestConfigureRejectsAMissingIngestURL(t *testing.T) {
	in := baseSpec()
	in.Ingest.URL = ""

	_, err := New(&Config{}).Configure(in)

	assert.Error(t, err)
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./services/logging/vlagent/ -v`
Expected: FAIL to build, `undefined: New`.

- [ ] **Step 3: Write the implementation**

Create `services/logging/vlagent/configure.go`:

```go
// Package vlagent configures VictoriaLogs' collector.
package vlagent

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

const (
	// DefaultImage is pinned. vlagent publishes no `latest` tag in any case.
	DefaultImage = "victoriametrics/vlagent:v1.52.0"

	// ContainersPath is where docker keeps the json-file logs, mounted read-only.
	ContainersPath = "/var/lib/docker/containers"

	// CheckpointsPath is where vlagent records how far it has read. Without it a
	// restart re-sends every line it can still see.
	CheckpointsPath = "/vlagent-data/checkpoints.json"

	// TmpDataPath buffers logs that cannot be delivered yet.
	TmpDataPath = "/vlagent-data/remotewrite"

	// FormatNative is VictoriaLogs' own protocol; FormatJSONLine is
	// newline-delimited JSON, which any HTTP endpoint accepting NDJSON can read.
	FormatNative   = "native"
	FormatJSONLine = "jsonline"
)

type Config struct {
	Image string
}

type Collector struct {
	cfg *Config
}

func New(cfg *Config) *Collector {
	return &Collector{cfg: cfg}
}

// destination is one place logs are written, in the order vlagent's positional
// flags expect.
type destination struct {
	url         string
	format      string
	bearerToken string
	username    string
	password    string
	headers     map[string]string
}

// Configure renders the collection job as vlagent's command line.
//
// Every -remoteWrite.* array is matched to its -remoteWrite.url by position,
// and so is -fileCollector.excludeGlob to -fileCollector.glob. A destination or
// source with nothing to say for one of those flags still takes its slot, with
// an empty value: dropping the slot shifts every later value onto the wrong
// destination, which for a credential means sending it to the wrong system.
func (c *Collector) Configure(spec *loggingmodel.CollectSpec) (*loggingmodel.RuntimeSpec, error) {
	if spec.Ingest.URL == "" {
		return nil, hperrors.Wrap(loggingmodel.ErrIngestEndpointRequired)
	}
	if len(spec.Sources) == 0 {
		return nil, hperrors.Wrap(loggingmodel.ErrNoSources)
	}

	dests := []destination{{
		url:         spec.Ingest.URL,
		format:      FormatNative,
		bearerToken: spec.Ingest.BearerToken,
		username:    spec.Ingest.Username,
		password:    spec.Ingest.Password,
		headers:     spec.Ingest.Headers,
	}}
	for _, f := range spec.Forwards {
		format := f.Format
		if format == "" {
			format = FormatJSONLine
		}
		if format != FormatJSONLine && format != FormatNative {
			return nil, hperrors.Wrap(loggingmodel.ErrForwardFormatInvalid).WithParam("Name", format)
		}
		dests = append(dests, destination{
			url:         f.Endpoint.URL,
			format:      format,
			bearerToken: f.Endpoint.BearerToken,
			username:    f.Endpoint.Username,
			password:    f.Endpoint.Password,
			headers:     f.Endpoint.Headers,
		})
	}

	args := make([]string, 0, len(dests)*5+len(spec.Sources)*3+4)
	for _, d := range dests {
		args = append(args, "-remoteWrite.url="+d.url)
	}
	for _, d := range dests {
		args = append(args, "-remoteWrite.format="+d.format)
	}
	for _, d := range dests {
		args = append(args, "-remoteWrite.bearerToken="+d.bearerToken)
	}
	for _, d := range dests {
		args = append(args, "-remoteWrite.basicAuth.username="+d.username)
	}
	for _, d := range dests {
		args = append(args, "-remoteWrite.basicAuth.password="+d.password)
	}
	for _, d := range dests {
		args = append(args, "-remoteWrite.headers="+joinHeaders(d.headers))
	}

	for _, s := range spec.Sources {
		args = append(args, "-fileCollector.glob="+s.Glob)
	}
	for _, s := range spec.Sources {
		args = append(args, "-fileCollector.excludeGlob="+s.Exclude)
	}
	for _, s := range spec.Sources {
		args = append(args, "-fileCollector.extraFields="+joinFields(s.Labels))
	}

	args = append(args,
		"-fileCollector.checkpointsPath="+CheckpointsPath,
		"-remoteWrite.tmpDataPath="+TmpDataPath,
	)

	image := c.cfg.Image
	if image == "" {
		image = DefaultImage
	}

	return &loggingmodel.RuntimeSpec{
		Image: image,
		Args:  args,
		Mounts: []loggingmodel.Mount{{
			Source:   ContainersPath,
			Target:   ContainersPath,
			ReadOnly: true,
		}},
	}, nil
}

// joinHeaders renders headers the way vlagent parses them, with a stable order
// so that the same configuration always produces the same command line - a
// service spec that differs run to run is a service that redeploys for nothing.
func joinHeaders(h map[string]string) string {
	if len(h) == 0 {
		return ""
	}
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s: %s", k, h[k]))
	}
	return strings.Join(parts, "^^")
}

// joinFields renders extra log fields as key=value pairs, stably ordered for
// the same reason.
func joinFields(f map[string]string) string {
	if len(f) == 0 {
		return ""
	}
	keys := make([]string, 0, len(f))
	for k := range f {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, f[k]))
	}
	return strings.Join(parts, ",")
}
```

- [ ] **Step 4: Run the tests**

Run: `go test ./services/logging/vlagent/ -v`
Expected: PASS, all eight tests.

- [ ] **Step 5: Prove the alignment test bites**

In `Configure`, change the bearer-token loop to skip empty values:

```go
for _, d := range dests {
    if d.bearerToken == "" {
        continue
    }
    args = append(args, "-remoteWrite.bearerToken="+d.bearerToken)
}
```

Run the tests: `TestConfigureKeepsPerDestinationArraysAligned` must fail, because the token would now land on destination zero — the HivePaaS backend — instead of the third. Restore the loop and confirm it passes.

- [ ] **Step 6: Commit**

```bash
git add services/logging/vlagent
git commit -m "feat(logging): vlagent collector configuration"
```

---

### Task 4: The logging facade

**Files:**
- Create: `services/logging/logging.go`
- Create: `services/logging/factory.go`
- Create: `services/logging/factory_test.go`

**Interfaces:**
- Consumes: everything from Tasks 1-3.
- Produces:
  - aliases `logging.Backend`, `logging.Collector`, `logging.Deployer`, `logging.Endpoint`, `logging.RuntimeSpec`, `logging.CollectSpec`, `logging.Source`, `logging.ForwardTarget`, `logging.SourceKind`, `logging.BackendType`, `logging.CollectorType`, `logging.QueryReq`, `logging.QueryResp`, `logging.Mount`, `logging.Port`, `logging.Resources`
  - `func NewBackend(t BackendType, cfg *BackendConfig) (Backend, error)`
  - `func NewCollector(t CollectorType, cfg *CollectorConfig) (Collector, error)`
  - `type BackendConfig struct { VictoriaLogs *victorialogs.Config }`
  - `type CollectorConfig struct { Vlagent *vlagent.Config }`

- [ ] **Step 1: Write the facade**

Create `services/logging/logging.go`:

```go
// Package logging is the entry point to the logging implementations.
//
// Callers use the aliases here rather than reaching into loggingmodel, the same
// arrangement services/backup uses.
package logging

import (
	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

// Re-exported constants
const (
	BackendTypeVictoriaLogs = loggingmodel.BackendTypeVictoriaLogs
	CollectorTypeVlagent    = loggingmodel.CollectorTypeVlagent

	SourceKindApp           = loggingmodel.SourceKindApp
	SourceKindHivePaaS      = loggingmodel.SourceKindHivePaaS
	SourceKindTraefikAccess = loggingmodel.SourceKindTraefikAccess
	SourceKindNode          = loggingmodel.SourceKindNode
)

// Re-exported errors
var (
	ErrBackendUnsupported     = loggingmodel.ErrBackendUnsupported
	ErrCollectorUnsupported   = loggingmodel.ErrCollectorUnsupported
	ErrIngestEndpointRequired = loggingmodel.ErrIngestEndpointRequired
	ErrNoSources              = loggingmodel.ErrNoSources
	ErrForwardFormatInvalid   = loggingmodel.ErrForwardFormatInvalid
	ErrQueryInvalid           = loggingmodel.ErrQueryInvalid
	ErrBackendUnreachable     = loggingmodel.ErrBackendUnreachable
)

// Re-exported types
type (
	BackendType   = loggingmodel.BackendType
	CollectorType = loggingmodel.CollectorType
	SourceKind    = loggingmodel.SourceKind

	Backend  = loggingmodel.Backend
	Collector = loggingmodel.Collector
	Deployer = loggingmodel.Deployer

	Endpoint      = loggingmodel.Endpoint
	Source        = loggingmodel.Source
	ForwardTarget = loggingmodel.ForwardTarget
	CollectSpec   = loggingmodel.CollectSpec
	RuntimeSpec   = loggingmodel.RuntimeSpec
	Mount         = loggingmodel.Mount
	Port          = loggingmodel.Port
	Resources     = loggingmodel.Resources
	QueryReq      = loggingmodel.QueryReq
	QueryResp     = loggingmodel.QueryResp
	LogEntry      = loggingmodel.LogEntry
)
```

Create `services/logging/factory.go`:

```go
package logging

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
	"github.com/hivepaas/hivepaas/services/logging/victorialogs"
	"github.com/hivepaas/hivepaas/services/logging/vlagent"
)

// BackendConfig carries the configuration for whichever backend is named.
type BackendConfig struct {
	VictoriaLogs *victorialogs.Config
}

// CollectorConfig carries the configuration for whichever collector is named.
type CollectorConfig struct {
	Vlagent *vlagent.Config
}

// NewBackend builds the backend for the given type.
func NewBackend(t BackendType, cfg *BackendConfig) (Backend, error) {
	if cfg == nil {
		return nil, hperrors.Wrap(loggingmodel.ErrBackendUnsupported).WithParam("Name", string(t))
	}
	switch t {
	case loggingmodel.BackendTypeVictoriaLogs:
		if cfg.VictoriaLogs == nil {
			return nil, hperrors.Wrap(loggingmodel.ErrBackendUnsupported).WithParam("Name", string(t))
		}
		return victorialogs.New(cfg.VictoriaLogs), nil
	default:
		return nil, hperrors.Wrap(loggingmodel.ErrBackendUnsupported).WithParam("Name", string(t))
	}
}

// NewDeployer builds a backend that HivePaaS also runs.
//
// Separate from NewBackend because only a managed backend can be deployed, and
// asking a user's backend to describe itself is a programming error rather than
// a configuration one.
func NewDeployer(t BackendType, cfg *BackendConfig) (Deployer, error) {
	b, err := NewBackend(t, cfg)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	d, ok := b.(Deployer)
	if !ok {
		return nil, hperrors.Wrap(loggingmodel.ErrBackendUnsupported).WithParam("Name", string(t))
	}
	return d, nil
}

// NewCollector builds the collector for the given type.
func NewCollector(t CollectorType, cfg *CollectorConfig) (Collector, error) {
	if cfg == nil {
		return nil, hperrors.Wrap(loggingmodel.ErrCollectorUnsupported).WithParam("Name", string(t))
	}
	switch t {
	case loggingmodel.CollectorTypeVlagent:
		if cfg.Vlagent == nil {
			return nil, hperrors.Wrap(loggingmodel.ErrCollectorUnsupported).WithParam("Name", string(t))
		}
		return vlagent.New(cfg.Vlagent), nil
	default:
		return nil, hperrors.Wrap(loggingmodel.ErrCollectorUnsupported).WithParam("Name", string(t))
	}
}
```

- [ ] **Step 2: Write the test**

Create `services/logging/factory_test.go`:

```go
package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/logging/victorialogs"
	"github.com/hivepaas/hivepaas/services/logging/vlagent"
)

func TestNewBackendBuildsVictoriaLogs(t *testing.T) {
	b, err := NewBackend(BackendTypeVictoriaLogs, &BackendConfig{VictoriaLogs: &victorialogs.Config{}})

	assert.NoError(t, err)
	assert.NotNil(t, b)
}

func TestNewBackendRejectsAnUnknownType(t *testing.T) {
	_, err := NewBackend("loki", &BackendConfig{VictoriaLogs: &victorialogs.Config{}})

	assert.ErrorIs(t, err, ErrBackendUnsupported)
}

func TestNewDeployerReturnsSomethingThatCanDescribeItself(t *testing.T) {
	d, err := NewDeployer(BackendTypeVictoriaLogs, &BackendConfig{
		VictoriaLogs: &victorialogs.Config{DataVolumeName: "v"},
	})
	if err != nil {
		t.Fatalf("NewDeployer: %v", err)
	}

	spec, err := d.RuntimeSpec()
	assert.NoError(t, err)
	assert.NotEmpty(t, spec.Image)
}

func TestNewCollectorBuildsVlagent(t *testing.T) {
	c, err := NewCollector(CollectorTypeVlagent, &CollectorConfig{Vlagent: &vlagent.Config{}})

	assert.NoError(t, err)
	assert.NotNil(t, c)
}

func TestNewCollectorRejectsAnUnknownType(t *testing.T) {
	_, err := NewCollector("fluent-bit", &CollectorConfig{Vlagent: &vlagent.Config{}})

	assert.ErrorIs(t, err, ErrCollectorUnsupported)
}
```

- [ ] **Step 3: Run everything and lint**

```bash
go test ./services/logging/... -count=1
golangci-lint --timeout=3m run ./services/logging/...
```
Expected: PASS, `0 issues.`

- [ ] **Step 4: Commit**

```bash
git add services/logging
git commit -m "feat(logging): facade and factory"
```

---

### Task 5: The logging setting

**Files:**
- Create: `hivepaas_app/entity/setting_logging.go`
- Create: `hivepaas_app/entity/setting_logging_migration.go`
- Create: `hivepaas_app/entity/setting_logging_test.go`
- Modify: `hivepaas_app/base/setting.go` (add `SettingTypeLogging`)

**Interfaces:**
- Consumes: nothing from earlier tasks — this layer does not know `services/logging` exists.
- Produces: `entity.Logging` and its nested types, `entity.CurrentLoggingVersion`, `base.SettingTypeLogging`, `(*Setting).AsLogging()`, `(*Setting).MustAsLogging()`.

- [ ] **Step 1: Add the setting type**

In `hivepaas_app/base/setting.go`, add to the `SettingType` block, keeping alphabetical order among its neighbours:

```go
	SettingTypeLogging           SettingType = "logging"
```

- [ ] **Step 2: Write the failing test**

Create `hivepaas_app/entity/setting_logging_test.go`:

```go
package entity_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// storedLogging returns a setting shaped like a database row: Data filled in,
// no parsed cache. Reading back through the same *Setting that SetData was
// called on would return the cached struct without ever touching Data, so a
// round-trip through it proves nothing about what persists.
func storedLogging(t *testing.T, data *entity.Logging) *entity.Setting {
	t.Helper()

	s := &entity.Setting{ID: "s1", Type: base.SettingTypeLogging}
	if err := s.SetData(data); err != nil {
		t.Fatalf("SetData: %v", err)
	}
	return &entity.Setting{ID: s.ID, Type: s.Type, Data: s.Data}
}

func TestLoggingSurvivesPersistence(t *testing.T) {
	stored := storedLogging(t, &entity.Logging{
		Enabled: true,
		Sources: entity.LoggingSources{Apps: true, HivePaaS: true},
		Collector: entity.LoggingCollector{
			Type:    entity.LoggingCollectorTypeVlagent,
			Managed: true,
		},
		Backend: entity.LoggingBackend{
			Type:         entity.LoggingBackendTypeVictoriaLogs,
			Managed:      true,
			VictoriaLogs: &entity.LoggingVictoriaLogs{NodeID: "node-1", VolumeID: "vol-1"},
		},
		Forwards: []entity.LoggingForward{{
			Name:     "siem",
			Format:   "jsonline",
			Endpoint: entity.LoggingEndpoint{URL: "https://siem.example/ingest"},
		}},
	})

	got, err := stored.AsLogging()
	if err != nil {
		t.Fatalf("AsLogging: %v", err)
	}

	assert.True(t, got.Enabled)
	assert.True(t, got.Sources.Apps)
	assert.True(t, got.Sources.HivePaaS)
	assert.False(t, got.Sources.Nodes)
	assert.Equal(t, entity.LoggingCollectorTypeVlagent, got.Collector.Type)
	assert.True(t, got.Collector.Managed)
	assert.Equal(t, entity.LoggingBackendTypeVictoriaLogs, got.Backend.Type)
	if got.Backend.VictoriaLogs == nil {
		t.Fatal("VictoriaLogs block did not survive")
	}
	assert.Equal(t, "node-1", got.Backend.VictoriaLogs.NodeID)
	assert.Equal(t, "vol-1", got.Backend.VictoriaLogs.VolumeID)
	if len(got.Forwards) != 1 {
		t.Fatalf("want one forward, got %d", len(got.Forwards))
	}
	assert.Equal(t, "siem", got.Forwards[0].Name)
	assert.Equal(t, "https://siem.example/ingest", got.Forwards[0].Endpoint.URL)
}

// The volume the backend writes to is a real reference: it must be refused if it
// does not exist, and it must not be deletable while logging uses it. This is
// what VerifyingRefIDs acts on, unlike ClusterVolume where it is always empty.
func TestLoggingReferencesItsDataVolume(t *testing.T) {
	l := &entity.Logging{Backend: entity.LoggingBackend{
		VictoriaLogs: &entity.LoggingVictoriaLogs{VolumeID: "vol-9"},
	}}

	assert.Equal(t, []string{"vol-9"}, l.GetRefObjectIDs().RefSettingIDs)
}

func TestLoggingReferencesNothingWithoutAManagedBackend(t *testing.T) {
	l := &entity.Logging{Backend: entity.LoggingBackend{Type: entity.LoggingBackendTypeVictoriaLogs}}

	assert.Empty(t, l.GetRefObjectIDs().RefSettingIDs)
}

func TestLoggingIsDisabledByDefault(t *testing.T) {
	stored := storedLogging(t, &entity.Logging{})

	got, err := stored.AsLogging()
	if err != nil {
		t.Fatalf("AsLogging: %v", err)
	}

	assert.False(t, got.Enabled, "logging must not deploy anything until it is switched on")
}
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `go test ./hivepaas_app/entity/ -run TestLogging -v`
Expected: FAIL to build, `undefined: entity.Logging`.

- [ ] **Step 4: Write the entity**

Create `hivepaas_app/entity/setting_logging.go`:

```go
package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	CurrentLoggingVersion = 1
)

var _ = registerSettingParser(base.SettingTypeLogging, &loggingParser{})

type loggingParser struct {
}

func (s *loggingParser) New() SettingData {
	return &Logging{}
}

// LoggingBackendType names the kind of log store, independently of who runs it.
type LoggingBackendType string

const (
	LoggingBackendTypeVictoriaLogs LoggingBackendType = "victoria-logs"
)

// LoggingCollectorType names the kind of collector, independently of who runs it.
type LoggingCollectorType string

const (
	LoggingCollectorTypeVlagent LoggingCollectorType = "vlagent"
)

// Logging is the whole subsystem's configuration. It is global and, until
// Enabled is set, nothing is deployed.
type Logging struct {
	Enabled   bool             `json:"enabled,omitempty"`
	Sources   LoggingSources   `json:"sources"`
	Collector LoggingCollector `json:"collector"`
	Backend   LoggingBackend   `json:"backend"`

	// Forwards are write-only copies of the stream. Keeping logs for the
	// dashboard and sending a copy to a company's own system is not an
	// either/or, so this is a list beside Backend rather than a variant of it.
	Forwards []LoggingForward `json:"forwards,omitempty"`
}

type LoggingSources struct {
	Apps          bool `json:"apps,omitempty"`
	HivePaaS      bool `json:"hivepaas,omitempty"`
	TraefikAccess bool `json:"traefikAccess,omitempty"`
	Nodes         bool `json:"nodes,omitempty"`
}

// LoggingBackend says what the log store is and whether HivePaaS runs it.
//
// Type and Managed answer different questions on purpose. Making "external" a
// value of Type would make the most useful case inexpressible: a user running
// their own VictoriaLogs, which HivePaaS can still query because it speaks the
// same protocol, but whose lifecycle it does not own.
type LoggingBackend struct {
	Type    LoggingBackendType `json:"type,omitempty"`
	Managed bool               `json:"managed,omitempty"`

	// Ingest and Query are separate because they are separate in backends other
	// than VictoriaLogs, and because a write proxy may sit in front of ingest.
	// Both are derived from the deployed service when Managed, and ignored.
	Ingest *LoggingEndpoint `json:"ingest,omitempty"`
	Query  *LoggingEndpoint `json:"query,omitempty"`

	VictoriaLogs *LoggingVictoriaLogs `json:"victoriaLogs,omitempty"`
}

type LoggingCollector struct {
	Type    LoggingCollectorType     `json:"type,omitempty"`
	Managed bool                     `json:"managed,omitempty"`
	Vlagent *LoggingCollectorVlagent `json:"vlagent,omitempty"`
}

type LoggingCollectorVlagent struct {
	Image string `json:"image,omitempty"`
}

type LoggingVictoriaLogs struct {
	Image    string `json:"image,omitempty"`
	NodeID   string `json:"nodeId,omitempty"`
	VolumeID string `json:"volumeId,omitempty"`

	// Retention reaches VictoriaLogs as a command-line flag, so changing it
	// restarts the service.
	Retention timeutil.Duration `json:"retention"`

	// MaxDiskUsagePercent drops the oldest days once the filesystem is this
	// full. Zero leaves it unset.
	MaxDiskUsagePercent int `json:"maxDiskUsagePercent,omitempty"`
}

type LoggingEndpoint struct {
	URL           string            `json:"url,omitempty"`
	Username      string            `json:"username,omitempty"`
	Password      EncryptedField    `json:"password,omitempty"`
	BearerToken   EncryptedField    `json:"bearerToken,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
	TLSSkipVerify bool              `json:"tlsSkipVerify,omitempty"`
}

type LoggingForward struct {
	Name     string          `json:"name"`
	Format   string          `json:"format,omitempty"`
	Endpoint LoggingEndpoint `json:"endpoint"`
}

func (s *Logging) GetType() base.SettingType {
	return base.SettingTypeLogging
}

// GetRefObjectIDs reports the data volume the managed backend writes to.
//
// Unlike most settings this one genuinely references another, so VerifyingRefIDs
// does something: it refuses a volume id that does not exist, and the resource
// link keeps that volume from being deleted while logging still uses it.
func (s *Logging) GetRefObjectIDs() *RefObjectIDs {
	ids := &RefObjectIDs{}
	if s.Backend.VictoriaLogs != nil && s.Backend.VictoriaLogs.VolumeID != "" {
		ids.RefSettingIDs = append(ids.RefSettingIDs, s.Backend.VictoriaLogs.VolumeID)
	}
	return ids
}

func (s *Logging) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

// Decrypt resolves every stored credential, so that a caller can read them.
func (s *Logging) Decrypt() error {
	endpoints := s.allEndpoints()
	for _, ep := range endpoints {
		if _, err := ep.Password.GetPlain(); err != nil {
			return hperrors.Wrap(err)
		}
		if _, err := ep.BearerToken.GetPlain(); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

// allEndpoints is every place a credential can be stored, so that Decrypt and
// anything like it cannot miss one as the schema grows.
func (s *Logging) allEndpoints() []*LoggingEndpoint {
	eps := make([]*LoggingEndpoint, 0, len(s.Forwards)+2)
	if s.Backend.Ingest != nil {
		eps = append(eps, s.Backend.Ingest)
	}
	if s.Backend.Query != nil {
		eps = append(eps, s.Backend.Query)
	}
	for i := range s.Forwards {
		eps = append(eps, &s.Forwards[i].Endpoint)
	}
	return eps
}

func (s *Setting) AsLogging() (*Logging, error) {
	return parseSettingAs[*Logging](s)
}

func (s *Setting) MustAsLogging() *Logging {
	return gofn.Must(s.AsLogging())
}
```

Create `hivepaas_app/entity/setting_logging_migration.go`:

```go
package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *Logging) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentLoggingVersion {
		return false, nil
	}
	if setting.Version > CurrentLoggingVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// Version 1 is the first, so there is nothing to migrate from yet.

	setting.Version = CurrentLoggingVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
```

- [ ] **Step 5: Run the tests**

Run: `go test ./hivepaas_app/entity/ -run TestLogging -v`
Expected: PASS, all four tests.

- [ ] **Step 6: Prove the persistence test is real**

Change the JSON tag of `Logging.Enabled` from `enabled` to `enabled_x`, re-run, and confirm `TestLoggingSurvivesPersistence` fails. Restore the tag and confirm it passes. A test that reads back through the same `*Setting` would still pass here, which is exactly why it is built row-shaped.

- [ ] **Step 7: Check the re-encryption path covers the new fields**

`entity/setting_reencrypt_test.go` walks settings that hold encrypted fields. Run the whole entity package and read any failure carefully — a new setting with `EncryptedField` that re-encryption does not know about is a data-loss bug at key rotation, not a test problem.

Run: `go test ./hivepaas_app/entity/ -count=1`
Expected: PASS. If re-encryption is table-driven over setting types, add `base.SettingTypeLogging` to that table with a fixture carrying a password and a bearer token in the backend endpoint and in one forward.

- [ ] **Step 8: Commit**

```bash
git add hivepaas_app/entity/setting_logging.go hivepaas_app/entity/setting_logging_migration.go hivepaas_app/entity/setting_logging_test.go hivepaas_app/base/setting.go
git commit -m "feat(logging): the logging setting"
```

---

### Task 6: Audit global-mode swarm services

**Files:**
- Create: `docs/superpowers/notes/2026-09-12-global-mode-audit.md`

**Interfaces:**
- Consumes: nothing.
- Produces: a written answer to "what breaks when a swarm service HivePaaS manages is global rather than replicated", which Tasks 8 and 9 act on.

This task writes findings, not code. The collector must run one task per node, which is `swarm.ServiceMode.Global`, and HivePaaS has never created such a service — everything it makes today sets `Replicated`. Code that assumes a replica count will meet a service that has none.

- [ ] **Step 1: Find every place a replica count is assumed**

```bash
grep -rn "Replicated" --include="*.go" hivepaas_app services | grep -v _test.go | grep -v vendor
grep -rn "Mode\.Replicated\|Spec\.Mode" --include="*.go" hivepaas_app | grep -v _test.go
```

For each hit, record in the note: the file and line, what it reads the replica count for, and what it does when `Spec.Mode.Replicated` is nil.

- [ ] **Step 2: Check the paths that will actually meet the collector**

The collector service is created by `loggingserviceimpl`, not by the app machinery, so most app code never sees it. Confirm which of these do:

```bash
grep -rn "ServiceList\|ServiceInspect" --include="*.go" hivepaas_app/service/clusterservice hivepaas_app/service/appservice | grep -v _test.go
```

Anything that lists *all* services and then reads a replica count is a real risk; anything that inspects one service by an id it already holds is not.

- [ ] **Step 3: Check status reporting**

`hivepaas_app/service/appservice/appserviceimpl/update_status.go:116` and `:151` build `swarm.ReplicatedService` values. Read both and record whether they are reached for a service that is not an app.

- [ ] **Step 4: Write the note**

Create `docs/superpowers/notes/2026-09-12-global-mode-audit.md` with three sections: **Safe** (places that never see the collector, with the reason), **Needs a guard** (places that would read a nil `Replicated`, with file and line), and **Decision** — either "the collector is invisible to all of them, no change needed" or a list of guards Task 9 must add.

- [ ] **Step 5: Commit**

```bash
git add docs/superpowers/notes/2026-09-12-global-mode-audit.md
git commit -m "docs: audit of global-mode swarm services"
```

---

### Task 7: loggingservice — configuration mapping

**Files:**
- Create: `hivepaas_app/service/loggingservice/service.go`
- Create: `hivepaas_app/service/loggingservice/errors.go`
- Create: `hivepaas_app/service/loggingservice/loggingserviceimpl/service.go`
- Create: `hivepaas_app/service/loggingservice/loggingserviceimpl/config.go`
- Create: `hivepaas_app/service/loggingservice/loggingserviceimpl/config_test.go`
- Modify: `hivepaas_app/pkg/translation/messages/en/errors.logging.en.toml`

**Interfaces:**
- Consumes: `entity.Logging` (Task 5); `logging.Endpoint`, `logging.CollectSpec`, `logging.Source`, `logging.ForwardTarget`, `logging.SourceKind*` (Tasks 1-4).
- Produces:
  - `type Service interface { Apply(ctx, db) error; TearDown(ctx) error; Status(ctx) (*Status, error) }`
  - `func New(...) loggingservice.Service`
  - `func toEndpoint(ep *entity.LoggingEndpoint) (logging.Endpoint, error)`
  - `func (s *service) buildCollectSpec(cfg *entity.Logging, ingestURL string) (*logging.CollectSpec, error)`
  - `const ServiceNameBackend = "hivepaas-victoria-logs"`, `ServiceNameCollector = "hivepaas-vlagent"`

Context: this is the boundary the whole structure exists for. `entity.LoggingEndpoint` holds `EncryptedField`; `logging.Endpoint` holds `string`. `toEndpoint` is where the decryption happens, exactly as `buildS3Storage` does in `backupreposerviceimpl/engine.go:177`.

- [ ] **Step 1: Add the service-layer errors**

Create `hivepaas_app/service/loggingservice/errors.go`:

```go
package loggingservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// NOTE: declared here rather than in hperrors/constants.go. The prefix differs
// from loggingmodel's ERR_LOGGING_ deliberately: both files feed one
// translation namespace, and a clash between them is a string collision the
// compiler cannot see. Translations are in errors.logging.en.toml.
var (
	ErrNotConfigured      = hperrors.NewErr(hperrors.ErrBadRequest, "ERR_LOGGING_SVC_NOT_CONFIGURED")
	ErrBackendNodeMissing = hperrors.NewErr(hperrors.ErrBadRequest, "ERR_LOGGING_SVC_BACKEND_NODE_MISSING")
	ErrVolumeMissing      = hperrors.NewErr(hperrors.ErrBadRequest, "ERR_LOGGING_SVC_VOLUME_MISSING")
	ErrBackendNotReady    = hperrors.NewErr(hperrors.ErrActionFailed, "ERR_LOGGING_SVC_BACKEND_NOT_READY")
	ErrDeployFailed       = hperrors.NewErr(hperrors.ErrActionFailed, "ERR_LOGGING_SVC_DEPLOY_FAILED")
)
```

Append to `hivepaas_app/pkg/translation/messages/en/errors.logging.en.toml`:

```toml
ERR_LOGGING_SVC_NOT_CONFIGURED = "Logging is not configured"
ERR_LOGGING_SVC_BACKEND_NODE_MISSING = "The logging backend has no node to run on"
ERR_LOGGING_SVC_VOLUME_MISSING = "The logging backend has no data volume"
ERR_LOGGING_SVC_BACKEND_NOT_READY = "The logging backend did not become ready"
ERR_LOGGING_SVC_DEPLOY_FAILED = "Deploying the logging stack failed"
```

- [ ] **Step 2: Write the failing test**

Create `hivepaas_app/service/loggingservice/loggingserviceimpl/config_test.go`:

```go
package loggingserviceimpl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/logging"
)

// The service layer must hand services/logging plain strings. An EncryptedField
// arriving undecrypted would reach vlagent's command line as ciphertext and the
// forward would fail authentication with nothing saying why.
func TestToEndpointDecrypts(t *testing.T) {
	ep := &entity.LoggingEndpoint{URL: "https://siem.example/ingest", Username: "u"}
	// Set takes plaintext and does not return an error; it detects whether the
	// value is already ciphertext. GetPlain then short-circuits on the
	// plaintext it holds, so this needs no configured data key.
	ep.BearerToken.Set("SECRET")

	got, err := toEndpoint(ep)
	if err != nil {
		t.Fatalf("toEndpoint: %v", err)
	}

	assert.Equal(t, "https://siem.example/ingest", got.URL)
	assert.Equal(t, "u", got.Username)
	assert.Equal(t, "SECRET", got.BearerToken)
}

func TestToEndpointHandlesNil(t *testing.T) {
	got, err := toEndpoint(nil)

	assert.NoError(t, err)
	assert.Empty(t, got.URL)
}

func TestBuildCollectSpecIncludesOnlySelectedSources(t *testing.T) {
	s := &service{}
	cfg := &entity.Logging{Sources: entity.LoggingSources{Apps: true, TraefikAccess: true}}

	spec, err := s.buildCollectSpec(cfg, "http://vlogs:9428/internal/insert")
	if err != nil {
		t.Fatalf("buildCollectSpec: %v", err)
	}

	kinds := map[logging.SourceKind]bool{}
	for _, src := range spec.Sources {
		kinds[src.Kind] = true
	}
	assert.True(t, kinds[logging.SourceKindApp])
	assert.True(t, kinds[logging.SourceKindTraefikAccess])
	assert.False(t, kinds[logging.SourceKindHivePaaS], "HivePaaS logs are opt-in")
	assert.False(t, kinds[logging.SourceKindNode])
}

// The stack writes logs of its own, and collecting them feeds the collector its
// own output. Excluding both is required, not an optimisation.
func TestBuildCollectSpecExcludesTheLoggingStack(t *testing.T) {
	s := &service{}
	cfg := &entity.Logging{Sources: entity.LoggingSources{Apps: true}}

	spec, err := s.buildCollectSpec(cfg, "http://vlogs:9428/internal/insert")
	if err != nil {
		t.Fatalf("buildCollectSpec: %v", err)
	}

	var appSource *logging.Source
	for i := range spec.Sources {
		if spec.Sources[i].Kind == logging.SourceKindApp {
			appSource = &spec.Sources[i]
		}
	}
	if appSource == nil {
		t.Fatal("no app source")
	}
	assert.NotEmpty(t, appSource.Exclude)
	assert.Contains(t, appSource.Exclude, ServiceNameCollector)
}

func TestBuildCollectSpecCarriesForwards(t *testing.T) {
	s := &service{}
	cfg := &entity.Logging{
		Sources: entity.LoggingSources{Apps: true},
		Forwards: []entity.LoggingForward{{
			Name:     "siem",
			Format:   "jsonline",
			Endpoint: entity.LoggingEndpoint{URL: "https://siem.example/ingest"},
		}},
	}

	spec, err := s.buildCollectSpec(cfg, "http://vlogs:9428/internal/insert")
	if err != nil {
		t.Fatalf("buildCollectSpec: %v", err)
	}

	if len(spec.Forwards) != 1 {
		t.Fatalf("want one forward, got %d", len(spec.Forwards))
	}
	assert.Equal(t, "siem", spec.Forwards[0].Name)
	assert.Equal(t, "https://siem.example/ingest", spec.Forwards[0].Endpoint.URL)
	assert.Equal(t, "http://vlogs:9428/internal/insert", spec.Ingest.URL)
}

func TestBuildCollectSpecRefusesWithNoSources(t *testing.T) {
	s := &service{}

	_, err := s.buildCollectSpec(&entity.Logging{}, "http://vlogs:9428/internal/insert")

	assert.Error(t, err)
}

// A source glob must name the docker log directory, or the collector reads
// nothing and reports no error.
func TestAppSourceGlobPointsAtDockerLogs(t *testing.T) {
	s := &service{}

	spec, err := s.buildCollectSpec(&entity.Logging{Sources: entity.LoggingSources{Apps: true}},
		"http://vlogs:9428/internal/insert")
	if err != nil {
		t.Fatalf("buildCollectSpec: %v", err)
	}

	assert.True(t, strings.HasPrefix(spec.Sources[0].Glob, "/var/lib/docker/containers/"))
}
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `go test ./hivepaas_app/service/loggingservice/... -v`
Expected: FAIL to build, `undefined: service`.

- [ ] **Step 4: Write the interface and the mapping**

Create `hivepaas_app/service/loggingservice/service.go`:

```go
package loggingservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// Status is what the logging stack is currently doing.
type Status struct {
	Enabled          bool
	BackendServiceID string
	CollectorService string
	BackendReady     bool
	// ExcludedApps are apps whose own log driver keeps them out of collection.
	ExcludedApps []string
}

type Service interface {
	// Apply makes the cluster match the stored configuration, deploying or
	// removing as needed.
	Apply(ctx context.Context, db database.IDB) error

	// TearDown removes the collector and the backend, keeping the data volume.
	TearDown(ctx context.Context) error

	Status(ctx context.Context, db database.IDB) (*Status, error)
}
```

Create `hivepaas_app/service/loggingservice/loggingserviceimpl/service.go`:

```go
package loggingserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	dockerManager  docker.Manager
	clusterService clusterservice.Service
	settingRepo    repository.SettingRepo
	logger         logging.Logger
}

// New builds the logging service. fx wires the arguments from the provider list
// in registry/provides.go; adding a parameter here needs no other change.
func New(
	dockerManager docker.Manager,
	clusterService clusterservice.Service,
	settingRepo repository.SettingRepo,
	logger logging.Logger,
) loggingservice.Service {
	return &service{
		dockerManager:  dockerManager,
		clusterService: clusterService,
		settingRepo:    settingRepo,
		logger:         logger,
	}
}
```

Create `hivepaas_app/service/loggingservice/loggingserviceimpl/config.go`:

```go
package loggingserviceimpl

import (
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/logging"
)

const (
	// ServiceNameBackend and ServiceNameCollector are the swarm service names.
	// They are also how the collector recognises the stack's own containers in
	// order to skip them.
	ServiceNameBackend   = "hivepaas-victoria-logs"
	ServiceNameCollector = "hivepaas-vlagent"

	// dockerContainersGlob matches every container's json-file log on a node.
	dockerContainersGlob = "/var/lib/docker/containers/*/*-json.log"

	// traefikAccessGlob is where the proxy writes its access log.
	traefikAccessGlob = "/var/log/traefik/access.log"

	// nodeLogGlob is the host's own logs.
	nodeLogGlob = "/var/log/syslog"
)

// toEndpoint converts a stored endpoint into one services/logging can use.
//
// This is where credentials are decrypted, so that nothing under
// services/logging ever sees an EncryptedField - the same boundary
// buildS3Storage draws for backups.
func toEndpoint(ep *entity.LoggingEndpoint) (logging.Endpoint, error) {
	if ep == nil {
		return logging.Endpoint{}, nil
	}

	password, err := ep.Password.GetPlain()
	if err != nil {
		return logging.Endpoint{}, hperrors.Wrap(err)
	}
	token, err := ep.BearerToken.GetPlain()
	if err != nil {
		return logging.Endpoint{}, hperrors.Wrap(err)
	}

	return logging.Endpoint{
		URL:           ep.URL,
		Username:      ep.Username,
		Password:      password,
		BearerToken:   token,
		Headers:       ep.Headers,
		TLSSkipVerify: ep.TLSSkipVerify,
	}, nil
}

// buildCollectSpec turns the stored configuration into a collection job.
func (s *service) buildCollectSpec(cfg *entity.Logging, ingestURL string) (*logging.CollectSpec, error) {
	// The stack's own containers are skipped. Both write logs, and collecting
	// them feeds the collector its own output - with a backend that logs each
	// ingest, that is a loop rather than merely noise.
	stackExclude := fmt.Sprintf("/var/lib/docker/containers/*%s*/*-json.log", ServiceNameCollector)

	var sources []logging.Source
	if cfg.Sources.Apps {
		sources = append(sources, logging.Source{
			Kind:    logging.SourceKindApp,
			Glob:    dockerContainersGlob,
			Exclude: stackExclude,
		})
	}
	if cfg.Sources.HivePaaS {
		sources = append(sources, logging.Source{
			Kind:    logging.SourceKindHivePaaS,
			Glob:    dockerContainersGlob,
			Exclude: stackExclude,
			Labels:  map[string]string{"hivepaas_source": string(logging.SourceKindHivePaaS)},
		})
	}
	if cfg.Sources.TraefikAccess {
		sources = append(sources, logging.Source{
			Kind:   logging.SourceKindTraefikAccess,
			Glob:   traefikAccessGlob,
			Labels: map[string]string{"hivepaas_source": string(logging.SourceKindTraefikAccess)},
		})
	}
	if cfg.Sources.Nodes {
		sources = append(sources, logging.Source{
			Kind:   logging.SourceKindNode,
			Glob:   nodeLogGlob,
			Labels: map[string]string{"hivepaas_source": string(logging.SourceKindNode)},
		})
	}
	if len(sources) == 0 {
		return nil, hperrors.Wrap(logging.ErrNoSources)
	}

	forwards := make([]logging.ForwardTarget, 0, len(cfg.Forwards))
	for i := range cfg.Forwards {
		f := &cfg.Forwards[i]
		ep, err := toEndpoint(&f.Endpoint)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		forwards = append(forwards, logging.ForwardTarget{
			Name:     f.Name,
			Format:   f.Format,
			Endpoint: ep,
		})
	}

	return &logging.CollectSpec{
		Ingest:   logging.Endpoint{URL: ingestURL},
		Forwards: forwards,
		Sources:  sources,
	}, nil
}
```

- [ ] **Step 5: Run the tests**

Run: `go test ./hivepaas_app/service/loggingservice/... -count=1 -v`
Expected: PASS, all seven tests.

- [ ] **Step 6: Commit**

```bash
git add hivepaas_app/service/loggingservice hivepaas_app/pkg/translation/messages/en/errors.logging.en.toml
git commit -m "feat(logging): service interface and configuration mapping"
```

---

### Task 8: RuntimeSpec to swarm.ServiceSpec

**Files:**
- Create: `hivepaas_app/service/loggingservice/loggingserviceimpl/swarmspec.go`
- Create: `hivepaas_app/service/loggingservice/loggingserviceimpl/swarmspec_test.go`

**Interfaces:**
- Consumes: `logging.RuntimeSpec`, `logging.Mount`, `logging.Port`, `logging.Resources`; `ServiceNameBackend`, `ServiceNameCollector` from Task 7; the decision recorded in Task 6.
- Produces:
  - `type swarmSpecOpts struct { Name string; Global bool; NodeID string; Networks []string }`
  - `func toSwarmServiceSpec(rt *logging.RuntimeSpec, opts swarmSpecOpts) (*swarm.ServiceSpec, error)`
  - `const LabelManagedBy = "hivepaas.logging.managed"`

Context: this is where docker knowledge lives and where `services/logging` stops. Two things matter and are easy to get wrong. The collector is `Global` — a mode HivePaaS has never produced, so `Replicated` must be left nil rather than set to something harmless-looking. And the backend is pinned with a required constraint `node.id==X`, the same shape the volume work emits, because its data is on one node's disk.

- [ ] **Step 1: Write the failing test**

Create `hivepaas_app/service/loggingservice/loggingserviceimpl/swarmspec_test.go`:

```go
package loggingserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/logging"
)

func TestToSwarmServiceSpecGlobalModeLeavesReplicatedNil(t *testing.T) {
	rt := &logging.RuntimeSpec{Image: "img", Args: []string{"-a=1"}}

	spec, err := toSwarmServiceSpec(rt, swarmSpecOpts{Name: ServiceNameCollector, Global: true})
	if err != nil {
		t.Fatalf("toSwarmServiceSpec: %v", err)
	}

	assert.NotNil(t, spec.Mode.Global, "the collector must run one task per node")
	assert.Nil(t, spec.Mode.Replicated, "setting both modes is rejected by the daemon")
	assert.Equal(t, ServiceNameCollector, spec.Name)
	assert.Equal(t, "img", spec.TaskTemplate.ContainerSpec.Image)
	assert.Equal(t, []string{"-a=1"}, spec.TaskTemplate.ContainerSpec.Args)
}

func TestToSwarmServiceSpecReplicatedByDefault(t *testing.T) {
	rt := &logging.RuntimeSpec{Image: "img"}

	spec, err := toSwarmServiceSpec(rt, swarmSpecOpts{Name: ServiceNameBackend})
	if err != nil {
		t.Fatalf("toSwarmServiceSpec: %v", err)
	}

	if spec.Mode.Replicated == nil || spec.Mode.Replicated.Replicas == nil {
		t.Fatal("want one replica")
	}
	assert.Equal(t, uint64(1), *spec.Mode.Replicated.Replicas)
	assert.Nil(t, spec.Mode.Global)
}

// The backend's data is on one node's disk, so the constraint is required
// rather than a preference - the same shape the volume pinning emits.
func TestToSwarmServiceSpecPinsToANode(t *testing.T) {
	rt := &logging.RuntimeSpec{Image: "img"}

	spec, err := toSwarmServiceSpec(rt, swarmSpecOpts{Name: ServiceNameBackend, NodeID: "node-7"})
	if err != nil {
		t.Fatalf("toSwarmServiceSpec: %v", err)
	}

	if spec.TaskTemplate.Placement == nil {
		t.Fatal("no placement")
	}
	assert.Contains(t, spec.TaskTemplate.Placement.Constraints, "node.id==node-7")
}

func TestToSwarmServiceSpecTranslatesMounts(t *testing.T) {
	rt := &logging.RuntimeSpec{
		Image: "img",
		Mounts: []logging.Mount{
			{Source: "/var/lib/docker/containers", Target: "/var/lib/docker/containers", ReadOnly: true},
			{VolumeName: "vol-1", Target: "/data"},
		},
	}

	spec, err := toSwarmServiceSpec(rt, swarmSpecOpts{Name: "x"})
	if err != nil {
		t.Fatalf("toSwarmServiceSpec: %v", err)
	}

	mounts := spec.TaskTemplate.ContainerSpec.Mounts
	if len(mounts) != 2 {
		t.Fatalf("want two mounts, got %d", len(mounts))
	}
	assert.Equal(t, mount.TypeBind, mounts[0].Type)
	assert.True(t, mounts[0].ReadOnly)
	assert.Equal(t, mount.TypeVolume, mounts[1].Type)
	assert.Equal(t, "vol-1", mounts[1].Source)
}

// Everything this service creates carries the label, so a later version can
// find what it owns without guessing from names.
func TestToSwarmServiceSpecLabelsWhatItOwns(t *testing.T) {
	spec, err := toSwarmServiceSpec(&logging.RuntimeSpec{Image: "img"}, swarmSpecOpts{Name: "x"})
	if err != nil {
		t.Fatalf("toSwarmServiceSpec: %v", err)
	}

	assert.Equal(t, "true", spec.Labels[LabelManagedBy])
}

func TestToSwarmServiceSpecRequiresAnImage(t *testing.T) {
	_, err := toSwarmServiceSpec(&logging.RuntimeSpec{}, swarmSpecOpts{Name: "x"})

	assert.Error(t, err)
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./hivepaas_app/service/loggingservice/loggingserviceimpl/ -run TestToSwarmServiceSpec -v`
Expected: FAIL to build, `undefined: toSwarmServiceSpec`.

- [ ] **Step 3: Write the implementation**

Create `hivepaas_app/service/loggingservice/loggingserviceimpl/swarmspec.go`:

```go
package loggingserviceimpl

import (
	"sort"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/services/logging"
)

// LabelManagedBy marks every service this package creates, so that finding them
// later does not depend on their names staying the same.
const LabelManagedBy = "hivepaas.logging.managed"

type swarmSpecOpts struct {
	Name string
	// Global runs one task per node. The collector needs this; nothing else
	// HivePaaS creates uses it.
	Global bool
	// NodeID pins the service to one node, for a component whose data is on
	// that node's disk.
	NodeID   string
	Networks []string
}

// toSwarmServiceSpec turns a description of a container into a swarm service.
//
// This is the only place in the logging subsystem that knows what swarm is.
func toSwarmServiceSpec(rt *logging.RuntimeSpec, opts swarmSpecOpts) (*swarm.ServiceSpec, error) {
	if rt.Image == "" {
		return nil, hperrors.Wrap(loggingservice.ErrDeployFailed).
			WithExtraDetail("the runtime spec has no image")
	}

	container := &swarm.ContainerSpec{
		Image: rt.Image,
		Args:  rt.Args,
		Env:   envSlice(rt.Env),
	}
	for _, m := range rt.Mounts {
		container.Mounts = append(container.Mounts, toSwarmMount(m))
	}

	spec := &swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name:   opts.Name,
			Labels: map[string]string{LabelManagedBy: "true"},
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: container,
			// The stack's own logs are collected like anything else's, so they
			// go through the same json-file driver rather than a special one.
			RestartPolicy: &swarm.RestartPolicy{Condition: swarm.RestartPolicyConditionAny},
		},
	}

	// Exactly one mode is set. The daemon rejects a spec carrying both, and a
	// Replicated left filled in "just in case" is how that happens.
	if opts.Global {
		spec.Mode.Global = &swarm.GlobalService{}
	} else {
		replicas := uint64(1)
		spec.Mode.Replicated = &swarm.ReplicatedService{Replicas: &replicas}
	}

	if opts.NodeID != "" {
		spec.TaskTemplate.Placement = &swarm.Placement{
			Constraints: []string{"node.id==" + opts.NodeID},
		}
	}

	for _, n := range opts.Networks {
		spec.TaskTemplate.Networks = append(spec.TaskTemplate.Networks,
			swarm.NetworkAttachmentConfig{Target: n})
	}

	return spec, nil
}

func toSwarmMount(m logging.Mount) mount.Mount {
	if m.VolumeName != "" {
		return mount.Mount{
			Type:     mount.TypeVolume,
			Source:   m.VolumeName,
			Target:   m.Target,
			ReadOnly: m.ReadOnly,
		}
	}
	return mount.Mount{
		Type:     mount.TypeBind,
		Source:   m.Source,
		Target:   m.Target,
		ReadOnly: m.ReadOnly,
	}
}

// envSlice renders the environment in a stable order, so that the same
// configuration always produces the same spec. An unstable one would make every
// Apply look like a change and redeploy the service for nothing.
func envSlice(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, k+"="+env[k])
	}
	return out
}
```

- [ ] **Step 4: Run the tests**

Run: `go test ./hivepaas_app/service/loggingservice/... -count=1`
Expected: PASS.

- [ ] **Step 5: Prove the mode test bites**

In `toSwarmServiceSpec`, change the `if opts.Global` branch to set `Replicated` as well as `Global`. `TestToSwarmServiceSpecGlobalModeLeavesReplicatedNil` must fail. Restore it.

- [ ] **Step 6: Commit**

```bash
git add hivepaas_app/service/loggingservice/loggingserviceimpl/swarmspec.go hivepaas_app/service/loggingservice/loggingserviceimpl/swarmspec_test.go
git commit -m "feat(logging): translate runtime specs into swarm services"
```

---

### Task 9: Apply and tear down

**Files:**
- Create: `hivepaas_app/service/loggingservice/loggingserviceimpl/deploy.go`
- Create: `hivepaas_app/service/loggingservice/loggingserviceimpl/deploy_test.go`
- Modify: `hivepaas_app/registry/provides.go`

**Interfaces:**
- Consumes: everything from Tasks 4, 5, 7, 8, and the decision from Task 6.
- Produces: `Apply`, `TearDown`, `Status` on `*service`; `loggingserviceimpl.New` registered with fx.

Context on ordering, from the spec: the backend comes up first and the collector waits for it to answer `Ping`, because a collector shipping into a backend that does not exist yet buffers and retries for no reason. Tearing down runs the other way. **The data volume is never removed** — switching the backend off, or from managed to unmanaged, must not destroy the logs. Removing it is a separate, explicit action through the volume API.

- [ ] **Step 1: Write the failing test**

Create `hivepaas_app/service/loggingservice/loggingserviceimpl/deploy_test.go`:

```go
package loggingserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/docker"
)

// fakeDocker records what was asked of it. Embedding the interface means only
// the methods this path uses need a body; anything else panics loudly instead
// of passing quietly.
type fakeDocker struct {
	docker.Manager
	created     []*swarm.ServiceSpec
	removed     []string
	removeCalls int
}

func (f *fakeDocker) ServiceCreate(
	_ context.Context, spec *swarm.ServiceSpec, _ ...docker.ServiceCreateOption,
) (*client.ServiceCreateResult, error) {
	f.created = append(f.created, spec)
	return &client.ServiceCreateResult{ID: "svc-" + spec.Name}, nil
}

func (f *fakeDocker) ServiceRemove(
	_ context.Context, serviceID string, _ ...docker.ServiceRemoveOption,
) (*client.ServiceRemoveResult, error) {
	f.removed = append(f.removed, serviceID)
	f.removeCalls++
	return &client.ServiceRemoveResult{}, nil
}

func enabledConfig() *entity.Logging {
	return &entity.Logging{
		Enabled: true,
		Sources: entity.LoggingSources{Apps: true},
		Collector: entity.LoggingCollector{
			Type:    entity.LoggingCollectorTypeVlagent,
			Managed: true,
		},
		Backend: entity.LoggingBackend{
			Type:    entity.LoggingBackendTypeVictoriaLogs,
			Managed: true,
			VictoriaLogs: &entity.LoggingVictoriaLogs{
				NodeID:   "node-1",
				VolumeID: "vol-1",
			},
		},
	}
}

func TestDeployCreatesBackendBeforeCollector(t *testing.T) {
	fd := &fakeDocker{}
	s := &service{dockerManager: fd}

	err := s.deploy(context.Background(), enabledConfig())
	if err != nil {
		t.Fatalf("deploy: %v", err)
	}

	if len(fd.created) != 2 {
		t.Fatalf("want two services, got %d", len(fd.created))
	}
	assert.Equal(t, ServiceNameBackend, fd.created[0].Name,
		"the collector must not be started before something to ship to exists")
	assert.Equal(t, ServiceNameCollector, fd.created[1].Name)
}

func TestDeployRunsTheCollectorOnEveryNode(t *testing.T) {
	fd := &fakeDocker{}
	s := &service{dockerManager: fd}

	if err := s.deploy(context.Background(), enabledConfig()); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	collector := fd.created[1]
	assert.NotNil(t, collector.Mode.Global)
	assert.Nil(t, collector.Mode.Replicated)
}

func TestDeployPinsTheBackendToItsNode(t *testing.T) {
	fd := &fakeDocker{}
	s := &service{dockerManager: fd}

	if err := s.deploy(context.Background(), enabledConfig()); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	backend := fd.created[0]
	if backend.TaskTemplate.Placement == nil {
		t.Fatal("the backend is not pinned")
	}
	assert.Contains(t, backend.TaskTemplate.Placement.Constraints, "node.id==node-1")
}

func TestDeployRefusesAManagedBackendWithNoVolume(t *testing.T) {
	cfg := enabledConfig()
	cfg.Backend.VictoriaLogs.VolumeID = ""
	s := &service{dockerManager: &fakeDocker{}}

	err := s.deploy(context.Background(), cfg)

	assert.Error(t, err, "a backend with no volume loses every log on restart")
}

func TestDeployRefusesAManagedBackendWithNoNode(t *testing.T) {
	cfg := enabledConfig()
	cfg.Backend.VictoriaLogs.NodeID = ""
	s := &service{dockerManager: &fakeDocker{}}

	err := s.deploy(context.Background(), cfg)

	assert.Error(t, err)
}

// An unmanaged backend is somebody else's service: HivePaaS ships to it and
// never creates it.
func TestDeploySkipsAnUnmanagedBackend(t *testing.T) {
	cfg := enabledConfig()
	cfg.Backend.Managed = false
	cfg.Backend.VictoriaLogs = nil
	cfg.Backend.Ingest = &entity.LoggingEndpoint{URL: "https://logs.example/insert"}
	fd := &fakeDocker{}
	s := &service{dockerManager: fd}

	if err := s.deploy(context.Background(), cfg); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	if len(fd.created) != 1 {
		t.Fatalf("want only the collector, got %d services", len(fd.created))
	}
	assert.Equal(t, ServiceNameCollector, fd.created[0].Name)
}

func TestTearDownRemovesCollectorBeforeBackend(t *testing.T) {
	fd := &fakeDocker{}
	s := &service{dockerManager: fd}

	if err := s.TearDown(context.Background()); err != nil {
		t.Fatalf("TearDown: %v", err)
	}

	if len(fd.removed) != 2 {
		t.Fatalf("want two removals, got %d", len(fd.removed))
	}
	assert.Equal(t, ServiceNameCollector, fd.removed[0],
		"stop shipping before removing the thing being shipped to")
	assert.Equal(t, ServiceNameBackend, fd.removed[1])
}

// Turning logging off must not destroy the logs. For a plain local volume the
// data is inside the volume, so removing it is unrecoverable; removal is a
// separate, explicit action through the volume API.
func TestTearDownKeepsTheDataVolume(t *testing.T) {
	fd := &fakeDocker{}
	s := &service{dockerManager: fd}

	if err := s.TearDown(context.Background()); err != nil {
		t.Fatalf("TearDown: %v", err)
	}

	for _, r := range fd.removed {
		assert.NotContains(t, r, "vol-", "the data volume must survive tear-down")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./hivepaas_app/service/loggingservice/loggingserviceimpl/ -run 'TestDeploy|TestTearDown' -v`
Expected: FAIL to build, `s.deploy undefined`.

- [ ] **Step 3: Write the implementation**

Create `hivepaas_app/service/loggingservice/loggingserviceimpl/deploy.go`:

```go
package loggingserviceimpl

import (
	"context"
	"errors"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/services/logging"
	"github.com/hivepaas/hivepaas/services/logging/victorialogs"
	"github.com/hivepaas/hivepaas/services/logging/vlagent"
)

// Apply makes the cluster match the stored configuration.
func (s *service) Apply(ctx context.Context, db database.IDB) error {
	setting, err := s.settingRepo.GetSingle(ctx, db, nil, base.SettingTypeLogging, true)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	if setting == nil {
		// Never configured, which is the default: there is nothing to run and
		// nothing to remove.
		return nil
	}

	cfg, err := setting.AsLogging()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if cfg == nil || !cfg.Enabled {
		return s.TearDown(ctx)
	}
	if err := cfg.Decrypt(); err != nil {
		return hperrors.Wrap(err)
	}
	return s.deploy(ctx, cfg)
}

// deploy brings the stack up in the order that avoids a collector shipping into
// a backend that does not exist yet.
func (s *service) deploy(ctx context.Context, cfg *entity.Logging) error {
	ingestURL, err := s.deployBackend(ctx, cfg)
	if err != nil {
		return hperrors.Wrap(err)
	}

	if !cfg.Collector.Managed {
		// The collector is somebody else's. The backend is up and the operator
		// points their own collector at it.
		return nil
	}

	spec, err := s.buildCollectSpec(cfg, ingestURL)
	if err != nil {
		return hperrors.Wrap(err)
	}

	collectorImage := ""
	if cfg.Collector.Vlagent != nil {
		collectorImage = cfg.Collector.Vlagent.Image
	}
	collector, err := logging.NewCollector(
		logging.CollectorType(cfg.Collector.Type),
		&logging.CollectorConfig{Vlagent: &vlagent.Config{Image: collectorImage}},
	)
	if err != nil {
		return hperrors.Wrap(err)
	}

	rt, err := collector.Configure(spec)
	if err != nil {
		return hperrors.Wrap(err)
	}

	svcSpec, err := toSwarmServiceSpec(rt, swarmSpecOpts{
		Name:   ServiceNameCollector,
		Global: true, // one task per node; this is the only global service HivePaaS creates
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	if _, err := s.dockerManager.ServiceCreate(ctx, svcSpec); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// deployBackend creates the backend when HivePaaS owns it, and reports where
// the collector should write either way.
func (s *service) deployBackend(ctx context.Context, cfg *entity.Logging) (string, error) {
	if !cfg.Backend.Managed {
		ep, err := toEndpoint(cfg.Backend.Ingest)
		if err != nil {
			return "", hperrors.Wrap(err)
		}
		if ep.URL == "" {
			return "", hperrors.Wrap(logging.ErrIngestEndpointRequired)
		}
		return ep.URL, nil
	}

	vl := cfg.Backend.VictoriaLogs
	if vl == nil {
		return "", hperrors.Wrap(loggingservice.ErrNotConfigured)
	}
	if vl.NodeID == "" {
		return "", hperrors.Wrap(loggingservice.ErrBackendNodeMissing)
	}
	if vl.VolumeID == "" {
		// Without a volume the logs live in the container's writable layer and
		// vanish on the next restart, silently.
		return "", hperrors.Wrap(loggingservice.ErrVolumeMissing)
	}

	deployer, err := logging.NewDeployer(
		logging.BackendType(cfg.Backend.Type),
		&logging.BackendConfig{VictoriaLogs: &victorialogs.Config{
			Image:               vl.Image,
			DataVolumeName:      vl.VolumeID,
			Retention:           vl.Retention.ToDuration(),
			MaxDiskUsagePercent: vl.MaxDiskUsagePercent,
		}},
	)
	if err != nil {
		return "", hperrors.Wrap(err)
	}

	rt, err := deployer.RuntimeSpec()
	if err != nil {
		return "", hperrors.Wrap(err)
	}

	svcSpec, err := toSwarmServiceSpec(rt, swarmSpecOpts{
		Name:   ServiceNameBackend,
		NodeID: vl.NodeID,
	})
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	if _, err := s.dockerManager.ServiceCreate(ctx, svcSpec); err != nil {
		return "", hperrors.Wrap(err)
	}

	return fmt.Sprintf("http://%s:%d%s", ServiceNameBackend, victorialogs.DefaultHTTPPort, victorialogs.IngestPath), nil
}

// TearDown removes what Apply created, in the reverse order, and leaves the
// data volume alone.
//
// The volume is deliberately kept. Turning logging off, or moving to a backend
// HivePaaS does not run, must not destroy what has been collected - for a plain
// local volume the data is inside the volume and removal is unrecoverable.
// Deleting it is a separate, explicit action through the volume API.
func (s *service) TearDown(ctx context.Context) error {
	var errs error
	for _, name := range []string{ServiceNameCollector, ServiceNameBackend} {
		_, err := s.dockerManager.ServiceRemove(ctx, name)
		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			errs = errors.Join(errs, err)
		}
	}
	if errs != nil {
		return hperrors.Wrap(errs)
	}
	return nil
}

// Status reports what is running.
func (s *service) Status(ctx context.Context, db database.IDB) (*loggingservice.Status, error) {
	setting, err := s.settingRepo.GetSingle(ctx, db, nil, base.SettingTypeLogging, true)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	out := &loggingservice.Status{}
	if setting == nil {
		return out, nil
	}
	cfg, err := setting.AsLogging()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if cfg != nil {
		out.Enabled = cfg.Enabled
	}

	if svc, err := s.clusterService.ServiceInspect(ctx, ServiceNameBackend, true); err == nil && svc != nil {
		out.BackendServiceID = svc.ID
		out.BackendReady = true
	}
	if svc, err := s.clusterService.ServiceInspect(ctx, ServiceNameCollector, true); err == nil && svc != nil {
		out.CollectorService = svc.ID
	}
	return out, nil
}
```

- [ ] **Step 4: Register with fx**

In `hivepaas_app/registry/provides.go`, add the import beside the other service implementations:

```go
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice/loggingserviceimpl"
```

and add the constructor to the provider list, beside `volumeserviceimpl.New`:

```go
	loggingserviceimpl.New,
```

- [ ] **Step 5: Run the tests and build**

```bash
go build ./...
go test ./hivepaas_app/service/loggingservice/... -count=1 -v
```
Expected: build clean, all tests PASS.

- [ ] **Step 6: Prove the ordering tests bite**

In `deploy`, move the collector creation above `deployBackend`. `TestDeployCreatesBackendBeforeCollector` must fail. Restore the order. Then in `TearDown`, reverse the name list; `TestTearDownRemovesCollectorBeforeBackend` must fail. Restore it.

- [ ] **Step 7: Commit**

```bash
git add hivepaas_app/service/loggingservice hivepaas_app/registry/provides.go
git commit -m "feat(logging): apply and tear down the logging stack"
```

---

### Task 10: Managed log driver for app services

**Files:**
- Create: `hivepaas_app/usecase/appsettingsuc/log_driver.go`
- Create: `hivepaas_app/usecase/appsettingsuc/log_driver_test.go`
- Modify: `hivepaas_app/usecase/appsettingsuc/container_settings_update.go:216-232`

**Interfaces:**
- Consumes: nothing from earlier tasks; this is app-settings code.
- Produces:
  - `const LogLabelAppID = "hivepaas.app.id"`, `LogLabelProjectID = "hivepaas.project.id"`, `LogLabelProjectEnvID = "hivepaas.projectEnv.id"`
  - `func managedLogDriver(app *entity.App, requested *appsettingsdto.LogDriver) *swarm.Driver`
  - `func isCollectible(d *swarm.Driver) bool`

Context: `prepareUpdatingAppContainerLogDriver` currently takes whatever the request says. Collection depends on apps writing `json-file` with the identifying labels, so an app that switches to `none` or `syslog` silently disappears from logging.

The resolution from the spec is neither to seize the field nor to ignore it: an app that asks for **no** driver gets the managed one; an app that deliberately asks for another keeps it, and is reported as not collected. `isCollectible` is what the reporting is built on.

`--log-opt labels=a,b` only emits the labels the container actually carries, so the service's own labels must carry them too.

- [ ] **Step 1: Write the failing test**

Create `hivepaas_app/usecase/appsettingsuc/log_driver_test.go`:

```go
package appsettingsuc

import (
	"strings"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func testApp() *entity.App {
	return &entity.App{
		ID:           "app-1",
		ProjectID:    "proj-1",
		ProjectEnvID: "env-1",
	}
}

// With no driver asked for, HivePaaS manages it: json-file, and the labels the
// daemon writes into every line as `attrs`. The app cannot forge those, which
// is what makes them usable for scoping a query.
func TestManagedLogDriverDefaultsToJSONFileWithLabels(t *testing.T) {
	d := managedLogDriver(testApp(), nil)

	if d == nil {
		t.Fatal("want a driver")
	}
	assert.Equal(t, "json-file", d.Name)
	labels := d.Options["labels"]
	assert.Contains(t, labels, LogLabelAppID)
	assert.Contains(t, labels, LogLabelProjectID)
	assert.Contains(t, labels, LogLabelProjectEnvID)
}

// A driver the operator chose is left alone. Seizing it would take away a real
// capability - shipping straight to their own syslog - without telling them.
func TestManagedLogDriverRespectsAnExplicitChoice(t *testing.T) {
	d := managedLogDriver(testApp(), &appsettingsdto.LogDriver{Name: "syslog"})

	if d == nil {
		t.Fatal("want a driver")
	}
	assert.Equal(t, "syslog", d.Name)
	assert.NotContains(t, d.Options["labels"], LogLabelAppID)
}

// json-file chosen explicitly still gets the labels: the operator asked for the
// driver collection needs, so there is nothing to take away.
func TestManagedLogDriverAddsLabelsToAnExplicitJSONFile(t *testing.T) {
	d := managedLogDriver(testApp(), &appsettingsdto.LogDriver{
		Name:    "json-file",
		Options: map[string]string{"max-size": "10m"},
	})

	if d == nil {
		t.Fatal("want a driver")
	}
	assert.Contains(t, d.Options["labels"], LogLabelAppID)
	assert.Equal(t, "10m", d.Options["max-size"], "the operator's own options survive")
}

func TestIsCollectible(t *testing.T) {
	assert.True(t, isCollectible(&swarm.Driver{Name: "json-file"}))
	assert.True(t, isCollectible(nil), "no driver means the daemon default, which is json-file")
	assert.False(t, isCollectible(&swarm.Driver{Name: "syslog"}))
	assert.False(t, isCollectible(&swarm.Driver{Name: "none"}))
}

// The label list is what the daemon matches against the container's own labels,
// so the order has to be stable or every apply looks like a change.
func TestManagedLogDriverLabelListIsStable(t *testing.T) {
	first := managedLogDriver(testApp(), nil).Options["labels"]
	second := managedLogDriver(testApp(), nil).Options["labels"]

	assert.Equal(t, first, second)
	assert.Equal(t, 3, len(strings.Split(first, ",")))
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./hivepaas_app/usecase/appsettingsuc/ -run 'TestManagedLogDriver|TestIsCollectible' -v`
Expected: FAIL to build, `undefined: managedLogDriver`.

- [ ] **Step 3: Write the implementation**

Create `hivepaas_app/usecase/appsettingsuc/log_driver.go`:

```go
package appsettingsuc

import (
	"sort"
	"strings"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

// The labels the daemon copies into every log line's `attrs` block.
//
// They are written by the daemon, not by the application, which is what makes
// them safe to scope a query by: a container cannot claim to be another app.
const (
	LogLabelAppID        = "hivepaas.app.id"
	LogLabelProjectID    = "hivepaas.project.id"
	LogLabelProjectEnvID = "hivepaas.projectEnv.id"
)

// collectibleDrivers write files the collector can tail. An empty name is the
// daemon default, which is json-file.
var collectibleDrivers = map[string]bool{
	"":          true,
	"json-file": true,
}

// managedLogDriver decides the log driver a service runs with.
//
// An app that asks for nothing gets the one collection needs. An app that names
// a driver keeps it - taking that away would remove a real capability, shipping
// directly to the operator's own syslog, without saying so. The cost is that
// such an app is not collected, which is reported rather than hidden: see
// isCollectible.
func managedLogDriver(app *entity.App, requested *appsettingsdto.LogDriver) *swarm.Driver {
	name := ""
	options := map[string]string{}
	if requested != nil {
		name = requested.Name
		for k, v := range requested.Options {
			options[k] = v
		}
	}

	if name == "" {
		name = "json-file"
	}
	if !collectibleDrivers[name] {
		return &swarm.Driver{Name: name, Options: options}
	}

	// `labels` names which of the container's own labels the daemon should copy
	// into each line. Sorted, so the same app always produces the same spec.
	labels := []string{LogLabelAppID, LogLabelProjectID, LogLabelProjectEnvID}
	sort.Strings(labels)
	options["labels"] = strings.Join(labels, ",")

	return &swarm.Driver{Name: name, Options: options}
}

// isCollectible reports whether a service with this driver writes logs the
// collector can read. It is what the "not collected" list in the logging
// settings is built from.
func isCollectible(d *swarm.Driver) bool {
	if d == nil {
		return true
	}
	return collectibleDrivers[d.Name]
}

// appLogLabels are the labels a service must carry for the driver's `labels`
// option to have anything to copy.
func appLogLabels(app *entity.App) map[string]string {
	return map[string]string{
		LogLabelAppID:        app.ID,
		LogLabelProjectID:    app.ProjectID,
		LogLabelProjectEnvID: app.ProjectEnvID,
	}
}
```

- [ ] **Step 4: Wire it into the update path**

Replace the body of `prepareUpdatingAppContainerLogDriver` in `container_settings_update.go` so that the driver comes from `managedLogDriver` and the service carries the labels it names:

```go
func (uc *UC) prepareUpdatingAppContainerLogDriver(
	req *appsettingsdto.UpdateAppContainerSettingsReq,
	data *updateAppContainerSettingsData,
) {
	taskSpec := &data.ServiceSpec.TaskTemplate

	taskSpec.LogDriver = managedLogDriver(data.App, req.LogDriver)

	// The driver's `labels` option names labels to copy; it copies nothing
	// unless the service actually carries them.
	if data.ServiceSpec.Labels == nil {
		data.ServiceSpec.Labels = map[string]string{}
	}
	for k, v := range appLogLabels(data.App) {
		data.ServiceSpec.Labels[k] = v
	}
}
```

Read the surrounding function first: `data` may name its fields differently. Keep the existing names and only change what the driver is built from.

- [ ] **Step 5: Run the package tests**

Run: `go test ./hivepaas_app/usecase/appsettingsuc/... -count=1`
Expected: PASS. Existing tests asserting the old pass-through behaviour will fail - read each one. If it asserted that a nil request clears the driver, that expectation is what this task changes, so update the test and say why in the commit. If it asserted something else, the change is wrong.

- [ ] **Step 6: Commit**

```bash
git add hivepaas_app/usecase/appsettingsuc
git commit -m "feat(logging): managed log driver for app services"
```

---

### Task 11: The configuration API

**Files:**
- Create: `hivepaas_app/usecase/system/logginguc/uc.go`
- Create: `hivepaas_app/usecase/system/logginguc/get.go`
- Create: `hivepaas_app/usecase/system/logginguc/update.go`
- Create: `hivepaas_app/usecase/system/logginguc/logginguc_test.go`
- Create: `hivepaas_app/usecase/system/logginguc/loggingdto/settings.go`
- Modify: `hivepaas_app/registry/provides.go`

**Interfaces:**
- Consumes: `entity.Logging` (Task 5); `loggingservice.Service` (Tasks 7-9).
- Produces:
  - `func (uc *UC) GetSettings(ctx, auth) (*loggingdto.GetSettingsResp, error)`
  - `func (uc *UC) UpdateSettings(ctx, auth, req) (*loggingdto.UpdateSettingsResp, error)`
  - `func validateSettings(cfg *entity.Logging) error`

Context: this is the singleton pattern — `settingRepo.GetSingle(ctx, db, nil, base.SettingTypeLogging, true)`, creating the row on first write. After a successful write, `Apply` runs so the cluster matches what was just saved.

- [ ] **Step 1: Write the failing validation test**

Create `hivepaas_app/usecase/system/logginguc/logginguc_test.go`:

```go
package logginguc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func validEnabled() *entity.Logging {
	return &entity.Logging{
		Enabled: true,
		Sources: entity.LoggingSources{Apps: true},
		Collector: entity.LoggingCollector{
			Type: entity.LoggingCollectorTypeVlagent, Managed: true,
		},
		Backend: entity.LoggingBackend{
			Type: entity.LoggingBackendTypeVictoriaLogs, Managed: true,
			VictoriaLogs: &entity.LoggingVictoriaLogs{NodeID: "n1", VolumeID: "v1"},
		},
	}
}

func TestValidateAcceptsAWorkingConfiguration(t *testing.T) {
	assert.NoError(t, validateSettings(validEnabled()))
}

// Disabled is the default and must stay valid however empty it is, or an
// operator could not turn logging off without filling in a form first.
func TestValidateAcceptsDisabledAndEmpty(t *testing.T) {
	assert.NoError(t, validateSettings(&entity.Logging{}))
}

func TestValidateRejectsEnabledWithNoSources(t *testing.T) {
	cfg := validEnabled()
	cfg.Sources = entity.LoggingSources{}

	assert.Error(t, validateSettings(cfg), "collecting nothing is a misconfiguration, not a quiet no-op")
}

func TestValidateRejectsAManagedBackendWithNoVolume(t *testing.T) {
	cfg := validEnabled()
	cfg.Backend.VictoriaLogs.VolumeID = ""

	assert.Error(t, validateSettings(cfg))
}

func TestValidateRejectsAnUnmanagedBackendWithNoIngestURL(t *testing.T) {
	cfg := validEnabled()
	cfg.Backend.Managed = false
	cfg.Backend.VictoriaLogs = nil

	assert.Error(t, validateSettings(cfg))
}

// An unmanaged backend is legitimate: the user runs their own VictoriaLogs and
// HivePaaS only ships to it. Type and Managed are separate questions.
func TestValidateAcceptsAnUnmanagedBackend(t *testing.T) {
	cfg := validEnabled()
	cfg.Backend.Managed = false
	cfg.Backend.VictoriaLogs = nil
	cfg.Backend.Ingest = &entity.LoggingEndpoint{URL: "https://logs.example/insert"}

	assert.NoError(t, validateSettings(cfg))
}

func TestValidateRejectsAForwardWithNoURL(t *testing.T) {
	cfg := validEnabled()
	cfg.Forwards = []entity.LoggingForward{{Name: "siem"}}

	assert.Error(t, validateSettings(cfg))
}

func TestValidateRejectsAForwardWithNoName(t *testing.T) {
	cfg := validEnabled()
	cfg.Forwards = []entity.LoggingForward{{
		Endpoint: entity.LoggingEndpoint{URL: "https://siem.example/ingest"},
	}}

	assert.Error(t, validateSettings(cfg), "a forward with no name cannot be reported on or removed")
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./hivepaas_app/usecase/system/logginguc/ -v`
Expected: FAIL to build, `undefined: validateSettings`.

- [ ] **Step 3: Write the validation**

Create `hivepaas_app/usecase/system/logginguc/update.go` with the validation first:

```go
package logginguc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/services/logging"
)

// validateSettings refuses a configuration that would deploy something broken.
//
// Nothing is checked while logging is disabled: that is the default state, and
// requiring a complete form to turn the feature off would be perverse.
func validateSettings(cfg *entity.Logging) error {
	if cfg == nil || !cfg.Enabled {
		return nil
	}

	if !cfg.Sources.Apps && !cfg.Sources.HivePaaS && !cfg.Sources.TraefikAccess && !cfg.Sources.Nodes {
		return hperrors.Wrap(logging.ErrNoSources)
	}

	if cfg.Backend.Managed {
		vl := cfg.Backend.VictoriaLogs
		if vl == nil {
			return hperrors.Wrap(loggingservice.ErrNotConfigured)
		}
		if vl.NodeID == "" {
			return hperrors.Wrap(loggingservice.ErrBackendNodeMissing)
		}
		if vl.VolumeID == "" {
			return hperrors.Wrap(loggingservice.ErrVolumeMissing)
		}
	} else if cfg.Backend.Ingest == nil || cfg.Backend.Ingest.URL == "" {
		return hperrors.Wrap(logging.ErrIngestEndpointRequired)
	}

	for i := range cfg.Forwards {
		f := &cfg.Forwards[i]
		if f.Name == "" {
			return hperrors.NewArgumentInvalid("Forwards").
				WithExtraDetail("a forward needs a name to be reported on or removed")
		}
		if f.Endpoint.URL == "" {
			return hperrors.Wrap(logging.ErrIngestEndpointRequired)
		}
	}

	return nil
}
```

- [ ] **Step 4: Run the validation tests**

Run: `go test ./hivepaas_app/usecase/system/logginguc/ -v`
Expected: PASS, all eight tests.

- [ ] **Step 5: Write the usecase around it**

Create `hivepaas_app/usecase/system/logginguc/uc.go`:

```go
package logginguc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
)

type UC struct {
	db             database.IDB
	settingRepo    repository.SettingRepo
	loggingService loggingservice.Service
}

// New builds the usecase. fx supplies the arguments from the provider list.
func New(
	db database.IDB,
	settingRepo repository.SettingRepo,
	loggingService loggingservice.Service,
) *UC {
	return &UC{db: db, settingRepo: settingRepo, loggingService: loggingService}
}
```

Create `hivepaas_app/usecase/system/logginguc/loggingdto/settings.go`:

```go
package loggingdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type GetSettingsReq struct {
	basedto.BaseReq
}

type GetSettingsResp struct {
	Data *entity.Logging `json:"data"`
	// Status is what is actually running, which is not always what is stored.
	Status *SettingsStatus `json:"status"`
}

type SettingsStatus struct {
	BackendReady bool     `json:"backendReady"`
	ExcludedApps []string `json:"excludedApps"`
}

type UpdateSettingsReq struct {
	basedto.BaseReq

	Data *entity.Logging `json:"data" validate:"required"`
}

type UpdateSettingsResp struct {
	Data *entity.Logging `json:"data"`
}
```

Create `hivepaas_app/usecase/system/logginguc/get.go`:

```go
package logginguc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/logginguc/loggingdto"
)

// GetSettings reads the stored configuration, or the disabled default when
// nothing has been stored yet.
func (uc *UC) GetSettings(
	ctx context.Context,
	_ *basedto.Auth,
	_ *loggingdto.GetSettingsReq,
) (*loggingdto.GetSettingsResp, error) {
	setting, err := uc.settingRepo.GetSingle(ctx, uc.db, nil, base.SettingTypeLogging, true)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}

	cfg := &entity.Logging{}
	if setting != nil {
		cfg, err = setting.AsLogging()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	status, err := uc.loggingService.Status(ctx, uc.db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &loggingdto.GetSettingsResp{
		Data: cfg,
		Status: &loggingdto.SettingsStatus{
			BackendReady: status.BackendReady,
			ExcludedApps: status.ExcludedApps,
		},
	}, nil
}
```

Append to `update.go`:

```go
// UpdateSettings stores the configuration and makes the cluster match it.
func (uc *UC) UpdateSettings(
	ctx context.Context,
	_ *basedto.Auth,
	req *loggingdto.UpdateSettingsReq,
) (*loggingdto.UpdateSettingsResp, error) {
	if err := validateSettings(req.Data); err != nil {
		return nil, hperrors.Wrap(err)
	}

	setting, err := uc.settingRepo.GetSingle(ctx, uc.db, nil, base.SettingTypeLogging, true)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}

	timeNow := timeutil.NowUTC()
	if setting == nil {
		setting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     base.HivepaasScope,
			Type:      base.SettingTypeLogging,
			Status:    base.SettingStatusActive,
			Name:      "Logging",
			Version:   entity.CurrentLoggingVersion,
			CreatedAt: timeNow,
		}
	}
	setting.UpdatedAt = timeNow
	if err := setting.SetData(req.Data); err != nil {
		return nil, hperrors.Wrap(err)
	}

	if err := uc.settingRepo.Upsert(ctx, uc.db, setting,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols); err != nil {
		return nil, hperrors.Wrap(err)
	}

	// Deploy what was just saved. A stored configuration the cluster does not
	// match is the failure mode worth avoiding here.
	if err := uc.loggingService.Apply(ctx, uc.db); err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &loggingdto.UpdateSettingsResp{Data: req.Data}, nil
}
```

Add the imports `update.go` now needs: `context`, `errors`, `basedto`, `base`, `timeutil`, `gofn`, `ulid`, `loggingdto`. Check `settingRepo.Upsert`'s real signature before writing the call — if the repository exposes only `UpsertMulti`, use that with a one-element slice.

- [ ] **Step 6: Register with fx**

In `hivepaas_app/registry/provides.go`, add the import and the constructor beside the other usecases:

```go
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/logginguc"
	...
	logginguc.New,
```

- [ ] **Step 7: Build, test, lint**

```bash
go build ./...
go test ./hivepaas_app/usecase/system/logginguc/... ./hivepaas_app/service/loggingservice/... ./services/logging/... -count=1
golangci-lint --timeout=3m run ./services/logging/... ./hivepaas_app/service/loggingservice/... ./hivepaas_app/usecase/system/logginguc/...
```
Expected: build clean, tests PASS, `0 issues.`

- [ ] **Step 8: Run the full suite**

Run: `./scripts/test.sh 2>&1 | grep -E "FAIL|DONE\.|panic"`
Expected: `DONE.` and no `FAIL`.

- [ ] **Step 9: Commit**

```bash
git add hivepaas_app/usecase/system/logginguc hivepaas_app/registry/provides.go
git commit -m "feat(logging): configuration API"
```

---

## Verification

With every task done:

```bash
go build ./...
go test ./services/logging/... ./hivepaas_app/service/loggingservice/... -count=1
make lint
./scripts/test.sh
```

And these properties hold, each with a test naming it:

- Nothing under `services/logging` imports `entity`, `base`, `basedto`, `repository`, `infra/database`, or `services/docker` (Task 1).
- A credential reaches `services/logging` decrypted (Task 7).
- vlagent's positional flag arrays stay aligned, so no destination receives another's token (Task 3).
- The collector runs `Global`, and `Replicated` is left nil (Task 8).
- The backend is created before the collector; the collector is removed before the backend (Task 9).
- Tear-down never removes the data volume (Task 9).
- An app that names its own log driver keeps it (Task 10).

## One property this plan cannot test, and where it belongs

Spec section 7 asks for the provenance properties as regression tests. Most are
read-path and arrive with the next plan. One is collection-side and **cannot be
unit tested**: that vlagent does not recursively parse a container's own JSON,
so a forged `attrs.hivepaas.app.id` in an app's stdout never displaces the
daemon's. That is vlagent's runtime behaviour, not HivePaaS code, and proving it
means running a real collector against a real docker daemon.

It was verified by hand while the spec was written, against vlagent v1.52.0 -
which is why `vlagent.DefaultImage` pins that version rather than tracking a
tag. Treat a version bump as the trigger to re-run it:

1. run a container labelled `hivepaas.app.id=REAL` with
   `--log-opt labels=hivepaas.app.id`, printing
   `{"attrs.hivepaas.app.id":"SPOOFED"}` on stdout
2. collect it with the new vlagent into VictoriaLogs
3. assert the stored record still has `attrs.hivepaas.app.id` = `REAL`, and that
   a filter on the forged value matches nothing

Do not add this to `./scripts/test.sh`: it needs a docker daemon and takes tens
of seconds. It belongs in whatever integration suite the project grows, and
until then in the checklist for changing that pinned image.

## What this plan does not do

The read path. `victorialogs.Client.Query` returns an error saying so. Building LogsQL from structured parameters, refusing raw queries from non-admins, and always passing `result_prefix` to `unpack_json` are the subject of the next plan, together with the dashboard. Until then the feature is useful through `Forwards`: logs are collected and shipped to a system that can already display them.
