package settings

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
)

type fakeClaims struct {
	settingmountservice.Service
	claimed map[string]string
	except  string
}

func (f *fakeClaims) ClaimedPaths(_ context.Context, _ database.IDB, _, except string) (map[string]string, error) {
	f.except = except
	return f.claimed, nil
}

// db_password and /run/secrets/db_password are one file.
func TestASecretNamedRelativelyCollidesWithAnEntry(t *testing.T) {
	claims := &fakeClaims{claimed: map[string]string{"/run/secrets/db_password": "setting mount cert"}}
	uc := &BaseUC{SettingMountService: claims}
	app := &entity.ObjectScope{ScopeType: base.ObjectScopeApp, AppID: "app_1"}

	err := uc.CheckMountPaths(context.Background(), nil, app, "s1",
		settingmountservice.SecretTarget("db_password"))

	assert.True(t, errors.Is(err, hperrors.ErrSettingMountPathTaken), "%v", err)
	assert.Equal(t, "s1", claims.except, "the secret being saved does not collide with itself")
	assert.NoError(t, uc.CheckMountPaths(context.Background(), nil,
		&entity.ObjectScope{ScopeType: base.ObjectScopeProject, ProjectID: "p1"}, "", "/run/secrets/db_password"),
		"a project's secret is no app's file")
	assert.NoError(t, uc.CheckMountPaths(context.Background(), nil, app, "", ""), "no file, nothing to check")
}
