package basesettinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/domainsettingsuc/domainsettingsdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/imagebuildsettingsuc/imagebuildsettingsdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/storagesettingsuc/storagesettingsdto"
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

	scope := &base.ObjectScope{}
	switch scopeType {
	case base.ObjectScopeGlobal:
		auth, _, err = h.GetAuthGlobalSettings(ctx, resType, base.ActionTypeWrite, "")
	case base.ObjectScopeProject:
		auth, scope.ProjectID, _, err = h.GetAuthProjectSettings(ctx, base.ActionTypeWrite, "")
	case base.ObjectScopeApp:
		auth, scope.ProjectID, scope.AppID, _, err = h.GetAuthAppSettings(ctx, base.ActionTypeWrite, "")
	case base.ObjectScopeUser:
		auth, scope.UserID, _, err = h.GetAuthUserSettings(ctx, base.ActionTypeWrite, "")
	default:
		err = apperrors.NewUnsupported("Setting scope 'none'")
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	var req any
	var ucFunc func() (any, error)
	reqCtx := h.RequestCtx(ctx)

	switch resType { //nolint:exhaustive
	case base.ResourceTypeImageBuildSettings:
		r := imagebuildsettingsdto.NewUpdateImageBuildSettingsReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.ImageBuildUC.UpdateImageBuildSettings(reqCtx, auth, r) }

	case base.ResourceTypeDomainSettings:
		r := domainsettingsdto.NewUpdateDomainSettingsReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.DomainSettingsUC.UpdateDomainSettings(reqCtx, auth, r) }

	case base.ResourceTypeStorageSettings:
		r := storagesettingsdto.NewUpdateStorageSettingsReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.StorageSettingsUC.UpdateStorageSettings(reqCtx, auth, r) }

	default:
		// NOTE: not implemented
		err = apperrors.NewNotImplementedNT()
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
