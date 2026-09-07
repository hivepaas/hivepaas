package hivepaashandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc/hpappsettingsdto"
)

// GetSecuritySettings Gets HivePaaS security settings
// @Summary Gets HivePaaS security settings
// @Description Gets HivePaaS security settings
// @Tags    system_hivepaas
// @Produce json
// @Id      getHivePaaSSecuritySettings
// @Success 200 {object} hpappsettingsdto.GetSecuritySettingsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/hivepaas/security-settings [get]
func (h *Handler) GetSecuritySettings(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeRead},
		Module:          base.ResourceModuleSystem,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, hperrors.NewForbidden("Get security settings").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := hpappsettingsdto.NewGetSecuritySettingsReq()
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.hpAppSettingsUC.GetSecuritySettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateSecuritySettings Updates HivePaaS security settings
// @Summary Updates HivePaaS security settings
// @Description Updates HivePaaS security settings
// @Tags    system_hivepaas
// @Produce json
// @Id      updateHivePaaSSecuritySettings
// @Param   body body hpappsettingsdto.UpdateSecuritySettingsReq true "request data"
// @Success 200 {object} hpappsettingsdto.UpdateSecuritySettingsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/hivepaas/security-settings [put]
func (h *Handler) UpdateSecuritySettings(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleSystem,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, hperrors.NewForbidden("Update security settings").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := hpappsettingsdto.NewUpdateSecuritySettingsReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.hpAppSettingsUC.UpdateSecuritySettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
