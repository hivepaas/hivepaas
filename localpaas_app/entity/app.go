package entity

import (
	"strings"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/base"
)

var (
	AppUpsertingConflictCols = []string{"id"}
	AppUpsertingUpdateCols   = []string{"name", "key", "project_id", "parent_id", "service_id",
		"status", "token", "note", "update_ver", "updated_at", "deleted_at"}
	AppDefaultExcludeColumns = []string{"note"}
)

const (
	appTokenLen = 24
)

type App struct {
	ID        string         `bun:",pk" json:"id"`
	Name      string         `json:"name"`
	Key       string         `json:"key"`
	ProjectID string         `json:"projectID"`
	ParentID  string         `bun:",nullzero" json:"parentID"`
	ServiceID string         `bun:",nullzero" json:"serviceID"`
	Status    base.AppStatus `json:"status"`
	Token     string         `json:"token"`
	Note      string         `bun:",nullzero" json:"note"`
	UpdateVer int            `json:"updateVer"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
	DeletedAt time.Time `bun:",soft_delete,nullzero" json:"deletedAt"`

	Project   *Project   `bun:"rel:has-one,join:project_id=id" json:"-"`
	ParentApp *App       `bun:"rel:has-one,join:parent_id=id" json:"-"`
	Settings  []*Setting `bun:"rel:has-many,join:id=object_id" json:"-"`
	Tags      []*AppTag  `bun:"rel:has-many,join:id=app_id" json:"-"`
}

// GetID implements IDEntity interface
func (app *App) GetID() string {
	return app.ID
}

// GetName implements NamedEntity interface
func (app *App) GetName() string {
	return app.Name
}

func (app *App) GetSettingScope() *base.SettingScope {
	return &base.SettingScope{
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

func (app *App) ResetToken() {
	app.Token = gofn.RandTokenAsHex(appTokenLen)
}
