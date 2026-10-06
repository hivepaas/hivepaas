package obi

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// Every task's container of the apps, each once, in one order.
func TestPatternsAndConfig(t *testing.T) {
	patterns := Patterns([]string{"p1_dev_a2", "", "p1_dev_a1", "p1_dev_a2"})
	assert.Equal(t, []string{"p1_dev_a1.*", "p1_dev_a2.*"}, patterns)
	assert.Equal(t, `discovery:
  instrument:
    - container_name: "p1_dev_a1.*"
    - container_name: "p1_dev_a2.*"
routes:
  unmatched: heuristic
  ignored_patterns:
    - "/_hivepaas/*"
ebpf:
  wakeup_len: 64
  maps_config:
    global_scale_factor: -2
`, string(Config(patterns, CapacitySmall)))
	assert.Contains(t, string(Config(patterns, CapacityMedium)), "wakeup_len: 128\n")
	assert.Contains(t, string(Config(patterns, CapacityMedium)), "global_scale_factor: -1\n")
	assert.Contains(t, string(Config(patterns, CapacityLarge)), "wakeup_len: 256\n")
	assert.Contains(t, string(Config(patterns, CapacityLarge)), "global_scale_factor: 0\n")

	small := ConfigHash(DefaultImage, false, Config(patterns, CapacitySmall))
	assert.NotEqual(t, small, ConfigHash(DefaultImage, false, Config(patterns[:1], CapacitySmall)))
	assert.NotEqual(t, small, ConfigHash(DefaultImage, false, Config(patterns, CapacityMedium)),
		"another capacity: another OBI")
	assert.NotEqual(t, small, ConfigHash("otel/ebpf-instrument:v0.15.0", false, Config(patterns, CapacitySmall)),
		"another release's OBI: another OBI")
	assert.Equal(t, small, ConfigHash(DefaultImage, false, Config(patterns, CapacitySmall)))
}

// HivePaaS recommends more capacity to a node with more memory; a capacity
// chosen is kept, auto follows the recommendation.
func TestCapacity(t *testing.T) {
	// What nodes sold with 1, 4, 8, 16, 32 and 128 GB read.
	for memMB, want := range map[int]Capacity{0: CapacitySmall, 961: CapacitySmall, 3912: CapacitySmall,
		7679: CapacitySmall, 7820: CapacityMedium, 15990: CapacityMedium, 30719: CapacityMedium,
		32090: CapacityLarge, 128700: CapacityLarge} {
		assert.Equal(t, want, Recommended(memMB), memMB)
	}
	assert.Equal(t, CapacityMedium, CapacityAuto.Effective(16384))
	assert.Equal(t, CapacityLarge, CapacityLarge.Effective(961), "chosen, whatever the memory")

	for in, want := range map[string]Capacity{"": CapacityAuto, "auto": CapacityAuto, "small": CapacitySmall,
		"medium": CapacityMedium, "large": CapacityLarge} {
		got, ok := ParseCapacity(in)
		assert.True(t, ok, in)
		assert.Equal(t, want, got, in)
	}
	_, ok := ParseCapacity("huge")
	assert.False(t, ok)

	assert.Equal(t, -2, CapacitySmall.ScaleFactor())
	assert.Equal(t, 215, CapacityLarge.MemoryMiB())
	assert.Equal(t, 15000, CapacityMedium.Tracked())
}

// OBI is not privileged: the host's PID namespace, the agent's network one,
// the measured capabilities, the Docker socket read-only, its memory bounded.
func TestContainerOptions(t *testing.T) {
	opts := ContainerOptions(DefaultImage, "agent123", "abc", false)
	host := opts.HostConfig
	assert.Equal(t, ContainerName, opts.Name)
	assert.Equal(t, DefaultImage, opts.Config.Image)
	assert.False(t, host.Privileged)
	assert.Equal(t, "host", string(host.PidMode))
	assert.Equal(t, "container:agent123", string(host.NetworkMode))
	assert.ElementsMatch(t, []string{"CAP_BPF", "CAP_PERFMON", "CAP_SYS_PTRACE", "CAP_NET_RAW",
		"CAP_DAC_READ_SEARCH", "CAP_CHECKPOINT_RESTORE"}, host.CapAdd)
	assert.Contains(t, host.Binds, "/var/run/docker.sock:/var/run/docker.sock:ro")
	assert.Equal(t, int64(512<<20), host.Memory)

	// A container made with other options - another limit, by an agent from
	// before - is another OBI, and is replaced.
	config := Config([]string{"p1_dev_a1.*"}, CapacitySmall)
	other := ContainerOptions(DefaultImage, "", "", false)
	other.HostConfig.Memory = 384 << 20
	assert.NotEqual(t, ConfigHash(DefaultImage, false, config), hashOf(other, config))
	assert.Equal(t, ConfigHash(DefaultImage, false, config),
		hashOf(ContainerOptions(DefaultImage, "", "", false), config))
	assert.Equal(t, base.LogComponentOBI, opts.Config.Labels[base.LabelLogComponent])
	assert.Equal(t, "abc", opts.Config.Labels[LabelConfig])
	assert.Contains(t, opts.Config.Env, "OTEL_EBPF_CONFIG_PATH=/hivepaas-obi.yaml")
}

// Where the kernel allows perf events to CAP_SYS_ADMIN alone, OBI has it too,
// and is still not privileged; a node where that changes gets another OBI.
func TestContainerOptionsWherePerfEventsAreRestricted(t *testing.T) {
	host := ContainerOptions(DefaultImage, "agent123", "abc", true).HostConfig
	assert.ElementsMatch(t, []string{"CAP_BPF", "CAP_PERFMON", "CAP_SYS_PTRACE", "CAP_NET_RAW",
		"CAP_DAC_READ_SEARCH", "CAP_CHECKPOINT_RESTORE", "CAP_SYS_ADMIN"}, host.CapAdd)
	assert.False(t, host.Privileged)

	config := Config([]string{"p1_dev_a1.*"}, CapacitySmall)
	assert.NotEqual(t, ConfigHash(DefaultImage, false, config), ConfigHash(DefaultImage, true, config))
}

// node writes a node's filesystem: a kernel, BTF, tracefs, memory.
func node(t *testing.T, kernel string, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	all := map[string]string{
		"proc/sys/kernel/osrelease":    kernel + "\n",
		"sys/kernel/btf/vmlinux":       "btf",
		"sys/kernel/tracing/trace":     "",
		"proc/meminfo":                 "MemTotal: 1000000 kB\nMemAvailable:  466944 kB\n",
		"sys/kernel/security/lockdown": "[none] integrity confidentiality\n",
	}
	for name, content := range files {
		all[name] = content
	}
	for name, content := range all {
		if content == "-" {
			continue
		}
		path := filepath.Join(root, name)
		assert.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		assert.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	}
	return root
}

func TestPreflight(t *testing.T) {
	good := Check(node(t, "6.8.0-124-generic", nil), false, CapacityAuto)
	assert.True(t, good.OK, good.Reasons)
	assert.Equal(t, "6.8.0-124-generic", good.Kernel)
	assert.Equal(t, 976, good.MemTotalMB)
	assert.Equal(t, 456, good.MemAvailableMB)
	assert.Equal(t, CapacitySmall, good.Recommended)
	assert.Equal(t, CapacitySmall, good.Capacity, "auto: the recommended one")

	// 400 MB free: enough for small, not for large - twice its 215 MiB.
	tight := map[string]string{"proc/meminfo": "MemTotal: 1000000 kB\nMemAvailable: 409600 kB\n"}
	assert.True(t, Check(node(t, "6.8.0", tight), false, CapacitySmall).OK)
	large := Check(node(t, "6.8.0", tight), false, CapacityLarge)
	assert.Equal(t, CapacityLarge, large.Capacity, "chosen")
	assert.Equal(t, CapacitySmall, large.Recommended)
	assert.Contains(t, large.Reasons, ReasonLowMemory)

	cases := map[string]struct {
		kernel string
		files  map[string]string
		want   string
	}{
		"old kernel": {"4.19.0-26-amd64", nil, ReasonKernelTooOld},
		"no btf":     {"6.1.0", map[string]string{"sys/kernel/btf/vmlinux": "-"}, ReasonNoBTF},
		"openvz":     {"5.10.0", map[string]string{"proc/vz/veinfo": ""}, ReasonContainerVirt},
		"lxc": {"5.15.0", map[string]string{"proc/1/environ": "PATH=/bin\x00container=lxc\x00"},
			ReasonContainerVirt},
		"no tracefs": {"6.8.0", map[string]string{"sys/kernel/tracing/trace": "-"}, ReasonNoTracefs},
		"locked down": {"6.8.0", map[string]string{"sys/kernel/security/lockdown": "none [confidentiality]"},
			ReasonLockdown},
		"short of mem": {"6.8.0", map[string]string{"proc/meminfo": "MemAvailable: 102400 kB\n"}, ReasonLowMemory},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			p := Check(node(t, c.kernel, c.files), false, CapacityAuto)
			assert.False(t, p.OK)
			assert.Contains(t, p.Reasons, c.want)
		})
	}

	// An OpenVZ host has /proc/bc too: it is no container.
	host := Check(node(t, "5.10.0", map[string]string{"proc/vz/veinfo": "", "proc/bc/0": ""}), false, CapacityAuto)
	assert.True(t, host.OK, host.Reasons)
	// Once OBI runs, its own memory is in the free memory: not a reason.
	running := Check(node(t, "6.8.0", map[string]string{"proc/meminfo": "MemAvailable: 102400 kB\n"}), true,
		CapacityAuto)
	assert.True(t, running.OK, running.Reasons)
}

// Above 2, perf_event_paranoid restricts perf events: to CAP_SYS_ADMIN alone on
// Debian's kernels. Not a reason: OBI is given it.
func TestPreflightSaysWherePerfEventsAreRestricted(t *testing.T) {
	for paranoid, want := range map[string]bool{"3\n": true, "4\n": true, "2\n": false, "-1\n": false,
		"-": false, "junk": false} {
		p := Check(node(t, "6.12.111+deb13-cloud-amd64",
			map[string]string{"proc/sys/kernel/perf_event_paranoid": paranoid}), false, CapacityAuto)
		assert.True(t, p.OK, p.Reasons)
		assert.Equal(t, want, p.PerfRestricted, paranoid)
	}
}

func TestParseText(t *testing.T) {
	in := `# HELP x a help
# TYPE x histogram
x_bucket{a="b\"c",d="e\\f\ng",le="+Inf"} 3
x_count{a="1"} 2.5e+00 1700000000000
other{a="1"} 1
broken{a="1" 2
`
	samples, err := parseText(strings.NewReader(in), func(name string) bool { return name != "other" })
	assert.NoError(t, err)
	if assert.Len(t, samples, 2) {
		assert.Equal(t, map[string]string{"a": `b"c`, "d": "e\\f\ng", "le": "+Inf"}, samples[0].Labels)
		assert.InDelta(t, 3, samples[0].Value, 0)
		assert.InDelta(t, 2.5, samples[1].Value, 0)
	}
}

// scrape is OBI v0.14.0's metrics as step 0 read them: a server, an HTTP
// client calling it, Postgres - n requests each, served in 1 ms or 7 ms.
func scrape(n, slow int) string {
	h := func(name, labels string, fast, slow int) string {
		var b strings.Builder
		total := fast + slow
		for _, le := range []struct {
			le string
			n  int
		}{{"0.005", fast}, {"0.01", total}, {"+Inf", total}} {
			b.WriteString(name + `_bucket{` + labels + `,le="` + le.le + `"} ` + itoa(le.n) + "\n")
		}
		b.WriteString(name + `_sum{` + labels + `} ` + ftoa(float64(fast)*0.001+float64(slow)*0.007) + "\n")
		b.WriteString(name + `_count{` + labels + `} ` + itoa(total) + "\n")
		return b.String()
	}
	return "# TYPE http_server_request_duration_seconds histogram\n" +
		h("http_server_request_duration_seconds", `container_name="p1_dev_a2.1.abc",error_type="",http_request_method="GET",`+
			`http_response_status_code="200",http_route="/u/*",instance="p1_dev_a2.1.abc",job="p1_dev_a2.1.abc",`+
			`server_address="p1_dev_a2.1.abc",server_port="18555",url_scheme="http"`, n, slow) +
		h("http_client_request_duration_seconds", `container_name="p1_dev_a1.1.def",error_type="",http_request_method="GET",`+
			`http_response_status_code="503",http_route="/c/*",server_address="a2",server_port="18555"`, n, 0) +
		h("db_client_operation_duration_seconds", `container_name="p1_dev_a2.1.abc",db_namespace="postgres",`+
			`db_operation_name="SELECT",db_response_status_code="",db_system_name="postgresql",error_type="",`+
			`server_address="outgoing",server_port=""`, n, 0) +
		h("http_server_request_duration_seconds", `container_name="stranger.1.xyz",error_type="",http_request_method="GET",`+
			`http_response_status_code="200",http_route="/",server_port="80"`, n, 0) +
		"obi_build_info{version=\"v0.14.0\"} 1\n"
}

func itoa(n int) string     { return strconv.Itoa(n) }
func ftoa(f float64) string { return strconv.FormatFloat(f, 'g', -1, 64) }

func appOf(c string) (string, bool) {
	switch {
	case strings.HasPrefix(c, "p1_dev_a2."):
		return "A2", true
	case strings.HasPrefix(c, "p1_dev_a1."):
		return "A1", true
	}
	return "", false
}

// The first scrape is a baseline; the next says what moved, per series, for
// the apps known; a stranger's container is dropped.
func TestDeltas(t *testing.T) {
	d := NewDeltas()
	rows, err := d.Read(strings.NewReader(scrape(100, 0)), appOf)
	assert.NoError(t, err)
	assert.Empty(t, rows, "a baseline")

	rows, err = d.Read(strings.NewReader(scrape(130, 10)), appOf)
	assert.NoError(t, err)
	if !assert.Len(t, rows, 3) {
		return
	}
	byPeer := map[string]*Row{}
	for _, r := range rows {
		byPeer[r.HP+"/"+r.Peer] = r
	}
	server := byPeer["routes/"]
	if assert.NotNil(t, server) {
		assert.Equal(t, "A2", server.App)
		assert.Equal(t, "/u/*", server.Route)
		assert.Equal(t, "2xx", server.Status)
		assert.Equal(t, int64(40), server.Count, "30 fast, 10 slow")
		assert.Zero(t, server.Errors)
		assert.InDelta(t, 30*1+10*7, server.SumMs, 1e-6)
		// OBI's 5 ms and 10 ms buckets, onto HivePaaS's bounds: every bound
		// from 10 ms on holds all 40.
		want := map[float64]int64{math.Inf(1): 40}
		for _, b := range Bounds {
			want[b] = 40
		}
		want[5] = 30
		assert.Equal(t, want, server.Buckets)
	}
	client := byPeer["calls/a2:18555"]
	if assert.NotNil(t, client) {
		assert.Equal(t, "A1", client.App)
		assert.Equal(t, KindHTTP, client.Kind)
		assert.Equal(t, int64(30), client.Count)
		assert.Equal(t, int64(30), client.Errors, "a 503")
	}
	db := byPeer["calls/postgresql/postgres"]
	if assert.NotNil(t, db) {
		assert.Equal(t, KindDB, db.Kind)
		assert.Equal(t, "SELECT", db.Operation)
	}

	// Nothing moved: nothing written.
	rows, err = d.Read(strings.NewReader(scrape(130, 10)), appOf)
	assert.NoError(t, err)
	assert.Empty(t, rows)

	// OBI restarted: its counts start again, and count from zero.
	rows, err = d.Read(strings.NewReader(scrape(5, 0)), appOf)
	assert.NoError(t, err)
	assert.Len(t, rows, 3)
	for _, r := range rows {
		assert.Equal(t, int64(5), r.Count)
	}
}

// A series first seen after the baseline is new since the last scrape: it
// counts whole.
func TestDeltasCountANewSeriesWhole(t *testing.T) {
	d := NewDeltas()
	_, _ = d.Read(strings.NewReader("obi_build_info 1\n"), appOf)
	rows, err := d.Read(strings.NewReader(scrape(7, 0)), appOf)
	assert.NoError(t, err)
	if assert.Len(t, rows, 3) {
		assert.Equal(t, int64(7), rows[0].Count)
	}
}

// A row is one flat line, its buckets as le fields, empty fields left out.
func TestRowJSON(t *testing.T) {
	var buf bytes.Buffer
	assert.NoError(t, WriteRows(&buf, []*Row{{HP: RowRoutes, App: "A2", Container: "c", Kind: KindHTTP,
		Method: "GET", Route: "/u/*", Status: "2xx", Count: 3, SumMs: 7.1234567,
		Buckets: map[float64]int64{5: 1, 7.5: 2, math.Inf(1): 3}}}))
	var got map[string]any
	assert.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.Equal(t, map[string]any{"hp": "routes", "app": "A2", "container": "c", "kind": "http", "method": "GET",
		"route": "/u/*", "status": "2xx", "count": 3.0, "errors": 0.0, "sumMs": 7.123, "le5": 1.0, "le7.5": 2.0,
		"leInf": 3.0}, got)
	assert.True(t, strings.HasSuffix(buf.String(), "}\n"))
}

// OBI's buckets map onto HivePaaS's bounds: at each, the count at the largest
// of OBI's bounds not above it - none below the first, every one at +Inf.
func TestNormalizedBuckets(t *testing.T) {
	got := normalized(map[string]float64{"0.002": 1, "0.007": 4, "0.3": 9, "+Inf": 10}, 10)
	assert.Equal(t, int64(1), got[5], "2 ms is under 5")
	assert.Equal(t, int64(4), got[10], "7 ms is under 10")
	assert.Equal(t, int64(4), got[250])
	assert.Equal(t, int64(9), got[500], "300 ms is under 500")
	assert.Equal(t, int64(9), got[10000])
	assert.Equal(t, int64(10), got[math.Inf(1)])
	assert.Len(t, got, len(Bounds)+1)
}

// Quantiles come from the summed buckets: within the one each falls in,
// linearly; past the last bound, that bound.
func TestQuantile(t *testing.T) {
	buckets := func(at map[float64]int64) map[string]int64 {
		out := map[string]int64{}
		var n int64
		for _, b := range append(slices.Clone(Bounds), math.Inf(1)) {
			n = max(n, at[b])
			out[BucketField(b)] = n
		}
		return out
	}
	b := buckets(map[float64]int64{10: 50, 25: 90, 50: 100})
	assert.InDelta(t, 10, *Quantile(b, 0.5), 1e-9)
	assert.InDelta(t, 37.5, *Quantile(b, 0.95), 1e-9)
	assert.InDelta(t, 47.5, *Quantile(b, 0.99), 1e-9)
	assert.InDelta(t, 2.5, *Quantile(buckets(map[float64]int64{5: 10}), 0.5), 1e-9)
	assert.InDelta(t, 10000, *Quantile(map[string]int64{"leInf": 10}, 0.5), 1e-9, "past the last bound")
	assert.Nil(t, Quantile(map[string]int64{}, 0.5))
	assert.Nil(t, Quantile(b, 1.5))

	assert.Equal(t, []string{"le5", "le10", "le25", "le50", "le100", "le250", "le500", "le1000", "le2500", "le5000",
		"le10000", "leInf"}, BucketFields())
	sum := map[string]int64{"le5": 1}
	AddBuckets(sum, map[string]int64{"le5": 2, "leInf": 3})
	assert.Equal(t, map[string]int64{"le5": 3, "leInf": 3}, sum)
}
