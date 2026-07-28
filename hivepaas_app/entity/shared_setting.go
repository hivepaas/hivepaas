package entity

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

var (
	SharedSettingUpsertingConflictCols = []string{"object_id", "setting_id"}
	SharedSettingUpsertingUpdateCols   = []string{"scope", "can_view_data", "deleted_at"}
)

type SharedSetting struct {
	Scope       base.ObjectScopeType `json:"scope"`
	ObjectID    string               `bun:",pk" json:"objectId"`
	SettingID   string               `bun:",pk" json:"settingId"`
	CanViewData bool                 `json:"canViewData"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	DeletedAt time.Time `bun:",soft_delete,nullzero" json:"deletedAt,omitzero"`

	Setting *Setting `bun:"rel:has-one,join:setting_id=id" json:"setting"`
}
