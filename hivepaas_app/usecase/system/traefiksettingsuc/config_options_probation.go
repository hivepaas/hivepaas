package traefiksettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingsprobationservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/traefiksettingsuc/traefiksettingsdto"
)

// ConfirmConfigOptions vouches for a traefik command change, which is what stops
// it from being undone.
//
// The proof is the request itself. Every route into HivePaaS goes through
// traefik, so a call that arrives here at all was served by the configuration on
// trial - which is the one thing /ping cannot tell us.
func (uc *UC) ConfirmConfigOptions(
	ctx context.Context,
	auth *basedto.Auth,
	req *traefiksettingsdto.ConfirmConfigOptionsReq,
) (*traefiksettingsdto.ConfirmConfigOptionsResp, error) {
	appID, err := uc.traefikAppID(ctx, uc.db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = uc.probationService.Confirm(ctx, auth, &settingsprobationservice.AnswerReq{
		AppID:           appID,
		SettingType:     base.SettingTypeTraefikConfig,
		ChangeID:        req.ChangeID,
		EnsureStillLive: uc.ensureTraefikRunsTheChange,
	}); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &traefiksettingsdto.ConfirmConfigOptionsResp{}, nil
}

// RevertConfigOptions undoes the change now instead of waiting for its deadline.
//
// For the operator who can still get in and can see the change was wrong. The one
// who cannot get in is served by the deadline, which needs nobody.
func (uc *UC) RevertConfigOptions(
	ctx context.Context,
	auth *basedto.Auth,
	req *traefiksettingsdto.RevertConfigOptionsReq,
) (*traefiksettingsdto.RevertConfigOptionsResp, error) {
	appID, err := uc.traefikAppID(ctx, uc.db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	output, err := uc.probationService.RevertNow(ctx, auth, &settingsprobationservice.AnswerReq{
		AppID:       appID,
		SettingType: base.SettingTypeTraefikConfig,
		ChangeID:    req.ChangeID,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp := &traefiksettingsdto.RevertConfigOptionsData{}
	if output != nil {
		resp.Reverted = output.Reverted
		resp.Reason = output.Reason
	}
	return &traefiksettingsdto.RevertConfigOptionsResp{Data: resp}, nil
}

// ensureTraefikRunsTheChange refuses a confirmation of a command traefik is not
// running.
//
// The case this catches is swarm undoing the change on its own. A command traefik
// cannot start under fails its healthcheck, and failure_action: rollback then
// restores the previous one - about two and a half minutes in, which is inside
// the confirmation window. From the operator's side everything looks fine at that
// point: traefik is serving again, the dashboard loads, the confirm button works.
// It is serving the old command.
//
// Confirming there would cancel the trial and leave the setting row saying one
// thing and the cluster running another, with nothing left to reconcile them.
// Refusing lets the deadline undo the row, which is what puts the two back in
// agreement.
//
// The setting row is the right expectation to compare against: pendingFor has
// already checked it still carries this trial's version, so it is the command
// this trial applied and not a later one.
func (uc *UC) ensureTraefikRunsTheChange(
	ctx context.Context,
	_ database.Tx,
	_ *entity.TaskSettingsRevertArgs,
) error {
	setting, err := uc.settingRepo.GetSingle(ctx, uc.db, nil, base.SettingTypeTraefikConfig, true)
	if err != nil {
		return hperrors.Wrap(err)
	}
	applied, err := setting.AsTraefikConfig()
	if err != nil {
		return hperrors.Wrap(err)
	}

	traefikSvc, err := uc.traefikService.GetTraefikSwarmService(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	var liveArgs []string
	if traefikSvc.Spec.TaskTemplate.ContainerSpec != nil {
		liveArgs = traefikSvc.Spec.TaskTemplate.ContainerSpec.Args
	}

	if !applied.SameArgsAs(liveArgs) {
		return hperrors.Wrap(hperrors.ErrSettingsChangeNotLive).
			WithMsgLog("traefik is running a command other than the one on trial, " +
				"most likely because swarm rolled the update back")
	}
	return nil
}
