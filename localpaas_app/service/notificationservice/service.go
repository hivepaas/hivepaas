package notificationservice

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

type Service interface {
	GetNotificationForEvent(ctx context.Context, db database.IDB, scope *base.ObjectScope,
		eventSetting *entity.BaseEventNotification, eventSuccess bool, refObjects *entity.RefObjects) (
		*entity.Notification, error)
	GetDefaultNotification(ctx context.Context, db database.IDB, scope *base.ObjectScope,
		refObjects *entity.RefObjects, errorIfRefObjectsUnavail bool) (
		*entity.Notification, error)
	BuildTitlePrefix(project *entity.Project, app *entity.App, user *entity.User) string

	NotifyForTaskResult(ctx context.Context, db database.IDB, data *TaskResultNotificationReq) (
		*TaskResultNotificationResp, error)
}
