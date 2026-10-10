package permissionimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
)

// envAppRepo holds apps by ID, each in the env of its key, as the real query finds
// them: by project and ID, whatever env an address names.
type envAppRepo struct {
	repository.AppRepo
	apps map[string]*entity.App
}

func (f *envAppRepo) GetByID(
	_ context.Context, _ database.IDB, projectID, id string, _ ...bunex.SelectQueryOption,
) (*entity.App, error) {
	app, ok := f.apps[id]
	if !ok || (projectID != "" && app.ProjectID != projectID) {
		return nil, hperrors.ErrNotFound
	}
	return app, nil
}

func appIn(projectID, envKey string) *entity.App {
	return &entity.App{
		ProjectID:    projectID,
		ProjectEnvID: projectID + ":" + envKey,
		ProjectEnv:   &entity.ProjectEnv{ID: projectID + ":" + envKey, Key: envKey},
	}
}

// An app is reached by the grant on the env it is in, not on the env an address
// names: a member given development reaching a production app at an address of
// development would otherwise act on it with development's grant.
func TestAppAccessFollowsTheAppsOwnEnv(t *testing.T) {
	p := newProjectManager([]*entity.ACLPermission{grant(base.ResourceTypeProjectEnv, "prj_1:dev", fullAccess)}, nil)
	p.appRepo = &envAppRepo{apps: map[string]*entity.App{
		"app_dev":  appIn("prj_1", "dev"),
		"app_prod": appIn("prj_1", "prod"),
	}}

	check := func(appID, env string) bool {
		t.Helper()
		allowed, err := p.CheckAccess(context.Background(), nil, plainAuth(), &permission.AppAccessCheck{
			BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
			ProjectID:       "prj_1",
			AppID:           appID,
			ProjectEnv:      env,
		})
		assert.NoError(t, err)
		return allowed
	}

	assert.True(t, check("app_dev", "prj_1:dev"), "an app of the env given")
	assert.False(t, check("app_prod", "prj_1:prod"), "an app of another env")
	assert.False(t, check("app_prod", "prj_1:dev"), "an app of another env, at an address of the env given")
	assert.True(t, check("app_new", "prj_1:dev"), "an app not made yet, in the env given")
	assert.False(t, check("app_new", "prj_1:prod"), "an app not made yet, in another env")
}
