package fileutil

import (
	"os"
	"path/filepath"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
)

const (
	certDirFileMode = 0o755
	certFileMode    = 0o644
)

func WriteCerts(cert, privateKey []byte, saveDir, certFile, keyFile string, overwrite bool) (err error) {
	err = os.MkdirAll(saveDir, certDirFileMode)
	if err != nil {
		return apperrors.New(err).WithMsgLog("failed to create directory to save ssl certificate")
	}

	certPath := filepath.Join(saveDir, certFile)
	keyPath := filepath.Join(saveDir, keyFile)

	var writtenCert, writtenKey bool
	defer func() {
		if err == nil {
			return
		}
		// Try to remove ONLY the files we just created during this run
		if writtenKey {
			_ = os.Remove(keyPath)
		}
		if writtenCert {
			_ = os.Remove(certPath)
		}
	}()

	if certFile != "" {
		if overwrite || !gofn.Head(FileExists(certPath, true)) {
			err = os.WriteFile(certPath, cert, certFileMode)
			if err != nil {
				return apperrors.New(err).WithMsgLog("failed to write cert file to %s", certPath)
			}
			writtenCert = true
		}
	}
	if keyFile != "" {
		if overwrite || !gofn.Head(FileExists(keyPath, true)) {
			err = os.WriteFile(keyPath, privateKey, certFileMode)
			if err != nil {
				return apperrors.New(err).WithMsgLog("failed to write key file to %s", keyPath)
			}
			writtenKey = true
		}
	}

	return nil
}
