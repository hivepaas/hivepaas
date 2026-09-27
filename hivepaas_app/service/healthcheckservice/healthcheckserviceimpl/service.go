package healthcheckserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/healthcheckservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobtriggerservice"
)

type service struct {
	db *database.DB

	healthcheckStateRepo cacherepository.HealthcheckStateRepo

	notificationService    notificationservice.Service
	schedJobTriggerService schedjobtriggerservice.Service
}

func New(
	db *database.DB,

	healthcheckStateRepo cacherepository.HealthcheckStateRepo,

	notificationService notificationservice.Service,
	schedJobTriggerService schedjobtriggerservice.Service,
) healthcheckservice.Service {
	return &service{
		db: db,

		healthcheckStateRepo: healthcheckStateRepo,

		notificationService:    notificationService,
		schedJobTriggerService: schedJobTriggerService,
	}
}
