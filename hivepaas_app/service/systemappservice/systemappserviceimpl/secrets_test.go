package systemappserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
)

type fakeSettingWrites struct {
	repository.SettingRepo
	upserted map[string]*entity.Setting
	deleted  []string
}

func (f *fakeSettingWrites) Upsert(
	_ context.Context, _ database.IDB, setting *entity.Setting, _, _ []string, _ ...bunex.InsertQueryOption,
) error {
	f.upserted[string(setting.Type)+"/"+setting.Name] = setting
	return nil
}

func (f *fakeSettingWrites) Update(
	_ context.Context, _ database.IDB, setting *entity.Setting, _ ...bunex.UpdateQueryOption,
) error {
	if !setting.DeletedAt.IsZero() {
		f.deleted = append(f.deleted, string(setting.Type)+"/"+setting.Name)
	}
	return nil
}

type fakeRefresh struct {
	settingmountservice.Service
	refreshed int
}

func (f *fakeRefresh) Refresh(context.Context, database.IDB, *entity.App) error {
	f.refreshed++
	return nil
}

func useDataKey(t *testing.T) {
	t.Helper()
	key, err := datakey.Generate()
	assert.NoError(t, err)
	datakey.SetActive(key)
}

// A system app's secret files are secrets, each mounted by a setting mount the
// way a template's swarmRef would make it; the service is refreshed once.
func TestSyncSecretsWritesSecretsAndTheirMounts(t *testing.T) {
	useDataKey(t)
	repo, mounts := &fakeSettingWrites{upserted: map[string]*entity.Setting{}}, &fakeRefresh{}
	svc := &service{settingRepo: repo, settingMountService: mounts}
	stale := &entity.Setting{ID: "old", Type: base.SettingTypeSecret, Name: "OLD_TOKEN"}
	assert.NoError(t, stale.SetData(&entity.Secret{Key: "OLD_TOKEN", Value: entity.NewEncryptedField("x")}))
	staleMount := &entity.Setting{ID: "old-mount", Type: base.SettingTypeAppSettingMount, Name: "old-token"}
	assert.NoError(t, staleMount.SetData(&entity.AppSettingMount{Source: entity.ObjectID{ID: "old"}}))
	app := &entity.App{ID: "collector", Settings: []*entity.Setting{stale, staleMount}}

	err := svc.SyncSecrets(context.Background(), nil, app, []*systemappservice.SecretFile{
		{Key: "LOKI_PASSWORD", Path: "/run/secrets/loki_password", Value: "s3cret"},
	})

	assert.NoError(t, err)
	secret, entry := repo.upserted["secret/LOKI_PASSWORD"], repo.upserted["app-setting-mount/loki-password"]
	if assert.NotNil(t, secret) && assert.NotNil(t, entry) {
		assert.Equal(t, &entity.AppSettingMount{Source: entity.ObjectID{ID: secret.ID},
			Files: []*entity.AppSettingMountFile{{Part: "value", Path: "/run/secrets/loki_password",
				Mode: fileutil.FileMode(0o444)}}}, entry.MustAsAppSettingMount())
	}
	assert.ElementsMatch(t, []string{"secret/OLD_TOKEN", "app-setting-mount/old-token"}, repo.deleted)
	assert.Equal(t, 1, mounts.refreshed)
}

// The same files again change nothing, and restart nothing.
func TestSyncSecretsLeavesTheSameFilesAlone(t *testing.T) {
	useDataKey(t)
	secret := &entity.Setting{ID: "s1", Type: base.SettingTypeSecret, Name: "LOKI_PASSWORD"}
	assert.NoError(t, secret.SetData(&entity.Secret{Key: "LOKI_PASSWORD", Value: entity.NewEncryptedField("s3cret")}))
	entry := &entity.Setting{ID: "m1", Type: base.SettingTypeAppSettingMount, Name: "loki-password"}
	assert.NoError(t, entry.SetData(&entity.AppSettingMount{Source: entity.ObjectID{ID: "s1"},
		Files: []*entity.AppSettingMountFile{{Part: "value", Path: "/run/secrets/loki_password",
			Mode: fileutil.FileMode(0o444)}}}))
	repo, mounts := &fakeSettingWrites{upserted: map[string]*entity.Setting{}}, &fakeRefresh{}
	svc := &service{settingRepo: repo, settingMountService: mounts}

	err := svc.SyncSecrets(context.Background(), nil, &entity.App{ID: "collector",
		Settings: []*entity.Setting{secret, entry}}, []*systemappservice.SecretFile{
		{Key: "LOKI_PASSWORD", Path: "/run/secrets/loki_password", Value: "s3cret"},
	})

	assert.NoError(t, err)
	assert.Empty(t, repo.upserted)
	assert.Zero(t, mounts.refreshed)
}
