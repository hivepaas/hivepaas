// Package obi runs OBI (OpenTelemetry eBPF Instrumentation) on a node for the
// agent, and turns what it measures into rows: an app's routes and the calls
// it makes. See docs/superpowers/specs/2026-10-03-obi-calls-and-routes-design.md.
package obi

import (
	"crypto/sha256"
	"encoding/hex"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	// Image is the OBI release HivePaaS runs, pinned by digest: its metric
	// names and labels are what the agent reads.
	Image = "otel/ebpf-instrument:v0.14.0@sha256:0b063b9ec47e10dd4503ce1dc21f273acd3c85ffc9c135f0b753ee64de9c0819"

	// ContainerName is OBI's container on a node: one a node, the agent's.
	ContainerName = "hivepaas-obi"

	// MetricsPort is where OBI serves its metrics, in the agent's network
	// namespace: the agent reads them on localhost.
	MetricsPort = 19410

	// ConfigDir and ConfigFile are where OBI reads its configuration, copied
	// into its container before it starts: its image's root, the one
	// directory sure to be there - it has no shell, nor /etc/obi.
	ConfigDir  = "/"
	ConfigFile = "hivepaas-obi.yaml"

	// LabelConfig carries the hash of the configuration a container was made
	// with: one made with another is replaced.
	LabelConfig = "hivepaas.obi.config"

	// mapsScaleFactor shrinks OBI's eBPF maps to a quarter: about 100 MiB on a
	// node instead of 215, every request still counted. -3 lost a third.
	mapsScaleFactor = -2

	// memoryLimit bounds OBI's container, its maps included: four times what
	// it was measured at, and a node's other containers safe from it.
	memoryLimit = 384 << 20
)

// capabilities are what OBI needs to watch processes and attach its probes,
// short of CAP_SYS_ADMIN: measured on Ubuntu 24.04, kernel 6.8. Without
// CAP_SYS_ADMIN it cannot join another process's network namespace, and logs
// it; no count is lost.
var capabilities = []string{
	"CAP_BPF", "CAP_PERFMON", "CAP_SYS_PTRACE", "CAP_NET_RAW", "CAP_DAC_READ_SEARCH", "CAP_CHECKPOINT_RESTORE",
}

// Patterns are the container names OBI watches: every task's container of the
// apps given by their swarm services' names, `<service>.<slot>.<task>`. A
// service's name holds no dot, so one never matches another's.
func Patterns(services []string) []string {
	out := make([]string, 0, len(services))
	seen := make(map[string]bool, len(services))
	for _, s := range services {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s+".*")
	}
	sort.Strings(out)
	return out
}

// Config is OBI's configuration: the containers it watches, its maps shrunk.
// Patterns are quoted as JSON strings, which YAML reads as they are.
func Config(patterns []string) []byte {
	var b strings.Builder
	b.WriteString("discovery:\n  instrument:\n")
	for _, p := range patterns {
		b.WriteString("    - container_name: " + strconv.Quote(p) + "\n")
	}
	b.WriteString("ebpf:\n  maps_config:\n    global_scale_factor: " + strconv.Itoa(mapsScaleFactor) + "\n")
	return []byte(b.String())
}

// ConfigHash names a configuration, for a container to say which it runs.
func ConfigHash(config []byte) string {
	sum := sha256.Sum256(append([]byte(Image+"\n"), config...))
	return hex.EncodeToString(sum[:8])
}

// ContainerOptions is OBI's container on a node:
//   - the host's PID namespace, to see every container's processes;
//   - the agent's network namespace, to serve its metrics to the agent alone
//     on localhost, and to be gone when the agent is;
//   - tracefs and debugfs, and the Docker socket read-only, to name the
//     containers it is told to watch;
//   - the capabilities above, not privileged;
//   - its memory bounded, and its own log lines marked as OBI's.
func ContainerOptions(agentContainerID, configHash string) client.ContainerCreateOptions {
	return client.ContainerCreateOptions{
		Name: ContainerName,
		Config: &container.Config{
			Image: Image,
			Env: []string{
				"OTEL_EBPF_CONFIG_PATH=" + path.Join(ConfigDir, ConfigFile),
				"OTEL_EBPF_PROMETHEUS_PORT=" + strconv.Itoa(MetricsPort),
				"OTEL_EBPF_PROMETHEUS_FEATURES=application",
			},
			Labels: map[string]string{
				base.LabelLogComponent: base.LogComponentOBI,
				LabelConfig:            configHash,
			},
		},
		HostConfig: &container.HostConfig{
			PidMode:     "host",
			NetworkMode: container.NetworkMode("container:" + agentContainerID),
			CapAdd:      capabilities,
			Binds: []string{
				"/sys/kernel/tracing:/sys/kernel/tracing",
				"/sys/kernel/debug:/sys/kernel/debug",
				"/var/run/docker.sock:/var/run/docker.sock:ro",
			},
			Resources:     container.Resources{Memory: memoryLimit},
			RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyUnlessStopped},
			LogConfig: container.LogConfig{Type: "json-file", Config: map[string]string{
				"max-size": "10m", "max-file": "2", "labels": base.LabelLogComponent,
			}},
		},
	}
}
