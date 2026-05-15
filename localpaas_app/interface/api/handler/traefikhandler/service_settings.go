package traefikhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/traefiksettingsuc/traefiksettingsdto"
)

// GetServiceSettings Gets Traefik service settings
// @Summary Gets Traefik service settings
// @Description Gets Traefik service settings
// @Tags    system_traefik
// @Produce json
// @Id      getTraefikServiceSettings
// @Success 200 {object} traefiksettingsdto.GetServiceSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/traefik/service-settings [get]
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

	req := traefiksettingsdto.NewGetServiceSettingsReq()
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.traefikSettingsUC.GetServiceSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateServiceSettings Updates Traefik service settings
// @Summary Updates Traefik service settings
// @Description Updates Traefik service settings
// @Tags    system_traefik
// @Produce json
// @Id      updateTraefikServiceSettings
// @Param   body body traefiksettingsdto.UpdateServiceSettingsReq true "request data"
// @Success 200 {object} traefiksettingsdto.UpdateServiceSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/traefik/service-settings [put]
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

	req := traefiksettingsdto.NewUpdateServiceSettingsReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.traefikSettingsUC.UpdateServiceSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
