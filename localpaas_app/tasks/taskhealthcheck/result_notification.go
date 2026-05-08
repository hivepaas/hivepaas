package taskhealthcheck

import (
	"context"
	"fmt"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity/cacheentity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/strutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
)

func (e *Executor) sendNotification(
	ctx context.Context,
	db database.IDB,
	data *taskData,
) (err error) {
	notifConfig := data.Healthcheck.Notification
	if notifConfig == nil {
		return nil
	}

	var scope *base.SettingScope
	switch {
	case data.App != nil:
		scope = data.App.GetSettingScope()
	case data.Project != nil:
		scope = data.Project.GetSettingScope()
	default:
		scope = base.NewSettingScopeGlobal()
	}

	notification, err := e.notificationService.GetNotificationForEvent(ctx, db,
		scope, notifConfig, data.Task.IsDone(), data.RefObjects)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if notification == nil {
		return nil
	}

	e.buildNotificationMsgData(data)
	req := &notificationservice.TaskResultNotificationReq{
		ActionSucceeded: data.Task.IsDone(),
		ScopeProject:    data.Project,
		ScopeApp:        data.App,
		RefObjects:      data.RefObjects,

		Notification: notification,
		TemplateName: notificationservice.TemplateHealthcheckNotification,
		TemplateData: data.NotifMsgData,
	}
	lastNotifSend := data.NotifEventMap[data.HealthcheckSetting.ID]
	if lastNotifSend != nil {
		req.LastSendEvent = lastNotifSend.Event
		req.LastSendTimestamp = lastNotifSend.Ts
		req.LastEmailSent = lastNotifSend.EmailSent
		req.LastSlackSent = lastNotifSend.SlackSent
		req.LastDiscordSent = lastNotifSend.DiscordSent
	}

	resp, err := e.notificationService.NotifyForTaskResult(ctx, db, req)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Update notification events in redis
	minSendingInterval := notification.MinSendInterval.ToDuration()
	if minSendingInterval > 0 {
		_ = e.notifEventRepo.Set(ctx, data.HealthcheckSetting.ID, &cacheentity.HealthcheckNotifEvent{
			Event:       gofn.If(req.ActionSucceeded, "success", "failure"),
			Ts:          timeutil.NowUTC(),
			EmailSent:   resp.EmailSent,
			SlackSent:   resp.SlackSent,
			DiscordSent: resp.DiscordSent,
		}, minSendingInterval)
	}

	return nil
}

func (e *Executor) buildNotificationMsgData(
	data *taskData,
) {
	isSucceeded := data.Task.IsDone()
	msgData := &notificationservice.TemplateDataHealthcheck{
		BaseTemplateData: notificationservice.BaseTemplateData{
			Title: e.notificationService.BuildTitlePrefix(data.Project, data.App, nil) +
				gofn.If(isSucceeded, " Healthcheck succeeded", " Healthcheck failed"),
		},
		Succeeded:       isSucceeded,
		HealthcheckName: data.HealthcheckSetting.Name,
		HealthcheckType: data.Healthcheck.HealthcheckType,
		StartedAt:       data.Task.StartedAt,
		Duration:        data.Task.GetDuration(),
		Retries:         data.Task.Config.Retry,
	}
	if data.Project != nil {
		msgData.ProjectName = data.Project.Name
	}
	if data.App != nil {
		msgData.AppName = data.App.Name
	}
	switch {
	case data.App != nil:
		msgData.DashboardLink = config.Current.DashboardAppHealthcheckDetailsURL(data.App.ID, data.App.ProjectID,
			data.HealthcheckSetting.ID, data.Task.ID)
	case data.Project != nil:
		msgData.DashboardLink = config.Current.DashboardProjectHealthcheckDetailsURL(data.Project.ID,
			data.HealthcheckSetting.ID, data.Task.ID)
	}

	output, _ := data.Task.OutputAsHealthcheck()
	if output.REST != nil && data.Healthcheck.REST != nil {
		input := data.Healthcheck.REST
		maxLen := 100
		pad := "..."
		if output.REST.ReturnCode != 0 {
			msgData.Expect = fmt.Sprintf("Status code = %v", input.ReturnCode)
			msgData.Actual = fmt.Sprintf("Status code = %v", output.REST.ReturnCode)
		}
		if output.REST.ReturnText != "" {
			msgData.Expect = fmt.Sprintf("Text = %v", strutil.CutShort(input.ReturnText, maxLen, pad))
			msgData.Actual = fmt.Sprintf("Text = %v", strutil.CutShort(output.REST.ReturnText, maxLen, pad))
		}
		if output.REST.ReturnJSON != "" {
			msgData.Expect = fmt.Sprintf("JSON = %v", strutil.CutShort(input.ReturnJSON, maxLen, pad))
			msgData.Actual = fmt.Sprintf("JSON = %v", strutil.CutShort(output.REST.ReturnJSON, maxLen, pad))
		}
	}
	if output.GRPC != nil && data.Healthcheck.GRPC != nil {
		if output.GRPC.ReturnStatus != 0 {
			msgData.Expect = fmt.Sprintf("Status = %v", data.Healthcheck.GRPC.ReturnStatus)
			msgData.Actual = fmt.Sprintf("Status = %v", output.GRPC.ReturnStatus)
		}
	}

	data.NotifMsgData = msgData
}
