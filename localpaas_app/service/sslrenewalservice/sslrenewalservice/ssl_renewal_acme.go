package sslrenewalserviceimpl

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

func (s *service) sslRenewByACME(
	ctx context.Context,
	ssl *entity.SSLCert,
	data *sslRenewalData,
) (err error) {
	if !ssl.AutoRenew {
		return nil
	}

	acmeClient, err := s.sslGetAcmeClient(ssl, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	certificates, renewalInfo, err := acmeClient.ObtainCertificateWithDetails(ctx, []string{ssl.Domain})
	if err != nil {
		return apperrors.Wrap(err)
	}

	ssl.Certificate = string(certificates.Certificate)
	ssl.PrivateKey = entity.NewEncryptedField(string(certificates.PrivateKey))
	if renewalInfo != nil {
		ssl.RenewableFrom = renewalInfo.SuggestedWindow.Start.UTC()
		if !ssl.RenewableFrom.IsZero() {
			// TODO: need a better method to have expiration date of SSLs
			ssl.ExpireAt = ssl.RenewableFrom.Add(base.SSLExpirationFromFirstRenewableDate)
		}
		if !ssl.ExpireAt.IsZero() {
			ssl.ValidPeriod = timeutil.Duration(ssl.ExpireAt.Sub(timeutil.NowUTC()))
		}
	}

	return nil
}
