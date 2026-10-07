package entity

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

var (
	hivepaasProject = &Project{ID: "hp", Key: base.HivepaasProjectKey, Name: base.HivepaasProjectName}
	userProject     = &Project{ID: "shop", Key: "shop", Name: "Shop"}
)

// The project HivePaaS runs in cannot be deleted or disabled, nor its
// environments: its apps are what runs this installation. Any other project
// can.
func TestTheHivePaaSProjectCannotBeDeletedOrDisabled(t *testing.T) {
	assert.True(t, errors.Is(CheckProjectChange(hivepaasProject, "deleted"), hperrors.ErrSystemProjectProtected))
	assert.ErrorIs(t, CheckProjectEnvChange(hivepaasProject, &ProjectEnv{Name: "production"}, "deleted"),
		hperrors.ErrSystemProjectEnvProtected)

	assert.NoError(t, CheckProjectChange(userProject, "deleted"))
	assert.NoError(t, CheckProjectEnvChange(userProject, &ProjectEnv{Name: "production"}, "deleted"))
	assert.NoError(t, CheckProjectChange(nil, "deleted"), "not known: nothing to say")
}

// None of the HivePaaS project's apps is deleted by hand: the stack's own are
// what runs this installation, and those HivePaaS provisions - the logs'
// backend and collector, the registry - go with their feature, turned off
// under System. Stopping or disabling is refused for the stack's own alone.
func TestHivePaaSsAppsCannotBeDeletedNorItsOwnStopped(t *testing.T) {
	traefik := &App{Key: base.HivepaasTraefikKey, Name: "traefik"}
	logs := &App{Key: base.HivepaasVictoriaLogsKey, Name: "victoria-logs"}
	shopWeb := &App{Key: base.HivepaasTraefikKey, Name: "traefik"} // a user's app may share a key

	assert.ErrorIs(t, CheckAppDeletion(hivepaasProject, traefik), hperrors.ErrSystemAppProtected)
	assert.ErrorIs(t, CheckAppDeletion(hivepaasProject, logs), hperrors.ErrSystemAppProtected)
	assert.NoError(t, CheckAppDeletion(userProject, shopWeb))

	assert.ErrorIs(t, CheckAppStop(hivepaasProject, traefik, "stopped"), hperrors.ErrSystemAppProtected)
	assert.NoError(t, CheckAppStop(hivepaasProject, logs, "stopped"), "its own page turns it off; stopping is allowed")
	assert.NoError(t, CheckAppStop(userProject, shopWeb, "stopped"))
	assert.NoError(t, CheckAppStop(nil, traefik, "stopped"), "not known: nothing to say")
}
