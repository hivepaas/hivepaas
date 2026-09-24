package dockerproxy

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// createWith sends the plain CLI create with the given binds and mounts.
func createWith(t *testing.T, w *world, binds, mounts []any) (int, string) {
	t.Helper()
	body := fixture(t, "cli-create-plain.json")
	set(body, "HostConfig.Binds", binds)
	set(body, "HostConfig.Mounts", mounts)
	status, raw := w.do(t, http.MethodPost, createPath, body)
	if status == http.StatusCreated {
		return status, ""
	}
	return status, refusalMessage(t, raw)
}

func forwardedStorage(t *testing.T, w *world) ([]any, []any) {
	t.Helper()
	_, body := w.posted(t, "/containers/create")
	host := object(body["HostConfig"])
	return list(host["Binds"]), list(host["Mounts"])
}

func TestCreateMountsASharedDirectoryFromTheAppsVolume(t *testing.T) {
	w := newWorld(t, testPolicy())
	status, raw := w.do(t, http.MethodPost, createPath, fixture(t, "cli-create-bind-host.json"))
	stop(t, assert.Equal(t, http.StatusCreated, status, string(raw)))

	binds, mounts := forwardedStorage(t, w)
	assert.Empty(t, binds)
	assert.Equal(t, []any{map[string]any{
		"Type": "volume", "Source": "hp-vol-data", "Target": "/tmp/ansible", "ReadOnly": false,
		"VolumeOptions": map[string]any{"NoCopy": true, "Subpath": "app1-key/ansible"},
	}}, mounts)
}

func TestCreateReachesNoOtherPathOfTheHost(t *testing.T) {
	w := newWorld(t, testPolicy())
	for _, bind := range []string{
		"/:/host",
		"/var/run/docker.sock:/var/run/docker.sock",
		// The app's own mount, but not a directory it shares.
		"/var/lib/autobase:/data",
		"/var/lib/autobase/ansiblex:/data",
		"/var/lib/autobase/ansible/../../..:/data",
	} {
		status, message := createWith(t, w, []any{bind}, nil)
		assert.Equal(t, http.StatusForbidden, status, bind)
		assert.Contains(t, message, "is not a shared directory of this app", bind)
	}
	status, message := createWith(t, w, nil, []any{map[string]any{"Type": "bind", "Source": "/etc", "Target": "/x"}})
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: /etc is not a shared directory of this app", message)
	assert.False(t, w.reached(http.MethodPost, "/containers/create"))
}

func TestCreateRefusesASharedDirectoryTheAppKeepsNoVolumeFor(t *testing.T) {
	policy := testPolicy()
	policy.SharedDirs = []string{"/srv/work"}
	w := newWorld(t, policy)
	status, message := createWith(t, w, []any{"/srv/work:/work"}, nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: shared directory /srv/work is not on a volume of the app", message)
}

func TestCreateFindsSharedDirectoriesOnlyInTheAppsOwnTask(t *testing.T) {
	w := newWorld(t, testPolicy())
	w.daemon.mu.Lock()
	delete(w.daemon.containers, "task1")
	// A child cannot set swarm labels through the proxy, but one made some other
	// way still must not stand in for the app.
	w.daemon.containers["spoof"] = &fakeContainer{running: true,
		labels: map[string]string{serviceIDLabel: "svc1", OwnerLabel: "app1"},
		mounts: []map[string]any{{"Type": "volume", "Source": "cache-app2", "Target": "/var/lib/autobase"}}}
	w.daemon.mu.Unlock()

	status, message := createWith(t, w, []any{"/var/lib/autobase/ansible:/tmp/ansible"}, nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: the app is not running on this node", message)
}

func TestCreateMountsTheAppsSocketOnlyWhenNestingIsAllowed(t *testing.T) {
	w := newWorld(t, testPolicy())
	status, message := createWith(t, w, []any{SocketPath + ":/var/run/docker.sock"}, nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: mounting the app's socket is not allowed for this app", message)
	status, _ = createWith(t, w, []any{"hp-dapi-sock-app1:/run/hp"}, nil)
	assert.Equal(t, http.StatusForbidden, status)

	policy := testPolicy()
	policy.Allow = []Group{GroupNestedSocket}
	w = newWorld(t, policy)
	status, message = createWith(t, w, []any{SocketPath + ":/var/run/docker.sock"}, nil)
	stop(t, assert.Equal(t, http.StatusCreated, status, message))
	_, mounts := forwardedStorage(t, w)
	assert.Equal(t, []any{map[string]any{
		"Type": "volume", "Source": "hp-dapi-sock-app1", "Target": "/var/run/docker.sock", "ReadOnly": false,
		"VolumeOptions": map[string]any{"NoCopy": true, "Subpath": "docker.sock"},
	}}, mounts)
}

func TestCreateTakesOnlyTheAppsNamedVolumes(t *testing.T) {
	w := newWorld(t, testPolicy())
	status, message := createWith(t, w, []any{"cache-app1:/cache"}, nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: volumes is not allowed for this app", message)

	policy := testPolicy()
	policy.Allow = []Group{GroupVolumes}
	w = newWorld(t, policy)

	status, message = createWith(t, w, []any{"cache-app1:/cache:ro"}, nil)
	stop(t, assert.Equal(t, http.StatusCreated, status, message))
	binds, _ := forwardedStorage(t, w)
	assert.Equal(t, []any{"cache-app1:/cache:ro"}, binds)

	status, message = createWith(t, w, []any{"cache-app2:/cache"}, nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: volume cache-app2 is not one this app created", message)

	// act names volumes in Binds and never creates them.
	status, message = createWith(t, w, []any{"act-toolcache:/opt/hostedtoolcache"}, nil)
	stop(t, assert.Equal(t, http.StatusCreated, status, message))
	w.daemon.mu.Lock()
	assert.Equal(t, map[string]string{OwnerLabel: "app1"}, w.daemon.volumes["act-toolcache"])
	w.daemon.mu.Unlock()

	status, _ = createWith(t, w, nil, []any{map[string]any{"Type": "volume", "Source": "cache-app2", "Target": "/c"}})
	assert.Equal(t, http.StatusForbidden, status)
	status, message = createWith(t, w, nil, []any{map[string]any{"Type": "volume", "Source": "cache-app1",
		"Target": "/c", "VolumeOptions": map[string]any{"DriverConfig": map[string]any{"Name": "local",
			"Options": map[string]any{"device": "/", "o": "bind", "type": "none"}}}}})
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: Mount.VolumeOptions.DriverConfig is not allowed", message)
	status, message = createWith(t, w, nil, []any{map[string]any{"Type": "volume", "Source": "cache-app1",
		"Target": "/c", "VolumeOptions": map[string]any{"Subpath": "../../etc"}}})
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: volume subpath ../../etc leaves the volume", message)
}

func TestCreateRefusesBindModesThatReachTheHost(t *testing.T) {
	policy := testPolicy()
	policy.Allow = []Group{GroupVolumes}
	w := newWorld(t, policy)
	for _, bind := range []string{"cache-app1:/c:z", "cache-app1:/c:Z", "cache-app1:/c:rshared", "a:b:c:d", "/x"} {
		status, _ := createWith(t, w, []any{bind}, nil)
		assert.Equal(t, http.StatusForbidden, status, bind)
	}
}

func TestCreateTakesTmpfsAndRefusesOtherMountTypes(t *testing.T) {
	w := newWorld(t, testPolicy())
	status, message := createWith(t, w, nil, []any{map[string]any{"Type": "tmpfs", "Target": "/scratch",
		"TmpfsOptions": map[string]any{"SizeBytes": 1 << 20}}})
	stop(t, assert.Equal(t, http.StatusCreated, status, message))

	for _, kind := range []string{"npipe", "cluster", "image"} {
		status, message = createWith(t, w, nil, []any{map[string]any{"Type": kind, "Source": "x", "Target": "/x"}})
		assert.Equal(t, http.StatusForbidden, status, kind)
		assert.Equal(t, "hivepaas: mount type "+kind+" is not allowed", message)
	}
	status, message = createWith(t, w, nil, []any{map[string]any{"Type": "bind",
		"Source": "/var/lib/autobase/ansible", "Target": "/x", "BindOptions": map[string]any{"Propagation": "rshared"}}})
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: Mount.BindOptions is not allowed", message)
}
