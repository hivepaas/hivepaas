package sslrenewalserviceimpl

import (
	"context"
	"crypto/x509/pkix"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func (s *service) sslRenewSelfSignedCert(
	_ context.Context,
	sslSetting *entity.Setting,
	_ *sslRenewalData,
) (err error) {
	sslCert := sslSetting.MustAsSSLCert()
	if !sslCert.AutoRenew {
		return nil
	}

	notBefore := timeutil.NowUTC()
	notAfter := notBefore.Add(sslCert.ValidPeriod.ToDuration())

	certBytes, keyBytes, err := s.sslService.GenerateCertAsPEM(&pkix.Name{CommonName: sslCert.Domain}, sslCert.KeyType,
		notBefore, notAfter, false)
	if err != nil {
		return apperrors.Wrap(err)
	}

	sslCert.Certificate = reflectutil.UnsafeBytesToStr(certBytes)
	sslCert.PrivateKey = entity.NewEncryptedField(reflectutil.UnsafeBytesToStr(keyBytes))
	sslCert.ExpireAt = notAfter
	sslCert.RenewableFrom = sslCert.ExpireAt.Add(-base.SSLSelfSignedRenewalPeriodDefault)
	sslCert.NotifyFrom = sslCert.RenewableFrom

	err = sslSetting.SetData(sslCert)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
