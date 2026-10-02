// Package cgroupstat reads a container's resource counters from cgroup v2, as
// the kernel keeps them: no daemon in the way, and nothing it does not already
// count. A rate - CPU, network, disk - is the difference of two samples.
package cgroupstat

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ErrNoCgroup is a container whose cgroup v2 directory is not found: cgroup v1,
// or a layout this does not know.
var ErrNoCgroup = errors.New("no cgroup v2 directory for the container")

// Sample is a container's counters at one moment.
type Sample struct {
	At time.Time
	// CPUUsageUsec and CPUThrottledUsec are cumulative.
	CPUUsageUsec     uint64
	CPUThrottledUsec uint64
	// CPULimit is in cores; 0 when there is none.
	CPULimit float64
	// MemoryCurrent counts the page cache too; MemoryInactiveFile is the part
	// of it the kernel reclaims first.
	MemoryCurrent      uint64
	MemoryInactiveFile uint64
	// MemoryLimit is 0 when there is none.
	MemoryLimit uint64
	// OOMKills is cumulative.
	OOMKills uint64
	// IORead, IOWrite, NetRx and NetTx are cumulative bytes.
	IORead  uint64
	IOWrite uint64
	NetRx   uint64
	NetTx   uint64
}

// Dir finds a container's cgroup directory under root - the host's
// /sys/fs/cgroup as the reader sees it - for the systemd driver
// (system.slice/docker-<id>.scope) and the cgroupfs one (docker/<id>), under
// the container's cgroup parent when it has one.
func Dir(root, containerID, cgroupParent string) (string, error) {
	parents := []string{"system.slice", "docker"}
	if cgroupParent != "" {
		parents = append([]string{strings.Trim(cgroupParent, "/")}, parents...)
	}
	for _, parent := range parents {
		for _, name := range []string{"docker-" + containerID + ".scope", containerID} {
			dir := filepath.Join(root, parent, name)
			if _, err := os.Stat(filepath.Join(dir, "cpu.stat")); err == nil {
				return dir, nil
			}
		}
	}
	return "", ErrNoCgroup
}

// Read takes a sample of the cgroup in dir; netDev is the container's
// /proc/<pid>/net/dev, as the reader sees it, or empty for none.
func Read(dir, netDev string, at time.Time) (*Sample, error) {
	s := &Sample{At: at}
	cpu, err := keyValues(filepath.Join(dir, "cpu.stat"))
	if err != nil {
		return nil, err
	}
	s.CPUUsageUsec, s.CPUThrottledUsec = cpu["usage_usec"], cpu["throttled_usec"]
	s.CPULimit = cpuLimit(filepath.Join(dir, "cpu.max"))

	if s.MemoryCurrent, err = single(filepath.Join(dir, "memory.current")); err != nil {
		return nil, err
	}
	s.MemoryLimit, _ = single(filepath.Join(dir, "memory.max")) // "max" reads as none
	if mem, err := keyValues(filepath.Join(dir, "memory.stat")); err == nil {
		s.MemoryInactiveFile = mem["inactive_file"]
	}
	if events, err := keyValues(filepath.Join(dir, "memory.events")); err == nil {
		s.OOMKills = events["oom_kill"]
	}
	s.IORead, s.IOWrite = ioBytes(filepath.Join(dir, "io.stat"))
	if netDev != "" {
		s.NetRx, s.NetTx = netBytes(netDev)
	}
	return s, nil
}

// Usage is what a container used between two samples of it.
type Usage struct {
	// CPU is in cores; CPUThrottled the share of the time it was held back.
	CPU          float64
	CPUThrottled float64
	CPULimit     float64
	// Memory is the working set: what it uses less what the kernel reclaims
	// first, as `docker stats` counts it.
	Memory      uint64
	MemoryLimit uint64
	// OOMKills happened between the samples.
	OOMKills uint64
	// Per second.
	IORead  float64
	IOWrite float64
	NetRx   float64
	NetTx   float64
}

// Between is the usage from prev to cur: false when they cannot be compared -
// out of order, or counters that went back, as after a restart in place.
func Between(prev, cur *Sample) (*Usage, bool) {
	secs := cur.At.Sub(prev.At).Seconds()
	if secs <= 0 || cur.CPUUsageUsec < prev.CPUUsageUsec {
		return nil, false
	}
	u := &Usage{
		CPU:          float64(cur.CPUUsageUsec-prev.CPUUsageUsec) / 1e6 / secs,
		CPUThrottled: float64(sub(cur.CPUThrottledUsec, prev.CPUThrottledUsec)) / 1e6 / secs,
		CPULimit:     cur.CPULimit,
		Memory:       sub(cur.MemoryCurrent, cur.MemoryInactiveFile),
		MemoryLimit:  cur.MemoryLimit,
		OOMKills:     sub(cur.OOMKills, prev.OOMKills),
		IORead:       float64(sub(cur.IORead, prev.IORead)) / secs,
		IOWrite:      float64(sub(cur.IOWrite, prev.IOWrite)) / secs,
		NetRx:        float64(sub(cur.NetRx, prev.NetRx)) / secs,
		NetTx:        float64(sub(cur.NetTx, prev.NetTx)) / secs,
	}
	return u, true
}

func sub(a, b uint64) uint64 {
	if a < b {
		return 0
	}
	return a - b
}

// keyValues reads a file of "key value" lines.
func keyValues(path string) (map[string]uint64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cgroupstat: %w", err)
	}
	defer func() { _ = f.Close() }()
	out := map[string]uint64{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		key, value, ok := strings.Cut(scanner.Text(), " ")
		if !ok {
			continue
		}
		if n, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64); err == nil {
			out[key] = n
		}
	}
	return out, nil
}

// single reads a file holding one number; "max" reads as 0.
func single(path string) (uint64, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("cgroupstat: %w", err)
	}
	value := strings.TrimSpace(string(raw))
	if value == "max" {
		return 0, nil
	}
	n, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("cgroupstat: %s: %w", path, err)
	}
	return n, nil
}

// cpuLimit reads cpu.max, "<quota> <period>" or "max <period>", in cores.
func cpuLimit(path string) float64 {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(raw))
	if len(fields) != 2 || fields[0] == "max" { //nolint:mnd // quota and period
		return 0
	}
	quota, err1 := strconv.ParseFloat(fields[0], 64)
	period, err2 := strconv.ParseFloat(fields[1], 64)
	if err1 != nil || err2 != nil || period <= 0 {
		return 0
	}
	return quota / period
}

// ioBytes sums io.stat's rbytes and wbytes over its devices.
func ioBytes(path string) (read, write uint64) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, 0
	}
	for line := range strings.SplitSeq(string(raw), "\n") {
		for _, field := range strings.Fields(line) {
			key, value, ok := strings.Cut(field, "=")
			if !ok {
				continue
			}
			n, _ := strconv.ParseUint(value, 10, 64)
			switch key {
			case "rbytes":
				read += n
			case "wbytes":
				write += n
			}
		}
	}
	return read, write
}

// netBytes sums /proc/<pid>/net/dev's received and sent bytes, but loopback's.
func netBytes(path string) (rx, tx uint64) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, 0
	}
	for line := range strings.SplitSeq(string(raw), "\n") {
		iface, counters, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(iface) == "lo" {
			continue
		}
		fields := strings.Fields(counters)
		const txBytesField = 8 // receive has 8 fields, then transmit's bytes
		if len(fields) <= txBytesField {
			continue
		}
		r, _ := strconv.ParseUint(fields[0], 10, 64)
		t, _ := strconv.ParseUint(fields[txBytesField], 10, 64)
		rx += r
		tx += t
	}
	return rx, tx
}
