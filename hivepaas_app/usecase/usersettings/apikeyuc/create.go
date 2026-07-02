package apikeyuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/usersettings/apikeyuc/apikeydto"
)

const (
	keyLen    = 16
	secretLen = 32
)

func (uc *UC) CreateAPIKey(
	ctx context.Context,
	auth *basedto.Auth,
	req *apikeydto.CreateAPIKeyReq,
) (*apikeydto.CreateAPIKeyResp, error) {
	if auth.User.IsDemoUser() {
		return nil, apperrors.New(apperrors.ErrUserDemoUnauthorized)
	}

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
				return apperrors.New(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &apikeydto.CreateAPIKeyResp{
		Data: &apikeydto.APIKeyDataResp{
			ID:        resp.Data.ID,
			KeyID:     keyID,
			SecretKey: secretKey,
		},
	}, nil
}
