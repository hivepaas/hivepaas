package sslrenewalserviceimpl

import (
	"context"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

func (s *service) sslRenewByAcme(
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

	keyType := gofn.Coalesce(ssl.KeyType, base.SSLKeyTypeDefault)
	certificates, renewalInfo, err := acmeClient.ObtainCertificateWithDetails(ctx, []string{ssl.Domain}, keyType)
	if err != nil {
		return apperrors.Wrap(err)
	}

	ssl.Certificate = string(certificates.Certificate)
	ssl.PrivateKey = entity.NewEncryptedField(string(certificates.PrivateKey))
	if renewalInfo != nil {
		ssl.RenewableFrom = renewalInfo.SuggestedWindow.Start.UTC()
	}
	x509Cert, err := certcrypto.ParsePEMCertificate(certificates.Certificate)
	if err != nil {
		return apperrors.Wrap(err)
	}
	ssl.ExpireAt = x509Cert.NotAfter.UTC()
	ssl.ValidPeriod = timeutil.Duration(ssl.ExpireAt.Sub(timeutil.NowUTC()))

	return nil
}
