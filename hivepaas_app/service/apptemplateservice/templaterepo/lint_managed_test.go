package templaterepo

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// managedLintTemplateYAML is a managed template whose data volume is mounted at
// MOUNT, so that a case can put the mount over the artifact's path.
const managedLintTemplateYAML = `
apiVersion: hivepaas.com/v1
kind: AppTemplate
metadata:
  name: pgm
  title: Postgres managed
  tagline: A managed Postgres
  description: Managed.
  categories: [databases/sql]
  icon: icons/demo.svg
  requires: {versionCode: v000002, features: [managed-databases]}
parameters:
  - {name: dataVolume, title: Data volume, type: volume}
managed: {engine: postgres, engineApi: 1}
artifacts:
  - name: wal-g
    version: 3.0.8
    mountPath: /opt/hivepaas/bin/wal-g
    files:
      amd64:
        url: https://github.com/wal-g/wal-g/releases/download/v3.0.8/wal-g-pg-20.04-amd64
        sha256: 9a09c2b1afad6a4e7d87444b34726dd098b60ec816032af921d2f887e6e285c5
        size: 66978616
versions:
  - {name: "17", release: "17.6", default: true, image: "postgres:17.6-trixie"}
app:
  deployment:
    source: {activeMethod: image, imageSource: {image: "${{ image }}"}}
    storage:
      mounts:
        MOUNT: {type: volume, source: "${{ params.dataVolume }}"}
  settings:
    kind: {category: database, engine: postgres, database: {}}
`

func managedLintTemplate(mount string) string {
	return strings.Replace(managedLintTemplateYAML, "MOUNT", mount, 1)
}

func lintManaged(t *testing.T, template string) string {
	t.Helper()
	_, problems := loadAndLint(t, withFile(validRepoFS(), "templates/pgm.yaml", template))
	messages := make([]string, 0, len(problems))
	for _, problem := range problems {
		messages = append(messages, problem.String())
	}
	return strings.Join(messages, "\n")
}

func TestLintAcceptsAManagedTemplate(t *testing.T) {
	assert.Empty(t, lintManaged(t, managedLintTemplate("/var/lib/postgresql/data")))
}

func TestLintRefusesAnArtifactUnderAVolumeMount(t *testing.T) {
	for name, mount := range map[string]string{
		"the artifact's directory": "/opt/hivepaas/bin",
		"a parent of it":           "/opt",
		"the artifact itself":      "/opt/hivepaas/bin/wal-g",
	} {
		t.Run(name, func(t *testing.T) {
			assert.Contains(t, lintManaged(t, managedLintTemplate(mount)),
				`artifact "wal-g" at /opt/hivepaas/bin/wal-g is under the mount at `+mount)
		})
	}
}

func TestLintRefusesAnArtifactUnderAVersionsMount(t *testing.T) {
	// The version is written out long, to give it an override adding a mount.
	template := managedLintTemplate("/var/lib/postgresql/data")
	template = strings.Replace(template,
		`  - {name: "17", release: "17.6", default: true, image: "postgres:17.6-trixie"}`, `  - name: "17"
    release: "17.6"
    default: true
    image: "postgres:17.6-trixie"
    override:
      app:
        deployment:
          storage:
            mounts:
              /opt/hivepaas: {type: volume, source: "${{ params.dataVolume }}", volumeOptions: {subpath: x}}`, 1)

	assert.Contains(t, lintManaged(t, template), `version 17: artifact "wal-g" at /opt/hivepaas/bin/wal-g is under`)
}

func TestBuildIndexMarksAManagedTemplate(t *testing.T) {
	repo, problems := loadAndLint(t,
		withFile(validRepoFS(), "templates/pgm.yaml", managedLintTemplate("/var/lib/postgresql/data")))
	assert.Empty(t, problems)

	index, err := BuildIndex(repo)
	assert.NoError(t, err)
	assert.Equal(t, "postgres", index.FindTemplate("pgm").Managed)
	assert.Equal(t, []string{"managed-databases"}, index.FindTemplate("pgm").Requires.Features)
	assert.Empty(t, index.FindTemplate("demo").Managed)
}
