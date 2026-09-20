package apptemplateuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type UC struct {
	db        *database.DB
	taskQueue queue.TaskQueue

	appRepo     repository.AppRepo
	settingRepo repository.SettingRepo

	appProvisionService appprovisionservice.Service
	appService          appservice.Service
	appTemplateService  apptemplateservice.Service
	auditService        auditservice.Service
	domainService       domainservice.Service
	specService         specservice.Service
}

func New(
	db *database.DB,
	taskQueue queue.TaskQueue,

	appRepo repository.AppRepo,
	settingRepo repository.SettingRepo,

	appProvisionService appprovisionservice.Service,
	appService appservice.Service,
	appTemplateService apptemplateservice.Service,
	auditService auditservice.Service,
	domainService domainservice.Service,
	specService specservice.Service,
) *UC {
	return &UC{
		db:        db,
		taskQueue: taskQueue,

		appRepo:     appRepo,
		settingRepo: settingRepo,

		appProvisionService: appProvisionService,
		appService:          appService,
		appTemplateService:  appTemplateService,
		auditService:        auditService,
		domainService:       domainService,
		specService:         specService,
	}
}
