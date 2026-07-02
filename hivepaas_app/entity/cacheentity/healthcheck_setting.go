package cacheentity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type HealthcheckSettings struct {
	Settings   []*entity.Setting  `json:"settings"`
	RefObjects *entity.RefObjects `json:"refObjects"`
}
