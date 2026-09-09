package taskuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/taskuc/taskdto"
)

func (uc *UC) ListTaskType(
	ctx context.Context,
	auth *basedto.Auth,
	req *taskdto.ListTaskTypeReq,
) (*taskdto.ListTaskTypeResp, error) {
	var taskTypes []base.TaskType
	switch req.Scope.ScopeType {
	case base.ObjectScopeGlobal:
		taskTypes = base.AllGlobalTaskTypes
	case base.ObjectScopeProject, base.ObjectScopeProjectEnv:
		taskTypes = base.AllProjectTaskTypes
	case base.ObjectScopeApp:
		taskTypes = base.AllAppTaskTypes
	case base.ObjectScopeHivepaas:
		taskTypes = base.AllHivepaasTaskTypes
	case base.ObjectScopeUser:
		taskTypes = base.AllUserTaskTypes
	}

	return &taskdto.ListTaskTypeResp{Data: taskTypes}, nil
}
