package healthcheckserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/funcutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/healthcheckservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
)

type healthcheckData struct {
	*healthcheckservice.HealthcheckReq
	Output               *entity.TaskPeriodicHealthcheckOutput
	LastHealthcheckState *cacheentity.HealthcheckState
	NotifMsgData         *notificationservice.TemplateDataHealthcheck
	LastNotifSendTs      time.Time
}

func (s *service) Healthcheck(
	ctx context.Context,
	req *healthcheckservice.HealthcheckReq,
) (resp *healthcheckservice.HealthcheckResp, err error) {
	resp = &healthcheckservice.HealthcheckResp{}
	data := &healthcheckData{
		HealthcheckReq: req,
		Output:         &entity.TaskPeriodicHealthcheckOutput{},
	}

	var testErr error
	defer func() {
		data.Task.Status = gofn.If(testErr == nil, base.TaskStatusDone, base.TaskStatusFailed)
		data.Task.EndedAt = timeutil.NowUTC()
		if testErr != nil {
			_ = data.Task.AddRun(&entity.TaskRun{
				StartedAt: data.Task.StartedAt,
				EndedAt:   data.Task.EndedAt,
				Error:     apperrors.GetErrorDetail(testErr, ""),
			})
		}
		data.Task.MustSetOutput(&entity.TaskPeriodicOutput{Healthcheck: data.Output})
		// Calculate state transition to decide whether we need to save the task or not
		_ = s.calculateStateTransition(ctx, data)
		// Send notification
		_ = s.sendNotification(ctx, s.db, data)
		// Save state in cache
		_ = s.saveStateInCache(ctx, data)
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
			periodicJob := data.PeriodicSetting.MustAsPeriodicJob()
			if time.Since(startTime)+5*time.Second > periodicJob.Interval.ToDuration() {
				break
			}
		} else {
			break
		}
	}

	return resp, err
}

func (s *service) calculateStateTransition(
	ctx context.Context,
	data *healthcheckData,
) (err error) {
	lastState, err := s.healthcheckStateRepo.Get(ctx, data.PeriodicSetting.ID)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	data.LastHealthcheckState = lastState

	// If the result state changes from `success -> failure` or `failure -> success`
	stateTransition := true
	if lastState != nil {
		stateTransition = (lastState.State == base.HealthcheckStateSuccess) != (data.Task.Status == base.TaskStatusDone)
	}
	if stateTransition {
		data.SaveTask = true // tell the scheduler to save the task in DB
	}

	return nil
}

func (s *service) saveStateInCache(
	ctx context.Context,
	data *healthcheckData,
) (err error) {
	state := data.LastHealthcheckState
	if state == nil {
		state = &cacheentity.HealthcheckState{
			State: gofn.If(data.Task.Status == base.TaskStatusDone,
				base.HealthcheckStateSuccess, base.HealthcheckStateFailure),
		}
	}
	if !data.LastNotifSendTs.IsZero() {
		state.LastNotifTs = data.LastNotifSendTs
	}

	exp := timeutil.Day
	periodicJob := data.PeriodicSetting.MustAsPeriodicJob()
	if exp < periodicJob.Interval.ToDuration() {
		exp = periodicJob.Interval.ToDuration() + timeutil.Day
	}
	err = s.healthcheckStateRepo.Set(ctx, data.PeriodicSetting.ID, state, exp)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
