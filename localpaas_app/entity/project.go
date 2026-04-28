package entity

import (
	"time"

	"github.com/localpaas/localpaas/localpaas_app/base"
)

var (
	ProjectUpsertingConflictCols = []string{"id"}
	ProjectUpsertingUpdateCols   = []string{"name", "key", "photo", "status", "note", "owner_id",
		"update_ver", "updated_at", "deleted_at"}
	ProjectDefaultExcludeColumns = []string{"note"}
)

type Project struct {
	ID        string             `bun:",pk" json:"id"`
	Name      string             `json:"name"`
	Key       string             `json:"key"`
	Photo     string             `bun:",nullzero" json:"photo"`
	Status    base.ProjectStatus `json:"status"`
	Note      string             `bun:",nullzero" json:"note"`
	OwnerID   string             `json:"ownerId"`
	UpdateVer int                `json:"updateVer"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
	DeletedAt time.Time `bun:",soft_delete,nullzero" json:"deletedAt"`

	Owner    *User            `bun:"rel:has-one,join:owner_id=id" json:"-"`
	Settings []*Setting       `bun:"rel:has-many,join:id=object_id" json:"-"`
	Apps     []*App           `bun:"rel:has-many,join:id=project_id" json:"-"`
	Tags     []*ProjectTag    `bun:"rel:has-many,join:id=project_id" json:"-"`
	Accesses []*ACLPermission `bun:"rel:has-many,join:id=resource_id" json:"-"`
}

// GetID implements IDEntity interface
func (p *Project) GetID() string {
	return p.ID
}

// GetName implements NamedEntity interface
func (p *Project) GetName() string {
	return p.Name
}

func (p *Project) GetSettingScope() *base.SettingScope {
	return &base.SettingScope{
		ProjectID: p.ID,
	}
}

func (p *Project) GetSettingsByType(typ base.SettingType) (resp []*Setting) {
	for _, setting := range p.Settings {
		if setting.Type == typ {
			resp = append(resp, setting)
		}
	}
	return resp
}

func (p *Project) GetSettingByType(typ base.SettingType) *Setting {
	for _, setting := range p.Settings {
		if setting.Type == typ {
			return setting
		}
	}
	return nil
}

func (p *Project) GetDefaultNetworkName() string {
	return p.Key + "_local_net"
}
