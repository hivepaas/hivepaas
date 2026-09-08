package traefiksettingsuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/traefiksettingsuc/traefiksettingsdto"
)

func (uc *UC) GetConfigOptions(
	ctx context.Context,
	auth *basedto.Auth,
	req *traefiksettingsdto.GetConfigOptionsReq,
) (*traefiksettingsdto.GetConfigOptionsResp, error) {
	traefikSvc, err := uc.traefikService.GetTraefikSwarmService(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData := traefiksettingsdto.TransformConfigOptions(traefikSvc)

	// The countdown, repeated. Applying these options replaces traefik's task, so
	// the response to the update that started the trial is sent down a connection
	// that is about to be cut; a dashboard that lost it, or was simply reloaded,
	// has no other way to find out there is something to confirm.
	pending, err := uc.pendingConfigOptionsChange(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	respData.PendingChange = traefiksettingsdto.TransformPendingChange(pending)

	return &traefiksettingsdto.GetConfigOptionsResp{
		Data: respData,
	}, nil
}

// pendingConfigOptionsChange is best effort on the app lookup only.
//
// An install whose traefik app row is missing has bigger problems, but they are
// not this endpoint's to report: the options themselves read fine from the swarm
// service, and failing the whole GET would leave the operator without the page
// they need to fix anything.
func (uc *UC) pendingConfigOptionsChange(ctx context.Context) (*entity.Task, error) {
	appID, err := uc.traefikAppID(ctx, uc.db)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return nil, nil //nolint:nilnil // no app means no trial
		}
		return nil, hperrors.Wrap(err)
	}

	task, err := uc.probationService.FindPending(ctx, uc.db, appID, base.SettingTypeTraefikConfig)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return task, nil
}
