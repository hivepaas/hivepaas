package registryserviceimpl

import (
	"context"
	"errors"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
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
//
// That holds across switching the registry off and on again. The credential
// outlives the registry for as long as an app names it, and the next registry
// takes it up again, whatever its address is now: making a second one would
// leave those apps pulling with a password the new registry does not know.
func (s *service) ensureCredential(ctx context.Context, db database.IDB, cfg *entity.RegistrySettings) (
	*entity.Setting, string, error) {
	timeNow := timeutil.NowUTC()

	setting, err := s.findOwnCredential(ctx, db, cfg.RegistryAuthID)
	if err != nil {
		return nil, "", err
	}
	if setting != nil {
		auth, err := setting.AsRegistryAuth()
		if err != nil {
			return nil, "", hperrors.Wrap(err)
		}
		password, err := auth.Password.GetPlain()
		if err != nil {
			return nil, "", hperrors.Wrap(err)
		}
		// A credential made before the marker existed gets it now, so it is
		// found again even once the registry has forgotten its id.
		stamp := auth.ManagedBy != entity.RegistryAuthManagedBySystemRegistry
		if auth.Address != cfg.Domain || stamp {
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

	password, err := generatePassword()
	if err != nil {
		return nil, "", hperrors.Wrap(err)
	}
	setting, err = newRegistryAuthSetting(gofn.Must(ulid.NewStringULID()), cfg.Domain, password, timeNow)
	if err != nil {
		return nil, "", hperrors.Wrap(err)
	}
	if err = s.settingRepo.Upsert(ctx, db, setting,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols); err != nil {
		return nil, "", hperrors.Wrap(err)
	}
	return setting, password, nil
}

// findOwnCredential is the credential the registry created for itself, or nil
// when it no longer exists.
//
// The id the registry remembers is asked first. The marker is for when that id
// is gone - a registry setting written before ids outlived a switch-off, or one
// reset by hand - and, among several marked, the one changed last is taken.
// Neither looks at the address or the name: both may have changed since.
func (s *service) findOwnCredential(ctx context.Context, db database.IDB, id string) (*entity.Setting, error) {
	if id != "" {
		setting, err := s.settingRepo.GetByID(ctx, db, entity.NewObjectScopeGlobal(),
			base.SettingTypeRegistryAuth, id, false)
		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return nil, hperrors.Wrap(err)
		}
		if setting != nil {
			return setting, nil
		}
		// Deleted by somebody, or with the last app that named it: looked for by
		// the marker below, and made anew when there is none.
	}

	settings, _, err := s.settingRepo.List(ctx, db, entity.NewObjectScopeGlobal(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeRegistryAuth),
		bunex.SelectOrder("setting.updated_at DESC"),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return pickOwnCredential(settings), nil
}

// pickOwnCredential is the credential marked as the registry's, the one changed
// last when there are several. A credential whose data cannot be read is
// passed over: it is nothing the registry could push with anyway.
func pickOwnCredential(settings []*entity.Setting) *entity.Setting {
	var own *entity.Setting
	for _, setting := range settings {
		if setting.Type != base.SettingTypeRegistryAuth || setting.Scope != base.ObjectScopeGlobal {
			continue
		}
		auth, err := setting.AsRegistryAuth()
		if err != nil || auth.ManagedBy != entity.RegistryAuthManagedBySystemRegistry {
			continue
		}
		if own == nil || setting.UpdatedAt.After(own.UpdatedAt) {
			own = setting
		}
	}
	return own
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
