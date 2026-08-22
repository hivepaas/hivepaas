package hpappserviceimpl

import (
	"context"
	"path/filepath"
	"time"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

const (
	authRateLimitAverage   = 20
	authRateLimitBurst     = 30
	authRateLimitMaxFlight = 10
)

func (s *service) GetHpAppSwarmService(ctx context.Context) (*swarm.Service, error) {
	service, err := s.dockerManager.ServiceGetByName(ctx, base.HivepaasAppServiceName, false)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return service, nil
}

func (s *service) RestartHpAppSwarmService(ctx context.Context) error {
	service, err := s.GetHpAppSwarmService(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}

	service.Spec.TaskTemplate.ForceUpdate++
	_, err = s.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *service) GetHpAppTasks(ctx context.Context) ([]swarm.Task, error) {
	service, err := s.GetHpAppSwarmService(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := s.dockerManager.ServiceTaskList(ctx, service.ID, nil)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return resp.Items, nil
}

func (s *service) SetupHttpSettingsDefault(
	httpSettings *entity.AppRoutingSettings,
) {
	for _, domain := range httpSettings.Domains {
		domain.ContainerPort = config.Current.HTTPServer.Port
		domain.ForceHttps = true
		domain.CompressionConfig = &entity.HTTPCompressionConfig{
			Enabled:         true,
			MinResponseBody: unit.KB, // 1kb
			DefaultEncoding: "br",    // brotli
		}

		var authPathConfig *entity.HTTPPathConfig
		authPath := filepath.Join(config.Current.HTTPServer.BasePath, "auth")
		for _, path := range domain.Paths {
			if path.Path == authPath {
				authPathConfig = path
				break
			}
		}
		if authPathConfig == nil {
			authPathConfig = &entity.HTTPPathConfig{
				Enabled: true,
				Path:    authPath,
				Mode:    base.HTTPPathModePrefix,
			}
			domain.Paths = append(domain.Paths, authPathConfig)
		}
		authPathConfig.RateLimitConfig = &entity.HTTPRateLimitConfig{
			Enabled:        true,
			Average:        authRateLimitAverage,
			Period:         timeutil.Duration(time.Minute),
			Burst:          authRateLimitBurst,
			MaxInFlightReq: authRateLimitMaxFlight,
		}
	}
}
