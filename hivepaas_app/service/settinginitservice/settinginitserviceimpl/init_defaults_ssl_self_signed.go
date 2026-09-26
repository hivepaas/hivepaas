package settinginitserviceimpl

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"os"
	"path/filepath"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

const (
	sslSelfSignedBaseName       = "self-signed"
	sslSelfSignedCN             = "*.swarm.localhost"
	sslSelfSignedKeyType        = base.SSLKeyTypeECP256
	sslSelfSignedValidPeriod    = timeutil.Day * 365
	sslSelfSignedRenewBeforeExp = timeutil.Day * 30
)

// InitSelfSignedCert creates the certificate an app is served with until a real
// one exists for its domain.
//
// It runs while HivePaaS is being installed and at no other time. It used to be
// one of the defaults, which are filled in whenever a unique setting is found
// missing - GetUniqueSettingOrEmpty does that the first time somebody opens a
// settings screen whose default did not exist yet - and a certificate was
// created on each of those visits. An operator who deletes this certificate has
// decided something, and nothing should undo that decision for them.
//
// Creating it twice for the same domain is refused as well, for an installation
// whose steps are run again.
func (s *service) InitSelfSignedCert(
	ctx context.Context,
	db database.Tx,
) (err error) {
	timeNow := timeutil.NowUTC()
	domain := gofn.Coalesce(config.Current().RootDomain, sslSelfSignedCN)
	existing, err := s.selfSignedCertFor(ctx, db, domain)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if existing {
		return nil
	}

	certDir := config.Current().DataPathSslCerts().AbsPath()
	certFile := filepath.Join(certDir, sslSelfSignedBaseName+".crt")
	keyFile := filepath.Join(certDir, sslSelfSignedBaseName+".key")
	certFileExists, _ := fileutil.FileExists(certFile, true)
	keyFileExists, _ := fileutil.FileExists(keyFile, true)
	regenerate := !certFileExists || !keyFileExists

	var certBytes, keyBytes []byte
	validTo := timeNow.Add(sslSelfSignedValidPeriod)
	if !regenerate {
		// The installer makes the certificate before the first boot so Traefik
		// serves it from its first second. It is recorded with the dates it has,
		// or the renewal would wait for dates it does not have.
		if certBytes, err = os.ReadFile(certFile); err != nil {
			return hperrors.Wrap(err)
		}
		notAfter, adoptable := adoptableSelfSigned(certBytes, timeNow)
		if adoptable {
			validTo = notAfter
			if keyBytes, err = os.ReadFile(keyFile); err != nil {
				return hperrors.Wrap(err)
			}
		} else {
			regenerate = true
		}
	}
	if regenerate {
		certBytes, keyBytes, err = s.sslService.GenerateCertAsPEM(&pkix.Name{CommonName: domain},
			sslSelfSignedKeyType, timeNow, validTo, false)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	// SSL cert settings
	sslSetting := &entity.Setting{
		ID:          gofn.Must(ulid.NewStringULID()),
		Scope:       base.ObjectScopeGlobal,
		Type:        base.SettingTypeSSLCert,
		Kind:        string(base.SSLCertTypeSelfSigned),
		Status:      base.SettingStatusActive,
		Name:        domain,
		Inheritable: true,
		Default:     true,
		Version:     entity.CurrentSSLCertVersion,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}
	sslCert := &entity.SSLCert{
		CertType:      base.SSLCertTypeSelfSigned,
		Domain:        domain,
		Certificate:   reflectutil.UnsafeBytesToStr(certBytes),
		PrivateKey:    entity.NewEncryptedField(reflectutil.UnsafeBytesToStr(keyBytes)),
		KeyType:       sslSelfSignedKeyType,
		ValidPeriod:   timeutil.Duration(sslSelfSignedValidPeriod),
		BaseFilename:  sslSelfSignedBaseName,
		AutoRenew:     true,
		RenewableFrom: validTo.Add(-sslSelfSignedRenewBeforeExp),
		ExpireAt:      validTo,
		Notification: &entity.BaseEventNotification{
			SuccessUseDefault: true,
			FailureUseDefault: true,
		},
	}
	sslSetting.MustSetData(sslCert)

	// Save the objects in DB
	err = s.settingRepo.Insert(ctx, db, sslSetting)
	if err != nil {
		return hperrors.Wrap(err)
	}

	if regenerate {
		err = s.sslService.WriteCertFiles(true, sslSetting)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	return nil
}

// selfSignedCertFor reports whether this installation already has the default
// self-signed certificate for a domain. The domain is part of the question: a
// system whose root domain changed needs one for the new name, and the old one
// stays for whatever still answers at the old.
func (s *service) selfSignedCertFor(ctx context.Context, db database.IDB, domain string) (bool, error) {
	certs, _, err := s.settingRepo.List(ctx, db, entity.NewObjectScopeGlobal(), nil,
		bunex.SelectColumns("id"),
		bunex.SelectWhere("setting.type = ?", base.SettingTypeSSLCert),
		bunex.SelectWhere("setting.kind = ?", string(base.SSLCertTypeSelfSigned)),
		bunex.SelectWhere("setting.name = ?", domain),
		bunex.SelectWhere("setting.is_default = ?", true),
	)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	return len(certs) > 0, nil
}

// adoptableSelfSigned says whether a certificate found on disk can be taken on
// as it is, and until when it is valid: it has to be one, and not be due for
// renewal already. Anything else is made anew.
func adoptableSelfSigned(certPEM []byte, timeNow time.Time) (time.Time, bool) {
	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return time.Time{}, false
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, false
	}
	notAfter := cert.NotAfter.UTC()
	if !notAfter.After(timeNow.Add(sslSelfSignedRenewBeforeExp)) {
		return time.Time{}, false
	}
	return notAfter, true
}
