package appdeploymentuc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appdeploymentuc/appdeploymentdto"
)

func notEnded(ids ...string) []*entity.Deployment {
	deployments := make([]*entity.Deployment, 0, len(ids))
	for _, id := range ids {
		deployments = append(deployments, &entity.Deployment{ID: id, Status: base.DeploymentStatusNotStarted})
	}
	return deployments
}

// No deployment that has not ended: nothing is active.
func TestNoDeploymentNotEndedIsNoneActive(t *testing.T) {
	assert.Nil(t, pickActiveDeployment(nil, nil))
}

// The deployment running is the active one, though older ones are queued: its
// row says not-started until it ends, and the cache says it runs.
func TestTheDeploymentRunningIsTheActiveOne(t *testing.T) {
	infos := map[string]*cacheentity.DeploymentInfo{
		"d2": {ID: "d2", Status: base.DeploymentStatusInProgress},
	}

	assert.Equal(t, &appdeploymentdto.ActiveDeploymentResp{ID: "d2", Status: base.DeploymentStatusInProgress},
		pickActiveDeployment(notEnded("d1", "d2", "d3"), infos))
}

// None running, the oldest queued is the active one: it runs next.
func TestTheNextToRunIsActiveWhenNoneRuns(t *testing.T) {
	assert.Equal(t, &appdeploymentdto.ActiveDeploymentResp{ID: "d1", Status: base.DeploymentStatusNotStarted},
		pickActiveDeployment(notEnded("d1", "d2"), nil))
}

// Only the deployments the user may see are kept.
func TestKeepDeploymentsKeepsThoseAllowed(t *testing.T) {
	kept := keepDeployments(notEnded("d1", "d2", "d3"), []string{"d3", "d1"})

	assert.Equal(t, []string{"d1", "d3"}, []string{kept[0].ID, kept[1].ID})
}
