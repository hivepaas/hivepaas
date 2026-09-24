package dockerproxy

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func volumesAndNetworks() *Policy {
	policy := testPolicy()
	policy.Allow = []Group{GroupVolumes, GroupNetworks}
	return policy
}

func TestVolumesAreTheAppsOwn(t *testing.T) {
	w := newWorld(t, volumesAndNetworks())

	status, raw := w.do(t, http.MethodGet, "/v1.51/volumes", nil)
	stop(t, assert.Equal(t, http.StatusOK, status))
	var answer struct {
		Volumes []struct{ Name string }
	}
	stop(t, assert.NoError(t, json.Unmarshal(raw, &answer)))
	stop(t, assert.Len(t, answer.Volumes, 1))
	assert.Equal(t, "cache-app1", answer.Volumes[0].Name)

	status, _ = w.do(t, http.MethodPost, "/v1.51/volumes/create", map[string]any{
		"Name": "cache2", "Labels": map[string]any{"keep": "yes", OwnerLabel: "app2"},
	})
	assert.Equal(t, http.StatusCreated, status)
	_, body := w.forwarded(t, http.MethodPost, "/volumes/create")
	assert.Equal(t, map[string]any{"keep": "yes", OwnerLabel: "app1"}, body["Labels"])

	status, raw = w.do(t, http.MethodPost, "/v1.51/volumes/create", fixture(t, "cli-volume-create-bind-opts.json"))
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: Volume.DriverOpts is not allowed", refusalMessage(t, raw))

	status, raw = w.do(t, http.MethodPost, "/v1.51/volumes/create", map[string]any{"Name": "x", "Driver": "nfs"})
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: volume driver nfs is not allowed", refusalMessage(t, raw))

	status, _ = w.do(t, http.MethodGet, "/v1.51/volumes/cache-app1", nil)
	assert.Equal(t, http.StatusOK, status)
	status, raw = w.do(t, http.MethodDelete, "/v1.51/volumes/cache-app2", nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: volume cache-app2 is not one this app created", refusalMessage(t, raw))
	status, _ = w.do(t, http.MethodDelete, "/v1.51/volumes/hp-dapi-sock-app1", nil)
	assert.Equal(t, http.StatusForbidden, status)
	status, _ = w.do(t, http.MethodDelete, "/v1.51/volumes/cache-app1", nil)
	assert.Equal(t, http.StatusOK, status)
}

func TestNetworksAreTheAppsOwn(t *testing.T) {
	w := newWorld(t, volumesAndNetworks())

	status, raw := w.do(t, http.MethodPost, "/v1.51/networks/create", fixture(t, "cli-network-create.json"))
	stop(t, assert.Equal(t, http.StatusOK, status, string(raw)))
	_, body := w.forwarded(t, http.MethodPost, "/networks/create")
	assert.Equal(t, map[string]any{OwnerLabel: "app1"}, body["Labels"])

	refused := map[string]map[string]any{
		"network driver overlay is not allowed": {"Name": "n", "Driver": "overlay"},
		"network scope swarm is not allowed":    {"Name": "n", "Scope": "swarm"},
		"Network.Options is not allowed": {
			"Name": "n", "Options": map[string]any{"com.docker.network.bridge.name": "docker0"},
		},
		"Network.Ingress is not allowed":    {"Name": "n", "Ingress": true},
		"IPAM driver custom is not allowed": {"Name": "n", "IPAM": map[string]any{"Driver": "custom"}},
		"Network.IPAM.Config is not allowed": {
			"Name": "n", "IPAM": map[string]any{"Config": []any{map[string]any{"Subnet": "10.0.0.0/8"}}},
		},
	}
	for message, request := range refused {
		status, raw = w.do(t, http.MethodPost, "/v1.51/networks/create", request)
		assert.Equal(t, http.StatusForbidden, status, message)
		assert.Equal(t, "hivepaas: "+message, refusalMessage(t, raw))
	}

	status, raw = w.do(t, http.MethodGet, "/v1.51/networks", nil)
	stop(t, assert.Equal(t, http.StatusOK, status))
	var networks []struct{ Name string }
	stop(t, assert.NoError(t, json.Unmarshal(raw, &networks)))
	names := make([]string, 0, len(networks))
	for _, n := range networks {
		names = append(names, n.Name)
	}
	assert.ElementsMatch(t, []string{"hp-dapi-app1", "jobnet", "proj_env_net"}, names)

	for path, want := range map[string]int{
		"/v1.51/networks/n-job":        http.StatusOK,
		"/v1.51/networks/hp-dapi-app1": http.StatusOK,
		"/v1.51/networks/othernet":     http.StatusForbidden,
		"/v1.51/networks/hivepaas_net": http.StatusForbidden,
	} {
		status, _ = w.do(t, http.MethodGet, path, nil)
		assert.Equal(t, want, status, path)
	}

	// The app may remove what it created, not the network HivePaaS gave it.
	status, _ = w.do(t, http.MethodDelete, "/v1.51/networks/jobnet", nil)
	assert.Equal(t, http.StatusOK, status)
	status, raw = w.do(t, http.MethodDelete, "/v1.51/networks/hp-dapi-app1", nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: network hp-dapi-app1 is not one this app created", refusalMessage(t, raw))
}

func TestConnectTakesTheAppsChildrenAndTheAppItself(t *testing.T) {
	w := newWorld(t, volumesAndNetworks())
	tests := []struct {
		network string
		body    map[string]any
		want    int
	}{
		{"jobnet", map[string]any{"Container": "child1"}, http.StatusOK},
		{"jobnet", map[string]any{"Container": "child1", "EndpointConfig": map[string]any{"Aliases": []string{"db"}}},
			http.StatusOK},
		// Appwrite's executor connects itself to the network of its runtimes.
		{"jobnet", map[string]any{"Container": "task1"}, http.StatusOK},
		{"proj_env_net", map[string]any{"Container": "child1"}, http.StatusOK},
		{"jobnet", map[string]any{"Container": "other1"}, http.StatusForbidden},
		{"othernet", map[string]any{"Container": "child1"}, http.StatusForbidden},
		{"jobnet", map[string]any{"Container": "child1",
			"EndpointConfig": map[string]any{"IPAMConfig": map[string]any{"IPv4Address": "10.0.0.9"}}},
			http.StatusForbidden},
	}
	for _, tt := range tests {
		status, raw := w.do(t, http.MethodPost, "/v1.51/networks/"+tt.network+"/connect", tt.body)
		assert.Equal(t, tt.want, status, "%s %v %s", tt.network, tt.body, raw)
	}
	status, _ := w.do(t, http.MethodPost, "/v1.51/networks/jobnet/disconnect",
		map[string]any{"Container": "child1", "Force": true})
	assert.Equal(t, http.StatusOK, status)
}

func TestReservedLabelsCannotBeSet(t *testing.T) {
	got := ownLabels(map[string]any{
		"keep": "yes", OwnerLabel: "app2", "com.docker.swarm.service.id": "svc1", "hivepaas.app.info": "{}",
	}, "app1")
	assert.Equal(t, map[string]any{"keep": "yes", OwnerLabel: "app1"}, got)
}
