package managedscripts

import (
	"slices"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
)

// mountPart is the part of a config file a script's mount puts in place.
const mountPart = "content"

// Patch is what replacing an app's scripts and artifacts changes of its
// settings: the settings to create, to update and to delete - stamped, their
// update versions moved - and the ids its engine setting's locks gain and lose.
type Patch struct {
	Created, Updated, Deleted []*entity.Setting
	LocksAdded, LocksRemoved  []string
}

// script is one of an app's scripts as its settings hold it.
type appScript struct {
	file, mount *entity.Setting
	data        *entity.ConfigFile
	mountData   *entity.AppSettingMount
	target      *entity.AppSettingMountFile
}

// Plan replaces the app's scripts and artifacts - its settings given - with
// available's. A script is known by its config file's name: a changed one is
// changed in place, a new one gets a config file and a setting mount, and one
// the set no longer has loses both. The app's other config files stay, and a
// new script whose name one of them has is refused.
func Plan(app *entity.App, settings []*entity.Setting, available *Set, now time.Time) (*Patch, error) {
	patch := &Patch{}
	scripts, files, mountNames, artifacts, err := readApp(settings)
	if err != nil {
		return nil, err
	}
	offered := map[string]bool{}
	for _, script := range available.Scripts {
		offered[script.Name] = true
		have := scripts[script.Name]
		if have == nil {
			if files[script.Name] {
				return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail(
					"the app has a config file called %s of its own, which the scripts would replace", script.Name)
			}
			file, mount, err := newScript(app, script, mountNames, now)
			if err != nil {
				return nil, err
			}
			patch.Created = append(patch.Created, file, mount)
			patch.LocksAdded = append(patch.LocksAdded, file.ID, mount.ID)
			continue
		}
		if have.data.Content != script.Content || have.data.Base64 {
			have.data.Content, have.data.Base64 = script.Content, false
			if err := have.file.SetData(have.data); err != nil {
				return nil, hperrors.Wrap(err)
			}
			patch.Updated = append(patch.Updated, stamp(have.file, now))
		}
		if have.target.Path != script.Path || have.target.Mode != script.Mode {
			have.target.Path, have.target.Mode = script.Path, script.Mode
			if err := have.mount.SetData(have.mountData); err != nil {
				return nil, hperrors.Wrap(err)
			}
			patch.Updated = append(patch.Updated, stamp(have.mount, now))
		}
	}
	for _, name := range sortedKeys(scripts) {
		if offered[name] {
			continue
		}
		have := scripts[name]
		for _, setting := range []*entity.Setting{have.file, have.mount} {
			setting.DeletedAt = now
			setting.UpdateVer++
			patch.Deleted = append(patch.Deleted, setting)
			patch.LocksRemoved = append(patch.LocksRemoved, setting.ID)
		}
	}
	if err := planArtifacts(app, artifacts, available, now, patch); err != nil {
		return nil, err
	}
	return patch, nil
}

// readApp reads an app's settings: its scripts by name, the names of all its
// config files and setting mounts, and its artifacts setting.
func readApp(settings []*entity.Setting) (scripts map[string]*appScript, files, mountNames map[string]bool,
	artifacts *entity.Setting, err error) {
	scripts, files, mountNames = map[string]*appScript{}, map[string]bool{}, map[string]bool{}
	byID := map[string]*entity.Setting{}
	for _, setting := range settings {
		if !setting.IsActive() || !setting.DeletedAt.IsZero() {
			continue
		}
		switch setting.Type { //nolint:exhaustive // what holds scripts and artifacts
		case base.SettingTypeConfigFile:
			byID[setting.ID] = setting
			files[setting.Name] = true
		case base.SettingTypeAppSettingMount:
			mountNames[setting.Name] = true
		case base.SettingTypeAppArtifacts:
			artifacts = setting
		}
	}
	for _, setting := range settings {
		if !setting.IsActive() || !setting.DeletedAt.IsZero() || setting.Type != base.SettingTypeAppSettingMount {
			continue
		}
		mount, err := setting.AsAppSettingMount()
		if err != nil {
			return nil, nil, nil, nil, hperrors.Wrap(err)
		}
		file := byID[mount.Source.ID]
		if file == nil {
			continue
		}
		for _, target := range mount.Files {
			if !isScriptPath(target.Path) {
				continue
			}
			data, err := file.AsConfigFile()
			if err != nil {
				return nil, nil, nil, nil, hperrors.Wrap(err)
			}
			scripts[data.Name] = &appScript{file: file, mount: setting, data: data, mountData: mount, target: target}
		}
	}
	return scripts, files, mountNames, artifacts, nil
}

// newScript is a script's config file and its setting mount, made as a
// template's swarmRef makes them, named after it.
func newScript(app *entity.App, script *Script, mountNames map[string]bool, now time.Time) (
	file, mount *entity.Setting, err error) {
	key := settingmountservice.EntryKeyFor(script.Name)
	if !settingmountservice.ValidEntryKey(key) || mountNames[key] {
		return nil, nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
			WithExtraDetail("%s: its mount would be called %q, which the app has or cannot have", script.Name, key)
	}
	mountNames[key] = true
	file = newAppSetting(app, base.SettingTypeConfigFile, script.Name, entity.CurrentConfigFileVersion, now)
	if err = file.SetData(&entity.ConfigFile{Name: script.Name, Content: script.Content}); err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	mount = newAppSetting(app, base.SettingTypeAppSettingMount, key, entity.CurrentAppSettingMountVersion, now)
	err = mount.SetData(&entity.AppSettingMount{Source: entity.ObjectID{ID: file.ID},
		Files: []*entity.AppSettingMountFile{{Part: mountPart, Path: script.Path, Mode: script.Mode}}})
	return file, mount, hperrors.Wrap(err)
}

// planArtifacts replaces the app's artifacts with available's: the setting is
// changed, made, or deleted when the set has none.
func planArtifacts(app *entity.App, setting *entity.Setting, available *Set, now time.Time, patch *Patch) error {
	switch {
	case available.Artifacts == nil && setting != nil:
		setting.DeletedAt = now
		setting.UpdateVer++
		patch.Deleted = append(patch.Deleted, setting)
	case available.Artifacts != nil && setting == nil:
		created := newAppSetting(app, base.SettingTypeAppArtifacts, "", entity.CurrentAppArtifactsVersion, now)
		if err := created.SetData(available.Artifacts); err != nil {
			return hperrors.Wrap(err)
		}
		patch.Created = append(patch.Created, created)
	case available.Artifacts != nil:
		current, err := setting.AsAppArtifactsSettings()
		if err != nil {
			return hperrors.Wrap(err)
		}
		if (&Set{Artifacts: current}).Fingerprint() == (&Set{Artifacts: available.Artifacts}).Fingerprint() {
			return nil
		}
		if err = setting.SetData(available.Artifacts); err != nil {
			return hperrors.Wrap(err)
		}
		patch.Updated = append(patch.Updated, stamp(setting, now))
	}
	return nil
}

func newAppSetting(app *entity.App, typ base.SettingType, name string, version int, now time.Time) *entity.Setting {
	return &entity.Setting{ID: gofn.Must(ulid.NewStringULID()), Scope: base.ObjectScopeApp, ObjectID: app.ID,
		Type: typ, Name: name, Status: base.SettingStatusActive, Version: version, CreatedAt: now, UpdatedAt: now}
}

// stamp marks a setting changed, once however many of its parts changed.
func stamp(setting *entity.Setting, now time.Time) *entity.Setting {
	if !setting.UpdatedAt.Equal(now) {
		setting.UpdateVer++
		setting.UpdatedAt = now
	}
	return setting
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}
