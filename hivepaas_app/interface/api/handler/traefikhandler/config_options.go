package traefikhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/traefiksettingsuc/traefiksettingsdto"
)

// GetConfigOptions Gets Traefik config options
// @Summary Gets Traefik config options
// @Description Gets Traefik config options
// @Tags    system_traefik
// @Produce json
// @Id      getTraefikConfigOptions
// @Success 200 {object} traefiksettingsdto.GetConfigOptionsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/traefik/config-options [get]
func (h *Handler) GetConfigOptions(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeRead},
		Module:          base.ResourceModuleSystem,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, apperrors.NewForbidden("Get config options").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := traefiksettingsdto.NewGetConfigOptionsReq()
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.traefikSettingsUC.GetConfigOptions(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateConfigOptions Updates Traefik config options
// @Summary Updates Traefik config options
// @Description Updates Traefik config options
// @Tags    system_traefik
// @Produce json
// @Id      updateTraefikConfigOptions
// @Param   body body traefiksettingsdto.UpdateConfigOptionsReq true "request data"
// @Success 200 {object} traefiksettingsdto.UpdateConfigOptionsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/traefik/config-options [put]
func (h *Handler) UpdateConfigOptions(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleSystem,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, apperrors.NewForbidden("Update config options").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := traefiksettingsdto.NewUpdateConfigOptionsReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.traefikSettingsUC.UpdateConfigOptions(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
