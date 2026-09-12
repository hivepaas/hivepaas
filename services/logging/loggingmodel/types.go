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
