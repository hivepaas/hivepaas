package settingservice

import (
	"context"
	"crypto/x509/pkix"
	"os"
	"path/filepath"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/fileutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
)

const (
	sslSelfSignedBaseName       = "self-signed"
	sslSelfSignedCN             = "*.swarm.localhost"
	sslSelfSignedKeyType        = base.SSLKeyTypeECP256
	sslSelfSignedValidDays      = time.Hour * 24 * 365
	sslSelfSignedRenewBeforeExp = time.Hour * 24 * 30
)

func (s *settingService) initDefaultSSLSelfSigned(
	ctx context.Context,
	db database.IDB,
	timeNow time.Time,
) (err error) {
	certDir := config.Current.DataPathSslCerts()
	err = os.MkdirAll(certDir, certDirFileMode)
	if err != nil {
		return apperrors.New(err).WithMsgLog("failed to create directory to save cert files")
	}

	certFile := filepath.Join(certDir, sslSelfSignedBaseName+".crt")
	keyFile := filepath.Join(certDir, sslSelfSignedBaseName+".key")
	certFileExists, _ := fileutil.FileExists(certFile, true)
	keyFileExists, _ := fileutil.FileExists(keyFile, true)
	regenerate := !certFileExists || !keyFileExists

	var certBytes, keyBytes []byte
	validTo := timeNow.Add(sslSelfSignedValidDays)
	if regenerate {
		certBytes, keyBytes, err = s.sslService.GenerateCert(&pkix.Name{CommonName: sslSelfSignedCN},
			sslSelfSignedKeyType, timeNow, validTo, false)
		if err != nil {
			return apperrors.Wrap(err)
		}
	} else {
		certBytes, err = os.ReadFile(certFile)
		if err != nil {
			return apperrors.Wrap(err)
		}
		keyBytes, err = os.ReadFile(keyFile)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	// SSL cert settings
	sslSetting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.SettingScopeGlobal,
		Type:      base.SettingTypeSSLCert,
		Kind:      string(base.SSLCertTypeSelfSigned),
		Status:    base.SettingStatusActive,
		Name:      sslSelfSignedCN,
		Default:   true,
		Version:   entity.CurrentSSLCertVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	sslCert := &entity.SSLCert{
		CertType:      base.SSLCertTypeSelfSigned,
		Domain:        sslSelfSignedCN,
		Certificate:   reflectutil.UnsafeBytesToStr(certBytes),
		PrivateKey:    entity.NewEncryptedField(reflectutil.UnsafeBytesToStr(keyBytes)),
		KeyType:       sslSelfSignedKeyType,
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
		return apperrors.Wrap(err)
	}

	if regenerate {
		err = s.PersistSSLCertFiles(true, sslSetting)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}
