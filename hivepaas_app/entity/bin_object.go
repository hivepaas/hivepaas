package entity

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

var (
	BinObjectUpsertingConflictCols = []string{"id"}
	BinObjectUpsertingUpdateCols   = []string{"type", "status", "name", "content_type", "data",
		"updated_at", "deleted_at"}
)

type BinObject struct {
	ID          string               `bun:",pk" json:"id"`
	Type        base.BinObjectType   `json:"type"`
	Status      base.BinObjectStatus `json:"status"`
	Name        string               `bun:",nullzero" json:"name,omitempty"`
	ContentType string               `bun:",nullzero" json:"contentType,omitempty"`
	Data        []byte               `json:"data,omitempty"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
	DeletedAt time.Time `bun:",soft_delete,nullzero" json:"deletedAt,omitzero"`
}

// GetID implements IDEntity interface
func (b *BinObject) GetID() string {
	return b.ID
}
