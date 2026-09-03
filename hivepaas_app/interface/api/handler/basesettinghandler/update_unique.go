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

type UpdateUniqueSettingOptions struct {
	PreRequestHandler func(auth *basedto.Auth, req any) error
}

type UpdateUniqueSettingOption func(*UpdateUniqueSettingOptions)

func UpdateUniqueSettingPreRequestHandler(fn func(auth *basedto.Auth, req any) error) UpdateUniqueSettingOption {
	return func(opts *UpdateUniqueSettingOptions) {
		opts.PreRequestHandler = fn
	}
}

func (h *Handler) UpdateUniqueSetting(
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
		r := appfeaturesettingsdto.NewUpdateAppFeatureSettingsReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.AppFeatureSettingsUC.UpdateAppFeatureSettings(reqCtx, auth, r) }

	case base.ResourceTypeAppPlacement:
		r := appplacementsettingsdto.NewUpdateAppPlacementSettingsReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.AppPlacementSettingsUC.UpdateAppPlacementSettings(reqCtx, auth, r) }

	case base.ResourceTypeDomainSettings:
		r := domainsettingsdto.NewUpdateDomainSettingsReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.DomainSettingsUC.UpdateDomainSettings(reqCtx, auth, r) }

	case base.ResourceTypeImageBuild:
		r := imagebuildsettingsdto.NewUpdateImageBuildSettingsReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.ImageBuildUC.UpdateImageBuildSettings(reqCtx, auth, r) }

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
