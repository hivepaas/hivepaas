package entity

import (
	"time"

	"github.com/localpaas/localpaas/localpaas_app/base"
)

var (
	ResLinkUpsertingConflictCols = []string{"src_type", "src_id", "dst_type", "dst_id"}
	ResLinkUpsertingUpdateCols   = []string{"data", "updated_at", "deleted_at"}
)

type ResLink struct {
	SrcType base.ResourceType `bun:",pk" json:"srcType"`
	SrcID   string            `bun:",pk" json:"srcId"`
	DstType base.ResourceType `bun:",pk" json:"dstType"`
	DstID   string            `bun:",pk" json:"dstId"`
	Data    string            `bun:",nullzero" json:"data"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
	DeletedAt time.Time `bun:",soft_delete,nullzero" json:"deletedAt,omitzero"`

	SrcUser    *User    `bun:"rel:has-one,join:src_id=id" json:"srcUser,omitempty"`
	SrcProject *Project `bun:"rel:has-one,join:src_id=id" json:"srcProject,omitempty"`
	SrcApp     *App     `bun:"rel:has-one,join:src_id=id" json:"srcApp,omitempty"`
}
