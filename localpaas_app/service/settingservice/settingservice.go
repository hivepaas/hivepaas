package settingservice

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/service/sslservice"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
	"github.com/localpaas/localpaas/services/docker"
)

type SettingService interface {
	PersistSettingData(ctx context.Context, db database.IDB, data *PersistingSettingData) error

	PersistSSLCertFiles(forceRecreate bool, settings ...*entity.Setting) error
	DeleteSSLCertFiles(settings ...*entity.Setting) error

	LoadReferenceObjects(ctx context.Context, db database.IDB, scope *base.SettingScope, requireActive bool,
		errorIfUnavail bool, inSettings ...*entity.Setting) (*entity.RefObjects, error)
	LoadReferenceObjectsByIDs(ctx context.Context, db database.IDB, scope *base.SettingScope, requireActive bool,
		errorIfUnavail bool, refIDs *entity.RefObjectIDs) (*entity.RefObjects, error)

	// Default settings
	InitDefaults(ctx context.Context, db database.IDB) error

	// Events
	OnCreate(ctx context.Context, db database.IDB, event *CreateEvent) error
	OnUpdate(ctx context.Context, db database.IDB, event *UpdateEvent) error
	OnUpdateMeta(ctx context.Context, db database.IDB, event *UpdateEvent) error
	OnDelete(ctx context.Context, db database.IDB, event *DeleteEvent) error
}

func NewSettingService(
	db *database.DB,
	settingRepo repository.SettingRepo,
	appRepo repository.AppRepo,
	healthcheckSettingsRepo cacherepository.HealthcheckSettingsRepo,
	userService userservice.UserService,
	taskService taskservice.TaskService,
	sslService sslservice.SSLService,
	permissionManager permission.Manager,
	dockerManager docker.Manager,
) SettingService {
	return &settingService{
		db:                      db,
		settingRepo:             settingRepo,
		appRepo:                 appRepo,
		healthcheckSettingsRepo: healthcheckSettingsRepo,
		userService:             userService,
		taskService:             taskService,
		sslService:              sslService,
		permissionManager:       permissionManager,
		dockerManager:           dockerManager,
	}
}

type settingService struct {
	db                      *database.DB
	settingRepo             repository.SettingRepo
	appRepo                 repository.AppRepo
	healthcheckSettingsRepo cacherepository.HealthcheckSettingsRepo
	userService             userservice.UserService
	taskService             taskservice.TaskService
	sslService              sslservice.SSLService
	permissionManager       permission.Manager
	dockerManager           docker.Manager
}
