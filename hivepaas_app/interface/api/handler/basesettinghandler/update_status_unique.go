package basesettinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/appfeaturesettingsuc/appfeaturesettingsdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/appplacementsettingsuc/appplacementsettingsdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/domainsettingsuc/domainsettingsdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/imagebuildsettingsuc/imagebuildsettingsdto"
)

func (h *Handler) UpdateUniqueSettingStatus(
	ctx *gin.Context,
	resType base.ResourceType,
	scopeType base.ObjectScopeType,
	opts ...UpdateUniqueSettingOption,
) {
	var auth *basedto.Auth
	var err error

	options := &UpdateUniqueSettingOptions{}
	for _, o := range opts {
		o(options)
	}

	scope := &entity.ObjectScope{ScopeType: scopeType}
	switch scopeType {
	case base.ObjectScopeProject:
		auth, scope.ProjectID, _, err = h.GetAuthProjectSettings(ctx, base.ActionTypeWrite, "")
	case base.ObjectScopeProjectEnv:
		auth, scope.ProjectID, scope.ProjectEnvID, _, err = h.GetAuthProjectEnvSettings(ctx, base.ActionTypeWrite, "")
	case base.ObjectScopeApp:
		auth, scope.ProjectID, scope.ProjectEnvID, scope.AppID, _, err = h.GetAuthAppSettings(ctx, base.ActionTypeWrite, "")
	case base.ObjectScopeUser:
		auth, scope.UserID, _, err = h.GetAuthUserSettings(ctx, base.ActionTypeWrite, "")
	case base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
		auth, _, err = h.GetAuthGlobalSettings(ctx, resType, base.ActionTypeWrite, "")
	default:
		err = hperrors.NewUnsupported("Setting scope 'none'")
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	var req any
	var ucFunc func() (any, error)
	reqCtx := h.RequestCtx(ctx)

	switch resType { //nolint:exhaustive
	case base.ResourceTypeAppFeatures:
		r := appfeaturesettingsdto.NewUpdateAppFeatureSettingsStatusReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.AppFeatureSettingsUC.UpdateAppFeatureSettingsStatus(reqCtx, auth, r) }

	case base.ResourceTypeAppPlacement:
		r := appplacementsettingsdto.NewUpdateAppPlacementSettingsStatusReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) {
			return h.AppPlacementSettingsUC.UpdateAppPlacementSettingsStatus(reqCtx, auth, r)
		}

	case base.ResourceTypeDomainSettings:
		r := domainsettingsdto.NewUpdateDomainSettingsStatusReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.DomainSettingsUC.UpdateDomainSettingsStatus(reqCtx, auth, r) }

	case base.ResourceTypeImageBuild:
		r := imagebuildsettingsdto.NewUpdateImageBuildSettingsStatusReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.ImageBuildUC.UpdateImageBuildSettingsStatus(reqCtx, auth, r) }

	default:
		// NOTE: not implemented
		err = hperrors.NewNotImplementedNT()
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	if options.PreRequestHandler != nil {
		if err = options.PreRequestHandler(auth, req); err != nil {
			h.RenderError(ctx, err)
			return
		}
	}

	resp, err := ucFunc()
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
