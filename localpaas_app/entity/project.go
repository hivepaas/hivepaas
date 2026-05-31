package entity

import (
	"time"

	"github.com/localpaas/localpaas/localpaas_app/base"
)

var (
	ProjectUpsertingConflictCols = []string{"id"}
	ProjectUpsertingUpdateCols   = []string{"name", "key", "photo", "photo_id", "status", "note", "owner_id",
		"update_ver", "updated_at", "deleted_at"}
	ProjectDefaultExcludeColumns = []string{"note"}
)

type Project struct {
	ID        string             `bun:",pk" json:"id"`
	Name      string             `json:"name"`
	Key       string             `json:"key"`
	Photo     string             `bun:",nullzero" json:"photo,omitempty"`
	PhotoID   string             `bun:",nullzero" json:"photoId,omitempty"`
	Status    base.ProjectStatus `json:"status"`
	Note      string             `bun:",nullzero" json:"note,omitempty"`
	OwnerID   string             `json:"ownerId"`
	UpdateVer int                `json:"updateVer"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
	DeletedAt time.Time `bun:",soft_delete,nullzero" json:"deletedAt,omitzero"`

	PhotoData *BinObject       `bun:"rel:has-one,join:photo_id=id" json:"photoData,omitempty"`
	Owner     *User            `bun:"rel:has-one,join:owner_id=id" json:"owner,omitempty"`
	Settings  []*Setting       `bun:"rel:has-many,join:id=object_id" json:"settings,omitempty"`
	Apps      []*App           `bun:"rel:has-many,join:id=project_id" json:"apps,omitempty"`
	Tags      []*ProjectTag    `bun:"rel:has-many,join:id=project_id" json:"tags,omitempty"`
	Accesses  []*ACLPermission `bun:"rel:has-many,join:id=res_id" json:"accesses,omitempty"`
}

// GetID implements IDEntity interface
func (p *Project) GetID() string {
	return p.ID
}

// GetName implements NamedEntity interface
func (p *Project) GetName() string {
	return p.Name
}

func (p *Project) GetSettingScope() *base.ObjectScope {
	return &base.ObjectScope{
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
