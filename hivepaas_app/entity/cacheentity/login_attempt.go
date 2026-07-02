package cacheentity

import "time"

type LoginAttempt struct {
	Fails       int       `json:"fails"`
	FirstFailAt time.Time `json:"firstFailAt"`
}
