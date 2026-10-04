package obi

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// The reasons a node cannot run OBI.
const (
	// ReasonKernelTooOld: OBI needs kernel 5.8 or later.
	ReasonKernelTooOld = "kernel-too-old"
	// ReasonNoBTF: the kernel describes no types (/sys/kernel/btf/vmlinux).
	ReasonNoBTF = "no-btf"
	// ReasonContainerVirt: the node is a container itself (OpenVZ, LXC),
	// whose kernel is the host's and not its to probe.
	ReasonContainerVirt = "container-virt"
	// ReasonNoTracefs: no tracefs to attach probes through.
	ReasonNoTracefs = "no-tracefs"
	// ReasonLockdown: the kernel is locked down for confidentiality, which
	// forbids reading its memory as probes do.
	ReasonLockdown = "lockdown"
	// ReasonLowMemory: too little memory is free for OBI's.
	ReasonLowMemory = "low-memory"
)

const minKernelMajor, minKernelMinor = 5, 8

// Preflight is whether a node can run OBI, and why not; and the capacity it
// runs with, beside the one HivePaaS recommends for its memory.
type Preflight struct {
	OK             bool     `json:"ok"`
	Reasons        []string `json:"reasons,omitempty"`
	Kernel         string   `json:"kernel,omitempty"`
	MemTotalMB     int      `json:"memTotalMb,omitempty"`
	MemAvailableMB int      `json:"memAvailableMb,omitempty"`
	Recommended    Capacity `json:"recommended"`
	Capacity       Capacity `json:"capacity"`
}

// Check reads a node's filesystem, mounted at root - the agent's /host - for
// what OBI needs at a capacity, auto for the recommended one. Free memory is
// checked only when OBI is not running yet - once it runs, its own use is in
// it - and against twice what the capacity takes.
func Check(root string, running bool, capacity Capacity) Preflight {
	p := Preflight{Kernel: readTrimmed(filepath.Join(root, "proc/sys/kernel/osrelease"))}
	if !kernelAtLeast(p.Kernel, minKernelMajor, minKernelMinor) {
		p.Reasons = append(p.Reasons, ReasonKernelTooOld)
	}
	if !exists(filepath.Join(root, "sys/kernel/btf/vmlinux")) {
		p.Reasons = append(p.Reasons, ReasonNoBTF)
	}
	if containerVirt(root) {
		p.Reasons = append(p.Reasons, ReasonContainerVirt)
	}
	if !exists(filepath.Join(root, "sys/kernel/tracing/trace")) &&
		!exists(filepath.Join(root, "sys/kernel/debug/tracing/trace")) {
		p.Reasons = append(p.Reasons, ReasonNoTracefs)
	}
	if strings.Contains(readTrimmed(filepath.Join(root, "sys/kernel/security/lockdown")), "[confidentiality]") {
		p.Reasons = append(p.Reasons, ReasonLockdown)
	}
	meminfo := filepath.Join(root, "proc/meminfo")
	p.MemTotalMB, p.MemAvailableMB = memInfoMB(meminfo, "MemTotal:"), memInfoMB(meminfo, "MemAvailable:")
	p.Recommended = Recommended(p.MemTotalMB)
	p.Capacity = capacity.Effective(p.MemTotalMB)
	if !running && p.MemAvailableMB > 0 && p.MemAvailableMB < 2*p.Capacity.MemoryMiB() {
		p.Reasons = append(p.Reasons, ReasonLowMemory)
	}
	p.OK = len(p.Reasons) == 0
	return p
}

// kernelAtLeast compares a release such as "6.8.0-124-generic". An unreadable
// one is not old: it is OBI's to refuse, then.
func kernelAtLeast(release string, major, minor int) bool {
	parts := strings.SplitN(release, ".", 3) //nolint:mnd // major, minor, the rest
	if len(parts) < 2 {                      //nolint:mnd // major and minor
		return true
	}
	ma, err1 := strconv.Atoi(parts[0])
	mi, err2 := strconv.Atoi(strings.TrimFunc(parts[1], func(r rune) bool { return r < '0' || r > '9' }))
	if err1 != nil || err2 != nil {
		return true
	}
	return ma > major || (ma == major && mi >= minor)
}

// containerVirt is whether the node is a container: OpenVZ's /proc/vz without
// its host's /proc/bc, or an init started by LXC.
func containerVirt(root string) bool {
	if exists(filepath.Join(root, "proc/vz")) && !exists(filepath.Join(root, "proc/bc")) {
		return true
	}
	environ, err := os.ReadFile(filepath.Join(root, "proc/1/environ"))
	return err == nil && bytes.Contains(environ, []byte("container=lxc"))
}

// memInfoMB is a /proc/meminfo field, in MB; 0 when it cannot be read.
func memInfoMB(path, field string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer func() { _ = f.Close() }()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && fields[0] == field {
			kb, err := strconv.Atoi(fields[1])
			if err != nil {
				return 0
			}
			return kb / 1024 //nolint:mnd // kB to MB
		}
	}
	return 0
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func readTrimmed(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// RowStatus is the "hp" of a node's status row.
const RowStatus = "obi"

// How often an agent writes its node's status row: every minute while the
// feature is on, every 10 minutes while it is off - so that the nodes can be
// chosen knowing which can run OBI. The API reads the rows of a little more.
const (
	StatusEvery    = time.Minute
	StatusEveryOff = 10 * time.Minute
)

// Status is what a node can run and runs, as the agent's status row says it
// every minute while the logs are stored: the settings show each node's
// latest.
type Status struct {
	HP        string    `json:"hp"`
	Node      string    `json:"node"`
	Wanted    bool      `json:"wanted"`
	Running   bool      `json:"running"`
	Apps      int       `json:"apps"`
	Preflight Preflight `json:"preflight"`
}
