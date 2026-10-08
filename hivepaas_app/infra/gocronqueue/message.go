package gocronqueue

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

const (
	taskQueueCtrlKey         = "task:queue:ctrl"
	taskQueueCtrlReadTimeout = 5 * time.Second
	// ctrlReadErrorBackoff is how long the listener waits after a read that
	// failed, not one that found the list empty.
	ctrlReadErrorBackoff = 10 * time.Second
)

type Message struct {
	StartScheduler bool `json:"startScheduler,omitempty"`
	StopScheduler  bool `json:"stopScheduler,omitempty"`
	// SentAt is when the message was sent. A stop sent before a process
	// started was meant for the one it replaced, and is ignored.
	SentAt time.Time `json:"sentAt,omitzero"`

	SchedTasks     []*entity.Task `json:"schedTasks,omitempty"`
	UnschedTaskIDs []string       `json:"unschedTaskIds,omitempty"`
}
