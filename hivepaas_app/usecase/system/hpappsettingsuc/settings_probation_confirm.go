package hpappsettingsuc

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingsprobationservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc/hpappsettingsdto"
)

const (
	appLabelsSweepMaxRetry   = 3
	appLabelsSweepRetryDelay = 30 * time.Second
)

// ConfirmRoutingSettings vouches for a routing change, which is what stops it
// from being undone.
func (uc *UC) ConfirmRoutingSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *hpappsettingsdto.ConfirmRoutingSettingsReq,
) (*hpappsettingsdto.ConfirmRoutingSettingsResp, error) {
	if err := uc.confirmSettingsChange(ctx, auth, base.SettingTypeAppRouting, req.ChangeID); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &hpappsettingsdto.ConfirmRoutingSettingsResp{}, nil
}

// RevertRoutingSettings undoes the change now instead of waiting for its deadline.
func (uc *UC) RevertRoutingSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *hpappsettingsdto.RevertRoutingSettingsReq,
) (*hpappsettingsdto.RevertRoutingSettingsResp, error) {
	output, err := uc.revertSettingsChange(ctx, auth, base.SettingTypeAppRouting, req.ChangeID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &hpappsettingsdto.RevertRoutingSettingsResp{Data: transformRevertOutput(output)}, nil
}

// ConfirmServiceSettings vouches for a HivePaaS service settings change.
func (uc *UC) ConfirmServiceSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *hpappsettingsdto.ConfirmServiceSettingsReq,
) (*hpappsettingsdto.ConfirmServiceSettingsResp, error) {
	if err := uc.confirmSettingsChange(ctx, auth, base.SettingTypeHivePaaSService, req.ChangeID); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &hpappsettingsdto.ConfirmServiceSettingsResp{}, nil
}

// RevertServiceSettings undoes an unconfirmed service settings change now.
func (uc *UC) RevertServiceSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *hpappsettingsdto.RevertServiceSettingsReq,
) (*hpappsettingsdto.RevertServiceSettingsResp, error) {
	output, err := uc.revertSettingsChange(ctx, auth, base.SettingTypeHivePaaSService, req.ChangeID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &hpappsettingsdto.RevertServiceSettingsResp{Data: transformRevertOutput(output)}, nil
}

func transformRevertOutput(output *entity.TaskSettingsRevertOutput) *hpappsettingsdto.RevertSettingsDataResp {
	resp := &hpappsettingsdto.RevertSettingsDataResp{}
	if output != nil {
		resp.Reverted = output.Reverted
		resp.Reason = output.Reason
	}
	return resp
}

func (uc *UC) confirmSettingsChange(
	ctx context.Context,
	auth *basedto.Auth,
	settingType base.SettingType,
	changeID string,
) error {
	appID, err := uc.hivePaaSAppID(ctx, uc.db)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return hperrors.Wrap(uc.probationService.Confirm(ctx, auth, &settingsprobationservice.AnswerReq{
		AppID:           appID,
		SettingType:     settingType,
		ChangeID:        changeID,
		OnConfirmed:     uc.appLabelsSweepOnConfirm,
		EnsureStillLive: uc.ensureProxySettingsAreLive,
	}))
}

func (uc *UC) revertSettingsChange(
	ctx context.Context,
	auth *basedto.Auth,
	settingType base.SettingType,
	changeID string,
) (*entity.TaskSettingsRevertOutput, error) {
	appID, err := uc.hivePaaSAppID(ctx, uc.db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	output, err := uc.probationService.RevertNow(ctx, auth, &settingsprobationservice.AnswerReq{
		AppID:       appID,
		SettingType: settingType,
		ChangeID:    changeID,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return output, nil
}

// appLabelsSweepOnConfirm records the fan-out the trial deferred.
//
// Only for the HivePaaS service settings: those are the ones whose proxy topology
// every other app's labels are derived from, and while the change was on trial
// only the HivePaaS app was given the new one. A routing change is one app's
// labels and has nothing to spread; a traefik command change is not in any app's
// labels at all.
//
// It runs inside the confirmation's transaction, so a confirmation that commits
// always has its sweep committed with it. See settingsprobationservice.AnswerReq.
func (uc *UC) appLabelsSweepOnConfirm(
	ctx context.Context,
	db database.Tx,
	args *entity.TaskSettingsRevertArgs,
) ([]*entity.Task, error) {
	if args.SettingType != base.SettingTypeHivePaaSService {
		return nil, nil
	}

	timeNow := timeutil.NowUTC()
	task := &entity.Task{
		ID:       gofn.Must(ulid.NewStringULID()),
		Scope:    base.ObjectScopeApp,
		ObjectID: args.AppID,
		Type:     base.TaskTypeAppLabelsSweep,
		Status:   base.TaskStatusNotStarted,
		Config: entity.TaskConfig{
			Priority:   base.TaskPriorityCritical,
			MaxRetry:   appLabelsSweepMaxRetry,
			RetryDelay: timeutil.Duration(appLabelsSweepRetryDelay),
		},
		Version:   entity.CurrentTaskVersion,
		RunAt:     timeNow,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	if err := task.SetArgs(&entity.TaskAppLabelsSweepArgs{AppID: args.AppID}); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err := uc.taskRepo.Insert(ctx, db, task); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return []*entity.Task{task}, nil
}

// ensureProxySettingsAreLive refuses a confirmation of proxy settings traefik is
// not running.
//
// Only the proxy settings, and only because of what applying them does: they are
// written onto traefik's web entrypoints as forwardedheaders.trustedips, which is
// a container spec change, which swarm answers by replacing the task. A task that
// never becomes healthy is rolled back - failure_action: rollback on that service
// - and the previous trusted IPs come back with it, quietly, around two and a
// half minutes in. That is inside the confirmation window, and by then traefik is
// serving again and the dashboard looks entirely healthy.
//
// Confirming there would cancel the trial with the setting row saying one thing
// and traefik running another. Refusing lets the deadline undo the row, which is
// what puts them back in agreement.
//
// Nothing to check for a routing change: it rewrites swarm service labels, which
// does not recreate a task, so there is no failed update for swarm to roll back.
func (uc *UC) ensureProxySettingsAreLive(
	ctx context.Context,
	_ database.Tx,
	args *entity.TaskSettingsRevertArgs,
) error {
	if args.SettingType != base.SettingTypeHivePaaSService {
		return nil
	}

	// pendingFor has already checked this row still carries the trial's version,
	// so it is the change on trial and not a later one.
	setting, err := uc.settingRepo.GetSingle(ctx, uc.db, nil, base.SettingTypeHivePaaSService, true)
	if err != nil {
		return hperrors.Wrap(err)
	}
	applied, err := setting.AsHivePaaSService()
	if err != nil {
		return hperrors.Wrap(err)
	}

	live, err := uc.traefikService.WebEntrypointsCarryTrustedIPs(ctx, applied.ProxySettings.TrustedIPs)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if !live {
		return hperrors.Wrap(hperrors.ErrSettingsChangeNotLive).
			WithMsgLog("traefik is not running the trusted IPs on trial, " +
				"most likely because swarm rolled the update back")
	}
	return nil
}
