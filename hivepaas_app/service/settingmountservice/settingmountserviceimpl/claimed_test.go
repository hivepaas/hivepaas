package settingmountserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

func claimant(t *testing.T, id string, typ base.SettingType, name string, data entity.SettingData) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: id, Type: typ, Name: name, ObjectID: testApp.ID, Status: base.SettingStatusActive}
	assert.NoError(t, setting.SetData(data))
	return setting
}

func TestClaimedPathsAreEveryFileOfTheApp(t *testing.T) {
	useDataKey(t)
	svc := &service{}
	claimants := []*entity.Setting{
		claimant(t, "s1", base.SettingTypeSecret, "DB_PASSWORD", &entity.Secret{Key: "DB_PASSWORD",
			Value:    entity.NewEncryptedField("x"),
			SwarmRef: &entity.SwarmSecretRef{File: &entity.SwarmRefFileTarget{Name: "db_password"}}}),
		claimant(t, "s2", base.SettingTypeSecret, "API_TOKEN", &entity.Secret{Key: "API_TOKEN",
			Value: entity.NewEncryptedField("y")}),
		claimant(t, "c1", base.SettingTypeConfigFile, "app.conf", &entity.ConfigFile{Name: "app.conf",
			SwarmRef: &entity.SwarmConfigRef{File: &entity.SwarmRefFileTarget{Name: "app.conf"}}}),
		claimant(t, "m1", base.SettingTypeAppSettingMount, "cert", certFiles("cert_1")),
	}
	svc.loadClaimants = func(context.Context, database.IDB, string) ([]*entity.Setting, error) {
		return claimants, nil
	}

	claimed, err := svc.ClaimedPaths(context.Background(), nil, testApp.ID, "m1")

	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		"/run/secrets/db_password": "secret DB_PASSWORD",
		"/app.conf":                "config file app.conf",
	}, claimed, "the entry being saved is left out, and a secret without a file claims nothing")
}
