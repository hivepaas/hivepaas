// Package vlagent configures VictoriaLogs' collector.
package vlagent

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

const (
	// DefaultImage is pinned. vlagent publishes no `latest` tag in any case,
	// and this is the version the spec's provenance properties were measured
	// against - that vlagent does not recursively parse a container's own JSON,
	// so a forged label in an app's stdout cannot displace the daemon's. A bump
	// is the trigger to re-run that check.
	DefaultImage = "victoriametrics/vlagent:v1.52.0"

	// ContainersPath is where docker keeps the json-file logs, mounted read-only.
	ContainersPath = "/var/lib/docker/containers"

	// CheckpointsPath is where vlagent records how far it has read. Without it a
	// restart re-sends every line it can still see.
	CheckpointsPath = "/vlagent-data/checkpoints.json"

	// TmpDataPath buffers logs that cannot be delivered yet.
	TmpDataPath = "/vlagent-data/remotewrite"

	// FormatNative is VictoriaLogs' own protocol; FormatJSONLine is
	// newline-delimited JSON, which any HTTP endpoint accepting NDJSON can read.
	FormatNative   = "native"
	FormatJSONLine = "jsonline"
)

type Config struct {
	Image string
}

type Collector struct {
	cfg *Config
}

func New(cfg *Config) *Collector {
	return &Collector{cfg: cfg}
}

// destination is one place logs are written, in the order vlagent's positional
// flags expect.
type destination struct {
	url         string
	format      string
	bearerToken string
	username    string
	password    string
	headers     map[string]string
}

// Configure renders the collection job as vlagent's command line.
//
// Every -remoteWrite.* array is matched to its -remoteWrite.url by position,
// and so is -fileCollector.excludeGlob to -fileCollector.glob. A destination or
// source with nothing to say for one of those flags still takes its slot, with
// an empty value: dropping the slot shifts every later value onto the wrong
// destination, which for a credential means sending it to the wrong system.
func (c *Collector) Configure(spec *loggingmodel.CollectSpec) (*loggingmodel.RuntimeSpec, error) {
	if spec.Ingest.URL == "" {
		return nil, hperrors.Wrap(loggingmodel.ErrIngestEndpointRequired)
	}
	if len(spec.Sources) == 0 {
		return nil, hperrors.Wrap(loggingmodel.ErrNoSources)
	}

	dests := []destination{{
		url:         spec.Ingest.URL,
		format:      FormatNative,
		bearerToken: spec.Ingest.BearerToken,
		username:    spec.Ingest.Username,
		password:    spec.Ingest.Password,
		headers:     spec.Ingest.Headers,
	}}
	for _, f := range spec.Forwards {
		format := f.Format
		if format == "" {
			format = FormatJSONLine
		}
		if format != FormatJSONLine && format != FormatNative {
			return nil, hperrors.Wrap(loggingmodel.ErrForwardFormatInvalid).WithParam("Name", format)
		}
		dests = append(dests, destination{
			url:         f.Endpoint.URL,
			format:      format,
			bearerToken: f.Endpoint.BearerToken,
			username:    f.Endpoint.Username,
			password:    f.Endpoint.Password,
			headers:     f.Endpoint.Headers,
		})
	}

	// One slot per destination for each of the six -remoteWrite.* arrays, one
	// per source for each of the three -fileCollector.* arrays, plus the two
	// path flags at the end.
	const (
		remoteWriteFlagsPerDest     = 6
		fileCollectorFlagsPerSource = 3
		trailingPathFlags           = 2
	)
	args := make([]string, 0,
		len(dests)*remoteWriteFlagsPerDest+
			len(spec.Sources)*fileCollectorFlagsPerSource+trailingPathFlags)
	for _, d := range dests {
		args = append(args, "-remoteWrite.url="+d.url)
	}
	for _, d := range dests {
		args = append(args, "-remoteWrite.format="+d.format)
	}
	for _, d := range dests {
		args = append(args, "-remoteWrite.bearerToken="+d.bearerToken)
	}
	for _, d := range dests {
		args = append(args, "-remoteWrite.basicAuth.username="+d.username)
	}
	for _, d := range dests {
		args = append(args, "-remoteWrite.basicAuth.password="+d.password)
	}
	for _, d := range dests {
		args = append(args, "-remoteWrite.headers="+joinHeaders(d.headers))
	}

	for _, s := range spec.Sources {
		args = append(args, "-fileCollector.glob="+s.Glob)
	}
	for _, s := range spec.Sources {
		args = append(args, "-fileCollector.excludeGlob="+s.Exclude)
	}
	for _, s := range spec.Sources {
		args = append(args, "-fileCollector.extraFields="+joinFields(s.Labels))
	}

	args = append(args,
		"-fileCollector.checkpointsPath="+CheckpointsPath,
		"-remoteWrite.tmpDataPath="+TmpDataPath,
	)

	image := c.cfg.Image
	if image == "" {
		image = DefaultImage
	}

	return &loggingmodel.RuntimeSpec{
		Image: image,
		Args:  args,
		Mounts: []loggingmodel.Mount{{
			Source:   ContainersPath,
			Target:   ContainersPath,
			ReadOnly: true,
		}},
	}, nil
}

// joinHeaders renders headers the way vlagent parses them, with a stable order
// so that the same configuration always produces the same command line - a
// service spec that differs run to run is a service that redeploys for nothing.
func joinHeaders(h map[string]string) string {
	if len(h) == 0 {
		return ""
	}
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s: %s", k, h[k]))
	}
	return strings.Join(parts, "^^")
}

// joinFields renders extra log fields as key=value pairs, stably ordered for
// the same reason.
func joinFields(f map[string]string) string {
	if len(f) == 0 {
		return ""
	}
	keys := make([]string, 0, len(f))
	for k := range f {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, f[k]))
	}
	return strings.Join(parts, ",")
}

// AttrField is the field a container label named in json-file's `labels`
// option is stored under. The daemon writes the label into the line's attrs
// object, and vlagent flattens that object with this prefix.
func AttrField(label string) string {
	return "attrs." + label
}
