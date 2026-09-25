package appcloneserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcloneservice"
)

func mountEntry(t *testing.T, name, source string, inheritable bool, parts ...string) *entity.Setting {
	t.Helper()
	mount := &entity.AppSettingMount{Source: entity.ObjectID{ID: source}}
	for _, part := range parts {
		mount.Files = append(mount.Files, &entity.AppSettingMountFile{Part: part, Path: "/etc/" + name + "/" + part})
	}
	setting := &entity.Setting{ID: "m-" + name, Type: base.SettingTypeAppSettingMount, Name: name,
		Status: base.SettingStatusActive, Inheritable: inheritable}
	assert.NoError(t, setting.SetData(mount))
	return setting
}

// A clone copies what its settings ask for, and an app's inheritable setting
// mounts: whoever made them said a copy may have them.
func TestTheDefaultCloneKeepsOnlyInheritableEntries(t *testing.T) {
	data := &appCloneData{AppCloneReq: &appcloneservice.AppCloneReq{CloneSettings: &entity.AppCloneSettings{}}}

	kept, err := (&service{}).onCloneSettingDefault(mountEntry(t, "conf", "cfg_1", true, "content"), data)
	assert.NoError(t, err)
	assert.NotNil(t, kept)
	dropped, err := (&service{}).onCloneSettingDefault(mountEntry(t, "key", "cert_1", false, "privateKey"), data)
	assert.NoError(t, err)
	assert.Nil(t, dropped)
}

// Each copied entry names the copy's setting where the source was the app's own
// and was copied; one whose own source was not copied is dropped; one mounting a
// setting from outside the app keeps it.
func TestACloneCopiesInheritableEntriesWithTheirSources(t *testing.T) {
	own := mountEntry(t, "conf", "cfg_1", true, "content")
	ownNotCopied := mountEntry(t, "token", "sec_1", true, "value")
	outside := mountEntry(t, "shared", "project_secret", true, "value")
	gated := mountEntry(t, "tls", "cert_1", true, "certificate", "privateKey")

	out := remapClonedMounts([]*entity.Setting{own, ownNotCopied, outside, gated},
		map[string]string{"cfg_1": "cfg_copy"},
		map[string]bool{"cfg_1": true, "sec_1": true},
		false)

	sources := map[string]string{}
	for _, setting := range out {
		sources[setting.Name] = setting.MustAsAppSettingMount().Source.ID
	}
	assert.Equal(t, map[string]string{"conf": "cfg_copy", "shared": "project_secret", "tls": "cert_1"}, sources)

	withoutGated := remapClonedMounts([]*entity.Setting{gated, outside}, nil, nil, true)
	if assert.Len(t, withoutGated, 1, "a denied gate leaves the private key's entry out") {
		assert.Equal(t, "shared", withoutGated[0].Name)
	}
}
