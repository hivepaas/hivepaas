// Package loggingmodel holds the types and interfaces the logging
// implementations share.
//
// It deliberately depends on nothing from the application: no entity, no
// database, no docker. Configuration arrives as plain values - credentials
// included, already decrypted by the caller - so that an implementation can be
// tested without any of that machinery, and so that the application can change
// how it stores things without touching this package.
package loggingmodel

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

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
	Forwards []*ForwardTarget
	Sources  []*Source
}

// Mount is something a container needs mounted: a volume, or a host path.
type Mount struct {
	// Source is a host path, bind-mounted.
	Source   string
	Target   string
	ReadOnly bool
	// Volume is set instead of Source when the mount is a volume. It is the
	// volume as the caller refers to it - HivePaaS gives the id of a volume
	// setting - and the caller turns it into what the orchestrator mounts.
	Volume string
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

// The memory limits the logging stack runs under unless configured otherwise.
// Both services always get one: it is what lets them be protected from the OOM
// killer without being able to take a node's memory from the apps beside them.
const (
	// DefaultCollectorMemoryLimit caps the collector, which has no setting of
	// its own.
	DefaultCollectorMemoryLimit = 512 * unit.MB
	// DefaultBackendMemoryLimit is what the backend gets when its own setting
	// is left empty. VictoriaLogs sizes its caches from its limit, so this is
	// also what it sizes itself against.
	DefaultBackendMemoryLimit = 1 * unit.GB
)

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

	// Secrets are values the container reads from files, so that they never
	// appear in its arguments, which anyone who can read the service can read.
	Secrets []*Secret

	// PerNode says one container runs on every node rather than one somewhere:
	// a collector reads the files of the node it is on.
	PerNode bool
}

// Secret is a value written to a file in the container.
type Secret struct {
	// Key names the secret, unique within one RuntimeSpec.
	Key   string
	Path  string
	Value string
}

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
	// Search is the free-text part of the query.
	Search *TextSearch
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

// TextSearch is free text to match a message against.
//
// The mode decides more than what matches. Plain text is a token filter the
// backend answers from its index; a regular expression is read row by row, and
// measured about five times slower over the same data. So plain is the default,
// and a regular expression is something a person turns on knowing why.
type TextSearch struct {
	// Value is matched from the start of a token when plain, and as a regular
	// expression when IsRegex. Empty means no text filter at all.
	Value         string
	IsRegex       bool
	CaseSensitive bool
}

func (s *TextSearch) IsEmpty() bool { return s == nil || s.Value == "" }

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
	Entries []*LogEntry
	// Truncated says the limit was reached: older matching lines exist.
	Truncated bool
}

// InvocationStatsReq counts a function's invocation lines - the line its
// runtime writes for every call - over [Start, End), by Step. Match scopes it,
// as it does a QueryReq: there is no query text here either.
type InvocationStatsReq struct {
	Match []FieldMatch
	Start time.Time
	End   time.Time
	// Step is the buckets' width, in whole seconds.
	Step time.Duration
	// TopPaths is how many of the most called paths are counted apart.
	TopPaths int
}

// InvocationCounts are a set of calls: how many, how many failed - an outcome
// other than ok - how many the handler answered 4xx and 5xx, and how long the
// handler ran, in milliseconds. The durations are nil without a call.
type InvocationCounts struct {
	Calls     int64
	Failed    int64
	Errors4xx int64
	Errors5xx int64
	P50       *float64
	P95       *float64
	P99       *float64
}

// InvocationBucket is the calls of one step, Time its start.
type InvocationBucket struct {
	Time time.Time
	InvocationCounts
}

// InvocationPath is the calls of one method and path.
type InvocationPath struct {
	Method string
	Path   string
	InvocationCounts
}

// InvocationStatsResp is a range's calls: by step, oldest first, a step
// without a call left out; in all; by outcome; and by method and path, the most
// called first, as many as TopPaths.
type InvocationStatsResp struct {
	Buckets   []*InvocationBucket
	Totals    InvocationCounts
	ByOutcome map[string]int64
	ByPath    []*InvocationPath
}

// HTTPStatsReq counts the requests the proxy's access log records for one
// app: the lines in Match - the proxy's own, by the identity the daemon wrote -
// whose service matches ServicePattern, a regular expression.
type HTTPStatsReq struct {
	Match          []FieldMatch
	ServicePattern string
	Start          time.Time
	End            time.Time
	// Step is the buckets' width, in whole seconds.
	Step time.Duration
	// TopPaths and TopReplicas are how many of the most requested paths and
	// replicas are counted apart.
	TopPaths    int
	TopReplicas int
}

// HTTPCounts are a set of requests: how many, how many the client got a 4xx
// and a 5xx for, how many the proxy could not get to the app at all - a 502,
// 503 or 504 with no answer from it - and how long they took end to end, in
// milliseconds. The durations are nil without a request.
type HTTPCounts struct {
	Requests    int64
	Errors4xx   int64
	Errors5xx   int64
	Unreachable int64
	P50         *float64
	P95         *float64
	P99         *float64
}

// HTTPBucket is the requests of one step, Time its start.
type HTTPBucket struct {
	Time time.Time
	HTTPCounts
}

// HTTPPath is the requests of one method and path, the path's numbers and ids
// replaced by :n and :id.
type HTTPPath struct {
	Method string
	Path   string
	HTTPCounts
}

// HTTPReplica is the requests one replica answered, by its address; an empty
// one is the requests no replica answered.
type HTTPReplica struct {
	Address string
	HTTPCounts
}

// HTTPStatsResp is a range's requests: by step, oldest first, a step without a
// request left out; in all; by method and path and by replica, the most
// requested first.
type HTTPStatsResp struct {
	Buckets   []*HTTPBucket
	Totals    HTTPCounts
	ByPath    []*HTTPPath
	ByReplica []*HTTPReplica
}

// ResourceStatsReq reads the rows the agent writes of an app's containers'
// usage: the lines in Match - the agent's, by the identity the daemon wrote -
// whose app is AppID.
type ResourceStatsReq struct {
	Match []FieldMatch
	AppID string
	Start time.Time
	End   time.Time
	// Step is the buckets' width, in whole seconds.
	Step time.Duration
	// TopContainers is how many of the app's containers are listed apart, the
	// last seen first.
	TopContainers int
}

// ResourceUsage is what an app's containers used, summed over them: CPU in
// cores and memory in bytes, each with its limit (0 for none), OOM kills, and
// network and disk in bytes a second. CPU and Memory are nil for a step without
// a row.
type ResourceUsage struct {
	CPU         *float64
	CPULimit    float64
	Memory      *float64
	MemoryLimit float64
	OOMKills    int64
	NetRx       float64
	NetTx       float64
	IORead      float64
	IOWrite     float64
}

// ResourceBucket is one step's usage, Time its start: each container's
// average over the step, its memory's peak, summed over the containers.
type ResourceBucket struct {
	Time time.Time
	ResourceUsage
}

// ResourceContainer is one container over the range: its CPU's average and
// peak, its memory's peak, its OOM kills, and when it was last seen.
type ResourceContainer struct {
	Container   string
	CPU         float64
	CPUPeak     float64
	Memory      float64
	MemoryLimit float64
	OOMKills    int64
	LastSeen    time.Time
}

// ResourceStatsResp is a range's usage: by step, oldest first, a step without
// a row left out; and by container.
type ResourceStatsResp struct {
	Buckets    []*ResourceBucket
	Containers []*ResourceContainer
}

// InvocationLoadReq asks how busy functions were over [Start, End): their
// invocation lines, by the app identity in Field the daemon wrote into them.
type InvocationLoadReq struct {
	Field  string
	AppIDs []string
	Start  time.Time
	End    time.Time
}

// InvocationLoad is one function's calls that ended in the range: the time
// they took, summed, in milliseconds - divided by the range, the calls it had
// in flight on average - how many, and how many it turned away for having
// Concurrency calls already ("throttled").
type InvocationLoad struct {
	BusyMs    float64
	Calls     int64
	Throttled int64
}

// InvocationLoadResp is each function's load, by app id; one with no call in
// the range is not in it.
type InvocationLoadResp struct {
	ByApp map[string]*InvocationLoad
}
