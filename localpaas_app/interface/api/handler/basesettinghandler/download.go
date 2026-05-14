package basesettinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/configfileuc/configfiledto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/secretuc/secretdto"
)

type DownloadOptions struct {
	PreRequestHandler func(auth *basedto.Auth, req any) error
}

type DownloadOption func(*DownloadOptions)

func DownloadPreRequestHandler(fn func(auth *basedto.Auth, req any) error) DownloadOption {
	return func(opts *DownloadOptions) {
		opts.PreRequestHandler = fn
	}
}

func (h *Handler) Download(
	ctx *gin.Context,
	resType base.ResourceType,
	scopeType base.SettingScopeType,
	dataType string,
	opts ...DownloadOption,
) {
	var auth *basedto.Auth
	var itemID string
	var err error

	options := &DownloadOptions{}
	for _, o := range opts {
		o(options)
	}

	scope := &base.SettingScope{}
	switch scopeType {
	case base.SettingScopeGlobal:
		itemID, err = h.GetParamGlobalSettings(ctx, "itemID")
	case base.SettingScopeProject:
		scope.ProjectID, itemID, err = h.GetParamProjectSettings(ctx, "itemID")
	case base.SettingScopeApp:
		scope.ProjectID, scope.AppID, itemID, err = h.GetParamAppSettings(ctx, "itemID")
	case base.SettingScopeUser:
		itemID, err = h.GetParamUserSettings(ctx, "itemID")
	default:
		err = apperrors.NewUnsupported("Setting scope 'none'")
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	var req any
	var ucFunc func() (*settings.BaseDownloadDataResp, error)
	reqCtx := h.RequestCtx(ctx)

	switch resType { //nolint:exhaustive
	case base.ResourceTypeSecret:
		r := secretdto.NewDownloadSecretReq()
		r.Scope, r.ID, r.DataType = scope, itemID, dataType
		req, ucFunc = r, func() (*settings.BaseDownloadDataResp, error) {
			resp, err := h.SecretUC.DownloadSecret(reqCtx, auth, r)
			if err != nil {
				return nil, apperrors.Wrap(err)
			}
			return resp.Data.BaseDownloadDataResp, nil
		}

	case base.ResourceTypeConfigFile:
		r := configfiledto.NewDownloadConfigFileReq()
		r.Scope, r.ID, r.DataType = scope, itemID, dataType
		req, ucFunc = r, func() (*settings.BaseDownloadDataResp, error) {
			resp, err := h.ConfigFileUC.DownloadConfigFile(reqCtx, auth, r)
			if err != nil {
				return nil, apperrors.Wrap(err)
			}
			return resp.Data.BaseDownloadDataResp, nil
		}

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

	defer resp.Content.Close()
	ctx.DataFromReader(http.StatusOK, resp.ContentLength, resp.ContentType, resp.Content, resp.ExtraHeaders)
}
