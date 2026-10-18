// Package appmanageduc is what a person does with a managed app as a whole: its
// engine's state and settings, restoring it, and unmanaging it. See
// docs/superpowers/specs/2026-10-17-managed-templates-design.md.
package appmanageduc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/artifactservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc"
)

// auditSectionManaged is the section a managed app's changes are audited under.
const auditSectionManaged = "managed"

type UC struct {
	db        *database.DB
	taskQueue queue.TaskQueue

	settingRepo repository.SettingRepo
	taskRepo    repository.TaskRepo
	// appRepo lists an environment's apps, for the consumers a switch repoints.
	appRepo repository.AppRepo

	appService     appservice.Service
	auditService   auditservice.Service
	managedService managedservice.Service
	// settingMountService mounts again the secrets a refresh changed.
	settingMountService settingmountservice.Service
	// appProvisionService, artifactService, licenseService and volumeService make
	// a verifier app.
	appProvisionService appprovisionservice.Service
	artifactService     artifactservice.Service
	licenseService      licenseservice.Service
	volumeService       volumeservice.Service
	// appTemplateService and appTemplateUC make a restored app the way its
	// source was made.
	appTemplateService apptemplateservice.Service
	appTemplateUC      *apptemplateuc.UC
	// appDeploymentService and envVarService deploy a switch's consumers, or
	// apply their variables.
	appDeploymentService appdeploymentservice.Service
	envVarService        envvarservice.Service
}

func New(
	db *database.DB,
	taskQueue queue.TaskQueue,

	settingRepo repository.SettingRepo,
	taskRepo repository.TaskRepo,

	appService appservice.Service,
	auditService auditservice.Service,
	managedService managedservice.Service,
	settingMountService settingmountservice.Service,
	appProvisionService appprovisionservice.Service,
	artifactService artifactservice.Service,
	licenseService licenseservice.Service,
	volumeService volumeservice.Service,
	appTemplateService apptemplateservice.Service,
	appTemplateUC *apptemplateuc.UC,
	appRepo repository.AppRepo,
	appDeploymentService appdeploymentservice.Service,
	envVarService envvarservice.Service,
) *UC {
	return &UC{
		db:        db,
		taskQueue: taskQueue,

		settingRepo: settingRepo,
		taskRepo:    taskRepo,

		appService:     appService,
		auditService:   auditService,
		managedService: managedService,

		settingMountService:  settingMountService,
		appProvisionService:  appProvisionService,
		artifactService:      artifactService,
		licenseService:       licenseService,
		volumeService:        volumeService,
		appTemplateService:   appTemplateService,
		appRepo:              appRepo,
		appDeploymentService: appDeploymentService,
		envVarService:        envVarService,
		appTemplateUC:        appTemplateUC,
	}
}
