package cacheentity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type PeriodicSettings struct {
	Settings   []*entity.Setting  `json:"settings"`
	RefObjects *entity.RefObjects `json:"refObjects"`
}
