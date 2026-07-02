package sshkeyuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sshkeyuc/sshkeydto"
)

func (uc *UC) GetSSHKey(
	ctx context.Context,
	auth *basedto.Auth,
	req *sshkeydto.GetSSHKeyReq,
) (*sshkeydto.GetSSHKeyResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	setting := resp.Data
	if setting.ObjectID == setting.CurrentObjectID { // not return sensitive data if setting is inherited
		if err := setting.MustAsSSHKey().Decrypt(); err != nil {
			return nil, apperrors.New(err)
		}
	}

	respData, err := sshkeydto.TransformSSHKey(setting, resp.RefObjects)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &sshkeydto.GetSSHKeyResp{
		Data: respData,
	}, nil
}
