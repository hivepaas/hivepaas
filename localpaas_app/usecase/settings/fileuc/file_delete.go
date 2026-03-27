package fileuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/fileuc/filedto"
)

func (uc *FileUC) DeleteFile(
	ctx context.Context,
	auth *basedto.Auth,
	req *filedto.DeleteFileReq,
) (*filedto.DeleteFileResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{
		AfterPersisting: func(
			ctx context.Context,
			tx database.Tx,
			data *settings.DeleteSettingData,
			pData *settings.PersistingSettingDeletionData,
		) error {
			// TODO: add implementation
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &filedto.DeleteFileResp{}, nil
}
