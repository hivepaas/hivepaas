package acme

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func DNS01ProviderTestAccess(
	ctx context.Context,
	providerKind base.AcmeDnsProvider,
	dnsConfig *entity.AcmeDnsProvider,
	testDomain string,
) (err error) {
	provider, err := NewDNS01Provider(providerKind, dnsConfig)
	if err != nil {
		return apperrors.Wrap(err)
	}
	err = provider.Present(ctx, testDomain, "test", "test")
	if err != nil {
		return apperrors.Wrap(err)
	}
	err = provider.CleanUp(ctx, testDomain, "test", "test")
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
