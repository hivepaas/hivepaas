package entity

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

type TaskLog struct {
	ID       int64           `bun:",pk,autoincrement" json:"id"`
	TaskID   string          `json:"taskId"`
	TargetID string          `bun:",nullzero" json:"targetId,omitempty"`
	Type     tasklog.LogType `bun:",nullzero" json:"type"`
	Data     string          `json:"data"`
	Ts       time.Time       `bun:",nullzero" json:"ts"`
}

func (t *TaskLog) GetID() int64 {
	return t.ID
}
