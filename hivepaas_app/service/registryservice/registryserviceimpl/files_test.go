package registryserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
)

type fakeUpsertRepo struct {
	repository.SettingRepo
	upserted []string
}

func (f *fakeUpsertRepo) Upsert(
	_ context.Context, _ database.IDB, setting *entity.Setting, _, _ []string, _ ...bunex.InsertQueryOption,
) error {
	f.upserted = append(f.upserted, setting.Name)
	return nil
}

type fakeMounts struct {
	settingmountservice.Service
	refreshed []string
}

func (f *fakeMounts) Refresh(_ context.Context, _ database.IDB, app *entity.App) error {
	f.refreshed = append(f.refreshed, app.ID)
	return nil
}

func registryApp(t *testing.T, config, htpasswd string) *entity.App {
	t.Helper()
	configFile := &entity.Setting{ID: "cfg", Type: base.SettingTypeConfigFile, Name: registryConfigFileName}
	assert.NoError(t, configFile.SetData(&entity.ConfigFile{Name: registryConfigFileName, Content: config}))
	secret := &entity.Setting{ID: "sec", Type: base.SettingTypeSecret, Name: registrySecretName}
	assert.NoError(t, secret.SetData(&entity.Secret{Key: registrySecretName, Value: entity.NewEncryptedField(htpasswd)}))
	return &entity.App{ID: "registry-app", Settings: []*entity.Setting{configFile, secret}}
}

// The registry's configuration and account are files its setting mounts put in
// its container: a change is written, then the mounts refreshed, which is what
// replaces the files.
func TestApplyingTheRegistryFilesRefreshesItsMountsOnlyOnChange(t *testing.T) {
	useDataKey(t)
	for name, tc := range map[string]struct {
		config, htpasswd string
		wantWritten      []string
	}{
		"nothing changed":   {"{}", "admin:x", nil},
		"the configuration": {"{\"log\":1}", "admin:x", []string{registryConfigFileName}},
		"the account":       {"{}", "admin:y", []string{registrySecretName}},
	} {
		repo, mounts := &fakeUpsertRepo{}, &fakeMounts{}
		svc := &service{settingRepo: repo, settingMountService: mounts}

		err := svc.applyFiles(context.Background(), nil, registryApp(t, "{}", "admin:x"),
			appDocInput{ZotConfig: tc.config, Htpasswd: tc.htpasswd})

		assert.NoError(t, err, name)
		assert.Equal(t, tc.wantWritten, repo.upserted, name)
		if tc.wantWritten == nil {
			assert.Empty(t, mounts.refreshed, name)
		} else {
			assert.Equal(t, []string{"registry-app"}, mounts.refreshed, name)
		}
	}
}
