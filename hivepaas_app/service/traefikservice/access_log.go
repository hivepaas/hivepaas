package traefikservice

import (
	"slices"
	"strings"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logidentity"
	"github.com/hivepaas/hivepaas/services/traefik/traefikhelper"
)

// WithAccessLogIdentity marks the proxy's lines as its own: see logidentity.
// It reports whether it changed the spec.
func WithAccessLogIdentity(spec *swarm.ServiceSpec) bool {
	return logidentity.WithComponent(spec, base.LogComponentTraefik)
}

// WithAccessLogArgs writes the Access Log option's arguments as this release
// has them, where the proxy writes an access log: what a save of the option
// would write - the fields it keeps among them. The operator's own arguments
// stay as they were, and the option's go where its first one was. It reports
// whether it changed the spec.
func WithAccessLogArgs(spec *swarm.ServiceSpec) bool {
	cs := spec.TaskTemplate.ContainerSpec
	if cs == nil {
		return false
	}
	on, at := false, -1
	args := make([]string, 0, len(cs.Args)+len(base.TraefikAccessLogArgs))
	for _, arg := range cs.Args {
		key, val, valid := traefikhelper.ParseCommandArg(arg)
		if !valid || !base.IsTraefikAccessLogArg(key) {
			args = append(args, arg)
			continue
		}
		if key == "accesslog" {
			on = isOn(val)
		}
		if at < 0 {
			at = len(args)
		}
	}
	if !on {
		return false
	}
	args = slices.Insert(args, at, base.TraefikAccessLogArgs...)
	if slices.Equal(args, cs.Args) {
		return false
	}
	cs.Args = args
	return true
}

// isOn reads a flag's value as traefik does: given with none, it is on.
func isOn(val string) bool {
	return val == "" || strings.EqualFold(val, "true")
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
			on = isOn(val)
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
	if !logidentity.HasComponent(spec, base.LogComponentTraefik) {
		return AccessLogUnlabelled
	}
	return ""
}
