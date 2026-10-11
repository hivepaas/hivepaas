package imagescanserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
)

// Status reads what runs. What cannot be read - a node list Docker does not
// answer - is left out rather than failing the screen.
func (s *service) Status(ctx context.Context, db database.IDB) (*imagescanservice.Status, error) {
	status := &imagescanservice.Status{Image: trivyImage(), ScanResources: defaultScanResources()}
	_, cfg, err := s.Settings(ctx, db)
	if err != nil {
		return nil, err
	}
	status.Enabled = cfg != nil && cfg.Enabled

	server, err := s.systemAppService.LoadApp(ctx, db, base.HivepaasTrivyKey)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	runner, err := s.systemAppService.LoadApp(ctx, db, base.HivepaasTrivyRunnerKey)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if server != nil {
		status.ServerAppID = server.ID
		status.ServerRunning = len(s.runningNodes(ctx, server)) > 0
	}
	if runner != nil {
		status.RunnerAppID = runner.ID
		status.RunnerNodes = len(s.runningNodes(ctx, runner))
		if res, err := s.systemAppService.ReadResources(ctx, runner); err == nil && res != nil {
			status.ScanResources = *res
		}
	}
	if nodes, err := s.dockerManager.NodeList(ctx); err == nil {
		status.Nodes = len(nodes.Items)
	}

	latest, err := s.imageScanRepo.LatestScan(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if latest != nil {
		status.LastScanAt, status.DBUpdatedAt, status.TrivyVersion = latest.ScannedAt, latest.DBUpdatedAt,
			latest.TrivyVersion
	}
	return status, nil
}

// runningNodes are the nodes an app's service has a task running on, by the
// container each runs; none when it cannot be read.
func (s *service) runningNodes(ctx context.Context, app *entity.App) map[string]string {
	running := map[string]string{}
	if app.ServiceID == "" {
		return running
	}
	tasks, err := s.dockerManager.ServiceTaskList(ctx, app.ServiceID, []swarm.TaskState{swarm.TaskStateRunning})
	if err != nil {
		return running
	}
	for i := range tasks.Items {
		task := &tasks.Items[i]
		if task.Status.State != swarm.TaskStateRunning || task.Status.ContainerStatus == nil ||
			task.Status.ContainerStatus.ContainerID == "" {
			continue
		}
		running[task.NodeID] = task.Status.ContainerStatus.ContainerID
	}
	return running
}
