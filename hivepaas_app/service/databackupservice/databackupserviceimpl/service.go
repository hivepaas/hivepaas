package databackupserviceimpl

import (
	"context"
	"io"
	"time"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/databackupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/nodeexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/scopeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	backupRepoService   backupreposervice.Service
	schedJobExecService schedjobexecservice.Service

	// repoScope is the scope a repository's storage is resolved in: its own.
	repoScope func(ctx context.Context, db database.IDB, repo *entity.Setting) (*entity.ObjectScope, error)
	// findAppVolume is FindAppVolume, replaced in tests.
	findAppVolume func(ctx context.Context, db database.IDB, app *entity.App,
		volumeID string) (*databackupservice.AppVolume, error)

	// apps, hostRun and now are what a restore stops and starts an app with,
	// runs commands on a node's host with, and names a directory moved aside
	// with; tests replace them.
	apps    appControl
	hostRun func(ctx context.Context, nodeID, nodeLabel string, cmd ...string) (int, error)
	now     func() time.Time

	clusterService clusterservice.Service
	settingService settingservice.Service
	volumeService  volumeservice.Service
}

func New(
	dockerManager docker.Manager,
	appService appservice.Service,
	backupRepoService backupreposervice.Service,
	clusterService clusterservice.Service,
	nodeExecService nodeexecservice.Service,
	schedJobExecService schedjobexecservice.Service,
	scopeService scopeservice.Service,
	settingService settingservice.Service,
	volumeService volumeservice.Service,
) databackupservice.Service {
	svc := &service{
		backupRepoService:   backupRepoService,
		schedJobExecService: schedJobExecService,
		clusterService:      clusterService,
		settingService:      settingService,
		volumeService:       volumeService,
	}
	svc.repoScope = func(ctx context.Context, db database.IDB, repo *entity.Setting) (*entity.ObjectScope, error) {
		scope, err := scopeService.LoadObjectScope(ctx, db, repo.Scope, repo.ObjectID, true)
		return scope, hperrors.Wrap(err)
	}
	svc.findAppVolume = svc.FindAppVolume
	svc.apps = &swarmAppControl{dockerManager: dockerManager, appService: appService}
	svc.hostRun = func(ctx context.Context, nodeID, nodeLabel string, cmd ...string) (int, error) {
		resp, err := nodeExecService.ExecCommand(ctx, &nodeexecservice.CommandExecReq{
			NodeID: nodeID, NodeLabel: nodeLabel,
			CommandExecOpts: &nodeexecservice.CommandExecOpts{Command: cmd, Stdout: io.Discard, Stderr: io.Discard},
		})
		if err != nil {
			return 0, hperrors.Wrap(err)
		}
		return int(resp.ExitCode), nil
	}
	svc.now = time.Now
	return svc
}

// swarmAppControl stops and starts an app's swarm service.
type swarmAppControl struct {
	dockerManager docker.Manager
	appService    appservice.Service
}

func (c *swarmAppControl) Running(ctx context.Context, app *entity.App) (bool, error) {
	if app.ServiceID == "" {
		return false, nil
	}
	tasks, err := c.dockerManager.ServiceTaskList(ctx, app.ServiceID, []swarm.TaskState{swarm.TaskStateRunning})
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	return len(tasks.Items) > 0, nil
}

func (c *swarmAppControl) Stop(ctx context.Context, app *entity.App) error {
	if err := c.appService.SetAppRunning(ctx, app, false); err != nil {
		return hperrors.Wrap(err)
	}
	_, err := c.dockerManager.ServiceWaitUntilStopped(ctx, app.ServiceID, 0)
	return hperrors.Wrap(err)
}

func (c *swarmAppControl) Start(ctx context.Context, app *entity.App) error {
	return hperrors.Wrap(c.appService.SetAppRunning(ctx, app, true))
}
