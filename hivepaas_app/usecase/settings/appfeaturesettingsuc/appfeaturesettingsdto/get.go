package appfeaturesettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetAppFeatureSettingsReq struct {
	settings.GetUniqueSettingReq
}

func NewGetAppFeatureSettingsReq() *GetAppFeatureSettingsReq {
	return &GetAppFeatureSettingsReq{}
}

func (req *GetAppFeatureSettingsReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppFeatureSettingsResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *AppFeatureSettingsResp `json:"data"`
}

type AppFeatureSettingsResp struct {
	*settings.BaseSettingResp
	TerminalSettings *AppFeatureTerminalSettingsResp `json:"terminalSettings"`
	LoggingSettings  *AppFeatureLoggingSettingsResp  `json:"loggingSettings"`
	SchedJobSettings *AppFeatureSchedJobSettingsResp `json:"schedJobSettings"`
	PreviewSettings  *AppFeaturePreviewSettingsResp  `json:"previewSettings"`
}

type AppFeatureTerminalSettingsResp struct {
	Enabled bool `json:"enabled"`
}

type AppFeatureLoggingSettingsResp struct {
	Enabled bool `json:"enabled"`
}

type AppFeatureSchedJobSettingsResp struct {
	Enabled bool `json:"enabled"`
}

type AppFeaturePreviewSettingsResp struct {
	Enabled       bool                  `json:"enabled"`
	CreationDelay timeutil.Duration     `json:"creationDelay,omitempty"`
	AppsToClone   []*appdto.AppBaseResp `json:"appsToClone,omitempty" copy:"-"`
	AutoCloneApps bool                  `json:"autoCloneApps,omitempty"`
}

type AppFeatureSettingsTransformInput struct {
	Setting    *entity.Setting
	RefObjects *entity.RefObjects
}

func TransformAppFeatureSettings(
	input *AppFeatureSettingsTransformInput,
) (resp *AppFeatureSettingsResp, err error) {
	config := input.Setting.MustAsAppFeatureSettings()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(input.Setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if input.RefObjects == nil {
		input.RefObjects = &entity.RefObjects{}
	}

	if resp.PreviewSettings != nil && config.PreviewSettings != nil {
		resp.PreviewSettings.AppsToClone = nil
		for _, appID := range config.PreviewSettings.AppsToClone {
			appResp := appdto.TransformAppBase(input.RefObjects.RefApps[appID.ID])
			if appResp == nil {
				appResp = appdto.NewMissingApp(appID.ID)
			}
			resp.PreviewSettings.AppsToClone = append(resp.PreviewSettings.AppsToClone, appResp)
		}
	}

	return resp, nil
}
