package systemappserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	projectRepo    repository.ProjectRepo
	projectEnvRepo repository.ProjectEnvRepo
	settingRepo    repository.SettingRepo

	appService        appservice.Service
	hpAppService      hpappservice.Service
	provisionService  appprovisionservice.Service
	specService       specservice.Service
	deploymentService appdeploymentservice.Service

	dockerManager docker.Manager
}

// New builds the system app service. fx wires the arguments from the provider
// list in registry/provides.go; adding a parameter here needs no other change.
//
//nolint:ireturn // the constructor of a service returns its interface
func New(
	projectRepo repository.ProjectRepo,
	projectEnvRepo repository.ProjectEnvRepo,
	settingRepo repository.SettingRepo,

	appService appservice.Service,
	hpAppService hpappservice.Service,
	provisionService appprovisionservice.Service,
	specService specservice.Service,
	deploymentService appdeploymentservice.Service,

	dockerManager docker.Manager,
) systemappservice.Service {
	return &service{
		projectRepo:       projectRepo,
		projectEnvRepo:    projectEnvRepo,
		settingRepo:       settingRepo,
		appService:        appService,
		hpAppService:      hpAppService,
		provisionService:  provisionService,
		specService:       specService,
		deploymentService: deploymentService,
		dockerManager:     dockerManager,
	}
}

func (s *service) LoadApp(ctx context.Context, db database.IDB, key string) (*entity.App, error) {
	app, err := s.hpAppService.LoadAppByKey(ctx, db, key,
		bunex.SelectRelation("Settings"),
		bunex.SelectRelation("Project"),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return nil, nil
		}
		return nil, hperrors.Wrap(err)
	}
	return app, nil
}

func (s *service) Remove(ctx context.Context, db database.IDB, app *entity.App, removeStorage bool) error {
	return hperrors.Wrap(s.appService.DeleteApp(ctx, db, app, removeStorage, true))
}
