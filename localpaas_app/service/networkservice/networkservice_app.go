package networkservice

import (
	"context"

	"github.com/docker/docker/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/entity"
)

func (s *networkService) UpdateAppGlobalRoutingNetwork(
	ctx context.Context,
	_ *entity.App,
	service *swarm.Service,
	dbHttpSettings *entity.Setting,
) error {
	httpSettings := dbHttpSettings.MustAsAppHttpSettings()
	globalNetworkID, err := s.FindGlobalRoutingNetworkID(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}

	spec := &service.Spec
	networks := make([]swarm.NetworkAttachmentConfig, 0, len(spec.TaskTemplate.Networks)+1)
	for _, net := range spec.TaskTemplate.Networks {
		if httpSettings.ExposePublicly && (net.Target == GlobalRoutingNetwork || net.Target == globalNetworkID) {
			return nil // app is attached to the external net already
		}
		if !httpSettings.ExposePublicly && (net.Target == GlobalRoutingNetwork || net.Target == globalNetworkID) {
			continue
		}
		networks = append(networks, net)
	}
	if httpSettings.ExposePublicly {
		networks = append(networks, swarm.NetworkAttachmentConfig{
			Target: globalNetworkID,
		})
	}
	spec.TaskTemplate.Networks = networks

	return nil
}
