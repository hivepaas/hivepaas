package templatemodel

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const dockerAPITemplateYAML = `
apiVersion: hivepaas.com/v1
kind: AppTemplate
metadata:
  name: autobase
  title: Autobase
  tagline: PostgreSQL clusters
  description: Test.
  categories: [databases/sql]
  icon: icons/autobase.svg
  requires: {versionCode: v000001}
versions:
  - {name: "2.11", release: "2.11.0", default: true, image: "autobase/console:2.11.0"}
app:
  deployment:
    source: {activeMethod: image, imageSource: {image: "${{ image }}"}}
    storage:
      mounts:
        /var/lib/autobase: {type: volume, source: "${{ params.dataVolume }}"}
  settings:
    dockerApi:
      images: [autobase/automation]
      sharedDirs: [/var/lib/autobase/ansible]
parameters:
  - {name: dataVolume, title: Data volume, type: volume}
`

func TestDockerAPIIsReadFromTheTemplate(t *testing.T) {
	tmpl, err := DecodeTemplate([]byte(dockerAPITemplateYAML))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	access, err := tmpl.DockerAPI()
	assert.NoError(t, err)
	if assert.NotNil(t, access) {
		assert.Equal(t, []string{"autobase/automation"}, access.Images)
	}
	assert.True(t, tmpl.RequiresDockerAPI())
}

func TestDockerAPIMayNotDependOnWhatAPersonFillsIn(t *testing.T) {
	tmpl, err := DecodeTemplate([]byte(strings.Replace(dockerAPITemplateYAML, "images: [autobase/automation]",
		`images: ["${{ params.image }}"]`, 1)))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Contains(t, errorDetail(t, tmpl.Validate("autobase.yaml")), "app.settings.dockerApi: a placeholder here")
}

// A template never gives its app the node's own socket: that is an
// administrator's choice, on the app's screen.
func TestATemplateMayNotAskForTheNodesSocket(t *testing.T) {
	tmpl, err := DecodeTemplate([]byte(strings.Replace(dockerAPITemplateYAML, "images: [autobase/automation]",
		"mode: host\n      images: [autobase/automation]", 1)))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Contains(t, errorDetail(t, tmpl.Validate("autobase.yaml")), "app.settings.dockerApi.mode")
}

func TestAVersionMayNotChangeTheDockerAPI(t *testing.T) {
	tmpl, err := DecodeTemplate([]byte(strings.Replace(dockerAPITemplateYAML,
		`default: true, image: "autobase/console:2.11.0"}`,
		`default: true, image: "autobase/console:2.11.0",`+
			` override: {app: {settings: {dockerApi: {images: ["*"]}}}}}`, 1)))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Contains(t, errorDetail(t, tmpl.Validate("autobase.yaml")), "the Docker API is declared once, in app")
}
