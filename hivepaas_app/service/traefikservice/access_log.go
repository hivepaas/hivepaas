package traefikservice

import (
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
	if !logidentity.HasComponent(spec, base.LogComponentTraefik) {
		return AccessLogUnlabelled
	}
	return ""
}
