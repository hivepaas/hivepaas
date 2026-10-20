package managedserviceimpl

import (
	"context"
	"strings"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// stopCheckInterval is how often a fenced database's tasks are looked at.
const stopCheckInterval = time.Second

func (s *service) RunCommand(ctx context.Context, db database.Tx, app *entity.App,
	command *entity.CommandTemplate) (*int, error) {
	resp, err := s.schedJobExecService.RunCommand(ctx, db, &schedjobexecservice.RunCommandReq{
		TaskExecData: &queue.TaskExecData{Task: &entity.Task{ID: gofn.Must(ulid.NewStringULID())},
			LogStore: tasklog.NewNullStore()},
		App:     app,
		Command: command,
	})
	if resp == nil {
		return nil, hperrors.Wrap(err)
	}
	return resp.ExitCode, hperrors.Wrap(err)
}

func (s *service) StopAndWait(ctx context.Context, app *entity.App, wait time.Duration) (bool, error) {
	if err := s.appService.SetAppRunning(ctx, app, false); err != nil {
		return false, hperrors.Wrap(err)
	}
	if app.ServiceID == "" {
		return true, nil
	}
	waitCtx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()
	stopped, err := s.dockerManager.ServiceWaitUntilStopped(waitCtx, app.ServiceID, stopCheckInterval)
	if err != nil && waitCtx.Err() != nil {
		return false, nil
	}
	return stopped, hperrors.Wrap(err)
}

func (s *service) Start(ctx context.Context, app *entity.App) error {
	return hperrors.Wrap(s.appService.SetAppRunning(ctx, app, true))
}

// aliasUpdateRetries is how many times a service update that meets a newer
// version of the service is tried again.
const aliasUpdateRetries = 3

func (s *service) MoveAlias(ctx context.Context, from, to *entity.App, alias string) error {
	if from != nil && from.ServiceID != "" {
		err := s.dockerManager.ServiceUpdateFunc(ctx, from.ServiceID, nil,
			func(_ int, service *swarm.Service) (bool, error) {
				return appservice.RemoveNetworkAlias(&service.Spec, alias), nil
			}, aliasUpdateRetries, 0)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}
	if to.ServiceID == "" {
		return nil
	}
	return hperrors.Wrap(s.dockerManager.ServiceUpdateFunc(ctx, to.ServiceID, nil,
		func(_ int, service *swarm.Service) (bool, error) {
			return appservice.AddNetworkAlias(&service.Spec, to.Key, alias), nil
		}, aliasUpdateRetries, 0))
}

func (s *service) AllowReplicas(ctx context.Context, db database.Tx, app *entity.App, allow *entity.CommandTemplate,
	password string) error {
	resp, err := s.schedJobExecService.RunCommand(ctx, db, &schedjobexecservice.RunCommandReq{
		TaskExecData: &queue.TaskExecData{Task: &entity.Task{ID: gofn.Must(ulid.NewStringULID())},
			LogStore: tasklog.NewNullStore()},
		App:     app,
		Command: allow,
		Stdin:   strings.NewReader(password + "\n"),
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	// It did not start: the database's service runs no container.
	if resp == nil || resp.ExitCode == nil {
		return hperrors.Wrap(hperrors.ErrActiveContainerNotFound).WithParam("App", app.Name)
	}
	if *resp.ExitCode != 0 {
		return hperrors.Wrap(hperrors.ErrActionFailed).
			WithExtraDetail("the database could not let replicas in: it exited with %d", *resp.ExitCode)
	}
	return nil
}
