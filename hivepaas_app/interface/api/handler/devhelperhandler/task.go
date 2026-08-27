package devhelperhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/devhelperuc/devhelperdto"
)

// LockTask Locks a task for a while
// @Summary Locks a task for a while
// @Description Locks a task for a while
// @Tags    dev_helper
// @Produce json
// @Id      devLockTask
// @Param   body body devhelperdto.LockTaskReq true "request data"
// @Success 200 {object} devhelperdto.LockTaskResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /dev-helper/lock-task [post]
func (h *Handler) LockTask(ctx *gin.Context) {
	req := devhelperdto.NewLockTaskReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.devHelperUC.LockTask(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
