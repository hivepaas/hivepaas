package cacheentity

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

type HealthcheckState struct {
	State       base.HealthcheckState `json:"state"`
	LastNotifTs time.Time             `json:"lastNotifTs"`
}
