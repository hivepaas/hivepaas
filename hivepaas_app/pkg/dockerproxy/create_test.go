package dockerproxy

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const createPath = "/v1.51/containers/create"

func TestCreateTakesWhatTheCLISendsForAPlainContainer(t *testing.T) {
	w := newWorld(t, testPolicy())
	status, raw := w.do(t, http.MethodPost, createPath+"?name=job1", fixture(t, "cli-create-plain.json"))
	stop(t, assert.Equal(t, http.StatusCreated, status, string(raw)))

	path, body := w.posted(t, "/containers/create")
	assert.Equal(t, createPath, path)
	host := object(body["HostConfig"])
	assert.Equal(t, "hp-dapi-app1", host["NetworkMode"])
	endpoints := object(object(body["NetworkingConfig"])["EndpointsConfig"])
	assert.Contains(t, endpoints, "hp-dapi-app1")
	assert.NotContains(t, endpoints, "default")
	assert.Equal(t, json.Number("1073741824"), host["Memory"])
	assert.Equal(t, json.Number("1000000000"), host["NanoCpus"])
	assert.Equal(t, json.Number("1024"), host["PidsLimit"])
	assert.Equal(t, map[string]any{OwnerLabel: "app1"}, body["Labels"])
}

func TestCreateRefusesEveryFieldThatReachesPastTheContainer(t *testing.T) {
	w := newWorld(t, testPolicy())
	fields := map[string]any{
		"HostConfig.Privileged": true,
		"HostConfig.CapAdd":     []any{"SYS_ADMIN"},
		"HostConfig.Devices": []any{map[string]any{
			"PathOnHost": "/dev/kmsg", "PathInContainer": "/dev/kmsg", "CgroupPermissions": "rwm",
		}},
		"HostConfig.DeviceRequests":    []any{map[string]any{"Count": -1, "Capabilities": []any{[]any{"gpu"}}}},
		"HostConfig.DeviceCgroupRules": []any{"c 1:3 rwm"},
		"HostConfig.PidMode":           "host",
		"HostConfig.IpcMode":           "host",
		"HostConfig.UTSMode":           "host",
		"HostConfig.UsernsMode":        "host",
		"HostConfig.CgroupnsMode":      "host",
		"HostConfig.Cgroup":            "container:other1",
		"HostConfig.SecurityOpt":       []any{"seccomp=unconfined"},
		"HostConfig.Runtime":           "runc",
		"HostConfig.Sysctls":           map[string]any{"net.ipv4.ip_forward": "1"},
		"HostConfig.CgroupParent":      "/",
		"HostConfig.OomKillDisable":    true,
		"HostConfig.OomScoreAdj":       -1000,
		"HostConfig.VolumesFrom":       []any{"other1"},
		"HostConfig.Links":             []any{"other1:db"},
		"HostConfig.PortBindings":      map[string]any{"22/tcp": []any{map[string]any{"HostPort": "2222"}}},
		"HostConfig.PublishAllPorts":   true,
		"HostConfig.VolumeDriver":      "local",
		"HostConfig.MemorySwappiness":  60,
		"HostConfig.MaskedPaths":       []any{},
		"HostConfig.ReadonlyPaths":     []any{},
	}
	for field, value := range fields {
		status, raw := w.do(t, http.MethodPost, createPath, set(fixture(t, "cli-create-plain.json"), field, value))
		assert.Equal(t, http.StatusForbidden, status, field)
		assert.Contains(t, refusalMessage(t, raw), field, field)
	}
	assert.False(t, w.reached(http.MethodPost, "/containers/create"))
}

func TestCreateTakesOnlyThePolicysImages(t *testing.T) {
	w := newWorld(t, testPolicy())
	status, raw := w.do(t, http.MethodPost, createPath, set(fixture(t, "cli-create-plain.json"), "Image", "busybox"))
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: image busybox is not allowed", refusalMessage(t, raw))
}

func TestCreatePutsTheDefaultNetworksOnTheAppsNetwork(t *testing.T) {
	w := newWorld(t, testPolicy())
	for _, mode := range []string{"", "default", "bridge", "host"} {
		status, raw := w.do(t, http.MethodPost, createPath,
			set(fixture(t, "cli-create-plain.json"), "HostConfig.NetworkMode", mode))
		stop(t, assert.Equal(t, http.StatusCreated, status, "%q %s", mode, raw))
		_, body := w.posted(t, "/containers/create")
		assert.Equal(t, "hp-dapi-app1", object(body["HostConfig"])["NetworkMode"], mode)
	}
}

func TestCreateJoinsOnlyNetworksTheAppMayUse(t *testing.T) {
	w := newWorld(t, testPolicy())
	for mode, want := range map[string]int{
		"jobnet":           http.StatusCreated,
		"n-job":            http.StatusCreated,
		"proj_env_net":     http.StatusCreated,
		"none":             http.StatusCreated,
		"othernet":         http.StatusForbidden,
		"hivepaas_net":     http.StatusForbidden,
		"container:other1": http.StatusForbidden,
	} {
		body := set(fixture(t, "cli-create-plain.json"), "HostConfig.NetworkMode", mode)
		body["NetworkingConfig"] = map[string]any{}
		status, raw := w.do(t, http.MethodPost, createPath, body)
		assert.Equal(t, want, status, "%s %s", mode, raw)
	}

	body := set(fixture(t, "cli-create-plain.json"), "HostConfig.NetworkMode", "jobnet")
	body["NetworkingConfig"] = map[string]any{"EndpointsConfig": map[string]any{"othernet": map[string]any{}}}
	status, raw := w.do(t, http.MethodPost, createPath, body)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: network othernet is not one this app may use", refusalMessage(t, raw))

	body["NetworkingConfig"] = map[string]any{"EndpointsConfig": map[string]any{
		"jobnet": map[string]any{"IPAMConfig": map[string]any{"IPv4Address": "10.0.0.9"}},
	}}
	status, raw = w.do(t, http.MethodPost, createPath, body)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: EndpointsConfig.jobnet.IPAMConfig is not allowed", refusalMessage(t, raw))
}

func TestCreateGivesTheLimitsAndRefusesMore(t *testing.T) {
	w := newWorld(t, testPolicy())
	refused := map[string]map[string]any{
		"memory":            {"HostConfig.Memory": 2 << 30},
		"cpus":              {"HostConfig.NanoCpus": 2_000_000_000},
		"cpu quota":         {"HostConfig.CpuQuota": 200000, "HostConfig.CpuPeriod": 100000},
		"cpu period":        {"HostConfig.CpuQuota": 1000, "HostConfig.CpuPeriod": 10},
		"pids limit":        {"HostConfig.PidsLimit": 99999},
		"unlimited swap":    {"HostConfig.MemorySwap": -1},
		"log driver syslog": {"HostConfig.LogConfig": map[string]any{"Type": "syslog"}},
	}
	for what, fields := range refused {
		body := fixture(t, "cli-create-plain.json")
		for field, value := range fields {
			set(body, field, value)
		}
		status, raw := w.do(t, http.MethodPost, createPath, body)
		assert.Equal(t, http.StatusForbidden, status, what)
		assert.Contains(t, refusalMessage(t, raw), what)
	}

	body := set(fixture(t, "cli-create-plain.json"), "HostConfig.CpuQuota", 50000)
	status, raw := w.do(t, http.MethodPost, createPath, body)
	stop(t, assert.Equal(t, http.StatusCreated, status, string(raw)))
	_, forwarded := w.posted(t, "/containers/create")
	assert.Equal(t, json.Number("0"), object(forwarded["HostConfig"])["NanoCpus"],
		"a quota is the child's own limit, and docker refuses both at once")
}

func TestCreateReplacesTheOwnerLabel(t *testing.T) {
	w := newWorld(t, testPolicy())
	body := set(fixture(t, "cli-create-plain.json"), "Labels", map[string]any{
		"keep": "yes", OwnerLabel: "app2", "com.docker.swarm.service.id": "svc1",
	})
	status, _ := w.do(t, http.MethodPost, createPath, body)
	stop(t, assert.Equal(t, http.StatusCreated, status))
	_, forwarded := w.posted(t, "/containers/create")
	assert.Equal(t, map[string]any{"keep": "yes", OwnerLabel: "app1"}, forwarded["Labels"])
}

func TestCreateStopsAtTheContainerLimit(t *testing.T) {
	policy := testPolicy()
	policy.Limits.Containers = 1
	w := newWorld(t, policy)
	status, raw := w.do(t, http.MethodPost, createPath, fixture(t, "cli-create-plain.json"))
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: this app already has 1 containers, its limit", refusalMessage(t, raw))
}

func TestCreateSpeaksAtLeastTheVersionSubpathsNeed(t *testing.T) {
	w := newWorld(t, testPolicy())
	for sent, want := range map[string]string{
		"/v1.44/containers/create": "/v1.45/containers/create",
		"/v1.51/containers/create": "/v1.51/containers/create",
		"/containers/create":       "/containers/create",
	} {
		status, _ := w.do(t, http.MethodPost, sent, fixture(t, "cli-create-plain.json"))
		stop(t, assert.Equal(t, http.StatusCreated, status, sent))
		path, _ := w.posted(t, "/containers/create")
		assert.Equal(t, want, path)
	}
}
