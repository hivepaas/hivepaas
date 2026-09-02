package backupreposerviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
)

func (s *service) ApplyRepoOptions(
	ctx context.Context,
	db database.IDB,
	req *backupreposervice.ApplyRepoOptionsReq,
) error {
	if !req.Options.HasData() {
		return nil
	}

	engine, err := s.buildEngine(ctx, db, req.Scope, req.Repo, req.RepoID, req.RefObjects)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// These settings live inside the repository, and the engine refuses to touch a repository it
	// is not connected to. A freshly started container has never written the config file.
	if err := engine.ConnectRepo(ctx); err != nil {
		return hperrors.Wrap(err)
	}

	if err := engine.ApplyRepoOptions(ctx, req.Options); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
