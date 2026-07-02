package sslserviceimpl

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
)

const (
	certDirFileMode = 0o755
)

func (s *service) WriteCertFiles(
	forceRecreate bool,
	settings ...*entity.Setting,
) error {
	if len(settings) == 0 {
		return nil
	}
	certDir := config.Current.DataPathSslCerts().AbsPath()
	err := os.MkdirAll(certDir, certDirFileMode)
	if err != nil {
		return apperrors.New(err).WithMsgLog("failed to create directory to save cert files")
	}

	for _, setting := range settings {
		sslCert := setting.MustAsSSLCert()
		baseFilename := gofn.Coalesce(sslCert.BaseFilename, setting.ID)

		certFile := baseFilename + ".crt"
		keyFile := baseFilename + ".key"
		certFileExists, _ := fileutil.FileExists(filepath.Join(certDir, certFile), true)
		keyFileExists, _ := fileutil.FileExists(filepath.Join(certDir, keyFile), true)

		if !forceRecreate && certFileExists && keyFileExists {
			continue
		}

		certBytes := reflectutil.UnsafeStrToBytes(sslCert.Certificate)
		privateKey, err := sslCert.PrivateKey.GetPlain()
		if err != nil {
			return apperrors.New(err)
		}
		keyBytes := reflectutil.UnsafeStrToBytes(privateKey)

		err = fileutil.WriteCerts(certBytes, keyBytes, certDir, certFile, keyFile, true)
		if err != nil {
			return apperrors.New(err)
		}
	}

	return nil
}

func (s *service) DeleteCertFiles(
	settings ...*entity.Setting,
) error {
	if len(settings) == 0 {
		return nil
	}
	certDir := config.Current.DataPathSslCerts().AbsPath()
	for _, setting := range settings {
		sslCert := setting.MustAsSSLCert()
		baseFilename := gofn.Coalesce(sslCert.BaseFilename, setting.ID)

		certFile := baseFilename + ".crt"
		keyFile := baseFilename + ".key"

		err := os.Remove(filepath.Join(certDir, certFile))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return apperrors.New(err)
		}

		err = os.Remove(filepath.Join(certDir, keyFile))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return apperrors.New(err)
		}
	}
	return nil
}
