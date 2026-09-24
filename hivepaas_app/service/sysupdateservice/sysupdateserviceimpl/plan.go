package sysupdateserviceimpl

import (
	"context"
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
}

func (s *service) PlanUpdate(
	ctx context.Context,
	db database.IDB,
	target *base.ReleaseInfo,
) (*sysupdateservice.UpdatePlan, error) {
	steps := []planStep{
		{key: base.HivepaasDbKey, image: target.DbImage, fetch: s.hpAppService.GetHpDbSwarmService},
		{key: base.HivepaasCacheKey, image: target.RedisImage, fetch: s.hpAppService.GetHpCacheSwarmService},
		{key: base.HivepaasTraefikKey, image: target.TraefikImage, fetch: s.traefikService.GetTraefikSwarmService},
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
		planStep{key: base.HivepaasAppKey, image: target.AppImage, fetch: s.hpAppService.GetHpAppSwarmService},
		planStep{key: base.HivepaasWorkerKey, image: target.AppImage, fetch: s.getWorkerSwarmService},
	)

	plan := &sysupdateservice.UpdatePlan{}
	for _, step := range steps {
		svc, err := step.fetch(ctx)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		current, deployed := "", svc != nil
		if deployed {
			current = svc.Spec.TaskTemplate.ContainerSpec.Image
		}
		change := planComponent(step.key, current, deployed, step.image, target.BlockMajorUpgrade)
		plan.Components = append(plan.Components, change)
		plan.Blocked = plan.Blocked || change.Change == sysupdateservice.ChangeBlocked
		plan.RequiresBackup = plan.RequiresBackup || change.RequiresBackup
	}
	return plan, nil
}

// planComponent decides what an update does to one component, by the rules
// updateServiceImage applies: imageref.IsUpgrade for whether to move at all,
// and the release's BlockMajorUpgrade for a move across a major.
func planComponent(
	key, current string,
	deployed bool,
	target string,
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
