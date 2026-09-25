package dockerproxy

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// HivePaaS names an app's socket volume and its network itself. A child that
// could create one of those names would be handing itself, or another app,
// something the name is supposed to guarantee: an app's socket is its own, and
// its network holds only its own children.
func TestChildrenMayNotCreateAReservedName(t *testing.T) {
	policy := volumesAndNetworks()
	w := newWorld(t, policy)

	for _, name := range []string{"hp-dapi-sock-app2", "hp-dapi-app2", "hp-dapi-sock-app1"} {
		status, raw := w.do(t, http.MethodPost, "/v1.51/volumes/create", map[string]any{"Name": name})
		assert.Equal(t, http.StatusForbidden, status, name)
		assert.Contains(t, refusalMessage(t, raw), "is a name HivePaaS uses", name)

		status, raw = w.do(t, http.MethodPost, "/v1.51/networks/create", map[string]any{"Name": name})
		assert.Equal(t, http.StatusForbidden, status, name)
		assert.Contains(t, refusalMessage(t, raw), "is a name HivePaaS uses", name)
	}
}

// The same name in a mount, which creates the volume when the node has none.
func TestAChildMayNotMountAReservedVolumeName(t *testing.T) {
	policy := volumesAndNetworks()
	w := newWorld(t, policy)

	status, raw := w.do(t, http.MethodPost, "/v1.51/containers/create", map[string]any{
		"Image": "alpine", "HostConfig": map[string]any{"Binds": []any{"hp-dapi-sock-app2:/sock"}},
	})

	assert.Equal(t, http.StatusForbidden, status)
	assert.Contains(t, refusalMessage(t, raw), "is a name HivePaaS uses")
}

// Its own socket is still its own: nested socket is how it is given, and the
// name is the app's by construction.
func TestAChildMayStillMountItsOwnSocketVolume(t *testing.T) {
	policy := volumesAndNetworks()
	policy.Allow = append(policy.Allow, GroupNestedSocket)
	w := newWorld(t, policy)

	status, _ := w.do(t, http.MethodPost, "/v1.51/containers/create", map[string]any{
		"Image": "alpine", "HostConfig": map[string]any{"Binds": []any{"hp-dapi-sock-app1:/var/run/docker.sock"}},
	})

	assert.Equal(t, http.StatusCreated, status)
}
