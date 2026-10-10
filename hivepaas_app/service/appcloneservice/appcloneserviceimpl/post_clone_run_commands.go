package appcloneserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/commandpipeexecservice"
)

// cloneCommandTaskFindRetryMax is how many times, two seconds apart, the
// commands run after a clone look for the clone's container of its own spec.
const cloneCommandTaskFindRetryMax = 30

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
		// The clone's container was replaced a moment ago - its placeholder by
		// its own - and is waited for longer than a running app's.
		TaskFindRetryMax: cloneCommandTaskFindRetryMax,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}
