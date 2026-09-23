package specserviceimpl

import (
	"context"
	"maps"
	"slices"
	"time"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
)

func (s *service) buildContainer(_ context.Context, state *buildState) error {
	return applyContainer(state.req.Doc.Deployment.Container, state.req.Spec)
}

// applyContainer writes the container block the way the container settings
// screen does, replacing what the spec held. Three things stay HivePaaS's: the
// image, which only a deployment changes; the labels HivePaaS put on the service
// and the container, which ApplyUserLabels keeps; and the log driver, which is
// HivePaaS's default when the block names none.
func applyContainer(c *specmodel.Container, spec *swarm.ServiceSpec) error {
	if c == nil {
		c = &specmodel.Container{}
	}
	task := &spec.TaskTemplate
	cs := task.ContainerSpec

	spec.Labels = dockerhelper.ApplyUserLabels(spec.Labels, c.ServiceLabels)
	cs.Labels = dockerhelper.ApplyUserLabels(cs.Labels, c.ContainerLabels)
	cs.Hostname, cs.User, cs.StopSignal = c.Hostname, c.User, c.StopSignal
	cs.Groups = slices.Clone(c.Groups)
	cs.TTY, cs.OpenStdin, cs.ReadOnly = c.TTY, c.OpenStdin, c.ReadOnly
	cs.Init = nil
	if c.Init != nil {
		cs.Init = new(*c.Init)
	}
	cs.StopGracePeriod = nil
	if c.StopGracePeriod != nil {
		cs.StopGracePeriod = new(time.Duration(*c.StopGracePeriod))
	}
	cs.Privileges = toDockerPrivileges(c.Privileges)
	if err := applyHealthcheck(c.Healthcheck, cs); err != nil {
		return err
	}
	task.RestartPolicy = toDockerRestartPolicy(c.RestartPolicy)
	task.LogDriver = appservice.DefaultLogDriver()
	if c.LogDriver != nil && c.LogDriver.Name != "" {
		task.LogDriver = &swarm.Driver{Name: c.LogDriver.Name, Options: maps.Clone(c.LogDriver.Options)}
	}
	return nil
}

func toDockerPrivileges(p *specmodel.Privileges) *swarm.Privileges {
	if p == nil {
		return nil
	}
	out := &swarm.Privileges{NoNewPrivileges: p.NoNewPrivileges}
	if ctx := p.SELinuxContext; ctx != nil {
		out.SELinuxContext = &swarm.SELinuxContext{
			Disable: ctx.Disable, User: ctx.User, Role: ctx.Role, Type: ctx.Type, Level: ctx.Level,
		}
	}
	if seccomp := p.Seccomp; seccomp != nil {
		out.Seccomp = &swarm.SeccompOpts{Mode: seccomp.Mode}
		if seccomp.Profile != "" {
			out.Seccomp.Profile = []byte(seccomp.Profile)
		}
	}
	if appArmor := p.AppArmor; appArmor != nil {
		out.AppArmor = &swarm.AppArmorOpts{Mode: appArmor.Mode}
	}
	return out
}

func toDockerRestartPolicy(p *specmodel.RestartPolicy) *swarm.RestartPolicy {
	if p == nil {
		return nil
	}
	out := &swarm.RestartPolicy{Condition: p.Condition}
	if p.MaxAttempts != nil {
		out.MaxAttempts = new(*p.MaxAttempts)
	}
	if p.Delay != nil {
		out.Delay = new(time.Duration(*p.Delay))
	}
	if p.Window != nil {
		out.Window = new(time.Duration(*p.Window))
	}
	return out
}
