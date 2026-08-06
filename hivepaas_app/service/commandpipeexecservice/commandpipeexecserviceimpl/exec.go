package commandpipeexecserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/funcutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/commandpipeexecservice"
)

type execData struct {
	*commandpipeexecservice.CommandPipeExecReq
}

func (s *service) CommandPipeExec(
	ctx context.Context,
	db database.IDB,
	req *commandpipeexecservice.CommandPipeExecReq,
) (_ *commandpipeexecservice.CommandPipeExecResp, err error) {
	defer funcutil.EnsureNoPanic(&err)

	if len(req.CommandPipes) == 0 {
		return &commandpipeexecservice.CommandPipeExecResp{}, nil
	}

	data := &execData{
		CommandPipeExecReq: req,
	}
	if req.LogStore == nil {
		req.LogStore = tasklog.NewNullStore()
	}
	err = s.loadCommandPipeData(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = s.waitUntilAppsRunning(ctx, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	for _, pipeSetting := range req.CommandPipes {
		err = s.commandPipeExec(ctx, db, pipeSetting, data)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	return &commandpipeexecservice.CommandPipeExecResp{}, nil
}

func (s *service) loadCommandPipeData(
	ctx context.Context,
	db database.IDB,
	data *execData,
) (err error) {
	err = s.settingService.LoadRefObjects(ctx, db, &data.RefObjects, nil, true, data.CommandPipes...)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
