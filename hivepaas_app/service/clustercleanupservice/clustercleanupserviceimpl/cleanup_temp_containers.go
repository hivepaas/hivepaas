package clustercleanupserviceimpl

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	tempResourceGracePeriod = 1 * time.Hour
)

func (s *service) cleanupTempContainers(ctx context.Context, data *clusterCleanupData) error {
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

	data.Output.TempContainersDeleted = deletedCount
	return nil
}

func (s *service) cleanupTempServices(ctx context.Context, data *clusterCleanupData) error {
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

	data.Output.TempServicesDeleted = deletedCount
	return nil
}
