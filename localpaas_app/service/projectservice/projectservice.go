package projectservice

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
	"github.com/localpaas/localpaas/localpaas_app/service/networkservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
	"github.com/localpaas/localpaas/services/docker"
)

type ProjectService interface {
	LoadProject(ctx context.Context, db database.IDB, projectID string, requireActive bool,
		extraLoadOpts ...bunex.SelectQueryOption) (*entity.Project, error)

	PersistProjectData(ctx context.Context, db database.IDB, data *PersistingProjectData) error
	DeleteProject(ctx context.Context, project *entity.Project) error

	SaveProjectPhoto(ctx context.Context, project *entity.Project, data []byte, fileExt string) error
}

func NewProjectService(
	projectRepo repository.ProjectRepo,
	appRepo repository.AppRepo,
	projectTagRepo repository.ProjectTagRepo,
	settingRepo repository.SettingRepo,
	permissionManager permission.Manager,
	userService userservice.UserService,
	appService appservice.AppService,
	networkService networkservice.NetworkService,
	dockerManager docker.Manager,
) ProjectService {
	return &projectService{
		projectRepo:       projectRepo,
		appRepo:           appRepo,
		projectTagRepo:    projectTagRepo,
		settingRepo:       settingRepo,
		permissionManager: permissionManager,
		userService:       userService,
		appService:        appService,
		networkService:    networkService,
		dockerManager:     dockerManager,
	}
}

type projectService struct {
	projectRepo       repository.ProjectRepo
	appRepo           repository.AppRepo
	projectTagRepo    repository.ProjectTagRepo
	settingRepo       repository.SettingRepo
	permissionManager permission.Manager
	userService       userservice.UserService
	appService        appservice.AppService
	networkService    networkservice.NetworkService
	dockerManager     docker.Manager
}
