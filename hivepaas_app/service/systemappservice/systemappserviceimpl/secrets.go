package systemappserviceimpl

import (
	"context"
	"maps"
	"reflect"
	"slices"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
)

// secretFileMode is readable by whatever user the image runs as, the mode the
// registry's files are mounted with too. The files hold credentials, but only
// the container they are mounted into can see them.
const secretFileMode = fileutil.FileMode(0o444)

// secretPart is the part of a secret its file holds.
const secretPart = "value"

// SyncSecrets makes the app's secrets exactly files: each a secret, mounted by a
// setting mount the way a template's swarmRef makes one, under the entry key
// its name makes. The mounts are refreshed once when anything changed, which is
// what puts the files in the container.
func (s *service) SyncSecrets(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	files []*systemappservice.SecretFile,
) error {
	secrets := settingsByName(app.GetSettingsByType(base.SettingTypeSecret))
	mounts := settingsByName(app.GetSettingsByType(base.SettingTypeAppSettingMount))

	timeNow := timeutil.NowUTC()
	changed := false
	wanted := make(map[string]bool, len(files))
	for _, file := range files {
		wanted[file.Key] = true
		secret, wrote, err := s.syncSecret(ctx, db, app, secrets[file.Key], file, timeNow)
		if err != nil {
			return hperrors.Wrap(err)
		}
		changed = changed || wrote
		key := settingmountservice.EntryKeyFor(file.Key)
		wrote, err = s.syncSecretMount(ctx, db, app, mounts[key], key, secret, file, timeNow)
		if err != nil {
			return hperrors.Wrap(err)
		}
		changed = changed || wrote
	}
	for _, key := range slices.Sorted(maps.Keys(secrets)) {
		if wanted[key] {
			continue
		}
		for _, setting := range []*entity.Setting{mounts[settingmountservice.EntryKeyFor(key)], secrets[key]} {
			if setting == nil {
				continue
			}
			if err := s.deleteSetting(ctx, db, setting, timeNow); err != nil {
				return hperrors.Wrap(err)
			}
		}
		changed = true
	}
	if !changed {
		return nil
	}
	return hperrors.Wrap(s.settingMountService.Refresh(ctx, db, app))
}

// syncSecret writes the file's value into its secret, creating the secret when
// there is none, and reports whether it wrote.
func (s *service) syncSecret(
	ctx context.Context, db database.IDB, app *entity.App, setting *entity.Setting,
	file *systemappservice.SecretFile, timeNow time.Time,
) (*entity.Setting, bool, error) {
	if setting != nil {
		current, err := setting.AsSecret()
		if err != nil {
			return nil, false, hperrors.Wrap(err)
		}
		plain, err := current.Value.GetPlain()
		if err != nil {
			return nil, false, hperrors.Wrap(err)
		}
		if plain == file.Value {
			return setting, false, nil
		}
	} else {
		setting = newAppSetting(app, base.SettingTypeSecret, file.Key, entity.CurrentSecretVersion, timeNow)
	}
	next := &entity.Secret{Key: file.Key, Value: entity.NewEncryptedField(file.Value)}
	if err := setting.SetData(next); err != nil {
		return nil, false, hperrors.Wrap(err)
	}
	size, err := next.ValueSize()
	if err != nil {
		return nil, false, hperrors.Wrap(err)
	}
	setting.Size = size
	return setting, true, s.upsertSetting(ctx, db, setting, timeNow)
}

// syncSecretMount writes the entry that mounts secret at the file's path, and
// reports whether it wrote.
func (s *service) syncSecretMount(
	ctx context.Context, db database.IDB, app *entity.App, setting *entity.Setting, key string,
	secret *entity.Setting, file *systemappservice.SecretFile, timeNow time.Time,
) (bool, error) {
	next := &entity.AppSettingMount{
		Source: entity.ObjectID{ID: secret.ID},
		Files:  []*entity.AppSettingMountFile{{Part: secretPart, Path: file.Path, Mode: secretFileMode}},
	}
	if setting != nil {
		current, err := setting.AsAppSettingMount()
		if err != nil {
			return false, hperrors.Wrap(err)
		}
		if reflect.DeepEqual(current, next) {
			return false, nil
		}
	} else {
		setting = newAppSetting(app, base.SettingTypeAppSettingMount, key,
			entity.CurrentAppSettingMountVersion, timeNow)
	}
	if err := setting.SetData(next); err != nil {
		return false, hperrors.Wrap(err)
	}
	return true, s.upsertSetting(ctx, db, setting, timeNow)
}

func (s *service) upsertSetting(
	ctx context.Context, db database.IDB, setting *entity.Setting, timeNow time.Time,
) error {
	setting.UpdateVer++
	setting.UpdatedAt = timeNow
	return hperrors.Wrap(s.settingRepo.Upsert(ctx, db, setting,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols))
}

func (s *service) deleteSetting(
	ctx context.Context, db database.IDB, setting *entity.Setting, timeNow time.Time,
) error {
	setting.UpdateVer++
	setting.UpdatedAt = timeNow
	setting.DeletedAt = timeNow
	return hperrors.Wrap(s.settingRepo.Update(ctx, db, setting,
		bunex.UpdateColumns("deleted_at", "update_ver", "updated_at")))
}

func settingsByName(settings []*entity.Setting) map[string]*entity.Setting {
	byName := make(map[string]*entity.Setting, len(settings))
	for _, setting := range settings {
		byName[setting.Name] = setting
	}
	return byName
}

// newAppSetting is a setting of the app, shaped the way the build of an app's
// document shapes one.
func newAppSetting(
	app *entity.App, typ base.SettingType, name string, version int, timeNow time.Time,
) *entity.Setting {
	return &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.ObjectScopeApp,
		ObjectID:  app.ID,
		Type:      typ,
		Name:      name,
		Status:    base.SettingStatusActive,
		Version:   version,
		CreatedAt: timeNow,
	}
}
