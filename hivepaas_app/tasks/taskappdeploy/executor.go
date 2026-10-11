package taskappdeploy

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type Executor struct {
	db        *database.DB
	taskQueue queue.TaskQueue
	appRepo   repository.AppRepo
	logger    logging.Logger

	appDeploymentService appdeploymentservice.Service
	imageScanService     imagescanservice.Service
}

func NewExecutor(
	taskQueue queue.TaskQueue,
	db *database.DB,
	appRepo repository.AppRepo,
	logger logging.Logger,
	appDeploymentService appdeploymentservice.Service,
	imageScanService imagescanservice.Service,
) *Executor {
	e := &Executor{
		db:                   db,
		taskQueue:            taskQueue,
		appRepo:              appRepo,
		logger:               logger,
		appDeploymentService: appDeploymentService,
		imageScanService:     imageScanService,
	}
	taskQueue.RegisterExecutor(base.TaskTypeAppDeploy, e.execute)
	return e
}

func (e *Executor) execute(
	ctx context.Context,
	db database.Tx,
	execData *queue.TaskExecData,
) error {
	resp, err := e.appDeploymentService.Deploy(ctx, db, &appdeploymentservice.AppDeploymentReq{
		TaskExecData: execData,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	// A deployment that ended done runs an image nobody has scanned yet: its
	// scan is queued once the deployment is committed. It never delays or fails
	// the deployment.
	if resp != nil && resp.Deployment != nil && resp.Deployment.Status == base.DeploymentStatusDone {
		appID := resp.Deployment.AppID
		execData.OnPostTx(func() { e.queueScan(context.Background(), appID) }) //nolint:contextcheck
	}
	return nil
}

// queueScan records the scan of a deployed app and schedules it, when image
// scanning is on. HivePaaS's own apps are not scanned: CI scans its images.
func (e *Executor) queueScan(ctx context.Context, appID string) {
	app, err := e.appRepo.GetByID(ctx, e.db, "", appID, bunex.SelectRelation("Project",
		bunex.SelectColumns("id", "key")))
	if err != nil {
		e.logger.Errorf("image scan after deployment: failed to load app %s: %v", appID, err)
		return
	}
	if app.Project != nil && app.Project.Key == base.HivepaasProjectKey {
		return
	}
	recorded, err := e.imageScanService.RecordScan(ctx, e.db, &imagescanservice.RecordScanReq{
		App: app, Trigger: base.ImageScanTriggerDeploy,
	})
	if err != nil {
		if !errors.Is(err, hperrors.ErrImageScanOff) {
			e.logger.Errorf("image scan after deployment: failed to record the scan of app %s: %v", appID, err)
		}
		return
	}
	if recorded.Pending {
		return
	}
	if err = e.taskQueue.ScheduleTask(ctx, recorded.Task); err != nil {
		e.logger.Errorf("image scan after deployment: failed to schedule the scan of app %s: %v", appID, err)
	}
}
