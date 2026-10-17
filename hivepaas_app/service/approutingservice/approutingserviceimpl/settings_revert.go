package approutingserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
)

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

	err = s.appService.PersistAppData(ctx, db, &appservice.PersistingAppData{
		UpsertingSettings: []*entity.Setting{setting},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// SkipMissingRefObjects, unlike the apply path, which refuses a change that
	// references a setting that is gone. Here refusing means staying on the
	// configuration that locked somebody out, so the settings are restored as far
	// as they can be and a missing reference is dropped rather than being made
	// fatal.
	//
	// It has to be said here rather than by pre-loading the references leniently:
	// the apply reloads them itself, and it re-queries exactly the ones a lenient
	// load left out.
	resp, err := s.ApplyRoutingSettings(ctx, db, revertApplyReq(req.App, routingSettings))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &approutingservice.RevertSettingsResp{Reverted: true, Service: resp.Service}, nil
}

func findSettingByID(app *entity.App, settingID string) *entity.Setting {
	for _, setting := range app.Settings {
		if setting.ID == settingID {
			return setting
		}
	}
	return nil
}

// revertApplyReq is the request the routing revert makes.
//
// The update rolls back when it fails, as the change it undoes did. It is the
// apply's own update that carries those terms, rather than a second one made
// afterwards, so that it is retried with the rest when swarm refuses it as out
// of sequence.
func revertApplyReq(
	app *entity.App,
	routingSettings *entity.AppRoutingSettings,
) *approutingservice.ApplyAppRoutingReq {
	return &approutingservice.ApplyAppRoutingReq{
		App:                   app,
		RoutingSettings:       routingSettings,
		RollbackOnFailure:     true,
		SkipMissingRefObjects: true,
	}
}
