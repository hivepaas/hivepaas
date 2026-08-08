package entity

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

var (
	ProjectEnvUpsertingConflictCols = []string{"id"}
	ProjectEnvUpsertingUpdateCols   = []string{"project_id", "name", "key", "status", "color", "index",
		"update_ver", "updated_at", "deleted_at"}
)

type ProjectEnv struct {
	ID        string             `bun:",pk" json:"id"`
	ProjectID string             `json:"projectId"`
	Name      string             `json:"name"`
	Key       string             `json:"key"`
	Status    base.ProjectStatus `json:"status"`
	Color     string             `json:"color"`
	Index     int                `json:"index"`
	UpdateVer int                `json:"updateVer"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
	DeletedAt time.Time `bun:",soft_delete,nullzero" json:"deletedAt,omitzero"`

	Project     *Project         `bun:"rel:has-one,join:project_id=id" json:"project,omitempty"`
	Settings    []*Setting       `bun:"rel:has-many,join:id=object_id" json:"settings,omitempty"`
	Apps        []*App           `bun:"rel:has-many,join:id=project_env_id" json:"apps,omitempty"`
	Accesses    []*ACLPermission `bun:"rel:has-many,join:id=res_id" json:"accesses,omitempty"`
	SrcResLinks []*ResLink       `bun:"rel:has-many,join:id=dst_id" json:"srcResLinks,omitempty"`
	DstResLinks []*ResLink       `bun:"rel:has-many,join:id=src_id" json:"dstResLinks,omitempty"`
}

// GetID implements IDEntity interface
func (p *ProjectEnv) GetID() string {
	return p.ID
}

// GetName implements NamedEntity interface
func (p *ProjectEnv) GetName() string {
	return p.Name
}

func (p *ProjectEnv) GetObjectScope() *ObjectScope {
	return &ObjectScope{
		ScopeType:    base.ObjectScopeProjectEnv,
		ProjectEnv:   p,
		ProjectEnvID: p.ID,
		Project:      p.Project,
		ProjectID:    p.ProjectID,
	}
}

func (p *ProjectEnv) GetSettingsByType(typ base.SettingType) (resp []*Setting) {
	for _, setting := range p.Settings {
		if setting.Type == typ {
			resp = append(resp, setting)
		}
	}
	return resp
}

func (p *ProjectEnv) GetSettingByType(typ base.SettingType) *Setting {
	for _, setting := range p.Settings {
		if setting.Type == typ {
			return setting
		}
	}
	return nil
}

func (p *ProjectEnv) GetChildAppsOfApp(appID string) (res []*App) {
	res = make([]*App, 0, 5) //nolint:mnd
	for _, app := range p.Apps {
		if app.ParentID != appID {
			continue
		}
		res = append(res, app)
	}
	return res
}
