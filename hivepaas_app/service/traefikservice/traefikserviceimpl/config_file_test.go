package traefikserviceimpl

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
)

// dynamicDir is a data path of its own for the test, with Traefik's dynamic
// configuration directory in it.
func dynamicDir(t *testing.T) string {
	t.Helper()
	config.SetCurrent(&config.Config{AppPath: t.TempDir()})
	t.Cleanup(func() { config.SetCurrent(nil) })
	dir := config.Current().DataPathTraefikEtcDynamic().AbsPath()
	assert.NoError(t, os.MkdirAll(dir, 0o755))
	return dir
}

func certConfig(certID string) *appConfigData {
	traefikConfig := &AppTraefikConfig{}
	(&service{}).addTLSCertificate(traefikConfig, certID)
	return &appConfigData{confData: traefikConfig}
}

func filesIn(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	assert.NoError(t, err)
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

// Two apps of one key - web in a project's development and production - each
// keep a file of their own: the one written last no longer replaces the other.
func TestAppsOfOneKeyKeepTheirOwnFiles(t *testing.T) {
	dir := dynamicDir(t)
	s := &service{}
	dev := certConfig("cert-dev")
	dev.ApplyAppConfigReq = &traefikservice.ApplyAppConfigReq{App: &entity.App{ID: "01DEV", Key: "web"}}
	prod := certConfig("cert-prod")
	prod.ApplyAppConfigReq = &traefikservice.ApplyAppConfigReq{App: &entity.App{ID: "01PROD", Key: "web"}}

	assert.NoError(t, s.writeAppConfigFile(dev))
	assert.NoError(t, s.writeAppConfigFile(prod))

	assert.ElementsMatch(t, []string{"01dev.yml", "01prod.yml"}, filesIn(t, dir))
	devFile, err := os.ReadFile(filepath.Join(dir, "01dev.yml"))
	assert.NoError(t, err)
	assert.Contains(t, string(devFile), "cert-dev.crt")
}

// The file an app had before, named after its key, goes once the app's own
// is written, or once the app is removed.
func TestTheFileNamedAfterTheKeyGoes(t *testing.T) {
	dir := dynamicDir(t)
	s := &service{}
	app := &entity.App{ID: "01WEB", Key: "web"}
	legacy := filepath.Join(dir, "web.yml")
	assert.NoError(t, os.WriteFile(legacy, []byte("tls: {}\n"), 0o600))

	data := certConfig("cert-web")
	data.ApplyAppConfigReq = &traefikservice.ApplyAppConfigReq{App: app}
	assert.NoError(t, s.writeAppConfigFile(data))
	assert.Equal(t, []string{"01web.yml"}, filesIn(t, dir))

	assert.NoError(t, os.WriteFile(legacy, []byte("tls: {}\n"), 0o600))
	_, err := s.RemoveAppConfig(context.Background(), nil, &traefikservice.RemoveAppConfigReq{App: app})
	assert.NoError(t, err)
	assert.Empty(t, filesIn(t, dir))
}
