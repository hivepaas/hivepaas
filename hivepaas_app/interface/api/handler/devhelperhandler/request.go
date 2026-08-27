package devhelperhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/devhelperuc/devhelperdto"
)

// SimulateLongRequest Simulates a long-running request
// @Summary Simulates a long-running request
// @Description Simulates a long-running request
// @Tags    dev_helper
// @Produce json
// @Id      devSimulateLongRequest
// @Param   body body devhelperdto.LongRequestReq true "request data"
// @Success 200 {object} devhelperdto.LongRequestResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /dev-helper/long-req [post]
func (h *Handler) SimulateLongRequest(ctx *gin.Context) {
	req := devhelperdto.NewLongRequestReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.devHelperUC.LongRequest(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
