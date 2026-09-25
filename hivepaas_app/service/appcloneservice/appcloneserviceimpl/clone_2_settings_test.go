package appcloneserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcloneservice"
)

// A clone copies what its settings ask for, and never an app's setting mounts:
// handing a private key to a copy takes the Reveal Secrets permission, and a
// clone runs as a task with nobody's session to ask it of.
func TestACloneLeavesSettingMountsBehind(t *testing.T) {
	data := &appCloneData{AppCloneReq: &appcloneservice.AppCloneReq{CloneSettings: &entity.AppCloneSettings{
		CloneEnvVars: true, CloneSecrets: true, CloneConfigFiles: true,
	}}}
	entry := &entity.Setting{ID: "m1", Type: base.SettingTypeAppSettingMount, Name: "cert"}

	got, err := (&service{}).onCloneSettingDefault(entry, data)

	assert.NoError(t, err)
	assert.Nil(t, got)
}
