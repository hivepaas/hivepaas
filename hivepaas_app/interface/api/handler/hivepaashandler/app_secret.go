package hivepaashandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc/hpappsettingsdto"
)

// UpdateAppSecret Updates the HivePaaS app secret
// @Summary Updates the HivePaaS app secret
// @Description Replaces the key every stored secret is encrypted with, then re-encrypts them
// @Tags    system_hivepaas
// @Produce json
// @Id      updateHivePaaSAppSecret
// @Param   body body hpappsettingsdto.UpdateAppSecretReq true "request data"
// @Success 200 {object} hpappsettingsdto.UpdateAppSecretResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 409 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/hivepaas/app-secret [put]
func (h *Handler) UpdateAppSecret(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleSystem,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, hperrors.NewForbidden("Update app secret").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := hpappsettingsdto.NewUpdateAppSecretReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.hpAppSettingsUC.UpdateAppSecret(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
