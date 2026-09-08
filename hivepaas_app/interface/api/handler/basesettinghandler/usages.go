package basesettinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type SettingUsageResp struct {
	Type base.ResourceType `json:"type"`
	ID   string            `json:"id"`
	Name string            `json:"name,omitempty"`

	SettingType base.SettingType     `json:"settingType,omitempty"`
	Scope       base.ObjectScopeType `json:"scope,omitempty"`

	// Enough for the dashboard to build the page this object lives on.
	// ProjectEnvKey rather than the env id, because that is what the URL carries.
	ProjectID     string `json:"projectId,omitempty"`
	ProjectEnvKey string `json:"projectEnvKey,omitempty"`
	AppID         string `json:"appId,omitempty"`
	AppName       string `json:"appName,omitempty"`
	UserID        string `json:"userId,omitempty"`
}

type GetSettingUsagesResp struct {
	Meta *basedto.Meta       `json:"meta"`
	Data []*SettingUsageResp `json:"data"`
}

// SettingUsages answers what still references a setting.
//
// A route of its own rather than a payload on the failed delete. The delete says
// only whether it happened - a 409 with ERR_SETTING_IN_USE when it did not - and
// leaves describing the blockers to a plain list resource, which is the shape the
// dashboard already knows how to fetch, cache and render. It is also useful before
// anybody tries to delete anything.
//
// It returns a handler rather than being one so the route can be registered under
// each settings group with that group's own resource type and scope. That is what
// makes it inherit exactly the authorization the group's other routes have,
// instead of deriving the scope from the setting row and inventing a second
// permission path to get wrong.
func (h *Handler) SettingUsages(resType base.ResourceType, scopeType base.ObjectScopeType) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		itemID, err := h.authorizeSettingUsages(ctx, resType, scopeType)
		if err != nil {
			h.RenderError(ctx, err)
			return
		}

		usages, err := h.BaseSettingUC.GetSettingUsages(h.RequestCtx(ctx), itemID)
		if err != nil {
			h.RenderError(ctx, err)
			return
		}

		data := make([]*SettingUsageResp, 0, len(usages))
		for _, usage := range usages {
			data = append(data, &SettingUsageResp{
				Type:          usage.Type,
				ID:            usage.ID,
				Name:          usage.Name,
				SettingType:   usage.SettingType,
				Scope:         usage.Scope,
				ProjectID:     usage.ProjectID,
				ProjectEnvKey: usage.ProjectEnvKey,
				AppID:         usage.AppID,
				AppName:       usage.AppName,
				UserID:        usage.UserID,
			})
		}

		ctx.JSON(http.StatusOK, &GetSettingUsagesResp{Data: data})
	}
}

// authorizeSettingUsages mirrors the scope switch the delete handler does, with
// read rather than write: seeing what uses a setting is a read of the setting.
func (h *Handler) authorizeSettingUsages(
	ctx *gin.Context,
	resType base.ResourceType,
	scopeType base.ObjectScopeType,
) (itemID string, err error) {
	var auth *basedto.Auth
	scope := &entity.ObjectScope{ScopeType: scopeType}

	switch scopeType {
	case base.ObjectScopeProject:
		auth, scope.ProjectID, itemID, err = h.GetAuthProjectSettings(ctx, base.ActionTypeRead, "itemID")
	case base.ObjectScopeProjectEnv:
		auth, scope.ProjectID, scope.ProjectEnvID, itemID, err = h.GetAuthProjectEnvSettings(ctx,
			base.ActionTypeRead, "itemID")
	case base.ObjectScopeApp:
		auth, scope.ProjectID, scope.ProjectEnvID, scope.AppID, itemID, err = h.GetAuthAppSettings(ctx,
			base.ActionTypeRead, "itemID")
	case base.ObjectScopeUser:
		auth, scope.UserID, itemID, err = h.GetAuthUserSettings(ctx, base.ActionTypeRead, "itemID")
	case base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
		auth, itemID, err = h.GetAuthGlobalSettings(ctx, resType, base.ActionTypeRead, "itemID")
	default:
		err = hperrors.NewUnsupported("Setting scope 'none'")
	}
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	// The scope is resolved for its permission checks, not for the query: a
	// setting id is unique on its own, and the checks above are what decide
	// whether this caller may look at it.
	_ = auth
	_ = scope

	return itemID, nil
}
