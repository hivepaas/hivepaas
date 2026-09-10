package sshkeyuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sshkeyuc/sshkeydto"
)

func (uc *UC) UpdateSSHKey(
	ctx context.Context,
	auth *basedto.Auth,
	req *sshkeydto.UpdateSSHKeyReq,
) (*sshkeydto.UpdateSSHKeyResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	sshKey := req.ToEntity()
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: sshKey.GetRefObjectIDs(),
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			current, err := data.Setting.AsSSHKey()
			if err != nil {
				return hperrors.Wrap(err)
			}
			// Restored before generateKey, which derives the public key from the
			// private one and would fail outright on the placeholder.
			req.KeepMaskedSecrets(sshKey, current)

			if err = generateKey(sshKey); err != nil {
				return hperrors.Wrap(err)
			}
			pData.Setting.Kind = gofn.Coalesce(string(req.Kind), pData.Setting.Kind)
			if err = pData.Setting.SetData(sshKey); err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &sshkeydto.UpdateSSHKeyResp{}, nil
}
