package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetAppKindSettingsReq struct {
	ProjectID     string `json:"-"`
	ProjectEnvID  string `json:"-"`
	AppID         string `json:"-"`
	RevealSecrets bool   `json:"-" mapstructure:"revealSecrets"`
}

func NewGetAppKindSettingsReq() *GetAppKindSettingsReq {
	return &GetAppKindSettingsReq{}
}

func (req *GetAppKindSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppKindSettingsResp struct {
	Meta *basedto.Meta        `json:"meta"`
	Data *AppKindSettingsResp `json:"data"`
}

type AppKindSettingsResp struct {
	Category base.AppCategory `json:"category"`
	Engine   string           `json:"engine"`
	Port     uint             `json:"port"`
	Version  string           `json:"version"`

	Webapp   *AppKindWebappResp   `json:"webapp,omitempty"`
	Database *AppKindDatabaseResp `json:"database,omitempty"`
	Cache    *AppKindCacheResp    `json:"cache,omitempty"`

	SecretMasked bool `json:"secretMasked"`
	UpdateVer    int  `json:"updateVer"`
}

type AppKindWebappResp struct {
}

type AppKindDatabaseResp struct {
	DbName         string                    `json:"dbName,omitempty"`
	Username       string                    `json:"username,omitempty"`
	Password       string                    `json:"password,omitzero" copy:"-"`
	RootPassword   string                    `json:"rootPassword,omitzero" copy:"-"`
	SSLMode        base.DatabaseSSLMode      `json:"sslMode,omitempty"`
	SSLCert        *settings.BaseSettingResp `json:"sslCert,omitempty"`
	TLSPassthrough bool                      `json:"tlsPassthrough,omitempty"`
}

func (resp *AppKindDatabaseResp) CopyPassword(field entity.EncryptedField) error {
	resp.Password = field.String()
	return nil
}

func (resp *AppKindDatabaseResp) CopyRootPassword(field entity.EncryptedField) error {
	resp.RootPassword = field.String()
	return nil
}

type AppKindCacheResp struct {
	Password        string                    `json:"password,omitempty" copy:"-"`
	MaxMemory       unit.DataSize             `json:"maxMemory,omitempty" swaggertype:"string"`
	EvictionRule    string                    `json:"evictionRule,omitempty"`
	PersistenceMode string                    `json:"persistenceMode,omitempty"`
	SSLCert         *settings.BaseSettingResp `json:"sslCert,omitempty"`
}

func (resp *AppKindCacheResp) CopyPassword(field entity.EncryptedField) error {
	resp.Password = field.String()
	return nil
}

type AppKindSettingsTransformInput struct {
	App            *entity.App
	KindSetting    *entity.Setting
	RoutingSetting *entity.Setting
	RefObjects     *entity.RefObjects
	MaskSecrets    bool
}

func TransformAppKindSettings(
	input *AppKindSettingsTransformInput,
) (resp *AppKindSettingsResp, err error) {
	resp = &AppKindSettingsResp{
		SecretMasked: input.MaskSecrets,
	}

	if err = copier.Copy(&resp, input.KindSetting); err != nil {
		return nil, hperrors.Wrap(err)
	}
	kindSettings := input.KindSetting.MustAsAppKindSettings()
	if err = copier.Copy(&resp, kindSettings); err != nil {
		return nil, hperrors.Wrap(err)
	}

	if input.RoutingSetting != nil {
		routingSettings := input.RoutingSetting.MustAsAppRoutingSettings()
		if routingSettings.Port > 0 && routingSettings.Port <= portMax {
			resp.Port = uint(routingSettings.Port)
		}
	}

	err = TransformAppKindDatabase(kindSettings, input, resp)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = TransformAppKindCache(kindSettings, input, resp)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return resp, nil
}

func TransformAppKindDatabase(
	kindSettings *entity.AppKindSettings,
	input *AppKindSettingsTransformInput,
	resp *AppKindSettingsResp,
) (err error) {
	if kindSettings.Database == nil {
		resp.Database = nil
		return nil
	}

	dbResp := resp.Database
	refObjects := input.RefObjects

	if input.RoutingSetting != nil {
		routingSettings := input.RoutingSetting.MustAsAppRoutingSettings()
		activeDomain, _ := gofn.First(routingSettings.GetActiveDomains())
		if activeDomain != nil && activeDomain.SSLCert.ID != "" {
			itemResp, _ := settings.TransformSettingBase(refObjects.RefSettings[activeDomain.SSLCert.ID])
			if itemResp == nil {
				itemResp = settings.NewMissingSetting(activeDomain.SSLCert.ID, base.SettingTypeSSLCert)
			}
			dbResp.SSLCert = itemResp
		}
	}

	if input.MaskSecrets {
		dbResp.Password = basedto.MaskedSecret
		dbResp.RootPassword = basedto.MaskedSecret
	}

	return nil
}

func TransformAppKindCache(
	kindSettings *entity.AppKindSettings,
	input *AppKindSettingsTransformInput,
	resp *AppKindSettingsResp,
) (err error) {
	if kindSettings.Cache == nil {
		resp.Cache = nil
		return nil
	}

	cacheResp := resp.Cache
	refObjects := input.RefObjects

	if input.RoutingSetting != nil {
		routingSettings := input.RoutingSetting.MustAsAppRoutingSettings()
		activeDomain, _ := gofn.First(routingSettings.GetActiveDomains())
		if activeDomain != nil && activeDomain.SSLCert.ID != "" {
			itemResp, _ := settings.TransformSettingBase(refObjects.RefSettings[activeDomain.SSLCert.ID])
			if itemResp == nil {
				itemResp = settings.NewMissingSetting(activeDomain.SSLCert.ID, base.SettingTypeSSLCert)
			}
			cacheResp.SSLCert = itemResp
		}
	}

	if input.MaskSecrets {
		cacheResp.Password = basedto.MaskedSecret
	}

	return nil
}
