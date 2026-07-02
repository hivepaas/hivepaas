package entity

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

var (
	SystemStatusUpsertingConflictCols = []string{"id"}
	SystemStatusUpsertingUpdateCols   = []string{"next_step", "update_ver", "updated_at"}
)

type SystemStatus struct {
	ID        string                `bun:",pk" json:"id"`
	NextStep  base.InstallationStep `json:"nextStep"`
	UpdateVer int                   `json:"updateVer"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
}

func (s *SystemStatus) HasNextStep() bool {
	return s.NextStep != "" && s.NextStep != base.InstallationStepNone
}
