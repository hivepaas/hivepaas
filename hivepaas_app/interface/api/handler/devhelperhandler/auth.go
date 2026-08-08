package devhelperhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

func (h *Handler) DevModeLogin(ctx *gin.Context) {
	req := sessiondto.NewDevModeLoginReq()
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.sessionUC.DevModeLogin(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
