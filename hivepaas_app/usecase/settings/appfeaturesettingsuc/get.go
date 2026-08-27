package appfeaturesettingsuc

import (
	"context"
	"errors"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/appfeaturesettingsuc/appfeaturesettingsdto"
)

func (uc *UC) GetAppFeatureSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appfeaturesettingsdto.GetAppFeatureSettingsReq,
) (*appfeaturesettingsdto.GetAppFeatureSettingsResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetUniqueSetting(ctx, auth, &req.GetUniqueSettingReq, &settings.GetUniqueSettingData{})
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}

	// If setting not found or a field of the settings not found, init default value for it
	if resp == nil || resp.Data == nil {
		timeNow := time.Now()
		resp = &settings.GetUniqueSettingResp{
			Data: &entity.Setting{
				ID:          gofn.Must(ulid.NewStringULID()),
				Scope:       req.Scope.ScopeType,
				Type:        base.SettingTypeAppFeatures,
				Status:      base.SettingStatusActive,
				ObjectID:    req.Scope.ScopeObjectID(),
				Inheritable: true,
				CreatedAt:   timeNow,
				UpdatedAt:   timeNow,
			},
		}
		featureSettings := &entity.AppFeatureSettings{}
		entity.InitAppFeatureSettingsDefault(featureSettings)
		resp.Data.MustSetData(featureSettings)
	}

	setting := resp.Data
	refObjects := entity.NewRefObjects()
	err = uc.SettingService.LoadRefObjectsSkipMissing(ctx, uc.DB, &refObjects, req.Scope, false, setting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	input := &appfeaturesettingsdto.AppFeatureSettingsTransformInput{
		Setting:    setting,
		RefObjects: refObjects,
	}

	respData, err := appfeaturesettingsdto.TransformAppFeatureSettings(input)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appfeaturesettingsdto.GetAppFeatureSettingsResp{
		Data: respData,
	}, nil
}
