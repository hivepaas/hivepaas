package cacheentity

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

type TaskInfo struct {
	ID              string          `json:"id"`
	Status          base.TaskStatus `json:"status"`
	ControlDisabled bool            `json:"controlDisabled,omitempty"`
	StartedAt       time.Time       `json:"startedAt"`
}
