package hivepaashandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc/hpappsettingsdto"
)

// GetRequestInfo Reports how a request reached HivePaaS
// @Summary Reports how a request reached HivePaaS
// @Description Echoes the proxy headers and the X-Forwarded-For chain this install receives,
// @Description and derives the proxyHops value that follows from it. Call it from outside the
// @Description cluster, over the path real traffic takes.
// @Tags    system_hivepaas
// @Produce json
// @Id      getHivePaaSRequestInfo
// @Success 200 {object} hpappsettingsdto.GetRequestInfoResp
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/hivepaas/request-info [get]
func (h *Handler) GetRequestInfo(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeRead},
		Module:          base.ResourceModuleSystem,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	// Admin only. The response describes the network path in front of this install,
	// which is reconnaissance for anyone deciding whether a forwarded header can be
	// forged - and it is of no use to anyone who is not configuring the proxy.
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, hperrors.NewForbidden("Get request info").
			WithMsgLog("only admin can perform this action"))
		return
	}

	resp := hpappsettingsdto.TransformRequestInfo(&hpappsettingsdto.RequestInfoTransformInput{
		RemoteAddr:   ctx.Request.RemoteAddr,
		ClientIP:     ctx.ClientIP(),
		HeaderValues: hpappsettingsdto.CollectProxyHeaders(ctx.Request.Header.Get),
	})

	ctx.JSON(http.StatusOK, &hpappsettingsdto.GetRequestInfoResp{Data: resp})
}
