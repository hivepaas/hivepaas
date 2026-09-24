package homehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/homeuc/homedto"
)

// GetHomeAttention Gets what needs attention
// @Summary Gets what needs attention
// @Description Gets what needs attention across the cluster - apps not running or restarting, nodes down
// @Description or given more memory than they have - narrowed to what the caller may see. Any signed-in
// @Description user may ask: each item is checked against the screen it leads to.
// @Tags    home
// @Produce json
// @Id      getHomeAttention
// @Success 200 {object} homedto.GetHomeAttentionResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /home/attention [get]
func (h *Handler) GetHomeAttention(ctx *gin.Context) {
	// No check here: what the page gathers spans every module, so each item is
	// checked on its own, in the usecase.
	auth, err := h.authHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.homeUC.GetHomeAttention(h.RequestCtx(ctx), auth)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
