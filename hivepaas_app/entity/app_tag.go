package entity

import (
	"time"
)

var (
	AppTagUpsertingConflictCols = []string{"app_id", "tag"}
	AppTagUpsertingUpdateCols   = []string{"display_order", "deleted_at"}
)

type AppTag struct {
	AppID        string `bun:",pk" json:"appId"`
	Tag          string `bun:",pk" json:"tag"`
	DisplayOrder int    `json:"displayOrder"`

	DeletedAt time.Time `bun:",soft_delete,nullzero" json:"deletedAt,omitzero"`
}
