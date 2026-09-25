package specserviceimpl

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func backendDeployment(bundle *specmodel.ImportBundle) *specmodel.Deployment {
	return bundle.Envs["project_a"]["dev"].Apps["backend"].Deployment
}

// A changed block reaches the running service in one update, after the
// commit, and the env's apps have their environment applied.
func TestPhaseTwoUpdatesAChangedServiceOnce(t *testing.T) {
	svc, bundle := planFixture(t)
	backendDeployment(bundle).Resources.Limits = &specmodel.ResourceLimits{CPUs: 2}

	resp := apply(t, svc, bundle, applyReq(t, svc, bundle))
	cluster := svc.clusterService.(*fakeClusterService)
	assert.Empty(t, cluster.updated, "nothing reaches docker before the commit")

	assert.NoError(t, resp.AfterCommit(context.Background(), nil))

	if updates := cluster.updated["svc_1"]; assert.Len(t, updates, 1) {
		assert.Equal(t, int64(2e9), updates[0].TaskTemplate.Resources.Limits.NanoCPUs)
	}
	assert.Equal(t, specmodel.OutcomeApplied, node(t, resp.Plan, backendPath).Outcome)
	assert.Equal(t, []string{"p1:dev"}, svc.envVarService.(*fakeEnvVarService).scopes)
}

// An app whose service cannot be updated is marked failed with the reason; its
// configuration is saved, and the other apps are applied.
func TestPhaseTwoMarksAFailedAppAndGoesOn(t *testing.T) {
	svc, bundle := planFixture(t)
	backendDeployment(bundle).Resources.Limits = &specmodel.ResourceLimits{CPUs: 2}
	bundle.Envs["project_a"]["dev"].Apps["frontend"].Deployment = &specmodel.Deployment{Source: imageSource("web:2")}
	svc.clusterService.(*fakeClusterService).failUpdate = map[string]error{"svc_1": errors.New("update out of sequence")}

	resp := apply(t, svc, bundle, applyReq(t, svc, bundle))
	assert.NoError(t, resp.AfterCommit(context.Background(), nil))

	backend := node(t, resp.Plan, backendPath)
	assert.Equal(t, specmodel.OutcomeFailed, backend.Outcome)
	assert.Contains(t, backend.Error, "update out of sequence")
	assert.Equal(t, specmodel.OutcomeApplied, node(t, resp.Plan, "projects/project_a/envs/dev/apps/frontend").Outcome)
}
