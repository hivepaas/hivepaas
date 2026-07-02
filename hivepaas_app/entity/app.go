package entity

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
)

var (
	AppUpsertingConflictCols = []string{"id"}
	AppUpsertingUpdateCols   = []string{"name", "key", "local_key", "project_id", "parent_id", "service_id",
		"status", "env", "note", "update_ver", "updated_at", "deleted_at"}
	AppDefaultExcludeColumns = []string{"note"}
)

type App struct {
	ID        string         `bun:",pk" json:"id"`
	Name      string         `json:"name"`
	Key       string         `json:"key"`
	LocalKey  string         `json:"localKey"`
	ProjectID string         `json:"projectId"`
	ParentID  string         `bun:",nullzero" json:"parentId,omitempty"`
	ServiceID string         `bun:",nullzero" json:"serviceId"`
	Status    base.AppStatus `json:"status"`
	Env       string         `bun:",nullzero" json:"env"`
	Note      string         `bun:",nullzero" json:"note,omitempty"`
	UpdateVer int            `json:"updateVer"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
	DeletedAt time.Time `bun:",soft_delete,nullzero" json:"deletedAt,omitzero"`

	Project     *Project   `bun:"rel:has-one,join:project_id=id" json:"project,omitempty"`
	ParentApp   *App       `bun:"rel:has-one,join:parent_id=id" json:"parentApp,omitempty"`
	Settings    []*Setting `bun:"rel:has-many,join:id=object_id" json:"settings,omitempty"`
	Tags        []*AppTag  `bun:"rel:has-many,join:id=app_id" json:"tags,omitempty"`
	SrcResLinks []*ResLink `bun:"rel:has-many,join:id=dst_id" json:"srcResLinks,omitempty"`
	DstResLinks []*ResLink `bun:"rel:has-many,join:id=src_id" json:"dstResLinks,omitempty"`
}

// GetID implements IDEntity interface
func (app *App) GetID() string {
	return app.ID
}

// GetName implements NamedEntity interface
func (app *App) GetName() string {
	return app.Name
}

func (app *App) GetObjectScope() *base.ObjectScope {
	return &base.ObjectScope{
		AppID:       app.ID,
		ParentAppID: app.ParentID,
		ProjectID:   app.ProjectID,
	}
}

func (app *App) GetSettingsByType(typ base.SettingType) (resp []*Setting) {
	for _, setting := range app.Settings {
		if setting.Type == typ {
			resp = append(resp, setting)
		}
	}
	return resp
}

func (app *App) GetSettingByType(typ base.SettingType) *Setting {
	for _, setting := range app.Settings {
		if setting.Type == typ {
			return setting
		}
	}
	return nil
}

func (app *App) GetAutoImageName() string {
	name := strings.NewReplacer("__", "_", "--", "-").Replace(app.Key)
	if len(name) > base.ImageNameMaxLen {
		name = name[:base.ImageNameMaxLen]
	}
	return name
}

func (app *App) TraefikConfigPath() string {
	return filepath.Join(config.Current.DataPathTraefikEtcDynamic().AbsPath(), app.Key+".yml")
}
