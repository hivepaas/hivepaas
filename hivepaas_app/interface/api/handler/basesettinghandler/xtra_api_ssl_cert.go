package basesettinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslcertuc/sslcertdto"
)

func (h *Handler) SSLCertRenew(ctx *gin.Context, scopeType base.ObjectScopeType) {
	var auth *basedto.Auth
	var itemID string
	var err error

	scope := &base.ObjectScope{}
	switch scopeType {
	case base.ObjectScopeGlobal:
		auth, itemID, err = h.GetAuthGlobalSettings(ctx, base.ResourceTypeSSLCert, base.ActionTypeWrite, "itemID")
	case base.ObjectScopeProject:
		auth, scope.ProjectID, itemID, err = h.GetAuthProjectSettings(ctx, base.ActionTypeWrite, "itemID")
	case base.ObjectScopeApp:
		auth, scope.ProjectID, scope.AppID, itemID, err = h.GetAuthAppSettings(ctx, base.ActionTypeWrite, "itemID")
	case base.ObjectScopeUser:
		auth, scope.UserID, itemID, err = h.GetAuthUserSettings(ctx, base.ActionTypeWrite, "itemID")
	default:
		err = apperrors.NewUnsupported("Setting scope 'none'")
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := sslcertdto.NewRenewSSLCertReq()
	req.Scope = scope
	req.Type = base.SettingTypeSSLCert
	req.ID = itemID
	if err = h.ParseJSONBody(ctx, req); err != nil {
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
