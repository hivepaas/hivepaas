package nginxservice

import (
	"context"
	"os"
	"path/filepath"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/config"
)

func (s *nginxService) ReloadNginxConfig(ctx context.Context, restartServiceOnFailure bool) error {
	err := s.reloadNginxConfig(ctx)
	if err == nil {
		return nil
	}
	if !restartServiceOnFailure {
		return apperrors.Wrap(err)
	}
	err = s.RestartNginxSwarmService(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *nginxService) reloadNginxConfig(ctx context.Context) error {
	service, err := s.dockerManager.ServiceGetByName(ctx, nginxServiceName)
	if err != nil {
		return apperrors.Wrap(err)
	}

	containers, err := s.dockerManager.ServiceContainerList(ctx, service.ID)
	if err != nil {
		return apperrors.Wrap(err)
	}

	containerIDs := make([]string, 0, len(containers))
	for _, container := range containers {
		containerIDs = append(containerIDs, container.ID)
	}
	if len(containerIDs) == 0 {
		return apperrors.NewNotFound("Nginx service")
	}

	errMap := s.dockerManager.ContainerKillMulti(ctx, containerIDs, "SIGHUP")
	for _, err := range errMap {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *nginxService) ResetNginxConfig(ctx context.Context) error {
	data, err := os.ReadFile("config/nginx/nginx.conf.template")
	if err != nil {
		return apperrors.Wrap(err)
	}

	confPath := filepath.Join(config.Current.DataPathNginxEtc(), "nginx.conf")
	err = os.WriteFile(confPath, data, defaultConfFileMode)
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = s.ReloadNginxConfig(ctx, true)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
