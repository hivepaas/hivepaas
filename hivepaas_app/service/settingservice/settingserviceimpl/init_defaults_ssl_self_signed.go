package settingserviceimpl

import (
	"context"
	"crypto/x509/pkix"
	"os"
	"path/filepath"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
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

func (s *service) initDefaultSSLSelfSigned(
	ctx context.Context,
	db database.IDB,
	timeNow time.Time,
) (err error) {
	certDir := config.Current.DataPathSslCerts().AbsPath()
	certFile := filepath.Join(certDir, sslSelfSignedBaseName+".crt")
	keyFile := filepath.Join(certDir, sslSelfSignedBaseName+".key")
	certFileExists, _ := fileutil.FileExists(certFile, true)
	keyFileExists, _ := fileutil.FileExists(keyFile, true)
	regenerate := !certFileExists || !keyFileExists

	var certBytes, keyBytes []byte
	domain := gofn.Coalesce(config.Current.RootDomain, sslSelfSignedCN)
	validTo := timeNow.Add(sslSelfSignedValidPeriod)
	if regenerate {
		certBytes, keyBytes, err = s.sslService.GenerateCertAsPEM(&pkix.Name{CommonName: domain},
			sslSelfSignedKeyType, timeNow, validTo, false)
		if err != nil {
			return apperrors.New(err)
		}
	} else {
		certBytes, err = os.ReadFile(certFile)
		if err != nil {
			return apperrors.New(err)
		}
		keyBytes, err = os.ReadFile(keyFile)
		if err != nil {
			return apperrors.New(err)
		}
	}

	// SSL cert settings
	sslSetting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.ObjectScopeGlobal,
		Type:      base.SettingTypeSSLCert,
		Kind:      string(base.SSLCertTypeSelfSigned),
		Status:    base.SettingStatusActive,
		Name:      domain,
		Default:   true,
		Version:   entity.CurrentSSLCertVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
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
		return apperrors.New(err)
	}

	if regenerate {
		err = s.sslService.WriteCertFiles(true, sslSetting)
		if err != nil {
			return apperrors.New(err)
		}
	}

	return nil
}
