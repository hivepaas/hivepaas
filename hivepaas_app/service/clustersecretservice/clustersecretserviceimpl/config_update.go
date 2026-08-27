package clustersecretserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

func (s *service) UpdateConfigForApp(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	oldConfig, newConfig *entity.ConfigFile,
) (err error) {
	if !s.HasConfigChanges(oldConfig, newConfig) {
		return nil
	}

	// Remove the old config from the service then delete it from the swarm
	err = s.RemoveConfigForApp(ctx, db, app, oldConfig)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Create a config in the swarm then add it to the service
	_, err = s.CreateConfigForApp(ctx, db, app, newConfig)
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}

func (s *service) UpdateConfigsForApp(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	oldConfigs, newConfigs []*entity.ConfigFile,
) (err error) {
	if len(oldConfigs) != len(newConfigs) {
		return hperrors.Wrap(hperrors.ErrArgumentInvalid).WithParam("Name", "Slice length")
	}

	removingConfigs := make([]*entity.ConfigFile, 0, len(oldConfigs))
	creatingConfigs := make([]*entity.ConfigFile, 0, len(newConfigs))
	for i, oldConfig := range oldConfigs {
		newConfig := newConfigs[i]
		if !s.HasConfigChanges(oldConfig, newConfig) {
			continue
		}
		if oldConfig != nil {
			removingConfigs = append(removingConfigs, oldConfig)
		}
		if newConfig != nil {
			creatingConfigs = append(creatingConfigs, newConfig)
		}
	}

	// Remove the old configs from the service then delete them from the swarm
	if len(removingConfigs) > 0 {
		err = s.RemoveConfigForApp(ctx, db, app, removingConfigs...)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	// Create configs in the swarm then add them to the service
	if len(creatingConfigs) > 0 {
		_, err = s.CreateConfigsForApp(ctx, db, app, creatingConfigs)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	return nil
}
