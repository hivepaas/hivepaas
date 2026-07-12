package sslrenewalserviceimpl

import (
	"context"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func (s *service) sslRenewByAcme(
	ctx context.Context,
	sslSetting *entity.Setting,
	data *sslRenewalData,
) (err error) {
	sslCert := sslSetting.MustAsSSLCert()
	if !sslCert.AutoRenew {
		return nil
	}

	acmeClient, err := s.sslService.GetAcmeClient(sslSetting, data.RefObjects)
	if err != nil {
		return apperrors.Wrap(err)
	}

	keyType := gofn.Coalesce(sslCert.KeyType, base.SSLKeyTypeDefault)
	certificates, renewalInfo, err := acmeClient.ObtainCertificateWithDetails(ctx, []string{sslCert.Domain}, keyType)
	if err != nil {
		return apperrors.Wrap(err)
	}

	sslCert.Certificate = string(certificates.Certificate)
	sslCert.PrivateKey = entity.NewEncryptedField(string(certificates.PrivateKey))
	if renewalInfo != nil {
		sslCert.RenewableFrom = renewalInfo.SuggestedWindow.Start.UTC()
	}
	x509Cert, err := certcrypto.ParsePEMCertificate(certificates.Certificate)
	if err != nil {
		return apperrors.Wrap(err)
	}
	sslCert.ExpireAt = x509Cert.NotAfter.UTC()
	sslCert.ValidPeriod = timeutil.Duration(sslCert.ExpireAt.Sub(timeutil.NowUTC()))

	err = sslSetting.SetData(sslCert)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
