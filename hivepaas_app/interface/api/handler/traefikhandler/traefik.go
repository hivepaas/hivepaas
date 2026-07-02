package traefikhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/traefikuc/traefikdto"
)

// ReloadTraefikConfig Reloads traefik config files
// @Summary Reloads traefik config files
// @Description Reloads traefik config files
// @Tags    system_traefik
// @Produce json
// @Id      reloadTraefikConfig
// @Param   body body traefikdto.ReloadTraefikConfigReq true "request data"
// @Success 200 {object} traefikdto.ReloadTraefikConfigResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/traefik/config/reload [post]
func (h *Handler) ReloadTraefikConfig(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		Action:         base.ActionTypeWrite,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, apperrors.NewForbidden("Reload traefik config").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := traefikdto.NewReloadTraefikConfigReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.traefikUC.ReloadTraefikConfig(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// ResetTraefikConfig Resets traefik config files
// @Summary Resets traefik config files
// @Description Resets traefik config files
// @Tags    system_traefik
// @Produce json
// @Id      resetTraefikConfig
// @Param   body body traefikdto.ResetTraefikConfigReq true "request data"
// @Success 200 {object} traefikdto.ResetTraefikConfigResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/traefik/config/reset [post]
func (h *Handler) ResetTraefikConfig(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		Action:         base.ActionTypeWrite,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, apperrors.NewForbidden("Reset traefik config").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := traefikdto.NewResetTraefikConfigReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.traefikUC.ResetTraefikConfig(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// RestartTraefik Restarts traefik containers
// @Summary Restarts traefik containers
// @Description Restarts traefik containers
// @Tags    system_traefik
// @Produce json
// @Id      restartTraefik
// @Param   body body traefikdto.RestartTraefikReq true "request data"
// @Success 200 {object} traefikdto.RestartTraefikResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/traefik/restart [post]
func (h *Handler) RestartTraefik(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		Action:         base.ActionTypeWrite,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, apperrors.NewForbidden("Restart traefik service").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := traefikdto.NewRestartTraefikReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.traefikUC.RestartTraefik(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
