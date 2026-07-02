package sshkeyuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sshkeyuc/sshkeydto"
)

func (uc *UC) DeleteSSHKey(
	ctx context.Context,
	auth *basedto.Auth,
	req *sshkeydto.DeleteSSHKeyReq,
) (*sshkeydto.DeleteSSHKeyResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &sshkeydto.DeleteSSHKeyResp{}, nil
}
