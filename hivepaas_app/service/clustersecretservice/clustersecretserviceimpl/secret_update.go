package clustersecretserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

func (s *service) UpdateSecretForApp(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	oldSecret, newSecret *entity.Secret,
) (err error) {
	// Remove the old secret from services then delete it from the swarm
	err = s.RemoveSecretForApp(ctx, db, app, oldSecret)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Create a secret in the swarm then add it to the services
	_, err = s.CreateSecretForApp(ctx, db, app, newSecret)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
