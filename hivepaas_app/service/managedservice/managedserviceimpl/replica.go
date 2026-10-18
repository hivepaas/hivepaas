package managedserviceimpl

import (
	"context"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

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
