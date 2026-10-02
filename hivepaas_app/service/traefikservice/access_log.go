package traefikservice

import (
	"maps"
	"sort"
	"strings"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/services/traefik/traefikhelper"
)

const logDriverJSONFile = "json-file"

// WithAccessLogIdentity marks the proxy's lines as its own: the container
// label naming it, and the json-file driver's `labels` option copying that
// label into every line. A service with no driver of its own gets json-file,
// the daemon's default, with the limits the stack gives it; one with another
// driver is left alone - its lines are not collected anyway. It reports
// whether it changed the spec.
func WithAccessLogIdentity(spec *swarm.ServiceSpec) bool {
	if spec == nil {
		return false
	}
	if spec.TaskTemplate.ContainerSpec == nil {
		spec.TaskTemplate.ContainerSpec = &swarm.ContainerSpec{}
	}
	cs := spec.TaskTemplate.ContainerSpec
	changed := false
	if cs.Labels[base.LabelLogComponent] != base.LogComponentTraefik {
		labels := make(map[string]string, len(cs.Labels)+1)
		maps.Copy(labels, cs.Labels)
		labels[base.LabelLogComponent] = base.LogComponentTraefik
		cs.Labels = labels
		changed = true
	}

	d := spec.TaskTemplate.LogDriver
	if d == nil || d.Name == "" {
		d = &swarm.Driver{Name: logDriverJSONFile,
			Options: map[string]string{"max-size": "50m", "max-file": "5", "compress": "true"}}
		changed = true
	}
	if d.Name != logDriverJSONFile {
		return changed
	}
	if !hasLabelsOption(d, base.LabelLogComponent) {
		names := []string{base.LabelLogComponent}
		for n := range strings.SplitSeq(d.Options["labels"], ",") {
			if n = strings.TrimSpace(n); n != "" && n != base.LabelLogComponent {
				names = append(names, n)
			}
		}
		// Sorted, so that the same service always has the same spec.
		sort.Strings(names)
		opts := make(map[string]string, len(d.Options)+1)
		maps.Copy(opts, d.Options)
		opts["labels"] = strings.Join(names, ",")
		d = &swarm.Driver{Name: d.Name, Options: opts}
		changed = true
	}
	spec.TaskTemplate.LogDriver = d
	return changed
}

func hasLabelsOption(d *swarm.Driver, label string) bool {
	for n := range strings.SplitSeq(d.Options["labels"], ",") {
		if strings.TrimSpace(n) == label {
			return true
		}
	}
	return false
}

// AccessLogNotReadyReason says why an app's HTTP numbers cannot be counted
// from the proxy's access log; empty when they can.
type AccessLogNotReadyReason string

const (
	// AccessLogOff: the proxy writes no access log.
	AccessLogOff AccessLogNotReadyReason = "access-log-off"
	// AccessLogNotJSON: the access log is not JSON, and has no field to count by.
	AccessLogNotJSON AccessLogNotReadyReason = "access-log-not-json"
	// AccessLogUnlabelled: the proxy's lines do not carry its identity, so they
	// cannot be told apart from a line an app printed.
	AccessLogUnlabelled AccessLogNotReadyReason = "access-log-unlabelled"
)

// AccessLogReadiness reads the proxy's live spec: whether it writes a JSON
// access log, and marks its lines as its own.
func AccessLogReadiness(spec *swarm.ServiceSpec) AccessLogNotReadyReason {
	if spec == nil || spec.TaskTemplate.ContainerSpec == nil {
		return AccessLogOff
	}
	on, json := false, false
	for _, arg := range spec.TaskTemplate.ContainerSpec.Args {
		key, val, valid := traefikhelper.ParseCommandArg(arg)
		if !valid {
			continue
		}
		switch key {
		case "accesslog":
			on = val == "" || strings.EqualFold(val, "true")
		case "accesslog.format":
			json = strings.EqualFold(val, "json")
		}
	}
	switch {
	case !on:
		return AccessLogOff
	case !json:
		return AccessLogNotJSON
	}
	cs := spec.TaskTemplate.ContainerSpec
	d := spec.TaskTemplate.LogDriver
	if cs.Labels[base.LabelLogComponent] != base.LogComponentTraefik ||
		d == nil || d.Name != logDriverJSONFile || !hasLabelsOption(d, base.LabelLogComponent) {
		return AccessLogUnlabelled
	}
	return ""
}
