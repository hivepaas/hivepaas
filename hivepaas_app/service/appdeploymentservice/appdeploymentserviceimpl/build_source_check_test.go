package appdeploymentserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
)

// fakeClusterService answers whether the cluster has several nodes.
type fakeClusterService struct {
	clusterservice.Service
	multiNode bool
}

func (f *fakeClusterService) IsMultiNode(context.Context) (bool, error) {
	return f.multiNode, nil
}

// A build on a cluster of several nodes pushes its image to a registry every
// node pulls from: without one, the settings are refused before they are saved.
func TestABuildOnSeveralNodesNeedsARegistry(t *testing.T) {
	several := &service{clusterService: &fakeClusterService{multiNode: true}}
	one := &service{clusterService: &fakeClusterService{}}
	withRegistry := &appdeploymentservice.CheckBuildSourceReq{PushToRegistry: entity.ObjectID{ID: "registry-1"}}

	assert.ErrorIs(t, several.CheckBuildSource(context.Background(), &appdeploymentservice.CheckBuildSourceReq{}),
		hperrors.ErrMultiNodeClusterRequireRegistryForImages)
	assert.NoError(t, several.CheckBuildSource(context.Background(), withRegistry))
	assert.NoError(t, one.CheckBuildSource(context.Background(), &appdeploymentservice.CheckBuildSourceReq{}))
}
