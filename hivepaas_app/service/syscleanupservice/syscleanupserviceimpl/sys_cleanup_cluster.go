package syscleanupserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clustercleanupservice"
)

func (s *service) sysCleanupCluster(
	ctx context.Context,
	data *sysCleanupData,
) (err error) {
	clusterCleanup := &data.SysCleanupSettings.ClusterCleanup
	if !clusterCleanup.Enabled {
		return nil
	}

	// Cleanup cluster from the current node (manager node)
	resp, err := s.clusterCleanupService.Cleanup(ctx, &clustercleanupservice.ClusterCleanupReq{
		TaskExecData:      data.TaskExecData,
		CleanupSettings:   clusterCleanup,
		CleanupContainers: data.CleanupClusterContainers,
		CleanupImages:     data.CleanupClusterImages,
		CleanupVolumes:    data.CleanupClusterVolumes,
		CleanupNetworks:   data.CleanupClusterNetworks,
		CleanupBuildCache: data.CleanupClusterBuildCache,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	if resp.Output != nil {
		if data.TaskOutput.ClusterCleanup == nil {
			data.TaskOutput.ClusterCleanup = &entity.ClusterCleanupOutput{}
		}
		data.TaskOutput.ClusterCleanup.Nodes = append(data.TaskOutput.ClusterCleanup.Nodes, resp.Output)
	}

	// TODO: run cluster cleanup on other nodes

	return nil
}
