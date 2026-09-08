package hpappsettingsuc

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingsprobationservice"
)

// Confirm-or-revert, as the HivePaaS settings pages use it.
//
// The mechanism itself is settingsprobationservice, shared with the traefik
// settings. What stays here is the part that is about these pages: which app the
// trials are filed under, which setting type each endpoint may answer for, and
// what confirming one of them releases.

// probationArgs and probationResult are the service's types under the names the
// call sites in this package already use.
type (
	probationArgs   = settingsprobationservice.ArmReq
	probationResult = settingsprobationservice.ArmResult
)

func resolveProbationWindow(requested, settleDelay time.Duration) time.Duration {
	return settingsprobationservice.ResolveWindow(requested, settleDelay)
}

func (uc *UC) armProbation(
	ctx context.Context,
	db database.Tx,
	auth *basedto.Auth,
	in *probationArgs,
	out *probationResult,
	upsertTask func(task *entity.Task),
) error {
	return hperrors.Wrap(uc.probationService.Arm(ctx, db, auth, in, out, upsertTask))
}

func (uc *UC) scheduleProbation(ctx context.Context, out *probationResult) {
	uc.probationService.Schedule(ctx, out)
}

func (uc *UC) findPendingProbation(
	ctx context.Context,
	db database.IDB,
	appID string,
	settingType base.SettingType,
) (*entity.Task, error) {
	task, err := uc.probationService.FindPending(ctx, db, appID, settingType)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return task, nil
}

// hivePaaSAppID is the object the trials armed from this package are filed under.
func (uc *UC) hivePaaSAppID(ctx context.Context, db database.IDB) (string, error) {
	app, err := uc.hpAppService.LoadAppByKey(ctx, db, base.HivepaasAppKey,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return app.ID, nil
}
