package appdeploymentserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/service/appdeploymentservice"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
)

func (s *service) notifyForDeployment(
	ctx context.Context,
	db database.IDB,
	data *appdeploymentservice.DeploymentData,
) (err error) {
	notifConfig := data.Deployment.Settings.Notification
	if notifConfig == nil {
		return nil
	}

	notification, err := s.notificationService.GetNotificationForEvent(ctx, db,
		data.App.GetSettingScope(), notifConfig, data.Deployment.IsDone(), data.RefObjects)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if notification == nil {
		return nil
	}

	s.buildDeploymentNotifMsgData(data)
	_, err = s.notificationService.NotifyForTaskResult(ctx, db, &notificationservice.TaskResultNotificationReq{
		ActionSucceeded: data.Deployment.IsDone(),
		ScopeProject:    data.Project,
		ScopeApp:        data.App,
		RefObjects:      data.RefObjects,
		Notification:    notification,
		TemplateName:    notificationservice.TemplateAppDeploymentNotification,
		TemplateData:    data.NotifMsgData,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *service) buildDeploymentNotifMsgData(
	data *appdeploymentservice.DeploymentData,
) {
	deployment := data.Deployment
	isSucceeded := deployment.IsDone()
	msgData := &notificationservice.TemplateDataAppDeployment{
		BaseTemplateData: notificationservice.BaseTemplateData{
			Title: s.notificationService.BuildTitlePrefix(data.Project, data.App, nil) +
				gofn.If(isSucceeded, " Deployment succeeded", " Deployment failed"),
		},
		ProjectName:   data.Project.Name,
		AppName:       data.App.Name,
		Succeeded:     isSucceeded,
		Method:        deployment.Settings.ActiveMethod,
		StartedAt:     deployment.StartedAt,
		Duration:      deployment.GetDuration(),
		DashboardLink: config.Current.DashboardDeploymentDetailsURL(deployment.ID),
	}
	data.NotifMsgData = msgData

	switch deployment.Settings.ActiveMethod {
	case base.DeploymentMethodRepo:
		msgData.RepoURL = deployment.Settings.RepoSource.RepoURL
		msgData.RepoRef = deployment.Settings.RepoSource.RepoRef
		if deployment.Output != nil {
			msgData.CommitMsg = deployment.Output.CommitMessage
		}
	case base.DeploymentMethodImage:
		msgData.Image = deployment.Settings.ImageSource.Image
	}
}
