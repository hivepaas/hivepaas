package specserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func imageSource(image string) map[string]any {
	return map[string]any{"activeMethod": "image", "imageSource": map[string]any{"image": image}}
}

// A new app is provisioned with the key it had, its settings and the service
// built from its deployment, and deployed when the options ask.
func TestApplyProvisionsANewApp(t *testing.T) {
	svc, bundle := planFixture(t)
	addWorker(bundle, routingAt("worker.example.com"))
	bundle.Envs["project_a"]["dev"].Apps["worker"].Deployment.Source = imageSource("nginx:1.27")

	resp := apply(t, svc, bundle, applyReqWith(t, svc, bundle, specmodel.ImportOptions{DeployCreated: true}))

	provision := svc.appProvisionService.(*fakeProvisionService)
	if !assert.Len(t, provision.reqs, 1) {
		return
	}
	req := provision.reqs[0]
	assert.Equal(t, []any{"worker", "Worker", "p1", "p1:dev"}, []any{req.Key, req.Name, req.ProjectID, req.ProjectEnvID})
	assert.NotEmpty(t, req.AppID)

	var types []base.SettingType
	for _, setting := range provision.settings[req.AppID] {
		assert.Equal(t, req.AppID, setting.ObjectID)
		types = append(types, setting.Type)
	}
	assert.ElementsMatch(t, []base.SettingType{base.SettingTypeAppRouting, base.SettingTypeAppDeployment}, types)

	spec := provision.specs[req.AppID]
	if assert.Len(t, spec.TaskTemplate.ContainerSpec.Mounts, 1) {
		assert.Equal(t, mount.Mount{Type: mount.TypeVolume, Source: "vol_setting_1", Target: "/data"},
			spec.TaskTemplate.ContainerSpec.Mounts[0])
	}

	assert.Equal(t, []*specservice.ImportDeployment{{AppID: req.AppID, DeploymentID: "dep_" + req.AppID}},
		resp.Deployments)
	assert.Len(t, resp.Tasks, 1)
	assert.Equal(t, specmodel.OutcomeApplied, node(t, resp.Plan, workerPath).Outcome)

	assert.NoError(t, resp.Cleanup(context.Background()))
	assert.True(t, provision.cleanedUp, "what provisioning made can be removed")
}

// An existing app whose source changed is deployed when the options ask, and its
// settings are persisted with the scopes'.
func TestApplyDeploysAnAppWhoseSourceChanged(t *testing.T) {
	svc, bundle := planFixture(t)
	bundle.Envs["project_a"]["dev"].Apps["frontend"].Deployment = &specmodel.Deployment{Source: imageSource("web:2")}

	resp := apply(t, svc, bundle, applyReqWith(t, svc, bundle, specmodel.ImportOptions{DeployChangedSource: true}))

	source := persistedSetting(t, svc, ofType(base.SettingTypeAppDeployment))
	assert.Equal(t, "app_2", source.ObjectID)
	assert.Equal(t, base.ObjectScopeApp, source.Scope)
	assert.Equal(t, []*specservice.ImportDeployment{{AppID: "app_2", DeploymentID: "dep_app_2"}}, resp.Deployments)
	persistedApps := svc.appService.(*fakeAppService).persisted
	if assert.Len(t, persistedApps, 1) {
		assert.Equal(t, []*entity.Task{{ID: "task_app_2", ObjectID: "app_2"}}, persistedApps[0].UpsertingTasks)
	}
	assert.Empty(t, svc.appProvisionService.(*fakeProvisionService).reqs, "nothing is provisioned")
}
