package sslrenewalserviceimpl

import (
	"context"
	"crypto/x509/pkix"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

func (s *service) sslRenewSelfSignedCert(
	_ context.Context,
	ssl *entity.SSLCert,
	_ *sslRenewalData,
) (err error) {
	if !ssl.AutoRenew {
		return nil
	}

	notBefore := timeutil.NowUTC()
	notAfter := notBefore.Add(ssl.ValidPeriod.ToDuration())

	certBytes, keyBytes, err := s.sslService.GenerateCertAsPEM(&pkix.Name{CommonName: ssl.Domain}, ssl.KeyType,
		notBefore, notAfter, false)
	if err != nil {
		return apperrors.Wrap(err)
	}

	ssl.Certificate = reflectutil.UnsafeBytesToStr(certBytes)
	ssl.PrivateKey = entity.NewEncryptedField(reflectutil.UnsafeBytesToStr(keyBytes))
	ssl.ExpireAt = notAfter
	ssl.RenewableFrom = ssl.ExpireAt.Add(-base.SSLSelfSignedRenewalPeriodDefault)
	ssl.NotifyFrom = ssl.RenewableFrom

	return nil
}
