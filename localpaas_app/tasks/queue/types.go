package queue

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/entity/cacheentity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

type TaskExecData struct {
	Task *entity.Task

	// RefObjects can be used as a cache to store objects
	RefObjects *entity.RefObjects

	NonCancelable bool
	NonRetryable  bool
	Canceled      bool
	Done          bool

	// Callback functions
	OnCommand         func(base.TaskCommand, ...any)
	OnEndTransaction  func()
	OnPostTransaction func()
}

func (t *TaskExecData) IsCanceled() bool {
	return t.Canceled
}

func (t *TaskExecData) IsDone() bool {
	return t.Done
}

func (t *TaskExecData) AddRefObjects(refObjects *entity.RefObjects) {
	if t.RefObjects == nil {
		t.RefObjects = refObjects
	} else {
		t.RefObjects.AddRefObjects(refObjects)
	}
}

type TaskExecFunc func(context.Context, database.Tx, *TaskExecData) error

type HealthcheckExecData struct {
	HealthcheckSetting *entity.Setting
	Healthcheck        *entity.Healthcheck
	Task               *entity.Task
	Project            *entity.Project
	App                *entity.App

	// RefObjects can be used as a cache to store objects
	RefObjects    *entity.RefObjects
	NotifEventMap map[string]*cacheentity.HealthcheckNotifEvent
}

func (t *HealthcheckExecData) AddRefObjects(refObjects *entity.RefObjects) {
	if t.RefObjects == nil {
		t.RefObjects = refObjects
	} else {
		t.RefObjects.AddRefObjects(refObjects)
	}
}

type HealthcheckExecFunc func(context.Context, *HealthcheckExecData) error
