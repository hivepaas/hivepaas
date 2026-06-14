package sslserviceimpl

import (
	"context"
	"crypto/x509/pkix"

	"github.com/go-acme/lego/v5/lego"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/services/ssl/acme"
)

func (s *service) ObtainCert(
	ctx context.Context,
	sslSetting *entity.Setting,
	refObjects *entity.RefObjects,
	writeFiles bool,
) (updated bool, err error) {
	ssl := sslSetting.MustAsSSLCert()
	switch ssl.CertType {
	case base.SSLCertTypeLetsEncrypt, base.SSLCertTypeZeroSSL, base.SSLCertTypeGoogleTrust:
		updated, err = s.obtainCertByAcme(ctx, sslSetting, refObjects)
	case base.SSLCertTypeSelfSigned:
		updated, err = s.obtainCertSelfSigned(ctx, sslSetting)
	case base.SSLCertTypeCustom:
		// No need to init as it's custom by user
		return false, nil
	default:
		return false, apperrors.NewUnsupported(apperrors.Fmt("Unknown cert type '%v'", ssl.CertType))
	}
	if err != nil {
		return false, apperrors.Wrap(err)
	}

	if updated && writeFiles {
		err = s.WriteCertFiles(true, sslSetting)
		if err != nil {
			return true, apperrors.Wrap(err)
		}
	}

	return updated, nil
}

func (s *service) obtainCertByAcme(
	ctx context.Context,
	sslSetting *entity.Setting,
	refObjects *entity.RefObjects,
) (updated bool, err error) {
	ssl := sslSetting.MustAsSSLCert()
	keyType := gofn.Coalesce(ssl.KeyType, base.SSLKeyTypeDefault)

	var provider *entity.SSLProvider
	if ssl.Provider.ID != "" {
		providerSetting := refObjects.RefSettings[ssl.Provider.ID]
		if providerSetting == nil {
			return false, apperrors.NewNotFound(apperrors.Fmt("SSL provider '%v'", ssl.Provider.ID))
		}
		provider = providerSetting.MustAsSSLProvider()
	}

	http01Provider, err := acme.NewHTTP01Provider(config.Current.DataPathSslAcme().AbsPath())
	if err != nil {
		return false, apperrors.Wrap(err)
	}

	acmeCfg := acme.ACMEConfig{
		Email:          ssl.Email,
		HTTP01Provider: http01Provider,
	}
	if provider != nil {
		switch ssl.CertType {
		case base.SSLCertTypeLetsEncrypt:
			// Do nothing for now
		case base.SSLCertTypeZeroSSL:
			acmeCfg.CACode = lego.CodeZeroSSL
			acmeCfg.EABKid = provider.ZeroSSL.EABKid
			acmeCfg.EABHmacKey = provider.ZeroSSL.EABHmacKey.MustGetPlain()
		case base.SSLCertTypeGoogleTrust:
			acmeCfg.CACode = lego.CodeGoogleTrust
			acmeCfg.EABKid = provider.GoogleTS.EABKid
			acmeCfg.EABHmacKey = provider.GoogleTS.EABHmacKey.MustGetPlain()
		case base.SSLCertTypeSelfSigned, base.SSLCertTypeCustom:
			// Do nothing
		}
	}

	acmeClient, err := acme.NewClient(acmeCfg)
	if err != nil {
		return false, apperrors.Wrap(err)
	}

	certificates, renewalInfo, err := acmeClient.ObtainCertificateWithDetails(ctx, []string{ssl.Domain}, keyType)
	if err != nil {
		return false, apperrors.Wrap(err)
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

	// Assign the update to the setting
	sslSetting.MustSetData(ssl)

	return true, nil
}

func (s *service) obtainCertSelfSigned(
	_ context.Context,
	sslSetting *entity.Setting,
) (updated bool, err error) {
	ssl := sslSetting.MustAsSSLCert()
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

	// Assign the update to the setting
	sslSetting.MustSetData(ssl)

	return true, nil
}
