package managedscripts

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

const templateYAML = `
apiVersion: hivepaas.com/v1
kind: AppTemplate
metadata:
  name: mariadb-managed
  title: MariaDB
  tagline: Test
  description: Test.
  categories: [databases/sql]
  icon: icons/mariadb.svg
  requires: {versionCode: v000003}
parameters:
  - {name: dataVolume, title: Data volume, type: volume}
versions:
  - name: "11.8"
    release: "11.8.9"
    default: true
    image: "mariadb:11.8.9-noble"
  - name: "11.4"
    release: "11.4.9"
    deprecated: true
    image: "mariadb:11.4.9-noble"
    override:
      app:
        settings:
          configFiles:
            walg-lib:
              content: "# 11.4's lib\n"
managed:
  engine: mariadb
  engineApi: 1
artifacts:
  - name: wal-g
    version: 3.0.9
    mountPath: /opt/hivepaas/bin/wal-g
    files:
      amd64: {url: "https://example.com/wal-g-amd64", sha256: "aaaa", size: 10}
      arm64: {url: "https://example.com/wal-g-arm64", sha256: "bbbb", size: 10}
app:
  deployment:
    source:
      activeMethod: image
      imageSource: {image: "${{ image }}"}
    storage:
      mounts:
        /var/lib/mysql: {type: volume, source: "${{ params.dataVolume }}"}
  settings:
    configFiles:
      walg-lib:
        content: "# lib\nbin=/opt/hivepaas/bin\n"
        swarmRef: {file: {name: /opt/hivepaas/bin/walg-lib, mode: 444}}
      walg-entrypoint:
        content: "#!/bin/bash\necho ${HOME} $${{ kept }}\n"
        swarmRef: {file: {name: /opt/hivepaas/bin/walg-entrypoint, mode: 555}}
      my-cnf:
        content: "[mysqld]\n"
        swarmRef: {file: {name: /etc/mysql/conf.d/hivepaas.cnf, mode: 444}}
`

func testTemplate(t *testing.T) *templatemodel.Template {
	t.Helper()
	tmpl, err := templatemodel.DecodeTemplate([]byte(templateYAML))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return tmpl
}

// appSettings are an app's settings as its creation made them from scripts:
// a config file and a mount for each, and its artifacts.
func appSettings(scripts, modes map[string]string, artifacts *entity.AppArtifactsSettings) []*entity.Setting {
	var out []*entity.Setting
	for name, content := range scripts {
		file := &entity.Setting{ID: "cf-" + name, Type: base.SettingTypeConfigFile, Name: name,
			Status: base.SettingStatusActive}
		file.MustSetData(&entity.ConfigFile{Name: name, Content: content})
		mode, _ := fileutil.ParseFileMode(modes[name])
		mount := &entity.Setting{ID: "mt-" + name, Type: base.SettingTypeAppSettingMount, Name: name,
			Status: base.SettingStatusActive}
		mount.MustSetData(&entity.AppSettingMount{Source: entity.ObjectID{ID: file.ID},
			Files: []*entity.AppSettingMountFile{{Part: "content", Path: "/opt/hivepaas/bin/" + name, Mode: mode}}})
		out = append(out, file, mount)
	}
	if artifacts != nil {
		setting := &entity.Setting{ID: "art", Type: base.SettingTypeAppArtifacts, Status: base.SettingStatusActive}
		setting.MustSetData(artifacts)
		out = append(out, setting)
	}
	return out
}

func walg(version, amd64 string) *entity.AppArtifactsSettings {
	return &entity.AppArtifactsSettings{Artifacts: []*entity.AppArtifact{{Name: "wal-g", Version: version,
		MountPath: "/opt/hivepaas/bin/wal-g", Files: map[base.CPUArch]*entity.AppArtifactFile{
			"amd64": {URL: "https://example.com/wal-g-amd64", SHA256: amd64, Size: 10},
			"arm64": {URL: "https://example.com/wal-g-arm64", SHA256: "bbbb", Size: 10}}}}}
}

// What a template offers are its config files mounted under the script
// directory, as an app made from it has them - escapes undone - and its
// artifacts. Another config file is the person's to change, not a script.
func TestFromTemplate(t *testing.T) {
	set, err := FromTemplate(testTemplate(t), "11.8")

	assert.NoError(t, err)
	if assert.Len(t, set.Scripts, 2) {
		assert.Equal(t, "walg-entrypoint", set.Scripts[0].Name, "by path")
		assert.Equal(t, "/opt/hivepaas/bin/walg-entrypoint", set.Scripts[0].Path)
		assert.Equal(t, "#!/bin/bash\necho ${HOME} ${{ kept }}\n", set.Scripts[0].Content)
		assert.Equal(t, fileutil.FileMode(0o555), set.Scripts[0].Mode)
		assert.Equal(t, "walg-lib", set.Scripts[1].Name)
	}
	assert.Equal(t, "3.0.9", set.Artifacts.Artifacts[0].Version)

	deprecated, err := FromTemplate(testTemplate(t), "11.4")
	assert.NoError(t, err, "a deprecated version is read")
	assert.Equal(t, "# 11.4's lib\n", deprecated.Scripts[1].Content, "with its override")
}

// An app made from the template has the same scripts, and the same fingerprint;
// any difference in a script's content, its mode, or an artifact changes it.
func TestFingerprints(t *testing.T) {
	offered, err := FromTemplate(testTemplate(t), "11.8")
	assert.NoError(t, err)
	scripts := map[string]string{"walg-lib": "# lib\nbin=/opt/hivepaas/bin\n",
		"walg-entrypoint": "#!/bin/bash\necho ${HOME} ${{ kept }}\n"}
	modes := map[string]string{"walg-lib": "444", "walg-entrypoint": "555"}

	same, err := FromSettings(appSettings(scripts, modes, walg("3.0.9", "aaaa")))
	assert.NoError(t, err)
	assert.Equal(t, offered.Fingerprint(), same.Fingerprint())
	assert.Len(t, offered.Fingerprint(), 64)

	for name, settings := range map[string][]*entity.Setting{
		"content": appSettings(map[string]string{"walg-lib": "# old lib\n",
			"walg-entrypoint": scripts["walg-entrypoint"]}, modes, walg("3.0.9", "aaaa")),
		"mode": appSettings(scripts, map[string]string{"walg-lib": "555", "walg-entrypoint": "555"},
			walg("3.0.9", "aaaa")),
		"artifact":     appSettings(scripts, modes, walg("3.0.8", "cccc")),
		"no artifacts": appSettings(scripts, modes, nil),
	} {
		other, err := FromSettings(settings)
		assert.NoError(t, err, name)
		assert.NotEqual(t, offered.Fingerprint(), other.Fingerprint(), name)
	}
}

// The changes are per script, by path, and per artifact, by name.
func TestDiff(t *testing.T) {
	offered, err := FromTemplate(testTemplate(t), "11.8")
	assert.NoError(t, err)
	current, err := FromSettings(appSettings(map[string]string{"walg-lib": "# old lib\n", "walg-old": "x"},
		map[string]string{"walg-lib": "444", "walg-old": "555"}, walg("3.0.8", "cccc")))
	assert.NoError(t, err)

	changes := Diff(current, offered)

	assert.Equal(t, []*Change{
		{Kind: KindScript, Name: "/opt/hivepaas/bin/walg-entrypoint", Change: ChangeAdded},
		{Kind: KindScript, Name: "/opt/hivepaas/bin/walg-lib", Change: ChangeChanged},
		{Kind: KindScript, Name: "/opt/hivepaas/bin/walg-old", Change: ChangeRemoved},
		{Kind: KindArtifact, Name: "wal-g", Change: ChangeChanged, From: "3.0.8", To: "3.0.9"},
	}, changes)
	assert.Empty(t, Diff(offered, offered))
}

// Whether the revision can be applied: the same engine and engine API, the
// version there, and scripts in it.
func TestEvaluate(t *testing.T) {
	managed := &entity.AppManagedSettings{Engine: "mariadb", EngineAPI: 1}
	offered, err := FromTemplate(testTemplate(t), "11.8")
	assert.NoError(t, err)
	old, err := FromSettings(appSettings(map[string]string{"walg-lib": "# old lib\n"},
		map[string]string{"walg-lib": "444"}, walg("3.0.9", "aaaa")))
	assert.NoError(t, err)

	offer := Evaluate(managed, testTemplate(t), "11.8", old)
	assert.Equal(t, StateUpdateAvailable, offer.State)
	assert.Equal(t, offered.Fingerprint(), offer.Available.Fingerprint())
	assert.Len(t, offer.Changes, 2)

	assert.Equal(t, StateUpToDate, Evaluate(managed, testTemplate(t), "11.8", offered).State)

	for name, tc := range map[string]struct {
		mutate func(*templatemodel.Template)
		manage *entity.AppManagedSettings
		ver    string
		state  string
		reason string
	}{
		"no longer managed": {func(t *templatemodel.Template) { t.Managed = nil }, managed, "11.8",
			StateUnavailable, "not a managed"},
		"another engine": {func(t *templatemodel.Template) { t.Managed.Engine = "postgres" }, managed, "11.8",
			StateIncompatible, "postgres"},
		"another engine API": {func(t *templatemodel.Template) { t.Managed.EngineAPI = 2 }, managed, "11.8",
			StateIncompatible, "engine API 2"},
		"the version gone": {nil, managed, "10.11", StateUnavailable, "10.11"},
		"a placeholder in a script": {func(t *templatemodel.Template) {
			files := t.App["settings"].(map[string]any)["configFiles"].(map[string]any)
			files["walg-lib"].(map[string]any)["content"] = "${{ params.dataVolume }}"
		}, managed, "11.8", StateIncompatible, "placeholder"},
		"no scripts": {func(t *templatemodel.Template) {
			delete(t.App["settings"].(map[string]any), "configFiles")
		}, managed, "11.8", StateIncompatible, "no scripts"},
	} {
		tmpl := testTemplate(t)
		if tc.mutate != nil {
			tc.mutate(tmpl)
		}
		offer := Evaluate(tc.manage, tmpl, tc.ver, old)
		assert.Equal(t, tc.state, offer.State, name)
		assert.True(t, strings.Contains(offer.Reason, tc.reason), "%s: %s", name, offer.Reason)
	}
}
