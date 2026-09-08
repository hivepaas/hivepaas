package approutingserviceimpl

import (
	"context"
	"sort"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
)

// ReapplyClientIPStrategy implements approutingservice.Service
//
// The set is filtered by AppRoutingSettings.UsesClientIP: an app with neither a
// rate limit nor an IP allowlist carries no ip-strategy depth in its labels, so
// the proxy topology cannot reach it and there is nothing to rewrite. In practice
// that leaves a handful of apps, not all of them.
//
// Every update here is labels only, which swarm applies without recreating the
// task - nothing restarts, and nothing serving traffic notices.
func (s *service) ReapplyClientIPStrategy(
	ctx context.Context,
	db database.Tx,
	req *approutingservice.ReapplyClientIPStrategyReq,
) (*approutingservice.ReapplyClientIPStrategyResp, error) {
	apps, _, err := s.appRepo.List(ctx, db, "", nil,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppRouting),
		),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp := &approutingservice.ReapplyClientIPStrategyResp{Failed: map[string]string{}}

	for _, app := range sweepTargets(apps, req.PrimaryAppID, req.PrimaryOnly) {
		applied, err := s.reapplyAppClientIPStrategy(ctx, db, app, req.SkipMissingRefObjects)
		switch {
		case err != nil && app.ID == req.PrimaryAppID:
			// The trial this runs inside exists to guard reachability of this one
			// app. Carrying on past it would leave the operator confirming a change
			// that never reached the thing they are confirming through.
			return resp, hperrors.Wrap(err)
		case err != nil:
			resp.Failed[app.ID] = err.Error()
		case applied:
			resp.Applied++
		default:
			resp.Skipped++
		}
	}

	return resp, nil
}

// sweepTargets picks the apps to sweep and the order to do it in.
//
// The HivePaaS app goes first, and the rest keep a stable order so a sweep that
// fails partway is repeatable rather than random. primaryOnly drops everything
// else - see ReapplyClientIPStrategyReq.PrimaryOnly.
func sweepTargets(apps []*entity.App, primaryAppID string, primaryOnly bool) []*entity.App {
	if primaryOnly {
		for _, app := range apps {
			if app.ID == primaryAppID {
				return []*entity.App{app}
			}
		}
		return nil
	}

	ordered := make([]*entity.App, len(apps))
	copy(ordered, apps)
	sort.SliceStable(ordered, func(i, j int) bool {
		if (ordered[i].ID == primaryAppID) != (ordered[j].ID == primaryAppID) {
			return ordered[i].ID == primaryAppID
		}
		return ordered[i].ID < ordered[j].ID
	})
	return ordered
}

func (s *service) reapplyAppClientIPStrategy(
	ctx context.Context,
	db database.Tx,
	app *entity.App,
	skipMissingRefObjects bool,
) (applied bool, _ error) {
	setting := app.GetSettingByType(base.SettingTypeAppRouting)
	if setting == nil {
		return false, nil
	}
	routingSettings, err := setting.AsAppRoutingSettings()
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	if !routingSettings.UsesClientIP() {
		return false, nil
	}

	// References are not pre-loaded here: the apply reloads them itself, and would
	// re-query exactly the ones a lenient load had left out. Whether a missing one
	// is fatal is decided by the flag instead - strict on the way forward, lenient
	// when this sweep is running as part of a revert.
	//
	// SkipUpdatingService is not set: the whole point is to push the regenerated
	// labels onto the swarm service.
	_, err = s.ApplyRoutingSettings(ctx, db, sweepApplyReq(app, routingSettings, skipMissingRefObjects))
	if err != nil {
		return false, hperrors.Wrap(err)
	}

	return true, nil
}

// sweepApplyReq is the request the sweep makes for one app.
//
// Split out so the choice it encodes - push the service, leave certs and networks
// alone, and be strict about references unless this is a revert - can be asserted
// without a docker daemon.
func sweepApplyReq(
	app *entity.App,
	routingSettings *entity.AppRoutingSettings,
	skipMissingRefObjects bool,
) *approutingservice.ApplyAppRoutingReq {
	return &approutingservice.ApplyAppRoutingReq{
		App:                   app,
		RoutingSettings:       routingSettings,
		SkipApplyingSslCerts:  true,
		SkipApplyingNetworks:  true,
		SkipMissingRefObjects: skipMissingRefObjects,
	}
}
