package commandpipeuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandpipeuc/commandpipedto"
)

func (uc *UC) ListCommandPipe(
	ctx context.Context,
	auth *basedto.Auth,
	req *commandpipedto.ListCommandPipeReq,
) (*commandpipedto.ListCommandPipeResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := commandpipedto.TransformCommandPipes(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &commandpipedto.ListCommandPipeResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
