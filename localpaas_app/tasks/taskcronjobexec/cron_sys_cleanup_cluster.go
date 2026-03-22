package taskcronjobexec

import (
	"context"
	"errors"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/entity"
)

func (e *Executor) sysClusterCleanup(
	ctx context.Context,
	clusterCleanup *entity.ClusterCleanup,
	data *sysCleanupTaskData,
) (err error) {
	if clusterCleanup == nil || !clusterCleanup.Enabled {
		return nil
	}

	objectsOlderThan := clusterCleanup.OnlyObjectsOlderThan.ToDuration()

	if clusterCleanup.PruneContainers {
		report, e := e.dockerManager.ContainersPrune(ctx, objectsOlderThan)
		if e != nil {
			data.TaskOutput.ClusterCleanup.ContainersPruneError = e.Error()
			err = errors.Join(err, e)
		} else {
			data.TaskOutput.ClusterCleanup.ContainersDeleted = len(report.ContainersDeleted)
			data.TaskOutput.ClusterCleanup.SpaceReclaimed += report.SpaceReclaimed
		}
	}

	if clusterCleanup.PruneImages {
		report, e := e.dockerManager.ImagesPrune(ctx, false, objectsOlderThan)
		if e != nil {
			data.TaskOutput.ClusterCleanup.ImagesPruneError = e.Error()
			err = errors.Join(err, e)
		} else {
			data.TaskOutput.ClusterCleanup.ImagesDeleted = len(report.ImagesDeleted)
			data.TaskOutput.ClusterCleanup.SpaceReclaimed += report.SpaceReclaimed
		}
	}

	if clusterCleanup.PruneVolumes {
		report, e := e.dockerManager.VolumesPrune(ctx, true)
		if e != nil {
			data.TaskOutput.ClusterCleanup.VolumesPruneError = e.Error()
			err = errors.Join(err, e)
		} else {
			data.TaskOutput.ClusterCleanup.VolumesDeleted = len(report.VolumesDeleted)
			data.TaskOutput.ClusterCleanup.SpaceReclaimed += report.SpaceReclaimed
		}
	}

	if clusterCleanup.PruneNetworks {
		report, e := e.dockerManager.NetworksPrune(ctx, objectsOlderThan)
		if e != nil {
			data.TaskOutput.ClusterCleanup.NetworksPruneError = e.Error()
			err = errors.Join(err, e)
		} else {
			data.TaskOutput.ClusterCleanup.NetworksDeleted = len(report.NetworksDeleted)
		}
	}

	return apperrors.Wrap(err)
}
