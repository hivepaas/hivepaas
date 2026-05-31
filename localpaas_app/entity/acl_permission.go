package entity

import (
	"time"

	"github.com/localpaas/localpaas/localpaas_app/base"
)

var (
	ACLPermissionUpsertingConflictCols = []string{"subj_id", "res_id"}
	ACLPermissionUpsertingUpdateCols   = []string{"subj_type", "res_type",
		"p_read", "p_exec", "p_write", "p_del", "updated_at", "deleted_at"}
)

type ACLPermission struct {
	SubjectType  base.SubjectType   `bun:"subj_type" json:"subjectType"`
	SubjectID    string             `bun:"subj_id,pk" json:"subjectId"`
	ResourceType base.ResourceType  `bun:"res_type" json:"resourceType"`
	ResourceID   string             `bun:"res_id,pk" json:"resourceId"`
	Actions      base.AccessActions `bun:"embed:p_" json:"actions"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
	DeletedAt time.Time `bun:",soft_delete,nullzero" json:"deletedAt,omitzero"`

	SubjectUser    *User    `bun:"rel:has-one,join:subj_id=id" json:"subjectUser,omitempty"`
	SubjectProject *Project `bun:"rel:has-one,join:subj_id=id" json:"subjectProject,omitempty"`
	SubjectApp     *App     `bun:"rel:has-one,join:subj_id=id" json:"subjectApp,omitempty"`

	ResourceProject *Project `bun:"rel:has-one,join:res_id=id" json:"resourceProject,omitempty"`
	ResourceApp     *App     `bun:"rel:has-one,join:res_id=id" json:"resourceApp,omitempty"`
}
