package imagescanuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/imagescanuc/imagescandto"
)

// GetImageScanSettings returns the stored configuration, or the default when
// nothing was saved, together with what runs.
func (uc *UC) GetImageScanSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *imagescandto.GetImageScanSettingsReq,
) (*imagescandto.GetImageScanSettingsResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetUniqueSettingOrEmpty(ctx, auth, &req.GetUniqueSettingReq, &settings.GetUniqueSettingData{})
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	status, err := uc.imageScanService.Status(ctx, uc.DB)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data, err := imagescandto.TransformImageScanSettings(resp.Data, resp.RefObjects, status)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &imagescandto.GetImageScanSettingsResp{Data: data}, nil
}
