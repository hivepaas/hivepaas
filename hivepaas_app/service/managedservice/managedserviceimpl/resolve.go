package managedserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/cloudstorageservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

func (s *service) ResolveS3(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	storage entity.ManagedStorage,
) (*managedmodel.S3Archive, error) {
	storageID := storage.CloudStorage.ID
	if storage.BackupRepo.ID != "" {
		setting, err := s.settingRepo.GetByID(ctx, db, scope, base.SettingTypeBackupRepo, storage.BackupRepo.ID, true)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		repo, err := setting.AsBackupRepo()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if repo.CloudStorage.ID == "" {
			return nil, hperrors.Wrap(hperrors.ErrStorageTypeUnsupported).WithParam("Type", "volume")
		}
		storageID = repo.CloudStorage.ID
	}

	storageSetting, err := s.settingRepo.GetByID(ctx, db, scope, base.SettingTypeCloudStorage, storageID, true)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	cloudStorage, err := storageSetting.AsCloudStorage()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	refs := entity.NewRefObjects()
	if cloudStorage.S3 != nil && cloudStorage.S3.KeyAuth.ID != "" {
		keySetting, keyErr := s.settingRepo.GetByID(ctx, db, scope, base.SettingTypeKeyAuth,
			cloudStorage.S3.KeyAuth.ID, true)
		if keyErr != nil {
			return nil, hperrors.Wrap(keyErr)
		}
		refs.RefSettings[keySetting.ID] = keySetting
	}
	cfg, err := cloudstorageservice.S3Config(ctx, db, storageSetting, refs)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	archive := &managedmodel.S3Archive{
		Endpoint: cfg.Endpoint, Region: cfg.Region, Bucket: cfg.Bucket, Prefix: storage.Prefix,
		AccessKey: cfg.AccessKeyID, SecretKey: cfg.SecretAccessKey,
		// An endpoint of its own is an S3-compatible store, most of which address
		// buckets by path; AWS's own is reached by its region alone.
		ForcePathStyle: cfg.Endpoint != "",
	}
	if storage.KeyAuth.ID != "" {
		keySetting, keyErr := s.settingRepo.GetByID(ctx, db, scope, base.SettingTypeKeyAuth, storage.KeyAuth.ID, true)
		if keyErr != nil {
			return nil, hperrors.Wrap(keyErr)
		}
		keyAuth, keyErr := keySetting.AsKeyAuth()
		if keyErr != nil {
			return nil, hperrors.Wrap(keyErr)
		}
		secret, keyErr := keyAuth.SecretKey.GetPlain()
		if keyErr != nil {
			return nil, hperrors.Wrap(keyErr)
		}
		archive.AccessKey, archive.SecretKey = keyAuth.KeyID, secret
	}
	return archive, nil
}
