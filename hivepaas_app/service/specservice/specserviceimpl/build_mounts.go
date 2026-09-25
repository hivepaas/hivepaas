package specserviceimpl

import (
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// addTemplateMount makes the setting mount entry a template's swarmRef.file
// stands for: one file of source's part, where the template put it, inheritable
// with its source.
func (state *buildState) addTemplateMount(
	block specmodel.Block, name string, source *entity.Setting, part string, file map[string]any, inheritable bool,
) error {
	key := settingmountservice.EntryKeyFor(name)
	if !settingmountservice.ValidEntryKey(key) {
		return invalidBlock(block, "%s: its name makes no entry key", name)
	}
	for _, existing := range state.settings {
		if existing.Type == base.SettingTypeAppSettingMount && existing.Name == key {
			return invalidBlock(block, "%s: its mount would be called %q, as another's is", name, key)
		}
	}
	mountFile := &entity.AppSettingMountFile{Part: part}
	mountFile.Path, _ = file["name"].(string)
	if uid, ok := file["uid"]; ok {
		mountFile.UID = fmt.Sprint(uid)
	}
	if gid, ok := file["gid"]; ok {
		mountFile.GID = fmt.Sprint(gid)
	}
	if mode, ok := file["mode"]; ok {
		parsed, err := fileutil.ParseFileMode(fmt.Sprint(mode))
		if err != nil {
			return invalidBlock(block, "%s: mode %v is not a file mode", name, mode)
		}
		mountFile.Mode = parsed
	}
	_, err := state.addNamedSetting(base.SettingTypeAppSettingMount, key, entity.CurrentAppSettingMountVersion,
		inheritable, &entity.AppSettingMount{
			Source: entity.ObjectID{ID: source.ID},
			Files:  []*entity.AppSettingMountFile{mountFile},
		})
	return err
}
