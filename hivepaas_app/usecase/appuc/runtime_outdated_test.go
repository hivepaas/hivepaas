package appuc

import (
	"context"
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

// fakeDeployments answers the deployments it holds, and counts the lists.
type fakeDeployments struct {
	repository.DeploymentRepo
	list  []*entity.Deployment
	lists int
}

func (f *fakeDeployments) List(_ context.Context, _ database.IDB, _ string, _ *basedto.Paging,
	_ ...bunex.SelectQueryOption) ([]*entity.Deployment, *basedto.PagingMeta, error) {
	f.lists++
	return f.list, nil, nil
}

func deployedOn(images map[string]string) []*entity.Deployment {
	return []*entity.Deployment{{Status: base.DeploymentStatusDone,
		Output: &entity.AppDeploymentOutput{RuntimeImages: images}}}
}

// A function is on an older runtime when the deployment it runs was built on
// other runtime images than this HivePaaS version's; one never deployed, or
// deployed before deployments recorded their runtime images, is not.
func TestAFunctionIsOnAnOlderRuntimeWhenItsDeploymentSaysSo(t *testing.T) {
	current := systemappservice.CurrentRelease().FunctionRuntimes
	older := maps.Clone(current)
	older["node24"] = "ghcr.io/hivepaas/function-runtime-node24:0.9.0@sha256:old"
	for name, tc := range map[string]struct {
		deployments []*entity.Deployment
		want        bool
	}{
		"never deployed":             {nil, false},
		"deployed before the record": {deployedOn(nil), false},
		"on this version's runtime":  {deployedOn(map[string]string{"node24": current["node24"]}), false},
		"on an older runtime":        {deployedOn(map[string]string{"node24": older["node24"]}), true},
	} {
		app := appWithSettings(t, base.AppCategoryFunction, nil)
		app.ProjectEnv = &entity.ProjectEnv{Name: "dev"}
		uc := &UC{appService: &fakeApps{app: app}, deploymentRepo: &fakeDeployments{list: tc.deployments}}

		resp, err := uc.GetApp(context.Background(), nil, &appdto.GetAppReq{ProjectID: "p", AppID: app.ID})

		if assert.NoError(t, err, name) {
			assert.Equal(t, tc.want, resp.Data.RuntimeOutdated, name)
		}
	}
}

// Only a function has a runtime: no other app's deployments are read for it.
func TestAnAppThatIsNoFunctionHasNoRuntimeToBeOlder(t *testing.T) {
	app := appWithSettings(t, base.AppCategoryWebapp, nil)
	app.ProjectEnv = &entity.ProjectEnv{Name: "dev"}
	deployments := &fakeDeployments{list: deployedOn(map[string]string{"node24": "old"})}
	uc := &UC{appService: &fakeApps{app: app}, deploymentRepo: deployments}

	resp, err := uc.GetApp(context.Background(), nil, &appdto.GetAppReq{ProjectID: "p", AppID: app.ID})

	if assert.NoError(t, err) {
		assert.False(t, resp.Data.RuntimeOutdated)
		assert.Zero(t, deployments.lists)
	}
}

// The deployment a function runs is its latest that succeeded: a failed or
// canceled one changed nothing.
func TestTheDeploymentAnAppRunsIsItsLatestDone(t *testing.T) {
	db := bun.NewDB(nil, pgdialect.New())
	var deployments []*entity.Deployment
	sql := bunex.ApplySelect(db.NewSelect().Model(&deployments), latestDoneDeployment()...).String()

	assert.Contains(t, sql, `WHERE (deployment.status = 'done')`)
	assert.Contains(t, sql, `ORDER BY "deployment"."created_at" DESC`)
	assert.Contains(t, sql, `LIMIT 1`)
}
