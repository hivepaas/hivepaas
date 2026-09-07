package hpappsettingsuc

import (
	"context"
	"net/netip"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reqinfo"
)

// ensureStillReachable refuses a routing change that would leave the caller with
// no way back in.
//
// HivePaaS configures the proxy that serves HivePaaS. Every other setting in the
// product can be fixed from the dashboard afterwards; these cannot, because the
// dashboard is what stops answering. The way back is a shell on a manager node,
// which the person making the change may not have.
//
// The invariant is deliberately narrow, so it refuses only what it can actually
// prove: after this change there must be at least one enabled domain that admits
// the address this request came from. It says nothing about DNS, certificates or
// anything else outside this setting - it only rules out the two shapes that are
// decided here and are unrecoverable.
func ensureStillReachable(ctx context.Context, settings *entity.AppRoutingSettings) error {
	enabled := enabledDomains(settings)
	if len(enabled) == 0 {
		// Every router label for the app is rebuilt from this list, so an empty one
		// leaves the service with no route at all - reachable by nothing, on any
		// address.
		return hperrors.Wrap(hperrors.ErrRoutingNoEnabledDomain)
	}

	clientIP := callerIP(ctx)
	if clientIP == "" {
		// Without knowing where the caller is, the check below cannot be made, and
		// an unverifiable answer is not one to act on when being wrong locks
		// somebody out of their own installation.
		return hperrors.Wrap(hperrors.ErrRoutingCallerUnknown)
	}

	for _, domain := range enabled {
		if domainAdmits(domain, clientIP) {
			return nil
		}
	}

	// The address is reported back because it is the thing most likely to be
	// surprising: behind a misconfigured proxy this is the proxy's address rather
	// than the operator's, and seeing it is what makes that obvious.
	return hperrors.Wrap(hperrors.ErrRoutingWouldLockYouOut).
		WithParam("ClientIP", clientIP)
}

func enabledDomains(settings *entity.AppRoutingSettings) []*entity.AppDomain {
	if settings == nil {
		return nil
	}
	res := make([]*entity.AppDomain, 0, len(settings.Domains))
	for _, domain := range settings.Domains {
		if domain != nil && domain.Enabled && domain.Domain != "" {
			res = append(res, domain)
		}
	}
	return res
}

// callerIP is the address this request came from, as the app resolved it.
//
// It is the app's answer, not Traefik's. The two agree when the proxy settings
// describe the deployment correctly and diverge when they do not - which is why
// the refusal reports the address it used, and why GET /system/hivepaas/request-info
// exists to check it.
func callerIP(ctx context.Context) string {
	info := reqinfo.From(ctx)
	if info == nil {
		return ""
	}
	return info.ClientIP
}

// domainAdmits reports whether a domain would still let this address through.
//
// Only the IP allowlist is considered. A rate limit slows a caller down but does
// not shut them out, and a wrong password is not a routing problem - the allowlist
// is the one setting here that answers "no" permanently and silently.
func domainAdmits(domain *entity.AppDomain, clientIP string) bool {
	cfg := domain.ClientConfig
	if cfg == nil || !cfg.Enabled || len(cfg.AllowedIPs) == 0 {
		return true
	}
	return ipAllowed(cfg.AllowedIPs, clientIP)
}

// ipAllowed reports whether an address falls inside any of the allowed entries,
// each of which is a single address or a CIDR block.
func ipAllowed(allowedIPs []string, clientIP string) bool {
	addr, err := netip.ParseAddr(clientIP)
	if err != nil {
		return false
	}

	for _, allowed := range allowedIPs {
		if single, err := netip.ParseAddr(allowed); err == nil {
			if single == addr {
				return true
			}
			continue
		}
		if prefix, err := netip.ParsePrefix(allowed); err == nil && prefix.Contains(addr) {
			return true
		}
	}
	return false
}
