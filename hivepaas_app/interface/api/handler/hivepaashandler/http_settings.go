package hivepaashandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc/hpappsettingsdto"
)

// GetRoutingSettings Gets HivePaaS routing settings
// @Summary Gets HivePaaS routing settings
// @Description Gets HivePaaS routing settings
// @Tags    system_hivepaas
// @Produce json
// @Id      getHivePaaSRoutingSettings
// @Success 200 {object} hpappsettingsdto.GetRoutingSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/hivepaas/http-settings [get]
func (h *Handler) GetRoutingSettings(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeRead},
		Module:          base.ResourceModuleSystem,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, apperrors.NewForbidden("Get routing settings").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := hpappsettingsdto.NewGetRoutingSettingsReq()
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

// UpdateRoutingSettings Updates HivePaaS routing settings
// @Summary Updates HivePaaS routing settings
// @Description Updates HivePaaS routing settings
// @Tags    system_hivepaas
// @Produce json
// @Id      updateHivePaaSRoutingSettings
// @Param   body body hpappsettingsdto.UpdateRoutingSettingsReq true "request data"
// @Success 200 {object} hpappsettingsdto.UpdateRoutingSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/hivepaas/http-settings [put]
func (h *Handler) UpdateRoutingSettings(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeRead},
		Module:          base.ResourceModuleSystem,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, apperrors.NewForbidden("Update routing settings").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := hpappsettingsdto.NewUpdateRoutingSettingsReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.hpAppSettingsUC.UpdateRoutingSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
