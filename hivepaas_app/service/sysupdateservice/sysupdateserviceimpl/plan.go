package sysupdateserviceimpl

import (
	"context"
	"encoding/json"
	"slices"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/imageref"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sysupdateservice"
)

// planStep is one component an update reaches, found the way the update finds it.
type planStep struct {
	key   string
	image string
	fetch func(ctx context.Context) (*swarm.Service, error)
	// align is the step's serviceImageUpdate.Align, if it has one.
	align func(spec *swarm.ServiceSpec) bool
}

func (s *service) PlanUpdate(
	ctx context.Context,
	db database.IDB,
	target *base.ReleaseInfo,
) (*sysupdateservice.UpdatePlan, error) {
	steps := []planStep{
		{key: base.HivepaasDbKey, image: target.DbImage, fetch: s.hpAppService.GetHpDbSwarmService},
		{key: base.HivepaasCacheKey, image: target.RedisImage, fetch: s.hpAppService.GetHpCacheSwarmService},
		{key: base.HivepaasTraefikKey, image: target.TraefikImage, fetch: s.traefikService.GetTraefikSwarmService,
			align: alignTraefik},
	}
	for _, app := range []struct{ key, image string }{
		{base.HivepaasVictoriaLogsKey, target.VictoriaLogsImage},
		{base.HivepaasVlagentKey, target.VlagentImage},
		{base.HivepaasRegistryKey, target.RegistryImage},
	} {
		loaded, err := s.systemAppService.LoadApp(ctx, db, app.key)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		steps = append(steps, planStep{key: app.key, image: app.image,
			fetch: func(ctx context.Context) (*swarm.Service, error) {
				return s.systemAppSwarmService(ctx, loaded)
			}})
	}
	steps = append(steps,
		planStep{key: base.HivepaasAgentKey, image: target.AgentImage, fetch: s.getAgentSwarmService},
		planStep{key: base.HivepaasAppKey, image: target.AppImage, fetch: s.hpAppService.GetHpAppSwarmService},
		planStep{key: base.HivepaasWorkerKey, image: target.AppImage, fetch: s.getWorkerSwarmService},
	)

	plan := &sysupdateservice.UpdatePlan{}
	for _, step := range steps {
		svc, err := step.fetch(ctx)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		current, deployed, aligns := "", svc != nil, false
		if deployed {
			current = svc.Spec.TaskTemplate.ContainerSpec.Image
			if aligns, err = wouldAlign(svc, step.align); err != nil {
				return nil, hperrors.Wrap(err)
			}
		}
		change := planComponent(step.key, current, deployed, step.image, aligns, target.BlockMajorUpgrade)
		plan.Components = append(plan.Components, change)
		plan.Blocked = plan.Blocked || change.Change == sysupdateservice.ChangeBlocked
		plan.RequiresBackup = plan.RequiresBackup || change.RequiresBackup
		if step.key == base.HivepaasAgentKey {
			// OBI moves with the agent: each node's runs its own release's.
			obiChange, err := s.obiChange(ctx, db, target)
			if err != nil {
				return nil, hperrors.Wrap(err)
			}
			plan.Components = append(plan.Components, obiChange)
			plan.Blocked = plan.Blocked || obiChange.Change == sysupdateservice.ChangeBlocked
		}
	}
	return plan, nil
}

// wouldAlign reports whether a step's Align would change the service, on a copy
// of its spec: the plan changes nothing. It is this version's Align, not the
// target's, so it sees only what this version writes; see ChangeSettings.
func wouldAlign(svc *swarm.Service, align func(spec *swarm.ServiceSpec) bool) (bool, error) {
	if align == nil {
		return false, nil
	}
	b, err := json.Marshal(svc.Spec)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	var spec swarm.ServiceSpec
	if err := json.Unmarshal(b, &spec); err != nil {
		return false, hperrors.Wrap(err)
	}
	return align(&spec), nil
}

// planComponent decides what an update does to one component, by the rules
// updateServiceImage applies: imageref.IsUpgrade for whether to move at all,
// the release's BlockMajorUpgrade for a move across a major, and whether the
// step's Align would change the service, aligns, for one whose image stays.
func planComponent(
	key, current string,
	deployed bool,
	target string,
	aligns bool,
	blockMajor []string,
) *sysupdateservice.ComponentChange {
	change := &sysupdateservice.ComponentChange{Key: key, CurrentImage: current, TargetImage: target}
	switch {
	case target == "":
		change.Change, change.Reason = sysupdateservice.ChangeNone, "the release names no image for it"
		return change
	case !deployed:
		change.Change, change.Reason = sysupdateservice.ChangeNotDeployed, "not deployed"
		return change
	}

	apply, reason := imageref.IsUpgrade(current, target)
	change.Reason = reason
	if !apply {
		change.Change = sysupdateservice.ChangeNone
		if aligns {
			change.Change = sysupdateservice.ChangeSettings
			change.Reason = reason + "; its settings are brought to this release's, which restarts it"
			change.InterruptsTraffic = key == base.HivepaasTraefikKey
		}
		return change
	}

	change.Change = sysupdateservice.ChangeUpdate
	from, fromOK := imageref.MajorVersion(current)
	to, toOK := imageref.MajorVersion(target)
	if fromOK && toOK && from != to {
		switch {
		case slices.Contains(blockMajor, key):
			change.Change = sysupdateservice.ChangeBlocked
			change.Reason = "this release moves it across a major version, " +
				"which needs a data migration the update does not perform"
		default:
			change.Change = sysupdateservice.ChangeMajor
			// The new postgres cluster starts empty and is loaded from the dump.
			change.RequiresBackup = key == base.HivepaasDbKey
		}
	}
	change.InterruptsTraffic = key == base.HivepaasTraefikKey
	return change
}
