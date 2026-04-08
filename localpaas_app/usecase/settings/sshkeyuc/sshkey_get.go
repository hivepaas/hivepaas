package sshkeyuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/sshkeyuc/sshkeydto"
)

func (uc *UC) GetSSHKey(
	ctx context.Context,
	auth *basedto.Auth,
	req *sshkeydto.GetSSHKeyReq,
) (*sshkeydto.GetSSHKeyResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.Data.MustAsSSHKey().MustDecrypt()
	respData, err := sshkeydto.TransformSSHKey(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &sshkeydto.GetSSHKeyResp{
		Data: respData,
	}, nil
}
