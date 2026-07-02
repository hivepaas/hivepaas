package hpappserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func (s *service) ReloadHpAppConfig(ctx context.Context) error {
	service, err := s.dockerManager.ServiceGetByName(ctx, base.HivepaasAppServiceName, false)
	if err != nil {
		return apperrors.New(err)
	}

	listResp, err := s.dockerManager.ServiceContainerList(ctx, service.ID)
	if err != nil {
		return apperrors.New(err)
	}

	containers := listResp.Items
	containerIDs := make([]string, 0, len(containers))
	for i := range containers {
		containerIDs = append(containerIDs, containers[i].ID)
	}

	errMap := s.dockerManager.ContainerKillMulti(ctx, containerIDs, "SIGHUP")
	for _, err := range errMap {
		return apperrors.New(err)
	}
	return nil
}
