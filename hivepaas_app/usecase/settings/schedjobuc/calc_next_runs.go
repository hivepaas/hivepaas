package schedjobuc

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
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
		return nil, apperrors.New(err)
	}

	nextRuns, err := sched.CalcNextRuns(initTime, req.Count)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &schedjobdto.CalcNextRunsResp{
		Data: nextRuns,
	}, nil
}
