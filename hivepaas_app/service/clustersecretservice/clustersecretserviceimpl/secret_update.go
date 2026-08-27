package clustersecretserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

func (s *service) UpdateSecretForApp(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	oldSecret, newSecret *entity.Secret,
) (err error) {
	if !s.HasSecretChanges(oldSecret, newSecret) {
		return nil
	}

	// Remove the old secret from the service then delete it from the swarm
	err = s.RemoveSecretForApp(ctx, db, app, oldSecret)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Create a secret in the swarm then add it to the service
	_, err = s.CreateSecretForApp(ctx, db, app, newSecret)
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}

func (s *service) UpdateSecretsForApp(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	oldSecrets, newSecrets []*entity.Secret,
) (err error) {
	if len(oldSecrets) != len(newSecrets) {
		return hperrors.Wrap(hperrors.ErrArgumentInvalid).WithParam("Name", "Slice length")
	}

	removingSecrets := make([]*entity.Secret, 0, len(oldSecrets))
	creatingSecrets := make([]*entity.Secret, 0, len(newSecrets))
	for i, oldSecret := range oldSecrets {
		newSecret := newSecrets[i]
		if !s.HasSecretChanges(oldSecret, newSecret) {
			continue
		}
		if oldSecret != nil {
			removingSecrets = append(removingSecrets, oldSecret)
		}
		if newSecret != nil {
			creatingSecrets = append(creatingSecrets, newSecret)
		}
	}

	// Remove the old secrets from the service then delete them from the swarm
	if len(removingSecrets) > 0 {
		err = s.RemoveSecretForApp(ctx, db, app, removingSecrets...)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	// Create secrets in the swarm then add them to the service
	if len(creatingSecrets) > 0 {
		_, err = s.CreateSecretsForApp(ctx, db, app, creatingSecrets)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	return nil
}
