package syscleanupserviceimpl

import (
	"context"
	"errors"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/service/syscleanupservice"
)

func (s *service) sysCleanupCluster(
	ctx context.Context,
	data *sysCleanupData,
) (err error) {
	clusterCleanup := data.SysCleanupSettings.ClusterCleanup
	if !clusterCleanup.Enabled {
		return nil
	}

	objectsOlderThan := clusterCleanup.OnlyObjectsOlderThan.ToDuration()

	if data.CleanupClusterContainers != syscleanupservice.CleanupFlagFalse && clusterCleanup.PruneContainers {
		resp, e := s.dockerManager.ContainerPrune(ctx, objectsOlderThan)
		if e != nil {
			data.TaskOutput.ClusterCleanup.ContainersPruneError = e.Error()
			err = errors.Join(err, e)
		} else {
			report := &resp.Report
			data.TaskOutput.ClusterCleanup.ContainersDeleted = len(report.ContainersDeleted)
			data.TaskOutput.ClusterCleanup.SpaceReclaimed += report.SpaceReclaimed
		}
	}

	if data.CleanupClusterImages != syscleanupservice.CleanupFlagFalse && clusterCleanup.PruneImages {
		resp, e := s.dockerManager.ImagePrune(ctx, false, objectsOlderThan)
		if e != nil {
			data.TaskOutput.ClusterCleanup.ImagesPruneError = e.Error()
			err = errors.Join(err, e)
		} else {
			report := &resp.Report
			data.TaskOutput.ClusterCleanup.ImagesDeleted = len(report.ImagesDeleted)
			data.TaskOutput.ClusterCleanup.SpaceReclaimed += report.SpaceReclaimed
		}
	}

	if data.CleanupClusterVolumes != syscleanupservice.CleanupFlagFalse && clusterCleanup.PruneVolumes {
		resp, e := s.dockerManager.VolumePrune(ctx, true)
		if e != nil {
			data.TaskOutput.ClusterCleanup.VolumesPruneError = e.Error()
			err = errors.Join(err, e)
		} else {
			report := &resp.Report
			data.TaskOutput.ClusterCleanup.VolumesDeleted = len(report.VolumesDeleted)
			data.TaskOutput.ClusterCleanup.SpaceReclaimed += report.SpaceReclaimed
		}
	}

	if data.CleanupClusterNetworks != syscleanupservice.CleanupFlagFalse && clusterCleanup.PruneNetworks {
		resp, e := s.dockerManager.NetworkPrune(ctx, objectsOlderThan)
		if e != nil {
			data.TaskOutput.ClusterCleanup.NetworksPruneError = e.Error()
			err = errors.Join(err, e)
		} else {
			report := &resp.Report
			data.TaskOutput.ClusterCleanup.NetworksDeleted = len(report.NetworksDeleted)
		}
	}

	return apperrors.Wrap(err)
}
