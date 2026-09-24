// Package victorialogs talks to VictoriaLogs and describes how to run it.
package victorialogs

import (
	"fmt"
	"math"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

const (
	// DefaultImage is pinned rather than tracking latest: an unannounced storage
	// format change arriving on a restart is not a surprise worth having. It is
	// also the version the provenance properties in the spec were measured
	// against, so a bump is the trigger to re-run them.
	DefaultImage = "victoriametrics/victoria-logs:v1.52.0"

	// DefaultHTTPPort is where VictoriaLogs serves ingest, query and health.
	DefaultHTTPPort = 9428

	// DataPath is where the data volume is mounted inside the container.
	DataPath = "/victoria-logs-data"

	// IngestPath is the native protocol endpoint vlagent writes to.
	IngestPath = "/internal/insert"

	// HealthPath answers OK once the service is serving.
	HealthPath = "/health"

	// QueryPath runs a LogsQL query.
	QueryPath = "/select/logsql/query"

	// minRetentionDays is VictoriaLogs' own floor; it rejects anything shorter.
	minRetentionDays = 1

	// hoursPerDay is what -retentionPeriod is expressed in.
	hoursPerDay = 24
)

type Config struct {
	// Image overrides DefaultImage. The image belongs to the release rather than
	// to anything the user configures, and it is what the system updater moves;
	// leaving it empty is what a caller with no release information does.
	Image string

	// DataVolume is the volume the store keeps its data on, as the caller refers
	// to it: see loggingmodel.Mount.Volume.
	DataVolume string
	Retention  time.Duration

	// MaxDiskUsagePercent caps the share of the filesystem the logs may take
	// before the oldest days are dropped. Zero leaves it unset.
	//
	// It measures the filesystem, not the subdirectory: a volume shared with
	// something else counts that something else against this cap.
	MaxDiskUsagePercent int

	// Resources caps what the container may take from its node. The zero value
	// is no cap, which is what running without this setting means.
	Resources loggingmodel.Resources

	// Endpoint is how to reach an already-running instance. It is what the read
	// path uses, and is unset while only RuntimeSpec is needed.
	Endpoint loggingmodel.Endpoint
}

type Client struct {
	cfg *Config
}

func New(cfg *Config) *Client {
	return &Client{cfg: cfg}
}

// RuntimeSpec describes the VictoriaLogs container.
func (c *Client) RuntimeSpec() (*loggingmodel.RuntimeSpec, error) {
	if c.cfg.DataVolume == "" {
		// Without a volume the logs live in the container's writable layer and
		// vanish on the next restart, silently.
		return nil, hperrors.Wrap(loggingmodel.ErrDataVolumeRequired)
	}

	args := []string{
		fmt.Sprintf("-storageDataPath=%s", DataPath),
		fmt.Sprintf("-retentionPeriod=%dd", retentionDays(c.cfg.Retention)),
		fmt.Sprintf("-httpListenAddr=:%d", DefaultHTTPPort),
	}
	if c.cfg.MaxDiskUsagePercent > 0 {
		args = append(args, fmt.Sprintf("-retention.maxDiskUsagePercent=%d", c.cfg.MaxDiskUsagePercent))
	}

	image := c.cfg.Image
	if image == "" {
		image = DefaultImage
	}

	return &loggingmodel.RuntimeSpec{
		Image:     image,
		Args:      args,
		Resources: c.cfg.Resources,
		Mounts: []loggingmodel.Mount{{
			Volume: c.cfg.DataVolume,
			Target: DataPath,
		}},
		// No published port. VictoriaLogs has no authentication and no tenancy,
		// so it is reachable only on the overlay network the API shares with it.
		Ports: nil,
	}, nil
}

// retentionDays converts a duration to whole days, never below VictoriaLogs'
// own one-day minimum: a shorter setting would otherwise produce a service that
// refuses to start.
func retentionDays(d time.Duration) int {
	days := int(math.Ceil(d.Hours() / hoursPerDay))
	if days < minRetentionDays {
		return minRetentionDays
	}
	return days
}
