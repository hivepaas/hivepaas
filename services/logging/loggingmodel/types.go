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
