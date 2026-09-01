package backupreposerviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
)

func (s *service) ListSnapshots(
	ctx context.Context,
	db database.IDB,
	req *backupreposervice.ListSnapshotsReq,
) (resp *backupreposervice.ListSnapshotsResp, err error) {
	engine, err := s.buildEngine(ctx, db, req.Scope, req.Repo, req.RepoID, req.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	snapshots, err := readSnapshots(ctx, engine, req.Options)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &backupreposervice.ListSnapshotsResp{Snapshots: snapshots}, nil
}
