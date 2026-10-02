// Package logidentity marks one of HivePaaS's own services in its log lines.
//
// The container label base.LabelLogComponent names the service, and the
// json-file driver's `labels` option copies it into every line the container
// writes - the daemon does, not the container - so that a query can trust a
// line is, say, the proxy's or the agent's, and not one an app printed to look
// like it.
package logidentity

import (
	"maps"
	"sort"
	"strings"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const logDriverJSONFile = "json-file"

// WithComponent marks a service's lines as the component's: the container
// label, and the json-file driver's `labels` option naming it. A service with
// no driver of its own gets json-file, the daemon's default, with the limits
// HivePaaS's stack gives its services; one with another driver keeps it - its
// lines are not collected anyway. It reports whether it changed the spec.
func WithComponent(spec *swarm.ServiceSpec, component string) bool {
	if spec == nil {
		return false
	}
	if spec.TaskTemplate.ContainerSpec == nil {
		spec.TaskTemplate.ContainerSpec = &swarm.ContainerSpec{}
	}
	cs := spec.TaskTemplate.ContainerSpec
	changed := false
	if cs.Labels[base.LabelLogComponent] != component {
		labels := make(map[string]string, len(cs.Labels)+1)
		maps.Copy(labels, cs.Labels)
		labels[base.LabelLogComponent] = component
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
	if !hasLabelsOption(d) {
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

// HasComponent reports whether a service's lines carry the component's mark.
func HasComponent(spec *swarm.ServiceSpec, component string) bool {
	if spec == nil || spec.TaskTemplate.ContainerSpec == nil {
		return false
	}
	d := spec.TaskTemplate.LogDriver
	return spec.TaskTemplate.ContainerSpec.Labels[base.LabelLogComponent] == component &&
		d != nil && d.Name == logDriverJSONFile && hasLabelsOption(d)
}

func hasLabelsOption(d *swarm.Driver) bool {
	for n := range strings.SplitSeq(d.Options["labels"], ",") {
		if strings.TrimSpace(n) == base.LabelLogComponent {
			return true
		}
	}
	return false
}
