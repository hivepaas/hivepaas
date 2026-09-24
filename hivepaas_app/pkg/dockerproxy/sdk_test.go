package dockerproxy

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/stretchr/testify/assert"
)

// sdkCreate is a create as a Go client - act, Woodpecker, Autobase - sends it:
// whole structs, with every field it did not set at its zero value.
func sdkCreate(t *testing.T, cfg *container.Config, host *container.HostConfig,
	networking *network.NetworkingConfig) []byte {
	t.Helper()
	raw, err := json.Marshal(container.CreateRequest{Config: cfg, HostConfig: host, NetworkingConfig: networking})
	stop(t, assert.NoError(t, err))
	return raw
}

func TestCreateTakesWhatGoClientsSend(t *testing.T) {
	policy := testPolicy()
	policy.Images = []string{"*"}
	policy.Allow = []Group{GroupVolumes, GroupNetworks, GroupNestedSocket}
	w := newWorld(t, policy)

	// act: a job container on the network it made, a cache volume, the socket.
	act := sdkCreate(t,
		&container.Config{Image: "node:20", Entrypoint: []string{"/bin/sleep", "10800"}},
		&container.HostConfig{
			NetworkMode: "jobnet",
			Binds:       []string{"act-toolcache:/opt/hostedtoolcache", SocketPath + ":/var/run/docker.sock"},
		},
		&network.NetworkingConfig{EndpointsConfig: map[string]*network.EndpointSettings{
			"jobnet": {Aliases: []string{"job"}},
		}})
	status, raw := w.do(t, http.MethodPost, "/v1.44/containers/create", act)
	stop(t, assert.Equal(t, http.StatusCreated, status, string(raw)))
	path, _ := w.posted(t, "/containers/create")
	assert.Equal(t, "/v1.45/containers/create", path)

	// Autobase: host networking and the log directory it shares with the playbook.
	autobase := sdkCreate(t,
		&container.Config{Image: "autobase/automation:2.11.0", Tty: true},
		&container.HostConfig{NetworkMode: "host", Mounts: []mount.Mount{{
			Type: mount.TypeBind, Source: "/var/lib/autobase/ansible", Target: "/tmp/ansible",
		}}},
		nil)
	status, raw = w.do(t, http.MethodPost, "/v1.47/containers/create", autobase)
	stop(t, assert.Equal(t, http.StatusCreated, status, string(raw)))
	_, body := w.posted(t, "/containers/create")
	host := object(body["HostConfig"])
	assert.Equal(t, "hp-dapi-app1", host["NetworkMode"])
	assert.Equal(t, "app1-key/ansible", object(object(list(host["Mounts"])[0])["VolumeOptions"])["Subpath"])
}
