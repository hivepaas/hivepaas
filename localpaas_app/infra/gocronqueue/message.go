package gocronqueue

import (
	"time"

	"github.com/localpaas/localpaas/localpaas_app/entity"
)

const (
	taskQueueCtrlKey         = "task:queue:ctrl"
	taskQueueCtrlReadTimeout = 10 * time.Minute
)

type Message struct {
	StartScheduler bool `json:"startScheduler,omitempty"`
	StopScheduler  bool `json:"stopScheduler,omitempty"`

	SchedTasks     []*entity.Task `json:"schedTasks,omitempty"`
	UnschedTaskIDs []string       `json:"unschedTaskIds,omitempty"`
}
