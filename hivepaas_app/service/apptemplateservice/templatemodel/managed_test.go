package templatemodel

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// managedTemplateYAML is the smallest managed template: one engine, its one
// artifact for both architectures, and Debian images only.
const managedTemplateYAML = `
apiVersion: hivepaas.com/v1
kind: AppTemplate
metadata:
  name: pgm
  title: Postgres managed
  tagline: A managed Postgres
  description: Managed.
  categories: [databases/sql]
  icon: icons/pgm.svg
  requires: {versionCode: v000002, features: [managed-databases]}
parameters:
  - {name: dataVolume, title: Data volume, type: volume}
managed:
  engine: postgres
  engineApi: 1
  postgres:
    archiveTimeout: 60s
    baseBackupSchedule: "0 2 * * *"
    retainFull: 7
    restoreTestSchedule: "0 4 * * 0"
artifacts:
  - name: wal-g
    version: 3.0.8
    mountPath: /opt/hivepaas/bin/wal-g
    files:
      amd64:
        url: https://github.com/wal-g/wal-g/releases/download/v3.0.8/wal-g-pg-20.04-amd64
        sha256: 9a09c2b1afad6a4e7d87444b34726dd098b60ec816032af921d2f887e6e285c5
        size: 66978616
      arm64:
        url: https://github.com/wal-g/wal-g/releases/download/v3.0.8/wal-g-pg-20.04-aarch64
        sha256: 1111111111111111111111111111111111111111111111111111111111111111
        size: 64000000
versions:
  - {name: "18", release: "18.0", default: true, image: "postgres:18.0-trixie"}
  - {name: "17", release: "17.6", image: "postgres:17.6-trixie"}
app:
  deployment:
    source: {activeMethod: image, imageSource: {image: "${{ image }}"}}
`

func mustDecodeManaged(t *testing.T) *Template {
	t.Helper()
	tmpl, err := DecodeTemplate([]byte(managedTemplateYAML))
	assert.NoError(t, err)
	return tmpl
}

func TestDecodeManagedTemplate(t *testing.T) {
	tmpl := mustDecodeManaged(t)
	assert.NoError(t, tmpl.Validate("pgm"))

	assert.True(t, tmpl.IsManaged())
	assert.Equal(t, ManagedEnginePostgres, tmpl.Managed.Engine)
	assert.Equal(t, 1, tmpl.Managed.EngineAPI)
	assert.Equal(t, 60*time.Second, tmpl.Managed.Postgres.ArchiveTimeout.Duration)
	assert.Equal(t, 7, tmpl.Managed.Postgres.RetainFull)
	assert.Equal(t, []string{"managed-databases"}, tmpl.Metadata.Requires.Features)

	walg := tmpl.FindArtifact("wal-g")
	assert.Equal(t, "/opt/hivepaas/bin/wal-g", walg.MountPath)
	assert.Equal(t, int64(66978616), walg.Files[base.CPUArchAMD64].Size)
	assert.NotNil(t, walg.Files[base.CPUArchARM64])
	assert.Nil(t, tmpl.FindArtifact("other"))
}

func TestAnOrdinaryTemplateIsNotManaged(t *testing.T) {
	assert.False(t, mustDecode(t).IsManaged())
}

func TestDecodeManagedTemplateRefusesUnknownFields(t *testing.T) {
	with := func(old, added string) string {
		return strings.Replace(managedTemplateYAML, old, old+"\n"+added, 1)
	}
	for name, yaml := range map[string]string{
		"in managed":        with("  engineApi: 1", "  surprise: 1"),
		"in an engine":      with("    retainFull: 7", "    surprise: 1"),
		"in an artifact":    with("    version: 3.0.8", "    surprise: 1"),
		"in a file":         with("        size: 66978616", "        surprise: 1"),
		"a duration typo":   strings.Replace(managedTemplateYAML, "archiveTimeout: 60s", "archiveTimeout: sixty", 1),
		"a number duration": strings.Replace(managedTemplateYAML, "archiveTimeout: 60s", "archiveTimeout: 60", 1),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := DecodeTemplate([]byte(yaml))
			assert.Error(t, err)
		})
	}
}

func TestValidateManagedReportsEveryProblem(t *testing.T) {
	file := func(tm *Template) *ArtifactFile { return tm.Artifacts[0].Files[base.CPUArchAMD64] }
	pg := func(tm *Template) *ManagedPostgresDefaults { return tm.Managed.Postgres }
	cases := map[string]struct {
		mutate func(tm *Template)
		want   string
	}{
		"an old version code": {func(tm *Template) { tm.Metadata.Requires.VersionCode = "v000001" },
			"requires.versionCode v000002 or later"},
		"features on an old code": {func(tm *Template) {
			tm.Managed, tm.Artifacts = nil, nil
			tm.Metadata.Requires.VersionCode = "v000001"
		}, "requires.features needs requires.versionCode v000002"},
		"a bad feature name": {func(tm *Template) { tm.Metadata.Requires.Features = []string{"Managed DB"} },
			"requires.features"},
		"a feature twice": {func(tm *Template) {
			tm.Metadata.Requires.Features = []string{"managed-databases", "managed-databases"}
		}, "listed twice"},
		"an unknown engine":   {func(tm *Template) { tm.Managed.Engine = "oracle" }, `engine "oracle"`},
		"an api too new":      {func(tm *Template) { tm.Managed.EngineAPI = 2 }, "engineApi 2"},
		"no api":              {func(tm *Template) { tm.Managed.EngineAPI = 0 }, "engineApi 0"},
		"a short archive":     {func(tm *Template) { pg(tm).ArchiveTimeout.Duration = time.Second }, "archiveTimeout"},
		"a long archive":      {func(tm *Template) { pg(tm).ArchiveTimeout.Duration = 2 * time.Hour }, "archiveTimeout"},
		"a bad schedule":      {func(tm *Template) { pg(tm).BaseBackupSchedule = "nightly" }, "baseBackupSchedule"},
		"a bad test schedule": {func(tm *Template) { pg(tm).RestoreTestSchedule = "* *" }, "restoreTestSchedule"},
		"no retention":        {func(tm *Template) { pg(tm).RetainFull = -1 }, "retainFull"},
		"the engine's artifact missing": {func(tm *Template) { tm.Artifacts[0].Name = "pgbackrest" },
			`needs the artifact "wal-g"`},
		"components": {func(tm *Template) {
			tm.Components = []*Component{{Name: "web", Title: "Web", Primary: true, App: tm.App}}
			tm.App = nil
		}, "a managed template has one app"},
		"an alpine image": {func(tm *Template) { tm.Versions[0].Image = "postgres:18.0-alpine3.22" },
			"glibc"},
		"artifacts without managed": {func(tm *Template) { tm.Managed = nil }, "artifacts are mounted by a managed engine"},
		"an artifact twice": {func(tm *Template) { tm.Artifacts = append(tm.Artifacts, tm.Artifacts[0]) },
			"declared twice"},
		"no version": {func(tm *Template) { tm.Artifacts[0].Version = "" }, "version is required"},
		"a relative path": {func(tm *Template) { tm.Artifacts[0].MountPath = "bin/wal-g" },
			"/opt/hivepaas/bin/"},
		"a path elsewhere": {func(tm *Template) { tm.Artifacts[0].MountPath = "/usr/bin/wal-g" },
			"/opt/hivepaas/bin/"},
		"a path not clean": {func(tm *Template) { tm.Artifacts[0].MountPath = "/opt/hivepaas/bin/../wal-g" },
			"/opt/hivepaas/bin/"},
		"no amd64":     {func(tm *Template) { delete(tm.Artifacts[0].Files, base.CPUArchAMD64) }, "amd64"},
		"an http url":  {func(tm *Template) { file(tm).URL = "http://example.com/wal-g" }, "https"},
		"a short hash": {func(tm *Template) { file(tm).SHA256 = "9a09" }, "sha256"},
		"an upper hash": {func(tm *Template) { file(tm).SHA256 = strings.ToUpper(file(tm).SHA256) },
			"sha256"},
		"no size":  {func(tm *Template) { file(tm).Size = 0 }, "size"},
		"too big":  {func(tm *Template) { file(tm).Size = MaxArtifactSize + 1 }, "size"},
		"no files": {func(tm *Template) { tm.Artifacts[0].Files = nil }, "amd64"},
		"an unknown arch": {func(tm *Template) {
			tm.Artifacts[0].Files["riscv64"] = tm.Artifacts[0].Files[base.CPUArchARM64]
		}, `architecture "riscv64"`},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			tmpl := mustDecodeManaged(t)
			c.mutate(tmpl)
			err := tmpl.Validate("pgm")
			assert.Error(t, err)
			assert.Contains(t, errorDetail(t, err), c.want)
		})
	}
}

func TestAManagedTemplateMayLeaveTheEngineDefaultsOut(t *testing.T) {
	tmpl := mustDecodeManaged(t)
	tmpl.Managed.Postgres = nil
	assert.NoError(t, tmpl.Validate("pgm"))
}

func TestAnAlpineVariantNameIsRefusedToo(t *testing.T) {
	tmpl := mustDecodeManaged(t)
	tmpl.Variants = []*Variant{{Name: "alpine", Title: "Alpine", Default: true}}
	for _, version := range tmpl.Versions {
		version.Images = map[string]string{"alpine": version.Image}
		version.Image = ""
	}
	err := tmpl.Validate("pgm")
	assert.Error(t, err)
	assert.Contains(t, errorDetail(t, err), "glibc")
}

func TestEveryEngineContractNamesItsArtifacts(t *testing.T) {
	for engine, contract := range ManagedEngines {
		assert.LessOrEqual(t, 1, contract.MinAPI, engine)
		assert.LessOrEqual(t, contract.MinAPI, contract.MaxAPI, engine)
		assert.NotEmpty(t, contract.Artifacts, engine)
	}
}
