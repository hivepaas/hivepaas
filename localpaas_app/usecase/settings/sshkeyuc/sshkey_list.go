package sshkeyuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/sshkeyuc/sshkeydto"
)

func (uc *UC) ListSSHKey(
	ctx context.Context,
	auth *basedto.Auth,
	req *sshkeydto.ListSSHKeyReq,
) (*sshkeydto.ListSSHKeyResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := sshkeydto.TransformSSHKeys(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &sshkeydto.ListSSHKeyResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
