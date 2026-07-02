package hivepaashandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
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
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/hivepaas/release-info [get]
func (h *Handler) GetAppReleaseInfo(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		Action:         base.ActionTypeWrite,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, apperrors.NewForbidden("Getting release info").
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
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/hivepaas/update-version [post]
func (h *Handler) UpdateAppVersion(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		Action:         base.ActionTypeWrite,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, apperrors.NewForbidden("Update app version").
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
