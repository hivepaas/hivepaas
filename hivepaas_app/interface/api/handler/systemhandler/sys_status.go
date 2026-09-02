package systemhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/sysstatusuc/sysstatusdto"
)

// GetDBStats Gets database connection pool stats
// @Summary Gets database connection pool stats
// @Description Reports the connection pool counters of the process serving the request. WaitCount
// @Description and waitDuration say whether the pool is a bottleneck: while they stay near zero,
// @Description raising the pool size would change nothing.
// @Tags    system
// @Produce json
// @Id      getSystemDBStats
// @Success 200 {object} sysstatusdto.GetDBStatsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/status/db [get]
func (h *Handler) GetDBStats(ctx *gin.Context) {
	// Pool internals say something about how the deployment is built, so this is admin-only like
	// the rest of the system module.
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.GeneralResourceAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeRead},
		Module:          base.ResourceModuleSystem,
		ResourceType:    base.ResourceTypeSystemStatus,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.sysStatusUC.GetDBStats(h.RequestCtx(ctx), auth)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
