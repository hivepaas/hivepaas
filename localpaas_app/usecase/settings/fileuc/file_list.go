package fileuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/fileuc/filedto"
)

func (uc *UC) ListFile(
	ctx context.Context,
	auth *basedto.Auth,
	req *filedto.ListFileReq,
) (*filedto.ListFileResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{
		ExtraLoadOpts: []bunex.SelectQueryOption{
			bunex.SelectWhereIf(len(req.StorageTypes) > 0,
				"setting.data->>'storageType' IN (?)", bunex.In(req.StorageTypes)),
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := filedto.TransformFiles(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &filedto.ListFileResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
