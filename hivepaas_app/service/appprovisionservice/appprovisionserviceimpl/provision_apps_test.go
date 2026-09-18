package appprovisionserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
)

// appReq is a request for one app, configured with what a deployment needs.
func appReq(name string, configure appprovisionservice.ConfigureFunc) *appprovisionservice.ProvisionAppReq {
	return &appprovisionservice.ProvisionAppReq{
		ProjectID: "p1", ProjectEnvID: "p1:prod", Name: name, Status: base.AppStatusActive,
		Configure: configure,
	}
}

func deployableConfigure(t *testing.T) appprovisionservice.ConfigureFunc {
	t.Helper()
	return func(_ context.Context, _ database.IDB, app *entity.App, _ *swarm.ServiceSpec) (
		[]*entity.Setting, error) {
		setting := &entity.Setting{ID: "set-deploy-" + app.Key, Type: base.SettingTypeAppDeployment,
			ObjectID: app.ID}
		assert.NoError(t, setting.SetData(&entity.AppDeploymentSettings{
			ActiveMethod: base.DeploymentMethodImage,
			ImageSource:  &entity.DeploymentImageSource{Image: "postgres:18.6-alpine3.24"},
		}))
		return []*entity.Setting{setting}, nil
	}
}

func errDetail(t *testing.T, err error) string {
	t.Helper()
	var hpErr hperrors.HPError
	if !assert.ErrorAs(t, err, &hpErr) {
		return err.Error()
	}
	return hpErr.Build("en").Detail
}

func TestProvisionAppCreatesTheFirstDeployment(t *testing.T) {
	svc, fakes := newProvisionTest(t)
	req := appReq("db", deployableConfigure(t))
	req.Deployment = &appprovisionservice.FirstDeployment{
		Source: base.DeploymentTriggerSourceAPI, SourceID: "user-1",
	}

	resp, err := svc.ProvisionApp(context.Background(), nil, req)

	assert.NoError(t, err)
	assert.Equal(t, "postgres:18.6-alpine3.24", resp.Deployment.Settings.ImageSource.Image)
	assert.Equal(t, base.DeploymentTriggerSourceAPI, resp.Deployment.Trigger.Source)
	assert.Equal(t, "user-1", resp.Deployment.Trigger.SourceID)
	assert.Equal(t, []*entity.Deployment{resp.Deployment}, fakes.apps.persisted.UpsertingDeployments)
	assert.Equal(t, []*entity.Task{resp.DeploymentTask}, fakes.apps.persisted.UpsertingTasks)
}

func TestProvisionAppDeploysNothingWhenNotAsked(t *testing.T) {
	svc, _ := newProvisionTest(t)

	resp, err := svc.ProvisionApp(context.Background(), nil, appReq("db", deployableConfigure(t)))

	assert.NoError(t, err)
	assert.Nil(t, resp.Deployment)
	assert.Nil(t, resp.DeploymentTask)
}

func TestProvisionAppsProvisionsInTheOrderGiven(t *testing.T) {
	svc, _ := newProvisionTest(t)

	resp, err := svc.ProvisionApps(context.Background(), nil, &appprovisionservice.ProvisionAppsReq{
		Apps: []*appprovisionservice.ProvisionAppReq{
			appReq("blog-db", deployableConfigure(t)),
			appReq("blog", deployableConfigure(t)),
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, []string{"blog-db", "blog"}, []string{resp.Apps[0].App.Name, resp.Apps[1].App.Name})
}

func TestProvisionAppsNamesTheAppThatFailed(t *testing.T) {
	svc, _ := newProvisionTest(t)
	failing := appReq("blog-db", func(context.Context, database.IDB, *entity.App, *swarm.ServiceSpec) (
		[]*entity.Setting, error) {
		return nil, errTestProvision
	})

	resp, err := svc.ProvisionApps(context.Background(), nil, &appprovisionservice.ProvisionAppsReq{
		Apps: []*appprovisionservice.ProvisionAppReq{failing, appReq("blog", deployableConfigure(t))},
	})

	assert.ErrorIs(t, err, errTestProvision)
	assert.Contains(t, errDetail(t, err), "blog-db")
	assert.Nil(t, resp.Apps[0].Created, "a service that was never created is not left for the caller to remove")
}

// A creation that fails leaves the transaction to roll back, and what it made in
// docker with nothing pointing at it. Cleanup is how the caller undoes that, and
// it goes newest first so that an app is removed before what it was created for.
func TestProvisionAppsCleanupRemovesEverythingItMadeNewestFirst(t *testing.T) {
	svc, fakes := newProvisionTest(t)
	withSecret := appReq("blog", func(_ context.Context, _ database.IDB, app *entity.App,
		_ *swarm.ServiceSpec) ([]*entity.Setting, error) {
		setting := &entity.Setting{ID: "set-secret", Type: base.SettingTypeSecret, ObjectID: app.ID}
		assert.NoError(t, setting.SetData(&entity.Secret{Key: "LICENSE",
			SwarmRef: &entity.SwarmSecretRef{File: &entity.SwarmRefFileTarget{Name: "/etc/license"}}}))
		return []*entity.Setting{setting}, nil
	})

	resp, err := svc.ProvisionApps(context.Background(), nil, &appprovisionservice.ProvisionAppsReq{
		Apps: []*appprovisionservice.ProvisionAppReq{appReq("blog-db", nil), withSecret},
	})
	assert.NoError(t, err)
	assert.NoError(t, resp.Cleanup(context.Background()))

	assert.Equal(t, []string{"svc-1", "svc-1"}, fakes.cluster.removed, "both services, the newest first")
	assert.Equal(t, []string{"docker-secret-LICENSE"}, fakes.clusterFiles.removedSecrets)
}
