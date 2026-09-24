package hivepaashandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappuc/hpappdto"
)

// GetAppReleaseInfo Gets release info of the app
// @Summary Gets release info of the app
// @Description Gets release info of the app
// @Tags    system_hivepaas
// @Produce json
// @Id      getHivePaaSReleaseInfo
// @Success 200 {object} hpappdto.GetHpAppReleaseInfoResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/hivepaas/release-info [get]
func (h *Handler) GetAppReleaseInfo(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleSystem,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, hperrors.NewForbidden("Getting release info").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := hpappdto.NewGetHpAppReleaseInfoReq()
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.hpAppUC.GetHpAppReleaseInfo(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateAppVersion Updates HivePaaS app
// @Summary Updates HivePaaS app
// @Description Updates HivePaaS app
// @Tags    system_hivepaas
// @Produce json
// @Id      updateHivePaaSAppVersion
// @Param   body body hpappdto.UpdateHpAppReq true "request data"
// @Success 201 {object} hpappdto.UpdateHpAppResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/hivepaas/update-version [post]
func (h *Handler) UpdateAppVersion(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleSystem,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, hperrors.NewForbidden("Update app version").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := hpappdto.NewUpdateHpAppReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.hpAppUC.UpdateHpApp(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

// GetAppUpdatePlan Gets what an update to a version would do
// @Summary Gets what an update to a version would do
// @Description Gets, component by component, what an update to a published version would do - the image each
// @Description runs now and would run, and whether the update would move it - without changing anything.
// @Tags    system_hivepaas
// @Produce json
// @Id      getHivePaaSUpdatePlan
// @Param   targetVersion query string true "a version published on the stable or beta channel"
// @Success 200 {object} hpappdto.GetHpAppUpdatePlanResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/hivepaas/update-plan [get]
func (h *Handler) GetAppUpdatePlan(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleSystem,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, hperrors.NewForbidden("Getting the update plan").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := hpappdto.NewGetHpAppUpdatePlanReq()
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.hpAppUC.GetHpAppUpdatePlan(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
