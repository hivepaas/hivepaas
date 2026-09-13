package loggingdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetLoggingSettingsReq struct {
	settings.GetUniqueSettingReq
}

func NewGetLoggingSettingsReq() *GetLoggingSettingsReq {
	return &GetLoggingSettingsReq{}
}

func (req *GetLoggingSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetLoggingSettingsResp struct {
	Meta *basedto.Meta        `json:"meta"`
	Data *LoggingSettingsResp `json:"data"`
}

type LoggingSettingsResp struct {
	*settings.BaseSettingResp

	Enabled   bool           `json:"enabled"`
	Sources   SourcesResp    `json:"sources"`
	Collector CollectorResp  `json:"collector"`
	Backend   BackendResp    `json:"backend"`
	Forwards  []*ForwardResp `json:"forwards"`

	// SecretMasked says the credentials came back as the placeholder. Sending
	// the placeholder back in an update keeps what is stored.
	SecretMasked  bool               `json:"secretMasked,omitempty"`
	LoggingStatus *LoggingStatusResp `json:"loggingStatus"`
}

type SourcesResp struct {
	Apps          bool `json:"apps"`
	HivePaaS      bool `json:"hivepaas"`
	TraefikAccess bool `json:"traefikAccess"`
	Nodes         bool `json:"nodes"`
}

type CollectorResp struct {
	Type    base.LoggingCollectorType `json:"type"`
	Managed bool                      `json:"managed"`
}

type BackendResp struct {
	Type         base.LoggingBackendType `json:"type"`
	Managed      bool                    `json:"managed"`
	Ingest       *EndpointResp           `json:"ingest,omitempty"`
	Query        *EndpointResp           `json:"query,omitempty"`
	VictoriaLogs *VictoriaLogsResp       `json:"victoriaLogs,omitempty"`
}

type EndpointResp struct {
	URL           string            `json:"url"`
	Username      string            `json:"username,omitempty"`
	Password      string            `json:"password,omitempty"`
	BearerToken   string            `json:"bearerToken,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
	TLSSkipVerify bool              `json:"tlsSkipVerify,omitempty"`
}

func (resp *EndpointResp) CopyPassword(field entity.EncryptedField) error {
	resp.Password = field.String()
	return nil
}

func (resp *EndpointResp) CopyBearerToken(field entity.EncryptedField) error {
	resp.BearerToken = field.String()
	return nil
}

type VictoriaLogsResp struct {
	Node                *settings.BaseSettingResp `json:"node"`
	Volume              *settings.BaseSettingResp `json:"volume"`
	VolumeSubpath       string                    `json:"volumeSubpath,omitempty"`
	Retention           timeutil.Duration         `json:"retention"`
	MaxDiskUsagePercent int                       `json:"maxDiskUsagePercent,omitempty"`
}

type ForwardResp struct {
	Name     string       `json:"name"`
	Format   string       `json:"format,omitempty"`
	Endpoint EndpointResp `json:"endpoint"`
}

type LoggingStatusResp struct {
	CollectorReady bool `json:"collectorReady"`
	BackendReady   bool `json:"backendReady"`
}

type LoggingSettingsTransformationInput struct {
	LoggingSetting *entity.Setting
	LoggingStatus  *loggingservice.Status
	RefObjects     *entity.RefObjects
	MaskSecrets    bool
}

func TransformLoggingSettings(
	input *LoggingSettingsTransformationInput,
) (resp *LoggingSettingsResp, err error) {
	resp = &LoggingSettingsResp{}
	if input.LoggingSetting == nil {
		return resp, nil
	}

	resp.SecretMasked = input.MaskSecrets
	if input.RefObjects == nil {
		input.RefObjects = entity.NewRefObjects()
	}

	if err = copier.Copy(&resp, input.LoggingSetting); err != nil {
		return nil, hperrors.Wrap(err)
	}
	loggingSettings := input.LoggingSetting.MustAsLogging()
	if err = copier.Copy(&resp, loggingSettings); err != nil {
		return nil, hperrors.Wrap(err)
	}

	TransformLoggingBackend(input, loggingSettings, resp)
	TransformLoggingForwards(input, loggingSettings, resp)

	resp.LoggingStatus = &LoggingStatusResp{}
	if input.LoggingStatus != nil {
		resp.LoggingStatus.CollectorReady = input.LoggingStatus.CollectorReady
		resp.LoggingStatus.BackendReady = input.LoggingStatus.BackendReady
	}

	return resp, nil
}

func TransformLoggingBackend(
	input *LoggingSettingsTransformationInput,
	loggingSettings *entity.Logging,
	resp *LoggingSettingsResp,
) {
	if loggingSettings == nil {
		return
	}

	if input.MaskSecrets {
		if resp.Backend.Ingest != nil {
			resp.Backend.Ingest.Password = basedto.MaskedSecret
			resp.Backend.Ingest.BearerToken = basedto.MaskedSecret
		}
		if resp.Backend.Query != nil {
			resp.Backend.Query.Password = basedto.MaskedSecret
			resp.Backend.Query.BearerToken = basedto.MaskedSecret
		}
	}

	vlogs := loggingSettings.Backend.VictoriaLogs
	if vlogs != nil {
		vlogsResp := resp.Backend.VictoriaLogs
		if vlogs.Node.ID != "" {
			itemResp, _ := settings.TransformSettingBase(input.RefObjects.RefSettings[vlogs.Node.ID])
			if itemResp == nil {
				itemResp = settings.NewMissingSetting(vlogs.Node.ID, base.SettingTypeClusterNode)
			}
			vlogsResp.Node = itemResp
		}
		if vlogs.Volume.ID != "" {
			itemResp, _ := settings.TransformSettingBase(input.RefObjects.RefSettings[vlogs.Volume.ID])
			if itemResp == nil {
				itemResp = settings.NewMissingSetting(vlogs.Volume.ID, base.SettingTypeClusterVolume)
			}
			vlogsResp.Volume = itemResp
		}
	}
}

func TransformLoggingForwards(
	input *LoggingSettingsTransformationInput,
	loggingSettings *entity.Logging,
	resp *LoggingSettingsResp,
) {
	if loggingSettings == nil {
		return
	}

	for _, forward := range resp.Forwards {
		if input.MaskSecrets {
			forward.Endpoint.Password = basedto.MaskedSecret
			forward.Endpoint.BearerToken = basedto.MaskedSecret
		}
	}
}
