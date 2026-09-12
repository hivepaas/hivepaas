package hivepaashandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/logginguc/loggingdto"
)

// GetLoggingSettings Gets HivePaaS logging settings
// @Summary Gets HivePaaS logging settings
// @Description Gets the logging configuration and what is actually running. Credentials always come back masked.
// @Tags    system_hivepaas
// @Produce json
// @Id      getHivePaaSLoggingSettings
// @Success 200 {object} loggingdto.GetSettingsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/hivepaas/logging-settings [get]
func (h *Handler) GetLoggingSettings(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeRead},
		Module:          base.ResourceModuleSystem,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, hperrors.NewForbidden("Get logging settings").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := loggingdto.NewGetSettingsReq()
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.loggingUC.GetSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateLoggingSettings Updates HivePaaS logging settings
// @Summary Updates HivePaaS logging settings
// @Description Stores the logging configuration and deploys, updates or removes the logging stack to match.
// @Description A credential sent as the masked placeholder keeps the stored value.
// @Tags    system_hivepaas
// @Produce json
// @Id      updateHivePaaSLoggingSettings
// @Param   body body loggingdto.UpdateSettingsReq true "request data"
// @Success 200 {object} loggingdto.UpdateSettingsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/hivepaas/logging-settings [put]
func (h *Handler) UpdateLoggingSettings(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleSystem,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, hperrors.NewForbidden("Update logging settings").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := loggingdto.NewUpdateSettingsReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.loggingUC.UpdateSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
