package specserviceimpl

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"filippo.io/age"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func testBundle() *specmodel.Bundle {
	return &specmodel.Bundle{
		Manifest: &specmodel.Manifest{
			APIVersion:  specmodel.APIVersion,
			Kind:        specmodel.KindSpec,
			ExportedAt:  time.Date(2026, 9, 16, 10, 15, 0, 0, time.UTC),
			Scope:       "global",
			SecretsMode: specmodel.SecretsModeOmit,
			Files: []string{
				"global.yaml",
				"projects/project_a/project.yaml",
				"projects/project_a/envs/dev.yaml",
			},
		},
		Files: map[string][]byte{
			"global.yaml":                      []byte("apiVersion: hivepaas.com/v1\n"),
			"projects/project_a/project.yaml":  []byte("apiVersion: hivepaas.com/v1\n"),
			"projects/project_a/envs/dev.yaml": []byte("apiVersion: hivepaas.com/v1\n"),
		},
	}
}

func TestWriteBundleLaysOutFilesPerEnv(t *testing.T) {
	dir := t.TempDir()

	path, err := writeBundle(dir, testBundle())
	assert.NoError(t, err)
	assert.FileExists(t, path)

	// The manifest is written even though it is not listed in Files.
	staged, err := os.ReadFile(filepath.Join(dir, stageDirName, manifestFilename))
	assert.NoError(t, err)
	assert.Contains(t, string(staged), "apiVersion: hivepaas.com/v1")
	assert.Contains(t, string(staged), "secretsMode: omit")

	// And the per-env path really is nested, not flattened.
	assert.FileExists(t, filepath.Join(dir, stageDirName, "projects", "project_a", "envs", "dev.yaml"))
}

func TestWriteBundleProducesAReadableArchive(t *testing.T) {
	dir := t.TempDir()
	path, err := writeBundle(dir, testBundle())
	assert.NoError(t, err)

	listing, err := exec.Command("tar", "-tzf", path).Output()
	assert.NoError(t, err)
	for _, want := range []string{
		"spec.yaml", "global.yaml", "projects/project_a/envs/dev.yaml",
	} {
		assert.Contains(t, string(listing), want)
	}
}

func TestEncryptBundleProducesSomethingAgeCanOpen(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "bundle.tar.gz")
	body := "not really a tarball, but bytes"
	assert.NoError(t, os.WriteFile(src, []byte(body), stageFilePerm))

	dst := filepath.Join(dir, "bundle.tar.gz.age")
	assert.NoError(t, encryptBundle(src, dst, "correct horse battery staple"))

	sealed, err := os.ReadFile(dst)
	assert.NoError(t, err)
	assert.NotContains(t, string(sealed), body)
	assert.True(t, strings.HasPrefix(string(sealed), "age-encryption.org/"),
		"the standard age CLI must recognize it")

	identity, err := age.NewScryptIdentity("correct horse battery staple")
	assert.NoError(t, err)
	opened, err := os.Open(dst)
	assert.NoError(t, err)
	defer opened.Close()

	reader, err := age.Decrypt(opened, identity)
	assert.NoError(t, err)
	plain, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, body, string(plain))
}

func TestEncryptBundleRefusesTheWrongPassphrase(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "bundle.tar.gz")
	assert.NoError(t, os.WriteFile(src, []byte("payload"), stageFilePerm))
	dst := filepath.Join(dir, "out.age")
	assert.NoError(t, encryptBundle(src, dst, "right"))

	identity, err := age.NewScryptIdentity("wrong")
	assert.NoError(t, err)
	opened, err := os.Open(dst)
	assert.NoError(t, err)
	defer opened.Close()

	_, err = age.Decrypt(opened, identity)
	assert.Error(t, err)
}

func TestEncryptBundleRefusesAnEmptyPassphrase(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "bundle.tar.gz")
	assert.NoError(t, os.WriteFile(src, []byte("x"), stageFilePerm))

	assert.Error(t, encryptBundle(src, filepath.Join(dir, "out.age"), ""))
}
