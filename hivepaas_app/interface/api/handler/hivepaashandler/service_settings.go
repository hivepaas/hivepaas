package hivepaashandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc/hpappsettingsdto"
)

// GetServiceSettings Gets HivePaaS service settings
// @Summary Gets HivePaaS service settings
// @Description Gets HivePaaS service settings
// @Tags    system_hivepaas
// @Produce json
// @Id      getHivePaaSServiceSettings
// @Success 200 {object} hpappsettingsdto.GetServiceSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/hivepaas/service-settings [get]
func (h *Handler) GetServiceSettings(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		Action:         base.ActionTypeRead,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, apperrors.NewForbidden("Get service settings").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := hpappsettingsdto.NewGetServiceSettingsReq()
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.hpAppSettingsUC.GetServiceSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateServiceSettings Updates HivePaaS service settings
// @Summary Updates HivePaaS service settings
// @Description Updates HivePaaS service settings
// @Tags    system_hivepaas
// @Produce json
// @Id      updateHivePaaSServiceSettings
// @Param   body body hpappsettingsdto.UpdateServiceSettingsReq true "request data"
// @Success 200 {object} hpappsettingsdto.UpdateServiceSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/hivepaas/service-settings [put]
func (h *Handler) UpdateServiceSettings(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		Action:         base.ActionTypeWrite,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, apperrors.NewForbidden("Update service settings").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := hpappsettingsdto.NewUpdateServiceSettingsReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.hpAppSettingsUC.UpdateServiceSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
