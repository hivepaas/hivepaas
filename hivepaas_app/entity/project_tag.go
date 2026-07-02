package entity

import (
	"time"
)

var (
	ProjectTagUpsertingConflictCols = []string{"project_id", "tag"}
	ProjectTagUpsertingUpdateCols   = []string{"display_order", "deleted_at"}
)

type ProjectTag struct {
	ProjectID    string `bun:",pk" json:"projectId"`
	Tag          string `bun:",pk" json:"tag"`
	DisplayOrder int    `json:"displayOrder"`

	DeletedAt time.Time `bun:",soft_delete,nullzero" json:"deletedAt,omitzero"`
}
