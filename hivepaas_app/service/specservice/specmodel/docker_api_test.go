package specmodel

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

func TestDockerAPIInReadsTheBlock(t *testing.T) {
	doc := decodeDoc(t, `
settings:
  dockerApi:
    images: [autobase/automation]
    sharedDirs: [/var/lib/autobase/ansible]
    limits: {containers: 3, memory: 2gb, cpus: 1.5}
`)
	got, err := DockerAPIIn(doc.Settings)
	assert.NoError(t, err)
	assert.Equal(t, &entity.AppDockerAPISettings{
		Images: []string{"autobase/automation"}, SharedDirs: []string{"/var/lib/autobase/ansible"},
		Limits: entity.AppDockerAPILimits{Containers: 3, Memory: 2 * unit.GB, CPUs: 1.5},
	}, got)

	got, err = DockerAPIIn(decodeDoc(t, "settings:\n  routing: {port: 80}\n").Settings)
	assert.NoError(t, err)
	assert.Nil(t, got)

	_, err = DockerAPIIn(decodeDoc(t, "settings:\n  dockerApi: {images: [a], privileged: true}\n").Settings)
	assert.ErrorIs(t, err, hperrors.ErrSpecBlockInvalid)
}

func TestDockerAPIProblem(t *testing.T) {
	ok := func() *entity.AppDockerAPISettings {
		return &entity.AppDockerAPISettings{Images: []string{"alpine"}}
	}
	assert.Empty(t, DockerAPIProblem(ok()))

	tooMany := make([]string, MaxDockerAPIImages+1)
	for i := range tooMany {
		tooMany[i] = "alpine"
	}
	sixDirs := []string{"/a", "/b", "/c", "/d", "/e", "/f"}
	cases := map[string]func(s *entity.AppDockerAPISettings){
		"images: at least one":      func(s *entity.AppDockerAPISettings) { s.Images = nil },
		"images: at most":           func(s *entity.AppDockerAPISettings) { s.Images = tooMany },
		"images[0]":                 func(s *entity.AppDockerAPISettings) { s.Images = []string{"alp ine"} },
		"sharedDirs: at most":       func(s *entity.AppDockerAPISettings) { s.SharedDirs = sixDirs },
		"sharedDirs[0]":             func(s *entity.AppDockerAPISettings) { s.SharedDirs = []string{"data"} },
		"sharedDirs[0]: / is":       func(s *entity.AppDockerAPISettings) { s.SharedDirs = []string{"/"} },
		"sharedDirs[0]: /a/../b is": func(s *entity.AppDockerAPISettings) { s.SharedDirs = []string{"/a/../b"} },
		"networks[0]":               func(s *entity.AppDockerAPISettings) { s.Networks = []string{"hivepaas_net"} },
		"allow[0]":                  func(s *entity.AppDockerAPISettings) { s.Allow = []string{"build"} },
		"limits.containers":         func(s *entity.AppDockerAPISettings) { s.Limits.Containers = 101 },
		"limits.memory":             func(s *entity.AppDockerAPISettings) { s.Limits.Memory = 1 * unit.MB },
		"limits.cpus":               func(s *entity.AppDockerAPISettings) { s.Limits.CPUs = -1 },
	}
	for want, change := range cases {
		settings := ok()
		change(settings)
		problem := DockerAPIProblem(settings)
		assert.True(t, strings.Contains(problem, want), "%s: %q", want, problem)
	}
}

// Host mode takes nothing of the proxy's policy, and there is no third mode.
func TestDockerAPIProblemOfTheMode(t *testing.T) {
	assert.Empty(t, DockerAPIProblem(&entity.AppDockerAPISettings{Mode: entity.DockerAPIModeHost}))
	assert.Empty(t, DockerAPIProblem(&entity.AppDockerAPISettings{
		Mode: entity.DockerAPIModeProxy, Images: []string{"alpine"},
	}))
	assert.Contains(t, DockerAPIProblem(&entity.AppDockerAPISettings{Mode: "root", Images: []string{"alpine"}}),
		"settings.dockerApi.mode")
}

// The node's own socket is given by an administrator, never by a document a
// template renders.
func TestCheckBuildableRefusesHostMode(t *testing.T) {
	err := CheckBuildable(decodeDoc(t, "settings:\n  dockerApi: {mode: host}\n"))

	assert.ErrorIs(t, err, hperrors.ErrSpecBlockInvalid)
	assert.Contains(t, buildableErrorDetail(t, err), "settings.dockerApi.mode")
}

func TestCheckBuildableTakesTheDockerAPIBlock(t *testing.T) {
	doc := decodeDoc(t, `
deployment:
  storage:
    mounts:
      /var/lib/autobase: {type: volume, source: vol-1}
settings:
  dockerApi: {images: [autobase/automation], sharedDirs: [/var/lib/autobase/ansible]}
`)
	assert.NoError(t, CheckBuildable(doc))

	doc.Settings["dockerApi"] = map[string]any{"images": []any{"alpine"}, "sharedDirs": []any{"/srv/work"}}
	err := CheckBuildable(doc)
	assert.ErrorIs(t, err, hperrors.ErrSpecBlockInvalid)
	assert.Contains(t, buildableErrorDetail(t, err), "/srv/work is on none of the app's storage mounts")

	doc.Settings["dockerApi"] = map[string]any{"images": []any{}}
	assert.ErrorIs(t, CheckBuildable(doc), hperrors.ErrSpecBlockInvalid)
}
