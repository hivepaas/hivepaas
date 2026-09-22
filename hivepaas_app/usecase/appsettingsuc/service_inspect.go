package appsettingsuc

import (
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// requireAppService turns a service that is not there into an answer.
//
// clusterservice.ServiceInspect returns (nil, nil) for an app whose ServiceID is
// empty, which is a real state: an app created but never deployed, or one whose
// service somebody removed. Every settings screen here reads the live spec, and
// each of them dereferenced that nil - so the screen answered 500 with a panic
// instead of saying what was wrong.
func requireAppService(service *swarm.Service, appID string) (*swarm.Service, error) {
	if service == nil {
		return nil, hperrors.Wrap(hperrors.ErrAppServiceUnavailable).
			WithMsgLog("app %v has no docker service to read its settings from", appID)
	}
	return service, nil
}
