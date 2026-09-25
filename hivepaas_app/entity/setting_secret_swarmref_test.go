package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// Nothing is carried over from swarmRef, which secrets and config files no
// longer have: a row that still holds one reads without it.
func TestARowWithASwarmRefReadsWithoutIt(t *testing.T) {
	secret := &Setting{Type: base.SettingTypeSecret, Data: `{"key":"A","swarmRef":{"file":{"name":"a"}}}`}
	gotSecret, err := secret.AsSecret()
	if assert.NoError(t, err) {
		assert.Equal(t, "A", gotSecret.Key)
	}

	configFile := &Setting{Type: base.SettingTypeConfigFile,
		Data: `{"name":"app.conf","content":"x","swarmRef":{"file":{"name":"/etc/app.conf"}}}`}
	gotConfig, err := configFile.AsConfigFile()
	if assert.NoError(t, err) {
		assert.Equal(t, "x", gotConfig.Content)
	}
}
