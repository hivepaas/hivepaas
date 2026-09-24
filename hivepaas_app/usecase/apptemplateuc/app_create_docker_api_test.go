package apptemplateuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func dockerAPIApps(t *testing.T, withAccess bool) []*appToProvision {
	t.Helper()
	settings := map[string]any{"routing": map[string]any{"port": 80}}
	if withAccess {
		settings["dockerApi"] = map[string]any{"images": []any{"autobase/automation"}}
	}
	return []*appToProvision{{
		id: "app-1", name: "autobase",
		result: &templaterender.Result{Doc: &specmodel.AppDoc{Settings: settings}},
	}}
}

func TestCheckDockerAPINeedsWriteOnTheClusterModule(t *testing.T) {
	permissions := &fakePermissionManager{granted: false}
	uc := &UC{permissionManager: permissions}

	err := uc.checkDockerAPI(context.Background(), &basedto.Auth{}, dockerAPIApps(t, true))
	assert.ErrorIs(t, err, hperrors.ErrUnauthorized)
	assert.Contains(t, errDetail(t, err), "autobase")
	if assert.Len(t, permissions.checked, 1) {
		check, ok := permissions.checked[0].(*permission.ModuleAccessCheck)
		if assert.True(t, ok) {
			assert.Equal(t, base.ResourceModuleCluster, check.Module)
			assert.Equal(t, base.ActionTypeWrite, check.Action)
		}
	}

	permissions.granted = true
	assert.NoError(t, uc.checkDockerAPI(context.Background(), &basedto.Auth{}, dockerAPIApps(t, true)))
}

func TestCheckDockerAPIAsksNothingOfATemplateWithout(t *testing.T) {
	permissions := &fakePermissionManager{granted: false}
	uc := &UC{permissionManager: permissions}
	assert.NoError(t, uc.checkDockerAPI(context.Background(), &basedto.Auth{}, dockerAPIApps(t, false)))
	assert.Empty(t, permissions.checked)
}
