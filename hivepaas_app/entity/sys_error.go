package entity

import (
	"time"
)

type SysError struct {
	ID         string `bun:",pk" json:"id"`
	RequestID  string `json:"requestId"`
	Status     int    `json:"status"`
	Code       string `json:"code"`
	Detail     string `json:"detail"`
	Cause      string `json:"cause"`
	DebugLog   string `json:"debugLog"`
	StackTrace string `json:"stackTrace"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
}

// GetID implements IDEntity interface
func (e *SysError) GetID() string {
	return e.ID
}
