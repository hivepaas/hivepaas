package syscleanupserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/syscleanupservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	tempResourceGracePeriod = 1 * time.Hour
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

	// Always cleanup orphaned temporary batch containers (hivepaas-cont-) older than grace period
	if e := s.cleanupTempContainers(ctx, data); e != nil {
		err = errors.Join(err, e)
	}

	// Always cleanup orphaned temporary batch services (hivepaas-svc-) older than grace period
	if e := s.cleanupTempServices(ctx, data); e != nil {
		err = errors.Join(err, e)
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

	// TODO: clean build cache in all nodes
	if data.CleanupClusterBuildCache == syscleanupservice.CleanupFlagForce {
		resp, e := s.dockerManager.BuildCachePrune(ctx)
		if e != nil {
			data.TaskOutput.ClusterCleanup.BuildCachesPruneError = e.Error()
			err = errors.Join(err, e)
		} else {
			report := &resp.Report
			data.TaskOutput.ClusterCleanup.BuildCachesDeleted = len(report.CachesDeleted)
			data.TaskOutput.ClusterCleanup.SpaceReclaimed += report.SpaceReclaimed
		}
	}

	return apperrors.Wrap(err)
}

func (s *service) cleanupTempContainers(ctx context.Context, data *sysCleanupData) error {
	opts := client.ContainerListOptions{
		All: true,
	}
	docker.FilterAdd(&opts.Filters, "label", docker.LabelTempResource+"="+docker.LabelTempResourceVal)

	resp, err := s.dockerManager.ContainerList(ctx, func(o *client.ContainerListOptions) {
		*o = opts
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	timeNow := time.Now().UTC()
	deletedCount := 0

	for _, cont := range resp.Items {
		// 1. Must not be running
		if cont.State == "running" {
			continue
		}

		// 2. Grace period check (at least 1 hour old)
		createdAt := time.Unix(cont.Created, 0)
		if createdStr, ok := cont.Labels[docker.LabelTempCreatedAt]; ok && createdStr != "" {
			if t, parseErr := time.Parse(time.RFC3339, createdStr); parseErr == nil {
				createdAt = t
			}
		}
		if timeNow.Sub(createdAt) < tempResourceGracePeriod {
			continue
		}

		// Remove container
		_, remErr := s.dockerManager.ContainerRemove(ctx, cont.ID, func(o *client.ContainerRemoveOptions) {
			o.Force = true
		})
		if remErr == nil {
			deletedCount++
		}
	}

	data.TaskOutput.ClusterCleanup.TempContainersDeleted = deletedCount
	return nil
}

func (s *service) cleanupTempServices(ctx context.Context, data *sysCleanupData) error {
	opts := client.ServiceListOptions{}
	docker.FilterAdd(&opts.Filters, "label", docker.LabelTempResource+"="+docker.LabelTempResourceVal)

	resp, err := s.dockerManager.ServiceList(ctx, func(o *client.ServiceListOptions) {
		*o = opts
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	timeNow := time.Now().UTC()
	deletedCount := 0

	for _, svc := range resp.Items {
		// 1. Grace period check (at least 1 hour old)
		createdAt := svc.CreatedAt
		if createdStr, ok := svc.Spec.Labels[docker.LabelTempCreatedAt]; ok && createdStr != "" {
			if t, parseErr := time.Parse(time.RFC3339, createdStr); parseErr == nil {
				createdAt = t
			}
		}
		if timeNow.Sub(createdAt) < tempResourceGracePeriod {
			continue
		}

		// 2. Check if service has any active running tasks
		tasksResp, taskErr := s.dockerManager.ServiceTaskList(ctx, svc.ID, nil)
		if taskErr == nil {
			hasActiveTask := false
			for _, task := range tasksResp.Items {
				if task.Status.State == swarm.TaskStateRunning ||
					task.Status.State == swarm.TaskStateStarting ||
					task.Status.State == swarm.TaskStatePreparing ||
					task.Status.State == swarm.TaskStateReady {
					hasActiveTask = true
					break
				}
			}
			if hasActiveTask {
				continue
			}
		}

		// Remove service
		_, remErr := s.dockerManager.ServiceRemove(ctx, svc.ID)
		if remErr == nil {
			deletedCount++
		}
	}

	data.TaskOutput.ClusterCleanup.TempServicesDeleted = deletedCount
	return nil
}
