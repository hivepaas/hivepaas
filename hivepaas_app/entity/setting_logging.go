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

	// VolumeSubpath is the directory inside the volume the store keeps its data
	// in, so one volume can serve more than logging. Empty means the volume's
	// root, which is where the store wrote before this existed - changing it on
	// a running backend points the store at a new, empty directory, and the
	// logs collected so far stay where they were.
	VolumeSubpath string `json:"volumeSubpath,omitempty"`

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
	for _, ep := range s.allEndpoints() {
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
	eps := make([]*LoggingEndpoint, 0, len(s.Forwards)+2) //nolint:mnd // Ingest and Query
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
