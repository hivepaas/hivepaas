package dockerproxy

import (
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// FuzzCreate sends the proxy arbitrary creates. Whatever it lets through must
// be confined: nothing that reaches past the container, no path of the host, no
// network or volume of anyone else, and the owner label.
func FuzzCreate(f *testing.F) {
	for _, name := range []string{"cli-create-plain.json", "cli-create-bind-host.json"} {
		raw, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			f.Fatal(err)
		}
		f.Add(raw)
	}
	f.Add([]byte(`{"Image":"alpine","HostConfig":{"Privileged":true}}`))
	f.Add([]byte(`{"Image":"alpine","HostConfig":{"Binds":["cache-app1:/c","/var/lib/autobase/ansible:/a"]}}`))
	f.Add([]byte(`{"Image":"alpine","HostConfig":{"Mounts":[{"Type":"volume","Source":"cache-app2","Target":"/x"}]}}`))
	f.Add([]byte(`{"Image":"alpine","HostConfig":{"NetworkMode":"jobnet","CpuQuota":1,"CpuPeriod":1000}}`))

	policy := testPolicy()
	policy.Allow = []Group{GroupVolumes, GroupNetworks, GroupNestedSocket}
	w := newWorld(f, policy)

	f.Fuzz(func(t *testing.T, raw []byte) {
		status, _ := w.do(t, http.MethodPost, createPath, raw)
		if status != http.StatusCreated {
			return
		}
		_, body := w.posted(t, "/containers/create")
		assertConfined(t, policy, body)
	})
}

// reachingFields are the HostConfig fields that reach past the container.
var reachingFields = []string{"Privileged", "CapAdd", "Devices", "DeviceRequests", "DeviceCgroupRules",
	"PidMode", "IpcMode", "UTSMode", "UsernsMode", "CgroupnsMode", "Cgroup", "SecurityOpt", "Runtime",
	"Sysctls", "CgroupParent", "OomKillDisable", "VolumesFrom", "Links", "PortBindings", "PublishAllPorts",
	"VolumeDriver"}

func assertConfined(t *testing.T, policy *Policy, body map[string]any) {
	t.Helper()
	host := object(body["HostConfig"])
	for _, field := range reachingFields {
		assert.True(t, isZero(host[field]), "HostConfig.%s reached the daemon: %v", field, host[field])
	}
	assert.Nil(t, host["MaskedPaths"])
	assert.Nil(t, host["ReadonlyPaths"])
	for _, bind := range list(host["Binds"]) {
		source := strings.Split(text(bind), ":")[0]
		assert.False(t, strings.HasPrefix(source, "/"), "a path of the host reached the daemon: %v", bind)
		assert.NotEqual(t, "cache-app2", source)
	}
	for _, raw := range list(host["Mounts"]) {
		m := object(raw)
		assert.Contains(t, []string{"volume", "tmpfs"}, text(m["Type"]), "mount %v", m)
		assert.NotEqual(t, "cache-app2", text(m["Source"]))
	}
	mode := text(host["NetworkMode"])
	usable := []string{"none", policy.Network, "jobnet", "n-job", "n-app", "proj_env_net", "n-env"}
	assert.True(t, slices.Contains(usable, mode), "network mode %q", mode)
	assert.Equal(t, policy.AppID, text(object(body["Labels"])[OwnerLabel]))
}
