package clustersecretserviceimpl

import (
	"context"
	"slices"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// mountedIDs are the secrets and configs setting mounts hold, whichever app's.
func (s *service) mountedIDs(ctx context.Context) (map[string]bool, error) {
	ids := map[string]bool{}
	secrets, err := s.dockerManager.SecretList(ctx, func(opts *client.SecretListOptions) {
		docker.FilterAdd(&opts.Filters, "label", settingmountservice.LabelEntry)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, secret := range secrets.Items {
		ids[secret.ID] = true
	}
	configs, err := s.dockerManager.ConfigList(ctx, func(opts *client.ConfigListOptions) {
		docker.FilterAdd(&opts.Filters, "label", settingmountservice.LabelEntry)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, config := range configs.Items {
		ids[config.ID] = true
	}
	return ids, nil
}

// makeRoomForSecret frees target for an ordinary secret: a setting mount there
// steps aside, as it does when the engine applies (an ordinary file wins). It
// reports whether an ordinary secret already has the path.
func makeRoomForSecret(spec *swarm.ContainerSpec, target string, mounted map[string]bool) (taken bool) {
	spec.Secrets = slices.DeleteFunc(spec.Secrets, func(ref *swarm.SecretReference) bool {
		if ref.File == nil || settingmountservice.SecretTarget(ref.File.Name) != target {
			return false
		}
		if mounted[ref.SecretID] {
			return true
		}
		taken = true
		return false
	})
	return taken
}

// makeRoomForConfig is makeRoomForSecret for a config.
func makeRoomForConfig(spec *swarm.ContainerSpec, target string, mounted map[string]bool) (taken bool) {
	spec.Configs = slices.DeleteFunc(spec.Configs, func(ref *swarm.ConfigReference) bool {
		if ref.File == nil || settingmountservice.ConfigTarget(ref.File.Name) != target {
			return false
		}
		if mounted[ref.ConfigID] {
			return true
		}
		taken = true
		return false
	})
	return taken
}
