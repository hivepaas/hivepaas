package basesettinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/configfileuc/configfiledto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/secretuc/secretdto"
)

type GetDownloadTokenOptions struct {
	PreRequestHandler func(auth *basedto.Auth, req any) error
}

type GetDownloadTokenOption func(*GetDownloadTokenOptions)

func GetDownloadTokenPreRequestHandler(fn func(auth *basedto.Auth, req any) error) GetDownloadTokenOption {
	return func(opts *GetDownloadTokenOptions) {
		opts.PreRequestHandler = fn
	}
}

func (h *Handler) GetDownloadToken(
	ctx *gin.Context,
	resType base.ResourceType,
	scopeType base.ObjectScopeType,
	dataType string,
	expiration timeutil.Duration,
	opts ...GetDownloadTokenOption,
) {
	var auth *basedto.Auth
	var itemID string
	var err error

	options := &GetDownloadTokenOptions{}
	for _, o := range opts {
		o(options)
	}

	scope := &base.ObjectScope{}
	switch scopeType {
	case base.ObjectScopeGlobal:
		auth, itemID, err = h.GetAuthGlobalSettings(ctx, resType, base.ActionTypeRead, "itemID")
	case base.ObjectScopeProject:
		auth, scope.ProjectID, itemID, err = h.GetAuthProjectSettings(ctx, base.ActionTypeRead, "itemID")
	case base.ObjectScopeApp:
		auth, scope.ProjectID, scope.AppID, itemID, err = h.GetAuthAppSettings(ctx, base.ActionTypeRead, "itemID")
	case base.ObjectScopeUser:
		auth, scope.UserID, itemID, err = h.GetAuthUserSettings(ctx, base.ActionTypeRead, "itemID")
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
	case base.ResourceTypeConfigFile:
		r := configfiledto.NewGetDownloadTokenReq()
		r.Scope, r.ID, r.DataType, r.Expiration = scope, itemID, dataType, expiration
		req, ucFunc = r, func() (any, error) { return h.ConfigFileUC.GetDownloadToken(reqCtx, auth, r) }

	case base.ResourceTypeSecret:
		r := secretdto.NewGetDownloadTokenReq()
		r.Scope, r.ID, r.DataType, r.Expiration = scope, itemID, dataType, expiration
		req, ucFunc = r, func() (any, error) { return h.SecretUC.GetDownloadToken(reqCtx, auth, r) }

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
