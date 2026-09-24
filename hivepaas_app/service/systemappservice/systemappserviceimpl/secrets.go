package systemappserviceimpl

import (
	"context"
	"maps"
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
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
)

// secretFileMode is readable by whatever user the image runs as, the mode the
// registry's files are mounted with too. The files hold credentials, but only
// the container they are mounted into can see them.
const secretFileMode = fileutil.FileMode(0o444)

// secretChange is one secret setting to write once the swarm has been changed.
type secretChange struct {
	setting *entity.Setting
	// next is nil for a secret being removed.
	next *entity.Secret
}

func (s *service) SyncSecrets(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	files []*systemappservice.SecretFile,
) error {
	existing := make(map[string]*entity.Setting)
	for _, setting := range app.GetSettingsByType(base.SettingTypeSecret) {
		existing[setting.Name] = setting
	}

	timeNow := timeutil.NowUTC()
	var olds, news []*entity.Secret
	var changes []*secretChange
	wanted := make(map[string]bool, len(files))
	for _, file := range files {
		wanted[file.Key] = true
		next := &entity.Secret{
			Key:   file.Key,
			Value: entity.NewEncryptedField(file.Value),
			SwarmRef: &entity.SwarmSecretRef{
				File: &entity.SwarmRefFileTarget{Name: file.Path, Mode: secretFileMode},
			},
		}

		setting := existing[file.Key]
		if setting == nil {
			olds, news = append(olds, nil), append(news, next)
			changes = append(changes, &secretChange{setting: newSecretSetting(app, file.Key, timeNow), next: next})
			continue
		}
		current, err := setting.AsSecret()
		if err != nil {
			return hperrors.Wrap(err)
		}
		same, err := sameSecretFile(current, file)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if same {
			continue
		}
		olds, news = append(olds, current), append(news, next)
		changes = append(changes, &secretChange{setting: setting, next: next})
	}
	for _, key := range slices.Sorted(maps.Keys(existing)) {
		if wanted[key] {
			continue
		}
		current, err := existing[key].AsSecret()
		if err != nil {
			return hperrors.Wrap(err)
		}
		olds, news = append(olds, current), append(news, nil)
		changes = append(changes, &secretChange{setting: existing[key]})
	}
	if len(changes) == 0 {
		return nil
	}

	// The swarm first: creating a secret is what fills in the reference the
	// setting then records.
	if err := s.clusterSecretService.UpdateSecretsForApp(ctx, db, app, olds, news); err != nil {
		return hperrors.Wrap(err)
	}
	for _, change := range changes {
		if err := s.persistSecretChange(ctx, db, change, timeNow); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

func (s *service) persistSecretChange(
	ctx context.Context,
	db database.IDB,
	change *secretChange,
	timeNow time.Time,
) error {
	setting := change.setting
	setting.UpdateVer++
	setting.UpdatedAt = timeNow
	if change.next == nil {
		setting.DeletedAt = timeNow
		return hperrors.Wrap(s.settingRepo.Update(ctx, db, setting,
			bunex.UpdateColumns("deleted_at", "update_ver", "updated_at")))
	}
	if err := setting.SetData(change.next); err != nil {
		return hperrors.Wrap(err)
	}
	size, err := change.next.ValueSize()
	if err != nil {
		return hperrors.Wrap(err)
	}
	setting.Size = size
	return hperrors.Wrap(s.settingRepo.Upsert(ctx, db, setting,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols))
}

// sameSecretFile says whether the secret already is the file: the same value,
// mounted at the same path.
func sameSecretFile(current *entity.Secret, file *systemappservice.SecretFile) (bool, error) {
	if current.SwarmRef == nil || current.SwarmRef.File == nil || current.SwarmRef.File.Name != file.Path {
		return false, nil
	}
	plain, err := current.Value.GetPlain()
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	return plain == file.Value, nil
}

// newSecretSetting is the setting a secret gets, shaped the way the build of an
// app's document shapes one.
func newSecretSetting(app *entity.App, key string, timeNow time.Time) *entity.Setting {
	return &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.ObjectScopeApp,
		ObjectID:  app.ID,
		Type:      base.SettingTypeSecret,
		Name:      key,
		Status:    base.SettingStatusActive,
		Version:   entity.CurrentSecretVersion,
		CreatedAt: timeNow,
	}
}
