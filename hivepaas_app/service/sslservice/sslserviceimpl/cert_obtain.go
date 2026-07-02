package sslserviceimpl

import (
	"context"
	"crypto/x509/pkix"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func (s *service) ObtainCert(
	ctx context.Context,
	sslSetting *entity.Setting,
	refObjects *entity.RefObjects,
	writeFiles bool,
) (updated bool, err error) {
	sslCert := sslSetting.MustAsSSLCert()
	switch sslCert.CertType {
	case base.SSLCertTypeLetsEncrypt, base.SSLCertTypeZeroSSL, base.SSLCertTypeGoogleTrust:
		updated, err = s.obtainCertByAcme(ctx, sslSetting, refObjects)
	case base.SSLCertTypeSelfSigned:
		updated, err = s.obtainCertSelfSigned(ctx, sslSetting)
	case base.SSLCertTypeCustom:
		// No need to init as it's custom by user
		return false, nil
	default:
		return false, apperrors.New(apperrors.ErrSSLTypeUnsupported).WithParam("Type", sslCert.CertType)
	}
	if err != nil {
		return false, apperrors.New(err)
	}

	if updated && writeFiles {
		err = s.WriteCertFiles(true, sslSetting)
		if err != nil {
			return true, apperrors.New(err)
		}
	}

	return updated, nil
}

func (s *service) obtainCertByAcme(
	ctx context.Context,
	sslSetting *entity.Setting,
	refObjects *entity.RefObjects,
) (updated bool, err error) {
	acmeClient, err := s.GetAcmeClient(sslSetting, refObjects)
	if err != nil {
		return false, apperrors.New(err)
	}

	sslCert := sslSetting.MustAsSSLCert()
	keyType := gofn.Coalesce(sslCert.KeyType, base.SSLKeyTypeDefault)
	certificates, renewalInfo, err := acmeClient.ObtainCertificateWithDetails(ctx, []string{sslCert.Domain}, keyType)
	if err != nil {
		return false, apperrors.New(err)
	}

	sslCert.Certificate = string(certificates.Certificate)
	sslCert.PrivateKey = entity.NewEncryptedField(string(certificates.PrivateKey))
	if renewalInfo != nil {
		sslCert.RenewableFrom = renewalInfo.SuggestedWindow.Start.UTC()
	}
	x509Cert, err := certcrypto.ParsePEMCertificate(certificates.Certificate)
	if err != nil {
		return false, apperrors.New(err)
	}
	sslCert.ExpireAt = x509Cert.NotAfter.UTC()
	sslCert.ValidPeriod = timeutil.Duration(sslCert.ExpireAt.Sub(timeutil.NowUTC()))

	// Assign the update to the setting
	sslSetting.MustSetData(sslCert)

	return true, nil
}

func (s *service) obtainCertSelfSigned(
	_ context.Context,
	sslSetting *entity.Setting,
) (updated bool, err error) {
	sslCert := sslSetting.MustAsSSLCert()
	notBefore := timeutil.NowUTC()
	notAfter := notBefore.Add(sslCert.ValidPeriod.ToDuration())

	certBytes, keyBytes, err := s.GenerateCertAsPEM(&pkix.Name{CommonName: sslCert.Domain}, sslCert.KeyType,
		notBefore, notAfter, false)
	if err != nil {
		return false, apperrors.New(err)
	}

	sslCert.Certificate = reflectutil.UnsafeBytesToStr(certBytes)
	sslCert.PrivateKey = entity.NewEncryptedField(reflectutil.UnsafeBytesToStr(keyBytes))
	sslCert.ExpireAt = notAfter
	sslCert.RenewableFrom = sslCert.ExpireAt.Add(-base.SSLSelfSignedRenewalPeriodDefault)
	sslCert.NotifyFrom = sslCert.RenewableFrom

	// Assign the update to the setting
	sslSetting.MustSetData(sslCert)

	return true, nil
}
