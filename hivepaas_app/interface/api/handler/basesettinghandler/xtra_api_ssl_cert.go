package basesettinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslcertuc/sslcertdto"
)

func (h *Handler) SSLCertRenew(ctx *gin.Context, scopeType base.ObjectScopeType) {
	var auth *basedto.Auth
	var itemID string
	var err error

	scope := &entity.ObjectScope{ScopeType: scopeType}
	switch scopeType {
	case base.ObjectScopeProject:
		auth, scope.ProjectID, itemID, err = h.GetAuthProjectSettings(ctx, base.ActionTypeWrite, "itemID")
	case base.ObjectScopeProjectEnv:
		auth, scope.ProjectID, scope.ProjectEnvID, itemID, err = h.GetAuthProjectEnvSettings(ctx,
			base.ActionTypeWrite, "itemID")
	case base.ObjectScopeApp:
		auth, scope.ProjectID, scope.ProjectEnvID, scope.AppID, itemID, err = h.GetAuthAppSettings(ctx,
			base.ActionTypeWrite, "itemID")
	case base.ObjectScopeUser:
		auth, scope.UserID, itemID, err = h.GetAuthUserSettings(ctx, base.ActionTypeWrite, "itemID")
	case base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
		auth, itemID, err = h.GetAuthGlobalSettings(ctx, base.ResourceTypeSSLCert, base.ActionTypeWrite, "itemID")
	default:
		err = hperrors.NewUnsupported("Setting scope 'none'")
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := sslcertdto.NewRenewSSLCertReq()
	req.Scope = scope
	req.Type = base.SettingTypeSSLCert
	req.ID = itemID
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.SSLCertUC.RenewSSLCert(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) SSLCertBundleDownload(ctx *gin.Context, scopeType base.ObjectScopeType) {
	var auth *basedto.Auth
	var itemID string
	var err error

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
		auth, itemID, err = h.GetAuthGlobalSettings(ctx, base.ResourceTypeSSLCert, base.ActionTypeRead, "itemID")
	default:
		err = hperrors.NewUnsupported("Setting scope 'none'")
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := sslcertdto.NewDownloadBundleReq()
	req.Scope = scope
	req.Type = base.SettingTypeSSLCert
	req.ID = itemID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.SSLCertUC.DownloadBundle(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	defer resp.Data.Content.Close()
	ctx.DataFromReader(http.StatusOK, resp.Data.ContentLength, resp.Data.ContentType,
		resp.Data.Content, resp.Data.ExtraHeaders)
}
