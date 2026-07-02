package entity

import (
	"time"
)

var (
	ProjectSharedSettingUpsertingConflictCols = []string{"project_id", "setting_id"}
	ProjectSharedSettingUpsertingUpdateCols   = []string{"data_view_allowed", "deleted_at"}
)

type ProjectSharedSetting struct {
	ProjectID       string `bun:",pk" json:"projectId"`
	SettingID       string `bun:",pk" json:"settingId"`
	DataViewAllowed bool   `json:"dataViewAllowed"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	DeletedAt time.Time `bun:",soft_delete,nullzero" json:"deletedAt,omitzero"`

	Project *Project `bun:"rel:has-one,join:project_id=id" json:"project"`
	Setting *Setting `bun:"rel:has-one,join:setting_id=id" json:"setting"`
}
