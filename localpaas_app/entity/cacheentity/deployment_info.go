package cacheentity

import (
	"time"

	"github.com/localpaas/localpaas/localpaas_app/base"
)

type DeploymentInfo struct {
	ID        string                `json:"id"`
	AppID     string                `json:"appID"`
	TaskID    string                `json:"taskID"`
	Status    base.DeploymentStatus `json:"status"`
	StartedAt time.Time             `json:"startedAt"`
}
