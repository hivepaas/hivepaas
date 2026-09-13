package loggingserviceimpl

import (
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
)

func isAppExcluded(app *entity.App, svc *swarm.Service) loggingservice.ExcludedReason {
	if app == nil || app.ServiceID == "" || svc == nil {
		return ""
	}
	driver := svc.Spec.TaskTemplate.LogDriver
	switch {
	case !appservice.IsLogDriverCollectible(driver):
		return loggingservice.ExcludedReasonDriverUnreadable
	case !appservice.HasLogIdentity(driver, svc.Spec.TaskTemplate.ContainerSpec, app.ID):
		return loggingservice.ExcludedReasonIdentityMissing
	}
	return ""
}
