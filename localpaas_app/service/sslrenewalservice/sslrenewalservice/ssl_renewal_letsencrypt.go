package sslrenewalserviceimpl

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

func (s *service) sslRenewByLetsEncrypt(
	ctx context.Context,
	ssl *entity.SSLCert,
	data *sslRenewalData,
) (err error) {
	if !ssl.AutoRenew {
		return nil
	}

	leClient, err := s.sslGetLeClient(ssl, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	certificates, renewalInfo, err := leClient.ObtainCertificateWithDetails(ctx, []string{ssl.Domain})
	if err != nil {
		return apperrors.Wrap(err)
	}

	ssl.Certificate = string(certificates.Certificate)
	ssl.PrivateKey = entity.NewEncryptedField(string(certificates.PrivateKey))
	if renewalInfo != nil {
		ssl.RenewableFrom = renewalInfo.SuggestedWindow.Start.UTC()
		if !ssl.RenewableFrom.IsZero() {
			// TODO: need a better method to have expiration date of SSLs from Let's encrypt.
			ssl.ExpireAt = ssl.RenewableFrom.Add(base.LetsEncryptExpirationFromFirstRenewableDate)
		}
		if !ssl.ExpireAt.IsZero() {
			ssl.ValidPeriod = timeutil.Duration(ssl.ExpireAt.Sub(timeutil.NowUTC()))
		}
	}

	return nil
}
