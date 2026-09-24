package loggingserviceimpl

import (
	"strings"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/executil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/services/logging"
)

// logDriverLocal keeps the logging stack's own output away from the collector.
//
// The collector globs *-json.log; `local` writes a different format under
// local-logs/, so the stack's containers simply do not match. Excluding them by
// path is not possible - the path is a container id, never a service name - and
// collecting them feeds the collector its own output. `docker service logs`
// still reads them, and so does the app's log screen.
const logDriverLocal = "local"

// secretFileMode is the mode a credential file is mounted with.
const secretFileMode = "0444"

// buildAppDoc turns what an app of the stack runs into the document it is built
// from - the same document a template's app is, so the stack cannot drift away
// from what an app can be.
//
// It says what a document can say. The rest - one task per node, a host path,
// a second network, the log driver - is customizeSpec's.
func buildAppDoc(rt *logging.RuntimeSpec, engine string) (*specmodel.AppDoc, error) {
	deployment := &specmodel.Deployment{
		Source: map[string]any{
			"activeMethod": string(base.DeploymentMethodImage),
			"imageSource":  map[string]any{"image": rt.Image},
			"command":      commandLine(rt.Args),
		},
		// Neither image has a shell, so there is nothing to run a check with.
		Container: &specmodel.Container{Healthcheck: &specmodel.Healthcheck{Enabled: false}},
		Resources: resourcesBlock(rt.Resources),
	}
	for _, m := range rt.Mounts {
		if m.Volume == "" {
			continue
		}
		if deployment.Storage == nil {
			deployment.Storage = &specmodel.Storage{Mounts: map[string]specmodel.Mount{}}
		}
		deployment.Storage.Mounts[m.Target] = specmodel.Mount{
			Type:     mount.TypeVolume,
			Source:   m.Volume,
			ReadOnly: m.ReadOnly,
		}
	}

	settings := map[string]any{
		specmodel.SingletonBlockName(base.SettingTypeAppKind): map[string]any{
			"category": string(base.AppCategoryWebapp),
			"engine":   engine,
			"webapp":   map[string]any{},
		},
	}
	if len(rt.Secrets) > 0 {
		secrets := make(map[string]any, len(rt.Secrets))
		for _, secret := range rt.Secrets {
			secrets[secret.Key] = map[string]any{
				"value": secret.Value,
				"swarmRef": map[string]any{
					"file": map[string]any{"name": secret.Path, "mode": secretFileMode},
				},
			}
		}
		settings[specmodel.CollectionBlockName(base.SettingTypeSecret)] = secrets
	}

	doc := &specmodel.AppDoc{Deployment: deployment, Settings: settings}
	if err := specmodel.CheckBuildable(doc); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return doc, nil
}

// commandLine is the arguments as the app's deployment settings store them.
//
// Every deployment sets the container's arguments from that string, split the
// way a shell would, so each argument is quoted: a JSON value, a header with a
// space in it, reach the container exactly as they were written.
func commandLine(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = executil.ArgQuote(arg)
	}
	return strings.Join(quoted, " ")
}

// resourcesBlock is the limits a document gives an app. The OOM priority goes
// only with a memory limit: see base.OomScoreAdjSystemAddon.
func resourcesBlock(r logging.Resources) *specmodel.Resources {
	if r.CPULimit <= 0 && r.MemoryLimit <= 0 {
		return nil
	}
	block := &specmodel.Resources{Limits: &specmodel.ResourceLimits{
		CPUs:   r.CPULimit,
		Memory: unit.DataSize(r.MemoryLimit),
	}}
	if r.MemoryLimit > 0 {
		block.Capabilities = &specmodel.Capabilities{OomScoreAdj: base.OomScoreAdjSystemAddon}
	}
	return block
}

// customizeSpec writes onto the service what a document is not allowed to say.
//
// A template may not ask for a host path, a network of the stack or one task on
// every node, because a template is written by somebody else. These apps are
// written here, and they need exactly those.
func customizeSpec(rt *logging.RuntimeSpec, networks []string, alias string) func(*swarm.ServiceSpec) error {
	return func(spec *swarm.ServiceSpec) error {
		if rt.PerNode {
			// Exactly one mode: the daemon rejects a spec carrying two.
			spec.Mode = swarm.ServiceMode{Global: &swarm.GlobalService{}}
		}
		container := spec.TaskTemplate.ContainerSpec
		for _, m := range rt.Mounts {
			if m.Source == "" {
				continue
			}
			container.Mounts = append(container.Mounts, mount.Mount{
				Type:     mount.TypeBind,
				Source:   m.Source,
				Target:   m.Target,
				ReadOnly: m.ReadOnly,
			})
		}
		// Replaced whole, options and all: the default driver's options name its
		// own format, which `local` does not write.
		spec.TaskTemplate.LogDriver = &swarm.Driver{Name: logDriverLocal}
		for _, network := range networks {
			spec.TaskTemplate.Networks = append(spec.TaskTemplate.Networks,
				swarm.NetworkAttachmentConfig{Target: network, Aliases: []string{alias}})
		}
		return nil
	}
}
