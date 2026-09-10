package taskuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/taskuc/taskdto"
)

func (uc *UC) ListTargetObjects(
	ctx context.Context,
	auth *basedto.Auth,
	req *taskdto.ListTargetObjectsReq,
) (*taskdto.ListTargetObjectsResp, error) {
	jobs, pagingMeta, err := uc.settingRepo.List(ctx, uc.db, req.Scope, &req.Paging,
		bunex.SelectWhereIn("setting.type IN (?)", base.SettingTypeSchedJob, base.SettingTypePeriodicJob),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp := make([]*taskdto.TargetObjectResp, 0, len(jobs))
	for _, obj := range jobs {
		objResp := &taskdto.TargetObjectResp{
			ID:   obj.ID,
			Name: obj.Name,
		}
		switch obj.Type { //nolint:exhaustive
		case base.SettingTypeSchedJob:
			objResp.Type = base.TaskTargetTypeSchedJob
		case base.SettingTypePeriodicJob:
			objResp.Type = base.TaskTargetTypePeriodicJob
		default:
			objResp.Type = base.TaskTargetType(obj.Type)
		}
		resp = append(resp, objResp)
	}

	return &taskdto.ListTargetObjectsResp{
		Meta: &basedto.ListMeta{Page: pagingMeta},
		Data: resp,
	}, nil
}
