package appdeploymentservice

import (
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
)

type DeploymentData struct {
	*queue.TaskExecData
	Project          *entity.Project
	App              *entity.App
	Deployment       *entity.Deployment
	DeploymentOutput *entity.AppDeploymentOutput
	Step             string
	LogStore         *applog.Store
	NotifMsgData     *notificationservice.TemplateDataAppDeployment
}
