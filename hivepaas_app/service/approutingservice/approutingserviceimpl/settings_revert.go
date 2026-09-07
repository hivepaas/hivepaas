package approutingserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
)

// updateMaxFailureRatio matches what every other swarm update in the codebase
// sets, so a revert converges on the same terms as the change it undoes.
const updateMaxFailureRatio = 0.5

// RevertSettings implements approutingservice.Service
func (s *service) RevertSettings(
	ctx context.Context,
	db database.Tx,
	req *approutingservice.RevertSettingsReq,
) (*approutingservice.RevertSettingsResp, error) {
	setting := findSettingByID(req.App, req.SettingID)
	if setting == nil {
		return &approutingservice.RevertSettingsResp{Reason: "setting no longer exists"}, nil
	}
	// The change under probation is not the change that is live any more, so
	// there is nothing here to undo - see RevertSettingsReq.ProbationVer.
	if setting.UpdateVer != req.ProbationVer {
		return &approutingservice.RevertSettingsResp{Reason: "settings changed since"}, nil
	}
	if req.Snapshot.Data == "" {
		return nil, hperrors.Wrap(hperrors.ErrInternal).
			WithMsgLog("routing revert has no snapshot to restore, setting %s", req.SettingID)
	}

	req.Snapshot.RestoreTo(setting)
	setting.UpdateVer++
	setting.UpdatedAt = timeutil.NowUTC()

	routingSettings, err := setting.AsAppRoutingSettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// SkipMissing, unlike the apply path, which refuses a change that references a
	// setting that is gone. Here refusing means staying on the configuration that
	// locked somebody out, so the settings are restored as far as they can be and
	// a missing reference is dropped rather than being made fatal.
	var refObjects *entity.RefObjects
	err = s.settingService.LoadRefObjectsByIDsSkipMissing(ctx, db, &refObjects,
		req.App.GetObjectScope(), true, routingSettings.GetRefObjectIDs())
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = s.appService.PersistAppData(ctx, db, &appservice.PersistingAppData{
		UpsertingSettings: []*entity.Setting{setting},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := s.ApplyRoutingSettings(ctx, db, &approutingservice.ApplyAppRoutingReq{
		App:                 req.App,
		RoutingSettings:     routingSettings,
		RefObjects:          refObjects,
		SkipUpdatingService: true,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	service := resp.Service
	if service.Spec.UpdateConfig == nil {
		service.Spec.UpdateConfig = &swarm.UpdateConfig{}
	}
	service.Spec.UpdateConfig.FailureAction = swarm.UpdateFailureActionRollback
	service.Spec.UpdateConfig.MaxFailureRatio = updateMaxFailureRatio

	_, err = s.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &approutingservice.RevertSettingsResp{Reverted: true, Service: service}, nil
}

func findSettingByID(app *entity.App, settingID string) *entity.Setting {
	for _, setting := range app.Settings {
		if setting.ID == settingID {
			return setting
		}
	}
	return nil
}
