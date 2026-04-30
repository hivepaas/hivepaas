package systemhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/authhandler"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/lpappuc/lpappdto"
)

// GetLocalPaaSReleaseInfo Gets release info of the app
// @Summary Gets release info of the app
// @Description Gets release info of the app
// @Tags    system_localpaas_app
// @Produce json
// @Id      getLocalPaaSReleaseInfo
// @Success 200 {object} lpappdto.GetLpAppReleaseInfoResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/localpaas/release-info [get]
func (h *Handler) GetLocalPaaSReleaseInfo(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, apperrors.NewForbidden("Getting release info").
			WithMsgLog("only admin can get release info"))
		return
	}

	req := lpappdto.NewGetLpAppReleaseInfoReq()
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.lpAppUC.GetLpAppReleaseInfo(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateLocalPaaSAppVersion Updates LocalPaaS app
// @Summary Updates LocalPaaS app
// @Description Updates LocalPaaS app
// @Tags    system_localpaas_app
// @Produce json
// @Id      updateLocalPaaSAppVersion
// @Param   body body lpappdto.UpdateLpAppReq true "request data"
// @Success 201 {object} lpappdto.UpdateLpAppResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/localpaas/update-version [post]
func (h *Handler) UpdateLocalPaaSAppVersion(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, apperrors.NewForbidden("Update app version").
			WithMsgLog("only admin can update app version"))
		return
	}

	req := lpappdto.NewUpdateLpAppReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.lpAppUC.UpdateLpApp(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}
