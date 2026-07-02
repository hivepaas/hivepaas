package cacheentity

import "github.com/hivepaas/hivepaas/hivepaas_app/base"

type TaskControl struct {
	ID  string           `json:"id"`
	Cmd base.TaskCommand `json:"cmd"`
}
