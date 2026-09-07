package traefikserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// getProxySettings returns the proxy topology, loading it at most once.
//
// It is memoized on the apply that asked for it rather than on the service,
// deliberately. A process-wide copy would be read long after an operator edited
// the proxy settings, and there is no hook that reliably fires on every such
// edit - so the stale copy would keep telling Traefik to read a forwarded header
// position that no longer describes the deployment. Scoped to one apply, the
// value cannot outlive the operation that read it, and the load still happens
// only once no matter how many domains and paths ask.
func (data *appConfigData) getProxySettings() *entity.HivePaaSProxySettings {
	if data.proxySettings == nil {
		// Nothing loaded it, which means nothing needed it. An empty topology reads
		// as "no proxy", and that is the safe answer: it makes Traefik identify
		// callers by the address it is talking to instead of by a header.
		return &entity.HivePaaSProxySettings{}
	}
	return data.proxySettings
}

// loadProxySettings reads the proxy topology, if this apply has any use for it.
//
// Most applies configure neither a rate limit nor an IP allowlist, and then
// nobody asks who the caller is - so the query is skipped rather than paid for on
// every deployment.
func (s *service) loadProxySettings(
	ctx context.Context,
	db database.IDB,
	data *appConfigData,
) error {
	if !needsClientIPStrategy(data.RoutingSettings) {
		return nil
	}

	setting, err := s.settingRepo.GetSingle(ctx, db, nil, base.SettingTypeHivePaaSService, true)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return nil
		}
		return hperrors.Wrap(err)
	}
	if setting == nil {
		return nil
	}

	svcSettings, err := setting.AsHivePaaSService()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if svcSettings == nil {
		return nil
	}

	data.proxySettings = &svcSettings.ProxySettings
	return nil
}

// needsClientIPStrategy reports whether any middleware in these routing settings
// has to decide which address a request came from.
func needsClientIPStrategy(routingSettings *entity.AppRoutingSettings) bool {
	if routingSettings == nil || !routingSettings.ExposePublicly {
		return false
	}

	for _, domain := range routingSettings.Domains {
		if usesClientIP(domain.RateLimitConfig, domain.ClientConfig) {
			return true
		}
		for _, pathCfg := range domain.Paths {
			if usesClientIP(pathCfg.RateLimitConfig, pathCfg.ClientConfig) {
				return true
			}
		}
	}
	return false
}

func usesClientIP(rateLimit *entity.HTTPRateLimitConfig, client *entity.HTTPClientConfig) bool {
	if rateLimit != nil && rateLimit.Enabled {
		return true
	}
	return client != nil && client.Enabled && len(client.AllowedIPs) > 0
}
