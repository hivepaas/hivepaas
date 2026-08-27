package devhelperhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/devhelperuc/devhelperdto"
)

// SimulateCrash Simulates a crash
// @Summary Simulates a crash
// @Description Simulates a crash
// @Tags    dev_helper
// @Produce json
// @Id      devSimulateCrash
// @Param   body body devhelperdto.SimulateCrashReq true "request data"
// @Success 200 {object} devhelperdto.SimulateCrashResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /dev-helper/crash-simulate [post]
func (h *Handler) SimulateCrash(ctx *gin.Context) {
	req := devhelperdto.NewSimulateCrashReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.devHelperUC.SimulateCrash(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
