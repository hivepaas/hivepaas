package hpappuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappuc/hpappdto"
)

const (
	lockIDSystemVersionUpdate = "lock:sys:version-update"
)

func (uc *UC) UpdateHpApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *hpappdto.UpdateHpAppReq,
) (*hpappdto.UpdateHpAppResp, error) {
	info, err := uc.hpAppService.GetAppReleaseInfo(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	var target *hpappservice.ReleaseInfo
	switch {
	case info.Stable != nil && info.Stable.AppVersion == req.TargetVersion:
		target = info.Stable
	case info.Beta != nil && info.Beta.AppVersion == req.TargetVersion:
		target = info.Beta
	default:
		return nil, hperrors.Wrap(hperrors.ErrUpdateVerMismatched)
	}
	// Naming a version release.json lists is not enough: a release.json that lists
	// an older one - stale, reverted, or not ours - would otherwise walk the install
	// back onto it, and swarm restores images, never the data a newer one migrated.
	// CanUpdate is the same comparison the dashboard shows the button by, so the
	// API refuses exactly what the UI does not offer.
	if !target.CanUpdate {
		return nil, hperrors.Wrap(hperrors.ErrVersionNotNewer)
	}
	targetVersion := &target.ReleaseInfo

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		_, err := uc.lockRepo.GetByID(ctx, db, lockIDSystemVersionUpdate,
			bunex.SelectFor("UPDATE"),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}
		err = uc.hpAppService.UpdateSystemVersion(ctx, db, targetVersion, req.SkipBackup)
		if err != nil {
			return hperrors.Wrap(err)
		}

		// The version being left behind goes in as well as the one being moved to:
		// after this there is nothing in the install that still says what it was,
		// and "what were we running before this went wrong" is the question an
		// upgrade entry is read to answer. It is taken from the binary serving
		// this request, which is the one actually running.
		return uc.recordHpAppAction(ctx, db, auth, "version-update", auditdetail.New().
			Compare("version", base.StableVersion.AppVersion, targetVersion.AppVersion).
			Set("channel", releaseChannelOf(info, targetVersion)).
			Set("appImage", targetVersion.AppImage))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &hpappdto.UpdateHpAppResp{}, nil
}

// releaseChannelOf names which channel the target came from, so a reader can tell
// a move onto beta from a move along stable.
func releaseChannelOf(info *hpappservice.AppReleaseInfo, target *base.ReleaseInfo) string {
	if info.Beta != nil && info.Beta.AppVersion == target.AppVersion {
		return "beta"
	}
	return "stable"
}
