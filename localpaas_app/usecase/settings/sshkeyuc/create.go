package sshkeyuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/sshutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/sshkeyuc/sshkeydto"
)

func (uc *UC) CreateSSHKey(
	ctx context.Context,
	auth *basedto.Auth,
	req *sshkeydto.CreateSSHKeyReq,
) (*sshkeydto.CreateSSHKeyResp, error) {
	req.Type = currentSettingType
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName: req.Name,
		Version:       currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			sshKey := req.ToEntity()
			if err := generateKey(sshKey); err != nil {
				return apperrors.Wrap(err)
			}
			if err := pData.Setting.SetData(sshKey); err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &sshkeydto.CreateSSHKeyResp{
		Data: &basedto.ObjectIDResp{ID: resp.Data.ID},
	}, nil
}

func generateKey(sshKey *entity.SSHKey) error {
	if sshKey.PublicKey == "" {
		keyType, pubKey, err := sshutil.GeneratePublicKey(sshKey.PrivateKey.MustGetPlain(),
			sshKey.Passphrase.MustGetPlain())
		if err != nil {
			return apperrors.Wrap(err)
		}
		sshKey.PublicKey = pubKey
		sshKey.KeyType = gofn.Coalesce(keyType, sshKey.KeyType)
	}
	return nil
}
