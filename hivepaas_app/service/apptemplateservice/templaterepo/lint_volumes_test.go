package templaterepo

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// filesTemplateYAML keeps its files and its data on two volume parameters, each
// with a subpath of its own. MOUNTS and OVERRIDE are where the cases differ.
const filesTemplateYAML = `
apiVersion: hivepaas.com/v1
kind: AppTemplate
metadata:
  name: files
  title: Files
  tagline: A file app
  description: Files.
  categories: [databases/sql]
  tags: [sql]
  icon: icons/demo.svg
  requires: {versionCode: v000001}
parameters:
  - {name: filesVolume, title: Files volume, type: volume}
  - {name: dataVolume, title: Data volume, type: volume}
versions:
  - name: "1"
    release: "1.0"
    default: true
    image: "files:1.0.0"
OVERRIDE
app:
  deployment:
    source: {activeMethod: image, imageSource: {image: "${{ image }}"}}
    storage:
      mounts:
MOUNTS
  settings:
    kind: {category: webapp, engine: files, webapp: {}}
`

const filesMountsYAML = `
        /files: {type: volume, source: "${{ params.filesVolume }}", volumeOptions: {subpath: files}}
        /data: {type: volume, source: "${{ params.dataVolume }}", volumeOptions: {subpath: data}}`

func filesTemplate(mounts, override string) string {
	return strings.NewReplacer("MOUNTS", strings.Trim(mounts, "\n"), "OVERRIDE", override).Replace(filesTemplateYAML)
}

func lintFiles(t *testing.T, files string) string {
	t.Helper()
	_, problems := loadAndLint(t, withFile(validRepoFS(), "templates/files.yaml", files))
	messages := make([]string, 0, len(problems))
	for _, problem := range problems {
		messages = append(messages, problem.String())
	}
	return strings.Join(messages, "\n")
}

func TestLintAcceptsVolumeMountsThatCannotMeet(t *testing.T) {
	cases := map[string]string{
		"a subpath each": filesMountsYAML,
		"one mount, whole": `
        /files: {type: volume, source: "${{ params.filesVolume }}"}`,
		// Another app's directory is shared on purpose.
		"through sourceApp": `
        /files: {type: volume, source: "${{ params.filesVolume }}", volumeOptions: {subpath: files}}
        /media: {type: volume, source: "${{ params.filesVolume }}", sourceApp: {app: media}}`,
		// A volume the template names itself is not one a user could give twice.
		"not a volume parameter": `
        /files: {type: volume, source: "${{ params.filesVolume }}", volumeOptions: {subpath: files}}
        /cache: {type: volume, source: cache}`,
	}
	for name, mounts := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Empty(t, lintFiles(t, filesTemplate(mounts, "")))
		})
	}
}

func TestLintRefusesVolumeMountsThatCouldMeet(t *testing.T) {
	cases := map[string]struct {
		mounts, override, want string
	}{
		"a whole mount beside others": {
			mounts: `
        /files: {type: volume, source: "${{ params.filesVolume }}"}
        /data: {type: volume, source: "${{ params.dataVolume }}", volumeOptions: {subpath: data}}`,
			want: "/files mounts a volume parameter whole, and the app mounts others",
		},
		"one subpath twice": {
			mounts: `
        /files: {type: volume, source: "${{ params.filesVolume }}", volumeOptions: {subpath: data}}
        /data: {type: volume, source: "${{ params.dataVolume }}", volumeOptions: {subpath: data}}`,
			want: `/data and /files both mount subpath "data" of a volume parameter`,
		},
		"in a version's override": {
			mounts: filesMountsYAML,
			override: `
    override:
      app:
        deployment:
          storage:
            mounts:
              /cache: {type: volume, source: "${{ params.dataVolume }}"}`,
			want: "version 1: /cache mounts a volume parameter whole",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Contains(t, lintFiles(t, filesTemplate(tc.mounts, tc.override)), tc.want)
		})
	}
}

// The templates published before the rule keep their layout: apps created from
// them have their data there.
func TestLintLeavesThePublishedVolumeLayoutsAlone(t *testing.T) {
	whole := filesTemplate(`
        /files: {type: volume, source: "${{ params.filesVolume }}"}
        /data: {type: volume, source: "${{ params.dataVolume }}", volumeOptions: {subpath: data}}`, "")
	exempt := strings.Replace(whole, "name: files", "name: jellyfin", 1)

	_, problems := loadAndLint(t, withFile(validRepoFS(), "templates/jellyfin.yaml", exempt))

	assert.Empty(t, problems)
}
