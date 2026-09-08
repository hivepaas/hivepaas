package settingsrevertserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingsrevertservice"
)

type reverter func(ctx context.Context, db database.Tx,
	args *entity.TaskSettingsRevertArgs) (*settingsrevertservice.RevertResp, error)

// Revert implements settingsrevertservice.Service
func (s *service) Revert(
	ctx context.Context,
	db database.Tx,
	args *entity.TaskSettingsRevertArgs,
) (*settingsrevertservice.RevertResp, error) {
	if args == nil {
		return nil, hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("settings revert has no args")
	}
	revert, known := s.reverters[args.SettingType]
	if !known {
		return nil, hperrors.Wrap(hperrors.ErrInternal).
			WithMsgLog("no reverter for setting type '%v'", args.SettingType)
	}
	// An empty snapshot would restore an empty payload, which for either setting
	// type means a HivePaaS with no way in at all - strictly worse than the change
	// this is trying to undo.
	if args.Snapshot.Data == "" {
		return nil, hperrors.Wrap(hperrors.ErrInternal).
			WithMsgLog("settings revert for '%v' has no snapshot to restore", args.SettingType)
	}

	resp, err := revert(ctx, db, args)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp, nil
}

// revertAppRouting puts the app's routing settings back.
//
// The apply half is the same one the update path uses, so a revert lands the
// labels exactly the way applying those settings would have.
func (s *service) revertAppRouting(
	ctx context.Context,
	db database.Tx,
	args *entity.TaskSettingsRevertArgs,
) (*settingsrevertservice.RevertResp, error) {
	app, err := s.hpAppService.LoadAppByKey(ctx, db, base.HivepaasAppKey,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", args.SettingType),
		),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := s.appRoutingService.RevertSettings(ctx, db, &approutingservice.RevertSettingsReq{
		App:          app,
		SettingID:    args.SettingID,
		ProbationVer: args.ProbationVer,
		Snapshot:     args.Snapshot,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &settingsrevertservice.RevertResp{Reverted: resp.Reverted, Reason: resp.Reason}, nil
}
