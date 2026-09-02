package backupreposerviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
)

func (s *service) ChangeRepoPassword(
	ctx context.Context,
	db database.IDB,
	req *backupreposervice.ChangeRepoPasswordReq,
) error {
	engine, err := s.buildEngineWithPassword(ctx, db, req.Scope, req.Repo, req.RepoID,
		req.RefObjects, req.OldPassword)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// The engine holds the connection state in a config file that a freshly started container has
	// never written, and it refuses to change the password of a repository it is not connected to.
	if err := engine.ConnectRepo(ctx); err != nil {
		return hperrors.Wrap(err)
	}

	if err := engine.ChangePassword(ctx, req.OldPassword, req.NewPassword); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
