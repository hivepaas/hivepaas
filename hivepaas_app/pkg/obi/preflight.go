package obi

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

const (
	minKernelMajor, minKernelMinor = 5, 8
	// minMemAvailableMB is the free memory a node needs to start OBI: what it
	// was measured at, twice.
	minMemAvailableMB = 200
)

// Preflight is whether a node can run OBI, and why not.
type Preflight struct {
	OK             bool     `json:"ok"`
	Reasons        []string `json:"reasons,omitempty"`
	Kernel         string   `json:"kernel,omitempty"`
	MemAvailableMB int      `json:"memAvailableMb,omitempty"`
}

// Check reads a node's filesystem, mounted at root - the agent's /host - for
// what OBI needs. Free memory is checked only when OBI is not running yet: once
// it runs, its own use is in it.
func Check(root string, running bool) Preflight {
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
	p.MemAvailableMB = memAvailableMB(filepath.Join(root, "proc/meminfo"))
	if !running && p.MemAvailableMB > 0 && p.MemAvailableMB < minMemAvailableMB {
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

func memAvailableMB(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer func() { _ = f.Close() }()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && fields[0] == "MemAvailable:" {
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
