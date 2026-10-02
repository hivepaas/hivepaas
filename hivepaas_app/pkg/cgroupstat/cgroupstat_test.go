package cgroupstat

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	assert.NoError(t, os.MkdirAll(dir, 0o755))
	for name, content := range files {
		assert.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
	}
}

func cgroupFiles(usage, throttled, current, oom, rbytes uint64) map[string]string {
	return map[string]string{
		"cpu.stat":       "usage_usec " + itoa(usage) + "\nthrottled_usec " + itoa(throttled) + "\n",
		"cpu.max":        "200000 100000\n",
		"memory.current": itoa(current) + "\n",
		"memory.max":     "max\n",
		"memory.stat":    "anon 100\nfile 50\ninactive_file 1000\n",
		"memory.events":  "low 0\noom 0\noom_kill " + itoa(oom) + "\n",
		"io.stat":        "8:0 rbytes=" + itoa(rbytes) + " wbytes=10 rios=1\n259:0 rbytes=5 wbytes=0\n",
	}
}

func itoa(n uint64) string { return strconv.FormatUint(n, 10) }

// The systemd and the cgroupfs layouts are both found; a container with
// neither is no cgroup.
func TestDir(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, filepath.Join(root, "system.slice", "docker-aaa.scope"), cgroupFiles(1, 0, 1, 0, 0))
	writeFiles(t, filepath.Join(root, "docker", "bbb"), cgroupFiles(1, 0, 1, 0, 0))

	dir, err := Dir(root, "aaa", "")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(root, "system.slice", "docker-aaa.scope"), dir)
	dir, err = Dir(root, "bbb", "")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(root, "docker", "bbb"), dir)
	_, err = Dir(root, "ccc", "")
	assert.ErrorIs(t, err, ErrNoCgroup)
}

// Two samples give cores, the working set, OOM kills and rates per second;
// loopback's traffic is not the container's.
func TestReadAndBetween(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "docker", "aaa")
	netDev := filepath.Join(t.TempDir(), "net-dev")
	writeNet := func(rx, tx uint64) {
		assert.NoError(t, os.WriteFile(netDev, []byte("Inter-|   Receive\n face |bytes\n"+
			"    lo: 999 1 0 0 0 0 0 0 999 1 0 0 0 0 0 0\n"+
			"  eth0: "+itoa(rx)+" 1 0 0 0 0 0 0 "+itoa(tx)+" 1 0 0 0 0 0 0\n"), 0o600))
	}
	at := time.Date(2026, 10, 3, 10, 0, 0, 0, time.UTC)

	writeFiles(t, dir, cgroupFiles(1_000_000, 0, 5000, 1, 100))
	writeNet(1000, 2000)
	first, err := Read(dir, netDev, at)
	assert.NoError(t, err)
	assert.Equal(t, 2.0, first.CPULimit)
	assert.Zero(t, first.MemoryLimit, "max is none")
	assert.Equal(t, uint64(105), first.IORead)

	writeFiles(t, dir, cgroupFiles(4_000_000, 1_500_000, 6000, 2, 1600))
	writeNet(31000, 2000)
	second, err := Read(dir, netDev, at.Add(15*time.Second))
	assert.NoError(t, err)

	u, ok := Between(first, second)
	assert.True(t, ok)
	assert.InDelta(t, 0.2, u.CPU, 1e-9)
	assert.InDelta(t, 0.1, u.CPUThrottled, 1e-9)
	assert.Equal(t, uint64(5000), u.Memory, "current less inactive_file")
	assert.Equal(t, uint64(1), u.OOMKills)
	assert.InDelta(t, 100.0, u.IORead, 1e-9)
	assert.InDelta(t, 2000.0, u.NetRx, 1e-9)
	assert.Zero(t, u.NetTx)

	_, ok = Between(second, first)
	assert.False(t, ok, "out of order")
}
