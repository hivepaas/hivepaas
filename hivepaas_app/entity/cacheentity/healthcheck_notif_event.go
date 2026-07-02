package cacheentity

import (
	"time"
)

type HealthcheckNotifEvent struct {
	Event      string    `json:"event"`
	LastSendTs time.Time `json:"lastSendTs"`
}
