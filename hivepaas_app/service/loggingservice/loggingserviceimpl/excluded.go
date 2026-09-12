package loggingserviceimpl

import (
	"sort"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
)

// excludedApps lists the apps whose logs will not reach the backend usable.
//
// An app with no service yet, or whose service is gone, is skipped: there is
// nothing running to collect from, and saying so is someone else's job.
func excludedApps(apps []*entity.App, services []swarm.Service) []loggingservice.ExcludedApp {
	byID := make(map[string]*swarm.Service, len(services))
	for i := range services {
		byID[services[i].ID] = &services[i]
	}

	out := []loggingservice.ExcludedApp{}
	for _, app := range apps {
		svc := byID[app.ServiceID]
		if app.ServiceID == "" || svc == nil {
			continue
		}
		driver := svc.Spec.TaskTemplate.LogDriver
		driverName := ""
		if driver != nil {
			driverName = driver.Name
		}

		switch {
		case !appservice.IsLogDriverCollectible(driver):
			out = append(out, loggingservice.ExcludedApp{
				AppID: app.ID, Name: app.Name, Driver: driverName,
				Reason: loggingservice.ExcludedReasonDriverUnreadable,
			})
		case !appservice.HasLogIdentity(driver, svc.Spec.TaskTemplate.ContainerSpec, app.ID):
			out = append(out, loggingservice.ExcludedApp{
				AppID: app.ID, Name: app.Name, Driver: driverName,
				Reason: loggingservice.ExcludedReasonIdentityMissing,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
