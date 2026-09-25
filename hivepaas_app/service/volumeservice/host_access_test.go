package volumeservice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The socket, and every directory holding it, is the same grant: the Docker API
// of the node, which is given through an app's Docker API settings alone.
func TestReachesDockerSocket(t *testing.T) {
	for path, want := range map[string]bool{
		"/var/run/docker.sock":     true,
		"/var/run/docker.sock/":    true,
		"/var/run/./docker.sock":   true,
		"/var/run":                 true,
		"/var/run/":                true,
		"/var":                     true,
		"/":                        true,
		"/var/run/docker.sock.bak": false,
		"/var/lib/docker":          false,
		"/srv/data":                false,
		"":                         false,
		"relative/path":            false,
	} {
		assert.Equal(t, want, ReachesDockerSocket(path), path)
	}
}

func TestDriverOptsReachDockerSocket(t *testing.T) {
	assert.True(t, DriverOptsReachDockerSocket(map[string]string{"device": "/var/run"}))
	assert.False(t, DriverOptsReachDockerSocket(map[string]string{"device": "/srv/data"}))
	assert.False(t, DriverOptsReachDockerSocket(nil))
}

func TestDriverOptsNameHostPath(t *testing.T) {
	assert.True(t, DriverOptsNameHostPath(map[string]string{"device": "/srv/data"}))
	assert.True(t, DriverOptsNameHostPath(map[string]string{"type": "none"}))
	assert.True(t, DriverOptsNameHostPath(map[string]string{"o": "bind,rw"}))
	assert.False(t, DriverOptsNameHostPath(map[string]string{"size": "100m"}))
	assert.False(t, DriverOptsNameHostPath(nil))
}
