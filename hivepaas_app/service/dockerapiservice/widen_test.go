package dockerapiservice

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

func TestWidens(t *testing.T) {
	base := func() *entity.AppDockerAPISettings {
		return &entity.AppDockerAPISettings{
			Images: []string{"autobase/automation"}, SharedDirs: []string{"/data"},
			Allow: []string{"exec"}, Limits: entity.AppDockerAPILimits{Containers: 3},
		}
	}
	for name, tc := range map[string]struct {
		prev, next *entity.AppDockerAPISettings
		change     func(*entity.AppDockerAPISettings)
		want       bool
	}{
		"access where there was none": {next: base(), want: true},
		"turning it off":              {prev: base(), want: false},
		"no change":                   {prev: base(), next: base(), want: false},
		"an image added": {prev: base(), next: base(), want: true, change: func(s *entity.AppDockerAPISettings) {
			s.Images = append(s.Images, "alpine")
		}},
		"an image removed": {prev: base(), next: base(), want: false, change: func(s *entity.AppDockerAPISettings) {
			s.Images = nil
		}},
		"a directory added": {prev: base(), next: base(), want: true, change: func(s *entity.AppDockerAPISettings) {
			s.SharedDirs = append(s.SharedDirs, "/cache")
		}},
		"a network added": {prev: base(), next: base(), want: true, change: func(s *entity.AppDockerAPISettings) {
			s.Networks = []string{entity.DockerAPINetworkEnv}
		}},
		"a group added": {prev: base(), next: base(), want: true, change: func(s *entity.AppDockerAPISettings) {
			s.Allow = append(s.Allow, "nestedSocket")
		}},
		"more containers": {prev: base(), next: base(), want: true, change: func(s *entity.AppDockerAPISettings) {
			s.Limits.Containers = 4
		}},
		"the default is more than three": {prev: base(), next: base(), want: true,
			change: func(s *entity.AppDockerAPISettings) { s.Limits.Containers = 0 }},
		"more memory than the default": {prev: base(), next: base(), want: true,
			change: func(s *entity.AppDockerAPISettings) { s.Limits.Memory = 2 * unit.GB }},
		"less memory than the default": {prev: base(), next: base(), want: false,
			change: func(s *entity.AppDockerAPISettings) { s.Limits.Memory = 512 * unit.MB }},
		"more cpus than the default": {prev: base(), next: base(), want: true,
			change: func(s *entity.AppDockerAPISettings) { s.Limits.CPUs = 1.5 }},
	} {
		if tc.change != nil {
			tc.change(tc.next)
		}
		assert.Equal(t, tc.want, Widens(tc.prev, tc.next), name)
	}
}

// "*" is every image already: naming one more lets the app run nothing new.
func TestAnImageAddedToEveryImageDoesNotWiden(t *testing.T) {
	prev := &entity.AppDockerAPISettings{Images: []string{"*"}}
	next := &entity.AppDockerAPISettings{Images: []string{"*", "alpine"}}

	assert.False(t, Widens(prev, next))
}

// Host mode covers everything the proxy could allow: entering it widens,
// leaving it never does.
func TestWidensAcrossModes(t *testing.T) {
	proxy := &entity.AppDockerAPISettings{Images: []string{"*"}, Limits: entity.AppDockerAPILimits{Containers: 50}}
	host := &entity.AppDockerAPISettings{Mode: entity.DockerAPIModeHost}

	assert.True(t, Widens(proxy, host))
	assert.True(t, Widens(nil, host))
	assert.False(t, Widens(host, proxy))
	assert.False(t, Widens(host, host))
}

func TestEntersHostMode(t *testing.T) {
	proxy := &entity.AppDockerAPISettings{Images: []string{"*"}}
	host := &entity.AppDockerAPISettings{Mode: entity.DockerAPIModeHost}

	assert.True(t, EntersHostMode(nil, host), "from no access, or from host mode turned off")
	assert.True(t, EntersHostMode(proxy, host))
	assert.False(t, EntersHostMode(host, host))
	assert.False(t, EntersHostMode(host, proxy))
	assert.False(t, EntersHostMode(nil, proxy))
}
