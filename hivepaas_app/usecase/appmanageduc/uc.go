// Package appmanageduc is what a person does with a managed app as a whole: its
// engine's state and settings, restoring it, and unmanaging it. See
// docs/superpowers/specs/2026-10-17-managed-templates-design.md.
package appmanageduc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type UC struct {
	db        *database.DB
	taskQueue queue.TaskQueue

	settingRepo repository.SettingRepo
	taskRepo    repository.TaskRepo

	appService     appservice.Service
	auditService   auditservice.Service
	managedService managedservice.Service
}

func New(
	db *database.DB,
	taskQueue queue.TaskQueue,

	settingRepo repository.SettingRepo,
	taskRepo repository.TaskRepo,

	appService appservice.Service,
	auditService auditservice.Service,
	managedService managedservice.Service,
) *UC {
	return &UC{
		db:        db,
		taskQueue: taskQueue,

		settingRepo: settingRepo,
		taskRepo:    taskRepo,

		appService:     appService,
		auditService:   auditService,
		managedService: managedService,
	}
}
