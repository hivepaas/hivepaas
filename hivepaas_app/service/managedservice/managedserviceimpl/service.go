package managedserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	registry    *registry
	db          *database.DB
	settingRepo repository.SettingRepo
	appService  appservice.Service
	// schedJobExecService runs the engine's commands in the database: a mark.
	schedJobExecService schedjobexecservice.Service
	// dockerManager sees a fenced database's tasks stop.
	dockerManager docker.Manager
}

//nolint:ireturn // the constructor of a service returns its interface
func New(
	db *database.DB,
	settingRepo repository.SettingRepo,
	appService appservice.Service,
	schedJobExecService schedjobexecservice.Service,
	dockerManager docker.Manager,
) managedservice.Service {
	return &service{registry: newRegistry(), db: db, settingRepo: settingRepo, appService: appService,
		schedJobExecService: schedJobExecService, dockerManager: dockerManager}
}

func (s *service) Engine(name templatemodel.ManagedEngine) (managedmodel.Engine, bool) {
	return s.registry.Engine(name)
}
