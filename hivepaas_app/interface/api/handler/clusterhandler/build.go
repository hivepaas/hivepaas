package clusterhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/builduc/builddto"
)

// ClearBuildCache Clears build cache
// @Summary Clears build cache
// @Description Clears build cache
// @Tags    cluster_build
// @Produce json
// @Id      clearClusterBuildCache
// @Param   body body builddto.ClearBuildCacheReq true "request data"
// @Success 200 {object} builddto.ClearBuildCacheResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /cluster/build/cache-clear [post]
func (h *Handler) ClearBuildCache(ctx *gin.Context) {
	auth, _, err := h.getAuth(ctx, base.ResourceTypeCluster, base.ActionTypeWrite, "")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := builddto.NewClearBuildCacheReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.buildUC.ClearBuildCache(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
