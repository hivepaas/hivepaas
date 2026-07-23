package entity

import (
	"fmt"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

var (
	ResLinkUpsertingConflictCols = []string{"src_type", "src_id", "dst_type", "dst_id"}
	ResLinkUpsertingUpdateCols   = []string{"index", "data", "updated_at", "deleted_at"}
)

type ResLink struct {
	SrcType base.ResourceType `bun:",pk" json:"srcType"`
	SrcID   string            `bun:",pk" json:"srcId"`
	DstType base.ResourceType `bun:",pk" json:"dstType"`
	DstID   string            `bun:",pk" json:"dstId"`
	Index   int               `json:"index"`
	Data    string            `bun:",nullzero" json:"data"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
	DeletedAt time.Time `bun:",soft_delete,nullzero" json:"deletedAt,omitzero"`

	SrcUser    *User    `bun:"rel:has-one,join:src_id=id" json:"srcUser,omitempty"`
	SrcProject *Project `bun:"rel:has-one,join:src_id=id" json:"srcProject,omitempty"`
	SrcApp     *App     `bun:"rel:has-one,join:src_id=id" json:"srcApp,omitempty"`
}

func (lnk *ResLink) GetKey() string {
	return fmt.Sprintf("%s:%s:%s:%s", lnk.SrcType, lnk.SrcID, lnk.DstType, lnk.DstID)
}

type ResLinks []*ResLink

func (links ResLinks) GetDstIDsByDstType(dstType base.ResourceType) (resp []string) {
	for _, link := range links {
		if link.DstType == dstType {
			resp = append(resp, link.DstID)
		}
	}
	return resp
}

func (links ResLinks) GetDstIDByDstType(dstType base.ResourceType) (string, bool) {
	for _, link := range links {
		if link.DstType == dstType {
			return link.DstID, true
		}
	}
	return "", false
}
