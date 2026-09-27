package cloudstorageuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// checkKeyAuth refuses a storage whose key auth is not one. The save has
// already checked the id names a setting its scope can see; this is that the
// setting is a key auth, not a basic auth or a secret that happens to be
// visible too.
func (uc *UC) checkKeyAuth(ctx context.Context, db database.IDB, storage *entity.CloudStorage) error {
	if storage.S3 == nil {
		return nil
	}
	_, err := uc.SettingRepo.GetByID(ctx, db, nil, base.SettingTypeKeyAuth, storage.S3.KeyAuth.ID, false)
	if errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(hperrors.ErrSettingNotFound).WithParam("ID", storage.S3.KeyAuth.ID).
			WithMsgLog("the cloud storage's key auth is not a key auth")
	}
	return hperrors.Wrap(err)
}
