package sslserviceimpl

import (
	"context"
	"crypto/x509/pkix"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/services/ssl/letsencrypt"
)

func (s *service) ObtainCert(
	ctx context.Context,
	sslSetting *entity.Setting,
	writeFiles bool,
) (updated bool, err error) {
	ssl := sslSetting.MustAsSSLCert()
	switch ssl.CertType {
	case base.SSLCertTypeLetsEncrypt:
		updated, err = s.obtainCertLetsEncrypt(ctx, ssl)
	case base.SSLCertTypeZeroSSL:
		// TODO: add implementation
	case base.SSLCertTypeGoogleTS:
		// TODO: add implementation
	case base.SSLCertTypeSelfSigned:
		updated, err = s.obtainCertSelfSigned(ctx, ssl)
	case base.SSLCertTypeCustom:
		// No need to init as it's custom by user
		return false, nil
	default:
		return false, apperrors.NewUnsupported(apperrors.Fmt("Unknown cert type '%v'", ssl.CertType))
	}
	if err != nil {
		return false, apperrors.Wrap(err)
	}

	if updated {
		sslSetting.MustSetData(ssl)
	}
	if updated && writeFiles {
		err = s.WriteCertFiles(true, sslSetting)
		if err != nil {
			return true, apperrors.Wrap(err)
		}
	}

	return updated, nil
}

func (s *service) obtainCertLetsEncrypt(
	ctx context.Context,
	ssl *entity.SSLCert,
) (updated bool, err error) {
	email := ssl.Email
	keyType := gofn.Coalesce(ssl.KeyType, base.SSLKeyTypeDefault)
	leClient, err := letsencrypt.NewClient(email, keyType, config.Current.DataPathSslLetsEncrypt().AbsPath())
	if err != nil {
		return false, apperrors.Wrap(err)
	}

	certificates, renewalInfo, err := leClient.ObtainCertificateWithDetails(ctx, []string{ssl.Domain})
	if err != nil {
		return false, apperrors.Wrap(err)
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

	return true, nil
}

func (s *service) obtainCertSelfSigned(
	_ context.Context,
	ssl *entity.SSLCert,
) (updated bool, err error) {
	notBefore := timeutil.NowUTC()
	notAfter := notBefore.Add(ssl.ValidPeriod.ToDuration())

	certBytes, keyBytes, err := s.GenerateCertAsPEM(&pkix.Name{CommonName: ssl.Domain}, ssl.KeyType,
		notBefore, notAfter, false)
	if err != nil {
		return false, apperrors.Wrap(err)
	}

	ssl.Certificate = reflectutil.UnsafeBytesToStr(certBytes)
	ssl.PrivateKey = entity.NewEncryptedField(reflectutil.UnsafeBytesToStr(keyBytes))
	ssl.ExpireAt = notAfter
	ssl.RenewableFrom = ssl.ExpireAt.Add(-base.SSLSelfSignedRenewalPeriodDefault)
	ssl.NotifyFrom = ssl.RenewableFrom

	return true, nil
}
