package devhelperhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/devhelperuc/devhelperdto"
)

// ExecuteCmd Executes a command
// @Summary Executes a command
// @Description Executes a command
// @Tags    dev_helper
// @Produce json
// @Id      devExecuteCmd
// @Param   body body devhelperdto.ExecuteCmdReq true "request data"
// @Success 200 {object} devhelperdto.ExecuteCmdResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /dev-helper/exec-cmd [post]
func (h *Handler) ExecuteCmd(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := devhelperdto.NewExecuteCmdReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.devHelperUC.ExecuteCmd(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
