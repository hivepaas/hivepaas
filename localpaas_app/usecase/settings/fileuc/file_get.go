package fileuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/fileuc/filedto"
)

const (
	currentSettingType = base.SettingTypeFile
)

func (uc *UC) GetFile(
	ctx context.Context,
	auth *basedto.Auth,
	req *filedto.GetFileReq,
) (*filedto.GetFileResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := filedto.TransformFile(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &filedto.GetFileResp{
		Data: respData,
	}, nil
}
