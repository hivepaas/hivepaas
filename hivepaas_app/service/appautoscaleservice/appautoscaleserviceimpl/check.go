package appautoscaleserviceimpl

import (
	"context"
	"errors"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logidentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appautoscaleservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
)

func (s *service) Check(
	ctx context.Context, db database.IDB, app *entity.App, svc *swarm.Service,
) (*appautoscaleservice.Check, error) {
	check := &appautoscaleservice.Check{}
	history, err := s.loggingService.ProxyHistory(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !history.Available {
		check.Requests, check.CPU = string(history.Reason), string(history.Reason)
	}

	if check.Requests == "" {
		if setting := app.GetSettingByType(base.SettingTypeAppRouting); setting == nil || !isExposed(setting) {
			check.Requests = appautoscaleservice.ReasonNotExposed
		} else if check.Requests, err = s.accessLogReadiness(ctx); err != nil {
			return nil, err
		}
	}
	if check.CPU == "" {
		if check.CPU, err = s.agentReadiness(ctx); err != nil {
			return nil, err
		}
	}
	if svc == nil {
		return check, nil
	}
	if check.CPU == "" && !hasAppLabel(&svc.Spec, app.ID) {
		check.CPU = appautoscaleservice.ReasonIdentityMissing
	}
	if check.CPU == "" && cpuLimit(&svc.Spec) == 0 && cpuReservation(&svc.Spec) == 0 {
		check.CPU = appautoscaleservice.ReasonNoCPULimit
	}

	switch {
	case svc.Spec.Mode.Replicated == nil:
		check.Refused = appautoscaleservice.RefusedNotReplicated
	case hasHostPorts(&svc.Spec):
		check.Refused = appautoscaleservice.RefusedHostPorts
	}
	check.WritableMounts = hasWritableMounts(&svc.Spec)
	if check.Pending, err = s.pending(ctx, svc.ID); err != nil {
		return nil, err
	}
	return check, nil
}

// pending is how many of a service's tasks are wanted and not running.
func (s *service) pending(ctx context.Context, serviceID string) (int, error) {
	list, err := s.dockerManager.ServiceListByIDs(ctx, []string{serviceID}, func(opts *client.ServiceListOptions) {
		opts.Status = true
	})
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	for _, svc := range list.Items {
		if svc.ServiceStatus != nil && svc.ServiceStatus.DesiredTasks > svc.ServiceStatus.RunningTasks {
			return int(svc.ServiceStatus.DesiredTasks - svc.ServiceStatus.RunningTasks), nil //nolint:gosec // tasks
		}
	}
	return 0, nil
}

// accessLogReadiness is why the proxy's access log cannot be read for
// requests, "" when it can.
func (s *service) accessLogReadiness(ctx context.Context) (string, error) {
	svc, err := s.traefikService.GetTraefikSwarmService(ctx)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return string(traefikservice.AccessLogReadiness(&svc.Spec)), nil
}

// agentReadiness is why the agent's rows cannot be read for CPU, "" when they
// can: the agent must mark them as its own.
func (s *service) agentReadiness(ctx context.Context) (string, error) {
	svc, err := s.hpAppService.GetHpAgentSwarmService(ctx)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return "", hperrors.Wrap(err)
	}
	if svc == nil || !logidentity.HasComponent(&svc.Spec, base.LogComponentAgent) {
		return appautoscaleservice.ReasonAgentUnlabelled, nil
	}
	return "", nil
}

// isExposed is whether an app's routing reaches it by a domain through the
// proxy.
func isExposed(setting *entity.Setting) bool {
	routing, err := setting.AsAppRoutingSettings()
	if err != nil || !routing.ExposePublicly {
		return false
	}
	for _, domain := range routing.Domains {
		if domain.Enabled {
			return true
		}
	}
	return false
}

// hasAppLabel is whether an app's containers carry its id, which the agent's
// rows name them by.
func hasAppLabel(spec *swarm.ServiceSpec, appID string) bool {
	cs := spec.TaskTemplate.ContainerSpec
	return cs != nil && cs.Labels[appservice.LabelLogAppID] == appID
}

// cpuLimit and cpuReservation are an instance's CPU limit and reservation, in
// cores; 0 for none.
func cpuLimit(spec *swarm.ServiceSpec) float64 {
	if r := spec.TaskTemplate.Resources; r != nil && r.Limits != nil {
		return float64(r.Limits.NanoCPUs) / 1e9 //nolint:mnd // nano
	}
	return 0
}

func cpuReservation(spec *swarm.ServiceSpec) float64 {
	if r := spec.TaskTemplate.Resources; r != nil && r.Reservations != nil {
		return float64(r.Reservations.NanoCPUs) / 1e9 //nolint:mnd // nano
	}
	return 0
}

// hasHostPorts is whether a service publishes a port on its node rather than
// through the routing mesh: one replica a node at most.
func hasHostPorts(spec *swarm.ServiceSpec) bool {
	if spec.EndpointSpec == nil {
		return false
	}
	for _, port := range spec.EndpointSpec.Ports {
		if port.PublishMode == swarm.PortConfigPublishModeHost {
			return true
		}
	}
	return false
}

// hasWritableMounts is whether a service writes to a volume or a bind mount:
// replicas on one node share it, and those on another each have their own.
func hasWritableMounts(spec *swarm.ServiceSpec) bool {
	cs := spec.TaskTemplate.ContainerSpec
	if cs == nil {
		return false
	}
	for _, m := range cs.Mounts {
		if (m.Type == mount.TypeVolume || m.Type == mount.TypeBind) && !m.ReadOnly {
			return true
		}
	}
	return false
}
