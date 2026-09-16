package specserviceimpl

import (
	"io"
	"os"
	"path/filepath"

	"filippo.io/age"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/filearchiver"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const (
	manifestFilename = "spec.yaml"
	stageDirName     = "stage"
	bundleFilename   = "bundle.tar.gz"

	stageDirPerm  = 0o700
	stageFilePerm = 0o600
)

// writeBundle stages the payload files and the manifest under dir and archives
// them, returning the archive path.
func writeBundle(dir string, bundle *specmodel.Bundle) (string, error) {
	stage := filepath.Join(dir, stageDirName)

	manifestBytes, err := marshalDoc(bundle.Manifest)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	if err = writeStagedFile(stage, manifestFilename, manifestBytes); err != nil {
		return "", hperrors.Wrap(err)
	}
	for name, content := range bundle.Files {
		if err = writeStagedFile(stage, name, content); err != nil {
			return "", hperrors.Wrap(err)
		}
	}

	archive := filepath.Join(dir, bundleFilename)
	cmdErr, err := filearchiver.CompressTarGz(stage, archive, filearchiver.CompressionLevelDefault)
	if err != nil {
		return "", hperrors.Wrap(err).WithMsgLog("archiving the spec bundle failed: %s", cmdErr)
	}
	return archive, nil
}

func writeStagedFile(stage, name string, content []byte) error {
	path := filepath.Join(stage, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), stageDirPerm); err != nil {
		return hperrors.Wrap(err)
	}
	if err := os.WriteFile(path, content, stageFilePerm); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// encryptBundle wraps the whole archive with age under a passphrase.
//
// This is the path sysbackupservice takes for an encrypted backup, and it is
// deliberately not the app secret. That secret protects every stored value in
// the installation, so shipping it to make an import possible would send the
// most sensitive credential there is along with the file - and it would not
// work anyway, since stored ciphertext is sealed by a data key generated per
// installation that the target does not have.
//
// age carries its own salt in its header, so the bundle needs no key material
// of its own, and the result opens with the standard age CLI without HivePaaS
// running at all.
func encryptBundle(src, dst, passphrase string) error {
	if passphrase == "" {
		return hperrors.Wrap(hperrors.ErrSpecPassphraseRequired)
	}

	recipient, err := age.NewScryptRecipient(passphrase)
	if err != nil {
		return hperrors.Wrap(err)
	}

	in, err := os.Open(src)
	if err != nil {
		return hperrors.Wrap(err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, stageFilePerm)
	if err != nil {
		return hperrors.Wrap(err)
	}
	defer out.Close()

	writer, err := age.Encrypt(out, recipient)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if _, err = io.Copy(writer, in); err != nil {
		return hperrors.Wrap(err)
	}
	if err = writer.Close(); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
