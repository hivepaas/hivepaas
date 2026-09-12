# Logging

HivePaaS has no way to read an app's logs after its container is gone, and no
way to read them at all from a node other than the one it runs on. This adds a
logging subsystem: a collector on every node, a queryable backend, and a read
path in the dashboard - built so that the collector and the backend are each
replaceable, and so that a user can point either of them at their own system.

Default is off. Nothing is deployed until an operator turns it on.

## Scope

**In:** app logs (the point of the feature), HivePaaS's own logs behind an
explicit toggle, Traefik access logs, and a global on/off. One managed backend
(VictoriaLogs), one managed collector (vlagent), plus forwarding a copy to
systems HivePaaS does not own.

**Out, deliberately:** node/system logs (recorded as low priority, not designed
here); per-app or per-project logging overrides, since the configuration is
global by decision - if a noisy app ever needs excluding, that belongs in a
separate app-scoped setting, not as an exclusion list bolted onto this one;
alerting on log contents.

## 1. Package structure

Two blocks, cut along one line: **what knows the protocol** versus **what knows
the cluster**.

```
services/logging/
  loggingmodel/          interfaces, types, errors - depends on no implementation
  victorialogs/          implements Backend + Deployer
  vlagent/               implements Collector
  logging.go             re-exported aliases (logging.Backend, logging.Collector)
  factory.go             NewBackend(type, cfg), NewCollector(type, cfg)

hivepaas_app/service/loggingservice/
  loggingserviceimpl/    lifecycle: swarm services, volumes, placement, settings
```

This follows `services/backup` exactly: `backupmodel` + `kopia` + `engine.go`
re-exports + `factory.go` switching on an engine type.

### Independence rule

`services/backup` imports only `hperrors` and `pkg/envutil` - no `entity`, no
`base`, no `basedto`. It defines its own `Storage`, `StorageS3`, `StorageLocal`,
and `backupreposervice` converts from `entity` at the boundary
(`backupreposerviceimpl/engine.go:136`).

`services/logging` holds to the same standard. It never imports `entity`. It
defines its own configuration types, and `loggingservice` closes over `entity`
and passes plain values across.

**Credentials cross the boundary decrypted.** `buildS3Storage` calls
`cloudStorage.Decrypt()` and `.GetPlain()` so that `backupmodel.StorageS3`
carries a plain `string`. `loggingmodel.Endpoint` does the same: the entity
holds `EncryptedField`, the service package never learns that type exists.

| Package | May import | Knows docker | Knows the database |
|---|---|---|---|
| `loggingmodel` | `hperrors`, `pkg/*` | no | no |
| `victorialogs`, `vlagent` | `hperrors`, `loggingmodel` | no | no |
| `loggingservice` | entity, docker, clusterservice, ... | yes | yes |

## 2. Configuration entity

A new setting type, `SettingTypeLogging`, rather than another block inside
`HivePaaSService`. `HivePaaSService` describes how HivePaaS itself runs
(replicas, workers, proxy); logging is a subsystem with its own lifecycle, its
own version/migration lane - which it will need, because the schema grows with
every backend added - and its own probation/revert.

```go
const CurrentLoggingVersion = 1

type Logging struct {
    Enabled   bool              `json:"enabled,omitempty"`
    Sources   LoggingSources    `json:"sources"`
    Collector LoggingCollector  `json:"collector"`
    Backend   LoggingBackend    `json:"backend"`
    Forwards  []LoggingForward  `json:"forwards,omitempty"`
}

type LoggingSources struct {
    Apps          bool `json:"apps,omitempty"`
    HivePaaS      bool `json:"hivepaas,omitempty"`
    TraefikAccess bool `json:"traefikAccess,omitempty"`
    Nodes         bool `json:"nodes,omitempty"`
}
```

### Type and Managed answer different questions

```go
type LoggingBackend struct {
    Type    LoggingBackendType `json:"type,omitempty"`
    Managed bool               `json:"managed,omitempty"`

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
    // ExtraArgs is an escape hatch for flags this schema does not model.
    // It is appended after the flags HivePaaS derives, and cannot replace them.
    ExtraArgs []string `json:"extraArgs,omitempty"`
}
```

`Type` says what it is; `Managed` says whether HivePaaS deploys it. Folding
these into one discriminator - making `external` a value of `Type` - makes the
most valuable case inexpressible: a user running their own VictoriaLogs, which
HivePaaS can still query because it speaks LogsQL, but does not own. That is
`Type: victoria-logs, Managed: false`.

This is the lesson from `entity.ClusterVolume`, where `Managed` carried two
meanings at once and had to be separated as soon as CSI was considered.

### Ingest and Query are separate endpoints

VictoriaLogs single-node serves both from one base URL, which makes the split
look redundant. Loki does not: its push and query paths differ, and a user may
put a write proxy in front of ingest. One field would force the second backend
to work around it.

When `Managed` is true both are derived from the deployed service's name and
neither is stored, so there is one source of truth.

```go
type LoggingEndpoint struct {
    URL           string            `json:"url,omitempty"`
    Username      string            `json:"username,omitempty"`
    Password      EncryptedField    `json:"password,omitempty"`
    BearerToken   EncryptedField    `json:"bearerToken,omitempty"`
    Headers       map[string]string `json:"headers,omitempty"`
    TLSSkipVerify bool              `json:"tlsSkipVerify,omitempty"`
}
```

### Forwards

```go
type LoggingForward struct {
    Name     string           `json:"name"`
    Format   string           `json:"format,omitempty"` // jsonline
    Endpoint LoggingEndpoint  `json:"endpoint"`
}
```

A forward is a write-only copy. The expected configuration is not "HivePaaS's
backend *or* the user's" but both at once: VictoriaLogs powering the dashboard,
and a copy going to the company's central system. vlagent supports this
directly - see the fan-out finding below - so the cost is schema and UI only.

### Reference integrity

```go
func (s *Logging) GetRefObjectIDs() *RefObjectIDs {
    ids := &RefObjectIDs{}
    if s.Backend.VictoriaLogs != nil && s.Backend.VictoriaLogs.VolumeID != "" {
        ids.RefSettingIDs = append(ids.RefSettingIDs, s.Backend.VictoriaLogs.VolumeID)
    }
    return ids
}
```

Unlike `ClusterVolume`, whose `GetRefObjectIDs` always returns an empty struct,
this setting genuinely references another. `VerifyingRefIDs` therefore does
something here: it refuses a volume id that does not exist, and the resource
link stops that volume being deleted while logging still uses it.

```go
type LoggingVictoriaLogs struct {
    Image     string            `json:"image,omitempty"`
    NodeID    string            `json:"nodeId,omitempty"`
    VolumeID  string            `json:"volumeId,omitempty"`
    Retention timeutil.Duration `json:"retention"`
}
```

Retention reaches VictoriaLogs as the `-retentionPeriod` flag, so changing it
restarts the service. The UI must say so rather than implying it takes effect
immediately.

## 3. Interfaces

```go
// Backend is the read path. Always required, whoever owns the backend.
type Backend interface {
    Query(ctx context.Context, req *QueryReq) (*QueryResp, error)
    Tail(ctx context.Context, req *TailReq) (Stream, error)
    Ping(ctx context.Context) error
}

// Collector turns "collect these sources, ship there" into concrete config.
type Collector interface {
    Configure(spec *CollectSpec) (*RuntimeSpec, error)
}

// Deployer is "describe how to run me". Implemented only when Managed.
type Deployer interface {
    RuntimeSpec() (*RuntimeSpec, error)
}
```

`Deployer` is separate from `Backend` because a backend the user runs needs only
`Backend`. `victorialogs` implements both; `vlagent` implements `Collector`.

```go
// RuntimeSpec describes a container to run, in terms no orchestrator owns.
type RuntimeSpec struct {
    Image     string
    Args      []string
    Env       map[string]string
    Files     map[string][]byte
    Mounts    []Mount
    Ports     []Port
    Resources Resources
}
```

This is what keeps docker out of `services/logging`: the package *describes*
what to run, and `loggingserviceimpl` translates `RuntimeSpec` into a
`swarm.ServiceSpec`, attaching placement constraints and volumes. Adding Loki
later means writing `services/logging/loki` and touching no app-layer code.

```go
type CollectSpec struct {
    Ingest   Endpoint
    Forwards []ForwardTarget
    Sources  []Source
}

// ForwardTarget is a write-only copy of the stream. Format is what the target
// accepts, not what the collector prefers, so it travels with the endpoint.
type ForwardTarget struct {
    Name     string
    Format   string
    Endpoint Endpoint
}

type SourceKind string

const (
    SourceKindApp           SourceKind = "app"
    SourceKindHivePaaS      SourceKind = "hivepaas"
    SourceKindTraefikAccess SourceKind = "traefik-access"
    SourceKindNode          SourceKind = "node"
)

type Source struct {
    Kind    SourceKind
    Glob    string
    Exclude string
    Labels  map[string]string
}
```

## 4. Errors

Errors are declared in the package that raises them, not in
`hperrors/constants.go`. `services/backup` and `services/git/*` already do this;
no app-layer package does, and logging extends the pattern there.

```
services/logging/loggingmodel/errors.go        ERR_LOGGING_*
hivepaas_app/service/loggingservice/errors.go  ERR_LOGGING_SVC_*
```

`loggingmodel` rather than `logging`, for the reason `backupmodel/errors.go`
gives: `logging` imports the implementations, so they cannot import it back.
`loggingservice` rather than `loggingserviceimpl`, so usecases can use the
errors without pulling in the implementation.

The two prefixes must differ. Both files feed one translation namespace and a
collision between them is a string clash the compiler cannot see.

Translations move to per-domain files: `errors.logging.en.toml`,
`errors.backup.en.toml`, and `validation_errors.en.toml` renamed to
`errors.validation.en.toml` for consistency. The loader already walks the
message directory recursively and parses every file it finds, so this needs no
code change, and `parsePath` reads the language from the second-to-last
dot-separated segment either way.

Each file opens with a comment naming the Go files it mirrors.

### Prerequisite: `tools/errcodelint`

Splitting the file removes a guarantee the single file was quietly providing,
so the lint has to replace it first. Measured against the vendored libraries:

- **Two ids repeated inside one file** are a TOML violation. BurntSushi returns
  `Key 'X' has already been defined`, and `translation.localizerMap` panics on
  that error during package initialisation. The process dies at startup - loud,
  immediate, impossible to miss.
- **The same id in two files** parses cleanly and silently overwrites, because
  `i18n.Bundle.AddMessages` is a plain map assignment. Which copy wins depends
  on directory walk order.

So one file makes duplicates a startup crash; several files make them a wrong
message nobody notices. That is a real regression, and it is the reason
`errcodelint` is a prerequisite rather than an improvement to add afterwards.

A new check under `tools/`, run from `make lint` beside `goroutinelint`:

1. every `hperrors.NewErr(_, "ERR_X")` in the repo has an entry in some message file
2. no id is defined in two different files (a repeat inside one file is already
   a TOML parse error, so the check only has to cover the cross-file case)
3. no message id is orphaned - present in a file with no Go declaration
4. every message file name parses to a valid language tag

Check 4 exists because `language.Make` returns an undefined tag for an
unparseable string rather than failing: a file named `logging.errors.toml`
would load every message under a language nothing resolves against, while the
build and the tests stay green.

**Ordering:** `errcodelint` lands first and passes. Only then are the message
files split, as a separate commit, so the lint is what proves the move dropped
and duplicated nothing.

## 5. Collection

Docker's `json-file` driver plus vlagent tailing the files those containers
write. `--log-opt labels=...` makes the daemon write a block of container labels
into every line, which is what identifies the app.

### Conflict: apps may already change their own log driver

`prepareUpdatingAppContainerLogDriver`
(`usecase/appsettingsuc/container_settings_update.go:216`) lets a user set
`taskSpec.LogDriver.Name` and `.Options` freely. An app switched to `none` or
`syslog` writes no json-file, so it silently disappears from logging.

The resolution is neither to seize the field nor to ignore the problem, but to
make the consequence visible:

- an app that sets no driver is managed - HivePaaS sets `json-file` and the
  required `labels` option, preserving any `max-size`/`max-file` the user set
- an app that deliberately sets another driver keeps it, and is listed as **not
  collected**, both on the logging settings page and as a badge on the app

Silently overriding the operator and silently dropping their app are equally
bad; what is fixable is that the exclusion is never a surprise.

### The collector must exclude itself

The logging stack's own containers write logs, which the collector then picks
up. Observed during verification: VictoriaLogs ingested the log lines of the
container receiving forwarded logs. With a chatty collector or a backend that
logs each ingest, this amplifies.

`-fileCollector.excludeGlob` must exclude the collector's and the backend's own
containers. This is required, not optional.

### Global mode is new ground

vlagent runs one task per node, which is `swarm.ServiceMode.Global`. Docker
supports it, but **HivePaaS has never used it** - every service it creates today
is `Replicated`. Code that assumes a replica count (status reporting, scaling,
placement application) has to be checked against a global service.

## 6. Read path

VictoriaLogs open-source single-node has no authentication and no tenancy.
Anything that can reach it can read every project's logs. Three constraints
follow:

1. **No published port.** It attaches to an internal overlay network; the
   HivePaaS API is its only client.
2. **The dashboard never queries it directly.** It calls HivePaaS, where
   per-project and per-app permissions already exist.
3. **No raw LogsQL from clients**, outside global/admin scope. The API takes
   structured parameters - app, time range, level, search text - and HivePaaS
   builds the query.

Point 3 is the one that is tempting to get wrong. Accepting a user's LogsQL and
prefixing a scope filter is not safe: LogsQL has boolean operators and pipes
that change how a prefix binds, the same shape of flaw as building SQL by
concatenation. Constructing the query from structured parameters has no such
surface.

## 7. Provenance

Verified experimentally rather than reasoned about; the results changed the
design. Each was run against real Docker (29.7.2), vlagent v1.52.0 and
VictoriaLogs.

### Storage and filtering cannot be spoofed

A container labelled `hivepaas.app.id=REAL-APP` printing
`{"hivepaas.app.id":"SPOOFED",...}` on stdout stores as:

```
_msg                        {"hivepaas.app.id":"SPOOFED","level":"info",...}
attrs.hivepaas.app.id       REAL-APP
attrs.hivepaas.project.id   REAL-PROJ
```

vlagent does **not** recursively parse the message. The app's JSON stays a
string in `_msg`, and the daemon's labels land under their own `attrs.` prefix.
Repeating the test with the app emitting the exact field name
`attrs.hivepaas.app.id` changed nothing. Filtering on the stored field returned
zero matches for the forged value.

### Query-time unpacking can be spoofed

```
| unpack_json from _msg                       -> attrs.hivepaas.app.id = SPOOFED-EXACT
| unpack_json from _msg result_prefix "app."  -> attrs.hivepaas.app.id = REAL-EVIL
                                                 app.attrs.hivepaas.app.id = SPOOFED-EXACT
```

**Rule: every `unpack_json` HivePaaS emits carries `result_prefix`.** Because
the read path builds queries server-side (section 6), this is enforceable in the
query builder - one place. The two decisions hold each other up.

### This property is vlagent's, not the pipeline's

Fluent Bit's default configuration sets `Merge_Log On`, which parses the message
content and merges it into the record - precisely the recursive parse vlagent
avoids. Any collector added later must demonstrate this property again rather
than inheriting the conclusion. It belongs in the `Collector` contract and in
the test suite for each implementation.

## 8. Why vlagent

Not throughput. The published benchmark showing vlagent at 143,000 logs/s
against Fluent Bit's 31,300 is not a fair comparison: it was run by
VictoriaMetrics, and Fluent Bit's default Helm chart it used enables the
`kubernetes` metadata filter, `Merge_Log On`, a `docker, cri` multiline parser,
a second systemd input, and a 5MB `Mem_Buf_Limit` that throttles throughput by
construction. Fluent Bit was doing strictly more work, and HivePaaS is not
Kubernetes, so none of that enrichment would apply here anyway.

The reasons that survive:

1. **Same vendor as the backend** - one protocol, one set of documentation.
2. **Lower memory floor** (27.9 MiB against 78.1 MiB in that benchmark), which
   matters because the collector runs on *every* node including small ones.
3. **Safe by default** - the no-recursive-parse property above is vlagent's
   default behaviour, where Fluent Bit requires turning `Merge_Log` off and
   proving it again.

Reason 3 is the strongest. The choice is not permanent: `Collector` is its own
interface precisely so a second one can be added without touching app code.

## 9. Forwarding to systems HivePaaS does not own

vlagent is not locked to VictoriaLogs. Verified: one vlagent, two destinations,
per-destination format and credentials.

```
-remoteWrite.url=http://victorialogs:9428/internal/insert  -remoteWrite.format=native
-remoteWrite.url=http://elsewhere/ingest                   -remoteWrite.format=jsonline
                                                           -remoteWrite.bearerToken=...
```

The second destination received newline-delimited JSON, the correct
`Authorization: Bearer` header, and `attrs.hivepaas.app.id` intact, while the
first kept receiving normally. `basicAuth.*`, `headers`, and `tlsCertFile` are
also position-matched per URL.

**What this covers:** any endpoint accepting NDJSON over HTTP. Better Stack and
Axiom both document exactly that with a bearer token. Vector and Fluent Bit
aggregators accept it, which is the documented route to anything else.

**What it does not cover:** backends with their own schema - Loki's push API,
Elasticsearch's `_bulk` with interleaved action lines, Splunk HEC's envelope.
For these, the documented answer for v1 is to forward into the user's existing
Vector or Fluent Bit aggregator and let it speak the protocol.

**Caveat to verify per target:** vlagent sends `Content-Type:
application/octet-stream` with `Content-Encoding: zstd`. Services expecting
`application/x-ndjson`, or not accepting zstd, need the content type overridden
through `-remoteWrite.headers` or compression disabled. This is not
plug-and-play and the UI should offer a connection test rather than implying it
is.

## 10. Deployment

```
On:   volume (if the backend is managed) -> VictoriaLogs, pinned -> Ping -> vlagent, global
Off:  vlagent -> VictoriaLogs -> the volume is KEPT
```

The collector starts last and stops first: it needs an ingest endpoint that
exists, and should not ship into a dead backend.

The data volume survives turning logging off, and survives switching the backend
from managed to unmanaged. Deleting it is a separate, explicit action. This
matters concretely: a plain local volume holds its data inside the volume, so
removing it destroys the logs, whereas a bind volume only loses a pointer.

## Testing

- The provenance properties in section 7 are regression tests, not one-off
  checks: storage keeps the daemon's label under a forged one, filtering ignores
  the forged value, and a query built by HivePaaS always carries `result_prefix`.
- The query builder is tested against attempts to widen scope through its
  structured parameters.
- `RuntimeSpec` translation is tested without docker, from spec to
  `swarm.ServiceSpec`.
- Each `Collector` implementation carries the section 7 tests for itself.
- `errcodelint` passes before the message files are split, and again after.

## Prerequisites

1. `tools/errcodelint`, wired into `make lint`
2. message files split per domain, `validation_errors.en.toml` renamed
3. global-mode swarm services reviewed against the code that assumes replicas
