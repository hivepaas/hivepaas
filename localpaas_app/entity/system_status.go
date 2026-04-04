package entity

import (
	"time"
)

var (
	SystemStatusUpsertingConflictCols = []string{"id"}
	SystemStatusUpsertingUpdateCols   = []string{"installation_complete", "update_ver", "updated_at"}
)

type SystemStatus struct {
	ID                   string `bun:",pk" json:"id"`
	InstallationComplete bool   `json:"installationComplete"`
	UpdateVer            int    `json:"updateVer"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
}
