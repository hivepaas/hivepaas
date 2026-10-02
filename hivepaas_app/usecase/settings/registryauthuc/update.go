package registryauthuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/registryauthuc/registryauthdto"
)

func (uc *UC) UpdateRegistryAuth(
	ctx context.Context,
	auth *basedto.Auth,
	req *registryauthdto.UpdateRegistryAuthReq,
) (*registryauthdto.UpdateRegistryAuthResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	regAuth := req.ToEntity()
	scheduleRenewal := func() {}
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: regAuth.GetRefObjectIDs(),
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			pData.Setting.Kind = req.Address
			current, err := data.Setting.AsRegistryAuth()
			if err != nil {
				return hperrors.Wrap(err)
			}
			req.KeepMaskedSecrets(regAuth, current)
			registryauthdto.KeepToken(regAuth, current)
			// Not the operator's to set or clear: an edit keeps what created it.
			regAuth.ManagedBy = current.ManagedBy

			err = pData.Setting.SetData(regAuth)
			if err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) (err error) {
			// New AWS keys: the services pulling with the credential get a
			// token from them now rather than at the next scheduled run.
			if regAuth.Kind != base.RegistryAuthKindAWSECR || !regAuth.Token.IsEmpty() ||
				pData.Setting.Status != base.SettingStatusActive {
				return nil
			}
			scheduleRenewal, err = uc.registryAuthRenewalUC.RenewOnSave(ctx, db, pData.Setting.ID)
			return hperrors.Wrap(err)
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	scheduleRenewal()

	return &registryauthdto.UpdateRegistryAuthResp{}, nil
}
