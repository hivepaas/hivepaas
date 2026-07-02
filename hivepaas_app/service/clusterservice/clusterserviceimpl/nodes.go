package clusterserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *service) IsMultiNode(ctx context.Context) (bool, error) {
	resp, err := s.dockerManager.SystemInfo(ctx)
	if err != nil {
		return false, apperrors.New(err)
	}
	return resp.Info.Swarm.Nodes > 1, nil
}
