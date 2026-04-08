package apikeyuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/usersettings/apikeyuc/apikeydto"
)

const (
	keyLen    = 16
	secretLen = 32
)

const (
	currentSettingType    = base.SettingTypeAPIKey
	currentSettingVersion = entity.CurrentAPIKeyVersion
)

func (uc *UC) CreateAPIKey(
	ctx context.Context,
	auth *basedto.Auth,
	req *apikeydto.CreateAPIKeyReq,
) (*apikeydto.CreateAPIKeyResp, error) {
	actingUser := auth.User.User
	// Generate key and secret
	keyID, secretKey := gofn.RandTokenAsHex(keyLen), gofn.RandTokenAsHex(secretLen)

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
			pData.Setting.ObjectID = actingUser.ID
			pData.Setting.Kind = keyID
			pData.Setting.ExpireAt = req.ExpireAt
			err := pData.Setting.SetData(&entity.APIKey{
				KeyID:        keyID,
				SecretKey:    entity.NewHashField(secretKey),
				AccessAction: req.AccessAction,
			})
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &apikeydto.CreateAPIKeyResp{
		Data: &apikeydto.APIKeyDataResp{
			ID:        resp.Data.ID,
			KeyID:     keyID,
			SecretKey: secretKey,
		},
	}, nil
}
