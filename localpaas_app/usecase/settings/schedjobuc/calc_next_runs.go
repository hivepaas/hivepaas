package schedjobuc

import (
	"context"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/schedjobuc/schedjobdto"
)

func (uc *UC) CalcNextRuns(
	_ context.Context,
	_ *basedto.Auth,
	req *schedjobdto.CalcNextRunsReq,
) (*schedjobdto.CalcNextRunsResp, error) {
	initTime := req.InitialTime
	if initTime.IsZero() {
		initTime = timeutil.NowUTC()
	}
	initTime = initTime.Truncate(time.Second)

	sched := &entity.SchedJobSchedule{
		Interval:    req.Interval,
		CronExpr:    req.CronExpr,
		InitialTime: initTime,
	}
	if err := sched.IsValid(); err != nil {
		return nil, apperrors.Wrap(err)
	}

	nextRuns, err := sched.CalcNextRuns(initTime, req.Count)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &schedjobdto.CalcNextRunsResp{
		Data: nextRuns,
	}, nil
}
