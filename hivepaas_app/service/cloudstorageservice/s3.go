// Package cloudstorageservice turns a cloud storage setting into what reaches its
// bucket. A cloud storage names its bucket; the key it is reached with is a key
// auth setting it links, so every user of a storage goes through here to put the
// two together.
package cloudstorageservice

import (
	"context"
	"database/sql"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/services/aws/s3"
)

// S3Config is what an S3 client of a cloud storage needs: its bucket, region and
// endpoint, and the key of the key auth it links.
//
// refs may hold the key auth already - a caller that loaded the storage as a
// reference loads what it references too - and it is read from db otherwise. The
// key auth must be active: a disabled one is a storage that cannot be used, and
// saying so here beats an S3 refusal later.
func S3Config(
	ctx context.Context, db database.IDB, storageSetting *entity.Setting, refs *entity.RefObjects,
) (*s3.Config, error) {
	if storageSetting.Type != base.SettingTypeCloudStorage || storageSetting.Kind != string(base.CloudStorageKindS3) {
		return nil, hperrors.Wrap(hperrors.ErrSettingTypeUnsupported).WithParam("Name", storageSetting.Type)
	}
	storage, err := storageSetting.AsCloudStorage()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if storage.S3 == nil {
		return nil, hperrors.Wrap(hperrors.ErrStorageTypeUnsupported).WithParam("Type", storageSetting.Kind)
	}
	keyAuth, err := loadKeyAuth(ctx, db, storage.S3.KeyAuth.ID, refs)
	if err != nil {
		return nil, err
	}
	secretKey, err := keyAuth.SecretKey.GetPlain()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &s3.Config{
		AccessKeyID:     keyAuth.KeyID,
		SecretAccessKey: secretKey,
		Endpoint:        storage.S3.Endpoint,
		Region:          storage.S3.Region,
		Bucket:          storage.S3.Bucket,
	}, nil
}

// NewS3Client is the S3 client of a cloud storage: S3Config, then a client.
func NewS3Client(
	ctx context.Context, db database.IDB, storageSetting *entity.Setting, refs *entity.RefObjects,
) (*s3.Client, error) {
	cfg, err := S3Config(ctx, db, storageSetting, refs)
	if err != nil {
		return nil, err
	}
	client, err := s3.NewClient(ctx, cfg)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return client, nil
}

func loadKeyAuth(ctx context.Context, db database.IDB, id string, refs *entity.RefObjects) (*entity.KeyAuth, error) {
	if id == "" {
		return nil, hperrors.Wrap(hperrors.ErrSettingNotFound).WithParam("ID", "key auth").
			WithMsgLog("the cloud storage links no key auth")
	}
	var setting *entity.Setting
	if refs != nil {
		setting = refs.RefSettings[id]
	}
	if setting == nil {
		// The storage's own save checked the key auth is visible from its scope;
		// reading it by id alone is reading what that check allowed.
		setting = &entity.Setting{}
		err := db.NewSelect().Model(setting).
			Where("setting.id = ?", id).
			Where("setting.type = ?", base.SettingTypeKeyAuth).
			Scan(ctx)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, hperrors.Wrap(hperrors.ErrSettingNotFound).WithParam("ID", id)
		}
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	if setting.Type != base.SettingTypeKeyAuth || setting.IsDeleted() {
		return nil, hperrors.Wrap(hperrors.ErrSettingNotFound).WithParam("ID", id)
	}
	if setting.Status != base.SettingStatusActive {
		return nil, hperrors.Wrap(hperrors.ErrSettingNotFound).WithParam("ID", id).
			WithMsgLog("the key auth %s of the cloud storage is not active", setting.Name)
	}
	keyAuth, err := setting.AsKeyAuth()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return keyAuth, nil
}
