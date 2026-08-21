package clustersecretserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

func (s *service) UpdateConfigForApp(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	oldConfig, newConfig *entity.ConfigFile,
) (err error) {
	// Remove the old config from services then delete it from the swarm
	err = s.RemoveConfigForApp(ctx, db, app, oldConfig)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Create a config in the swarm then add it to the services
	_, err = s.CreateConfigForApp(ctx, db, app, newConfig)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
