package hpappsettingsdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// The apps in host mode are listed whatever the switch says: turning it off
// takes the socket from none of them, and this is where an administrator finds
// them.
func TestSecuritySettingsListTheAppsGivenTheNodesSocket(t *testing.T) {
	cfg := &config.Config{}
	resp, err := TransformSecuritySettings(&SecuritySettingsTransformInput{Config: cfg})
	assert.NoError(t, err)
	assert.NotNil(t, resp.PrivilegedApps, "no apps is an empty list, not null")
	assert.Empty(t, resp.PrivilegedApps)

	resp, err = TransformSecuritySettings(&SecuritySettingsTransformInput{Config: cfg,
		PrivilegedApps: []*entity.App{{
			ID: "app-1", Name: "Portainer", ProjectID: "p1", ProjectEnvID: "p1:prod",
			Project: &entity.Project{Name: "Ops"}, ProjectEnv: &entity.ProjectEnv{Key: "prod", Name: "Production"},
		}}})
	assert.NoError(t, err)
	assert.Equal(t, []*PrivilegedAppResp{{
		AppID: "app-1", AppName: "Portainer", ProjectID: "p1", ProjectName: "Ops",
		ProjectEnvKey: "prod", ProjectEnvName: "Production",
	}}, resp.PrivilegedApps)
}
