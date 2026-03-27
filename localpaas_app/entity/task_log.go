package entity

import (
	"time"

	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
)

type TaskLog struct {
	ID       int64          `bun:",pk,autoincrement" json:"id"`
	TaskID   string         `json:"taskID"`
	TargetID string         `bun:",nullzero" json:"targetID"`
	Type     applog.LogType `bun:",nullzero" json:"type"`
	Data     string         `json:"data"`
	Ts       time.Time      `bun:",nullzero" json:"ts"`
}

func (t *TaskLog) GetID() int64 {
	return t.ID
}
