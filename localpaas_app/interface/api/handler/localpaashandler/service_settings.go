package localpaashandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/lpappsettingsuc/lpappsettingsdto"
)

// GetServiceSettings Gets LocalPaaS service settings
// @Summary Gets LocalPaaS service settings
// @Description Gets LocalPaaS service settings
// @Tags    system_localpaas
// @Produce json
// @Id      getLocalPaaSServiceSettings
// @Success 200 {object} lpappsettingsdto.GetServiceSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/localpaas/service-settings [get]
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

	req := lpappsettingsdto.NewGetServiceSettingsReq()
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.lpAppSettingsUC.GetServiceSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateServiceSettings Updates LocalPaaS service settings
// @Summary Updates LocalPaaS service settings
// @Description Updates LocalPaaS service settings
// @Tags    system_localpaas
// @Produce json
// @Id      updateLocalPaaSServiceSettings
// @Param   body body lpappsettingsdto.UpdateServiceSettingsReq true "request data"
// @Success 200 {object} lpappsettingsdto.UpdateServiceSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/localpaas/service-settings [put]
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

	req := lpappsettingsdto.NewUpdateServiceSettingsReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.lpAppSettingsUC.UpdateServiceSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
