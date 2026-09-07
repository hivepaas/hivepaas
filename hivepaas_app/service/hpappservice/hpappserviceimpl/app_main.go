package hpappserviceimpl

import (
	"context"
	"path/filepath"
	"time"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

// apiRateLimits caps the request rate on the API paths where a high rate is
// never legitimate, keyed by their path under the API base path.
//
// Deliberately not a blanket limit over the whole API. The dashboard polls -
// /system/tasks/:id/status during a deployment above all - so a number tight
// enough to matter on the configuration endpoints would break the pages that
// poll. That is also why /system/hivepaas is listed rather than /system: Traefik
// ranks path rules by length, so the longer prefix catches only the subtree that
// carries the app secret, the security settings and the restart.
//
// Two properties of the underlying Traefik middleware shape these numbers. It
// counts per source IP, in memory, per instance: with N Traefik replicas the real
// ceiling is N times what is written here, and behind a proxy whose address is
// not among the entrypoint's trusted IPs every request appears to come from the
// proxy, which turns a per-caller limit into one ceiling shared by everybody. See
// traefikservice.ApplyTrustedIPsToWebEntrypoints.
//
// None of this is a defense against someone guessing a secret from an
// authenticated session - they can change address, and 20 a minute is still tens
// of thousands a day. That belongs with the actor, and lives in the usecase; see
// appSecretBackoff in hpappsettingsuc.
//
//nolint:mnd // tuned limits, not arithmetic - the numbers are the content here
var apiRateLimits = []struct {
	SubPath        string
	Average        int
	Burst          int
	MaxInFlightReq int
}{
	// Login, password reset, SSO callback: unauthenticated, and worth guessing at.
	{SubPath: "auth", Average: 20, Burst: 30, MaxInFlightReq: 10},

	// The app's own configuration. Reached from one settings page a handful of
	// requests at a time, and every endpoint on it is expensive or dangerous.
	{SubPath: "system/hivepaas", Average: 30, Burst: 20, MaxInFlightReq: 5},

	// The traefik's configuration
	{SubPath: "system/traefik", Average: 30, Burst: 20, MaxInFlightReq: 5},

	// The system configuration
	{SubPath: "system/settings", Average: 30, Burst: 20, MaxInFlightReq: 10},

	// Nodes, volumes, networks: infrastructure changes, not a polled view.
	{SubPath: "cluster", Average: 60, Burst: 40, MaxInFlightReq: 10},
}

func (s *service) GetHpAppSwarmService(ctx context.Context) (*swarm.Service, error) {
	service, err := s.dockerManager.ServiceGetByName(ctx, base.HivepaasAppServiceName, false)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return service, nil
}

func (s *service) RestartHpAppSwarmService(ctx context.Context) error {
	service, err := s.GetHpAppSwarmService(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}

	service.Spec.TaskTemplate.ForceUpdate++
	_, err = s.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *service) GetHpAppTasks(ctx context.Context) ([]swarm.Task, error) {
	service, err := s.GetHpAppSwarmService(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := s.dockerManager.ServiceTaskList(ctx, service.ID, nil)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp.Items, nil
}

func (s *service) SetupRoutingSettingsDefault(
	routingSettings *entity.AppRoutingSettings,
) {
	cfg := config.Current()

	for _, domain := range routingSettings.Domains {
		domain.ContainerPort = cfg.HTTPServer.Port
		domain.ForceHttps = true
		domain.CompressionConfig = &entity.HTTPCompressionConfig{
			Enabled:         true,
			MinResponseBody: unit.KB, // 1kb
			DefaultEncoding: "br",    // brotli
		}

		for _, limit := range apiRateLimits {
			pathCfg := ensurePathConfig(domain, filepath.Join(cfg.HTTPServer.BasePath, limit.SubPath))
			pathCfg.RateLimitConfig = &entity.HTTPRateLimitConfig{
				Enabled:        true,
				Average:        limit.Average,
				Period:         timeutil.Duration(time.Minute),
				Burst:          limit.Burst,
				MaxInFlightReq: limit.MaxInFlightReq,
			}
		}
	}
}

// ensurePathConfig returns the domain's config for a path, adding one if that
// path is not configured yet.
//
// An existing entry is reused rather than replaced. This function runs on every
// routing settings update, not only at first setup, so replacing would silently
// drop whatever else the operator had put on that path - a basic auth, a header
// rule - each time they saved anything.
func ensurePathConfig(domain *entity.AppDomain, path string) *entity.HTTPPathConfig {
	for _, pathCfg := range domain.Paths {
		if pathCfg.Path == path {
			return pathCfg
		}
	}

	pathCfg := &entity.HTTPPathConfig{
		Enabled: true,
		Path:    path,
		Mode:    base.HTTPPathModePrefix,
	}
	domain.Paths = append(domain.Paths, pathCfg)
	return pathCfg
}
