package clusterhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/builduc/builddto"
)

// ClearBuildCache Clears build cache
// @Summary Clears build cache
// @Description Clears build cache
// @Tags    cluster_build
// @Produce json
// @Id      clearClusterBuildCache
// @Param   body body builddto.ClearBuildCacheReq true "request data"
// @Success 200 {object} builddto.ClearBuildCacheResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
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
