package commandpipeuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandpipeuc/commandpipedto"
)

func (uc *UC) DeleteCommandPipe(
	ctx context.Context,
	auth *basedto.Auth,
	req *commandpipedto.DeleteCommandPipeReq,
) (*commandpipedto.DeleteCommandPipeResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &commandpipedto.DeleteCommandPipeResp{}, nil
}
