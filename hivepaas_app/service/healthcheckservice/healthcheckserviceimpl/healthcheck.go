package healthcheckserviceimpl

import (
	"context"
	"fmt"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/funcutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/healthcheckservice"
)

type healthcheckData struct {
	*healthcheckservice.HealthcheckReq
	Output *entity.TaskHealthcheckOutput
}

func (s *service) Healthcheck(
	ctx context.Context,
	req *healthcheckservice.HealthcheckReq,
) (resp *healthcheckservice.HealthcheckResp, err error) {
	resp = &healthcheckservice.HealthcheckResp{}
	data := &healthcheckData{
		HealthcheckReq: req,
		Output:         &entity.TaskHealthcheckOutput{},
	}

	var testErr error
	defer func() {
		data.Task.Status = gofn.If(testErr == nil, base.TaskStatusDone, base.TaskStatusFailed)
		data.Task.EndedAt = timeutil.NowUTC()
		data.Task.MustSetOutput(data.Output)
	}()
	defer funcutil.EnsureNoPanic(&err)

	retries := 0
	startTime := time.Now()
	for {
		switch data.Healthcheck.HealthcheckType {
		case base.HealthcheckTypeREST:
			testErr = s.doHealthcheckREST(ctx, data)
		case base.HealthcheckTypeGRPC:
			testErr = s.doHealthcheckGRPC(ctx, data)
		default:
			testErr = apperrors.NewUnsupported(
				fmt.Sprintf("Healthcheck type '%v'", data.Healthcheck.HealthcheckType))
		}
		if testErr != nil {
			retries++
			if retries > data.Task.Config.MaxRetry {
				break
			}
			data.Task.Config.Retry = retries
			if data.Task.Config.RetryDelay > 0 {
				time.Sleep(data.Task.Config.RetryDelay.ToDuration())
			}
			if time.Since(startTime)+5*time.Second > data.Healthcheck.Interval.ToDuration() {
				break
			}
		} else {
			break
		}
	}

	return resp, err
}
