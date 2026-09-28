package backupsnapshotuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backupsnapshotuc/backupsnapshotdto"
)

func (uc *UC) GetBackupSnapshot(
	ctx context.Context,
	auth *basedto.Auth,
	req *backupsnapshotdto.GetBackupSnapshotReq,
) (*backupsnapshotdto.GetBackupSnapshotResp, error) {
	db := uc.DB
	if err := uc.ScopeService.LoadObjectScopeData(ctx, db, req.Scope); err != nil {
		return nil, hperrors.Wrap(err)
	}
	record, repos, err := uc.findInReach(ctx, db, auth, req.Scope, req.ID, base.ActionTypeRead)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	tags, err := uc.tagsOf(ctx, db, []*entity.Setting{record})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	refs, err := uc.snapshotRefs(ctx, db, repos, tags)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &backupsnapshotdto.GetBackupSnapshotResp{
		Data: backupsnapshotdto.TransformBackupSnapshot(record, tags[record.ID], refs),
	}, nil
}

// findInReach is a snapshot of the view's reach, for a viewer acting with action;
// not found when it is outside, as a snapshot that does not exist is.
func (uc *UC) findInReach(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	scope *entity.ObjectScope,
	id string,
	action base.ActionType,
) (*entity.Setting, []*entity.Setting, error) {
	reach, repos, err := uc.reachOf(ctx, db, auth, scope, action)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	records, _, err := uc.SettingRepo.List(ctx, db, nil, nil, snapshotQueryOpts(reach, &snapshotFilter{ID: id})...)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	if len(records) == 0 {
		return nil, nil, hperrors.NewNotFound("Backup snapshot")
	}
	return records[0], repos, nil
}
