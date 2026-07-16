package hivepaashandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc/hpappsettingsdto"
)

// GetHttpSettings Gets HivePaaS HTTP settings
// @Summary Gets HivePaaS HTTP settings
// @Description Gets HivePaaS HTTP settings
// @Tags    system_hivepaas
// @Produce json
// @Id      getHivePaaSHttpSettings
// @Success 200 {object} hpappsettingsdto.GetHttpSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/hivepaas/http-settings [get]
func (h *Handler) GetHttpSettings(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		Action:         base.ActionTypeRead,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, apperrors.NewForbidden("Get HTTP settings").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := hpappsettingsdto.NewGetHttpSettingsReq()
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.hpAppSettingsUC.GetHttpSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateHttpSettings Updates HivePaaS HTTP settings
// @Summary Updates HivePaaS HTTP settings
// @Description Updates HivePaaS HTTP settings
// @Tags    system_hivepaas
// @Produce json
// @Id      updateHivePaaSHttpSettings
// @Param   body body hpappsettingsdto.UpdateHttpSettingsReq true "request data"
// @Success 200 {object} hpappsettingsdto.UpdateHttpSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/hivepaas/http-settings [put]
func (h *Handler) UpdateHttpSettings(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		Action:         base.ActionTypeWrite,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, apperrors.NewForbidden("Update HTTP settings").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := hpappsettingsdto.NewUpdateHttpSettingsReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.hpAppSettingsUC.UpdateHttpSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
