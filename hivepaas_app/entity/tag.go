package entity

import (
	"time"
)

var (
	TagUpsertingConflictCols = []string{"object_id", "tag"}
	TagUpsertingUpdateCols   = []string{"index", "deleted_at"}
)

type Tag struct {
	ObjectID string `bun:",pk" json:"objectId"`
	Tag      string `bun:",pk" json:"tag"`
	Index    int    `json:"index"`

	DeletedAt time.Time `bun:",soft_delete,nullzero" json:"deletedAt,omitzero"`
}
