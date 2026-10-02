// Package resourcesampler writes, for every app container on a node, a row of
// what it used since the last: CPU, memory, OOM kills, network and disk. The
// rows go to the agent's stdout, where the log collector takes them with the
// rest of its lines; the agent's container carries its identity, so a query
// can trust a row is the agent's and not one an app printed.
package resourcesampler

import (
	"context"
	"encoding/json"
	"io"
	"math"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/cgroupstat"
)

// Kind is the "hp" field of a row: what the queries look for.
const Kind = "resources"

// Container is an app's container on the node.
type Container struct {
	ID           string
	AppID        string
	Pid          int
	CgroupParent string
}

// Row is one container's usage over one interval.
type Row struct {
	HP        string  `json:"hp"`
	App       string  `json:"app"`
	Container string  `json:"container"`
	CPU       float64 `json:"cpu"`
	CPULimit  float64 `json:"cpuLimit,omitempty"`
	Throttled float64 `json:"throttled,omitempty"`
	Memory    uint64  `json:"memory"`
	MemLimit  uint64  `json:"memoryLimit,omitempty"`
	OOMKills  uint64  `json:"oomKills,omitempty"`
	NetRx     float64 `json:"netRx"`
	NetTx     float64 `json:"netTx"`
	IORead    float64 `json:"ioRead"`
	IOWrite   float64 `json:"ioWrite"`
}

// Sampler keeps each container's last sample, to tell what it used since.
type Sampler struct {
	// CgroupRoot and ProcRoot are the host's /sys/fs/cgroup and /proc as the
	// agent sees them.
	CgroupRoot string
	ProcRoot   string
	Out        io.Writer
	Now        func() time.Time

	mu   sync.Mutex
	prev map[string]*cgroupstat.Sample
}

// Tick samples the containers, and writes a row for each that was sampled the
// last time too. One that went is forgotten; one that is new gets its row next
// time. A container whose cgroup cannot be read is passed over.
func (s *Sampler) Tick(_ context.Context, containers []Container) (rows int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.prev == nil {
		s.prev = map[string]*cgroupstat.Sample{}
	}
	now := s.Now()
	seen := make(map[string]bool, len(containers))
	for _, c := range containers {
		dir, err := cgroupstat.Dir(s.CgroupRoot, c.ID, c.CgroupParent)
		if err != nil {
			continue
		}
		netDev := ""
		if c.Pid > 0 {
			netDev = filepath.Join(s.ProcRoot, strconv.Itoa(c.Pid), "net", "dev")
		}
		cur, err := cgroupstat.Read(dir, netDev, now)
		if err != nil {
			continue
		}
		seen[c.ID] = true
		prev := s.prev[c.ID]
		s.prev[c.ID] = cur
		if prev == nil {
			continue
		}
		usage, ok := cgroupstat.Between(prev, cur)
		if !ok {
			continue
		}
		if s.write(rowOf(c, usage)) {
			rows++
		}
	}
	for id := range s.prev {
		if !seen[id] {
			delete(s.prev, id)
		}
	}
	return rows
}

// Reset forgets every sample: after a pause, the next rows would otherwise
// cover the whole of it.
func (s *Sampler) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prev = nil
}

func (s *Sampler) write(row *Row) bool {
	line, err := json.Marshal(row)
	if err != nil {
		return false
	}
	// One write per line, so that lines from elsewhere in the process are not
	// interleaved into it.
	_, err = s.Out.Write(append(line, '\n'))
	return err == nil
}

// shortIDLen is the container id's length as docker shows it.
const shortIDLen = 12

func rowOf(c Container, u *cgroupstat.Usage) *Row {
	id := c.ID
	if len(id) > shortIDLen {
		id = id[:shortIDLen]
	}
	return &Row{
		HP: Kind, App: c.AppID, Container: id,
		CPU: round(u.CPU), CPULimit: round(u.CPULimit), Throttled: round(u.CPUThrottled),
		Memory: u.Memory, MemLimit: u.MemoryLimit, OOMKills: u.OOMKills,
		NetRx: math.Round(u.NetRx), NetTx: math.Round(u.NetTx),
		IORead: math.Round(u.IORead), IOWrite: math.Round(u.IOWrite),
	}
}

// round keeps four decimals: a row is a log line, and a thousandth of a core
// is past what anyone reads.
func round(v float64) float64 {
	return math.Round(v*1e4) / 1e4 //nolint:mnd
}
