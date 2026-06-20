package taskuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/taskuc/taskdto"
)

func (uc *UC) ListTask(
	ctx context.Context,
	auth *basedto.Auth,
	req *taskdto.ListTaskReq,
) (*taskdto.ListTaskResp, error) {
	targetIDs := req.TargetID

	if req.JobName != "" {
		var settingType base.SettingType
		switch req.JobName {
		case base.SystemJobNameDataBackup:
			settingType = base.SettingTypeSystemBackup
		case base.SystemJobNameDataCleanup:
			settingType = base.SettingTypeSystemCleanup
		case base.SystemJobNameSslRenewal:
			settingType = base.SettingTypeSSLRenewal
		default:
			return nil, apperrors.NewNotFound(apperrors.Fmt("Job setting of '%v'", req.JobName))
		}
		setting, err := uc.settingRepo.GetSingle(ctx, uc.db, base.NewObjectScopeGlobal(), settingType, false,
			bunex.SelectColumns("id"),
		)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		targetIDs = append(targetIDs, setting.ID)
	}

	listResp, err := uc.taskService.ListTask(ctx, uc.db, &taskservice.ListTaskReq{
		Scope:     base.NewObjectScopeGlobal(),
		TargetIDs: targetIDs,
		Statuses:  req.Status,
		Search:    req.Search,
		Paging:    req.Paging,
		ExtraSelectOpts: []bunex.SelectQueryOption{
			bunex.SelectRelation("TargetJob",
				bunex.SelectColumns("id", "type", "kind", "name", "status"),
			),
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := taskdto.TransformTasks(listResp.Tasks, listResp.TaskInfoMap)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &taskdto.ListTaskResp{
		Meta: &basedto.ListMeta{Page: listResp.PagingMeta},
		Data: resp,
	}, nil
}
