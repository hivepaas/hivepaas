package appservice

import (
	"os"
	"path/filepath"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/fileutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/htpasswd"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
)

const (
	certDirFileMode      = 0o755
	basicAuthDirFileMode = 0o755
)

func (s *appService) EnsureSSLConfigFiles(
	sslIDs []string,
	forceRecreate bool,
	refObjects *entity.RefObjects,
) error {
	if len(sslIDs) == 0 {
		return nil
	}
	certDir := config.Current.DataPathCerts()
	err := os.MkdirAll(certDir, certDirFileMode)
	if err != nil {
		return apperrors.New(err).WithMsgLog("failed to create directory to save cert files")
	}

	for _, sslID := range sslIDs {
		certFile := sslID + ".crt"
		keyFile := sslID + ".key"
		certFileExists, _ := fileutil.FileExists(filepath.Join(certDir, certFile), true)
		keyFileExists, _ := fileutil.FileExists(filepath.Join(certDir, keyFile), true)

		if !forceRecreate && certFileExists && keyFileExists {
			continue
		}

		dbSSL := refObjects.RefSettings[sslID]
		if dbSSL == nil {
			return apperrors.NewNotFound("SSL").WithMsgLog("ssl %s not found", sslID)
		}
		ssl := dbSSL.MustAsSSL()
		certBytes := reflectutil.UnsafeStrToBytes(ssl.Certificate)
		keyBytes := reflectutil.UnsafeStrToBytes(ssl.PrivateKey.MustGetPlain())

		err := fileutil.WriteCerts(certBytes, keyBytes, certDir, certFile, keyFile, true)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}

func (s *appService) EnsureBasicAuthConfigFiles(
	basicAuthIDs []string,
	forceRecreate bool,
	refObjects *entity.RefObjects,
) error {
	if len(basicAuthIDs) == 0 {
		return nil
	}
	basicAuthDir := config.Current.DataPathNginxShareBasicAuth()
	err := os.MkdirAll(basicAuthDir, basicAuthDirFileMode)
	if err != nil {
		return apperrors.New(err).WithMsgLog("failed to create directory to save basic auth files")
	}

	for _, authID := range basicAuthIDs {
		authFile := filepath.Join(basicAuthDir, authID)
		authFileExists, _ := fileutil.FileExists(authFile, true)

		if !forceRecreate && authFileExists {
			continue
		}

		dbBasicAuth := refObjects.RefSettings[authID]
		if dbBasicAuth == nil {
			return apperrors.NewNotFound("BasicAuth").WithMsgLog("basic-auth %s not found", authID)
		}
		basicAuth := dbBasicAuth.MustAsBasicAuth()

		hashedPasswd := htpasswd.HashedPasswords{}
		err = hashedPasswd.SetPassword(basicAuth.Username, basicAuth.Password.MustGetPlain(), htpasswd.HashBCrypt)
		if err != nil {
			return apperrors.Wrap(err)
		}

		err = hashedPasswd.WriteToFile(authFile)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}
