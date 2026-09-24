package loggingdto

import (
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/logging"
)

const (
	nameMaxLen = 100
	urlMaxLen  = 255

	// maxBackendCPULimit and maxBackendMemoryLimit are sanity bounds, not a
	// judgement about the right size: they exist so a typo cannot ask swarm for
	// a limit no node could ever satisfy, leaving the task unschedulable.
	maxBackendCPULimit    = 256
	maxBackendMemoryLimit = 4 * unit.TB
)

type UpdateLoggingSettingsReq struct {
	settings.UpdateUniqueSettingReq
	*UpdateSettingsBaseReq

	// RemoveApp and RemoveStorage are asked of this request, not stored by it.
	// Switching logging off - or handing the backend or the collector to a
	// system HivePaaS does not run - takes an app down, and the apps have no
	// screen of their own to be removed from, so the confirmation arrives here.
	RemoveApp bool `json:"removeApp,omitempty"`
	// RemoveStorage deletes the stored logs with the backend: its own directory
	// inside the volume, not the volume.
	RemoveStorage bool `json:"removeStorage,omitempty"`
}

// BackendResources is what the managed backend is to run under. It is not part
// of the stored settings: it goes to the backend's service, where the app's own
// resource screen reads and writes it too. An empty memory limit is the
// default one - the backend always runs capped, which is what lets it be
// protected from the OOM killer.
func (req *UpdateLoggingSettingsReq) BackendResources() *logging.Resources {
	if req.UpdateSettingsBaseReq == nil || !req.Backend.Managed || req.Backend.VictoriaLogs == nil {
		return nil
	}
	vl := req.Backend.VictoriaLogs
	memoryLimit := vl.MemoryLimit
	if memoryLimit == 0 {
		memoryLimit = logging.DefaultBackendMemoryLimit
	}
	return &logging.Resources{CPULimit: vl.CPULimit, MemoryLimit: memoryLimit.Bytes()}
}

type UpdateSettingsBaseReq struct {
	Enabled   bool          `json:"enabled"`
	Sources   SourcesReq    `json:"sources"`
	Collector CollectorReq  `json:"collector"`
	Backend   BackendReq    `json:"backend"`
	Forwards  []*ForwardReq `json:"forwards"`
}

func (req *UpdateSettingsBaseReq) ToEntity() *entity.LoggingSettings {
	if req == nil {
		return nil
	}
	return &entity.LoggingSettings{
		Enabled:   req.Enabled,
		Sources:   req.Sources.ToEntity(),
		Collector: req.Collector.ToEntity(),
		Backend:   req.Backend.ToEntity(),
		Forwards: gofn.MapSlice(req.Forwards, func(fw *ForwardReq) entity.LoggingForward {
			return *fw.ToEntity()
		}),
	}
}

func (req *UpdateSettingsBaseReq) KeepMaskedSecrets(newSettings, current *entity.LoggingSettings) {
	if newSettings == nil || current == nil {
		return
	}

	// Backend ingest
	if newSettings.Backend.Ingest != nil && current.Backend.Ingest != nil {
		if req.Backend.Ingest != nil && basedto.IsMaskedSecret(req.Backend.Ingest.Password) {
			newSettings.Backend.Ingest.Password = current.Backend.Ingest.Password
		}
		if req.Backend.Ingest != nil && basedto.IsMaskedSecret(req.Backend.Ingest.BearerToken) {
			newSettings.Backend.Ingest.BearerToken = current.Backend.Ingest.BearerToken
		}
	}
	// Backend query
	if newSettings.Backend.Query != nil && current.Backend.Query != nil {
		if req.Backend.Query != nil && basedto.IsMaskedSecret(req.Backend.Query.Password) {
			newSettings.Backend.Query.Password = current.Backend.Query.Password
		}
		if req.Backend.Query != nil && basedto.IsMaskedSecret(req.Backend.Query.BearerToken) {
			newSettings.Backend.Query.BearerToken = current.Backend.Query.BearerToken
		}
	}

	// Logging forwards
	curForwards := make(map[string]*entity.LoggingForward, len(current.Forwards))
	for i := range current.Forwards {
		curForwards[current.Forwards[i].Name] = &current.Forwards[i]
	}
	for i, reqForward := range req.Forwards {
		newForward := &newSettings.Forwards[i]
		currForward, ok := curForwards[newForward.Name]
		if !ok {
			continue
		}
		if basedto.IsMaskedSecret(reqForward.Endpoint.Password) {
			newForward.Endpoint.Password = currForward.Endpoint.Password
		}
		if basedto.IsMaskedSecret(reqForward.Endpoint.BearerToken) {
			newForward.Endpoint.BearerToken = currForward.Endpoint.BearerToken
		}
	}
}

func (req *UpdateSettingsBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, req.Backend.validate(field+"backend")...)
	for _, forward := range req.Forwards {
		res = append(res, forward.validate(field+"forward")...)
	}
	return res
}

type SourcesReq struct {
	Apps          bool `json:"apps"`
	HivePaaS      bool `json:"hivepaas"`
	TraefikAccess bool `json:"traefikAccess"`
	Nodes         bool `json:"nodes"`
}

func (req SourcesReq) ToEntity() entity.LoggingSources {
	return entity.LoggingSources{
		Apps:          req.Apps,
		HivePaaS:      req.HivePaaS,
		TraefikAccess: req.TraefikAccess,
		Nodes:         req.Nodes,
	}
}

type CollectorReq struct {
	Type    base.LoggingCollectorType `json:"type"`
	Managed bool                      `json:"managed"`
}

func (req CollectorReq) ToEntity() entity.LoggingCollector {
	return entity.LoggingCollector{
		Type:    req.Type,
		Managed: req.Managed,
	}
}

type BackendReq struct {
	Type         base.LoggingBackendType `json:"type"`
	Managed      bool                    `json:"managed"`
	Ingest       *EndpointReq            `json:"ingest,omitempty"`
	Query        *EndpointReq            `json:"query,omitempty"`
	VictoriaLogs *VictoriaLogsReq        `json:"victoriaLogs,omitempty"`
}

func (req BackendReq) ToEntity() entity.LoggingBackend {
	return entity.LoggingBackend{
		Type:         req.Type,
		Managed:      req.Managed,
		Ingest:       req.Ingest.ToEntity(),
		Query:        req.Query.ToEntity(),
		VictoriaLogs: req.VictoriaLogs.ToEntity(),
	}
}

func (req *BackendReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStrIn(&req.Type, true, base.AllLoggingBackendTypes, field+"type")...)
	res = append(res, req.Ingest.validate(field+"ingest")...)
	res = append(res, req.Query.validate(field+"query")...)

	res = append(res, req.VictoriaLogs.validate(field+"victoriaLogs")...)
	return res
}

type VictoriaLogsReq struct {
	Volume              basedto.ObjectIDReq `json:"volume"`
	Retention           timeutil.Duration   `json:"retention"`
	MaxDiskUsagePercent int                 `json:"maxDiskUsagePercent,omitempty"`
	// CPULimit is in cores, zero for no cap. MemoryLimit carries its own unit -
	// "1gb", "512mb" - and also accepts a bare number of bytes; zero is the
	// default limit. Neither is stored with the settings: see BackendResources.
	CPULimit    float64       `json:"cpuLimit,omitempty"`
	MemoryLimit unit.DataSize `json:"memoryLimit,omitempty" swaggertype:"string"`
}

func (req *VictoriaLogsReq) ToEntity() *entity.LoggingVictoriaLogs {
	if req == nil {
		return nil
	}
	return &entity.LoggingVictoriaLogs{
		Volume:              *req.Volume.ToEntity(),
		Retention:           req.Retention,
		MaxDiskUsagePercent: req.MaxDiskUsagePercent,
	}
}

func (req *VictoriaLogsReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateObjectIDReq(&req.Volume, true, field+"volume")...)
	res = append(res, basedto.ValidateNumber(&req.MaxDiskUsagePercent, false, 0, 100, //nolint:mnd
		field+"maxDiskUsagePercent")...)
	res = append(res, basedto.ValidateNumber(&req.CPULimit, false, 0, maxBackendCPULimit, field+"cpuLimit")...)
	res = append(res, basedto.ValidateNumber(&req.MemoryLimit, false, 0, maxBackendMemoryLimit,
		field+"memoryLimit")...)
	return res
}

type ForwardReq struct {
	Name     string       `json:"name"`
	Format   string       `json:"format,omitempty"`
	Endpoint *EndpointReq `json:"endpoint"`
}

func (req *ForwardReq) ToEntity() *entity.LoggingForward {
	if req == nil {
		return nil
	}
	return &entity.LoggingForward{
		Name:     req.Name,
		Format:   req.Format,
		Endpoint: *req.Endpoint.ToEntity(),
	}
}

func (req *ForwardReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, nameMaxLen, field+"name")...)
	res = append(res, basedto.ValidateStr(&req.Format, true, 1, nameMaxLen, field+"format")...)
	res = append(res, req.Endpoint.validate(field+"endpoint")...)
	return res
}

type EndpointReq struct {
	URL           string            `json:"url"`
	Username      string            `json:"username,omitempty"`
	Password      string            `json:"password,omitempty"`
	BearerToken   string            `json:"bearerToken,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
	TLSSkipVerify bool              `json:"tlsSkipVerify,omitempty"`
}

func (req *EndpointReq) ToEntity() *entity.LoggingEndpoint {
	if req == nil {
		return nil
	}
	return &entity.LoggingEndpoint{
		URL:           req.URL,
		Username:      req.Username,
		Password:      entity.NewEncryptedField(req.Password),
		BearerToken:   entity.NewEncryptedField(req.BearerToken),
		Headers:       req.Headers,
		TLSSkipVerify: req.TLSSkipVerify,
	}
}

func (req *EndpointReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.URL, true, 1, urlMaxLen, field+"url")...)
	res = append(res, basedto.ValidateStr(&req.Username, false, 1, nameMaxLen, field+"username")...)
	res = append(res, basedto.ValidatePlainSecret(&req.Password, field+"password")...)
	res = append(res, basedto.ValidatePlainSecret(&req.BearerToken, field+"bearerToken")...)
	return res
}

func NewUpdateLoggingSettingsReq() *UpdateLoggingSettingsReq {
	return &UpdateLoggingSettingsReq{}
}

func (req *UpdateLoggingSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateLoggingSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
