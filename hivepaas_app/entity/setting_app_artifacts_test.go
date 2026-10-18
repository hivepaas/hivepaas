package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func TestAppArtifactsSettingsRoundTrip(t *testing.T) {
	setting := &Setting{Type: base.SettingTypeAppArtifacts, Version: CurrentAppArtifactsVersion}
	setting.MustSetData(&AppArtifactsSettings{Artifacts: []*AppArtifact{{
		Name: "wal-g", Version: "3.0.8", MountPath: "/opt/hivepaas/bin/wal-g",
		Files: map[base.CPUArch]*AppArtifactFile{
			base.CPUArchAMD64: {URL: "https://example.com/a", SHA256: "aa", Size: 1},
			base.CPUArchARM64: {URL: "https://example.com/b", SHA256: "bb", Size: 2},
		},
	}}})

	stored := &Setting{Type: base.SettingTypeAppArtifacts, Version: setting.Version, Data: setting.Data}
	artifacts, err := stored.AsAppArtifactsSettings()

	assert.NoError(t, err)
	assert.Equal(t, "wal-g", artifacts.Artifacts[0].Name)
	assert.Equal(t, int64(2), artifacts.Artifacts[0].Files[base.CPUArchARM64].Size)
	assert.Equal(t, []string{"aa", "bb"}, artifacts.SHA256s(), "every file, whatever the node")
	assert.Nil(t, (*AppArtifactsSettings)(nil).SHA256s())
}
