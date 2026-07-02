package hivepaashandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappuc/hpappdto"
)

// ReloadHivePaaSAppConfig Reloads HivePaaS config files
// @Summary Reloads HivePaaS config files
// @Description Reloads HivePaaS config files
// @Tags    system_hivepaas
// @Produce json
// @Id      reloadHivePaaSAppConfig
// @Param   body body hpappdto.ReloadHpAppConfigReq true "request data"
// @Success 200 {object} hpappdto.ReloadHpAppConfigResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/hivepaas/config/reload [post]
func (h *Handler) ReloadHivePaaSAppConfig(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		Action:         base.ActionTypeWrite,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, apperrors.NewForbidden("Reload hivepaas config").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := hpappdto.NewReloadHpAppConfigReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.hpAppUC.ReloadHpAppConfig(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// RestartHivePaaSApp Restarts hivepaas app containers
// @Summary Restarts hivepaas app containers
// @Description Restarts hivepaas app containers
// @Tags    system_hivepaas
// @Produce json
// @Id      restartHivePaaSApp
// @Param   body body hpappdto.RestartHpAppReq true "request data"
// @Success 200 {object} hpappdto.RestartHpAppResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/hivepaas/restart [post]
func (h *Handler) RestartHivePaaSApp(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		Action:         base.ActionTypeWrite,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, apperrors.NewForbidden("Restart hivepaas service").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := hpappdto.NewRestartHpAppReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.hpAppUC.RestartHpApp(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
