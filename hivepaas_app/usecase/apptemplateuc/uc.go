package apptemplateuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clustersecretservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type UC struct {
	db        *database.DB
	taskQueue queue.TaskQueue

	settingRepo repository.SettingRepo

	appDeploymentService appdeploymentservice.Service
	appProvisionService  appprovisionservice.Service
	appRoutingService    approutingservice.Service
	appService           appservice.Service
	appTemplateService   apptemplateservice.Service
	auditService         auditservice.Service
	clusterSecretService clustersecretservice.Service
	clusterService       clusterservice.Service
	envVarService        envvarservice.Service
	specService          specservice.Service
}

func New(
	db *database.DB,
	taskQueue queue.TaskQueue,

	settingRepo repository.SettingRepo,

	appDeploymentService appdeploymentservice.Service,
	appProvisionService appprovisionservice.Service,
	appRoutingService approutingservice.Service,
	appService appservice.Service,
	appTemplateService apptemplateservice.Service,
	auditService auditservice.Service,
	clusterSecretService clustersecretservice.Service,
	clusterService clusterservice.Service,
	envVarService envvarservice.Service,
	specService specservice.Service,
) *UC {
	return &UC{
		db:        db,
		taskQueue: taskQueue,

		settingRepo: settingRepo,

		appDeploymentService: appDeploymentService,
		appProvisionService:  appProvisionService,
		appRoutingService:    appRoutingService,
		appService:           appService,
		appTemplateService:   appTemplateService,
		auditService:         auditService,
		clusterSecretService: clusterSecretService,
		clusterService:       clusterService,
		envVarService:        envVarService,
		specService:          specService,
	}
}
