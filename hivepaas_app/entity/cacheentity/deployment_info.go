package cacheentity

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

type DeploymentInfo struct {
	ID        string                `json:"id"`
	AppID     string                `json:"appId"`
	TaskID    string                `json:"taskId"`
	Status    base.DeploymentStatus `json:"status"`
	StartedAt time.Time             `json:"startedAt"`
}
