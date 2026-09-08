package apikeyuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
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
		return nil, hperrors.Wrap(hperrors.ErrUserDemoUnauthorized)
	}
	if err := uc.authorizeAPIKeyCreate(ctx, auth); err != nil {
		return nil, hperrors.Wrap(err)
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
				return hperrors.Wrap(err)
			}

			// Written inside the creating transaction, so the key and the record
			// of it either both exist or neither does. A key nobody can account
			// for is what this is here to prevent, and it outlives the session
			// that made it by up to a year.
			return uc.AuditService.Record(ctx, db, &auditservice.Entry{
				Type:     base.AuditLogTypeAPIKeyCreate,
				Scope:    base.ObjectScopeUser,
				ObjectID: actingUser.ID,
				Source:   base.AuditLogSourceAPICreate,
				Result:   base.AuditLogResultAllowed,
				Auth:     auth,
				ResType:  base.ResourceTypeAPIKey,
				ResID:    pData.Setting.ID,
				ResName:  req.Name,
			})
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &apikeydto.CreateAPIKeyResp{
		Data: &apikeydto.APIKeyDataResp{
			ID:        resp.Data.ID,
			KeyID:     keyID,
			SecretKey: secretKey,
		},
	}, nil
}

// authorizeAPIKeyCreate checks the capability and records the attempt either way.
//
// The refusals are the half worth keeping: a key that was minted leaves a key
// behind to find, while an attempt that was turned down leaves nothing at all
// unless it is written down here.
func (uc *UC) authorizeAPIKeyCreate(ctx context.Context, auth *basedto.Auth) error {
	hasCap, capErr := uc.HasCapability(ctx, uc.DB, auth, base.ResourceCapAPIKeyCreate)
	if capErr != nil {
		return hperrors.Wrap(capErr)
	}
	if hasCap {
		return nil
	}

	err := uc.AuditService.Record(ctx, uc.DB, &auditservice.Entry{
		Type:     base.AuditLogTypeAPIKeyCreate,
		Scope:    base.ObjectScopeUser,
		ObjectID: auth.UserID(),
		Source:   base.AuditLogSourceAPICreate,
		Result:   base.AuditLogResultDenied,
		Auth:     auth,
		ResType:  base.ResourceTypeAPIKey,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return hperrors.Wrap(hperrors.ErrUserNotHavePermissionOnCreateAPIKey)
}
