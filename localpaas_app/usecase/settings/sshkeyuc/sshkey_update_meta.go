package sshkeyuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/sshkeyuc/sshkeydto"
)

func (uc *UC) UpdateSSHKeyMeta(
	ctx context.Context,
	auth *basedto.Auth,
	req *sshkeydto.UpdateSSHKeyMetaReq,
) (*sshkeydto.UpdateSSHKeyMetaResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingMeta(ctx, &req.UpdateSettingMetaReq, &settings.UpdateSettingMetaData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &sshkeydto.UpdateSSHKeyMetaResp{}, nil
}
