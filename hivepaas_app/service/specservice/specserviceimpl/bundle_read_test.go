package specserviceimpl

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func exportedBytes(t *testing.T, mode specmodel.SecretsMode, passphrase string) []byte {
	t.Helper()
	path, _ := runExport(t, mode, passphrase)
	content, err := os.ReadFile(path)
	assert.NoError(t, err)
	return content
}

// tarGz builds an archive the way export's tar does, with ./ in front of every
// entry.
func tarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		assert.NoError(t, tw.WriteHeader(&tar.Header{Name: name, Mode: 0o600, Size: int64(len(content))}))
		_, err := tw.Write([]byte(content))
		assert.NoError(t, err)
	}
	assert.NoError(t, tw.Close())
	assert.NoError(t, gz.Close())
	return buf.Bytes()
}

const manifestYAML = "apiVersion: hivepaas.com/v1\nkind: Spec\nscope: global\nsecretsMode: omit\n"

// Whatever export writes, the reader reads back: the manifest, global.yaml, and
// every project and env document, keyed by their keys.
func TestReadBundleReadsWhatExportWrites(t *testing.T) {
	for _, mode := range []specmodel.SecretsMode{
		specmodel.SecretsModeOmit, specmodel.SecretsModePlaintext, specmodel.SecretsModeEncrypted,
	} {
		t.Run(string(mode), func(t *testing.T) {
			passphrase := ""
			if mode == specmodel.SecretsModeEncrypted {
				passphrase = "correct horse battery staple"
			}
			content := exportedBytes(t, mode, passphrase)

			bundle, err := readBundle(content, passphrase)

			assert.NoError(t, err)
			assert.Equal(t, mode, bundle.Manifest.SecretsMode)
			assert.NotNil(t, bundle.Global)
			if assert.Contains(t, bundle.Projects, "project_a") {
				assert.Equal(t, "p1", bundle.Projects["project_a"].ID)
			}
			if assert.Contains(t, bundle.Envs["project_a"], "dev") {
				assert.Contains(t, bundle.Envs["project_a"]["dev"].Apps, "backend")
			}
			assert.Len(t, bundle.Digest, 64)
		})
	}
}

func TestReadBundleRefuses(t *testing.T) {
	encrypted := exportedBytes(t, specmodel.SecretsModeEncrypted, "correct horse battery staple")

	cases := map[string]struct {
		content    []byte
		passphrase string
		want       error
	}{
		"an encrypted bundle with no passphrase": {encrypted, "", hperrors.ErrSpecPassphraseRequired},
		"the wrong passphrase":                   {encrypted, "wrong", hperrors.ErrSpecPassphraseInvalid},
		"something that is not a bundle":         {[]byte("hello"), "", hperrors.ErrSpecBundleInvalid},
		"no manifest": {tarGz(t, map[string]string{"./global.yaml": "apiVersion: hivepaas.com/v1\n"}), "",
			hperrors.ErrSpecBundleInvalid},
		"an entry leaving the archive": {tarGz(t, map[string]string{
			"./spec.yaml": manifestYAML, "../../etc/passwd": "x",
		}), "", hperrors.ErrSpecBundleInvalid},
		"a format this installation does not read": {tarGz(t, map[string]string{
			"./spec.yaml": "apiVersion: hivepaas.com/v99\nkind: Spec\n",
		}), "", hperrors.ErrSpecAPIVersionUnsupported},
		"a document that is not YAML": {tarGz(t, map[string]string{
			"./spec.yaml": manifestYAML, "./global.yaml": "settings: [unclosed",
		}), "", hperrors.ErrSpecBundleInvalid},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := readBundle(tc.content, tc.passphrase)
			assert.ErrorIs(t, err, tc.want)
		})
	}
}
