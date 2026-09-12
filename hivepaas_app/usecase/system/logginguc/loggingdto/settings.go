// Package loggingdto is the logging configuration as the API reads and writes it.
//
// It is not entity.Logging on the wire. The entity's credentials are
// EncryptedField, which would marshal as ciphertext; here they are strings the
// response always masks and the request carries in the clear.
package loggingdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

type EndpointData struct {
	URL           string            `json:"url"`
	Username      string            `json:"username,omitempty"`
	Password      string            `json:"password,omitempty"`
	BearerToken   string            `json:"bearerToken,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
	TLSSkipVerify bool              `json:"tlsSkipVerify,omitempty"`
}

type SourcesData struct {
	Apps          bool `json:"apps"`
	HivePaaS      bool `json:"hivepaas"`
	TraefikAccess bool `json:"traefikAccess"`
	Nodes         bool `json:"nodes"`
}

type CollectorData struct {
	Type    string `json:"type"`
	Managed bool   `json:"managed"`
	Image   string `json:"image,omitempty"`
}

type VictoriaLogsData struct {
	Image               string            `json:"image,omitempty"`
	NodeID              string            `json:"nodeId"`
	VolumeID            string            `json:"volumeId"`
	Retention           timeutil.Duration `json:"retention"`
	MaxDiskUsagePercent int               `json:"maxDiskUsagePercent,omitempty"`
}

type BackendData struct {
	Type         string            `json:"type"`
	Managed      bool              `json:"managed"`
	Ingest       *EndpointData     `json:"ingest,omitempty"`
	Query        *EndpointData     `json:"query,omitempty"`
	VictoriaLogs *VictoriaLogsData `json:"victoriaLogs,omitempty"`
}

type ForwardData struct {
	Name     string       `json:"name"`
	Format   string       `json:"format,omitempty"`
	Endpoint EndpointData `json:"endpoint"`
}

type SettingsData struct {
	Enabled   bool          `json:"enabled"`
	Sources   SourcesData   `json:"sources"`
	Collector CollectorData `json:"collector"`
	Backend   BackendData   `json:"backend"`
	Forwards  []ForwardData `json:"forwards"`
}

type SettingsStatus struct {
	BackendReady bool              `json:"backendReady"`
	ExcludedApps []ExcludedAppData `json:"excludedApps"`
}

// ExcludedAppData is an app logging will not show, and why.
type ExcludedAppData struct {
	AppID  string `json:"appId"`
	Name   string `json:"name"`
	Reason string `json:"reason"`
	Driver string `json:"driver,omitempty"`
}

type GetSettingsReq struct{}

func NewGetSettingsReq() *GetSettingsReq { return &GetSettingsReq{} }

func (req *GetSettingsReq) Validate() hperrors.ValidationErrors { return nil }

type GetSettingsResp struct {
	Data *SettingsData `json:"data"`
	// SecretMasked says the credentials came back as the placeholder. Sending
	// the placeholder back in an update keeps what is stored.
	SecretMasked bool            `json:"secretMasked,omitempty"`
	Status       *SettingsStatus `json:"status"`
}

type UpdateSettingsReq struct {
	Data *SettingsData `json:"data"`
}

func NewUpdateSettingsReq() *UpdateSettingsReq { return &UpdateSettingsReq{} }

// Validate rejects a request missing its data, and any credential that looks
// like ciphertext already: EncryptedField.Set would store such a value verbatim
// and it could never be decrypted.
func (req *UpdateSettingsReq) Validate() hperrors.ValidationErrors {
	validators := []vld.Validator{
		vld.Must(req.Data != nil).OnError(
			vld.SetField("data", nil),
			vld.SetCustomKey("ERR_VLD_FIELD_REQUIRED"),
		),
	}
	if req.Data == nil {
		return hperrors.NewValidationErrors(vld.Validate(validators...))
	}
	addEndpoint := func(ep *EndpointData, path string) {
		if ep == nil {
			return
		}
		validators = append(validators, basedto.ValidatePlainSecret(&ep.Password, path+".password")...)
		validators = append(validators, basedto.ValidatePlainSecret(&ep.BearerToken, path+".bearerToken")...)
	}
	addEndpoint(req.Data.Backend.Ingest, "data.backend.ingest")
	addEndpoint(req.Data.Backend.Query, "data.backend.query")
	for i := range req.Data.Forwards {
		addEndpoint(&req.Data.Forwards[i].Endpoint, "data.forwards.endpoint")
	}
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateSettingsResp struct {
	Data         *SettingsData `json:"data"`
	SecretMasked bool          `json:"secretMasked,omitempty"`
}

// FromEntity renders the stored configuration for a response.
//
// Every credential that is set comes back as the placeholder, never in the
// clear. Showing one would mean going through the audited reveal path the other
// settings use, which this does not offer yet.
func FromEntity(l *entity.Logging) (data *SettingsData, secretMasked bool) {
	if l == nil {
		l = &entity.Logging{}
	}
	mask := func(f entity.EncryptedField) string {
		if f.IsEmpty() {
			return ""
		}
		secretMasked = true
		return basedto.MaskedSecret
	}
	endpoint := func(ep *entity.LoggingEndpoint) *EndpointData {
		if ep == nil {
			return nil
		}
		return &EndpointData{
			URL: ep.URL, Username: ep.Username,
			Password: mask(ep.Password), BearerToken: mask(ep.BearerToken),
			Headers: ep.Headers, TLSSkipVerify: ep.TLSSkipVerify,
		}
	}

	data = &SettingsData{
		Enabled: l.Enabled,
		Sources: SourcesData{
			Apps: l.Sources.Apps, HivePaaS: l.Sources.HivePaaS,
			TraefikAccess: l.Sources.TraefikAccess, Nodes: l.Sources.Nodes,
		},
		Collector: CollectorData{Type: string(l.Collector.Type), Managed: l.Collector.Managed},
		Backend: BackendData{
			Type: string(l.Backend.Type), Managed: l.Backend.Managed,
			Ingest: endpoint(l.Backend.Ingest), Query: endpoint(l.Backend.Query),
		},
		Forwards: make([]ForwardData, 0, len(l.Forwards)),
	}
	if l.Collector.Vlagent != nil {
		data.Collector.Image = l.Collector.Vlagent.Image
	}
	if vl := l.Backend.VictoriaLogs; vl != nil {
		data.Backend.VictoriaLogs = &VictoriaLogsData{
			Image: vl.Image, NodeID: vl.NodeID, VolumeID: vl.VolumeID,
			Retention: vl.Retention, MaxDiskUsagePercent: vl.MaxDiskUsagePercent,
		}
	}
	for i := range l.Forwards {
		f := &l.Forwards[i]
		data.Forwards = append(data.Forwards, ForwardData{
			Name: f.Name, Format: f.Format, Endpoint: *endpoint(&f.Endpoint),
		})
	}
	return data, secretMasked
}

// ToEntity turns a request into the configuration to store.
//
// A credential sent back as the placeholder keeps the stored value it stands
// for: the backend's by position, a forward's by name. The placeholder with
// nothing stored behind it is refused rather than saved as a literal password.
func ToEntity(d *SettingsData, current *entity.Logging) (*entity.Logging, error) {
	if current == nil {
		current = &entity.Logging{}
	}
	curForwards := make(map[string]*entity.LoggingEndpoint, len(current.Forwards))
	for i := range current.Forwards {
		curForwards[current.Forwards[i].Name] = &current.Forwards[i].Endpoint
	}

	out := &entity.Logging{
		Enabled: d.Enabled,
		Sources: entity.LoggingSources{
			Apps: d.Sources.Apps, HivePaaS: d.Sources.HivePaaS,
			TraefikAccess: d.Sources.TraefikAccess, Nodes: d.Sources.Nodes,
		},
		Collector: entity.LoggingCollector{
			Type: entity.LoggingCollectorType(d.Collector.Type), Managed: d.Collector.Managed,
		},
		Backend: entity.LoggingBackend{
			Type: entity.LoggingBackendType(d.Backend.Type), Managed: d.Backend.Managed,
		},
	}
	if d.Collector.Image != "" {
		out.Collector.Vlagent = &entity.LoggingCollectorVlagent{Image: d.Collector.Image}
	}
	if vl := d.Backend.VictoriaLogs; vl != nil {
		out.Backend.VictoriaLogs = &entity.LoggingVictoriaLogs{
			Image: vl.Image, NodeID: vl.NodeID, VolumeID: vl.VolumeID,
			Retention: vl.Retention, MaxDiskUsagePercent: vl.MaxDiskUsagePercent,
		}
	}

	var err error
	if out.Backend.Ingest, err = toEndpoint(d.Backend.Ingest, current.Backend.Ingest, "backend.ingest"); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if out.Backend.Query, err = toEndpoint(d.Backend.Query, current.Backend.Query, "backend.query"); err != nil {
		return nil, hperrors.Wrap(err)
	}
	for i := range d.Forwards {
		f := &d.Forwards[i]
		ep, err := toEndpoint(&f.Endpoint, curForwards[f.Name], "forwards."+f.Name)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		out.Forwards = append(out.Forwards, entity.LoggingForward{Name: f.Name, Format: f.Format, Endpoint: *ep})
	}
	return out, nil
}

func toEndpoint(ep *EndpointData, cur *entity.LoggingEndpoint, path string) (*entity.LoggingEndpoint, error) {
	if ep == nil {
		return nil, nil
	}
	secret := func(value string, stored func(*entity.LoggingEndpoint) entity.EncryptedField, name string,
	) (entity.EncryptedField, error) {
		if !basedto.IsMaskedSecret(value) {
			if value == "" {
				return entity.EncryptedField{}, nil
			}
			return entity.NewEncryptedField(value), nil
		}
		if cur != nil {
			if kept := stored(cur); !kept.IsEmpty() {
				return kept, nil
			}
		}
		return entity.EncryptedField{}, hperrors.NewArgumentInvalid(path + "." + name).
			WithExtraDetail("the masked placeholder was sent, but nothing is stored for it")
	}

	password, err := secret(ep.Password, func(e *entity.LoggingEndpoint) entity.EncryptedField { return e.Password },
		"password")
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	token, err := secret(ep.BearerToken, func(e *entity.LoggingEndpoint) entity.EncryptedField { return e.BearerToken },
		"bearerToken")
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &entity.LoggingEndpoint{
		URL: ep.URL, Username: ep.Username, Password: password, BearerToken: token,
		Headers: ep.Headers, TLSSkipVerify: ep.TLSSkipVerify,
	}, nil
}
