package appcloneserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/commandpipeexecservice"
)

func (s *service) runCommands(
	ctx context.Context,
	db database.IDB,
	data *appCloneData,
) (err error) {
	if data.CloneSettings == nil || len(data.CloneSettings.CommandPipes) == 0 {
		return nil
	}
	if data.SrcApp.ServiceID == "" || data.DestApp.ServiceID == "" {
		return nil
	}

	commandPipeSettings := make([]*entity.Setting, 0, len(data.CloneSettings.CommandPipes))
	for _, pipeObj := range data.CloneSettings.CommandPipes {
		if pipeObj == nil || pipeObj.ID == "" {
			continue
		}
		pipeSetting := data.RefObjects.RefSettings[pipeObj.ID]
		if pipeSetting != nil {
			commandPipeSettings = append(commandPipeSettings, pipeSetting)
		}
	}
	if len(commandPipeSettings) == 0 {
		return nil
	}

	_, err = s.commandPipeExecService.CommandPipeExec(ctx, db, &commandpipeexecservice.CommandPipeExecReq{
		TaskExecData: data.TaskExecData,
		CommandPipes: commandPipeSettings,
		SrcApp:       data.SrcApp,
		DestApp:      data.DestApp,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
