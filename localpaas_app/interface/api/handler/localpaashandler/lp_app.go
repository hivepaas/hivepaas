package localpaashandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/lpappuc/lpappdto"
)

// ReloadLocalPaaSAppConfig Reloads LocalPaaS config files
// @Summary Reloads LocalPaaS config files
// @Description Reloads LocalPaaS config files
// @Tags    system_localpaas
// @Produce json
// @Id      reloadLocalPaaSAppConfig
// @Param   body body lpappdto.ReloadLpAppConfigReq true "request data"
// @Success 200 {object} lpappdto.ReloadLpAppConfigResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/localpaas/config/reload [post]
func (h *Handler) ReloadLocalPaaSAppConfig(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		Action:         base.ActionTypeWrite,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, apperrors.NewForbidden("Reload localpaas config").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := lpappdto.NewReloadLpAppConfigReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.lpAppUC.ReloadLpAppConfig(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// RestartLocalPaaSApp Restarts localpaas app containers
// @Summary Restarts localpaas app containers
// @Description Restarts localpaas app containers
// @Tags    system_localpaas
// @Produce json
// @Id      restartLocalPaaSApp
// @Param   body body lpappdto.RestartLpAppReq true "request data"
// @Success 200 {object} lpappdto.RestartLpAppResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/localpaas/restart [post]
func (h *Handler) RestartLocalPaaSApp(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		Action:         base.ActionTypeWrite,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, apperrors.NewForbidden("Restart localpaas service").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := lpappdto.NewRestartLpAppReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.lpAppUC.RestartLpApp(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
