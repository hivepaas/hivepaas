package entity

import (
	"time"
)

var (
	SharedSettingUpsertingConflictCols = []string{"object_id", "setting_id"}
	SharedSettingUpsertingUpdateCols   = []string{"data_view_allowed", "deleted_at"}
)

type SharedSetting struct {
	ObjectID        string `bun:",pk" json:"objectId"`
	SettingID       string `bun:",pk" json:"settingId"`
	DataViewAllowed bool   `json:"dataViewAllowed"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	DeletedAt time.Time `bun:",soft_delete,nullzero" json:"deletedAt,omitzero"`

	Setting *Setting `bun:"rel:has-one,join:setting_id=id" json:"setting"`
}
