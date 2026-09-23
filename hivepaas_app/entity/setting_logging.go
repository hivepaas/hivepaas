package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

const (
	CurrentLoggingSettingsVersion = 1
)

var _ = registerSettingParser(base.SettingTypeLogging, &loggingSettingsParser{})

type loggingSettingsParser struct {
}

// New is both the never-configured default and the struct stored data is
// unmarshalled into, so every field defaulted here to a non-zero value must be
// written unconditionally - with `omitempty`, a stored false is left out of the
// JSON and the default below survives the unmarshal as a silent true.
func (s *loggingSettingsParser) New() SettingData {
	return &LoggingSettings{
		Collector: LoggingCollector{
			Type:    base.LoggingCollectorTypeVlagent,
			Managed: true,
		},
		Sources: LoggingSources{Apps: true},
		Backend: LoggingBackend{
			Type:    base.LoggingBackendTypeVictoriaLogs,
			Managed: true,
		},
	}
}

// LoggingSettings is the whole subsystem's configuration. It is global and, until
// Enabled is set, nothing is deployed.
type LoggingSettings struct {
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
	// Apps defaults to true in New, so it carries no omitempty.
	Apps          bool `json:"apps"`
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
	Type base.LoggingBackendType `json:"type,omitempty"`
	// Managed defaults to true in New, so it carries no omitempty.
	Managed bool `json:"managed"`

	// Ingest and Query are separate because they are separate in backends other
	// than VictoriaLogs, and because a write proxy may sit in front of ingest.
	// Both are derived from the deployed service when Managed, and ignored.
	Ingest *LoggingEndpoint `json:"ingest,omitempty"`
	Query  *LoggingEndpoint `json:"query,omitempty"`

	VictoriaLogs *LoggingVictoriaLogs `json:"victoriaLogs,omitempty"`
}

type LoggingCollector struct {
	Type base.LoggingCollectorType `json:"type,omitempty"`
	// Managed defaults to true in New, so it carries no omitempty.
	Managed bool                     `json:"managed"`
	Vlagent *LoggingCollectorVlagent `json:"vlagent,omitempty"`
}

type LoggingCollectorVlagent struct {
}

type LoggingVictoriaLogs struct {
	// Volume decides both what the backend writes to and where it runs: the
	// placement constraint is derived from the volume's own pin, so the two can
	// never disagree. A volume pinned to nowhere leaves the backend unpinned
	// too - which on a cluster of more than one node means swarm may move it,
	// and a local volume on the new node is empty.
	Volume ObjectID `json:"volume,omitempty"`

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

	// CPULimit is in cores, the way app resource settings express it. Zero is
	// no cap, which is what running without this setting meant: a heavy query
	// could take whatever the node had, from the apps running beside it.
	CPULimit float64 `json:"cpuLimit,omitempty"`

	// MemoryLimit is written the way every other size in HivePaaS is - "1gb",
	// "512mb" - so the unit travels with the value and cannot be mistaken for
	// another. Zero takes logging.DefaultBackendMemoryLimit: the backend always
	// runs capped, which is what lets it be protected from the OOM killer.
	//
	// It is more than a ceiling: VictoriaLogs sizes its caches from the memory
	// it is allowed.
	MemoryLimit unit.DataSize `json:"memoryLimit,omitempty"`
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

func (s *LoggingSettings) GetType() base.SettingType {
	return base.SettingTypeLogging
}

// GetRefObjectIDs reports the data volume the managed backend writes to.
//
// Unlike most settings this one genuinely references another, so VerifyingRefIDs
// does something: it refuses a volume id that does not exist, and the resource
// link keeps that volume from being deleted while logging still uses it.
func (s *LoggingSettings) GetRefObjectIDs() *RefObjectIDs {
	ids := &RefObjectIDs{}
	if s.Backend.VictoriaLogs != nil && s.Backend.VictoriaLogs.Volume.ID != "" {
		ids.RefSettingIDs = append(ids.RefSettingIDs, s.Backend.VictoriaLogs.Volume.ID)
	}
	return ids
}

func (s *LoggingSettings) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

// Decrypt resolves every stored credential, so that a caller can read them.
func (s *LoggingSettings) Decrypt() error {
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
func (s *LoggingSettings) allEndpoints() []*LoggingEndpoint {
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

func (s *Setting) AsLoggingSettings() (*LoggingSettings, error) {
	return parseSettingAs[*LoggingSettings](s)
}

func (s *Setting) MustAsLoggingSettings() *LoggingSettings {
	return gofn.Must(s.AsLoggingSettings())
}
