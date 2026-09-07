package hpappsettingsuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reqinfo"
)

func ctxFromIP(clientIP string) context.Context {
	return reqinfo.NewContext(context.Background(), &reqinfo.RequestInfo{ClientIP: clientIP})
}

func domainWithAllowlist(name string, enabled bool, allowedIPs ...string) *entity.AppDomain {
	domain := &entity.AppDomain{Enabled: enabled, Domain: name}
	if len(allowedIPs) > 0 {
		domain.ClientConfig = &entity.HTTPClientConfig{Enabled: true, AllowedIPs: allowedIPs}
	}
	return domain
}

func routing(domains ...*entity.AppDomain) *entity.AppRoutingSettings {
	return &entity.AppRoutingSettings{ExposePublicly: true, Domains: domains}
}

func TestEnsureStillReachable(t *testing.T) {
	ctx := ctxFromIP("203.0.113.7")

	t.Run("a domain with no allowlist always admits", func(t *testing.T) {
		assert.NoError(t, ensureStillReachable(ctx, routing(
			domainWithAllowlist("app.example.com", true))))
	})

	t.Run("an allowlist covering the caller", func(t *testing.T) {
		assert.NoError(t, ensureStillReachable(ctx, routing(
			domainWithAllowlist("app.example.com", true, "203.0.113.0/24"))))
	})

	t.Run("an allowlist naming the caller exactly", func(t *testing.T) {
		assert.NoError(t, ensureStillReachable(ctx, routing(
			domainWithAllowlist("app.example.com", true, "203.0.113.7"))))
	})

	// The case this exists for: valid settings, saved successfully, and nobody can
	// reach the dashboard afterwards.
	t.Run("an allowlist that shuts the caller out", func(t *testing.T) {
		err := ensureStillReachable(ctx, routing(
			domainWithAllowlist("app.example.com", true, "10.0.0.0/8")))
		assert.ErrorIs(t, err, hperrors.ErrRoutingWouldLockYouOut)
	})

	// Locking yourself out of one domain is survivable while another still lets
	// you in, so the check is about the whole result rather than each domain.
	t.Run("another enabled domain still admits", func(t *testing.T) {
		assert.NoError(t, ensureStillReachable(ctx, routing(
			domainWithAllowlist("locked.example.com", true, "10.0.0.0/8"),
			domainWithAllowlist("open.example.com", true),
		)))
	})

	// A disabled domain routes nothing, so it cannot be the way back in.
	t.Run("a disabled domain does not count as a way in", func(t *testing.T) {
		err := ensureStillReachable(ctx, routing(
			domainWithAllowlist("locked.example.com", true, "10.0.0.0/8"),
			domainWithAllowlist("open.example.com", false),
		))
		assert.ErrorIs(t, err, hperrors.ErrRoutingWouldLockYouOut)
	})

	// Every router label is rebuilt from this list; an empty one leaves no route.
	t.Run("no enabled domain at all", func(t *testing.T) {
		assert.ErrorIs(t, ensureStillReachable(ctx, routing()),
			hperrors.ErrRoutingNoEnabledDomain)
		assert.ErrorIs(t, ensureStillReachable(ctx, routing(
			domainWithAllowlist("app.example.com", false))),
			hperrors.ErrRoutingNoEnabledDomain)
	})

	t.Run("an unverifiable caller is refused", func(t *testing.T) {
		err := ensureStillReachable(context.Background(), routing(
			domainWithAllowlist("app.example.com", true, "10.0.0.0/8")))
		assert.ErrorIs(t, err, hperrors.ErrRoutingCallerUnknown)
	})

	t.Run("IPv6", func(t *testing.T) {
		ctx := ctxFromIP("2001:db8::1")
		assert.NoError(t, ensureStillReachable(ctx, routing(
			domainWithAllowlist("app.example.com", true, "2001:db8::/32"))))
		assert.ErrorIs(t, ensureStillReachable(ctx, routing(
			domainWithAllowlist("app.example.com", true, "2001:db9::/32"))),
			hperrors.ErrRoutingWouldLockYouOut)
	})
}

func TestIPAllowed(t *testing.T) {
	t.Run("an unparseable caller matches nothing", func(t *testing.T) {
		assert.False(t, ipAllowed([]string{"0.0.0.0/0"}, "not-an-address"))
	})

	t.Run("an unparseable entry is skipped, not fatal", func(t *testing.T) {
		assert.True(t, ipAllowed([]string{"garbage", "203.0.113.7"}, "203.0.113.7"))
	})

	t.Run("a v4 address is not inside a v6 block", func(t *testing.T) {
		assert.False(t, ipAllowed([]string{"2001:db8::/32"}, "203.0.113.7"))
	})
}
