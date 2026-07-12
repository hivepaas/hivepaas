package sshkeyuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/sshutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sshkeyuc/sshkeydto"
)

func (uc *UC) CreateSSHKey(
	ctx context.Context,
	auth *basedto.Auth,
	req *sshkeydto.CreateSSHKeyReq,
) (*sshkeydto.CreateSSHKeyResp, error) {
	req.Type = currentSettingType
	sshKey := req.ToEntity()
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: sshKey.GetRefObjectIDs(),
		Version:         currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			if err := generateKey(sshKey); err != nil {
				return apperrors.Wrap(err)
			}
			pData.Setting.Kind = string(req.Kind)
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
		privateKey, err := sshKey.PrivateKey.GetPlain()
		if err != nil {
			return apperrors.Wrap(err)
		}
		passphrase, err := sshKey.Passphrase.GetPlain()
		if err != nil {
			return apperrors.Wrap(err)
		}

		keyType, pubKey, err := sshutil.GeneratePublicKey(privateKey, passphrase)
		if err != nil {
			return apperrors.Wrap(err)
		}
		sshKey.PublicKey = pubKey
		sshKey.KeyType = gofn.Coalesce(keyType, sshKey.KeyType)
	}
	return nil
}
