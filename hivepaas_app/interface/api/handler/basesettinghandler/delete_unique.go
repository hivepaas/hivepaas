package basesettinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/appfeaturesettingsuc/appfeaturesettingsdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/domainsettingsuc/domainsettingsdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/imagebuildsettingsuc/imagebuildsettingsdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/storagesettingsuc/storagesettingsdto"
)

type DeleteUniqueSettingOptions struct {
	PreRequestHandler func(auth *basedto.Auth, req any) error
}

type DeleteUniqueSettingOption func(*DeleteUniqueSettingOptions)

func DeleteUniqueSettingPreRequestHandler(fn func(auth *basedto.Auth, req any) error) DeleteUniqueSettingOption {
	return func(opts *DeleteUniqueSettingOptions) {
		opts.PreRequestHandler = fn
	}
}

func (h *Handler) DeleteUniqueSetting(
	ctx *gin.Context,
	resType base.ResourceType,
	scopeType base.ObjectScopeType,
	opts ...DeleteUniqueSettingOption,
) {
	var auth *basedto.Auth
	var err error

	options := &DeleteUniqueSettingOptions{}
	for _, o := range opts {
		o(options)
	}

	scope := &base.ObjectScope{}
	switch scopeType {
	case base.ObjectScopeGlobal:
		auth, _, err = h.GetAuthGlobalSettings(ctx, resType, base.ActionTypeDelete, "")
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
	case base.ResourceTypeAppFeatures:
		r := appfeaturesettingsdto.NewDeleteAppFeatureSettingsReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.AppFeatureSettingsUC.DeleteAppFeatureSettings(reqCtx, auth, r) }

	case base.ResourceTypeDomainSettings:
		r := domainsettingsdto.NewDeleteDomainSettingsReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.DomainSettingsUC.DeleteDomainSettings(reqCtx, auth, r) }

	case base.ResourceTypeImageBuildSettings:
		r := imagebuildsettingsdto.NewDeleteImageBuildSettingsReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.ImageBuildUC.DeleteImageBuildSettings(reqCtx, auth, r) }

	case base.ResourceTypeStorageSettings:
		r := storagesettingsdto.NewDeleteStorageSettingsReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.StorageSettingsUC.DeleteStorageSettings(reqCtx, auth, r) }

	default:
		// NOTE: not implemented
		err = apperrors.NewNotImplementedNT()
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
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
