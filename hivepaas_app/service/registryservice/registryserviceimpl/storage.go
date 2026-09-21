package registryserviceimpl

import (
	"context"
	"errors"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

// ensureCredential returns the credential the registry is pushed to with, and the
// password in plaintext - which the htpasswd file needs and nothing else in the
// flow has.
//
// It creates the credential once. Every save after that keeps the password and
// only rewrites the address, because a save is not a rotation: changing the
// password under apps that are already using it is what RotateCredential does,
// deliberately and with a grace period.
func (s *service) ensureCredential(ctx context.Context, db database.IDB, cfg *entity.RegistrySettings) (
	*entity.Setting, string, error) {
	timeNow := timeutil.NowUTC()

	if cfg.RegistryAuthID != "" {
		setting, err := s.settingRepo.GetByID(ctx, db, entity.NewObjectScopeGlobal(),
			base.SettingTypeRegistryAuth, cfg.RegistryAuthID, false)
		if err == nil && setting != nil {
			auth, parseErr := setting.AsRegistryAuth()
			if parseErr != nil {
				return nil, "", hperrors.Wrap(parseErr)
			}
			password, plainErr := auth.Password.GetPlain()
			if plainErr != nil {
				return nil, "", hperrors.Wrap(plainErr)
			}
			if auth.Address != cfg.Domain {
				if err = updateRegistryAuthSetting(setting, cfg.Domain, "", timeNow); err != nil {
					return nil, "", hperrors.Wrap(err)
				}
				if err = s.settingRepo.Upsert(ctx, db, setting,
					entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols); err != nil {
					return nil, "", hperrors.Wrap(err)
				}
			}
			return setting, password, nil
		}
		// A credential somebody deleted is not a reason to refuse: one is made
		// below, and the setting is pointed at it.
	}

	password, err := generatePassword()
	if err != nil {
		return nil, "", hperrors.Wrap(err)
	}
	setting, err := newRegistryAuthSetting(gofn.Must(ulid.NewStringULID()), cfg.Domain, password, timeNow)
	if err != nil {
		return nil, "", hperrors.Wrap(err)
	}
	if err = s.settingRepo.Upsert(ctx, db, setting,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols); err != nil {
		return nil, "", hperrors.Wrap(err)
	}
	return setting, password, nil
}

// resolveStorage turns the setting's reference into what the configuration and
// the document need: a volume's name, or a bucket's address and credentials.
func (s *service) resolveStorage(ctx context.Context, db database.IDB, cfg *entity.RegistrySettings) (
	planInput, error) {
	if cfg.Storage.Type == base.RegistryStorageTypeS3 {
		s3, err := s.resolveCloudStorage(ctx, db, cfg.Storage.CloudStorage.ID)
		if err != nil {
			return planInput{}, hperrors.Wrap(err)
		}
		return planInput{S3: s3}, nil
	}

	setting, err := s.settingRepo.GetByID(ctx, db, entity.NewObjectScopeGlobal(),
		base.SettingTypeClusterVolume, cfg.Storage.Volume.ID, true)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return planInput{}, hperrors.Wrap(err)
	}
	if setting == nil {
		return planInput{}, hperrors.Wrap(hperrors.ErrRegistrySettingsInvalid).
			WithExtraDetail("The volume the registry keeps its images on no longer exists.")
	}
	// A volume reaches an app only when it is inheritable: that is the rule
	// applyAppFilter enforces, and the registry's app is in a project of its own.
	// Without this the failure lands inside BuildAppMounts as "Volume not found",
	// which says nothing about what to do.
	if !setting.Inheritable {
		return planInput{}, hperrors.Wrap(hperrors.ErrRegistrySettingsInvalid).WithExtraDetail(
			"The volume %q is not shared with apps. Edit it in Cluster > Volumes and make it "+
				"available to apps, or choose one that already is.", setting.Name)
	}
	// A mount names the volume by the setting's id; the load above is what turns
	// a missing or unusable volume into a sentence instead of a failure inside
	// the build.
	return planInput{VolumeID: setting.ID}, nil
}

func (s *service) resolveCloudStorage(ctx context.Context, db database.IDB, id string) (*zotS3Input, error) {
	setting, err := s.settingRepo.GetByID(ctx, db, entity.NewObjectScopeGlobal(),
		base.SettingTypeCloudStorage, id, true)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	if setting == nil {
		return nil, hperrors.Wrap(hperrors.ErrRegistrySettingsInvalid).
			WithExtraDetail("The cloud storage the registry keeps its images in no longer exists.")
	}

	storage, err := setting.AsCloudStorage()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if storage.S3 == nil || storage.S3.CloudProviderAWS == nil || storage.S3.Bucket == "" {
		return nil, hperrors.Wrap(hperrors.ErrRegistrySettingsInvalid).
			WithExtraDetail("The chosen cloud storage names no S3 bucket.")
	}
	secretKey, err := storage.S3.SecretKey.GetPlain()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &zotS3Input{
		Bucket:    storage.S3.Bucket,
		Region:    gofn.Coalesce(storage.S3.Region, storage.S3.CloudProviderAWS.Region),
		Endpoint:  storage.S3.Endpoint,
		AccessKey: storage.S3.AccessKeyID,
		SecretKey: secretKey,
		// Anything but a plain-HTTP endpoint speaks TLS, and a registry's bucket
		// is not somewhere to make that optional.
		Secure: true,
	}, nil
}
