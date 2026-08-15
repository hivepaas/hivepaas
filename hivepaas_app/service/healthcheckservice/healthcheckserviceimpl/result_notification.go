package healthcheckserviceimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/strutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
)

func (s *service) sendNotification(
	ctx context.Context,
	db database.IDB,
	data *healthcheckData,
) (err error) {
	periodicJob := data.PeriodicSetting.MustAsPeriodicJob()
	notifConfig := periodicJob.Notification
	if notifConfig == nil {
		return nil
	}

	notification, err := s.notificationService.GetNotificationForEvent(ctx, db,
		data.Scope, notifConfig.BaseEventNotification, data.Task.IsDone(), data.RefObjects)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if notification == nil {
		return nil
	}
	if notifConfig.MinSendInterval > 0 {
		notification.MinSendInterval = notifConfig.MinSendInterval
	}

	s.buildNotificationMsgData(data)
	req := &notificationservice.TaskResultNotificationReq{
		ActionSucceeded: data.Task.IsDone(),
		Scope:           data.Scope,
		RefObjects:      data.RefObjects,

		Notification: notification,
		TemplateName: notificationservice.TemplateHealthcheckNotification,
		TemplateData: data.NotifMsgData,
	}
	if data.LastHealthcheckState != nil {
		req.LastEvent = string(data.LastHealthcheckState.State)
		req.LastSendTs = data.LastHealthcheckState.LastNotifTs
	}

	resp, err := s.notificationService.NotifyForTaskResult(ctx, db, req)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if resp.HasSend() {
		data.LastNotifSendTs = resp.SendTs
	}

	return nil
}

func (s *service) buildNotificationMsgData(
	data *healthcheckData,
) {
	healthcheck := data.Healthcheck
	isSucceeded := data.Task.IsDone()
	msgData := &notificationservice.TemplateDataHealthcheck{
		BaseTemplateData: notificationservice.BaseTemplateData{
			Title: s.notificationService.BuildTitlePrefixForScope(data.Scope) +
				gofn.If(isSucceeded, " Healthcheck succeeded", " Healthcheck failed"),
		},
		Succeeded:       isSucceeded,
		HealthcheckName: data.PeriodicSetting.Name,
		HealthcheckType: healthcheck.HealthcheckType,
		StartedAt:       data.Task.StartedAt.Truncate(time.Second),
		Duration:        data.Task.GetDuration().Truncate(time.Millisecond),
		Retries:         data.Task.Config.Retry,
	}

	if project := data.Scope.GetProject(); project != nil {
		msgData.ProjectName = project.Name
	}
	if app := data.Scope.GetApp(); app != nil {
		msgData.AppName = app.Name
	}
	msgData.DashboardLink = config.Current.DashboardPeriodicTaskDetailsURL(data.Scope.GetBaseURLPath(),
		data.PeriodicSetting.ID, data.Task.ID)

	taskOutput, _ := data.Task.OutputAsPeriodicJob()
	output := taskOutput.Healthcheck
	if output.REST != nil && healthcheck.REST != nil {
		input := healthcheck.REST
		maxLen := 100
		pad := "..."
		if output.REST.ReturnCode != 0 {
			msgData.Expect = fmt.Sprintf("Status code = %v",
				gofn.StringJoinBy(input.ReturnCode, ", ", strconv.Itoa))
			msgData.Actual = fmt.Sprintf("Status code = %v", output.REST.ReturnCode)
		}
		if output.REST.ReturnText != "" && input.ReturnText != nil {
			expectStr := input.ReturnText.Exact
			if input.ReturnText.Regex != "" {
				expectStr = "Regex: " + input.ReturnText.Regex
			}
			msgData.Expect = strutil.CutShort(expectStr, maxLen, pad)
			msgData.Actual = strutil.CutShort(output.REST.ReturnText, maxLen, pad)
		}
		if output.REST.ReturnText != "" && input.ReturnJSON != nil {
			var expectStr string
			if input.ReturnJSON.Exact != nil {
				expectBytes, _ := json.Marshal(input.ReturnJSON.Exact)
				expectStr = "JSON Exact: " + reflectutil.UnsafeBytesToStr(expectBytes)
			} else if input.ReturnJSON.Contain != nil {
				expectBytes, _ := json.Marshal(input.ReturnJSON.Contain)
				expectStr = "JSON Contain: " + reflectutil.UnsafeBytesToStr(expectBytes)
			}
			msgData.Expect = strutil.CutShort(expectStr, maxLen, pad)
			msgData.Actual = strutil.CutShort(output.REST.ReturnText, maxLen, pad)
		}
	}
	if output.GRPC != nil && healthcheck.GRPC != nil {
		if output.GRPC.ReturnStatus != 0 {
			msgData.Expect = fmt.Sprintf("Status = %v", healthcheck.GRPC.ReturnStatus)
			msgData.Actual = fmt.Sprintf("Status = %v", output.GRPC.ReturnStatus)
		}
	}

	data.NotifMsgData = msgData
}
