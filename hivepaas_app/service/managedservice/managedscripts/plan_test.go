package managedscripts

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
)

var planNow = time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC)

func byID(settings []*entity.Setting) map[string]*entity.Setting {
	out := map[string]*entity.Setting{}
	for _, setting := range settings {
		out[setting.ID] = setting
	}
	return out
}

// An app's scripts become the revision's: a changed one's file and mount are
// changed in place, a new one gets a file and a mount, one no longer shipped
// loses both, and the artifacts are replaced. The app's own config files stay.
func TestPlanReplacesTheScripts(t *testing.T) {
	app := &entity.App{ID: "app-1", Key: "shop-db"}
	settings := appSettings(map[string]string{"walg-lib": "# old lib\n", "walg-old": "x",
		"walg-entrypoint": "#!/bin/bash\necho ${HOME} ${{ kept }}\n"},
		map[string]string{"walg-lib": "444", "walg-old": "555", "walg-entrypoint": "555"}, walg("3.0.8", "cccc"))
	own := &entity.Setting{ID: "cf-mine", Type: base.SettingTypeConfigFile, Name: "my-cnf",
		Status: base.SettingStatusActive}
	own.MustSetData(&entity.ConfigFile{Name: "my-cnf", Content: "[mysqld]\n"})
	settings = append(settings, own)
	available, err := FromTemplate(testTemplate(t), "11.8")
	assert.NoError(t, err)
	available.Scripts = append(available.Scripts, &Script{Name: "walg-new", Path: "/opt/hivepaas/bin/walg-new",
		Mode: 0o555, Content: "#!/bin/sh\n", SHA256: sha256Hex("#!/bin/sh\n")})

	patch, err := Plan(app, settings, available, planNow)

	assert.NoError(t, err)
	updated := byID(patch.Updated)
	if assert.Contains(t, updated, "cf-walg-lib") {
		file := updated["cf-walg-lib"].MustAsConfigFile()
		assert.Equal(t, "# lib\nbin=/opt/hivepaas/bin\n", file.Content)
		assert.Equal(t, planNow, updated["cf-walg-lib"].UpdatedAt)
	}
	assert.NotContains(t, updated, "cf-walg-entrypoint", "unchanged")
	assert.NotContains(t, updated, "cf-mine", "the app's own")
	if assert.Contains(t, updated, "art") {
		assert.Equal(t, "3.0.9", updated["art"].MustAsAppArtifactsSettings().Artifacts[0].Version)
	}

	if assert.Len(t, patch.Created, 2) {
		file, mount := patch.Created[0], patch.Created[1]
		assert.Equal(t, base.SettingTypeConfigFile, file.Type)
		assert.Equal(t, "walg-new", file.Name)
		assert.Equal(t, "app-1", file.ObjectID)
		assert.Equal(t, base.ObjectScopeApp, file.Scope)
		assert.Equal(t, "#!/bin/sh\n", file.MustAsConfigFile().Content)
		assert.Equal(t, base.SettingTypeAppSettingMount, mount.Type)
		data := mount.MustAsAppSettingMount()
		assert.Equal(t, file.ID, data.Source.ID)
		assert.Equal(t, []*entity.AppSettingMountFile{{Part: "content", Path: "/opt/hivepaas/bin/walg-new",
			Mode: fileutil.FileMode(0o555)}}, data.Files)
		assert.ElementsMatch(t, []string{file.ID, mount.ID}, patch.LocksAdded)
	}

	deleted := byID(patch.Deleted)
	assert.Len(t, deleted, 2)
	assert.Contains(t, deleted, "cf-walg-old")
	assert.Contains(t, deleted, "mt-walg-old")
	assert.Equal(t, planNow, deleted["cf-walg-old"].DeletedAt)
	assert.ElementsMatch(t, []string{"cf-walg-old", "mt-walg-old"}, patch.LocksRemoved)

	// Applied, the app has what the revision offers.
	var after []*entity.Setting
	for _, setting := range settings {
		if setting.DeletedAt.IsZero() {
			after = append(after, setting)
		}
	}
	after = append(after, patch.Created...)
	set, err := FromSettings(after)
	assert.NoError(t, err)
	assert.Equal(t, available.Fingerprint(), set.Fingerprint())
}

// A script's mode moves with the revision, on its mount.
func TestPlanChangesAMode(t *testing.T) {
	settings := appSettings(map[string]string{"walg-lib": "# lib\nbin=/opt/hivepaas/bin\n",
		"walg-entrypoint": "#!/bin/bash\necho ${HOME} ${{ kept }}\n"},
		map[string]string{"walg-lib": "555", "walg-entrypoint": "555"}, walg("3.0.9", "aaaa"))
	available, err := FromTemplate(testTemplate(t), "11.8")
	assert.NoError(t, err)

	patch, err := Plan(&entity.App{ID: "app-1"}, settings, available, planNow)

	assert.NoError(t, err)
	if assert.Len(t, patch.Updated, 1) {
		assert.Equal(t, "mt-walg-lib", patch.Updated[0].ID)
		assert.Equal(t, fileutil.FileMode(0o444), patch.Updated[0].MustAsAppSettingMount().Files[0].Mode)
	}
	assert.Empty(t, patch.Created)
	assert.Empty(t, patch.Deleted)
}

// A new script whose name an app's own config file already has is refused:
// the file is the person's.
func TestPlanRefusesANameTheAppUses(t *testing.T) {
	own := &entity.Setting{ID: "cf-mine", Type: base.SettingTypeConfigFile, Name: "walg-new",
		Status: base.SettingStatusActive}
	own.MustSetData(&entity.ConfigFile{Name: "walg-new", Content: "mine"})
	available := &Set{Scripts: []*Script{{Name: "walg-new", Path: "/opt/hivepaas/bin/walg-new", Content: "x"}}}

	_, err := Plan(&entity.App{ID: "app-1"}, []*entity.Setting{own}, available, planNow)

	assert.Error(t, err)
}
