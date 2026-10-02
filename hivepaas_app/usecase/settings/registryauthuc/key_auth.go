package registryauthuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// checkKeyAuth refuses an ECR credential whose key auth is not one. The save
// has already checked the id names a setting its scope can see; this is that
// the setting is a key auth, not a basic auth or a secret that happens to be
// visible too.
func (uc *UC) checkKeyAuth(ctx context.Context, db database.IDB, regAuth *entity.RegistryAuth) error {
	if regAuth.ECR == nil {
		return nil
	}
	_, err := uc.keyAuthOf(ctx, db, regAuth)
	return err
}

// keyAuthOf is the key auth an ECR credential names.
func (uc *UC) keyAuthOf(ctx context.Context, db database.IDB, regAuth *entity.RegistryAuth) (*entity.Setting, error) {
	keyAuth, err := uc.SettingRepo.GetByID(ctx, db, nil, base.SettingTypeKeyAuth, regAuth.ECR.KeyAuth.ID, false)
	if errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(hperrors.ErrSettingNotFound).WithParam("Name", "key auth "+regAuth.ECR.KeyAuth.ID).
			WithMsgLog("the registry credential's key auth is not a key auth")
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return keyAuth, nil
}
